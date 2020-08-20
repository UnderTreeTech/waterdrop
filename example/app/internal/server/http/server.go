package http

import (
	"fmt"
	"net"

	ip "github.com/UnderTreeTech/waterdrop/utils/net"

	"github.com/UnderTreeTech/waterdrop/pkg/registry"

	"github.com/UnderTreeTech/waterdrop/pkg/server/http"
)

type ServerInfo struct {
	Server      *http.Server
	ServiceInfo *registry.ServiceInfo
}

func New() *ServerInfo {
	server := http.New("Server.HTTP")

	router(server)
	middlewares(server)

	addr := server.Start()
	_, port, _ := net.SplitHostPort(addr.String())
	serviceInfo := &registry.ServiceInfo{
		Name:    "server.http.example",
		Scheme:  "http",
		Addr:    fmt.Sprintf("%s://%s:%s", "http", ip.InternalIP(), port),
		Version: "1.0.0",
	}

	return &ServerInfo{Server: server, ServiceInfo: serviceInfo}
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
