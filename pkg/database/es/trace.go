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

package es

import (
	"net/http"
	"strconv"

	"github.com/opentracing/opentracing-go/log"

	"github.com/opentracing/opentracing-go/ext"

	"github.com/UnderTreeTech/waterdrop/pkg/trace"
)

type Transport struct {
	// The actual RoundTripper to use for the request. A nil
	// RoundTripper defaults to http.DefaultTransport.
	rt     http.RoundTripper
	config *Config
}

func NewTransport(config *Config) *Transport {
	return &Transport{
		config: config,
	}
}

func (t *Transport) SetRoundTripper(rt http.RoundTripper) *Transport {
	t.rt = rt
	return t
}

func (t *Transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	span, ctx := trace.StartSpanFromContext(req.Context(), "es")
	req = req.WithContext(ctx)
	ext.Component.Set(span, t.config.Version)
	ext.HTTPUrl.Set(span, req.URL.String())
	ext.HTTPMethod.Set(span, req.Method)
	ext.PeerHostname.Set(span, req.URL.Hostname())
	port, _ := strconv.Atoi(req.URL.Port())
	ext.PeerPort.Set(span, uint16(port))
	defer span.Finish()

	if t.rt != nil {
		resp, err = t.rt.RoundTrip(req)
	} else {
		resp, err = http.DefaultTransport.RoundTrip(req)
	}

	if err != nil {
		ext.Error.Set(span, true)
		span.LogFields(log.String("event", "error"), log.String("message", err.Error()))
	}

	if resp != nil {
		ext.HTTPStatusCode.Set(span, uint16(resp.StatusCode))
	}

	return
}
