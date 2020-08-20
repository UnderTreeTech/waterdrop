package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/UnderTreeTech/waterdrop/example/app/internal/server/grpc"

	"github.com/UnderTreeTech/waterdrop/example/app/internal/server/http"

	"google.golang.org/grpc/resolver"

	"github.com/UnderTreeTech/waterdrop/pkg/registry/etcd"

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

	http := http.New()
	rpc := grpc.New()

	etcdConf := &etcd.Config{}
	if err := conf.Unmarshal("Etcd", etcdConf); err != nil {
		panic(fmt.Sprintf("unmarshal etcd config fail, err msg %s", err.Error()))
	}
	etcd := etcd.New(etcdConf)
	etcd.Register(context.Background(), rpc.ServiceInfo)
	etcd.Register(context.Background(), http.ServiceInfo)
	resolver.Register(etcd)

	<-c

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	etcd.Close()
	http.Server.Stop(ctx)
	rpc.Server.Stop()
}
