/*
 *
 * Copyright 2020 waterdrop authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/LICENSE.org/LICENSE-2.0
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
	"context"
	"net/http"

	"github.com/UnderTreeTech/waterdrop/pkg/trace"
	"github.com/UnderTreeTech/waterdrop/pkg/trace/otel"

	"github.com/UnderTreeTech/waterdrop/pkg/server/http/metadata"
	"github.com/opentracing/opentracing-go/ext"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/UnderTreeTech/waterdrop/pkg/server/http/config"

	"github.com/gin-gonic/gin"
)

// Trace trace incoming request details.
//
// Starts the opentracing/jaeger server span (unchanged) and, when OTel is
// enabled (otel.Enabled(), set by [trace.otel] in jaeger.Init), also
// starts an OTel server span extracted from the composite propagator
// (traceparent + uber-trace-id). The two spans coexist on the context via
// different context keys. When OTel is enabled the X-Trace-Id response header
// is overwritten with the OTel TraceID so the outward id matches the
// langfuse/OTel trace; otherwise it stays the jaeger id (current behavior).
func Trace(config *config.ServerConfig) gin.HandlerFunc {
	tracer := otel.Tracer("waterdrop/http")
	return func(c *gin.Context) {
		// ---- jaeger span (unchanged) ----
		span, ctx := trace.StartSpanFromContext(
			c.Request.Context(),
			c.Request.Method+" "+c.Request.URL.Path,
			trace.HeaderExtractor(c.Request.Header),
		)
		ext.Component.Set(span, "http")
		ext.SpanKind.Set(span, ext.SpanKindRPCServerEnum)
		ext.HTTPMethod.Set(span, c.Request.Method)
		ext.HTTPUrl.Set(span, c.FullPath())
		ext.PeerHostIPv4.SetString(span, c.ClientIP())

		// adjust request timeout
		timeout := config.Timeout
		reqTimeout := metadata.GetTimeout(c.Request)
		if reqTimeout > 0 && timeout > reqTimeout {
			timeout = reqTimeout
		}

		// if zero timeout config means never timeout
		var cancel func()
		if timeout > 0 {
			ctx, cancel = context.WithTimeout(ctx, timeout)
		} else {
			cancel = func() {}
		}

		// ---- OTel span (only when enabled) ----
		var ospan oteltrace.Span
		if otel.Enabled() {
			octx := otel.Propagator().Extract(ctx, propagation.HeaderCarrier(c.Request.Header))
			octx, ospan = tracer.Start(octx, c.Request.Method+" "+c.FullPath(),
				oteltrace.WithSpanKind(oteltrace.SpanKindServer),
				oteltrace.WithAttributes(
					semconv.HTTPRequestMethodKey.String(c.Request.Method),
					semconv.URLPath(c.Request.URL.Path),
					semconv.HTTPRoute(c.FullPath()),
				),
			)
			ctx = octx // stack OTel span on top of jaeger ctx (different keys, both kept)
		}

		defer func() {
			if ospan != nil {
				ospan.End()
			}
			span.Finish()
			cancel()
		}()

		c.Request = c.Request.WithContext(ctx)

		// X-Trace-Id: OTel id when enabled (matches langfuse trace), else jaeger id.
		if ospan != nil {
			c.Writer.Header().Set(metadata.HeaderHttpTraceId, ospan.SpanContext().TraceID().String())
		} else {
			c.Writer.Header().Set(metadata.HeaderHttpTraceId, trace.TraceID(ctx))
		}

		c.Next()

		if ospan != nil {
			status := c.Writer.Status()
			ospan.SetAttributes(semconv.HTTPResponseStatusCode(status))
			if status >= http.StatusInternalServerError {
				ospan.SetStatus(codes.Error, http.StatusText(status))
			}
		}
	}
}
