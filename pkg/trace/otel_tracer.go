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
	"go.opentelemetry.io/otel/attribute"
	otelbridge "go.opentelemetry.io/otel/bridge/opentracing"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"github.com/UnderTreeTech/waterdrop/pkg/trace/otel"
)

// newOtelTracer sets up the pure-OTel backend: it initializes the OTel
// provider/propagator (otel.InitWithConfig), then installs an
// opentracing->OTel BridgeTracer as the global opentracing tracer so that
// legacy code using StartSpanFromContext (db/broker layers) produces OTel
// spans via the bridge. Returns an otelTracer + shutdown func.
func newOtelTracer(oconf *otel.Config) (Tracer, func()) {
	otelShutdown := otel.InitWithConfig(oconf)

	bridge := otelbridge.NewBridgeTracer()
	bridge.SetOpenTelemetryTracer(otel.Tracer("waterdrop"))
	bridge.SetTextMapPropagator(otel.Propagator())
	opentracing.SetGlobalTracer(bridge)

	t := &otelTracer{
		tracer:  otel.Tracer("waterdrop"),
		bridge:  bridge,
		prop:    otel.Propagator(),
	}
	return t, func() {
		otelShutdown()
	}
}

// otelTracer is the pure-OTel backend. Server/client spans are started with
// the OTel tracer (placing the OTel span on the OTel context key so TraceID
// works) and also wrapped as a bridge span on the opentracing context key (via
// ContextWithBridgeSpan) so db/broker spans started through the global
// opentracing tracer (the bridge) parent correctly to the server span.
type otelTracer struct {
	tracer trace.Tracer
	bridge *otelbridge.BridgeTracer
	prop   propagation.TextMapPropagator
}

func (t *otelTracer) Mode() Mode { return ModeOTel }

func (t *otelTracer) StartServerSpan(ctx context.Context, name string, kind SpanKind, carrier Carrier) (Span, context.Context) {
	ctx = t.extract(ctx, carrier)
	octx, span := t.tracer.Start(ctx, name, trace.WithSpanKind(toOtelKind(kind)))
	octx = t.bridge.ContextWithBridgeSpan(octx, span)
	return &otelSpanWrap{span: span}, octx
}

func (t *otelTracer) StartClientSpan(ctx context.Context, name string, kind SpanKind, carrier Carrier) (Span, context.Context) {
	octx, span := t.tracer.Start(ctx, name, trace.WithSpanKind(toOtelKind(kind)))
	octx = t.bridge.ContextWithBridgeSpan(octx, span)
	t.inject(octx, carrier)
	return &otelSpanWrap{span: span}, octx
}

func (t *otelTracer) StartSpan(ctx context.Context, name string) (Span, context.Context) {
	octx, span := t.tracer.Start(ctx, name)
	octx = t.bridge.ContextWithBridgeSpan(octx, span)
	return &otelSpanWrap{span: span}, octx
}

func (t *otelTracer) SpanFromContext(ctx context.Context) Span {
	if sp := trace.SpanFromContext(ctx); sp != nil && sp.SpanContext().IsValid() {
		return &otelSpanWrap{span: sp}
	}
	return noopSpan{}
}

func (t *otelTracer) TraceID(ctx context.Context) string {
	sc := trace.SpanContextFromContext(ctx)
	if !sc.HasTraceID() {
		return ""
	}
	return sc.TraceID().String()
}

func (t *otelTracer) Shutdown() {}

// extract extracts a parent SpanContext from carrier into ctx via the composite
// propagator.
func (t *otelTracer) extract(ctx context.Context, carrier Carrier) context.Context {
	if carrier == nil {
		return ctx
	}
	return t.prop.Extract(ctx, toOtelCarrier(carrier))
}

// inject injects the active span on ctx into carrier via the composite propagator.
func (t *otelTracer) inject(ctx context.Context, carrier Carrier) {
	if carrier == nil {
		return
	}
	t.prop.Inject(ctx, toOtelCarrier(carrier))
}

// otelSpanWrap adapts an OTel trace.Span to the Span interface.
type otelSpanWrap struct {
	span trace.Span
}

func (s *otelSpanWrap) End() { s.span.End() }

func (s *otelSpanWrap) SetStatus(ok bool, msg string) {
	if !ok {
		s.span.SetStatus(codes.Error, msg)
	}
}

func (s *otelSpanWrap) SetStringAttr(key, val string) {
	s.span.SetAttributes(keyVal(key, val))
}

func (s *otelSpanWrap) SetIntAttr(key string, val int64) {
	s.span.SetAttributes(keyInt(key, val))
}

func (s *otelSpanWrap) TraceID() string {
	sc := s.span.SpanContext()
	if !sc.HasTraceID() {
		return ""
	}
	return sc.TraceID().String()
}

// toOtelKind maps a SpanKind to an OTel trace.SpanKind.
func toOtelKind(kind SpanKind) trace.SpanKind {
	switch kind {
	case SpanServer:
		return trace.SpanKindServer
	case SpanClient:
		return trace.SpanKindClient
	default:
		return trace.SpanKindInternal
	}
}

// toOtelCarrier converts the Carrier abstraction to an OTel TextMapCarrier.
func toOtelCarrier(carrier Carrier) propagation.TextMapCarrier {
	switch c := carrier.(type) {
	case HTTPCarrier:
		return propagation.HeaderCarrier(c.Header)
	case GRPCCarrier:
		return otel.MDCarrier{MD: c.MD}
	}
	return nil
}

// keyVal returns an OTel string attribute. Shared with dualTracer.
func keyVal(key, val string) attribute.KeyValue {
	return attribute.String(key, val)
}

// keyInt returns an OTel int64 attribute. Shared with dualTracer.
func keyInt(key string, val int64) attribute.KeyValue {
	return attribute.Int64(key, val)
}
