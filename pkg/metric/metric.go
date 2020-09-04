package metric

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func init() {
	//gin.SetMode(gin.ReleaseMode)
	engine := gin.Default()

	metric := engine.Group("/metrics")
	{
		metric.GET("/", metricHandler(promhttp.Handler().ServeHTTP))
	}

	go func() {
		if err := engine.Run("localhost:20829"); err != nil {
			panic(fmt.Sprintf("start profile server fail, error %s", err.Error()))
		}
	}()
}

func metricHandler(handler http.HandlerFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		handler.ServeHTTP(ctx.Writer, ctx.Request)
	}
}
