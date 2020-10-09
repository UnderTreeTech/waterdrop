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

package profile

import (
	"net/http"
	"net/http/pprof"

	"github.com/gin-gonic/gin"
)

func RegisterProfile(engine *gin.Engine) {
	perf := engine.Group("/debug/profile")
	{
		perf.GET("/", profileHandler(pprof.Index))
		perf.GET("/cmdline", profileHandler(pprof.Cmdline))
		perf.GET("/profile", profileHandler(pprof.Profile))
		perf.GET("/symbol", profileHandler(pprof.Symbol))
		perf.GET("/trace", profileHandler(pprof.Trace))
		perf.GET("/allocs", profileHandler(pprof.Handler("allocs").ServeHTTP))
		perf.GET("/block", profileHandler(pprof.Handler("block").ServeHTTP))
		perf.GET("/goroutine", profileHandler(pprof.Handler("goroutine").ServeHTTP))
		perf.GET("/heap", profileHandler(pprof.Handler("heap").ServeHTTP))
		perf.GET("/mutex", profileHandler(pprof.Handler("mutex").ServeHTTP))
		perf.GET("/threadcreate", profileHandler(pprof.Handler("threadcreate").ServeHTTP))
	}
}

func profileHandler(handler http.HandlerFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		handler.ServeHTTP(ctx.Writer, ctx.Request)
	}
}
