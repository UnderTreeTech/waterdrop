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
	"flag"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/UnderTreeTech/waterdrop/pkg/conf"
	"github.com/UnderTreeTech/waterdrop/pkg/trace"

	"github.com/UnderTreeTech/waterdrop/pkg/server/http/config"
	"github.com/gin-gonic/gin"
)

// loadTraceConfig writes a TOML config to a temp file, points the conf flag at
// it, loads it, and runs the trace factory. Returns the factory shutdown func.
func loadTraceConfig(t *testing.T, body string) func() {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "app.toml")
	require.NoError(t, os.WriteFile(path, []byte(body), 0644))
	require.NoError(t, flag.Set("conf", path))
	require.NoError(t, flag.Set("watch", "false"))
	conf.Init()
	return trace.Init()
}

// TestTrace: with no trace backend initialized, the middleware writes an empty
// X-Trace-Id (noopTracer) and TraceID is "".
func TestTrace(t *testing.T) {
	engine := gin.New()
	engine.Use(Trace(config.DefaultServerConfig()))
	engine.GET("/trace/mock", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, trace.TraceID(ctx.Request.Context()))
	})

	req := httptest.NewRequest(http.MethodGet, "/trace/mock", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, "", w.Body.String())
	assert.Equal(t, "", w.Header().Get("X-Trace-Id"))
}

// TestJaegerTrace: with [trace.jaeger] configured, the middleware starts a
// jaeger server span, TraceID is non-empty, and X-Trace-Id carries it.
func TestJaegerTrace(t *testing.T) {
	closeFn := loadTraceConfig(t, `
[trace]
    [trace.jaeger]
        serviceName = "trace"
        samplerType = "const"
        samplerParam = 1
        agentAddr = "127.0.0.1:6831"
`)
	defer closeFn()

	engine := gin.New()
	engine.Use(Trace(config.DefaultServerConfig()))
	engine.GET("/trace/jaeger", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, trace.TraceID(ctx.Request.Context()))
	})

	req := httptest.NewRequest(http.MethodGet, "/trace/jaeger", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.NotEqual(t, 0, len(w.Body.String()))
	assert.Equal(t, w.Body.String(), w.Header().Get("X-Trace-Id"))
	assert.Len(t, w.Body.String(), 16, "jaeger 64-bit id")
}
