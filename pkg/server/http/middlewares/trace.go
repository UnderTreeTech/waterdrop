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
	"context"
	"net/http"

	"github.com/UnderTreeTech/waterdrop/pkg/trace"

	"github.com/UnderTreeTech/waterdrop/pkg/server/http/metadata"

	"github.com/gin-gonic/gin"

	"github.com/UnderTreeTech/waterdrop/pkg/server/http/config"
)

// Trace trace incoming request details.
//
// Starts a server span via the factory-selected trace.Tracer (jaeger / otel /
// dual, decided by config in trace.Init). It extracts the parent from the
// incoming HTTP headers, sets standard HTTP attributes, and writes the span's
// TraceID to the X-Trace-Id response header. The backend-specific logic lives
// in the Tracer implementation; this middleware holds no jaeger/OTel branches.
func Trace(config *config.ServerConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		carrier := trace.HTTPCarrier{Header: c.Request.Header}
		name := c.Request.Method + " " + c.Request.URL.Path
		span, ctx := trace.GlobalTracer().StartServerSpan(c.Request.Context(), name, trace.SpanServer, carrier)

		// standard HTTP attributes
		span.SetStringAttr("http.method", c.Request.Method)
		span.SetStringAttr("http.url", c.FullPath())
		span.SetStringAttr("http.route", c.FullPath())
		span.SetStringAttr("http.client_ip", c.ClientIP())

		// adjust request timeout: framework timeout clamped down by the
		// X-Request-Timeout header; zero means never timeout.
		timeout := config.Timeout
		if reqTimeout := metadata.GetTimeout(c.Request); reqTimeout > 0 && timeout > reqTimeout {
			timeout = reqTimeout
		}
		var cancel func()
		if timeout > 0 {
			ctx, cancel = context.WithTimeout(ctx, timeout)
		} else {
			cancel = func() {}
		}

		defer func() {
			span.End()
			cancel()
		}()

		c.Request = c.Request.WithContext(ctx)

		// X-Trace-Id: the active backend's TraceID (jaeger 64-bit or OTel
		// 128-bit), so the outward id matches the trace the backend records.
		c.Writer.Header().Set(metadata.HeaderHttpTraceId, span.TraceID())

		c.Next()

		status := c.Writer.Status()
		span.SetIntAttr("http.status_code", int64(status))
		if status >= http.StatusInternalServerError {
			span.SetStatus(false, http.StatusText(status))
		}
	}
}
