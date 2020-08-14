package http

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/UnderTreeTech/waterdrop/pkg/conf"

	"github.com/gin-gonic/gin"
)

type Config struct {
	Addr string `conf:"addr"`

	Timeout time.Duration `conf:"timeout"`
	Mode    string        `conf:"mode"`

	SlowRequestTimeout time.Duration
}

func defaultConfig() *Config {
	return &Config{
		Addr:               "0.0.0.0:9090",
		Mode:               gin.ReleaseMode,
		Timeout:            time.Millisecond * 1000,
		SlowRequestTimeout: 500 * time.Millisecond,
	}
}

func srvConfig(name string) *Config {
	config := defaultConfig()

	if err := conf.Unmarshal(name, config); err != nil {
		panic(fmt.Sprintf("reload server.http fail, err msg %s", err.Error()))
	}

	log.Printf("reload http server config, %+v", config)

	conf.OnChange(func(config *conf.Config) {
		err := config.Unmarshal(name, config)
		if err != nil {
			log.Printf("reload server.http fail, err msg %s", err.Error())
		}
	})

	return config
}

type Server struct {
	*gin.Engine
	Server *http.Server
	config *Config
}

func New(confName string) *Server {
	config := srvConfig(confName)
	srv := &Server{
		Engine: gin.New(),
		config: config,
	}

	gin.SetMode(config.Mode)
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

	log.Printf("HTTP Server: start http listen addr: %s", s.config.Addr)
	return listener.Addr()
}

// shutdown server graceful
func (s *Server) Stop(ctx context.Context) error {
	return s.Server.Shutdown(ctx)
}
