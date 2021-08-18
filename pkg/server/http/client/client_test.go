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

package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/UnderTreeTech/waterdrop/pkg/server/http/metadata"

	"github.com/UnderTreeTech/waterdrop/pkg/log"

	"github.com/stretchr/testify/assert"

	"github.com/UnderTreeTech/waterdrop/pkg/server/http/config"
)

// result response
type result struct {
	Method string `json:"method"`
}

// TestHttp test http GET, POST, PUT, DELETE method
func TestHttp(t *testing.T) {
	defer log.New(nil).Sync()
	srv := createTestServer(t)
	defer srv.Close()

	cfg := &config.ClientConfig{
		HostURL: srv.URL,
	}
	client := New(cfg)

	rget, err := client.RawGet(context.Background(), &Request{URI: "/rawget"})
	assert.Nil(t, err)
	assert.Equal(t, string(rget), "rawget")
	rpost, err := client.RawPost(context.Background(), &Request{URI: "/rawpost"})
	assert.Nil(t, err)
	assert.Equal(t, string(rpost), "rawpost")
	rput, err := client.RawPut(context.Background(), &Request{URI: "/rawput"})
	assert.Nil(t, err)
	assert.Equal(t, string(rput), "rawput")
	rdel, err := client.RawDelete(context.Background(), &Request{URI: "/rawdelete"})
	assert.Nil(t, err)
	assert.Equal(t, string(rdel), "rawdelete")

	get := &result{}
	err = client.Get(context.Background(), &Request{URI: "/get"}, get)
	assert.Nil(t, err)
	assert.Equal(t, get.Method, "get")
	post := &result{}
	err = client.Get(context.Background(), &Request{URI: "/post"}, post)
	assert.Nil(t, err)
	assert.Equal(t, post.Method, "post")
	put := &result{}
	err = client.Get(context.Background(), &Request{URI: "/put"}, put)
	assert.Nil(t, err)
	assert.Equal(t, put.Method, "put")
	del := &result{}
	err = client.Get(context.Background(), &Request{URI: "/delete"}, del)
	assert.Nil(t, err)
	assert.Equal(t, del.Method, "delete")
}

// TestRequestMiddleware test client use middleware
func TestRequestMiddleware(t *testing.T) {
	defer log.New(nil).Sync()
	srv := createTestServer(t)
	defer srv.Close()

	cfg := &config.ClientConfig{
		HostURL: srv.URL,
	}
	client := New(cfg)
	client.Use(Signature)

	rget, err := client.RawGet(context.Background(), &Request{URI: "/rawget"})
	assert.Nil(t, err)
	assert.Equal(t, string(rget), "rawget")
}

// createTestServer create test server
func createTestServer(t *testing.T) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Method: %v", r.Method)
		t.Logf("Path: %v", r.URL.Path)

		switch r.URL.Path {
		case "/rawget":
			sign := r.Header.Get(metadata.HeaderSign)
			fmt.Println("sign", sign)
			_, _ = w.Write([]byte("rawget"))
		case "/rawpost":
			_, _ = w.Write([]byte("rawpost"))
		case "/rawput":
			_, _ = w.Write([]byte("rawput"))
		case "/rawdelete":
			_, _ = w.Write([]byte("rawdelete"))
		case "/get":
			w.Header().Set("Content-Type", "application/json")
			get := &result{Method: "get"}
			ret, _ := json.Marshal(get)
			_, _ = w.Write(ret)
		case "/post":
			w.Header().Set("Content-Type", "application/json")
			post := &result{Method: "post"}
			ret, _ := json.Marshal(post)
			_, _ = w.Write(ret)
		case "/put":
			w.Header().Set("Content-Type", "application/json")
			put := &result{Method: "put"}
			ret, _ := json.Marshal(put)
			_, _ = w.Write(ret)
		case "/delete":
			w.Header().Set("Content-Type", "application/json")
			del := &result{Method: "delete"}
			ret, _ := json.Marshal(del)
			_, _ = w.Write(ret)
		}
	}))

	return ts
}
