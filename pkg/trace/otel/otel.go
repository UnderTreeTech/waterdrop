/*
 *
 * Copyright 2026 waterdrop authors.
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

// Package otel provides the OpenTelemetry backend for waterdrop's trace
// factory (pkg/trace). The global propagator is a composite of W3C
// TraceContext + Jaeger (uber-trace-id) + Baggage so that, in dual mode, an
// OTel span is injected as BOTH `traceparent` and `uber-trace-id` and the
// jaeger stack can extract the latter, sharing one TraceID without a bridge.
//
// Ownership / initialization:
// InitWithConfig is called by trace.Init() (the factory) when the [trace.otel]
// section is present and enabled. Services opt in purely via config.
//   - Self-built provider: [trace.otel] with an endpoint -> InitWithConfig
//     builds an OTLP/HTTP TracerProvider, sets it global, and installs the
//     composite propagator.
//   - External provider: [trace.otel] with enable=true, endpoint="", and
//     external=true -> InitWithConfig installs only the composite propagator
//     and flips Enabled(); a global TracerProvider is set by an external
//     component (e.g. an ADK langfuse.Setup).
//   - Disabled: [trace.otel] with enable=true, endpoint="", external=false ->
//     InitWithConfig warns and leaves OTel disabled to avoid emitting spans
//     with an all-zero TraceID from the noop provider.
//   - enable=false or no [trace.otel] section: no-op, the factory selects the
//     pure-jaeger mode (forward-compatible fallback).
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
	// Enable=true, no provider is built; see External for how the provider is
	// then resolved.
	Endpoint string
	// External, only meaningful with Enable=true and Endpoint="", declares that
	// the global TracerProvider is owned and set by an external component (e.g.
	// an ADK langfuse.Setup). InitWithConfig then installs only the composite
	// propagator and flips Enabled(), trusting the external provider is set
	// later. If Endpoint="" and External=false with Enable=true, InitWithConfig
	// logs a warning and leaves OTel disabled (no propagator, Enabled()=false)
	// to avoid emitting spans with an all-zero TraceID from the noop provider.
	External bool
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

// Active reports whether the config would actually enable OTel: enable=true
// AND (an endpoint to build a provider OR external=true declaring a provider is
// set elsewhere). The trace factory uses this to select the otel/dual backend;
// enable=true with neither endpoint nor external is treated as inactive (OTel
// would emit zero-TraceID spans from the noop provider).
func (c *Config) Active() bool {
	return c.Enable && (c.Endpoint != "" || c.External)
}

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
// built). Called by trace.Init when [trace.otel] is present and enabled.
//
// enable=true + endpoint!="": full provider + propagator (self-built).
// enable=true + endpoint=""  + external=true: propagator only (global provider
// owned and set elsewhere, e.g. ADK langfuse.Setup).
// enable=true + endpoint=""  + external=false: warn and disable (avoid zero
// TraceID spans from the noop provider).
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
		if !cfg.External {
			// Enabled but neither an endpoint to build a provider nor a declared
			// external provider: the global TracerProvider would be the noop
			// one, so every OTel span would carry an all-zero TraceID. Disable
			// OTel instead of emitting misleading zero-id spans. Roll back the
			// propagator/enabled flag set above so middlewares behave as the
			// non-OTel path.
			otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator())
			enabled.Store(false)
			log.Printf("otel: enabled but no endpoint and external=false; OTel disabled to avoid zero TraceID (set external=true if a global provider is set elsewhere)")
			return func() {}
		}
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
