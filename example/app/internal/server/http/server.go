package http

import (
	"fmt"
	"net"

	"github.com/UnderTreeTech/waterdrop/pkg/conf"

	"github.com/UnderTreeTech/waterdrop/utils/xnet"

	"github.com/UnderTreeTech/waterdrop/pkg/registry"

	"github.com/UnderTreeTech/waterdrop/pkg/server/http"
)

type ServerInfo struct {
	Server      *http.Server
	ServiceInfo *registry.ServiceInfo
}

func New() *ServerInfo {
	srvConfig := &http.ServerConfig{}
	parseConfig("server.http", srvConfig)
	if srvConfig.WatchConfig {
		conf.OnChange(func(config *conf.Config) {
			parseConfig("server.http", srvConfig)
		})
	}

	server := http.NewServer(srvConfig)

	router(server)
	middlewares(server)

	addr := server.Start()
	_, port, _ := net.SplitHostPort(addr.String())
	serviceInfo := &registry.ServiceInfo{
		Name:    "server.http.example",
		Scheme:  "http",
		Addr:    fmt.Sprintf("%s://%s:%s", "http", xnet.InternalIP(), port),
		Version: "1.0.0",
	}

	return &ServerInfo{Server: server, ServiceInfo: serviceInfo}
}

func parseConfig(configName string, srvConfig *http.ServerConfig) {
	if err := conf.Unmarshal(configName, srvConfig); err != nil {
		panic(fmt.Sprintf("unmarshal http server config fail, err msg %s", err.Error()))
	}
}

func middlewares(s *http.Server) {
	//jwt token middleware
	//s.Use(jwt.JWT())
}

func router(s *http.Server) {
	g := s.Group("/api")
	{
		g.GET("/app/secrets", getAppInfo)
		g.GET("/app/skips", getSkipUrls)
		g.POST("/app/validate/:id", validateApp)
	}
}
