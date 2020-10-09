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
