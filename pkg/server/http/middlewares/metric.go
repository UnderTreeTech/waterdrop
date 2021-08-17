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
	"strconv"
	"time"

	"github.com/UnderTreeTech/waterdrop/pkg/stats/metric"
	"github.com/gin-gonic/gin"
)

var namespace = "appkey"

const _defaultNamespace = "default"

// Metric http request metric middleware
func Metric() gin.HandlerFunc {
	return func(c *gin.Context) {
		now := time.Now()

		c.Next()

		ns := c.Request.Header.Get(namespace)
		if ns == "" {
			ns = _defaultNamespace
		}
		metric.HTTPServerHandleCounter.Inc(c.FullPath(), c.Request.Method, ns, strconv.Itoa(c.Writer.Status()))
		metric.HTTPServerReqDuration.Observe(time.Since(now).Seconds(), c.FullPath(), c.Request.Method, ns)
	}
}

// SetHttpMetricNamespace set http metric namespace get from which http header
// default namespace is get from http header: appkey
func SetMetricNamespace(header string) {
	namespace = header
}
