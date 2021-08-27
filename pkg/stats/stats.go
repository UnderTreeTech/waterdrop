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

	"github.com/UnderTreeTech/waterdrop/pkg/stats/metric"
	"github.com/UnderTreeTech/waterdrop/pkg/stats/profile"
	"github.com/gin-gonic/gin"
)

// StatsConfig
type StatsConfig struct {
	Addr string
	Mode string

	EnableMetric  bool
	EnableProfile bool
}

func defaultStatsConfig() *StatsConfig {
	return &StatsConfig{
		Addr: "0.0.0.0:20828",
		Mode: "release",

		EnableMetric:  true,
		EnableProfile: true,
	}
}

// StartStats start stats server
func StartStats(config *StatsConfig) {
	if config == nil {
		config = defaultStatsConfig()
	}

	gin.SetMode(config.Mode)
	engine := gin.Default()

	if config.EnableProfile {
		profile.RegisterProfile(engine)
	}

	if config.EnableMetric {
		metric.RegisterMetric(engine)
	}

	go func() {
		if err := engine.Run(config.Addr); err != nil {
			panic(fmt.Sprintf("start profile server fail, error %s", err.Error()))
		}
	}()
}
