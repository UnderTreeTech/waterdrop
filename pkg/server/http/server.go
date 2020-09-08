package http

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type ServerConfig struct {
	Addr string

	Timeout time.Duration
	Mode    string

	SlowRequestDuration time.Duration
	WatchConfig         bool
}

func defaultServerConfig() *ServerConfig {
	return &ServerConfig{
		Addr:                "0.0.0.0:9000",
		Mode:                gin.ReleaseMode,
		Timeout:             time.Millisecond * 1000,
		SlowRequestDuration: 500 * time.Millisecond,
	}
}

type Server struct {
	*gin.Engine
	Server *http.Server
	config *ServerConfig
}

func NewServer(config *ServerConfig) *Server {
	if config == nil {
		config = defaultServerConfig()
	}

	gin.SetMode(config.Mode)
	srv := &Server{
		Engine: gin.New(),
		config: config,
	}

	srv.Use(srv.recovery(), srv.trace(), srv.logger())

	return srv
}

//start server
func (s *Server) Start() net.Addr {
	listener, err := net.Listen("tcp", s.config.Addr)
	if err != nil {
		panic(fmt.Sprintf("http server: listen tcp fail,err msg %s", err.Error()))
	}

	s.Server = &http.Server{
		Addr:    s.config.Addr,
		Handler: s,
	}

	go func() {
		if err := s.Server.Serve(listener); err != nil {
			if err == http.ErrServerClosed {
				log.Printf("waterdrop: http server closed")
				return
			}
			panic(fmt.Sprintf("HTTP Server serve fail,err msg %s", err.Error()))
		}
	}()

	log.Printf("HTTP Server: start http listen addr: %s", listener.Addr().String())
	return listener.Addr()
}

// shutdown server graceful
func (s *Server) Stop(ctx context.Context) error {
	return s.Server.Shutdown(ctx)
}
