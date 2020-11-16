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
	"testing"

	"github.com/UnderTreeTech/waterdrop/pkg/trace"

	"github.com/opentracing/opentracing-go"
	jconfig "github.com/uber/jaeger-client-go/config"

	"github.com/UnderTreeTech/waterdrop/pkg/log"

	"github.com/stretchr/testify/assert"

	"google.golang.org/grpc"
)

func newJaegerClient() (opentracing.Tracer, func()) {
	var configuration = jconfig.Configuration{
		ServiceName: "trace",
	}

	tracer, closer, err := configuration.NewTracer()
	if err != nil {
		panic(fmt.Sprintf("new jaeger trace fail, err msg %s", err.Error()))
	}

	return tracer, func() { closer.Close() }
}

func TestTraceForUnaryServer(t *testing.T) {
	interceptor := TraceForUnaryServer()
	handler := func(ctx context.Context, req interface{}) (resp interface{}, err error) {
		log.Info(ctx, "test server trace")
		return
	}
	info := &grpc.UnaryServerInfo{
		FullMethod: "/grpc.testing.TestService/UnaryCall",
	}

	t.Run("mock trace", func(t *testing.T) {
		resp, err := interceptor(context.Background(), nil, info, handler)
		assert.Nil(t, resp)
		assert.Nil(t, err)
	})

	tracer, close := newJaegerClient()
	trace.SetGlobalTracer(tracer)
	defer close()
	t.Run("jaeger trace", func(t *testing.T) {
		resp, err := interceptor(context.Background(), nil, info, handler)
		assert.Nil(t, resp)
		assert.Nil(t, err)
	})
}

func TestTraceForUnaryClient(t *testing.T) {
	interceptor := TraceForUnaryClient()
	invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) (err error) {
		log.Info(ctx, "test client trace")
		return
	}

	t.Run("mock trace", func(t *testing.T) {
		err := interceptor(context.Background(), "/grpc.testing.TestService/UnaryCall", nil, nil, nil, invoker)
		assert.Nil(t, err)
	})

	tracer, close := newJaegerClient()
	trace.SetGlobalTracer(tracer)
	defer close()
	t.Run("jaeger trace", func(t *testing.T) {
		err := interceptor(context.Background(), "/grpc.testing.TestService/UnaryCall", nil, nil, nil, invoker)
		assert.Nil(t, err)
	})
}
