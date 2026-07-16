/*
 *
 * Copyright 2026 waterdrop authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package trace

import (
	"context"
	"flag"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/UnderTreeTech/waterdrop/pkg/conf"
	"google.golang.org/grpc/metadata"
)

// modePair enumerates the client->server trace backend combinations to test.
var modePairs = []struct{ client, server string }{
	{"jaeger", "jaeger"},
	{"jaeger", "otel"},
	{"otel", "jaeger"},
	{"otel", "otel"},
}

// initMode writes a config that selects the given mode and runs the factory.
// jaeger mode: [trace.jaeger] only. otel mode: [trace.otel] external=true. The
// caller is responsible for installing a global in-memory OTel provider (via
// setOtelInMemoryProvider) BEFORE calling initMode for otel/dual modes, so the
// returned exporter matches the provider the factory wires to.
func initMode(t *testing.T, mode string) func() {
	t.Helper()
	var body string
	switch mode {
	case "jaeger":
		body = `
[trace]
    [trace.jaeger]
        serviceName = "svc"
        samplerType = "const"
        samplerParam = 1
        agentAddr = "127.0.0.1:6831"
`
	case "otel":
		body = `
[trace]
    [trace.otel]
        serviceName = "svc"
        enable = true
        endpoint = ""
        external = true
        sampleRatio = 1.0
`
	case "dual":
		body = `
[trace]
    [trace.jaeger]
        serviceName = "svc"
        samplerType = "const"
        samplerParam = 1
        agentAddr = "127.0.0.1:6831"
    [trace.otel]
        serviceName = "svc"
        enable = true
        endpoint = ""
        external = true
        sampleRatio = 1.0
`
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "app.toml")
	require.NoError(t, os.WriteFile(path, []byte(body), 0644))
	require.NoError(t, flag.Set("conf", path))
	require.NoError(t, flag.Set("watch", "false"))
	conf.Init()
	return Init()
}

// initModeWithExporter runs initMode and, for otel/dual modes, installs the
// in-memory provider first and returns its exporter so callers can inspect
// recorded spans. For jaeger mode the returned exporter is nil.
func initModeWithExporter(t *testing.T, mode string) (func(), *tracetest.InMemoryExporter) {
	t.Helper()
	var exp *tracetest.InMemoryExporter
	if mode == "otel" || mode == "dual" {
		exp = setOtelInMemoryProvider()
	}
	closeFn := initMode(t, mode)
	return closeFn, exp
}

// expectTraceIDMatch asserts the client and server share one trace. jaeger
// ids are 64-bit (16 hex); otel ids are 128-bit (32 hex). When a jaeger side
// meets an otel side, the otel side left-pads the 64-bit id to 128-bit, so the
// ids match after normalizing to the longer (padded) form.
func expectTraceIDMatch(t *testing.T, clientTID, serverTID string) {
	t.Helper()
	require.NotEmpty(t, clientTID, "client trace id must not be empty (log trace_id would be empty)")
	require.NotEmpty(t, serverTID, "server trace id must not be empty (log trace_id would be empty)")

	c, s := clientTID, serverTID
	if len(c) < len(s) {
		c = strings.Repeat("0", len(s)-len(c)) + c
	} else if len(s) < len(c) {
		s = strings.Repeat("0", len(c)-len(s)) + s
	}
	t.Logf("client(%d hex)=%s  server(%d hex)=%s", len(clientTID), clientTID, len(serverTID), serverTID)
	assert.Equal(t, c, s, "client and server are on the same trace (after left-pad normalization)")
}

// TestE2EHTTP exercises the HTTP client->server trace propagation across all
// four mode combinations. The client starts a client span and injects into an
// http.Header (the path the HTTP client + middleware use); the server extracts
// from the same headers and starts a server span (the path the Trace middleware
// uses). Both trace ids must be non-empty (so log.go's trace_id prints) and
// belong to one trace.
func TestE2EHTTP(t *testing.T) {
	for _, mp := range modePairs {
		t.Run(mp.client+"->"+mp.server, func(t *testing.T) {
			defer resetGlobal()

			// client side
			clientClose, _ := initModeWithExporter(t, mp.client)
			defer clientClose()
			hdr := http.Header{}
			cSpan, _ := GlobalTracer().StartClientSpan(context.Background(), "http.client", SpanClient, HTTPCarrier{Header: hdr})
			clientTID := cSpan.TraceID()
			require.NotEmpty(t, clientTID)
			cSpan.End()

			// server side (re-init to server mode, like a separate process)
			serverClose, _ := initModeWithExporter(t, mp.server)
			defer serverClose()
			sSpan, serverCtx := GlobalTracer().StartServerSpan(context.Background(), "http.server", SpanServer, HTTPCarrier{Header: hdr})
			defer sSpan.End()
			serverTID := TraceID(serverCtx)

			expectTraceIDMatch(t, clientTID, serverTID)
		})
	}
}

// TestE2EGRPC exercises the gRPC client->server trace propagation across all
// four mode combinations, over a gRPC metadata.MD (the carrier the real
// TraceForUnaryClient/Server interceptors use). Same assertions as HTTP.
func TestE2EGRPC(t *testing.T) {
	for _, mp := range modePairs {
		t.Run(mp.client+"->"+mp.server, func(t *testing.T) {
			defer resetGlobal()

			// client side
			clientClose, _ := initModeWithExporter(t, mp.client)
			defer clientClose()
			md := metadata.New(nil)
			cSpan, _ := GlobalTracer().StartClientSpan(context.Background(), "grpc.client", SpanClient, GRPCCarrier{MD: md})
			clientTID := cSpan.TraceID()
			require.NotEmpty(t, clientTID)
			cSpan.End()

			// server side
			serverClose, _ := initModeWithExporter(t, mp.server)
			defer serverClose()
			sSpan, serverCtx := GlobalTracer().StartServerSpan(context.Background(), "grpc.server", SpanServer, GRPCCarrier{MD: md})
			defer sSpan.End()
			serverTID := TraceID(serverCtx)

			expectTraceIDMatch(t, clientTID, serverTID)
		})
	}
}

// TestE2ELogTraceID asserts that trace.TraceID(ctx) - which log.go uses for the
// "trace_id" log field - is non-empty inside a server span in each mode, so
// logs printed from a request handler carry a trace_id. Covers the "日志能打印出
// trace_id" requirement directly.
func TestE2ELogTraceID(t *testing.T) {
	for _, mode := range []string{"jaeger", "otel"} {
		t.Run(mode, func(t *testing.T) {
			defer resetGlobal()
			closeFn, _ := initModeWithExporter(t, mode)
			defer closeFn()

			span, ctx := GlobalTracer().StartServerSpan(context.Background(), "op", SpanServer, nil)
			defer span.End()
			tid := TraceID(ctx)
			t.Logf("mode=%s  log trace_id=%s (%d hex)", mode, tid, len(tid))
			assert.NotEmpty(t, tid, "log trace_id would be empty in %s mode", mode)
			assert.Equal(t, tid, span.TraceID(), "TraceID(ctx) == span.TraceID()")
		})
	}
}
