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
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/UnderTreeTech/waterdrop/pkg/server/http/config"
	"github.com/gin-gonic/gin"
)

func TestLogger(t *testing.T) {
	engine := gin.New()
	engine.Use(Logger(config.DefaultServerConfig()))

	engine.GET("/log/normal", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "normal")
	})
	engine.GET("/log/slow", func(ctx *gin.Context) {
		time.Sleep(time.Second)
		ctx.String(http.StatusOK, "slow")
	})

	var tests = []string{"/log/normal", "/log/slow"}

	for _, test := range tests {
		req := httptest.NewRequest(http.MethodGet, test, nil)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		assert.Contains(t, test, w.Body.String())
	}
}
