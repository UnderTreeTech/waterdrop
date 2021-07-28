/*
 *
 * Copyright 2021 waterdrop authors.
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

	"github.com/gin-contrib/cors"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func newTestRouter(config interface{}) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	if c, ok := config.(CORSConfig); ok {
		router.Use(NewCORS(c))
	} else {
		router.Use(DefaultCORS())
	}
	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "get")
	})
	router.POST("/", func(c *gin.Context) {
		c.String(http.StatusOK, "post")
	})
	router.PATCH("/", func(c *gin.Context) {
		c.String(http.StatusOK, "patch")
	})
	return router
}

func performRequest(r http.Handler, method, origin string) *httptest.ResponseRecorder {
	return performRequestWithHeaders(r, method, origin, http.Header{})
}

func performRequestWithHeaders(r http.Handler, method, origin string, header http.Header) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, "/", nil)
	req.Host = header.Get("Host")
	header.Del("Host")
	if len(origin) > 0 {
		header.Set("Origin", origin)
	}
	req.Header = header
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestConfigAddAllow(t *testing.T) {
	cc := CORSConfig{}
	config := cors.Config(cc)
	config.AddAllowMethods("POST")
	config.AddAllowMethods("GET", "PUT")
	config.AddExposeHeaders()

	config.AddAllowHeaders("Some", " cool")
	config.AddAllowHeaders("header")
	config.AddExposeHeaders()

	config.AddExposeHeaders()
	config.AddExposeHeaders("exposed", "header")
	config.AddExposeHeaders("hey")

	assert.Equal(t, config.AllowMethods, []string{"POST", "GET", "PUT"})
	assert.Equal(t, config.AllowHeaders, []string{"Some", " cool", "header"})
	assert.Equal(t, config.ExposeHeaders, []string{"exposed", "header", "hey"})

}

func TestNewCORS(t *testing.T) {
	router := newTestRouter(CORSConfig{
		AllowOrigins:     []string{"http://google.com"},
		AllowMethods:     []string{" GeT ", "get", "post", "PUT  ", "Head", "POST"},
		AllowHeaders:     []string{"Content-type", "timeStamp "},
		ExposeHeaders:    []string{"Data", "x-User"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
		AllowOriginFunc: func(origin string) bool {
			return origin == "http://github.com"
		},
	})

	// no CORS request, origin == ""
	w := performRequest(router, "GET", "")
	assert.Equal(t, "get", w.Body.String())
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Credentials"))
	assert.Empty(t, w.Header().Get("Access-Control-Expose-Headers"))

	// no CORS request, origin == host
	h := http.Header{}
	h.Set("Host", "facebook.com")
	w = performRequestWithHeaders(router, "GET", "http://facebook.com", h)
	assert.Equal(t, "get", w.Body.String())
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Credentials"))
	assert.Empty(t, w.Header().Get("Access-Control-Expose-Headers"))

	// allowed CORS request
	w = performRequest(router, "GET", "http://google.com")
	assert.Equal(t, "get", w.Body.String())
	assert.Equal(t, "http://google.com", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "", w.Header().Get("Access-Control-Allow-Credentials"))
	assert.Equal(t, "Data,X-User", w.Header().Get("Access-Control-Expose-Headers"))

	w = performRequest(router, "GET", "http://github.com")
	assert.Equal(t, "get", w.Body.String())
	assert.Equal(t, "http://github.com", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "", w.Header().Get("Access-Control-Allow-Credentials"))
	assert.Equal(t, "Data,X-User", w.Header().Get("Access-Control-Expose-Headers"))

	// deny CORS request
	w = performRequest(router, "GET", "https://google.com")
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Credentials"))
	assert.Empty(t, w.Header().Get("Access-Control-Expose-Headers"))

	// allowed CORS prefligh request
	w = performRequest(router, "OPTIONS", "http://github.com")
	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "http://github.com", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "", w.Header().Get("Access-Control-Allow-Credentials"))
	assert.Equal(t, "GET,POST,PUT,HEAD", w.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Content-Type,Timestamp", w.Header().Get("Access-Control-Allow-Headers"))
	assert.Equal(t, "43200", w.Header().Get("Access-Control-Max-Age"))

	// deny CORS prefligh request
	w = performRequest(router, "OPTIONS", "http://example.com")
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Credentials"))
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Methods"))
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Headers"))
	assert.Empty(t, w.Header().Get("Access-Control-Max-Age"))
}

func TestDefaultCORS(t *testing.T) {
	router := newTestRouter(nil)
	// no CORS request, origin == ""
	w := performRequest(router, "GET", "")
	assert.Equal(t, "get", w.Body.String())
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Credentials"))
	assert.Empty(t, w.Header().Get("Access-Control-Expose-Headers"))
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Credentials"))

	// allowed CORS request
	w = performRequest(router, "POST", "example.com")
	assert.Equal(t, "post", w.Body.String())
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Credentials"))
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))

	// allowed CORS prefligh request
	w = performRequest(router, "OPTIONS", "https://facebook.com")
	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET,POST,PUT,PATCH,DELETE,HEAD", w.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "43200", w.Header().Get("Access-Control-Max-Age"))
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Credentials"))

}

func TestWildcard(t *testing.T) {
	router := newTestRouter(CORSConfig{
		AllowOrigins:  []string{"https://*.github.com", "https://api.*", "http://*", "https://facebook.com", "*.golang.org"},
		AllowMethods:  []string{"GET"},
		AllowWildcard: true,
	})

	w := performRequest(router, "GET", "https://gist.github.com")
	assert.Equal(t, 200, w.Code)

	w = performRequest(router, "GET", "https://api.github.com/v1/users")
	assert.Equal(t, 200, w.Code)

	w = performRequest(router, "GET", "https://giphy.com/")
	assert.Equal(t, 403, w.Code)

	w = performRequest(router, "GET", "http://hard-to-find-http-example.com")
	assert.Equal(t, 200, w.Code)

	w = performRequest(router, "GET", "https://facebook.com")
	assert.Equal(t, 200, w.Code)

	w = performRequest(router, "GET", "https://something.golang.org")
	assert.Equal(t, 200, w.Code)

	w = performRequest(router, "GET", "https://something.go.org")
	assert.Equal(t, 403, w.Code)

	router = newTestRouter(CORSConfig{
		AllowOrigins: []string{"https://github.com", "https://facebook.com"},
		AllowMethods: []string{"GET"},
	})

	w = performRequest(router, "GET", "https://gist.github.com")
	assert.Equal(t, 403, w.Code)

	w = performRequest(router, "GET", "https://github.com")
	assert.Equal(t, 200, w.Code)
}
