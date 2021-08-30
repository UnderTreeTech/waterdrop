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

package server

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/UnderTreeTech/waterdrop/pkg/server/rpc/config"

	"github.com/UnderTreeTech/waterdrop/pkg/server/rpc/metadata"

	"github.com/UnderTreeTech/waterdrop/pkg/server/rpc/interceptors"

	"google.golang.org/grpc/reflection"

	"google.golang.org/grpc/keepalive"

	// Automatically set GOMAXPROCS to match Linux container CPU quota
	_ "go.uber.org/automaxprocs"
	"google.golang.org/grpc"
)

// Server rpc server definition
type Server struct {
	server *grpc.Server
	config *config.ServerConfig

	serverOptions     []grpc.ServerOption
	unaryInterceptors []grpc.UnaryServerInterceptor
}

// New returns a rpc Server instance
func New(cfg *config.ServerConfig) *Server {
	if cfg == nil {
		cfg = config.DefaultServerConfig()
	}

	srv := &Server{
		config:            cfg,
		serverOptions:     make([]grpc.ServerOption, 0),
		unaryInterceptors: make([]grpc.UnaryServerInterceptor, 0),
	}

	keepaliveOpts := grpc.KeepaliveParams(keepalive.ServerParameters{
		MaxConnectionIdle:     cfg.IdleTimeout,
		MaxConnectionAgeGrace: cfg.ForceCloseWait,
		Time:                  cfg.KeepAliveInterval,
		Timeout:               cfg.KeepAliveTimeout,
		MaxConnectionAge:      cfg.MaxLifeTime,
	})

	srv.Use(
		interceptors.RecoveryForUnaryServer(srv.config),
		interceptors.TraceForUnaryServer(),
		interceptors.LoggerForUnaryServer(srv.config),
		interceptors.Metric(),
	)
	srv.serverOptions = append(srv.serverOptions, keepaliveOpts, srv.WithUnaryServerChain())
	srv.server = grpc.NewServer(srv.serverOptions...)
	return srv
}

// Start rpc server
func (s *Server) Start() net.Addr {
	listener, err := net.Listen("tcp", s.config.Addr)
	if err != nil {
		panic(fmt.Sprintf("grpc server: listen tcp fail,err msg %s", err.Error()))
	}

	reflection.Register(s.server)
	go func() {
		if err := s.server.Serve(listener); err != nil {
			if err == grpc.ErrServerStopped {
				log.Printf("waterdrop: grpc server closed")
				return
			}

			panic(fmt.Sprintf("GRPC Server serve fail,err msg %s", err.Error()))
		}
	}()

	log.Printf("GRPC Server: start grpc listen addr: %s", listener.Addr().String())
	return listener.Addr()
}

// Stop stops the gRPC server gracefully. It stops the server from
// accepting new connections and RPCs and blocks until all the pending RPCs are
// finished.
func (s *Server) Stop(ctx context.Context) error {
	var err error
	ch := make(chan struct{})

	go func() {
		s.server.GracefulStop()
		close(ch)
	}()

	select {
	case <-ctx.Done():
		s.server.Stop()
		err = ctx.Err()
	case <-ch:
	}

	return err
}

// Server returns underlying grpc Server
func (s *Server) Server() *grpc.Server {
	return s.server
}

// ChainUnaryServer creates a single interceptor out of a chain of many interceptors.
// Execution is done in left-to-right order, including passing of context.
// For example ChainUnaryServer(one, two, three) will execute one before two before three, and three
// will see context changes of one and two.
func (s *Server) ChainUnaryServer() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		interceptors := s.unaryInterceptors
		n := len(interceptors)

		chainer := func(currentInter grpc.UnaryServerInterceptor, currentHandler grpc.UnaryHandler) grpc.UnaryHandler {
			return func(currentCtx context.Context, currentReq interface{}) (interface{}, error) {
				return currentInter(currentCtx, currentReq, info, currentHandler)
			}
		}

		chainedHandler := handler
		for i := n - 1; i >= 0; i-- {
			chainedHandler = chainer(interceptors[i], chainedHandler)
		}

		return chainedHandler(ctx, req)
	}
}

// Chain creates a single interceptor out of a chain of many interceptors.
// WithUnaryServerChain is a grpc.Server config option that accepts multiple unary interceptors.
// Basically syntactic sugar.
func (s *Server) WithUnaryServerChain() grpc.ServerOption {
	return grpc.UnaryInterceptor(s.ChainUnaryServer())
}

// Use attaches a global interceptor to the server. ie. the interceptor attached through Use() will be
// included in the interceptors chain for every single request.
// For example, this is the right place for a logger or error management interceptor.
func (s *Server) Use(interceptors ...grpc.UnaryServerInterceptor) {
	finalSize := len(s.unaryInterceptors) + len(interceptors)
	if finalSize >= metadata.MaxInterceptors {
		panic("waterdrop: server use too many interceptors")
	}

	mergedInterceptors := make([]grpc.UnaryServerInterceptor, finalSize)
	copy(mergedInterceptors, s.unaryInterceptors)
	copy(mergedInterceptors[len(s.unaryInterceptors):], interceptors)

	s.unaryInterceptors = mergedInterceptors
}
