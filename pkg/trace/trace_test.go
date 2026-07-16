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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	jaegerprop "go.opentelemetry.io/contrib/propagators/jaeger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	oteltrace "go.opentelemetry.io/otel/trace"

	jaeger "github.com/uber/jaeger-client-go"
	jconfig "github.com/uber/jaeger-client-go/config"

	"github.com/UnderTreeTech/waterdrop/pkg/conf"
)

// writeConfig writes a TOML config to a temp file and points the conf flag at
// it, then (re)loads the global config. Returns the shutdown func from
// trace.Init (the factory under test).
func writeConfig(t *testing.T, body string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "app.toml")
	require.NoError(t, os.WriteFile(path, []byte(body), 0644))
	require.NoError(t, flag.Set("conf", path))
	require.NoError(t, flag.Set("watch", "false"))
	conf.Init()
}

// resetGlobal sets the factory's globalTracer back to noop and clears the
// global opentracing/otel state so tests are isolated.
func resetGlobal() {
	globalTracer = noopTracer{}
	opentracing.SetGlobalTracer(opentracing.NoopTracer{})
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator())
	otel.SetTracerProvider(oteltrace.NewNoopTracerProvider())
}

// setOtelInMemoryProvider installs a global OTel TracerProvider backed by an
// in-memory span exporter, returning the exporter so tests can inspect spans.
// AlwaysSample so every span is recorded (default sampler would drop them).
func setOtelInMemoryProvider() *tracetest.InMemoryExporter {
	exp := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exp), // synchronous for deterministic tests
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	otel.SetTracerProvider(tp)
	return exp
}

// TestFactoryJaegerMode: [trace.jaeger] only -> ModeJaeger, TraceID is a
// non-empty 16-hex (64-bit) jaeger id.
func TestFactoryJaegerMode(t *testing.T) {
	defer resetGlobal()
	writeConfig(t, `
[trace]
    [trace.jaeger]
        serviceName = "svc"
        samplerType = "const"
        samplerParam = 1
        agentAddr = "127.0.0.1:6831"
`)
	closeFn := Init()
	defer closeFn()

	assert.Equal(t, ModeJaeger, GlobalTracer().Mode())

	span, ctx := GlobalTracer().StartServerSpan(context.Background(), "op", SpanServer, nil)
	defer span.End()
	tid := TraceID(ctx)
	assert.NotEmpty(t, tid)
	assert.Len(t, tid, 16, "jaeger TraceID is 64-bit (16 hex)")
	assert.Equal(t, tid, span.TraceID())
}

// TestFactoryOTelMode: [trace.otel] external=true (provider set by test) ->
// ModeOTel, TraceID is a non-empty 32-hex (128-bit) OTel id.
func TestFactoryOTelMode(t *testing.T) {
	defer resetGlobal()
	exp := setOtelInMemoryProvider()
	writeConfig(t, `
[trace]
    [trace.otel]
        serviceName = "svc"
        enable = true
        endpoint = ""
        external = true
        sampleRatio = 1.0
`)
	closeFn := Init()
	defer closeFn()

	assert.Equal(t, ModeOTel, GlobalTracer().Mode())

	span, ctx := GlobalTracer().StartServerSpan(context.Background(), "op", SpanServer, nil)
	tid := TraceID(ctx)
	assert.NotEmpty(t, tid)
	assert.Len(t, tid, 32, "otel TraceID is 128-bit (32 hex)")
	assert.Equal(t, tid, span.TraceID())
	span.End()
	assert.NotEmpty(t, exp.GetSpans(), "otel server span recorded by exporter")
}

// TestFactoryOTelBridgeDBSpan: in otel mode, a span started via the legacy
// StartSpanFromContext (as db/broker layers do) is recorded by the OTel
// exporter and shares the server span's TraceID (correct parent linking
// through the bridge).
func TestFactoryOTelBridgeDBSpan(t *testing.T) {
	defer resetGlobal()
	exp := setOtelInMemoryProvider()
	writeConfig(t, `
[trace]
    [trace.otel]
        serviceName = "svc"
        enable = true
        endpoint = ""
        external = true
        sampleRatio = 1.0
`)
	closeFn := Init()
	defer closeFn()

	// server span (middleware path)
	srvSpan, srvCtx := GlobalTracer().StartServerSpan(context.Background(), "http.server", SpanServer, nil)
	srvTID := TraceID(srvCtx)
	require.NotEmpty(t, srvTID)

	// db/broker path: legacy opentracing facade -> global bridge -> OTel span
	dbSpan, dbCtx := StartSpanFromContext(srvCtx, "mongo.query")
	ext.Component.Set(dbSpan, "mongo")
	dbSpan.Finish()
	srvSpan.End()

	_ = dbCtx
	spans := exp.GetSpans()
	require.Len(t, spans, 2, "server + db span recorded")
	// both spans share the server TraceID
	assert.Equal(t, srvTID, spans[0].SpanContext.TraceID().String())
	assert.Equal(t, srvTID, spans[1].SpanContext.TraceID().String())
}

// TestFactoryDualMode: [trace.jaeger] + [trace.otel] external -> ModeDual,
// TraceID is the OTel 128-bit id.
func TestFactoryDualMode(t *testing.T) {
	defer resetGlobal()
	exp := setOtelInMemoryProvider()
	writeConfig(t, `
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
`)
	closeFn := Init()
	defer closeFn()

	assert.Equal(t, ModeDual, GlobalTracer().Mode())

	span, ctx := GlobalTracer().StartServerSpan(context.Background(), "op", SpanServer, nil)
	tid := TraceID(ctx)
	assert.NotEmpty(t, tid)
	assert.Len(t, tid, 32, "dual TraceID is the OTel 128-bit id")
	span.End()
	assert.NotEmpty(t, exp.GetSpans(), "dual: otel side recorded")
}

// TestFactoryOTelDisabledFallback: enable=true, endpoint="", external=false
// -> OTel inactive (zero-TraceID protection) -> factory falls back to jaeger.
func TestFactoryOTelDisabledFallback(t *testing.T) {
	defer resetGlobal()
	writeConfig(t, `
[trace]
    [trace.otel]
        serviceName = "svc"
        enable = true
        endpoint = ""
        external = false
        sampleRatio = 1.0
`)
	closeFn := Init()
	defer closeFn()

	// OTel is inactive (no endpoint, no external) -> factory should not select
	// otel mode. With no [trace.jaeger] either, it falls back to default
	// jaeger, so TraceID is a 16-hex jaeger id (non-zero, not the all-zero OTel
	// noop id).
	assert.NotEqual(t, ModeOTel, GlobalTracer().Mode())
	span, ctx := GlobalTracer().StartServerSpan(context.Background(), "op", SpanServer, nil)
	defer span.End()
	tid := TraceID(ctx)
	assert.NotEmpty(t, tid)
	assert.NotEqual(t, "00000000000000000000000000000000", tid)
}

// TestCarrierRoundTripHTTP: a client span injects into HTTP headers; a server
// span started on a fresh context extracting from those headers continues the
// same trace (same TraceID).
func TestCarrierRoundTripHTTP(t *testing.T) {
	defer resetGlobal()
	setOtelInMemoryProvider()
	writeConfig(t, `
[trace]
    [trace.otel]
        serviceName = "svc"
        enable = true
        endpoint = ""
        external = true
        sampleRatio = 1.0
`)
	closeFn := Init()
	defer closeFn()

	// client side: start a client span, inject into headers
	hdr := http.Header{}
	cSpan, _ := GlobalTracer().StartClientSpan(context.Background(), "http.call", SpanClient, HTTPCarrier{Header: hdr})
	clientTID := cSpan.TraceID()
	require.NotEmpty(t, clientTID)

	// server side: fresh context, extract parent from the injected headers
	srvSpan, srvCtx := GlobalTracer().StartServerSpan(context.Background(), "http.server", SpanServer, HTTPCarrier{Header: hdr})
	defer srvSpan.End()
	assert.Equal(t, clientTID, TraceID(srvCtx), "server span continues the client's trace")
	assert.Equal(t, clientTID, srvSpan.TraceID())
}

// TestJaegerFacadeStartSpan: in jaeger mode, the legacy StartSpanFromContext
// facade returns a real opentracing span whose ext tags work, and TraceID reads
// it. Guards the db/broker code path.
func TestJaegerFacadeStartSpan(t *testing.T) {
	defer resetGlobal()
	writeConfig(t, `
[trace]
    [trace.jaeger]
        serviceName = "svc"
        samplerType = "const"
        samplerParam = 1
        agentAddr = "127.0.0.1:6831"
`)
	closeFn := Init()
	defer closeFn()

	span, ctx := StartSpanFromContext(context.Background(), "mongo.query")
	ext.Component.Set(span, "mongo")
	span.SetTag("db.system", "mongodb")
	tid := TraceID(ctx)
	assert.NotEmpty(t, tid)
	assert.Len(t, tid, 16)
	span.Finish()
}

// newJaegerTracerForTest builds a standalone jaeger tracer for cross-backend
// tests (the jaeger backend, not wired through the factory).
func newJaegerTracerForTest(t *testing.T, service string) (opentracing.Tracer, func()) {
	t.Helper()
	tracer, closer, err := (&jconfig.Configuration{ServiceName: service}).NewTracer()
	require.NoError(t, err)
	return tracer, func() { closer.Close() }
}

// installOtelCompositePropagator installs the W3C+Jaeger+Baggage composite
// propagator as the global OTel propagator (as the factory does in otel/dual
// mode via otel.InitWithConfig).
func installOtelCompositePropagator() {
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		jaegerprop.Jaeger{},
		propagation.Baggage{},
	))
}

// TestCrossJaegerToOtel: an OLD service (jaeger mode, no composite propagator)
// injects only uber-trace-id; a NEW service (otel composite propagator) extracts
// it. The trace continues: the new service's TraceID is the old 64-bit id
// left-padded to 128-bit.
func TestCrossJaegerToOtel(t *testing.T) {
	defer resetGlobal()

	// --- old service side: jaeger tracer, inject to HTTP header ---
	jtracer, jclose := newJaegerTracerForTest(t, "old")
	defer jclose()
	sp := jtracer.StartSpan("old.outbound")
	hdr := http.Header{}
	require.NoError(t, jtracer.Inject(sp.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(hdr)))
	jaegerTID := sp.Context().(jaeger.SpanContext).TraceID().String()
	sp.Finish()
	require.Contains(t, hdr, "Uber-Trace-Id")
	_, hasTraceparent := hdr["Traceparent"]
	assert.False(t, hasTraceparent, "old jaeger service must NOT emit traceparent")

	// --- new service side: otel composite propagator extracts ---
	installOtelCompositePropagator()
	ctx := otel.GetTextMapPropagator().Extract(context.Background(), propagation.HeaderCarrier(hdr))
	sc := oteltrace.SpanContextFromContext(ctx)
	require.True(t, sc.IsValid(), "new service must extract a parent from uber-trace-id")
	otelTID := sc.TraceID().String()

	// old 64-bit id left-padded with 16 zeros == new 128-bit id -> same trace
	assert.Equal(t, "0000000000000000"+jaegerTID, otelTID,
		"jaeger 64-bit TraceID left-padded == otel 128-bit TraceID")
}

// TestCrossOtelToJaeger: a NEW service (otel composite propagator) injects both
// traceparent + uber-trace-id; an OLD service (jaeger) extracts uber-trace-id.
// The trace continues with the same 128-bit TraceID (jaeger carries 128-bit).
func TestCrossOtelToJaeger(t *testing.T) {
	defer resetGlobal()
	installOtelCompositePropagator()

	// --- new service side: otel tracer, inject to HTTP header ---
	// Use a noop-safe tracer provider so Start produces a real span context.
	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	defer tp.Shutdown(context.Background())
	otel.SetTracerProvider(tp)
	otracer := tp.Tracer("new")
	_, ospan := otracer.Start(context.Background(), "new.outbound")
	otelTID := ospan.SpanContext().TraceID().String()
	require.Len(t, otelTID, 32)
	ctx := oteltrace.ContextWithSpan(context.Background(), ospan)
	hdr := http.Header{}
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(hdr))
	ospan.End()
	require.Contains(t, hdr, "Traceparent")
	require.Contains(t, hdr, "Uber-Trace-Id", "composite propagator also emits uber-trace-id")

	// --- old service side: jaeger tracer extracts uber-trace-id ---
	jtracer, jclose := newJaegerTracerForTest(t, "old")
	defer jclose()
	sc, err := jtracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(hdr))
	require.NoError(t, err, "jaeger must extract parent from uber-trace-id")
	jaegerTID := sc.(jaeger.SpanContext).TraceID().String()

	// 128-bit otel id passes through unchanged (jaeger carries 128-bit) -> identical
	assert.Equal(t, otelTID, jaegerTID,
		"otel 128-bit TraceID passes through to jaeger unchanged")
}

// TestForwardCompatNoTraceSection: a legacy service with NO trace config section
// at all still gets the default jaeger backend and a non-empty TraceID.
func TestForwardCompatNoTraceSection(t *testing.T) {
	defer resetGlobal()
	writeConfig(t, `[etcd]
key = "v"
`)
	closeFn := Init()
	defer closeFn()

	assert.Equal(t, ModeJaeger, GlobalTracer().Mode(), "no trace section -> default jaeger")
	span, ctx := GlobalTracer().StartServerSpan(context.Background(), "op", SpanServer, nil)
	defer span.End()
	tid := TraceID(ctx)
	assert.NotEmpty(t, tid)
	assert.Len(t, tid, 16, "default jaeger 64-bit id")
}

// TestForwardCompatJaegerOnly: a legacy service with only [trace.jaeger] behaves
// exactly as before the refactor (pure jaeger, 64-bit TraceID).
func TestForwardCompatJaegerOnly(t *testing.T) {
	defer resetGlobal()
	writeConfig(t, `
[trace]
    [trace.jaeger]
        serviceName = "oldsrv"
        samplerType = "const"
        samplerParam = 0.001
        agentAddr = "127.0.0.1:6831"
`)
	closeFn := Init()
	defer closeFn()

	assert.Equal(t, ModeJaeger, GlobalTracer().Mode())
	span, ctx := GlobalTracer().StartServerSpan(context.Background(), "op", SpanServer, nil)
	defer span.End()
	tid := TraceID(ctx)
	assert.NotEmpty(t, tid)
	assert.Len(t, tid, 16)
}
