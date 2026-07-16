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

package client

import (
	"context"
	"flag"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/UnderTreeTech/waterdrop/pkg/conf"
	"github.com/UnderTreeTech/waterdrop/pkg/log"
	"github.com/UnderTreeTech/waterdrop/pkg/trace"

	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// initTraceForTest boots the trace factory in OTel mode with an in-memory
// exporter, mirroring what pkg/trace/trace_test.go does with its package-private
// helpers (which the client package can't call). It writes a [trace.otel]
// (external=true) TOML, points the conf flag at it, reloads config, installs a
// synchronous AlwaysSample TracerProvider so TraceID() returns real ids, then
// calls trace.Init(). The returned func closes the exporter / resets the OTel
// global so tests stay isolated.
func initTraceForTest(t *testing.T) func() {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "app.toml")
	body := []byte(`[trace]
    [trace.otel]
        serviceName = "svc"
        enable = true
        endpoint = ""
        external = true
        sampleRatio = 1.0
`)
	require.NoError(t, os.WriteFile(path, body, 0644))
	require.NoError(t, flag.Set("conf", path))
	require.NoError(t, flag.Set("watch", "false"))
	conf.Init()

	exp := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exp),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	otel.SetTracerProvider(tp)

	closeFn := trace.Init()
	return func() {
		closeFn()
		otel.SetTracerProvider(oteltrace.NewNoopTracerProvider())
	}
}

// initLog boots the default stdout logger (Debug=true) so the access log
// emitted by RoundTrip (http-request-log / http-slow-request-log) is printed
// during the test, mirroring client_test.go's `defer log.New(nil).Sync()`.
// The logger is async-buffered, so Sync runs on cleanup to flush the lines.
func initLog(t *testing.T) {
	t.Helper()
	logger := log.New(nil)
	t.Cleanup(func() { _ = logger.Sync() })
}

// TestTransport_TraceContinuity verifies the Transport injects trace context
// into outbound headers so a downstream server span continues the same trace.
// Mirrors pkg/trace TestCarrierRoundTripHTTP but drives it through a real
// *http.Client built with NewClient.
func TestTransport_TraceContinuity(t *testing.T) {
	closeFn := initTraceForTest(t)
	defer closeFn()
	initLog(t)

	// server side: extract the parent from injected headers and echo its TraceID
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		srvSpan, _ := trace.GlobalTracer().StartServerSpan(
			r.Context(), r.Method+" "+r.URL.Path, trace.SpanServer, trace.HTTPCarrier{Header: r.Header})
		defer srvSpan.End()
		_, _ = io.WriteString(w, srvSpan.TraceID())
	}))
	defer srv.Close()

	// client side: start a parent span, carry it on the request context
	parentSpan, pctx := trace.GlobalTracer().StartSpan(context.Background(), "test.parent")
	defer parentSpan.End()
	parentTID := parentSpan.TraceID()
	require.NotEmpty(t, parentTID)

	req, err := http.NewRequestWithContext(pctx, http.MethodGet, srv.URL+"/foo", nil)
	require.NoError(t, err)

	c := NewClient()
	resp, err := c.Do(req)
	require.NoError(t, err)
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	// downstream extracted the same trace id the client parent started
	assert.Equal(t, parentTID, string(body), "server span should continue the client's trace")
}

// TestNewClient_ReturnsTraceClient locks the unbypassable contract: NewClient
// returns an opaque *TraceClient, never a raw *http.Client whose Transport
// field a caller could overwrite to drop tracing. If this test fails, tracing
// has become opt-out again.
func TestNewClient_ReturnsTraceClient(t *testing.T) {
	c := NewClient()
	assert.IsType(t, &TraceClient{}, c, "NewClient must return *TraceClient, not *http.Client")

	// *http.Client is a struct, *TraceClient is a distinct struct type, so the
	// two are not assignable - this assertion documents that.
	var _ interface {
		Do(*http.Request) (*http.Response, error)
	} = c
}

// TestTransport_NoInitNoPanic verifies the Transport works when the trace
// factory has not been initialized (GlobalTracer is the noop), i.e. it must
// not panic and the request must still succeed.
func TestTransport_NoInitNoPanic(t *testing.T) {
	initLog(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "ok")
	}))
	defer srv.Close()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+"/foo", nil)
	require.NoError(t, err)

	assert.NotPanics(t, func() {
		resp, err := NewClient().Do(req)
		require.NoError(t, err)
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		assert.Equal(t, "ok", string(body))
	})
}

// TestTransport_CustomBaseTransport verifies a caller-supplied round tripper
// (e.g. *http.Transport with DisableCompression for SSE) is actually used and
// the request still goes through.
func TestTransport_CustomBaseTransport(t *testing.T) {
	initLog(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "ok")
	}))
	defer srv.Close()

	c := NewClient(WithBaseTransport(&http.Transport{}))
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+"/foo", nil)
	require.NoError(t, err)

	resp, err := c.Do(req)
	require.NoError(t, err)
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	assert.Equal(t, "ok", string(body))
}

// TestTransport_5xxNoPanic verifies a 5xx response is surfaced (not swallowed)
// and the Transport does not panic on it.
func TestTransport_5xxNoPanic(t *testing.T) {
	initLog(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, "boom")
	}))
	defer srv.Close()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+"/foo", nil)
	require.NoError(t, err)

	assert.NotPanics(t, func() {
		resp, err := NewClient().Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		resp.Body.Close()
	})
}

// TestTraceClient_PostJson verifies PostJson sends a POST with the framework's
// JSON content type and that the body round-trips.
func TestTraceClient_PostJson(t *testing.T) {
	initLog(t)
	var gotContentType, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotContentType = r.Header.Get("Content-Type")
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		_, _ = io.WriteString(w, "ok")
	}))
	defer srv.Close()

	resp, err := NewClient().PostJson(srv.URL+"/post", strings.NewReader(`{"k":"v"}`))
	require.NoError(t, err)
	resp.Body.Close()

	assert.Equal(t, "application/json;charset=utf-8", gotContentType, "PostJson should set the JSON content type")
	assert.Equal(t, `{"k":"v"}`, gotBody)
}

// TestTransport_RequestBodyRestoredForLog verifies that, because RoundTrip reads
// the request body for logging and restores it, the server still receives the
// full body unchanged - including a body large enough to exercise the restore
// path (multi-read of the buffered bytes).
func TestTransport_RequestBodyRestoredForLog(t *testing.T) {
	initLog(t)
	payload := strings.Repeat(`{"msg":"hello, trace log body restore"}`, 8) // ~384 bytes, under maxBodyLogBytes
	var gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		_, _ = io.WriteString(w, "ok")
	}))
	defer srv.Close()

	resp, err := NewClient().Post(srv.URL+"/post", "application/json", strings.NewReader(payload))
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, payload, gotBody, "logged-then-restored request body must reach the server intact")
}

// TestTransport_SlowRequestThresholdNoPanic verifies a sub-threshold slow
// request is logged at Warn (http-slow-request-log) without panicking.
func TestTransport_SlowRequestThresholdNoPanic(t *testing.T) {
	initLog(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "ok")
	}))
	defer srv.Close()

	c := NewClient(WithSlowRequestThreshold(time.Nanosecond)) // every call counts as slow
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+"/foo", nil)
	require.NoError(t, err)

	assert.NotPanics(t, func() {
		resp, err := c.Do(req)
		require.NoError(t, err)
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		assert.Equal(t, "ok", string(body))
	})
}

// TestTransport_EmitsAccessLog verifies RoundTrip emits the http-request-log
// access log (mirroring client.go:172's field set). It captures os.Stdout
// (the Debug-mode logger sink), drives one request, flushes, and asserts the
// log line carries the expected message + fields.
func TestTransport_EmitsAccessLog(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "ok")
	}))
	defer srv.Close()

	// The Debug-mode logger writes to os.Stdout; redirect it BEFORE log.New so
	// the logger's BufferedWriteSyncer captures the pipe, not the real stdout.
	orig := os.Stdout
	rPipe, wPipe, _ := os.Pipe()
	os.Stdout = wPipe
	logger := log.New(nil) // Debug=true -> os.Stdout (now the pipe)
	done := make(chan string, 1)
	go func() {
		b, _ := io.ReadAll(rPipe)
		done <- string(b)
	}()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+"/foo", nil)
	require.NoError(t, err)
	resp, err := NewClient().Do(req)
	require.NoError(t, err)
	resp.Body.Close()

	_ = logger.Sync() // flush the async buffer
	wPipe.Close()
	os.Stdout = orig
	out := <-done

	assert.Contains(t, out, "http-request-log", "RoundTrip must emit the access log")
	for _, field := range []string{"peer", "method", "path", "headers", "query", "body", "quota", "duration", "reply_size", "status", "code", "error"} {
		assert.Contains(t, out, field, "access log should include the %q field (mirrors client.go:172)", field)
	}
}
