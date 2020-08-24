package rpc

import (
	"context"
	"time"

	"github.com/UnderTreeTech/waterdrop/pkg/status"

	"github.com/UnderTreeTech/waterdrop/pkg/log"

	"google.golang.org/grpc/peer"

	"google.golang.org/grpc"
)

func (s *Server) logger() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		now := time.Now()
		var clientIP string
		if peer, ok := peer.FromContext(ctx); ok {
			clientIP = peer.Addr.String()
		}

		var quota float64
		if deadline, ok := ctx.Deadline(); ok {
			quota = time.Until(deadline).Seconds()
		}

		// call server interceptor
		resp, err = handler(ctx, req)
		var errmsg, retcode string
		if err != nil {
			estatus := status.ExtractStatus(err)
			retcode = estatus.Error()
			errmsg = estatus.Message()
		}

		duration := time.Since(now)

		fields := make([]log.Field, 0, 7)
		fields = append(
			fields,
			log.String("client_ip", clientIP),
			log.String("method", info.FullMethod),
			log.Any("req", req),
			log.Float64("quota", quota),
			log.Float64("duration", duration.Seconds()),
			log.String("code", retcode),
			log.String("error", errmsg),
		)

		if duration >= s.config.SlowRequestDuration {
			log.Warn(ctx, "grpc-slow-access-log", fields...)
		} else {
			log.Info(ctx, "grpc-access-log", fields...)
		}

		return
	}
}
