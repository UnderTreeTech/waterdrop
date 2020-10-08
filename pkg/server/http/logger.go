package http

import (
	"time"

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

		fields := make([]log.Field, 0, 10)
		fields = append(
			fields,
			log.String("client_ip", c.ClientIP()),
			log.String("method", c.Request.Method),
			log.String("path", c.Request.URL.Path),
			log.Any("headers", c.Request.Header),
			log.String("req", c.Request.Form.Encode()),
			log.Float64("quota", quota),
			log.Float64("duration", duration.Seconds()),
			log.Int("status", c.Writer.Status()),
			log.Int("size", c.Writer.Size()),
			log.String("error", c.Errors.ByType(gin.ErrorTypePrivate).String()),
		)

		if duration >= s.config.SlowRequestDuration {
			log.Warn(c.Request.Context(), "http-slow-access-log", fields...)
		} else {
			log.Info(c.Request.Context(), "http-access-log", fields...)
		}
	}
}
