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
	"fmt"

	"github.com/UnderTreeTech/waterdrop/pkg/log"

	"github.com/UnderTreeTech/waterdrop/pkg/status"

	"github.com/alibaba/sentinel-golang/core/base"

	"github.com/alibaba/sentinel-golang/api"

	"google.golang.org/grpc"
)

// SentinelForUnaryClient is client side sentinel
func SentinelForUnaryClient() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
		entry, blockErr := api.Entry(method, api.WithResourceType(base.ResTypeRPC), api.WithTrafficType(base.Outbound))
		if blockErr != nil {
			log.Warn(ctx,
				"rpc hit rate limit",
				log.String("kind", "client"),
				log.String("method", method),
				log.String("error", blockErr.Error()),
			)
			return status.LimitExceed
		}
		defer entry.Exit()
		fmt.Println("method", method)
		err = invoker(ctx, method, req, reply, cc, opts...)
		return
	}
}

// SentinelForUnaryServer is server side sentinel
func SentinelForUnaryServer() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		entry, blockErr := api.Entry(info.FullMethod, api.WithResourceType(base.ResTypeRPC), api.WithTrafficType(base.Inbound))
		if blockErr != nil {
			log.Warn(ctx,
				"rpc hit rate limit",
				log.String("kind", "server"),
				log.String("method", info.FullMethod),
				log.String("error", blockErr.Error()),
			)
			return nil, status.LimitExceed
		}
		defer entry.Exit()

		resp, err = handler(ctx, req)
		return
	}
}
