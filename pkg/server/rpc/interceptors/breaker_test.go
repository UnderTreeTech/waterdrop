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
	"testing"

	"github.com/UnderTreeTech/waterdrop/pkg/breaker"

	"github.com/UnderTreeTech/waterdrop/pkg/status"
	"github.com/stretchr/testify/assert"

	"google.golang.org/grpc"
)

func TestGoogleSREBreakerUnaryClientInterceptor(t *testing.T) {
	interceptor := GoogleSREBreaker(breaker.NewBreakerGroup())
	invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		return nil
	}

	t.Run("success", func(t *testing.T) {
		err := interceptor(context.Background(), "/grpc.testing.TestService/UnaryCall", nil, nil, nil, invoker)
		assert.Nil(t, err)
	})

	invoker = func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		return status.ServiceUnavailable
	}
	t.Run("blocked", func(t *testing.T) {
		var err error
		for i := 0; i < 10; i++ {
			err = interceptor(context.Background(), "/grpc.testing.TestService/UnaryCall", nil, nil, nil, invoker)
		}
		assert.IsType(t, &status.Status{}, err)
	})
}
