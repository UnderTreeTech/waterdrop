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
	"net/http"

	"github.com/UnderTreeTech/waterdrop/pkg/server/http/config"

	"github.com/UnderTreeTech/waterdrop/pkg/server/http/websocket"

	"github.com/UnderTreeTech/waterdrop/pkg/server/http/middlewares"

	"github.com/gin-gonic/gin"
)

type Server struct {
	*gin.Engine
	Server *http.Server
	config *config.ServerConfig
}

func New(cfg *config.ServerConfig) *Server {
	if cfg == nil {
		cfg = config.DefaultServerConfig()
	}

	gin.SetMode(cfg.Mode)
	srv := &Server{
		Engine: gin.New(),
		config: cfg,
	}

	srv.Use(middlewares.Recovery(), middlewares.Trace(srv.config), middlewares.Logger(srv.config))
	if cfg.EnableMetric {
		srv.Use(middlewares.Metric())
	}

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

// upgrade http to websocket
func (s *Server) Upgrade(ws *websocket.WebSocket) gin.IRoutes {
	return s.GET(ws.Path, func(c *gin.Context) {
		ws.Upgrade(c.Writer, c.Request)
	})
}
