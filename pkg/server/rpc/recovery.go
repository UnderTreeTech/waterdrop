package rpc

import (
	"context"
	"runtime"

	"github.com/UnderTreeTech/waterdrop/pkg/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"google.golang.org/grpc"
)

const size = 4 << 10

func (s *Server) recovery() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if rerr := recover(); rerr != nil {
				stack := make([]byte, size)
				stack = stack[:runtime.Stack(stack, true)]
				log.Error(ctx, "panic request", log.Any("req", req), log.Any("err", rerr), log.Bytes("stack", stack))
				err = status.Errorf(codes.Unknown, "500")
			}
		}()

		resp, err = handler(ctx, req)
		return
	}
}
