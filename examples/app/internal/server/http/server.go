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

package http

import (
	"fmt"
	"net"

	"github.com/UnderTreeTech/waterdrop/pkg/server/http/config"

	"github.com/UnderTreeTech/waterdrop/pkg/server/http/server"

	"github.com/UnderTreeTech/waterdrop/pkg/server/http/middlewares"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"

	"github.com/UnderTreeTech/waterdrop/examples/app/internal/dao"

	"github.com/UnderTreeTech/waterdrop/pkg/conf"

	"github.com/UnderTreeTech/waterdrop/pkg/utils/xnet"

	"github.com/UnderTreeTech/waterdrop/pkg/registry"
)

type ServerInfo struct {
	Server      *server.Server
	ServiceInfo *registry.ServiceInfo
}

func New() *ServerInfo {
	srvConfig := &config.ServerConfig{}
	parseConfig("server.http", srvConfig)
	if srvConfig.WatchConfig {
		conf.OnChange(func(config *conf.Config) {
			parseConfig("server.http", srvConfig)
		})
	}

	server := server.New(srvConfig)

	registerMiddlewares(server)
	router(server)

	addr := server.Start()
	_, port, _ := net.SplitHostPort(addr.String())
	serviceInfo := &registry.ServiceInfo{
		Name:    "server.http.example",
		Scheme:  "http",
		Addr:    fmt.Sprintf("%s://%s:%s", "http", xnet.InternalIP(), port),
		Version: "1.0.0",
	}
	binding.Validator.Engine().(*validator.Validate).SetTagName("validate")
	return &ServerInfo{Server: server, ServiceInfo: serviceInfo}
}

func parseConfig(configName string, srvConfig *config.ServerConfig) {
	if err := conf.Unmarshal(configName, srvConfig); err != nil {
		panic(fmt.Sprintf("unmarshal http server config fail, err msg %s", err.Error()))
	}
}

func registerMiddlewares(s *server.Server) {
	//jwt token middleware
	//s.Use(jwt.JWT())
	s.Use(middlewares.Header())

	signClientConfig := &config.ClientConfig{}
	if err := conf.Unmarshal("client.http.app", signClientConfig); err != nil {
		panic(fmt.Sprintf("unmarshal signature client config fail, err msg %s", err.Error()))
	}
	signVerify := middlewares.NewSignatureVerify(signClientConfig, dao.NewRedis())
	s.Use(signVerify.Signature())
}

func router(s *server.Server) {
	g := s.Group("/api")
	{
		g.GET("/app/secrets", getAppInfo)
		g.GET("/app/skips", getSkipUrls)
		g.POST("/app/validate/:id", validateApp)
	}
}
