/*
 *
 * Copyright 2020 waterdrop authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/LICENSE.org/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package otel provides OpenTelemetry tracing that runs alongside waterdrop's
// existing opentracing/jaeger stack, sharing the same TraceID without a bridge.
//
// Design (trace unification, route W):
// waterdrop keeps its opentracing `Trace` middleware / gRPC interceptors
// (jaeger, uber-trace-id). This package adds a parallel OTel stack whose
// global propagator is a composite of W3C TraceContext + Jaeger
// (uber-trace-id) + Baggage. Because both dialects are spoken, an OTel
// span is injected as BOTH `traceparent` and `uber-trace-id`; the existing
// jaeger middleware extracts the latter and therefore jaeger and OTel
// share the same TraceID without an opentracing->OTel bridge and without
// replacing the global opentracing tracer (so pkg/trace.TraceID keeps
// working for downstream services that receive uber-trace-id carrying the
// OTel id).
//
// Ownership / initialization:
// OTel is initialized inside jaeger.Init() when the [trace.otel] section is
// present and enabled, so services opt in purely via config (no code change).
// Services without that section (or with enable=false) stay pure-jaeger (==
// current behavior), preserving forward compatibility.
//   - Services without their own OTel provider (e.g. anchor): set
//     [trace.otel] with an endpoint -> InitWithConfig builds a
//     TracerProvider exporting to langfuse and sets the composite propagator.
//   - Services whose provider is owned elsewhere (e.g. maestro's ADK
//     langfuse.Setup sets the global TracerProvider): set
//     [trace.otel] with enable=true but empty endpoint ->
//     InitWithConfig only installs the composite propagator and flips
//     Enabled(); the later langfuse.Setup sets the provider.
package otel

import (
	"context"
	"log"
	"sync/atomic"
	"time"

	jaegerprop "go.opentelemetry.io/contrib/propagators/jaeger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// Config is the [trace.otel] configuration.
type Config struct {
	// ServiceName reported as resource attribute `service.name`. Falls back to
	// the jaeger serviceName when empty (caller's responsibility).
	ServiceName string
	// Enable gates the OTel stack. When false, InitWithConfig is a no-op and
	// Enabled() stays false (middlewares behave as pure jaeger).
	Enable bool
	// Endpoint is the full OTLP/HTTP trace endpoint, e.g.
	// "https://cloud.langfuse.com/api/public/otel/v1/traces". When empty with
	// Enable=true, no provider is built (caller's global provider is used,
	// e.g. maestro's ADK langfuse.Setup).
	Endpoint string
	// Headers attached to OTLP export requests, e.g.
	// {"Authorization": "Basic <base64(public:secret)>"}.
	Headers map[string]string
	// Insecure skips TLS (for plain-http endpoints).
	Insecure bool
	// SampleRatio in [0,1]. <=0 or >=1 => always sample.
	SampleRatio float64
}

// enabled is set true once the composite propagator is installed.
var enabled atomic.Bool

// Enabled reports whether the OTel stack has been initialized. The merged
// Trace middleware / gRPC interceptors use this to decide whether to also
// start an OTel span (otherwise they behave as pure jaeger).
func Enabled() bool { return enabled.Load() }

// compositePropagator returns the W3C+Jaeger+Baggage composite propagator
// that lets OTel and waterdrop's jaeger share a TraceID without a bridge.
func compositePropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, // W3C traceparent (RFC standard)
		jaegerprop.Jaeger{},        // uber-trace-id (interop with waterdrop jaeger)
		propagation.Baggage{},
	)
}

// InitWithConfig installs the composite propagator and, when cfg.Endpoint is
// non-empty, builds an OTel TracerProvider exporting to the OTLP/HTTP
// endpoint and sets it global. Returns a shutdown func (noop when nothing was
// built). Called by jaeger.Init when [trace.otel] is present and enabled.
//
// enable=true + endpoint="": propagator only (provider owned elsewhere, e.g.
// maestro's ADK langfuse.Setup which sets the global provider later).
// enable=true + endpoint!="": full provider + propagator (e.g. anchor).
// enable=false: no-op, stays pure jaeger.
func InitWithConfig(cfg *Config) func() {
	if cfg == nil || !cfg.Enable {
		return func() {}
	}

	// Always install the composite propagator so cross-process propagation
	// works regardless of who owns the provider.
	otel.SetTextMapPropagator(compositePropagator())
	enabled.Store(true)

	if cfg.Endpoint == "" {
		// Provider owned elsewhere; only the propagator was needed.
		return func() {}
	}

	ctx := context.Background()
	opts := []otlptracehttp.Option{otlptracehttp.WithEndpointURL(cfg.Endpoint)}
	if len(cfg.Headers) > 0 {
		opts = append(opts, otlptracehttp.WithHeaders(cfg.Headers))
	}
	if cfg.Insecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	exporter, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		log.Printf("otel: create otlp exporter fail, err msg %s", err.Error())
		return func() {}
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(semconv.ServiceNameKey.String(cfg.ServiceName)),
	)
	if err != nil {
		log.Printf("otel: create resource fail, err msg %s", err.Error())
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler(cfg.SampleRatio)),
	)
	otel.SetTracerProvider(tp)

	return func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tp.Shutdown(shutdownCtx); err != nil {
			log.Printf("otel: tracer provider shutdown fail, err msg %s", err.Error())
		}
	}
}

func sampler(ratio float64) sdktrace.Sampler {
	if ratio <= 0 || ratio >= 1 {
		return sdktrace.AlwaysSample()
	}
	return sdktrace.TraceIDRatioBased(ratio)
}

// OtelTraceID returns the OTel TraceID from the context's active span, or ""
// when there is none.
func OtelTraceID(ctx context.Context) string {
	sc := oteltrace.SpanContextFromContext(ctx)
	if !sc.HasTraceID() {
		return ""
	}
	return sc.TraceID().String()
}

// SpanContextFromContext exposes the OTel SpanContext for callers that need to
// re-parent onto an isolated context (e.g. anchor's proxyMaestroChat which
// derives its ctx from context.Background() to avoid inheriting gin cancellation).
func SpanContextFromContext(ctx context.Context) oteltrace.SpanContext {
	return oteltrace.SpanContextFromContext(ctx)
}

// ContextWithSpanContext returns a copy of ctx carrying the given OTel
// SpanContext as parent, without inheriting any cancellation from the original
// span's context.
func ContextWithSpanContext(ctx context.Context, sc oteltrace.SpanContext) context.Context {
	return oteltrace.ContextWithSpanContext(ctx, sc)
}

// Propagator returns the globally configured propagator (composite when
// initialized, OTel's default noop otherwise).
func Propagator() propagation.TextMapPropagator {
	return otel.GetTextMapPropagator()
}

// Tracer returns a named tracer from the global provider.
func Tracer(name string) oteltrace.Tracer {
	return otel.Tracer(name)
}
