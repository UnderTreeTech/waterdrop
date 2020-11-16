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
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/assert/v2"

	"github.com/UnderTreeTech/waterdrop/pkg/trace"

	"github.com/UnderTreeTech/waterdrop/pkg/server/http/config"
	"github.com/gin-gonic/gin"

	"github.com/opentracing/opentracing-go"
	jconfig "github.com/uber/jaeger-client-go/config"
)

func newJaegerClient() (opentracing.Tracer, func()) {
	var configuration = jconfig.Configuration{
		ServiceName: "trace",
	}

	tracer, closer, err := configuration.NewTracer()
	if err != nil {
		panic(fmt.Sprintf("new jaeger trace fail, err msg %s", err.Error()))
	}

	return tracer, func() { closer.Close() }
}

func TestTrace(t *testing.T) {
	engine := gin.New()

	engine.Use(Trace(config.DefaultServerConfig()))
	engine.GET("/trace/mock", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, trace.TraceID(ctx.Request.Context()))
	})

	req := httptest.NewRequest(http.MethodGet, "/trace/mock", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, w.Body.String(), "")
}

func TestJaegerTrace(t *testing.T) {
	engine := gin.New()
	engine.Use(Trace(config.DefaultServerConfig()))
	engine.GET("/trace/jaeger", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, trace.TraceID(ctx.Request.Context()))
	})

	tracer, close := newJaegerClient()
	trace.SetGlobalTracer(tracer)
	defer close()

	req := httptest.NewRequest(http.MethodGet, "/trace/jaeger", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.NotEqual(t, 0, len(w.Body.String()))
}
