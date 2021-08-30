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

package stats

import (
	"fmt"
	"net"

	"github.com/UnderTreeTech/waterdrop/pkg/registry"
	"github.com/UnderTreeTech/waterdrop/pkg/utils/xnet"

	"github.com/gin-gonic/gin"

	"github.com/UnderTreeTech/waterdrop/pkg/stats/metric"
	"github.com/UnderTreeTech/waterdrop/pkg/stats/profile"
)

// StartStats start stats server
func StartStats() (si *registry.ServiceInfo, err error) {
	gin.SetMode("release")
	engine := gin.Default()

	profile.RegisterProfile(engine)
	metric.RegisterMetric(engine)
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, err
	}

	go func() {
		if err = engine.RunListener(listener); err != nil {
			return
		}
	}()

	_, port, _ := net.SplitHostPort(listener.Addr().String())
	si = &registry.ServiceInfo{
		Name:    "server.http.stats",
		Scheme:  "http",
		Addr:    fmt.Sprintf("%s://%s:%s", "http", xnet.InternalIP(), port),
		Version: "1.0.0",
	}
	return
}
