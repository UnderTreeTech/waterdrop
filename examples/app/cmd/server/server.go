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
	syslog "log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/UnderTreeTech/waterdrop/pkg/stats"

	"github.com/UnderTreeTech/waterdrop/examples/app/internal/server/grpc"

	"github.com/UnderTreeTech/waterdrop/examples/app/internal/server/http"

	"google.golang.org/grpc/resolver"

	"github.com/UnderTreeTech/waterdrop/pkg/registry/etcd"

	"github.com/UnderTreeTech/waterdrop/examples/app/internal/dao"
	"github.com/UnderTreeTech/waterdrop/pkg/conf"

	"github.com/UnderTreeTech/waterdrop/pkg/trace/jaeger"

	"github.com/UnderTreeTech/waterdrop/pkg/log"
)

func main() {
	flag.Parse()
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)

	conf.Init()
	defer initLog().Sync()
	defer jaeger.Init()()
	defer dao.NewDao().Close()

	http := http.New()
	rpc := grpc.New()

	etcdConf := &etcd.Config{}
	if err := conf.Unmarshal("etcd", etcdConf); err != nil {
		panic(fmt.Sprintf("unmarshal etcd config fail, err msg %s", err.Error()))
	}
	etcd := etcd.New(etcdConf)
	etcd.Register(context.Background(), rpc.ServiceInfo)
	etcd.Register(context.Background(), http.ServiceInfo)
	resolver.Register(etcd)
	startStats()

	<-c

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	etcd.Close()
	http.Server.Stop(ctx)
	rpc.Server.Stop(ctx)
}

func initLog() *log.Logger {
	logConfig := &log.Config{}
	if err := conf.Unmarshal("log", logConfig); err != nil {
		syslog.Printf("parse log config fail, err msg %s", err.Error())
	}

	logger := log.New(logConfig)
	conf.OnChange(func(config *conf.Config) {
		if err := conf.Unmarshal("log", logConfig); err != nil {
			syslog.Printf("parse log config fail, err msg %s", err.Error())
		}
		logger.SetLevel(logConfig.Level)
	})

	return logger
}

func startStats() {
	statsConfig := &stats.StatsConfig{}
	if err := conf.Unmarshal("stats", statsConfig); err != nil {
		panic(fmt.Sprintf("unmarshal stats config fail, err msg %s", err.Error()))
	}
	stats.StartStats(statsConfig)
}
