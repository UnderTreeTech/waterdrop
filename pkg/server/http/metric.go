package http

import (
	"strconv"
	"time"

	"github.com/UnderTreeTech/waterdrop/pkg/stats/metric"
	"github.com/gin-gonic/gin"
)

func (s *Server) Metric() gin.HandlerFunc {
	return func(c *gin.Context) {
		now := time.Now()
		c.Next()

		appkey := c.Request.Header.Get("appkey")
		metric.HTTPServerHandleCounter.Inc(c.FullPath(), c.Request.Method, appkey, strconv.Itoa(c.Writer.Status()))
		metric.HTTPServerReqDuration.Observe(time.Since(now).Seconds(), c.FullPath(), c.Request.Method, appkey)
	}
}
