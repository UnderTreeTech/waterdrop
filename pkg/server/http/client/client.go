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

package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/UnderTreeTech/waterdrop/pkg/server/http/metadata"

	"github.com/UnderTreeTech/waterdrop/pkg/utils/xcrypto"

	"github.com/UnderTreeTech/waterdrop/pkg/server/http/config"

	"github.com/UnderTreeTech/waterdrop/pkg/breaker"

	"github.com/UnderTreeTech/waterdrop/pkg/status"
	"github.com/UnderTreeTech/waterdrop/pkg/trace"
	"github.com/opentracing/opentracing-go/ext"

	"github.com/UnderTreeTech/waterdrop/pkg/utils/xtime"

	"github.com/UnderTreeTech/waterdrop/pkg/log"
	"github.com/UnderTreeTech/waterdrop/pkg/utils/xstring"
	tlog "github.com/opentracing/opentracing-go/log"

	"github.com/go-resty/resty/v2"
)

// Request request params
type Request struct {
	URI        string
	QueryParam url.Values
	Body       interface{}
	PathParam  map[string]string
}

// RequestMiddleware http request middleware
type RequestMiddleware func(client *Client) resty.RequestMiddleware

// Client http client
type Client struct {
	client   *resty.Client
	config   *config.ClientConfig
	breakers *breaker.BreakerGroup
}

// New return a http client
func New(config *config.ClientConfig) *Client {
	cli := resty.New()
	cli.SetTimeout(config.Timeout)
	cli.SetDebug(config.EnableDebug)
	cli.SetHostURL(config.HostURL)

	return &Client{
		client:   cli,
		config:   config,
		breakers: breaker.NewBreakerGroup(),
	}
}

// Use set client request middleware
func (c *Client) Use(m RequestMiddleware) *Client {
	rm := m(c)
	c.client = c.client.OnBeforeRequest(rm)
	return c
}

// NewRequest return a resty Request object
func (c *Client) NewRequest(method string, req *Request, reply interface{}) *resty.Request {
	request := c.client.NewRequest()
	request.URL = req.URI
	request.Method = method
	request.SetQueryParamsFromValues(req.QueryParam)
	request.SetBody(req.Body)
	request.SetPathParams(req.PathParam)
	request.SetResult(reply)
	if method != http.MethodGet {
		request.SetHeader(metadata.HeaderContentType, metadata.DefaultContentTypeJson)
	}
	request.SetHeader(metadata.HeaderAppkey, c.config.Key)
	request.SetHeader(metadata.HeaderUserAgent, metadata.DefaultUserAgentVal)
	request.SetHeader(metadata.HeaderAcceptLanguage, metadata.DefaultLocale)
	return request
}

// execute send http request
func (c *Client) execute(ctx context.Context, request *resty.Request) error {
	err := c.breakers.Do(request.URL,
		func() error {
			// adjust request timeout
			timeout := c.config.Timeout
			if deadline, ok := ctx.Deadline(); ok {
				derivedTimeout := time.Until(deadline)
				if timeout > derivedTimeout {
					timeout = derivedTimeout
				}
			}

			request.SetHeader(metadata.HeaderHttpTimeout, strconv.Itoa(int(timeout.Milliseconds())))
			span, sctx := trace.StartSpanFromContext(ctx, request.Method+" "+request.URL)
			sctx = trace.HeaderInjector(sctx, request.Header)
			ext.Component.Set(span, "http")
			ext.SpanKind.Set(span, ext.SpanKindRPCClientEnum)
			ext.HTTPMethod.Set(span, request.Method)
			ext.HTTPUrl.Set(span, request.URL)
			request.SetContext(sctx)
			// zero timeout config means never timeout
			var cancel func()
			if timeout > 0 {
				sctx, cancel = context.WithTimeout(sctx, timeout)
			} else {
				cancel = func() {}
			}
			defer func() {
				span.Finish()
				cancel()
			}()

			now := time.Now()
			var quota float64
			if deadline, ok := sctx.Deadline(); ok {
				quota = time.Until(deadline).Seconds()
			}
			response, err := request.Execute(request.Method, request.URL)
			estatus := status.OK
			if err != nil {
				if uerr, ok := err.(*url.Error); ok {
					err = uerr.Unwrap()
				}
				estatus = status.ExtractContextStatus(err)
			}

			if estatus.Code() != status.OK.Code() {
				ext.Error.Set(span, true)
				span.LogFields(tlog.String("event", "error"), tlog.Int("code", estatus.Code()), tlog.String("message", estatus.Message()))
			}

			duration := time.Since(now)
			fields := make([]log.Field, 0, 12)
			fields = append(
				fields,
				log.String("host", c.client.HostURL),
				log.String("method", request.Method),
				log.String("path", strings.TrimPrefix(request.URL, c.client.HostURL)),
				log.Any("headers", request.Header),
				log.String("query", request.QueryParam.Encode()),
				log.Any("body", request.Body),
				log.Float64("quota", quota),
				log.Float64("duration", duration.Seconds()),
				log.Bytes("reply", response.Body()),
				log.Int("status", response.StatusCode()),
				log.Int("code", estatus.Code()),
				log.String("error", estatus.Message()),
			)

			if duration >= c.config.SlowRequestDuration {
				log.Warn(sctx, "http-slow-request-log", fields...)
			} else {
				log.Info(sctx, "http-request-log", fields...)
			}

			if estatus.Code() == status.OK.Code() {
				return nil
			}
			return estatus
		},
		accept)

	return err
}

// accept calculate request success/failure ratio
func accept(err error) bool {
	if err != nil {
		switch status.ExtractContextStatus(err).Code() {
		case status.Deadline.Code(), status.LimitExceed.Code(),
			status.ServerErr.Code(), status.Canceled.Code(),
			status.ServiceUnavailable.Code():
			return false
		default:
			return true
		}
	}
	return true
}

// Get http get request
func (c *Client) Get(ctx context.Context, req *Request, reply interface{}) error {
	request := c.NewRequest(http.MethodGet, req, reply)
	return c.execute(ctx, request)
}

// Post http post request
func (c *Client) Post(ctx context.Context, req *Request, reply interface{}) error {
	request := c.NewRequest(http.MethodPost, req, reply)
	return c.execute(ctx, request)
}

// Put http put request
func (c *Client) Put(ctx context.Context, req *Request, reply interface{}) error {
	request := c.NewRequest(http.MethodPut, req, reply)
	return c.execute(ctx, request)
}

// Delete http delete request
func (c *Client) Delete(ctx context.Context, req *Request, reply interface{}) error {
	request := c.NewRequest(http.MethodDelete, req, reply)
	return c.execute(ctx, request)
}

// Signature a example of RequestMiddleware
// sign algorithm:md5(query params + body + secret + timestamp + nonce)
// Notice:stuff body only when HTTP METHOD is not GET.
// Encode query params to `"bar=baz&foo=quux"` sorted by key in any case
func Signature(client *Client) resty.RequestMiddleware {
	return func(cli *resty.Client, request *resty.Request) error {
		ts := strconv.Itoa(int(xtime.Now().CurrentUnixTime()))
		nonce := xstring.RandomString(metadata.DefaultNonceLen)
		sb := strings.Builder{}
		query := request.QueryParam.Encode()
		bodyStr := ""
		if request.Method != http.MethodGet {
			jsonReq, err := json.Marshal(request.Body)
			if err != nil {
				return err
			}
			bodyStr = xstring.BytesToString(jsonReq)
		}

		sb.WriteString(query)
		sb.WriteString(bodyStr)
		sb.WriteString(client.config.Secret)
		sb.WriteString(ts)
		sb.WriteString(nonce)
		signStr := sb.String()
		sign, err := xcrypto.HashToString(signStr, xcrypto.MD5, xcrypto.HEX)
		if err != nil {
			return err
		}
		request.SetHeader(metadata.HeaderSign, sign)
		request.SetHeader(metadata.HeaderNonce, nonce)
		request.SetHeader(metadata.HeaderTimestamp, ts)
		return nil
	}
}
