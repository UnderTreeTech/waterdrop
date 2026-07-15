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
	"net/http"

	"github.com/opentracing/opentracing-go/log"
	"google.golang.org/grpc/metadata"

	opentracing "github.com/opentracing/opentracing-go"
)

type NullStartSpanOption struct{}

func (n NullStartSpanOption) Apply(options *opentracing.StartSpanOptions) {}

// SetGlobalTracer set global trace instance
func SetGlobalTracer(tracer opentracing.Tracer) {
	opentracing.SetGlobalTracer(tracer)
}

// StartSpanFromContext start a span from current context
func StartSpanFromContext(ctx context.Context, op string, opts ...opentracing.StartSpanOption) (opentracing.Span, context.Context) {
	return opentracing.StartSpanFromContext(ctx, op, opts...)
}

// SpanFromContext return the span attached to current context
func SpanFromContext(ctx context.Context) opentracing.Span {
	return opentracing.SpanFromContext(ctx)
}

// ContextWithSpan return a new `context.Context` that holds a reference to the span
func ContextWithSpan(ctx context.Context, span opentracing.Span) context.Context {
	return opentracing.ContextWithSpan(ctx, span)
}

// FromIncomingContext extract trace info from span
func FromIncomingContext(ctx context.Context) opentracing.StartSpanOption {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	}

	sc, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, CarrierMD{md: md})
	if err != nil {
		return NullStartSpanOption{}
	}
	return opentracing.ChildOf(sc)
}

// MetadataInjector inject trace info to span
func MetadataInjector(ctx context.Context, md metadata.MD) context.Context {
	span := opentracing.SpanFromContext(ctx)
	err := opentracing.GlobalTracer().Inject(span.Context(), opentracing.HTTPHeaders, CarrierMD{md: md})
	if err != nil {
		span.LogFields(log.String("event", "inject failed"), log.Error(err))
		return ctx
	}
	return metadata.NewOutgoingContext(ctx, md)
}

type hdOutgoingKey struct{}

// HeaderInjector inject trace info to span
func HeaderInjector(ctx context.Context, carrier http.Header) context.Context {
	md := opentracing.HTTPHeadersCarrier(carrier)
	span := opentracing.SpanFromContext(ctx)
	err := opentracing.GlobalTracer().Inject(span.Context(), opentracing.HTTPHeaders, md)
	if err != nil {
		span.LogFields(log.String("event", "inject failed"), log.Error(err))
		return ctx
	}
	return context.WithValue(ctx, hdOutgoingKey{}, carrier)
}

// HeaderExtractor extract trace info from span
func HeaderExtractor(carrier http.Header) opentracing.StartSpanOption {
	md := opentracing.HTTPHeadersCarrier(carrier)
	sc, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, md)
	if err != nil {
		return NullStartSpanOption{}
	}
	return opentracing.ChildOf(sc)
}

// TraceID returns the trace id of the active span on ctx as a string.
//
// It delegates to the factory-selected global Tracer so it returns the correct
// id for the active backend: the jaeger 64-bit id in jaeger mode, or the OTel
// 128-bit id in otel/dual mode. Before Init() it returns "" (noopTracer).
func TraceID(ctx context.Context) string {
	return globalTracer.TraceID(ctx)
}
