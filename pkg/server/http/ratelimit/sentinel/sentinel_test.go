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

package sentinel

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/UnderTreeTech/waterdrop/pkg/log"

	"github.com/UnderTreeTech/waterdrop/pkg/server/http/ratelimit"

	"github.com/go-playground/assert/v2"

	"github.com/gin-gonic/gin"

	"github.com/alibaba/sentinel-golang/core/flow"

	"github.com/alibaba/sentinel-golang/api"
)

func TestMain(m *testing.M) {
	defer log.New(nil).Sync()
	if err := api.InitDefault(); err != nil {
		panic(fmt.Sprintf("init sentinel entity fail, error is %s", err.Error()))
	}

	_, err := flow.LoadRules([]*flow.Rule{
		{
			Resource:               "GET:/ping",
			Threshold:              1.0,
			TokenCalculateStrategy: flow.Direct,
			ControlBehavior:        flow.Reject,
			StatIntervalInMs:       1000,
		},
		{
			Resource:               "/api/app/validate/:id",
			Threshold:              0.0,
			TokenCalculateStrategy: flow.Direct,
			ControlBehavior:        flow.Reject,
			StatIntervalInMs:       1000,
		},
		{
			Resource:               "GET:/api/app/secrets",
			Threshold:              0.0,
			TokenCalculateStrategy: flow.Direct,
			ControlBehavior:        flow.Reject,
			StatIntervalInMs:       1000,
		},
	})

	if err != nil {
		panic(fmt.Sprintf("load rules fail, error is %s", err.Error()))
	}

	code := m.Run()
	os.Exit(code)
}

func TestSentinel(t *testing.T) {
	// default sentinel without WithMethods
	t.Run("default sentinel", func(t *testing.T) {
		engine := gin.New()
		engine.Use(Sentinel())
		engine.Handle(http.MethodGet, "/ping", func(c *gin.Context) {
			c.String(http.StatusOK, "pong")
		})
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		resp := httptest.NewRecorder()
		engine.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)
	})

	// sentinel with Option WithResourceStrategy
	t.Run("resource strategy", func(t *testing.T) {
		opts := []ratelimit.Option{
			ratelimit.WithResourceStrategy(func(c *gin.Context) string {
				return c.FullPath()
			}),
		}
		engine := gin.New()
		engine.Use(Sentinel(opts...))
		engine.Handle(http.MethodPost, "/api/app/validate/:id", func(c *gin.Context) {
			c.String(http.StatusOK, "validate")
		})
		req := httptest.NewRequest(http.MethodPost, "/api/app/validate/1", nil)
		resp := httptest.NewRecorder()
		engine.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusTooManyRequests, resp.Code)
	})

	// sentinel with Option WithFallback
	t.Run("block fallback", func(t *testing.T) {
		opts := []ratelimit.Option{
			ratelimit.WithFallback(func(c *gin.Context) {
				c.String(http.StatusBadRequest, "blocked: request exceed")
			}),
		}
		engine := gin.New()
		engine.Use(Sentinel(opts...))
		engine.Handle(http.MethodGet, "/api/app/secrets", func(c *gin.Context) {
			c.String(http.StatusOK, "secrets")
		})
		req := httptest.NewRequest(http.MethodGet, "/api/app/secrets", nil)
		resp := httptest.NewRecorder()
		engine.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

}
