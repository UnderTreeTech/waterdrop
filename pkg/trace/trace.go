package trace

import (
	"context"

	jaeger "github.com/uber/jaeger-client-go"

	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"google.golang.org/grpc/metadata"

	opentracing "github.com/opentracing/opentracing-go"
)

type NullStartSpanOption struct{}

func (sso NullStartSpanOption) Apply(options *opentracing.StartSpanOptions) {}

func SetGlobalTracer(tracer opentracing.Tracer) {
	opentracing.SetGlobalTracer(tracer)
}

func StartSpanFromContext(ctx context.Context, op string, opts ...opentracing.StartSpanOption) (opentracing.Span, context.Context) {
	return opentracing.StartSpanFromContext(ctx, op, opts...)
}

func SpanFromContext(ctx context.Context) opentracing.Span {
	return opentracing.SpanFromContext(ctx)
}

// FromIncomingContext ...
func FromIncomingContext(ctx context.Context) opentracing.StartSpanOption {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	}
	sc, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, md)
	if err != nil {
		return NullStartSpanOption{}
	}

	return ext.RPCServerOption(sc)
}

type httpCarrierKey struct{}

// HeaderExtractor ...
func HeaderExtractor(carrier opentracing.HTTPHeadersCarrier) opentracing.StartSpanOption {
	sc, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, carrier)
	if err != nil {
		return NullStartSpanOption{}
	}

	return opentracing.ChildOf(sc)
}

func HeaderInjector(ctx context.Context, carrier opentracing.HTTPHeadersCarrier) context.Context {
	span := opentracing.SpanFromContext(ctx)
	err := opentracing.GlobalTracer().Inject(span.Context(), opentracing.HTTPHeaders, carrier)
	if err != nil {
		span.LogFields(log.String("event", "inject failed"), log.Error(err))
		return ctx
	}

	return context.WithValue(ctx, httpCarrierKey{}, carrier)
}

// MetadataExtractor ...
func MetadataExtractor(carrier metadata.MD) opentracing.StartSpanOption {
	sc, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, carrier)
	if err != nil {
		return NullStartSpanOption{}
	}

	return opentracing.ChildOf(sc)
}

// MetadataInjector ...
func MetadataInjector(ctx context.Context, carrier metadata.MD) context.Context {
	span := opentracing.SpanFromContext(ctx)
	err := opentracing.GlobalTracer().Inject(span.Context(), opentracing.HTTPHeaders, carrier)
	if err != nil {
		span.LogFields(log.String("event", "inject failed"), log.Error(err))
		return ctx
	}

	return metadata.NewOutgoingContext(ctx, carrier)
}

func TraceID(ctx context.Context) string {
	sp := SpanFromContext(ctx)
	if sp == nil {
		return ""
	}

	if jsc, ok := sp.Context().(jaeger.SpanContext); ok {
		return jsc.TraceID().String()
	}

	return ""
}
