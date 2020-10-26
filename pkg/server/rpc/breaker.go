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
			})
	}
}
