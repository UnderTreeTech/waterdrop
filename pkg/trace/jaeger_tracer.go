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
	"fmt"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	jaeger "github.com/uber/jaeger-client-go"
	jconfig "github.com/uber/jaeger-client-go/config"

	"google.golang.org/grpc/metadata"
)

// newJaegerTracer builds an opentracing/jaeger tracer from the [trace.jaeger]
// config, sets it as the global opentracing tracer (so legacy
// StartSpanFromContext code, including db/broker layers, produces jaeger
// spans), and returns a jaegerTracer + shutdown func.
func newJaegerTracer(jconf *jaegerFlatConfig) (Tracer, func()) {
	sampler := &jconfig.SamplerConfig{
		Type:  jconf.SamplerType,
		Param: jconf.SamplerParam,
	}
	reporter := &jconfig.ReporterConfig{
		LocalAgentHostPort:  jconf.AgentAddr,
		LogSpans:            jconf.ReporterLogSpans,
		BufferFlushInterval: jconf.ReporterBufferFlushInterval,
	}
	cfg := &jconfig.Configuration{
		ServiceName: jconf.ServiceName,
		Sampler:     sampler,
		Reporter:    reporter,
		RPCMetrics:  jconf.EnableRPCMetrics,
	}
	opts := []jconfig.Option{jconfig.MaxTagValueLength(jconf.MaxTagValueLength)}
	tracer, closer, err := cfg.NewTracer(opts...)
	if err != nil {
		// jaeger requires a non-empty service name; the factory gates on that,
		// but panic on any other misconfiguration to surface it loudly rather
		// than silently running noop.
		panic(fmt.Sprintf("new jaeger tracer fail, err msg %s", err.Error()))
	}
	opentracing.SetGlobalTracer(tracer)
	return &jaegerTracer{tracer: tracer}, func() { _ = closer.Close() }
}

// newJaegerTracerOnly builds a jaeger tracer without setting it global. Used
// by dualTracer, which needs a private jaeger tracer while the global
// opentracing tracer is the OTel bridge (for db/broker spans).
func newJaegerTracerOnly(jconf *jaegerFlatConfig) (*jaegerTracer, func()) {
	sampler := &jconfig.SamplerConfig{Type: jconf.SamplerType, Param: jconf.SamplerParam}
	reporter := &jconfig.ReporterConfig{
		LocalAgentHostPort:  jconf.AgentAddr,
		LogSpans:            jconf.ReporterLogSpans,
		BufferFlushInterval: jconf.ReporterBufferFlushInterval,
	}
	if reporter.BufferFlushInterval == 0 {
		reporter.BufferFlushInterval = 1 * time.Second
	}
	cfg := &jconfig.Configuration{
		ServiceName: jconf.ServiceName, Sampler: sampler, Reporter: reporter, RPCMetrics: jconf.EnableRPCMetrics,
	}
	tracer, closer, err := cfg.NewTracer(jconfig.MaxTagValueLength(jconf.MaxTagValueLength))
	if err != nil {
		panic(fmt.Sprintf("new jaeger tracer fail, err msg %s", err.Error()))
	}
	return &jaegerTracer{tracer: tracer}, func() { _ = closer.Close() }
}

// jaegerTracer is the pure-jaeger backend. It holds a private opentracing
// tracer (the jaeger tracer) and uses it directly for span creation and
// propagation. In jaeger mode the same tracer is also set global so db/broker
// code using StartSpanFromContext produces jaeger spans that parent correctly;
// in dual mode the global is the OTel bridge and this tracer is used privately
// for the jaeger side of dual spans.
type jaegerTracer struct {
	tracer opentracing.Tracer
}

func (t *jaegerTracer) Mode() Mode { return ModeJaeger }

func (t *jaegerTracer) StartServerSpan(ctx context.Context, name string, kind SpanKind, carrier Carrier) (Span, context.Context) {
	opts := []opentracing.StartSpanOption{opentracing.Tag{Key: string(ext.Component), Value: componentForCarrier(carrier)}}
	if carrier != nil {
		if parent := extractWith(t.tracer, carrier); parent != nil {
			opts = append(opts, parent)
		}
	}
	span := t.tracer.StartSpan(name, opts...)
	ext.SpanKind.Set(span, spanKindOpentracing(kind))
	ctx = opentracing.ContextWithSpan(ctx, span)
	return &jaegerSpan{span: span}, ctx
}

func (t *jaegerTracer) StartClientSpan(ctx context.Context, name string, kind SpanKind, carrier Carrier) (Span, context.Context) {
	span := t.tracer.StartSpan(name, opentracing.Tag{Key: string(ext.Component), Value: componentForCarrier(carrier)})
	ext.SpanKind.Set(span, spanKindOpentracing(kind))
	ctx = opentracing.ContextWithSpan(ctx, span)
	if carrier != nil {
		injectWith(t.tracer, span, carrier)
	}
	return &jaegerSpan{span: span}, ctx
}

func (t *jaegerTracer) StartSpan(ctx context.Context, name string) (Span, context.Context) {
	span := t.tracer.StartSpan(name)
	ctx = opentracing.ContextWithSpan(ctx, span)
	return &jaegerSpan{span: span}, ctx
}

func (t *jaegerTracer) SpanFromContext(ctx context.Context) Span {
	if sp := opentracing.SpanFromContext(ctx); sp != nil {
		return &jaegerSpan{span: sp}
	}
	return noopSpan{}
}

func (t *jaegerTracer) TraceID(ctx context.Context) string {
	sp := opentracing.SpanFromContext(ctx)
	if sp == nil {
		return ""
	}
	if jsc, ok := sp.Context().(jaeger.SpanContext); ok {
		return jsc.TraceID().String()
	}
	return ""
}

func (t *jaegerTracer) Shutdown() {}

// jaegerSpan adapts an opentracing.Span to the Span interface.
type jaegerSpan struct {
	span opentracing.Span
}

func (s *jaegerSpan) End() { s.span.Finish() }

func (s *jaegerSpan) SetStatus(ok bool, msg string) {
	if !ok {
		ext.Error.Set(s.span, true)
		if msg != "" {
			s.span.SetTag("error.message", msg)
		}
	}
}

func (s *jaegerSpan) SetStringAttr(key, val string) {
	s.span.SetTag(key, val)
}

func (s *jaegerSpan) SetIntAttr(key string, val int64) {
	s.span.SetTag(key, val)
}

func (s *jaegerSpan) TraceID() string {
	if jsc, ok := s.span.Context().(jaeger.SpanContext); ok {
		return jsc.TraceID().String()
	}
	return ""
}

// extractWith extracts a parent SpanContext from carrier via the given tracer,
// returning a ChildOf option (or nil if no parent). Mirrors the legacy
// HeaderExtractor / FromIncomingContext.
func extractWith(tracer opentracing.Tracer, carrier Carrier) opentracing.StartSpanOption {
	switch c := carrier.(type) {
	case HTTPCarrier:
		sc, err := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(c.Header))
		if err != nil {
			return NullStartSpanOption{}
		}
		return opentracing.ChildOf(sc)
	case GRPCCarrier:
		sc, err := tracer.Extract(opentracing.HTTPHeaders, CarrierMD{md: c.MD})
		if err != nil {
			return NullStartSpanOption{}
		}
		return opentracing.ChildOf(sc)
	}
	return NullStartSpanOption{}
}

// injectWith injects span's context into carrier via the given tracer. Mirrors
// the legacy HeaderInjector / MetadataInjector.
func injectWith(tracer opentracing.Tracer, span opentracing.Span, carrier Carrier) {
	switch c := carrier.(type) {
	case HTTPCarrier:
		if err := tracer.Inject(span.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(c.Header)); err != nil {
			span.SetTag("inject.error", err.Error())
		}
	case GRPCCarrier:
		if err := tracer.Inject(span.Context(), opentracing.HTTPHeaders, CarrierMD{md: c.MD}); err != nil {
			span.SetTag("inject.error", err.Error())
		}
	}
}

// componentForCarrier returns the opentracing component tag value for the
// transport identified by the carrier (nil carrier -> internal).
func componentForCarrier(carrier Carrier) string {
	switch carrier.(type) {
	case HTTPCarrier:
		return "http"
	case GRPCCarrier:
		return "grpc"
	}
	return "internal"
}

// spanKindOpentracing maps a SpanKind to an opentracing ext span kind.
func spanKindOpentracing(kind SpanKind) ext.SpanKindEnum {
	switch kind {
	case SpanServer:
		return ext.SpanKindRPCServerEnum
	case SpanClient:
		return ext.SpanKindRPCClientEnum
	default:
		return ext.SpanKindEnum("internal")
	}
}

// noopMD is a non-nil empty metadata.MD for carriers that need one.
var noopMD = metadata.New(nil)
