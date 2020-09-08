package stats

import (
	"fmt"

	"github.com/UnderTreeTech/waterdrop/pkg/stats/metric"
	"github.com/UnderTreeTech/waterdrop/pkg/stats/profile"
	"github.com/gin-gonic/gin"
)

func init() {
	//gin.SetMode(gin.ReleaseMode)
	engine := gin.Default()

	profile.RegisterProfile(engine)
	metric.RegisterMetric(engine)

	go func() {
		if err := engine.Run("localhost:20828"); err != nil {
			panic(fmt.Sprintf("start profile server fail, error %s", err.Error()))
		}
	}()
}
