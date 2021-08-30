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

package middlewares

import (
	"time"

	"github.com/UnderTreeTech/waterdrop/pkg/server/http/config"

	"github.com/UnderTreeTech/waterdrop/pkg/log"

	"github.com/gin-gonic/gin"
)

// Logger log request details
func Logger(config *config.ServerConfig) gin.HandlerFunc {
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

		if duration >= config.SlowRequestDuration {
			log.Warn(c.Request.Context(), "http-slow-access-log", fields...)
		} else {
			log.Info(c.Request.Context(), "http-access-log", fields...)
		}
	}
}
