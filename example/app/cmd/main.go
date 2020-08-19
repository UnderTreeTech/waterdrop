package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	net2 "github.com/UnderTreeTech/waterdrop/utils/net"

	"google.golang.org/grpc/resolver"

	"github.com/UnderTreeTech/waterdrop/pkg/registry"
	"github.com/UnderTreeTech/waterdrop/pkg/registry/etcd"

	"github.com/UnderTreeTech/protobuf/demo"
	"github.com/UnderTreeTech/waterdrop/example/app/internal/service"

	"github.com/UnderTreeTech/waterdrop/example/app/internal/server"

	"github.com/UnderTreeTech/waterdrop/example/app/internal/dao"
	"github.com/UnderTreeTech/waterdrop/pkg/conf"
	"github.com/UnderTreeTech/waterdrop/pkg/server/rpc"

	"github.com/UnderTreeTech/waterdrop/pkg/trace/jaeger"

	"github.com/UnderTreeTech/waterdrop/pkg/log"
)

func main() {
	flag.Parse()
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)

	conf.Init()
	defer log.Init("Log")()
	defer jaeger.Init("Trace.Jaeger")()
	defer dao.NewDao().Close()

	srv := server.NewHTTPServer()
	httpAddr := srv.Start()

	rpcConf := &rpc.Config{}
	if err := conf.Unmarshal("Server.RPC", rpcConf); err != nil {
		panic(fmt.Sprintf("unmarshal grpc server config fail, err msg %s", err.Error()))
	}

	rpcSrv := rpc.New(rpcConf)
	demo.RegisterDemoServer(rpcSrv.Server(), &service.Service{})
	addr := rpcSrv.Start()

	_, port, _ := net.SplitHostPort(addr.String())
	etcdConf := &etcd.Config{}
	if err := conf.Unmarshal("Etcd", etcdConf); err != nil {
		panic(fmt.Sprintf("unmarshal etcd config fail, err msg %s", err.Error()))
	}
	etcd := etcd.New(etcdConf)
	serviceInfo := &registry.ServiceInfo{
		Name:    "service.user.v1",
		Scheme:  "grpc",
		Addr:    fmt.Sprintf("%s://%s:%s", "grpc", net2.InternalIP(), port),
		Version: "1.0.0",
	}

	_, httpPort, _ := net.SplitHostPort(httpAddr.String())
	httpInfo := &registry.ServiceInfo{
		Name:    "waterdrop.example.http",
		Scheme:  "http",
		Addr:    fmt.Sprintf("%s://%s:%s", "http", net2.InternalIP(), httpPort),
		Version: "1.0.0",
	}

	etcd.Register(context.Background(), serviceInfo)
	etcd.Register(context.Background(), httpInfo)
	resolver.Register(etcd)

	<-c

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	etcd.Close()
	srv.Stop(ctx)
	rpcSrv.Stop()
}
