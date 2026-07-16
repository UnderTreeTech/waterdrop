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

package tokenbucket

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/UnderTreeTech/waterdrop/pkg/server/http/middlewares/ratelimit"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
)

func newEngine(handler gin.HandlerFunc) *gin.Engine {
	engine := gin.New()
	engine.Use(handler)
	return engine
}

func do(engine *gin.Engine, method, target string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	return w
}

// Global limit applies to every route: the burst is served, the rest rejected.
func TestGlobalLimit(t *testing.T) {
	engine := newEngine(New(&Config{
		Global: Rate{Rate: rate.Limit(1), Burst: 1},
	}))
	engine.GET("/", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	assert.Equal(t, http.StatusOK, do(engine, http.MethodGet, "/").Code)              // burst token
	assert.Equal(t, http.StatusTooManyRequests, do(engine, http.MethodGet, "/").Code) // no token
}

// A route-specific limit applies to that route, while an unconfigured route
// falls back to the (separate) global limit.
func TestRouteLimitPrecedenceAndFallback(t *testing.T) {
	engine := newEngine(New(&Config{
		Global: Rate{Rate: rate.Limit(100), Burst: 100}, // generous global
		Routes: map[string]Rate{
			"GET:/expensive": {Rate: rate.Limit(1), Burst: 1}, // tight route limit
		},
	}))
	engine.GET("/expensive", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	engine.GET("/cheap", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	// /expensive is capped at its own burst of 1.
	assert.Equal(t, http.StatusOK, do(engine, http.MethodGet, "/expensive").Code)
	assert.Equal(t, http.StatusTooManyRequests, do(engine, http.MethodGet, "/expensive").Code)

	// /cheap has no route rule, so it falls back to the global limit and is allowed.
	assert.Equal(t, http.StatusOK, do(engine, http.MethodGet, "/cheap").Code)
}

// A zero-value Global means unlimited: only configured routes are limited.
func TestZeroGlobalIsUnlimited(t *testing.T) {
	engine := newEngine(New(&Config{
		// Global left zero -> unlimited.
		Routes: map[string]Rate{
			"GET:/expensive": {Rate: rate.Limit(1), Burst: 1},
		},
	}))
	engine.GET("/expensive", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	engine.GET("/open", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	// /expensive is limited; /open is unlimited (10 rapid requests all pass).
	assert.Equal(t, http.StatusOK, do(engine, http.MethodGet, "/expensive").Code)
	assert.Equal(t, http.StatusTooManyRequests, do(engine, http.MethodGet, "/expensive").Code)
	for i := 0; i < 10; i++ {
		assert.Equal(t, http.StatusOK, do(engine, http.MethodGet, "/open").Code)
	}
}

// WithFallback overrides the default 429 response.
func TestWithFallback(t *testing.T) {
	opts := []ratelimit.Option{
		ratelimit.WithFallback(func(c *gin.Context) {
			c.String(http.StatusServiceUnavailable, "slow down")
		}),
	}
	engine := newEngine(New(&Config{Global: Rate{Rate: rate.Limit(1), Burst: 1}}, opts...))
	engine.GET("/", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	assert.Equal(t, http.StatusOK, do(engine, http.MethodGet, "/").Code)
	resp := do(engine, http.MethodGet, "/")
	assert.Equal(t, http.StatusServiceUnavailable, resp.Code)
	assert.Equal(t, "slow down", resp.Body.String())
}

// WithResourceStrategy lets the caller key the limiter on a custom value
// (here: FullPath only, dropping the method prefix).
func TestWithResourceStrategy(t *testing.T) {
	opts := []ratelimit.Option{
		ratelimit.WithResourceStrategy(func(c *gin.Context) string {
			return c.FullPath()
		}),
	}
	engine := newEngine(New(&Config{
		Routes: map[string]Rate{
			"/api/app/validate/:id": {Rate: rate.Limit(1), Burst: 1},
		},
	}, opts...))
	engine.POST("/api/app/validate/:id", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	// Routed via the strategy key "/api/app/validate/:id", not the default
	// "POST:/api/app/validate/:id".
	assert.Equal(t, http.StatusOK, do(engine, http.MethodPost, "/api/app/validate/1").Code)
	assert.Equal(t, http.StatusTooManyRequests, do(engine, http.MethodPost, "/api/app/validate/2").Code)
}

// A nil config yields an unlimited middleware (no request is rejected).
func TestNilConfigUnlimited(t *testing.T) {
	engine := newEngine(New(nil))
	engine.GET("/", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	for i := 0; i < 50; i++ {
		assert.Equal(t, http.StatusOK, do(engine, http.MethodGet, "/").Code)
	}
}
