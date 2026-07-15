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

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otelbridge "go.opentelemetry.io/otel/bridge/opentracing"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/UnderTreeTech/waterdrop/pkg/trace/otel"
)

// newDualTracer builds the dual backend: a private jaeger tracer (for jaeger
// server/client spans, exported to the jaeger backend) AND the OTel stack +
// opentracing->OTel bridge as the global opentracing tracer (so db/broker
// spans go to the OTel backend and parent to the OTel server span). jaeger and
// OTel share one TraceID via the composite propagator. Returns a dualTracer +
// shutdown func.
func newDualTracer(jconf *jaegerFlatConfig, oconf *otel.Config) (Tracer, func()) {
	jt, jaegerClose := newJaegerTracerOnly(jconf)

	otelShutdown := otel.InitWithConfig(oconf)
	bridge := otelbridge.NewBridgeTracer()
	bridge.SetOpenTelemetryTracer(otel.Tracer("waterdrop"))
	bridge.SetTextMapPropagator(otel.Propagator())
	opentracing.SetGlobalTracer(bridge)

	t := &dualTracer{
		jaeger: jt,
		otel: &otelTracer{
			tracer: otel.Tracer("waterdrop"),
			bridge: bridge,
			prop:   otel.Propagator(),
		},
	}
	return t, func() {
		otelShutdown()
		jaegerClose()
	}
}

// dualTracer runs jaeger and OTel side by side. The OTel side is the primary
// one for db/broker parent linking (via the global bridge), so StartServerSpan
// places the OTel span on the OTel context key and a bridge span on the
// opentracing context key. The jaeger span is started with the private jaeger
// tracer and tracked in the returned dualSpan (it has no db/broker children -
// those go to OTel; this is the documented dual-mode tradeoff).
type dualTracer struct {
	jaeger *jaegerTracer
	otel   *otelTracer
}

func (t *dualTracer) Mode() Mode { return ModeDual }

func (t *dualTracer) StartServerSpan(ctx context.Context, name string, kind SpanKind, carrier Carrier) (Span, context.Context) {
	// OTel side: extract parent, start OTel server span, wrap as bridge span so
	// db/broker (global=bridge) parent to it.
	octx := t.otel.extract(ctx, carrier)
	octx, ospan := t.otel.tracer.Start(octx, name, oteltrace.WithSpanKind(toOtelKind(kind)))
	octx = t.otel.bridge.ContextWithBridgeSpan(octx, ospan)

	// Jaeger side: extract parent (uber-trace-id via the private jaeger tracer),
	// start a jaeger server span. It is NOT placed on the opentracing context
	// key (the bridge span is, for db/broker), so it has no children in the
	// jaeger backend; tracked in dualSpan for Finish.
	var jspan opentracing.Span
	opts := []opentracing.StartSpanOption{opentracing.Tag{Key: string(ext.Component), Value: componentForCarrier(carrier)}}
	if carrier != nil {
		if parent := extractWith(t.jaeger.tracer, carrier); parent != nil {
			opts = append(opts, parent)
		}
	}
	jspan = t.jaeger.tracer.StartSpan(name, opts...)
	ext.SpanKind.Set(jspan, spanKindOpentracing(kind))

	return &dualSpan{otel: ospan, jaeger: jspan}, octx
}

func (t *dualTracer) StartClientSpan(ctx context.Context, name string, kind SpanKind, carrier Carrier) (Span, context.Context) {
	octx, ospan := t.otel.tracer.Start(ctx, name, oteltrace.WithSpanKind(toOtelKind(kind)))
	octx = t.otel.bridge.ContextWithBridgeSpan(octx, ospan)
	t.otel.inject(octx, carrier)

	jspan := t.jaeger.tracer.StartSpan(name, opentracing.Tag{Key: string(ext.Component), Value: componentForCarrier(carrier)})
	ext.SpanKind.Set(jspan, spanKindOpentracing(kind))
	if carrier != nil {
		injectWith(t.jaeger.tracer, jspan, carrier)
	}
	return &dualSpan{otel: ospan, jaeger: jspan}, octx
}

func (t *dualTracer) StartSpan(ctx context.Context, name string) (Span, context.Context) {
	octx, ospan := t.otel.tracer.Start(ctx, name)
	octx = t.otel.bridge.ContextWithBridgeSpan(octx, ospan)
	jspan := t.jaeger.tracer.StartSpan(name)
	return &dualSpan{otel: ospan, jaeger: jspan}, octx
}

func (t *dualTracer) SpanFromContext(ctx context.Context) Span {
	if sp := oteltrace.SpanFromContext(ctx); sp != nil && sp.SpanContext().IsValid() {
		return &otelSpanWrap{span: sp}
	}
	return noopSpan{}
}

func (t *dualTracer) TraceID(ctx context.Context) string {
	sc := oteltrace.SpanContextFromContext(ctx)
	if !sc.HasTraceID() {
		return ""
	}
	return sc.TraceID().String()
}

func (t *dualTracer) Shutdown() {}

// dualSpan wraps a jaeger span and an OTel span. End/SetStatus/SetAttr are
// forwarded to both so both backends receive the same span data. TraceID
// returns the OTel id (the canonical cross-backend id).
type dualSpan struct {
	otel   oteltrace.Span
	jaeger opentracing.Span
}

func (s *dualSpan) End() {
	if s.otel != nil {
		s.otel.End()
	}
	if s.jaeger != nil {
		s.jaeger.Finish()
	}
}

func (s *dualSpan) SetStatus(ok bool, msg string) {
	if !ok {
		if s.otel != nil {
			s.otel.SetStatus(codes.Error, msg)
		}
		if s.jaeger != nil {
			ext.Error.Set(s.jaeger, true)
			if msg != "" {
				s.jaeger.SetTag("error.message", msg)
			}
		}
	}
}

func (s *dualSpan) SetStringAttr(key, val string) {
	if s.otel != nil {
		s.otel.SetAttributes(keyVal(key, val))
	}
	if s.jaeger != nil {
		s.jaeger.SetTag(key, val)
	}
}

func (s *dualSpan) SetIntAttr(key string, val int64) {
	if s.otel != nil {
		s.otel.SetAttributes(keyInt(key, val))
	}
	if s.jaeger != nil {
		s.jaeger.SetTag(key, val)
	}
}

func (s *dualSpan) TraceID() string {
	if s.otel != nil {
		sc := s.otel.SpanContext()
		if sc.HasTraceID() {
			return sc.TraceID().String()
		}
	}
	return ""
}
