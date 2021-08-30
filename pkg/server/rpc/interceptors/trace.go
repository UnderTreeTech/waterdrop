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

package interceptors

import (
	"context"

	"google.golang.org/grpc/metadata"

	"google.golang.org/grpc/peer"

	"github.com/opentracing/opentracing-go/ext"

	"github.com/UnderTreeTech/waterdrop/pkg/status"
	"github.com/UnderTreeTech/waterdrop/pkg/trace"
	"github.com/opentracing/opentracing-go/log"

	"google.golang.org/grpc"
)

// TraceForUnaryServer trace unary server side details
func TraceForUnaryServer() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		opt := trace.FromIncomingContext(ctx)
		span, ctx := trace.StartSpanFromContext(ctx, info.FullMethod, opt)
		ext.Component.Set(span, "grpc")
		ext.SpanKind.Set(span, ext.SpanKindRPCServerEnum)
		if peer, ok := peer.FromContext(ctx); ok {
			ext.PeerAddress.Set(span, peer.Addr.String())
		}
		defer span.Finish()

		resp, err = handler(ctx, req)
		if err != nil {
			estatus := status.ExtractStatus(err)
			ext.Error.Set(span, true)
			span.LogFields(log.String("event", "error"), log.Int("code", estatus.Code()), log.String("message", estatus.Message()))
		}

		return
	}
}

// TraceForUnaryClient trace unary client side details
func TraceForUnaryClient() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		}
		span, ctx := trace.StartSpanFromContext(ctx, method)
		ext.Component.Set(span, "grpc")
		ext.SpanKind.Set(span, ext.SpanKindRPCClientEnum)
		defer span.Finish()

		ctx = trace.MetadataInjector(ctx, md)
		err = invoker(ctx, method, req, reply, cc, opts...)
		if err != nil {
			estatus := status.ExtractStatus(err)
			ext.Error.Set(span, true)
			span.LogFields(log.String("event", "error"), log.Int("code", estatus.Code()), log.String("message", estatus.Message()))
		}
		return
	}
}
