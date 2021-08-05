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
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/UnderTreeTech/waterdrop/examples/proto/demo"

	"github.com/UnderTreeTech/waterdrop/pkg/server/rpc/config"

	"github.com/UnderTreeTech/waterdrop/pkg/server/rpc/server"

	"github.com/UnderTreeTech/waterdrop/pkg/stats"

	"google.golang.org/grpc/resolver"

	"github.com/UnderTreeTech/waterdrop/pkg/registry/etcd"

	"github.com/UnderTreeTech/waterdrop/pkg/log"

	"github.com/UnderTreeTech/waterdrop/pkg/utils/xnet"

	"github.com/UnderTreeTech/waterdrop/pkg/conf"

	"google.golang.org/grpc"

	"github.com/UnderTreeTech/waterdrop/pkg/registry"
)

func main() {
	flag.Parse()
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)

	conf.Init()
	defer log.New(nil).Sync()

	srvConfig := &config.ServerConfig{}
	parseConfig("server.rpc", srvConfig)
	if srvConfig.WatchConfig {
		conf.OnChange(func(config *conf.Config) {
			parseConfig("server.rpc", srvConfig)
		})
	}

	server := server.New(srvConfig)
	registerServers(server.Server(), &Service{})

	addr := server.Start()
	_, port, _ := net.SplitHostPort(addr.String())
	serviceInfo := &registry.ServiceInfo{
		Name:    "service.user.v1",
		Scheme:  "grpc",
		Addr:    fmt.Sprintf("%s://%s:%s", "grpc", xnet.InternalIP(), port),
		Version: "1.0.0",
	}

	etcdConf := &etcd.Config{}
	if err := conf.Unmarshal("etcd", etcdConf); err != nil {
		panic(fmt.Sprintf("unmarshal etcd config fail, err msg %s", err.Error()))
	}
	etcd := etcd.New(etcdConf)
	etcd.Register(context.Background(), serviceInfo)
	resolver.Register(etcd)
	startStats()

	<-c

	etcd.Close()
	server.Stop(context.Background())
}

func registerServers(g *grpc.Server, s *Service) {
	demo.RegisterDemoServer(g, s)
}

func parseConfig(configName string, srvConfig *config.ServerConfig) {
	if err := conf.Unmarshal(configName, srvConfig); err != nil {
		panic(fmt.Sprintf("unmarshal grpc server config fail, err msg %s", err.Error()))
	}
}

func startStats() {
	statsConfig := &stats.StatsConfig{}
	if err := conf.Unmarshal("stats", statsConfig); err != nil {
		panic(fmt.Sprintf("unmarshal stats config fail, err msg %s", err.Error()))
	}
	stats.StartStats(statsConfig)
}

type Service struct{}

func (s *Service) SayHello(ctx context.Context, req *demo.HelloReq) (reply *emptypb.Empty, err error) {
	reply = &emptypb.Empty{}
	return reply, nil
}
func (s *Service) SayHelloURL(ctx context.Context, req *demo.HelloReq) (reply *demo.HelloResp, err error) {
	reply = &demo.HelloResp{Content: "Hello " + req.Name}
	return reply, nil
}
