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
	"testing"

	"github.com/stretchr/testify/assert"

	"google.golang.org/grpc"
)

func TestLoggerForUnaryServer(t *testing.T) {
	interceptor := loggerForUnaryServer(defaultServerConfig())
	handler := func(ctx context.Context, req interface{}) (resp interface{}, err error) {
		return
	}
	info := &grpc.UnaryServerInfo{
		FullMethod: "/grpc.testing.TestService/UnaryCall",
	}

	t.Run("server log", func(t *testing.T) {
		resp, err := interceptor(context.Background(), nil, info, handler)
		assert.Nil(t, resp, nil)
		assert.Nil(t, err, nil)
	})
}

func TestLoggerForUnaryClient(t *testing.T) {
	interceptor := loggerForUnaryClient(defaultClientConfig())
	invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) (err error) {
		return nil
	}

	t.Run("client log", func(t *testing.T) {
		err := interceptor(context.Background(), "/grpc.testing.TestService/UnaryCall", nil, nil, nil, invoker)
		assert.Nil(t, err, nil)
	})
}