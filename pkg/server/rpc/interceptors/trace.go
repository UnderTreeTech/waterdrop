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

package interceptors

import (
	"context"

	"google.golang.org/grpc/metadata"

	"google.golang.org/grpc/peer"

	"github.com/opentracing/opentracing-go/ext"

	"github.com/UnderTreeTech/waterdrop/pkg/status"
	"github.com/UnderTreeTech/waterdrop/pkg/trace"
	"github.com/UnderTreeTech/waterdrop/pkg/trace/otel"
	"github.com/opentracing/opentracing-go/log"

	otelcodes "go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"

	"google.golang.org/grpc"
)

// TraceForUnaryServer trace unary server side details.
//
// Starts the opentracing/jaeger server span (unchanged) and, when OTel is
// enabled, also starts an OTel server span extracted from incoming gRPC
// metadata via the composite propagator (traceparent + uber-trace-id). The
// OTel span shares the caller's TraceID so it lands in the same langfuse
// trace as the upstream anchor span and the ADK LLM spans.
func TraceForUnaryServer() grpc.UnaryServerInterceptor {
	tracer := otel.Tracer("waterdrop/grpc")
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		opt := trace.FromIncomingContext(ctx)
		span, ctx := trace.StartSpanFromContext(ctx, info.FullMethod, opt)
		ext.Component.Set(span, "grpc")
		ext.SpanKind.Set(span, ext.SpanKindRPCServerEnum)
		if peer, ok := peer.FromContext(ctx); ok {
			ext.PeerAddress.Set(span, peer.Addr.String())
		}

		var ospan oteltrace.Span
		if otel.Enabled() {
			md, ok := metadata.FromIncomingContext(ctx)
			if !ok {
				md = metadata.New(nil)
			}
			octx := otel.Propagator().Extract(ctx, otel.MDCarrier{MD: md})
			octx, ospan = tracer.Start(octx, info.FullMethod,
				oteltrace.WithSpanKind(oteltrace.SpanKindServer),
			)
			ctx = octx
		}
		defer func() {
			if ospan != nil {
				if err != nil {
					ospan.SetStatus(otelcodes.Error, err.Error())
				}
				ospan.End()
			}
			span.Finish()
		}()

		resp, err = handler(ctx, req)
		if err != nil {
			estatus := status.ExtractStatus(err)
			ext.Error.Set(span, true)
			span.LogFields(log.String("event", "error"), log.Int("code", estatus.Code()), log.String("message", estatus.Message()))
		}

		return
	}
}

// TraceForUnaryClient trace unary client side details.
//
// Starts the opentracing/jaeger client span (unchanged) and, when OTel is
// enabled, also starts an OTel client span and injects it into outgoing gRPC
// metadata via the composite propagator (traceparent + uber-trace-id) so the
// downstream server continues the same trace.
func TraceForUnaryClient() grpc.UnaryClientInterceptor {
	tracer := otel.Tracer("waterdrop/grpc")
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		}
		span, ctx := trace.StartSpanFromContext(ctx, method)
		ext.Component.Set(span, "grpc")
		ext.SpanKind.Set(span, ext.SpanKindRPCClientEnum)

		var ospan oteltrace.Span
		if otel.Enabled() {
			ctx, ospan = tracer.Start(ctx, method,
				oteltrace.WithSpanKind(oteltrace.SpanKindClient),
			)
			// inject OTel context (traceparent + uber-trace-id) into outgoing metadata
			otel.Propagator().Inject(ctx, otel.MDCarrier{MD: md})
		}

		ctx = trace.MetadataInjector(ctx, md)
		defer func() {
			if ospan != nil {
				if err != nil {
					ospan.SetStatus(otelcodes.Error, err.Error())
				}
				ospan.End()
			}
			span.Finish()
		}()

		err = invoker(ctx, method, req, reply, cc, opts...)
		if err != nil {
			estatus := status.ExtractStatus(err)
			ext.Error.Set(span, true)
			span.LogFields(log.String("event", "error"), log.Int("code", estatus.Code()), log.String("message", estatus.Message()))
		}
		return
	}
}
