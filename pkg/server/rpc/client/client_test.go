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

package client

import (
	"context"
	"testing"
	"time"

	"github.com/UnderTreeTech/waterdrop/examples/proto/demo"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/UnderTreeTech/waterdrop/pkg/log"

	"github.com/stretchr/testify/assert"

	"github.com/UnderTreeTech/waterdrop/pkg/server/rpc/config"
	"github.com/UnderTreeTech/waterdrop/pkg/server/rpc/server"
)

// TestClient test grpc client
func TestClient(t *testing.T) {
	defer log.New(nil).Sync()
	srv := server.New(&config.ServerConfig{
		Addr: "0.0.0.0:21819",
	})
	demo.RegisterDemoServer(srv.Server(), &service{})
	srv.Start()
	time.Sleep(time.Millisecond * 100)
	defer srv.Stop(context.Background())

	cfg := &config.ClientConfig{
		DialTimeout: 150 * time.Millisecond,
		Block:       false,
		Balancer:    "round_robin",
		Target:      "127.0.0.1:21819",
	}
	client := New(cfg)
	rpc := demo.NewDemoClient(client.GetConn())
	reply, err := rpc.SayHelloURL(context.Background(), &demo.HelloReq{Name: "waterdrop"})
	assert.Equal(t, reply.Content, "Hello waterdrop")
	assert.Nil(t, err)
}

// TestDialTimeout test dial timeout
func TestDialTimeout(t *testing.T) {
	defer log.New(nil).Sync()
	flag := false
	defer func() {
		if r := recover(); r != nil {
			flag = true
		}
		assert.Equal(t, flag, true)
	}()

	srv := server.New(&config.ServerConfig{
		Addr: "0.0.0.0:21819",
	})
	demo.RegisterDemoServer(srv.Server(), &service{})
	go func() {
		time.Sleep(time.Millisecond * 100)
		srv.Start()
	}()
	defer srv.Stop(context.Background())

	cfg := &config.ClientConfig{
		DialTimeout: 50 * time.Millisecond,
		Block:       true,
		Balancer:    "round_robin",
		Target:      "127.0.0.1:21819",
	}
	client := New(cfg)
	demo.NewDemoClient(client.GetConn())
}

type service struct{}

func (s *service) SayHello(ctx context.Context, req *demo.HelloReq) (reply *emptypb.Empty, err error) {
	reply = &emptypb.Empty{}
	return reply, nil
}

func (s *service) SayHelloURL(ctx context.Context, req *demo.HelloReq) (reply *demo.HelloResp, err error) {
	reply = &demo.HelloResp{Content: "Hello " + req.Name}
	return
}
