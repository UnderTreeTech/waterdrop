/*
 *
 * Copyright 2020 waterdrop authors.
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
	"log"
	"net/http"
	"os"
	"time"

	"github.com/UnderTreeTech/waterdrop/pkg/conf"
	"github.com/UnderTreeTech/waterdrop/pkg/trace/otel"

	"google.golang.org/grpc/metadata"
)

// Mode is the active trace backend selected by the factory.
type Mode string

const (
	// ModeJaeger: pure opentracing/jaeger (uber-trace-id). Default/fallback.
	ModeJaeger Mode = "jaeger"
	// ModeOTel: pure OpenTelemetry (traceparent + uber-trace-id via the
	// composite propagator). db/broker spans reach the OTel backend via the
	// opentracing->OTel bridge set as the global opentracing tracer.
	ModeOTel Mode = "otel"
	// ModeDual: jaeger + OTel run together sharing one TraceID (composite
	// propagator). db/broker spans go to the OTel side via the bridge.
	ModeDual Mode = "dual"
)

// SpanKind classifies a span.
type SpanKind int

const (
	SpanServer SpanKind = iota
	SpanClient
	SpanInternal
)

// Span is the cross-backend span abstraction used by the Trace middleware /
// gRPC interceptors. db/broker code does NOT use this interface - it keeps
// using the legacy opentracing Span via StartSpanFromContext (which, in
// otel/dual mode, is a bridge span backed by a real OTel span).
type Span interface {
	// End finishes the span.
	End()
	// SetStatus records the span's status; ok=false marks it as an error with
	// the given message.
	SetStatus(ok bool, msg string)
	// SetStringAttr sets a string attribute/tag on the span.
	SetStringAttr(key, val string)
	// SetIntAttr sets an int64 attribute/tag on the span.
	SetIntAttr(key string, val int64)
	// TraceID returns the span's TraceID as a hex string (jaeger 64-bit or
	// OTel 128-bit depending on the active backend).
	TraceID() string
}

// Carrier carries trace context across process boundaries. It is either an
// HTTP header set or a gRPC metadata.MD; the concrete Tracer implementations
// type-switch on it to use the appropriate native carrier, preserving the
// exact wire format of each backend.
type Carrier interface {
	isCarrier()
}

// HTTPCarrier wraps an http.Header for HTTP transport propagation.
type HTTPCarrier struct{ Header http.Header }

func (HTTPCarrier) isCarrier() {}

// GRPCCarrier wraps a gRPC metadata.MD for gRPC transport propagation.
type GRPCCarrier struct{ MD metadata.MD }

func (GRPCCarrier) isCarrier() {}

// Tracer is the trace backend selected by Init() based on config. The Trace
// middleware and gRPC interceptors call this interface; they hold no
// jaeger/OTel-specific logic.
type Tracer interface {
	// Mode reports the active backend.
	Mode() Mode
	// StartServerSpan extracts the parent from carrier (when non-nil) and
	// starts a server span of the given kind. The returned context carries the
	// span so downstream code (and db/broker spans) parent to it.
	StartServerSpan(ctx context.Context, name string, kind SpanKind, carrier Carrier) (Span, context.Context)
	// StartClientSpan starts a client span of the given kind and injects its
	// context into carrier (when non-nil) so the downstream server continues
	// the trace.
	StartClientSpan(ctx context.Context, name string, kind SpanKind, carrier Carrier) (Span, context.Context)
	// StartSpan starts an internal/root span (e.g. cron entry) with no carrier.
	StartSpan(ctx context.Context, name string) (Span, context.Context)
	// SpanFromContext returns the active span from ctx, or a noop span.
	SpanFromContext(ctx context.Context) Span
	// TraceID returns the TraceID of the active span on ctx, or "".
	TraceID(ctx context.Context) string
	// Shutdown releases backend resources.
	Shutdown()
}

// globalTracer is the backend selected by Init. Defaults to noopTracer so
// that calling TraceID/etc. before Init is safe.
var globalTracer Tracer = noopTracer{}

// GlobalTracer returns the active trace backend. Before Init() it is a noop.
func GlobalTracer() Tracer { return globalTracer }

// noopTracer is the zero-value backend: it produces noop spans and empty
// TraceIDs, and does no propagation.
type noopTracer struct{}

func (noopTracer) Mode() Mode { return ModeJaeger }
func (noopTracer) StartServerSpan(ctx context.Context, _ string, _ SpanKind, _ Carrier) (Span, context.Context) {
	return noopSpan{}, ctx
}
func (noopTracer) StartClientSpan(ctx context.Context, _ string, _ SpanKind, _ Carrier) (Span, context.Context) {
	return noopSpan{}, ctx
}
func (noopTracer) StartSpan(ctx context.Context, _ string) (Span, context.Context) {
	return noopSpan{}, ctx
}
func (noopTracer) SpanFromContext(context.Context) Span { return noopSpan{} }
func (noopTracer) TraceID(context.Context) string       { return "" }
func (noopTracer) Shutdown()                            {}

// noopSpan implements Span as a set of no-ops.
type noopSpan struct{}

func (noopSpan) End()                       {}
func (noopSpan) SetStatus(bool, string)     {}
func (noopSpan) SetStringAttr(string, string) {}
func (noopSpan) SetIntAttr(string, int64)   {}
func (noopSpan) TraceID() string            { return "" }

// Init is the trace factory entry point. It reads the [trace.jaeger] and
// [trace.otel] config sections and selects a backend:
//
//   - [trace.jaeger] + [trace.otel] enabled -> dual (jaeger + OTel, same TraceID)
//   - [trace.otel] enabled only            -> pure OTel
//   - [trace.jaeger] only                  -> pure jaeger
//   - neither                              -> default jaeger (forward-compatible)
//
// It sets the global opentracing tracer (jaeger tracer in jaeger mode, the
// opentracing->OTel bridge in otel/dual mode) so legacy code using
// StartSpanFromContext keeps working, and installs the globalTracer used by
// the Trace middleware / gRPC interceptors. Returns a combined shutdown func.
// Services call it via jaeger.Init() (which delegates here).
func Init() func() {
	hasOtel, oconf := readOtelConfig()
	hasJaeger, jconf := readJaegerConfig()

	var tracer Tracer
	var shutdown func()

	switch {
	case hasJaeger && hasOtel:
		t, closeFn := newDualTracer(jconf, oconf)
		tracer, shutdown = t, closeFn
	case hasOtel:
		t, closeFn := newOtelTracer(oconf)
		tracer, shutdown = t, closeFn
	case hasJaeger:
		t, closeFn := newJaegerTracer(jconf)
		tracer, shutdown = t, closeFn
	default:
		// Forward-compatible fallback: default jaeger (serviceName = hostname).
		t, closeFn := newJaegerTracer(defaultJaegerFlatConfig())
		tracer, shutdown = t, closeFn
	}

	globalTracer = tracer
	return func() {
		if shutdown != nil {
			shutdown()
		}
	}
}

// readOtelConfig reads [trace.otel]. conf.Unmarshal is lenient (returns nil
// with a zero-valued target when the key is missing). Active() is the real
// enable test: enable=true AND (endpoint != "" OR external=true). enable=true
// with neither is treated as inactive (zero-TraceID protection, see otel.go).
func readOtelConfig() (bool, *otel.Config) {
	oconf := &otel.Config{}
	_ = conf.Unmarshal("trace.otel", oconf)
	return oconf.Active(), oconf
}

// readJaegerConfig reads [trace.jaeger]. jaeger has no enable flag, so a
// non-empty ServiceName is the presence test (a configured section always
// carries serviceName; a missing section leaves it empty).
func readJaegerConfig() (bool, *jaegerFlatConfig) {
	jconf := &jaegerFlatConfig{}
	_ = conf.Unmarshal("trace.jaeger", jconf)
	return jconf.ServiceName != "", jconf
}

// jaegerFlatConfig is the [trace.jaeger] config shape (mirrors the former
// pkg/trace/jaeger.Config, kept local to the trace package so the factory does
// not import the jaeger package and create a cycle). newJaegerTracer consumes
// it.
type jaegerFlatConfig struct {
	ServiceName                 string
	EnableRPCMetrics            bool
	SamplerType                 string
	SamplerParam                float64
	AgentAddr                   string
	ReporterLogSpans            bool
	ReporterBufferFlushInterval time.Duration
	TraceBaggageHeaderPrefix    string
	TraceContextHeaderName      string
	MaxTagValueLength           int
}

// defaultJaegerFlatConfig mirrors the former pkg/trace/jaeger.defaultJaegerConfig
// for the forward-compatible fallback (no [trace.jaeger] section).
func defaultJaegerFlatConfig() *jaegerFlatConfig {
	return &jaegerFlatConfig{
		ServiceName:  defaultJaegerServiceName(),
		SamplerType:  "const",
		SamplerParam: 0.001,
		AgentAddr:    defaultJaegerAgentAddr(),
	}
}

// defaultJaegerServiceName returns the OS hostname (jaeger requires a
// non-empty service name).
func defaultJaegerServiceName() string {
	name, _ := os.Hostname()
	return name
}

// defaultJaegerAgentAddr returns the jaeger agent UDP address, overridable via
// the JAEGER_AGENT_ADDR env var.
func defaultJaegerAgentAddr() string {
	addr := "127.0.0.1:6831"
	if env := os.Getenv("JAEGER_AGENT_ADDR"); env != "" {
		addr = env
	}
	return addr
}

// failf logs a fatal-ish init error. The factory does not panic to avoid
// taking down services on a misconfigured trace section; it logs and falls
// back to noop.
func warnInit(format string, args ...interface{}) {
	log.Printf("trace init: "+format, args...)
}
