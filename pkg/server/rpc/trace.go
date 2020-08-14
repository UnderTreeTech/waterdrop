package rpc

import (
	"context"
	"time"

	"github.com/opentracing/opentracing-go/log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"google.golang.org/grpc/peer"

	"github.com/opentracing/opentracing-go/ext"

	"github.com/UnderTreeTech/waterdrop/pkg/trace"

	"google.golang.org/grpc"
)

func (s *Server) trace() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		span, ctx := trace.StartSpanFromContext(ctx, info.FullMethod, trace.FromIncomingContext(ctx))
		ext.Component.Set(span, "grpc")
		ext.SpanKind.Set(span, ext.SpanKindRPCServerEnum)
		if peer, ok := peer.FromContext(ctx); ok {
			ext.PeerAddress.Set(span, peer.Addr.String())
		}

		// adjust request timeout
		timeout := s.config.Timeout
		if deadline, ok := ctx.Deadline(); ok {
			reqTimeout := time.Until(deadline)
			if timeout > reqTimeout {
				timeout = reqTimeout
			}
		}

		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer func() {
			span.Finish()
			cancel()
		}()

		resp, err = handler(ctx, req)
		if err != nil {
			code := codes.Unknown
			if s, ok := status.FromError(err); ok {
				code = s.Code()
			}
			span.SetTag("code", code)
			ext.Error.Set(span, true)
			span.LogFields(log.String("event", "error"), log.String("message", err.Error()))
		}

		return
	}
}
