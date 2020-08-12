package http

import (
	"context"

	"github.com/UnderTreeTech/waterdrop/pkg/trace"
	"github.com/gin-gonic/gin"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

func Trace(conf *Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// start trace
		span, ctx := trace.StartSpanFromContext(c.Request.Context(),
			c.Request.Method+" "+c.Request.URL.Path,
			trace.HeaderExtractor(opentracing.HTTPHeadersCarrier(c.Request.Header)))
		ext.Component.Set(span, "http")
		ext.SpanKind.Set(span, ext.SpanKindRPCServerEnum)
		ext.HTTPMethod.Set(span, c.Request.Method)
		ext.HTTPUrl.Set(span, c.Request.URL.Path)
		ext.PeerHostIPv4.SetString(span, c.ClientIP())

		// adjust request timeout
		timeout := conf.Timeout
		reqTimeout := getTimeout(c.Request)
		if reqTimeout > 0 && conf.Timeout > reqTimeout {
			timeout = reqTimeout
		}

		ctx, cancel := context.WithTimeout(ctx, timeout)

		defer func() {
			span.Finish()
			cancel()
		}()

		c.Request = c.Request.WithContext(ctx)
		c.Writer.Header().Set(_httpHeaderTraceId, trace.TraceID(ctx))

		c.Next()
	}
}
