package metric

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func RegisterMetric(engine *gin.Engine) {
	metric := engine.Group("/metrics")
	{
		metric.GET("/", metricHandler(promhttp.Handler().ServeHTTP))
	}
}

func metricHandler(handler http.HandlerFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		handler.ServeHTTP(ctx.Writer, ctx.Request)
	}
}
