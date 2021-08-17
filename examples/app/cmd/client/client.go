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

package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	rpcConfig "github.com/UnderTreeTech/waterdrop/pkg/server/rpc/config"

	rpcClient "github.com/UnderTreeTech/waterdrop/pkg/server/rpc/client"

	"github.com/UnderTreeTech/waterdrop/pkg/server/rpc/interceptors"

	"github.com/UnderTreeTech/waterdrop/examples/proto/demo"
	"github.com/UnderTreeTech/waterdrop/pkg/log"

	"github.com/UnderTreeTech/waterdrop/pkg/trace/jaeger"

	"google.golang.org/grpc/resolver"

	"github.com/UnderTreeTech/waterdrop/pkg/registry/etcd"

	_ "github.com/UnderTreeTech/waterdrop/examples/app/internal/ecode"
	"github.com/UnderTreeTech/waterdrop/pkg/conf"
)

func main() {
	flag.Parse()

	conf.Init()
	defer log.New(nil).Sync()
	defer jaeger.Init()()

	etcdConf := &etcd.Config{}
	if err := conf.Unmarshal("etcd", etcdConf); err != nil {
		panic(fmt.Sprintf("unmarshal etcd config fail, err msg %s", err.Error()))
	}
	etcd := etcd.New(etcdConf)
	resolver.Register(etcd)

	cliConf := &rpcConfig.ClientConfig{}
	if err := conf.Unmarshal("client.rpc.demo", cliConf); err != nil {
		panic(fmt.Sprintf("unmarshal demo client config fail, err msg %s", err.Error()))
	}
	fmt.Println(cliConf)
	rpcCli := rpcClient.New(cliConf)
	rpcCli.Use(interceptors.GoogleSREBreaker(rpcCli.GetBreakers()))
	demoRPC := demo.NewDemoClient(rpcCli.GetConn())

	for i := 0; i < 1; i++ {
		_, err := demoRPC.SayHelloURL(context.Background(), &demo.HelloReq{Name: "John Sun"})
		if err != nil {
			fmt.Println("request error", err)
		}
	}

	time.Sleep(time.Hour * 30)
}
