package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/UnderTreeTech/protobuf/demo"

	"google.golang.org/grpc/resolver"

	"github.com/UnderTreeTech/waterdrop/pkg/registry/etcd"

	"github.com/UnderTreeTech/waterdrop/pkg/conf"
	"github.com/UnderTreeTech/waterdrop/pkg/log"
	"google.golang.org/grpc"
)

func main() {
	flag.Parse()

	conf.Init()
	defer log.Init("Log")()

	etcdConf := &etcd.Config{}
	if err := conf.Unmarshal("Etcd", etcdConf); err != nil {
		panic(fmt.Sprintf("unmarshal etcd config fail, err msg %s", err.Error()))
	}
	etcd := etcd.New(etcdConf)
	resolver.Register(etcd)

	client := demo.NewDemoClient(newClient())
	now := time.Now()
	for i := 0; i < 1; i++ {
		reply, err := client.SayHelloURL(context.Background(), &demo.HelloReq{Name: "John Sun"})
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(reply)
	}
	fmt.Println(time.Since(now))
	time.Sleep(time.Hour * 30)
}

func newClient() *grpc.ClientConn {
	cc, err := grpc.DialContext(
		context.Background(),
		"etcd://default/service.user.v1",
		grpc.WithBlock(),
		grpc.WithInsecure(),
	)
	if err != nil {
		panic(err)
	}

	return cc
}
