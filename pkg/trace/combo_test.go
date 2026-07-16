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
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

// allModes runs the cross-transport combination tests across all three backends.
var allModes = []string{"jaeger", "otel", "dual"}

// startInbound starts an inbound server span (simulating a request entering the
// service). transport is "http" or "grpc" and determines the carrier type (none
// for inbound - no extraction needed, we start a root server span).
func startInbound(t *testing.T, mode, transport, name string) (Span, context.Context, string) {
	t.Helper()
	closeFn, _ := initModeWithExporter(t, mode)
	t.Cleanup(func() { closeFn(); resetGlobal() })

	var carrier Carrier
	if transport == "http" {
		carrier = HTTPCarrier{Header: http.Header{}}
	} else {
		carrier = GRPCCarrier{MD: metadata.New(nil)}
	}
	span, ctx := GlobalTracer().StartServerSpan(context.Background(), name, SpanServer, carrier)
	tid := TraceID(ctx)
	require.NotEmpty(t, tid, "inbound server span must have a trace id (%s/%s)", mode, transport)
	return span, ctx, tid
}

// startOutbound simulates the service making an outbound call. transport is
// "http" or "grpc". Returns the client span, the carrier (with injected trace
// context), and the client span's trace id.
func startOutbound(t *testing.T, ctx context.Context, transport, name string) (Span, Carrier, string) {
	t.Helper()
	var carrier Carrier
	if transport == "http" {
		carrier = HTTPCarrier{Header: http.Header{}}
	} else {
		carrier = GRPCCarrier{MD: metadata.New(nil)}
	}
	span, _ := GlobalTracer().StartClientSpan(ctx, name, SpanClient, carrier)
	ctid := span.TraceID()
	return span, carrier, ctid
}

// startDownstream simulates the downstream service receiving the outbound call.
// It extracts the trace context from carrier and starts a server span. Returns
// the downstream server span's trace id.
func startDownstream(t *testing.T, mode, transport, name string, carrier Carrier) (Span, string) {
	t.Helper()
	span, ctx := GlobalTracer().StartServerSpan(context.Background(), name, SpanServer, carrier)
	return span, TraceID(ctx)
}

// assertSameTrace asserts the inbound, outbound-client, and downstream-server
// all share one trace (after left-pad normalization for jaeger<->otel crossing).
func assertSameTrace(t *testing.T, inboundTID, clientTID, downstreamTID, mode string) {
	t.Helper()
	require.NotEmpty(t, clientTID, "client span trace id non-empty")
	require.NotEmpty(t, downstreamTID, "downstream trace id non-empty")
	// same mode: ids should be identical
	assert.Equal(t, normalizeTID(inboundTID), normalizeTID(clientTID),
		"client span inherits inbound server span (%s)", mode)
	assert.Equal(t, normalizeTID(inboundTID), normalizeTID(downstreamTID),
		"downstream continues the same trace (%s)", mode)
	t.Logf("[%s] inbound=%s client=%s downstream=%s",
		mode, inboundTID, clientTID, downstreamTID)
}

// TestComboHTTPToRPC: HTTP request handler makes an RPC call to another service.
// inbound=HTTP server span -> outbound=gRPC client span -> downstream=gRPC server span.
func TestComboHTTPToRPC(t *testing.T) {
	for _, mode := range allModes {
		t.Run(mode, func(t *testing.T) {
			inSpan, ctx, inTID := startInbound(t, mode, "http", "POST /handler")
			cSpan, carrier, cTID := startOutbound(t, ctx, "grpc", "/svc/Method")
			dSpan, dTID := startDownstream(t, mode, "grpc", "/svc/Method", carrier)
			assertSameTrace(t, inTID, cTID, dTID, mode)
			cSpan.End()
			dSpan.End()
			inSpan.End()
		})
	}
}

// TestComboRPCToHTTP: RPC handler makes an HTTP call to another service.
// inbound=gRPC server span -> outbound=HTTP client span -> downstream=HTTP server span.
func TestComboRPCToHTTP(t *testing.T) {
	for _, mode := range allModes {
		t.Run(mode, func(t *testing.T) {
			inSpan, ctx, inTID := startInbound(t, mode, "grpc", "/svc/Method")
			cSpan, carrier, cTID := startOutbound(t, ctx, "http", "POST /api")
			dSpan, dTID := startDownstream(t, mode, "http", "POST /api", carrier)
			assertSameTrace(t, inTID, cTID, dTID, mode)
			cSpan.End()
			dSpan.End()
			inSpan.End()
		})
	}
}

// TestComboHTTPToHTTP: HTTP handler makes an HTTP call to another service.
// inbound=HTTP server span -> outbound=HTTP client span -> downstream=HTTP server span.
func TestComboHTTPToHTTP(t *testing.T) {
	for _, mode := range allModes {
		t.Run(mode, func(t *testing.T) {
			inSpan, ctx, inTID := startInbound(t, mode, "http", "POST /handler")
			cSpan, carrier, cTID := startOutbound(t, ctx, "http", "POST /downstream")
			dSpan, dTID := startDownstream(t, mode, "http", "POST /downstream", carrier)
			assertSameTrace(t, inTID, cTID, dTID, mode)
			cSpan.End()
			dSpan.End()
			inSpan.End()
		})
	}
}

// TestComboRPCToRPC: RPC handler makes an RPC call to another service.
// inbound=gRPC server span -> outbound=gRPC client span -> downstream=gRPC server span.
func TestComboRPCToRPC(t *testing.T) {
	for _, mode := range allModes {
		t.Run(mode, func(t *testing.T) {
			inSpan, ctx, inTID := startInbound(t, mode, "grpc", "/svc/A")
			cSpan, carrier, cTID := startOutbound(t, ctx, "grpc", "/svc/B")
			dSpan, dTID := startDownstream(t, mode, "grpc", "/svc/B", carrier)
			assertSameTrace(t, inTID, cTID, dTID, mode)
			cSpan.End()
			dSpan.End()
			inSpan.End()
		})
	}
}

// TestStartClientSpanInheritsParent is the direct regression test for the bug
// where jaegerTracer.StartClientSpan (and dualTracer's jaeger side) did not
// inherit the ctx parent, starting a root span instead. Asserts the client span
// trace id == the server span trace id in all three modes.
func TestStartClientSpanInheritsParent(t *testing.T) {
	for _, mode := range allModes {
		t.Run(mode, func(t *testing.T) {
			closeFn, _ := initModeWithExporter(t, mode)
			defer func() { closeFn(); resetGlobal() }()

			srvSpan, ctx := GlobalTracer().StartServerSpan(context.Background(), "inbound", SpanServer, nil)
			srvTID := TraceID(ctx)
			require.NotEmpty(t, srvTID)

			// client span should inherit the server span as parent
			cliSpan, _ := GlobalTracer().StartClientSpan(ctx, "outbound", SpanClient, nil)
			cliTID := cliSpan.TraceID()
			cliSpan.End()
			srvSpan.End()

			assert.Equal(t, normalizeTID(srvTID), normalizeTID(cliTID),
				"client span must inherit server span's trace (not start a root) in %s mode", mode)
		})
	}
}
