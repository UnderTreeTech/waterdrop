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

package jaeger

import (
	"fmt"
	"os"
	"time"

	"github.com/UnderTreeTech/waterdrop/pkg/trace"
	"github.com/UnderTreeTech/waterdrop/pkg/trace/otel"

	"github.com/UnderTreeTech/waterdrop/pkg/conf"

	opentracing "github.com/opentracing/opentracing-go"
	jconfig "github.com/uber/jaeger-client-go/config"
)

// JaegerConfig
type JaegerConfig struct {
	ServiceName      string
	Sampler          *jconfig.SamplerConfig
	Reporter         *jconfig.ReporterConfig
	EnableRPCMetrics bool
	options          []jconfig.Option
}

// Config
type Config struct {
	ServiceName      string
	EnableRPCMetrics bool

	SamplerType  string
	SamplerParam float64

	AgentAddr                   string
	ReporterLogSpans            bool
	ReporterBufferFlushInterval time.Duration

	TraceBaggageHeaderPrefix string
	TraceContextHeaderName   string

	MaxTagValueLength int
}

func defaultJaegerConfig() *JaegerConfig {
	agentAddr := "127.0.0.1:6831"
	if addr := os.Getenv("JAEGER_AGENT_ADDR"); addr != "" {
		agentAddr = addr
	}
	hostname, _ := os.Hostname()
	return &JaegerConfig{
		ServiceName: hostname,
		Sampler: &jconfig.SamplerConfig{
			Type:  "const",
			Param: 0.001,
		},
		Reporter: &jconfig.ReporterConfig{
			LogSpans:            false,
			BufferFlushInterval: 1 * time.Second,
			LocalAgentHostPort:  agentAddr,
		},
		EnableRPCMetrics: true,
	}
}

// WithOption apply jaeger option
func (jc *JaegerConfig) WithOption(options ...jconfig.Option) *JaegerConfig {
	if jc.options == nil {
		jc.options = make([]jconfig.Option, 0)
	}
	jc.options = append(jc.options, options...)
	return jc
}

func newJaegerClient(traceConf *JaegerConfig) (opentracing.Tracer, func()) {
	var configuration = jconfig.Configuration{
		ServiceName: traceConf.ServiceName,
		Sampler:     traceConf.Sampler,
		Reporter:    traceConf.Reporter,
		RPCMetrics:  traceConf.EnableRPCMetrics,
	}

	tracer, closer, err := configuration.NewTracer(traceConf.options...)
	if err != nil {
		panic(fmt.Sprintf("new jaeger trace fail, err msg %s", err.Error()))
	}

	return tracer, func() { closer.Close() }
}

// buildJaeger builds the opentracing/jaeger tracer from a parsed [trace.jaeger]
// config and sets it global. Returns the shutdown func.
func buildJaeger(jconf *Config) func() {
	traceConf := &JaegerConfig{}
	sampler := &jconfig.SamplerConfig{}
	sampler.Type = jconf.SamplerType
	sampler.Param = jconf.SamplerParam
	reporter := &jconfig.ReporterConfig{}
	reporter.LocalAgentHostPort = jconf.AgentAddr
	reporter.LogSpans = jconf.ReporterLogSpans
	reporter.BufferFlushInterval = jconf.ReporterBufferFlushInterval
	traceConf.ServiceName = jconf.ServiceName
	traceConf.EnableRPCMetrics = jconf.EnableRPCMetrics
	traceConf.Sampler = sampler
	traceConf.Reporter = reporter
	traceConf.WithOption(jconfig.MaxTagValueLength(jconf.MaxTagValueLength))

	tracer, close := newJaegerClient(traceConf)
	trace.SetGlobalTracer(tracer)
	return close
}

// Init initializes tracing from config. It is the single entry point services
// call (defer jaeger.Init()()), and decides what to start based on which config
// sections are present, so services opt in purely via config:
//
//   - [trace.jaeger] only          -> pure jaeger (UDP agent). Current behavior.
//   - [trace.otel] only            -> pure OTel (OTLP HTTP). New services.
//   - both                          -> dual-track (jaeger + OTel, shared TraceID).
//   - neither                       -> default jaeger (forward-compatible fallback).
//
// OTel is initialized via otel.InitWithConfig using the [trace.otel] section.
// The OTel global propagator is the composite (W3C traceparent + uber-trace-id
// + Baggage) so OTel and jaeger share a TraceID without a bridge.
func Init() func() {
	// Read optional [trace.otel] section.
	oconf := &otel.Config{}
	hasOtel := conf.Unmarshal("trace.otel", oconf) == nil
	otelEnabled := hasOtel && oconf.Enable

	// Read optional [trace.jaeger] section.
	jconf := &Config{}
	errJ := conf.Unmarshal("trace.jaeger", jconf)

	var jaegerClose func()
	switch {
	case errJ == nil:
		// [trace.jaeger] present -> build jaeger.
		jaegerClose = buildJaeger(jconf)
	case !otelEnabled:
		// Neither configured -> default jaeger (forward-compatible fallback).
		tracer, close := newJaegerClient(defaultJaegerConfig())
		trace.SetGlobalTracer(tracer)
		jaegerClose = close
	}
	// else: pure OTel (no jaeger tracer; opentracing global stays default noop,
	// the jaeger branches in the Trace middleware / gRPC interceptors run as noop).

	var otelClose func()
	if otelEnabled {
		// Default OTel service name to the jaeger service name when omitted.
		if oconf.ServiceName == "" && errJ == nil {
			oconf.ServiceName = jconf.ServiceName
		}
		otelClose = otel.InitWithConfig(oconf)
	}

	return func() {
		otelClose()
		if jaegerClose != nil {
			jaegerClose()
		}
	}
}
