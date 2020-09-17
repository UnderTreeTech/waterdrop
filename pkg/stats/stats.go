package stats

import (
	"fmt"

	"github.com/UnderTreeTech/waterdrop/pkg/stats/metric"
	"github.com/UnderTreeTech/waterdrop/pkg/stats/profile"
	"github.com/gin-gonic/gin"
)

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
