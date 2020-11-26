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
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/UnderTreeTech/waterdrop/pkg/server/http/config"

	md "github.com/UnderTreeTech/waterdrop/pkg/server/http/metadata"

	"github.com/UnderTreeTech/waterdrop/pkg/breaker"

	"google.golang.org/grpc/metadata"

	"github.com/UnderTreeTech/waterdrop/pkg/status"
	"github.com/UnderTreeTech/waterdrop/pkg/trace"
	"github.com/opentracing/opentracing-go/ext"

	"github.com/UnderTreeTech/waterdrop/pkg/utils/xtime"

	"github.com/UnderTreeTech/waterdrop/pkg/log"
	"github.com/UnderTreeTech/waterdrop/pkg/utils/xstring"
	tlog "github.com/opentracing/opentracing-go/log"

	"github.com/go-resty/resty/v2"
)

type Request struct {
	URI         string
	QueryParams url.Values
	Body        interface{}
	PathParams  map[string]string
}

type Client struct {
	client   *resty.Client
	config   *config.ClientConfig
	breakers *breaker.BreakerGroup
}

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

func (c *Client) NewRequest(ctx context.Context, method string, req *Request, reply interface{}) (*resty.Request, error) {
	request := c.client.NewRequest()
	request.URL = req.URI
	request.Method = method
	request.SetQueryParamsFromValues(req.QueryParams)
	request.SetBody(req.Body)
	request.SetPathParams(req.PathParams)
	request.SetResult(reply)

	if c.config.EnableSign {
		ts := strconv.Itoa(int(xtime.Now().CurrentUnixTime()))
		nonce := xstring.RandomString(md.DefaultNonceLen)
		sign, err := c.sign(ctx, method, ts, nonce, req)
		if err != nil {
			return nil, err
		}

		request.SetHeader(md.HeaderSign, sign)
		request.SetHeader(md.HeaderNonce, nonce)
		request.SetHeader(md.HeaderTimestamp, ts)
	}

	if method != http.MethodGet {
		request.SetHeader(md.HeaderContentType, md.DefaultContentTypeJson)
	}

	request.SetHeader(md.HeaderAppkey, c.config.Key)
	request.SetHeader(md.HeaderUserAgent, md.DefaultUserAgentVal)
	request.SetHeader(md.HeaderAcceptLanguage, md.DefaultLocale)

	return request, nil
}

func (c *Client) execute(ctx context.Context, request *resty.Request) error {
	err := c.breakers.Do(request.URL,
		func() error {
			span, ctx := trace.StartSpanFromContext(ctx, request.Method+" "+request.URL)
			ext.Component.Set(span, "http")
			ext.SpanKind.Set(span, ext.SpanKindRPCClientEnum)
			ext.HTTPMethod.Set(span, request.Method)

			// adjust request timeout
			timeout := c.config.Timeout
			if deadline, ok := ctx.Deadline(); ok {
				derivedTimeout := time.Until(deadline)
				if timeout > derivedTimeout {
					timeout = derivedTimeout
				}
			}

			ctx = metadata.NewOutgoingContext(ctx, metadata.MD(request.Header))
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer func() {
				span.Finish()
				cancel()
			}()

			request.SetHeader(md.HeaderHttpTimeout, strconv.Itoa(int(timeout)))
			request.SetContext(ctx)

			trace.MetadataInjector(ctx, metadata.MD(request.Header))

			now := time.Now()
			var quota float64
			if deadline, ok := ctx.Deadline(); ok {
				quota = time.Until(deadline).Seconds()
			}

			response, err := request.Execute(request.Method, request.URL)
			ext.HTTPUrl.Set(span, request.URL)
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
				log.Warn(ctx, "http-slow-request-log", fields...)
			} else {
				log.Info(ctx, "http-request-log", fields...)
			}

			if estatus.Code() == status.OK.Code() {
				return nil
			}
			return estatus
		},
		accept)

	return err
}

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

func (c *Client) Get(ctx context.Context, req *Request, reply interface{}) error {
	request, err := c.NewRequest(ctx, http.MethodGet, req, reply)
	if err != nil {
		return err
	}

	return c.execute(ctx, request)
}

func (c *Client) Post(ctx context.Context, req *Request, reply interface{}) error {
	request, err := c.NewRequest(ctx, http.MethodPost, req, reply)
	if err != nil {
		return err
	}

	return c.execute(ctx, request)
}

func (c *Client) Put(ctx context.Context, req *Request, reply interface{}) error {
	request, err := c.NewRequest(ctx, http.MethodPut, req, reply)
	if err != nil {
		return err
	}

	return c.execute(ctx, request)
}

func (c *Client) Delete(ctx context.Context, req *Request, reply interface{}) error {
	request, err := c.NewRequest(ctx, http.MethodDelete, req, reply)
	if err != nil {
		return err
	}

	return c.execute(ctx, request)
}

// sign algorithm:md5(query params + body + secret + timestamp + nonce)
// Notice:stuff body only when HTTP METHOD is not GET.
// Encode query params to `"bar=baz&foo=quux"` sorted by key in any case
func (c *Client) sign(ctx context.Context, method string, timestamp string, nonce string, req *Request) (string, error) {
	sb := strings.Builder{}
	query := req.QueryParams.Encode()
	bodyStr := ""
	if method != http.MethodGet {
		jsonReq, err := json.Marshal(req.Body)
		if err != nil {
			return "", err
		}
		bodyStr = xstring.BytesToString(jsonReq)
	}

	sb.WriteString(query)
	sb.WriteString(bodyStr)
	sb.WriteString(c.config.Secret)
	sb.WriteString(timestamp)
	sb.WriteString(nonce)
	signStr := sb.String()

	digest := md5.Sum(xstring.StringToBytes(signStr))
	sign := hex.EncodeToString(digest[:])
	log.Debug(ctx, "signature info", log.String("sign_str", signStr), log.String("sign", sign))
	return sign, nil
}
