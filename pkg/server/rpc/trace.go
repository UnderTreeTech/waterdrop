package rpc

import (
	"context"
	"time"

	"google.golang.org/grpc/peer"

	"github.com/opentracing/opentracing-go/ext"

	"github.com/UnderTreeTech/waterdrop/pkg/status"
	"github.com/UnderTreeTech/waterdrop/pkg/trace"
	"github.com/opentracing/opentracing-go/log"

	"google.golang.org/grpc"
)

func (s *Server) trace() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		ctx, opt := trace.FromIncomingContext(ctx)
		span, ctx := trace.StartSpanFromContext(ctx, info.FullMethod, opt)
		ext.Component.Set(span, "grpc")
		ext.SpanKind.Set(span, ext.SpanKindRPCServerEnum)
		if peer, ok := peer.FromContext(ctx); ok {
			ext.PeerAddress.Set(span, peer.Addr.String())
		}

		// adjust request timeout
		timeout := s.config.Timeout
		if deadline, ok := ctx.Deadline(); ok {
			derivedTimeout := time.Until(deadline)
			// reduce 10ms network transmission time for every request
			if derivedTimeout-5*time.Millisecond > 0 {
				derivedTimeout = derivedTimeout - 5*time.Millisecond
			}

			if timeout > derivedTimeout {
				timeout = derivedTimeout
			}
		}

		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer func() {
			span.Finish()
			cancel()
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

func (c *Client) trace() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
		ctx, md := trace.FromOutgoingContext(ctx)
		span, ctx := trace.StartSpanFromContext(ctx, method)
		ext.Component.Set(span, "grpc")
		ext.SpanKind.Set(span, ext.SpanKindRPCClientEnum)

		// adjust request timeout
		timeout := c.config.Timeout
		if deadline, ok := ctx.Deadline(); ok {
			derivedTimeout := time.Until(deadline)
			if timeout > derivedTimeout {
				timeout = derivedTimeout
			}
		}

		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer func() {
			span.Finish()
			cancel()
		}()

		trace.MetadataInjector(ctx, md)

		err = invoker(ctx, method, req, reply, cc, opts...)
		if err != nil {
			estatus := status.ExtractStatus(err)
			ext.Error.Set(span, true)
			span.LogFields(log.String("event", "error"), log.Int("code", estatus.Code()), log.String("message", estatus.Message()))
		}

		return
	}
}
