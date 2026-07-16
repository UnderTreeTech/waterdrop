/*
 *
 * Copyright 2026 waterdrop authors.
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
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/UnderTreeTech/waterdrop/pkg/log"
	"github.com/UnderTreeTech/waterdrop/pkg/server/http/metadata"
	"github.com/UnderTreeTech/waterdrop/pkg/status"
	"github.com/UnderTreeTech/waterdrop/pkg/trace"
)

const (
	// defaultSlowRequestDuration is the threshold above which an HTTP call is
	// logged at Warn as a slow request. Mirrors the resty client default.
	defaultSlowRequestDuration = 500 * time.Millisecond
	// maxBodyLogBytes caps request-body logging: bodies whose ContentLength is
	// known to exceed this are not buffered for logging (their size is still
	// implicit from the sent request), to avoid OOM on large uploads. Bodies
	// with unknown length are still read in full (typical JSON requests are
	// small). The response body is NEVER buffered - only its size is logged -
	// so streaming (SSE) responses keep flowing.
	maxBodyLogBytes = 8 * 1024
)

// transport is an http.RoundTripper that starts a client span for every
// outbound request, injects its trace context into the request headers, and
// ends the span when the response is received. It is the HTTP outbound
// counterpart of the gRPC client interceptor (TraceForUnaryClient) and the
// HTTP server Trace middleware: all three go through the factory-selected
// trace.GlobalTracer(), so tracing works in jaeger / otel / dual mode.
//
// It also emits a per-request access log (http-request-log / slow ->
// http-slow-request-log), mirroring the resty client in client.go. Tracing is
// mandatory and not bypassable: transport is unexported, and the only way to
// get a traced HTTP client is NewClient, which bakes a transport into an
// opaque TraceClient.
type transport struct {
	// base is the underlying round tripper. If nil, http.DefaultTransport is used.
	base http.RoundTripper
	// slow is the duration threshold above which a request is logged at Warn
	// as http-slow-request-log.
	slow time.Duration
}

// RoundTrip implements http.RoundTripper. It starts a client span, injects the
// trace context into req's headers, delegates to base, records the HTTP status
// (and marks the span as errored on transport error or 5xx), emits the access
// log, then ends the span.
func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	base := t.base
	if base == nil {
		base = http.DefaultTransport
	}

	// Capture the request body for logging (small/unknown sizes only), then
	// restore it so base sends the body unchanged. Done before base.RoundTrip
	// so the bytes are available regardless of the call's outcome.
	var reqBody []byte
	if req.Body != nil && req.Body != http.NoBody && (req.ContentLength <= 0 || req.ContentLength <= maxBodyLogBytes) {
		reqBody, _ = io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewReader(reqBody))
	}

	carrier := trace.HTTPCarrier{Header: req.Header}
	name := req.Method + " " + req.URL.Path
	span, sctx := trace.GlobalTracer().StartClientSpan(req.Context(), name, trace.SpanClient, carrier)
	defer span.End()

	span.SetStringAttr("http.method", req.Method)
	span.SetStringAttr("http.url", req.URL.String())

	now := time.Now()
	resp, err := base.RoundTrip(req)
	duration := time.Since(now)
	if err != nil {
		if uerr, ok := err.(*url.Error); ok {
			err = uerr.Unwrap()
		}
		span.SetStatus(false, err.Error())
		t.logHTTP(sctx, req, reqBody, nil, duration, status.ExtractContextStatus(err))
		return nil, err
	}

	span.SetIntAttr("http.status_code", int64(resp.StatusCode))
	if resp.StatusCode >= http.StatusInternalServerError {
		span.SetStatus(false, http.StatusText(resp.StatusCode))
	}
	// estatus reflects transport-level status only (mirrors resty client): a
	// non-2xx HTTP response is surfaced via the "status" field, not as an error.
	t.logHTTP(sctx, req, reqBody, resp, duration, status.OK)
	return resp, nil
}

// logHTTP emits the per-request HTTP access log, mirroring the field set and
// slow-request threshold of the resty client (client.go). The response body is
// intentionally NOT buffered - only reply_size is logged - so streaming (SSE)
// responses keep flowing.
func (t *transport) logHTTP(ctx context.Context, req *http.Request, reqBody []byte, resp *http.Response, duration time.Duration, estatus *status.Status) {
	var quota float64
	if deadline, ok := req.Context().Deadline(); ok {
		quota = time.Until(deadline).Seconds()
	}

	fields := make([]log.Field, 0, 12)
	fields = append(fields,
		log.String("peer", req.URL.Host),
		log.String("method", req.Method),
		log.String("path", req.URL.Path),
		log.Any("headers", req.Header),
		log.String("query", req.URL.RawQuery),
		log.Bytes("body", reqBody),
		log.Float64("quota", quota),
		log.Float64("duration", duration.Seconds()),
	)
	if resp != nil {
		fields = append(fields,
			log.Int64("reply_size", resp.ContentLength),
			log.Int("status", resp.StatusCode),
		)
	}
	fields = append(fields,
		log.Int("code", estatus.Code()),
		log.String("error", estatus.Message()),
	)

	if duration >= t.slow {
		log.Warn(ctx, "http-slow-request-log", fields...)
	} else {
		log.Info(ctx, "http-request-log", fields...)
	}
}

// TraceClient is an HTTP client whose every outbound request is traced. The
// underlying *http.Client and its transport are not exposed, so tracing cannot
// be turned off or bypassed - this is intentional: tracing is the default and
// not the caller's decision.
//
// Build it with NewClient; configure timeout / base transport / slow-request
// threshold via options. The parent span is read from req.Context(), so build
// requests with http.NewRequestWithContext(ctx, ...) to carry the inbound span.
type TraceClient struct {
	c *http.Client
}

// Do sends an HTTP request and returns an HTTP response, tracing the call.
// It mirrors (*http.Client).Do.
func (tc *TraceClient) Do(req *http.Request) (*http.Response, error) {
	return tc.c.Do(req)
}

// Get issues a GET to the specified URL. It mirrors (*http.Client).Get.
func (tc *TraceClient) Get(url string) (*http.Response, error) {
	return tc.c.Get(url)
}

// Post issues a POST to the specified URL. It mirrors (*http.Client).Post.
func (tc *TraceClient) Post(url, contentType string, body io.Reader) (*http.Response, error) {
	return tc.c.Post(url, contentType, body)
}

// PostJson issues a POST with a JSON content type (metadata.DefaultContentTypeJson).
// It is a convenience over Post so callers don't have to repeat the content type.
func (tc *TraceClient) PostJson(url string, body io.Reader) (*http.Response, error) {
	return tc.c.Post(url, metadata.DefaultContentTypeJson, body)
}

// PostForm issues a POST to the specified URL with data's keys and values
// URL-encoded as the request body. It mirrors (*http.Client).PostForm.
func (tc *TraceClient) PostForm(url string, data url.Values) (*http.Response, error) {
	return tc.c.PostForm(url, data)
}

// Head issues a HEAD to the specified URL. It mirrors (*http.Client).Head.
func (tc *TraceClient) Head(url string) (*http.Response, error) {
	return tc.c.Head(url)
}

// CloseIdleConnections closes any connections on its underlying transport
// that were previously connected from previous requests but are now idle.
func (tc *TraceClient) CloseIdleConnections() {
	tc.c.CloseIdleConnections()
}

// clientConfig holds the build-time configuration for a TraceClient. Options
// mutate it; NewClient then freezes it into a TraceClient with a transport
// layered on top of base - callers can configure the base transport (e.g. an
// *http.Transport with DisableCompression for SSE) but never the tracing one.
type clientConfig struct {
	timeout             time.Duration
	base                http.RoundTripper
	slowRequestDuration time.Duration
}

// Option configures the TraceClient returned by NewClient.
type Option func(*clientConfig)

// WithTimeout sets the client's overall request timeout.
func WithTimeout(d time.Duration) Option {
	return func(cfg *clientConfig) { cfg.timeout = d }
}

// WithBaseTransport sets the underlying round tripper that the tracing
// transport delegates to. Pass an *http.Transport (e.g. with
// DisableCompression for SSE) or any custom http.RoundTripper; the tracing
// layer is always applied on top, so tracing cannot be removed from here.
func WithBaseTransport(rt http.RoundTripper) Option {
	return func(cfg *clientConfig) { cfg.base = rt }
}

// WithSlowRequestThreshold sets the duration above which a request is logged
// at Warn as http-slow-request-log (default 500ms, matching the resty client).
func WithSlowRequestThreshold(d time.Duration) Option {
	return func(cfg *clientConfig) { cfg.slowRequestDuration = d }
}

// NewClient returns a TraceClient whose every request is traced. Tracing is
// always on and cannot be bypassed. Use the options to set a timeout, a custom
// base transport, and/or the slow-request log threshold:
//
//	c := client.NewClient(
//	    client.WithTimeout(10*time.Second),
//	    client.WithBaseTransport(&http.Transport{DisableCompression: true}),
//	)
//	resp, err := c.Do(httpReq) // client span auto-started/ended, headers injected, access log emitted
func NewClient(opts ...Option) *TraceClient {
	cfg := &clientConfig{
		base:                http.DefaultTransport,
		slowRequestDuration: defaultSlowRequestDuration,
	}
	for _, o := range opts {
		o(cfg)
	}
	return &TraceClient{
		c: &http.Client{
			Timeout:   cfg.timeout,
			Transport: &transport{base: cfg.base, slow: cfg.slowRequestDuration},
		},
	}
}
