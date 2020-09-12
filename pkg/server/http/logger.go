package http

import (
	"time"

	"github.com/UnderTreeTech/waterdrop/pkg/stats/metric"
	"github.com/UnderTreeTech/waterdrop/pkg/status"

	"github.com/UnderTreeTech/waterdrop/pkg/log"

	"github.com/gin-gonic/gin"
)

func (s *Server) logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		now := time.Now()
		var quota float64
		if deadline, ok := c.Request.Context().Deadline(); ok {
			quota = time.Until(deadline).Seconds()
		}

		c.Next()

		duration := time.Since(now)
		estatus := status.OK
		if len(c.Errors) > 0 {
			estatus = status.ExtractStatus(c.Errors.Last().Err)
		}

		metric.HTTPServerHandleCounter.Inc(c.Request.URL.Path, c.Request.Method, c.ClientIP(), estatus.Error())
		metric.HTTPServerReqDuration.Observe(duration.Seconds(), c.Request.URL.Path, c.Request.Method, c.ClientIP())

		fields := make([]log.Field, 0, 9)
		fields = append(
			fields,
			log.String("client_ip", c.ClientIP()),
			log.String("method", c.Request.Method),
			log.String("path", c.Request.URL.Path),
			log.Any("headers", c.Request.Header),
			log.String("req", c.Request.Form.Encode()),
			log.Float64("quota", quota),
			log.Float64("duration", duration.Seconds()),
			log.Int("code", estatus.Code()),
			log.String("error", estatus.Message()),
		)

		if duration >= s.config.SlowRequestDuration {
			log.Warn(c.Request.Context(), "http-slow-access-log", fields...)
		} else {
			log.Info(c.Request.Context(), "http-access-log", fields...)
		}
	}
}
