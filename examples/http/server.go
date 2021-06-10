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
	"time"

	"github.com/UnderTreeTech/waterdrop/pkg/server/http/server"

	"github.com/alibaba/sentinel-golang/core/flow"

	"github.com/UnderTreeTech/waterdrop/pkg/ratelimit/setinel"

	"github.com/UnderTreeTech/waterdrop/pkg/server/http/middlewares/ratelimit/sentinel"

	"github.com/UnderTreeTech/waterdrop/pkg/log"

	"github.com/gin-gonic/gin"
)

func main() {
	defer log.New(nil).Sync()

	config := &setinel.Config{
		AppName: "sentinel-http",
	}
	config.FlowRules = append(
		config.FlowRules,
		&flow.Rule{
			Resource:               "GET:/api/ping",
			Threshold:              1000.0,
			TokenCalculateStrategy: flow.Direct,
			ControlBehavior:        flow.Reject,
			StatIntervalInMs:       1000,
		},
	)
	setinel.InitSentinel(config)

	srv := server.New(nil)
	srv.Use(sentinel.Sentinel())

	g := srv.Group("/api")
	{
		g.GET("/ping", ping)
		g.GET("/waterdrop", waterdrop)
	}

	srv.Start()

	time.Sleep(time.Minute * 5)
	srv.Stop(context.Background())
}

func ping(c *gin.Context) {
	c.JSON(200, "ping")
}

func waterdrop(c *gin.Context) {
	c.JSON(200, "Framwork waterdrop")
}
