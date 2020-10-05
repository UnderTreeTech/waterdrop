package rpc

import (
	"context"

	"github.com/UnderTreeTech/waterdrop/pkg/status"

	"google.golang.org/grpc"
)

func (c *Client) breaker() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		return c.breakers.Do(
			method,
			func() error {
				return invoker(ctx, method, req, reply, cc, opts...)
			},
			func(err error) bool {
				switch status.ExtractStatus(err).Code() {
				case status.Deadline.Code(), status.LimitExceed.Code(),
					status.ServerErr.Code(), status.Canceled.Code(),
					status.ServiceUnavailable.Code():
					return false
				default:
					return true
				}
				return true
			})
	}
}
