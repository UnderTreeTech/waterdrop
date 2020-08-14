package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/UnderTreeTech/waterdrop/example/app/internal/server"

	"github.com/UnderTreeTech/waterdrop/example/app/internal/dao"
	"github.com/UnderTreeTech/waterdrop/pkg/conf"

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
	srv.Start()

	//rpcConf := &rpc.Config{}
	//if err := conf.Unmarshal("Server.RPC", rpcConf); err != nil {
	//	panic(fmt.Sprintf("unmarshal grpc server config fail, err msg %s", err.Error()))
	//}

	//rpcSrv := rpc.New(rpcConf)
	//demo.RegisterDemoServer(rpcSrv.Server(), &service.Service{})
	//rpcSrv.Start()
	<-c

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Stop(ctx)
	//rpcSrv.Stop()
}
