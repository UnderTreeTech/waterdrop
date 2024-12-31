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
	"fmt"

	"github.com/UnderTreeTech/waterdrop/pkg/server/rpc/config"

	"github.com/UnderTreeTech/waterdrop/pkg/server/rpc/metadata"

	"github.com/UnderTreeTech/waterdrop/pkg/server/rpc/interceptors"

	"github.com/UnderTreeTech/waterdrop/pkg/breaker"

	"google.golang.org/grpc/keepalive"

	"google.golang.org/grpc"
)

// Client grpc client definition
type Client struct {
	conn          *grpc.ClientConn
	config        *config.ClientConfig
	clientOptions []grpc.DialOption
	breakers      *breaker.BreakerGroup

	unaryInterceptors []grpc.UnaryClientInterceptor
}

// New returns a Client instance
func New(config *config.ClientConfig) *Client {
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

	if config.MaxCallSendMsgSize > 0 {
		cli.clientOptions = append(cli.clientOptions, grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(config.MaxCallSendMsgSize)))
	}

	keepaliveOpts := grpc.WithKeepaliveParams(keepalive.ClientParameters{
		Time:    config.KeepAliveInterval,
		Timeout: config.KeepAliveTimeout,
	})

	cli.Use(
		interceptors.RecoveryForUnaryClient(cli.config),
		interceptors.TraceForUnaryClient(),
		interceptors.LoggerForUnaryClient(cli.config),
		interceptors.GoogleSREBreaker(cli.breakers),
	)

	cli.clientOptions = append(
		cli.clientOptions,
		keepaliveOpts,
		grpc.WithInsecure(),
		// use WithDefaultServiceConfig to fix golinter staticcheck error
		// maybe it's better to use balancer config struct
		// you can get more detail at here: https://github.com/grpc/grpc-go/issues/3003
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"`+config.Balancer+`"}`),
		cli.WithUnaryServerChain(),
	)

	cc, err := grpc.DialContext(ctx, config.Target, cli.clientOptions...)
	if err != nil {
		panic(fmt.Sprintf("dial peer service fail, target %s, error %s", config.Target, err.Error()))
	}

	cli.conn = cc
	return cli
}

// ChainUnaryClient creates a single interceptor out of a chain of many interceptors.
// Execution is done in left-to-right order, including passing of context.
// For example ChainUnaryClient(one, two, three) will execute one before two before three.
func (c *Client) ChainUnaryClient() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		interceptors := c.unaryInterceptors
		n := len(interceptors)

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

// WithUnaryServerChain is a grpc.Client dial option that accepts multiple unary interceptors.
func (c *Client) WithUnaryServerChain() grpc.DialOption {
	return grpc.WithUnaryInterceptor(c.ChainUnaryClient())
}

// Use attaches a global interceptor to the client. ie. the interceptor attached through Use() will be
// included in the interceptors chain for every single request.
// For example, this is the right place for a logger or error management interceptor.
func (c *Client) Use(interceptors ...grpc.UnaryClientInterceptor) {
	finalSize := len(c.unaryInterceptors) + len(interceptors)
	if finalSize >= metadata.MaxInterceptors {
		panic("waterdrop: server use too many interceptors")
	}

	mergedInterceptors := make([]grpc.UnaryClientInterceptor, finalSize)
	copy(mergedInterceptors, c.unaryInterceptors)
	copy(mergedInterceptors[len(c.unaryInterceptors):], interceptors)

	c.unaryInterceptors = mergedInterceptors
}

// GetConn return the client connection
func (c *Client) GetConn() *grpc.ClientConn {
	return c.conn
}

// GetBreakers return the client breakers
func (c *Client) GetBreakers() *breaker.BreakerGroup {
	return c.breakers
}
