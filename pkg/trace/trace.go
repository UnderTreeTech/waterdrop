package trace

import (
	"context"

	jaeger "github.com/uber/jaeger-client-go"

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

// rpc: FromIncomingContext
func FromIncomingContext(ctx context.Context) (context.Context, opentracing.StartSpanOption) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(nil)
		ctx = metadata.NewIncomingContext(ctx, md)
	}

	sc, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, &Metadata{md: md})
	if err != nil {
		return ctx, NullStartSpanOption{}
	}

	return ctx, opentracing.ChildOf(sc)
}

// rpc: FromOutgoingContext
func FromOutgoingContext(ctx context.Context) (context.Context, metadata.MD) {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.New(nil)
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	return metadata.NewOutgoingContext(ctx, md), md
}

// rpc: MetadataInjector
func MetadataInjector(ctx context.Context, md metadata.MD) error {
	span := opentracing.SpanFromContext(ctx)
	err := opentracing.GlobalTracer().Inject(span.Context(), opentracing.HTTPHeaders, &Metadata{md: md})
	if err != nil {
		span.LogFields(log.String("event", "inject failed"), log.Error(err))
		return err
	}

	return nil
}

type httpCarrierKey struct{}

// HTTP HeaderExtractor
func HeaderExtractor(ctx context.Context, carrier map[string][]string) (context.Context, opentracing.StartSpanOption) {
	md := metadata.MD(carrier)
	ctx = metadata.NewIncomingContext(ctx, md)

	return FromIncomingContext(ctx)
}

// HTTP HeaderInjector
func HeaderInjector(ctx context.Context, carrier map[string][]string) context.Context {

	span := opentracing.SpanFromContext(ctx)
	err := opentracing.GlobalTracer().Inject(span.Context(), opentracing.HTTPHeaders, carrier)
	if err != nil {
		span.LogFields(log.String("event", "inject failed"), log.Error(err))
		return ctx
	}

	return context.WithValue(ctx, httpCarrierKey{}, carrier)
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
