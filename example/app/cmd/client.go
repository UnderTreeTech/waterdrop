package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/UnderTreeTech/waterdrop/pkg/trace/jaeger"

	"github.com/UnderTreeTech/waterdrop/pkg/server/rpc"

	"github.com/UnderTreeTech/waterdrop/pkg/status"

	"github.com/UnderTreeTech/protobuf/demo"

	"google.golang.org/grpc/resolver"

	"github.com/UnderTreeTech/waterdrop/pkg/registry/etcd"

	_ "github.com/UnderTreeTech/waterdrop/example/app/internal/ecode"
	"github.com/UnderTreeTech/waterdrop/pkg/conf"
	"github.com/UnderTreeTech/waterdrop/pkg/log"
)

func main() {
	flag.Parse()

	conf.Init()
	defer log.Init("Log")()
	defer jaeger.Init("Trace.Jaeger")()

	etcdConf := &etcd.Config{}
	if err := conf.Unmarshal("Etcd", etcdConf); err != nil {
		panic(fmt.Sprintf("unmarshal etcd config fail, err msg %s", err.Error()))
	}
	etcd := etcd.New(etcdConf)
	resolver.Register(etcd)

	cliConf := &rpc.ClientConfig{}
	if err := conf.Unmarshal("Client.RPC.Stardust", cliConf); err != nil {
		panic(fmt.Sprintf("unmarshal demo client config fail, err msg %s", err.Error()))
	}
	fmt.Println(cliConf)
	client := demo.NewDemoClient(rpc.NewClient(cliConf))
	now := time.Now()
	for i := 0; i < 1; i++ {
		_, err := client.SayHelloURL(context.Background(), &demo.HelloReq{Name: "John Sun"})
		if err != nil {
			fmt.Println("err", status.ExtractStatus(err))
		}
		//fmt.Println(reply)
	}
	fmt.Println(time.Since(now))
	time.Sleep(time.Hour * 30)
}
