package jaeger

import (
	"fmt"
	"os"
	"time"

	"github.com/UnderTreeTech/waterdrop/pkg/trace"

	"github.com/UnderTreeTech/waterdrop/pkg/conf"

	opentracing "github.com/opentracing/opentracing-go"
	jconfig "github.com/uber/jaeger-client-go/config"
)

// Config ...
type JaegerConfig struct {
	ServiceName      string
	Sampler          *jconfig.SamplerConfig
	Reporter         *jconfig.ReporterConfig
	EnableRPCMetrics bool
	options          []jconfig.Option
}

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

func Init(moduleName string) func() {
	traceConf := defaultJaegerConfig()
	if moduleName != "" {
		jconf := &Config{}
		err := conf.Unmarshal(moduleName, jconf)
		if err != nil {
			panic(fmt.Sprintf("reload server.http fail, err msg %s", err.Error()))
		}

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

		maxTagValueOpt := jconfig.MaxTagValueLength(jconf.MaxTagValueLength)
		traceConf.WithOption(maxTagValueOpt)
	}

	tracer, close := newJaegerClient(traceConf)
	trace.SetGlobalTracer(tracer)

	return close
}
