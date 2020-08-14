package rpc

import (
	"context"

	"github.com/go-playground/validator/v10"

	"google.golang.org/grpc"
)

var v = validator.New()

// validate request params
func (s *Server) validate() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		if err = v.Struct(req); err != nil {
			return
		}
		return handler(ctx, req)
	}
}

// GetValidator returns the underlying validator engine which powers the
// StructValidator implementation.
func (s *Server) GetValidator() *validator.Validate {
	return v
}
