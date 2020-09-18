package rpc

import (
	"context"
	"time"

	"github.com/UnderTreeTech/waterdrop/pkg/status"

	"google.golang.org/grpc/peer"

	"github.com/UnderTreeTech/waterdrop/pkg/stats/metric"

	"google.golang.org/grpc"
)

func (s *Server) Metric() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		now := time.Now()
		var ip string
		if peer, ok := peer.FromContext(ctx); ok {
			ip = peer.Addr.String()
		}

		// call server interceptor
		resp, err = handler(ctx, req)
		estatus := status.ExtractStatus(err)

		metric.UnaryServerHandleCounter.Inc(ip, info.FullMethod, estatus.Error())
		metric.UnaryServerReqDuration.Observe(time.Since(now).Seconds(), ip, info.FullMethod)

		return
	}
}
