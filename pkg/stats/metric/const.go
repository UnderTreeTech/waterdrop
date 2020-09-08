package metric

const (
	_httpClientNamespace = "http_client"
	_httpServerNamespace = "http_server"

	_unaryClientNamespace = "unary_client"
	_unaryServerNamespace = "unary_server"

	_streamClientNamespace = "stream_client"
	_streamServerNamespace = "stream_server"

	_redisClientNamespace = "redis"

	_mysqlClientNamespace = "mysql"

	_rocketmqClientNamespace = "rocketmq"
	_kafkaClientNamespace    = "kafka"
)

// http metrics
var (
	HTTPClientReqDuration = NewHistogramVec(&HistogramVecOpts{
		Namespace: _httpClientNamespace,
		Subsystem: "requests",
		Name:      "duration_ms",
		Help:      "http client requests duration(ms).",
		Labels:    []string{"path", "method", "peer"},
		Buckets:   []float64{5, 10, 25, 50, 100, 250, 500, 1000},
	})

	HTTPClientHandleCounter = NewCounterVec(&CounterVecOpts{
		Namespace: _httpClientNamespace,
		Subsystem: "requests",
		Name:      "code_total",
		Help:      "http client requests code count.",
		Labels:    []string{"path", "method", "peer", "code"},
	})

	HTTPServerReqDuration = NewHistogramVec(&HistogramVecOpts{
		Namespace: _httpServerNamespace,
		Subsystem: "requests",
		Name:      "duration_ms",
		Help:      "http server requests duration(ms).",
		Labels:    []string{"path", "method", "peer"},
		Buckets:   []float64{5, 10, 25, 50, 100, 250, 500, 1000},
	})

	HTTPServerHandleCounter = NewCounterVec(&CounterVecOpts{
		Namespace: _httpServerNamespace,
		Subsystem: "requests",
		Name:      "code_total",
		Help:      "http server requests error count.",
		Labels:    []string{"path", "method", "peer", "code"},
	})
)

// unary metrics
var (
	UnaryClientReqDuration = NewHistogramVec(&HistogramVecOpts{
		Namespace: _unaryClientNamespace,
		Subsystem: "requests",
		Name:      "duration_ms",
		Help:      "unary client requests duration(ms).",
		Labels:    []string{"peer", "method"},
		Buckets:   []float64{5, 10, 25, 50, 100, 250, 500, 1000},
	})

	UnaryClientHandleCounter = NewCounterVec(&CounterVecOpts{
		Namespace: _unaryClientNamespace,
		Subsystem: "requests",
		Name:      "code_total",
		Help:      "unary client requests code count.",
		Labels:    []string{"peer", "method", "code"},
	})

	UnaryServerReqDuration = NewHistogramVec(&HistogramVecOpts{
		Namespace: _unaryServerNamespace,
		Subsystem: "requests",
		Name:      "duration_ms",
		Help:      "unary server requests duration(ms).",
		Labels:    []string{"peer", "method"},
		Buckets:   []float64{5, 10, 25, 50, 100, 250, 500, 1000},
	})

	UnaryServerHandleCounter = NewCounterVec(&CounterVecOpts{
		Namespace: _unaryServerNamespace,
		Subsystem: "requests",
		Name:      "code_total",
		Help:      "unary server requests error count.",
		Labels:    []string{"peer", "method", "code"},
	})
)

// redis metrics
var (
	RedisClientReqDuration = NewHistogramVec(&HistogramVecOpts{
		Namespace: _redisClientNamespace,
		Subsystem: "requests",
		Name:      "duration_ms",
		Help:      "redis client requests duration(ms).",
		Labels:    []string{"name", "peer", "command"},
		Buckets:   []float64{5, 10, 25, 50, 100, 250, 500, 1000},
	})

	RedisClientErrCounter = NewCounterVec(&CounterVecOpts{
		Namespace: _redisClientNamespace,
		Subsystem: "requests",
		Name:      "error_total",
		Help:      "redis client requests error count.",
		Labels:    []string{"name", "peer", "command", "error"},
	})

	RedisHitCounter = NewCounterVec(&CounterVecOpts{
		Namespace: _redisClientNamespace,
		Subsystem: "requests",
		Name:      "hits_total",
		Help:      "redis client hits total.",
		Labels:    []string{"name", "peer", "command"},
	})

	RedisMissCounter = NewCounterVec(&CounterVecOpts{
		Namespace: _redisClientNamespace,
		Subsystem: "requests",
		Name:      "misses_total",
		Help:      "redis client misses total.",
		Labels:    []string{"name", "peer", "command"},
	})
)

// mysql metrics
var (
	MySQLClientReqDuration = NewHistogramVec(&HistogramVecOpts{
		Namespace: _mysqlClientNamespace,
		Subsystem: "requests",
		Name:      "duration_ms",
		Help:      "mysql client requests duration(ms).",
		Labels:    []string{"name", "addr", "command"},
		Buckets:   []float64{5, 10, 25, 50, 100, 250, 500, 1000},
	})
)

// rocketmq metrics
var (
	RocketMQClientReqDuration = NewHistogramVec(&HistogramVecOpts{
		Namespace: _rocketmqClientNamespace,
		Subsystem: "requests",
		Name:      "duration_ms",
		Help:      "rocketmq client requests duration(ms).",
		Labels:    []string{"name", "addr", "command"},
		Buckets:   []float64{5, 10, 25, 50, 100, 250, 500, 1000},
	})
)
