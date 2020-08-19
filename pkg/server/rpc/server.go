package rpc

import (
	"context"
	"fmt"
	"log"
	"math"
	"net"
	"time"

	"google.golang.org/grpc/reflection"

	"google.golang.org/grpc/keepalive"

	"google.golang.org/grpc"
)

// borrow from gin
var _abortIndex int8 = math.MaxInt8 / 2

type Config struct {
	Network string
	Addr    string

	Timeout        time.Duration
	IdleTimeout    time.Duration
	MaxLifeTime    time.Duration
	ForceCloseWait time.Duration

	KeepAliveInterval time.Duration
	KeepAliveTimeout  time.Duration

	SlowRequestDuration time.Duration
}

type Server struct {
	server *grpc.Server
	config *Config

	serverOptions     []grpc.ServerOption
	unaryInterceptors []grpc.UnaryServerInterceptor
}

func defaultConfig() *Config {
	return &Config{
		Network:             "tcp",
		Addr:                "0.0.0.0:20812",
		Timeout:             time.Second,
		IdleTimeout:         180 * time.Second,
		MaxLifeTime:         2 * time.Hour,
		ForceCloseWait:      20 * time.Second,
		KeepAliveInterval:   60 * time.Second,
		KeepAliveTimeout:    20 * time.Second,
		SlowRequestDuration: 500 * time.Millisecond,
	}
}

func New(config *Config) *Server {
	if config == nil {
		config = defaultConfig()
	}

	srv := &Server{
		config:            config,
		serverOptions:     make([]grpc.ServerOption, 0),
		unaryInterceptors: make([]grpc.UnaryServerInterceptor, 0),
	}

	keepaliveOpts := grpc.KeepaliveParams(keepalive.ServerParameters{
		MaxConnectionIdle:     config.IdleTimeout,
		MaxConnectionAgeGrace: config.ForceCloseWait,
		Time:                  config.KeepAliveInterval,
		Timeout:               config.KeepAliveTimeout,
		MaxConnectionAge:      config.MaxLifeTime,
	})

	srv.Use(srv.recovery(), srv.trace(), srv.logger(), srv.validate(), srv.validate())
	unaryOpts := srv.WithUnaryServerChain(srv.unaryInterceptors...)

	srv.serverOptions = append(srv.serverOptions, keepaliveOpts, unaryOpts)
	srv.server = grpc.NewServer(srv.serverOptions...)

	return srv
}

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

// GracefulStop stops the gRPC server gracefully. It stops the server from
// accepting new connections and RPCs and blocks until all the pending RPCs are
// finished.
func (s *Server) Stop() {
	s.server.GracefulStop()
}

func (s *Server) Server() *grpc.Server {
	return s.server
}

// ChainUnaryServer creates a single interceptor out of a chain of many interceptors.
//
// Execution is done in left-to-right order, including passing of context.
// For example ChainUnaryServer(one, two, three) will execute one before two before three, and three
// will see context changes of one and two.
func (s *Server) ChainUnaryServer(interceptors ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	n := len(interceptors)

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
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
//
// WithUnaryServerChain is a grpc.Server config option that accepts multiple unary interceptors.
// Basically syntactic sugar.
func (s *Server) WithUnaryServerChain(interceptors ...grpc.UnaryServerInterceptor) grpc.ServerOption {
	return grpc.UnaryInterceptor(s.ChainUnaryServer(interceptors...))
}

// Use attaches a global interceptor to the server. ie. the interceptor attached through Use() will be
// included in the interceptors chain for every single request.
// For example, this is the right place for a logger or error management interceptor.
func (s *Server) Use(interceptors ...grpc.UnaryServerInterceptor) {
	finalSize := len(s.unaryInterceptors) + len(interceptors)
	if finalSize >= int(_abortIndex) {
		panic("waterdrop: server use too many interceptors")
	}

	mergedInterceptors := make([]grpc.UnaryServerInterceptor, finalSize)
	copy(mergedInterceptors, s.unaryInterceptors)
	copy(mergedInterceptors[len(s.unaryInterceptors):], interceptors)

	s.unaryInterceptors = mergedInterceptors
}
