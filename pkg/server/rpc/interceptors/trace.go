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

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"

	"github.com/UnderTreeTech/waterdrop/pkg/status"
	"github.com/UnderTreeTech/waterdrop/pkg/trace"
)

// TraceForUnaryServer trace unary server side details.
//
// Starts a server span via the factory-selected trace.Tracer, extracting the
// parent from incoming gRPC metadata. The backend-specific extract/span logic
// lives in the Tracer implementation; this interceptor holds no
// jaeger/OTel branches.
func TraceForUnaryServer() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		}
		carrier := trace.GRPCCarrier{MD: md}
		span, ctx := trace.GlobalTracer().StartServerSpan(ctx, info.FullMethod, trace.SpanServer, carrier)
		span.SetStringAttr("rpc.system", "grpc")
		if p, ok := peer.FromContext(ctx); ok {
			span.SetStringAttr("rpc.peer", p.Addr.String())
		}
		defer func() {
			if err != nil {
				span.SetStatus(false, err.Error())
			}
			span.End()
		}()

		resp, err = handler(ctx, req)
		if err != nil {
			estatus := status.ExtractStatus(err)
			span.SetIntAttr("rpc.code", int64(estatus.Code()))
			span.SetStringAttr("rpc.message", estatus.Message())
		}
		return
	}
}

// TraceForUnaryClient trace unary client side details.
//
// Starts a client span via the factory-selected trace.Tracer and injects its
// context into the outgoing gRPC metadata so the downstream server continues
// the trace.
func TraceForUnaryClient() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		}
		carrier := trace.GRPCCarrier{MD: md}
		span, ctx := trace.GlobalTracer().StartClientSpan(ctx, method, trace.SpanClient, carrier)
		span.SetStringAttr("rpc.system", "grpc")
		ctx = metadata.NewOutgoingContext(ctx, md)
		defer func() {
			if err != nil {
				span.SetStatus(false, err.Error())
			}
			span.End()
		}()

		err = invoker(ctx, method, req, reply, cc, opts...)
		if err != nil {
			estatus := status.ExtractStatus(err)
			span.SetIntAttr("rpc.code", int64(estatus.Code()))
			span.SetStringAttr("rpc.message", estatus.Message())
		}
		return
	}
}
