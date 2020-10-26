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
	"fmt"
	"time"

	"github.com/UnderTreeTech/waterdrop/pkg/breaker"

	"google.golang.org/grpc/keepalive"

	"google.golang.org/grpc"
)

type ClientConfig struct {
	DialTimeout time.Duration
	Block       bool
	Balancer    string
	Target      string

	Timeout time.Duration

	KeepAliveInterval time.Duration
	KeepAliveTimeout  time.Duration

	SlowRequestDuration time.Duration
}

func defaultClientConfig() *ClientConfig {
	return &ClientConfig{
		DialTimeout: 5 * time.Second,
		Block:       true,
		Balancer:    "round_robin",

		Timeout: 500 * time.Millisecond,

		KeepAliveInterval: 60 * time.Second,
		KeepAliveTimeout:  20 * time.Second,
	}
}

type Client struct {
	conn          *grpc.ClientConn
	config        *ClientConfig
	clientOptions []grpc.DialOption
	breakers      *breaker.BreakerGroup

	unaryInterceptors []grpc.UnaryClientInterceptor
}

func NewClient(config *ClientConfig) *grpc.ClientConn {
	cli := &Client{
		config:   config,
		breakers: breaker.NewBreakerGroup(),

		clientOptions:     make([]grpc.DialOption, 0),
		unaryInterceptors: make([]grpc.UnaryClientInterceptor, 0),
	}

	ctx := context.Background()
	if config.Block {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, config.DialTimeout)
		defer cancel()

		cli.clientOptions = append(cli.clientOptions, grpc.WithBlock())
	}

	keepaliveOpts := grpc.WithKeepaliveParams(keepalive.ClientParameters{
		Time:    config.KeepAliveInterval,
		Timeout: config.KeepAliveTimeout,
	})

	cli.Use(cli.recovery(), cli.breaker(), cli.trace(), cli.logger())
	cli.clientOptions = append(
		cli.clientOptions,
		keepaliveOpts,
		grpc.WithInsecure(),
		// grpc.WithBalancerName(config.Balancer),
		// use WithDefaultServiceConfig to fix golinter staticcheck error
		// maybe it's better to use balancer config struct, you can get more detail at here: https://github.com/grpc/grpc-go/issues/3003
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"`+config.Balancer+`"}`),
		cli.WithUnaryServerChain(cli.unaryInterceptors...),
	)

	cc, err := grpc.DialContext(ctx, config.Target, cli.clientOptions...)
	if err != nil {
		panic(fmt.Sprintf("dial peer service fail, target %s, error %s", config.Target, err.Error()))
	}

	return cc
}

// ChainUnaryClient creates a single interceptor out of a chain of many interceptors.
//
// Execution is done in left-to-right order, including passing of context.
// For example ChainUnaryClient(one, two, three) will execute one before two before three.
func (c *Client) ChainUnaryClient(interceptors ...grpc.UnaryClientInterceptor) grpc.UnaryClientInterceptor {
	n := len(interceptors)

	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		chainer := func(currentInter grpc.UnaryClientInterceptor, currentInvoker grpc.UnaryInvoker) grpc.UnaryInvoker {
			return func(currentCtx context.Context, currentMethod string, currentReq, currentRepl interface{}, currentConn *grpc.ClientConn, currentOpts ...grpc.CallOption) error {
				return currentInter(currentCtx, currentMethod, currentReq, currentRepl, currentConn, currentInvoker, currentOpts...)
			}
		}

		chainedInvoker := invoker
		for i := n - 1; i >= 0; i-- {
			chainedInvoker = chainer(interceptors[i], chainedInvoker)
		}

		return chainedInvoker(ctx, method, req, reply, cc, opts...)
	}
}

// Chain creates a single interceptor out of a chain of many interceptors.
//
// WithUnaryServerChain is a grpc.Client dial option that accepts multiple unary interceptors.
// Basically syntactic sugar.
func (c *Client) WithUnaryServerChain(interceptors ...grpc.UnaryClientInterceptor) grpc.DialOption {
	return grpc.WithUnaryInterceptor(c.ChainUnaryClient(interceptors...))
}

// Use attaches a global interceptor to the client. ie. the interceptor attached through Use() will be
// included in the interceptors chain for every single request.
// For example, this is the right place for a logger or error management interceptor.
func (c *Client) Use(interceptors ...grpc.UnaryClientInterceptor) {
	finalSize := len(c.unaryInterceptors) + len(interceptors)
	if finalSize >= int(_abortIndex) {
		panic("waterdrop: server use too many interceptors")
	}

	mergedInterceptors := make([]grpc.UnaryClientInterceptor, finalSize)
	copy(mergedInterceptors, c.unaryInterceptors)
	copy(mergedInterceptors[len(c.unaryInterceptors):], interceptors)

	c.unaryInterceptors = mergedInterceptors
}
