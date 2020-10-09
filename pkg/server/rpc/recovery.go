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

package rpc

import (
	"context"
	"runtime"

	"github.com/UnderTreeTech/waterdrop/pkg/status"

	"github.com/UnderTreeTech/waterdrop/pkg/log"

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
				err = status.ServerErr
			}
		}()

		resp, err = handler(ctx, req)
		return
	}
}

func (c *Client) recovery() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
		defer func() {
			if rerr := recover(); rerr != nil {
				stack := make([]byte, size)
				stack = stack[:runtime.Stack(stack, true)]
				log.Error(ctx, "panic request", log.Any("req", req), log.Any("err", rerr), log.Bytes("stack", stack))
				err = status.ServerErr
			}
		}()

		err = invoker(ctx, method, req, reply, cc, opts...)
		return
	}
}
