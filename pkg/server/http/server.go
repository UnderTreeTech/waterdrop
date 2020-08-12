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
}

func defaultConfig() *Config {
	return &Config{
		Addr:    "0.0.0.0:9090",
		Mode:    gin.ReleaseMode,
		Timeout: time.Millisecond * 1000,
	}
}

func srvConfig(name string) *Config {
	serverConf := defaultConfig()
	if name != "" {
		err := conf.Unmarshal(name, serverConf)
		if err != nil {
			panic(fmt.Sprintf("reload server.http fail, err msg %s", err.Error()))
		}

		log.Printf("reload http server config, %+v", serverConf)

		conf.OnChange(func(config *conf.Config) {
			err := config.Unmarshal(name, serverConf)
			if err != nil {
				log.Printf("reload server.http fail, err msg %s", err.Error())
			}
		})
	}

	return serverConf
}

type Server struct {
	*gin.Engine
	Server *http.Server
	config *Config
}

func NewServer(confName string) *Server {
	config := srvConfig(confName)

	gin.SetMode(config.Mode)
	engine := gin.New()
	engine.Use(Trace(config), Recovery(), Logger())

	return &Server{
		Engine: engine,
		config: config,
	}
}

//start server
func (s *Server) Start() {
	listener, err := net.Listen("tcp", s.config.Addr)
	if err != nil {
		panic(fmt.Sprintf("listen tcp fail,err msg %s", err.Error()))
	}

	s.Server = &http.Server{
		Addr:    s.config.Addr,
		Handler: s,
	}

	go func() {
		if err := s.Server.Serve(listener); err != nil {
			if err == http.ErrServerClosed {
				log.Printf("waterdrop: server closed")
				return
			}
			panic(fmt.Sprintf("Server serve fail,err msg %s", err.Error()))
		}
	}()

	log.Printf("HTTP Server:start http listen addr: %s", s.config.Addr)
}

// shutdown server graceful
func (s *Server) Stop(ctx context.Context) error {
	return s.Server.Shutdown(ctx)
}
