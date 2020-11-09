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
	"time"

	"github.com/UnderTreeTech/waterdrop/pkg/status"

	"github.com/UnderTreeTech/waterdrop/pkg/log"

	"google.golang.org/grpc"
)

const size = 4 << 10

func recoveryForUnaryServer(config *ServerConfig) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if rerr := recover(); rerr != nil {
				stack := make([]byte, size)
				stack = stack[:runtime.Stack(stack, true)]
				log.Error(ctx, "panic request", log.Any("req", req), log.Any("err", rerr), log.Bytes("stack", stack))
				err = status.ServerErr
			}
		}()

		// adjust request timeout
		timeout := config.Timeout
		if deadline, ok := ctx.Deadline(); ok {
			derivedTimeout := time.Until(deadline)
			// reduce 5ms network transmission time for every request
			if derivedTimeout-5*time.Millisecond > 0 {
				derivedTimeout = derivedTimeout - 5*time.Millisecond
			}

			if timeout > derivedTimeout {
				timeout = derivedTimeout
			}
		}

		// if zero timeout config means never timeout
		var cancel func()
		if timeout > 0 {
			ctx, cancel = context.WithTimeout(ctx, timeout)
		} else {
			cancel = func() {}
		}
		defer cancel()

		resp, err = handler(ctx, req)
		return
	}
}

func recoveryForUnaryClient(config *ClientConfig) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
		defer func() {
			if rerr := recover(); rerr != nil {
				stack := make([]byte, size)
				stack = stack[:runtime.Stack(stack, true)]
				log.Error(ctx, "panic request", log.Any("req", req), log.Any("err", rerr), log.Bytes("stack", stack))
				err = status.ServerErr
			}
		}()

		// adjust request timeout
		timeout := config.Timeout
		if deadline, ok := ctx.Deadline(); ok {
			derivedTimeout := time.Until(deadline)
			if timeout > derivedTimeout {
				timeout = derivedTimeout
			}
		}

		// if zero timeout config means never timeout
		var cancel func()
		if timeout > 0 {
			ctx, cancel = context.WithTimeout(ctx, timeout)
		} else {
			cancel = func() {}
		}
		defer cancel()

		err = invoker(ctx, method, req, reply, cc, opts...)
		return
	}
}
