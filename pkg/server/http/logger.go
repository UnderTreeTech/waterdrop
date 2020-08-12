package http

import (
	"time"

	"github.com/UnderTreeTech/waterdrop/pkg/log"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		now := time.Now()
		var quota float64
		if deadline, ok := c.Request.Context().Deadline(); ok {
			quota = time.Until(deadline).Seconds()
		}

		c.Next()

		err := c.Errors
		dt := time.Since(now)

		log.Info(c.Request.Context(),
			"http-access-log",
			log.String("ip", c.ClientIP()),
			log.String("method", c.Request.Method),
			log.String("path", c.Request.URL.Path),
			log.Any("headers", c.Request.Header),
			log.String("params", c.Request.Form.Encode()),
			//log.Bytes("body", "body"),
			//log.Int("ret", cerr.Code()),
			//log.String("msg", cerr.Message()),
			//log.String("stack", fmt.Sprintf("%+v", err)),
			log.String("err", err.String()),
			log.Float64("quota", quota),
			log.Float64("exec_time", dt.Seconds()),
		)
	}
}
