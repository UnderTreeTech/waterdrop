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

package trace

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// initModeWithExporter lives in e2e_test.go (same package).

// dbBrokerModes runs the db/broker coverage across all three backends. otel
// and dual assert via an in-memory exporter; jaeger asserts structurally (jaeger
// spans are not capturable in-process, but correctness follows from the shared
// opentracing API + uber-trace-id propagation validated elsewhere).
var dbBrokerModes = []string{"jaeger", "otel", "dual"}

// TestDBBrokerSpanBasics drives the exact opentracing API the db/broker layers
// use (StartSpanFromContext + ext.* tags + SetTag + LogFields + ext.Error) and
// asserts, per mode:
//   - the span is non-nil (so ext.* / SetTag never nil-derefs)
//   - trace.TraceID on the db-span ctx is non-empty (log trace_id prints)
//   - the db span shares the server span's TraceID (correct parent linking)
//
// For otel/dual it additionally asserts the span was recorded by the OTel
// exporter with the expected attributes and a slow_query event.
func TestDBBrokerSpanBasics(t *testing.T) {
	for _, mode := range dbBrokerModes {
		t.Run(mode, func(t *testing.T) {
			defer resetGlobal()
			closeFn, exp := initModeWithExporter(t, mode)
			defer closeFn()

			// server span (the middleware would start this)
			srvSpan, srvCtx := GlobalTracer().StartServerSpan(context.Background(), "http.server", SpanServer, nil)
			srvTID := TraceID(srvCtx)
			require.NotEmpty(t, srvTID, "server span has a trace id")

			// --- mimic the db/broker layer (mongo collection.Aggregate etc.) ---
			span, dbCtx := StartSpanFromContext(srvCtx, "aggregate")
			require.NotNil(t, span, "StartSpanFromContext returns a non-nil span")
			// the exact ext.* tags db/broker code sets:
			ext.PeerAddress.Set(span, "127.0.0.1:27017")
			ext.DBType.Set(span, "mongo")
			ext.SpanKind.Set(span, ext.SpanKindRPCClientEnum)
			ext.DBInstance.Set(span, "jarvis")
			span.SetTag("db.collection", "t_user")
			// slow_query logging path (mongo aggregate.go / query.go)
			span.LogFields(log.String("event", "slow_query"), log.Int64("elapse", 42))
			// error path (mongo aggregate.go)
			ext.Error.Set(span, true)

			dbTID := TraceID(dbCtx)
			span.Finish()
			srvSpan.End()

			// non-empty + same trace as the server span (parent linking)
			assert.NotEmpty(t, dbTID, "db span trace id non-empty (log trace_id prints)")
			assertTraceIDsSameTrace(t, srvTID, dbTID)

			// otel/dual: the db span reached the OTel backend with full data
			if exp != nil {
				spans := exp.GetSpans()
				require.Len(t, spans, 2, "server + db span recorded")
				var dbSpan *tracetest.SpanStub
				for i := range spans {
					if spans[i].Name == "aggregate" {
						dbSpan = &spans[i]
					}
				}
				require.NotNil(t, dbSpan, "db span recorded by otel exporter")
				assertAttr(t, dbSpan.Attributes, "peer.address", "127.0.0.1:27017")
				assertAttr(t, dbSpan.Attributes, "db.type", "mongo")
				assertAttr(t, dbSpan.Attributes, "db.instance", "jarvis")
				assertAttr(t, dbSpan.Attributes, "db.collection", "t_user")
				// ext.Error.Set -> bridge maps to otel span status (codes.Error),
				// not an "error" attribute.
				assert.Equal(t, otelcodes.Error, dbSpan.Status.Code, "ext.Error -> otel error status")
				// slow_query LogFields -> otel event with event=slow_query attr
				require.NotEmpty(t, dbSpan.Events, "LogFields recorded as otel event")
				assertAttr(t, dbSpan.Events[0].Attributes, "event", "slow_query")
				// parent TraceID == server TraceID
				assert.Equal(t, normalizeTID(srvTID), normalizeTID(dbSpan.SpanContext.TraceID().String()))
				// NOTE: ext.SpanKind.Set is called AFTER span creation; the bridge
				// ignores post-creation span.kind (TODO in bridgeSpan.SetTag), so
				// dbSpan.SpanKind stays Internal. This is a known, cosmetic bridge
				// limitation; it does not affect trace continuity or trace_id.
			}
		})
	}
}

// TestDBBrokerRedisPattern covers the redis hook pattern: Component="redis",
// SpanKind=client, and a command-name operation. Same assertions as the mongo
// case, proving the pattern works across db drivers.
func TestDBBrokerRedisPattern(t *testing.T) {
	for _, mode := range dbBrokerModes {
		t.Run(mode, func(t *testing.T) {
			defer resetGlobal()
			closeFn, exp := initModeWithExporter(t, mode)
			defer closeFn()

			srvSpan, srvCtx := GlobalTracer().StartServerSpan(context.Background(), "http.server", SpanServer, nil)
			srvTID := TraceID(srvCtx)

			// redis hook (redis.go:199): StartSpanFromContext(cmd.Name())
			span, dbCtx := StartSpanFromContext(srvCtx, "GET")
			ext.Component.Set(span, "redis")
			ext.SpanKind.Set(span, ext.SpanKindRPCClientEnum)
			ext.PeerAddress.Set(span, "127.0.0.1:6379")
			ext.DBInstance.Set(span, "0")
			dbTID := TraceID(dbCtx)
			span.Finish()
			srvSpan.End()

			assert.NotEmpty(t, dbTID)
			assertTraceIDsSameTrace(t, srvTID, dbTID)

			if exp != nil {
				spans := exp.GetSpans()
				var redis *tracetest.SpanStub
				for i := range spans {
					if spans[i].Name == "GET" {
						redis = &spans[i]
					}
				}
				require.NotNil(t, redis)
				assertAttr(t, redis.Attributes, "component", "redis")
			}
		})
	}
}

// TestDBBrokerSQLTxPattern covers the sql Tx path (sql.go): a span is held on a
// struct (opentracing.Span field) and tagged with Component=driverName. Proves
// the "span stored on a struct, tagged later" pattern works.
func TestDBBrokerSQLTxPattern(t *testing.T) {
	for _, mode := range dbBrokerModes {
		t.Run(mode, func(t *testing.T) {
			defer resetGlobal()
			closeFn, exp := initModeWithExporter(t, mode)
			defer closeFn()

			srvSpan, srvCtx := GlobalTracer().StartServerSpan(context.Background(), "http.server", SpanServer, nil)
			srvTID := TraceID(srvCtx)

			// sql conn.begin (sql.go:346): span held on struct, Component=driverName
			span, dbCtx := StartSpanFromContext(srvCtx, "conn.transaction")
			ext.PeerAddress.Set(span, "127.0.0.1:3306")
			ext.Component.Set(span, "mysql")
			ext.SpanKind.Set(span, ext.SpanKindRPCClientEnum)
			ext.DBInstance.Set(span, "db_jarvis")
			// simulate a nested exec within the tx using the same ctx
			execSpan, _ := StartSpanFromContext(dbCtx, "conn.exec")
			ext.Component.Set(execSpan, "mysql")
			execSpan.Finish()
			dbTID := TraceID(dbCtx)
			span.Finish()
			srvSpan.End()

			assert.NotEmpty(t, dbTID)
			assertTraceIDsSameTrace(t, srvTID, dbTID)

			if exp != nil {
				spans := exp.GetSpans()
				var tx, exec *tracetest.SpanStub
				for i := range spans {
					switch spans[i].Name {
					case "conn.transaction":
						tx = &spans[i]
					case "conn.exec":
						exec = &spans[i]
					}
				}
				require.NotNil(t, tx)
				require.NotNil(t, exec, "nested exec span recorded")
				assertAttr(t, tx.Attributes, "component", "mysql")
				// nested exec is a child of tx -> same trace
				assert.Equal(t, normalizeTID(srvTID), normalizeTID(exec.SpanContext.TraceID().String()))
			}
		})
	}
}

// TestDBBrokerRocketMQPattern covers the rocketmq consumer/producer pattern:
// FromIncomingContext (extract) + StartSpanFromContext + ext.Component="MQ" +
// SpanKind=consumer, plus TraceID for the producer's log.
func TestDBBrokerRocketMQPattern(t *testing.T) {
	for _, mode := range dbBrokerModes {
		t.Run(mode, func(t *testing.T) {
			defer resetGlobal()
			closeFn, exp := initModeWithExporter(t, mode)
			defer closeFn()

			// upstream: a server span whose context is "incoming"
			srvSpan, srvCtx := GlobalTracer().StartServerSpan(context.Background(), "http.server", SpanServer, nil)
			srvTID := TraceID(srvCtx)

			// rocketmq interceptor.go:87 - FromIncomingContext extracts parent
			opt := FromIncomingContext(srvCtx)
			span, mqCtx := StartSpanFromContext(srvCtx, "consumer:order_topic", opt)
			ext.Component.Set(span, "MQ")
			ext.SpanKind.Set(span, ext.SpanKindConsumerEnum)
			mqTID := TraceID(mqCtx)
			span.Finish()
			srvSpan.End()

			assert.NotEmpty(t, mqTID)
			assertTraceIDsSameTrace(t, srvTID, mqTID)

			if exp != nil {
				spans := exp.GetSpans()
				var mq *tracetest.SpanStub
				for i := range spans {
					if spans[i].Name == "consumer:order_topic" {
						mq = &spans[i]
					}
				}
				require.NotNil(t, mq)
				assertAttr(t, mq.Attributes, "component", "MQ")
			}
		})
	}
}

// TestDBBrokerTraceIDInLogContext asserts trace.TraceID(ctx) (used by log.go's
// assembleFields) returns the db span's trace id inside a db operation, so log
// lines emitted from within db/broker code carry the correct trace_id.
func TestDBBrokerTraceIDInLogContext(t *testing.T) {
	for _, mode := range dbBrokerModes {
		t.Run(mode, func(t *testing.T) {
			defer resetGlobal()
			closeFn, _ := initModeWithExporter(t, mode)
			defer closeFn()

			srvSpan, srvCtx := GlobalTracer().StartServerSpan(context.Background(), "http.server", SpanServer, nil)
			srvTID := TraceID(srvCtx)

			_, dbCtx := StartSpanFromContext(srvCtx, "mongo.query")
			// log.go would call trace.TraceID(dbCtx) here
			dbTID := TraceID(dbCtx)
			srvSpan.End()

			assert.NotEmpty(t, dbTID, "log trace_id non-empty inside db op")
			assertTraceIDsSameTrace(t, srvTID, dbTID)
		})
	}
}

// TestDBBrokerSpanFromContextFacade asserts trace.SpanFromContext (used by
// MetadataInjector/HeaderInjector to read the current span) returns a non-nil
// span inside a db op, in all modes.
func TestDBBrokerSpanFromContextFacade(t *testing.T) {
	for _, mode := range dbBrokerModes {
		t.Run(mode, func(t *testing.T) {
			defer resetGlobal()
			closeFn, _ := initModeWithExporter(t, mode)
			defer closeFn()

			_, ctx := StartSpanFromContext(context.Background(), "op")
			sp := SpanFromContext(ctx)
			assert.NotNil(t, sp, "SpanFromContext non-nil inside a span (injectors depend on this)")
		})
	}
}

// TestDBBrokerNilContextSafe asserts calling StartSpanFromContext on a bare
// context.Background() (no parent span, e.g. a background/cron job with no
// request span) does not panic and still records a span in any mode. In
// jaeger mode the standalone span also yields a TraceID; in otel/dual the span
// is recorded by the OTel backend (via the bridge) even though trace.TraceID
// returns "" (the bridge span is not on the OTel context key without an
// upstream server span) - the db/broker code itself is unaffected.
func TestDBBrokerNilContextSafe(t *testing.T) {
	for _, mode := range dbBrokerModes {
		t.Run(mode, func(t *testing.T) {
			defer resetGlobal()
			closeFn, exp := initModeWithExporter(t, mode)
			defer closeFn()

			span, ctx := StartSpanFromContext(context.Background(), "standalone")
			require.NotNil(t, span)
			ext.Component.Set(span, "mongo") // must not panic
			span.SetTag("db.collection", "x")
			span.LogFields(log.String("event", "slow_query"))
			span.Finish()

			if mode == "jaeger" {
				// jaeger: standalone span has a real trace id
				assert.NotEmpty(t, TraceID(ctx))
			}
			if exp != nil {
				// otel/dual: the standalone span reached the backend
				require.Len(t, exp.GetSpans(), 1, "standalone span recorded by otel backend")
			}
		})
	}
}

// assertTraceIDsSameTrace asserts two trace ids belong to one trace, accounting
// for the jaeger(16 hex) <-> otel(32 hex) left-pad when crossing backends.
func assertTraceIDsSameTrace(t *testing.T, a, b string) {
	t.Helper()
	require.NotEmpty(t, a)
	require.NotEmpty(t, b)
	assert.Equal(t, normalizeTID(a), normalizeTID(b), "same trace (after left-pad normalization)")
}

// normalizeTID left-pads a trace id to 32 hex so jaeger 64-bit and otel
// 128-bit ids compare equal when they are the same trace.
func normalizeTID(s string) string {
	if len(s) < 32 {
		pad := 32 - len(s)
		bytes := make([]byte, 0, 32)
		for i := 0; i < pad; i++ {
			bytes = append(bytes, '0')
		}
		bytes = append(bytes, s...)
		return string(bytes)
	}
	return s
}

// assertAttr asserts a string attribute key/value pair exists in the slice.
func assertAttr(t *testing.T, attrs []attribute.KeyValue, key, val string) {
	t.Helper()
	for _, a := range attrs {
		if string(a.Key) == key {
			assert.Equal(t, val, a.Value.AsString(), "attr %s", key)
			return
		}
	}
	t.Errorf("attr %q not found in %v", key, attrs)
}
