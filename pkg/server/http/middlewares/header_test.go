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
	"strings"
	"testing"

	"github.com/go-playground/assert/v2"

	"github.com/gin-gonic/gin"
)

func TestHeader(t *testing.T) {
	engine := gin.New()
	engine.Use(Header())
	engine.POST("/header", func(ctx *gin.Context) {})

	headers := make([][]string, 0)
	headers = append(
		headers,
		[]string{"Content-Type:application/text"},
		[]string{"Content-Type:application/json", "appkey"},
		[]string{"Content-Type:application/json", "appkey:aaa"},
		[]string{"Content-Type:application/json", "appkey:xHf74ZfV43cAUsUl"},
	)

	for index, hds := range headers {
		req := httptest.NewRequest(http.MethodPost, "/header", nil)
		for _, header := range hds {
			kvs := strings.Split(header, ":")
			if len(kvs) > 1 {
				req.Header.Add(kvs[0], kvs[1])
			} else {
				req.Header.Add(kvs[0], "")
			}
		}
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		if index != len(headers)-1 {
			assert.Equal(t, http.StatusBadRequest, w.Code)
		} else {
			assert.Equal(t, http.StatusOK, w.Code)
		}
	}
}
