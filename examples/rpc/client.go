package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/UnderTreeTech/protobuf/demo"
	"github.com/UnderTreeTech/waterdrop/pkg/server/rpc"

	"github.com/UnderTreeTech/waterdrop/pkg/log"

	"github.com/UnderTreeTech/waterdrop/pkg/conf"
	"github.com/UnderTreeTech/waterdrop/pkg/registry/etcd"
	"google.golang.org/grpc/resolver"
)

func main() {
	flag.Parse()

	conf.Init()
	defer log.New(nil).Sync()

	etcdConf := &etcd.Config{}
	if err := conf.Unmarshal("etcd", etcdConf); err != nil {
		panic(fmt.Sprintf("unmarshal etcd config fail, err msg %s", err.Error()))
	}
	etcd := etcd.New(etcdConf)
	resolver.Register(etcd)

	cliConf := &rpc.ClientConfig{}
	if err := conf.Unmarshal("client.rpc.demo", cliConf); err != nil {
		panic(fmt.Sprintf("unmarshal demo client config fail, err msg %s", err.Error()))
	}

	client := demo.NewDemoClient(rpc.NewClient(cliConf))
	for i := 0; i < 100; i++ {
		reply, err := client.SayHelloURL(context.Background(), &demo.HelloReq{Name: "John Sun"})
		if err != nil {
			fmt.Println("error", err)
		}

		fmt.Println("reply", reply)
	}

}
