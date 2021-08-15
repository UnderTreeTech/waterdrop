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
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gin-gonic/gin"

	"github.com/UnderTreeTech/waterdrop/pkg/log"
)

func TestMain(m *testing.M) {
	defer log.New(nil).Sync()

	code := m.Run()
	os.Exit(code)
}

func TestRecovery(t *testing.T) {
	engine := gin.New()
	engine.Use(Recovery())

	engine.GET("/recovery/panic", func(ctx *gin.Context) {
		panic("internal server error occurred")
	})

	req := httptest.NewRequest(http.MethodGet, "/recovery/panic", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestPanicWithAbort(t *testing.T) {
	engine := gin.New()
	engine.Use(Recovery())

	engine.GET("/recovery/panic", func(ctx *gin.Context) {
		ctx.AbortWithStatus(http.StatusBadRequest)
		panic("internal server error occurred")
	})

	req := httptest.NewRequest(http.MethodGet, "/recovery/panic", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPanicWithBrokenPipe(t *testing.T) {
	const expectCode = 204

	expectMsgs := map[syscall.Errno]string{
		syscall.EPIPE:      "broken pipe",
		syscall.ECONNRESET: "connection reset by peer",
	}

	for errno, expectMsg := range expectMsgs {
		t.Run(expectMsg, func(t *testing.T) {
			engine := gin.New()
			engine.Use(Recovery())
			engine.GET("/recovery/panic", func(ctx *gin.Context) {
				// Start writing response
				ctx.Header("X-Test", "Value")
				ctx.Status(expectCode)

				// Client connection closed
				e := &net.OpError{Err: &os.SyscallError{Err: errno}}
				panic(e)
			})
			// RUN
			req := httptest.NewRequest(http.MethodGet, "/recovery/panic", nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			assert.Equal(t, expectCode, w.Code)
		})
	}
}
