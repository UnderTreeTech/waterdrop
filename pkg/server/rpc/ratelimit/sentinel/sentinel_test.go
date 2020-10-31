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

package sentinel

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/UnderTreeTech/waterdrop/pkg/status"

	"github.com/stretchr/testify/assert"

	"github.com/UnderTreeTech/waterdrop/pkg/log"

	"github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/flow"

	"google.golang.org/grpc"
)

func TestMain(m *testing.M) {
	defer log.New(nil).Sync()
	if err := api.InitDefault(); err != nil {
		panic(fmt.Sprintf("init sentinel entity fail, error is %s", err.Error()))
	}
	code := m.Run()
	os.Exit(code)
}

func TestSentinelUnaryClientInterceptor(t *testing.T) {
	interceptor := SentinelForUnaryClient()
	invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		return nil
	}

	t.Run("success", func(t *testing.T) {
		_, err := flow.LoadRules([]*flow.Rule{
			{
				Resource:               "/grpc.testing.TestService/UnaryCall",
				Threshold:              1.0,
				TokenCalculateStrategy: flow.Direct,
				ControlBehavior:        flow.Reject,
				StatIntervalInMs:       1000,
			},
		})
		assert.Nil(t, err)
		err = interceptor(context.Background(), "/grpc.testing.TestService/UnaryCall", nil, nil, nil, invoker)
		assert.Nil(t, err)
		t.Run("request too fast", func(t *testing.T) {
			err = interceptor(context.Background(), "/grpc.testing.TestService/UnaryCall", nil, nil, nil, invoker)
			assert.IsType(t, &status.Status{}, err)
		})
	})

	t.Run("blocked", func(t *testing.T) {
		_, err := flow.LoadRules([]*flow.Rule{
			{
				Resource:               "/grpc.testing.TestService/UnaryCall",
				Threshold:              0.0,
				TokenCalculateStrategy: flow.Direct,
				ControlBehavior:        flow.Reject,
				StatIntervalInMs:       1000,
			},
		})
		assert.Nil(t, err)
		err = interceptor(context.Background(), "/grpc.testing.TestService/UnaryCall", nil, nil, nil, invoker)
		assert.IsType(t, &status.Status{}, err)
	})
}

func TestSentinelUnaryServerInterceptor(t *testing.T) {
	interceptor := SentinelForUnaryServer()
	handler := func(ctx context.Context, req interface{}) (resp interface{}, err error) {
		return nil, nil
	}
	info := &grpc.UnaryServerInfo{
		FullMethod: "/grpc.testing.TestService/UnaryCall",
	}

	t.Run("success", func(t *testing.T) {
		_, err := flow.LoadRules([]*flow.Rule{
			{
				Resource:               "/grpc.testing.TestService/UnaryCall",
				Threshold:              2.0,
				TokenCalculateStrategy: flow.Direct,
				ControlBehavior:        flow.Reject,
				StatIntervalInMs:       1000,
			},
		})
		assert.Nil(t, err)
		resp, err := interceptor(context.Background(), nil, info, handler)
		assert.Nil(t, resp)
		assert.Nil(t, err)

		t.Run("request too fast", func(t *testing.T) {
			resp, err := interceptor(context.Background(), nil, info, handler)
			assert.Nil(t, resp)
			assert.IsType(t, &status.Status{}, err)
		})
	})

	t.Run("blocked", func(t *testing.T) {
		_, err := flow.LoadRules([]*flow.Rule{
			{
				Resource:               "/grpc.testing.TestService/UnaryCall",
				Threshold:              0.0,
				TokenCalculateStrategy: flow.Direct,
				ControlBehavior:        flow.Reject,
				StatIntervalInMs:       1000,
			},
		})
		assert.Nil(t, err)
		resp, err := interceptor(context.Background(), nil, info, handler)
		assert.Nil(t, resp)
		assert.IsType(t, &status.Status{}, err)
	})
}
