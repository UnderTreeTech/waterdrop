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

package metric

const (
	_httpServerNamespace  = "http_server"
	_unaryServerNamespace = "unary_server"

	_redisClientNamespace = "redis"
	_mysqlClientNamespace = "mysql"

	_rocketmqClientNamespace = "rocketmq"
	_kafkaClientNamespace    = "kafka"
)

// http metrics
var (
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
	RocketMQClientHandleCounter = NewCounterVec(&CounterVecOpts{
		Namespace: _rocketmqClientNamespace,
		Subsystem: "requests",
		Name:      "code_total",
		Help:      "rocketmq client requests code count.",
		Labels:    []string{"peer", "type", "name", "command", "code"},
	})

	RocketMQClientReqDuration = NewHistogramVec(&HistogramVecOpts{
		Namespace: _rocketmqClientNamespace,
		Subsystem: "requests",
		Name:      "duration_ms",
		Help:      "rocketmq client requests duration(ms).",
		Labels:    []string{"peer", "type", "name", "command"},
		Buckets:   []float64{5, 10, 25, 50, 100, 250, 500, 1000},
	})
)

// kafka metrics
var (
	KafkaClientHandleCounter = NewCounterVec(&CounterVecOpts{
		Namespace: _kafkaClientNamespace,
		Subsystem: "requests",
		Name:      "code_total",
		Help:      "kafka client requests code count.",
		Labels:    []string{"peer", "type", "name", "command", "code"},
	})

	KafkaClientReqDuration = NewHistogramVec(&HistogramVecOpts{
		Namespace: _kafkaClientNamespace,
		Subsystem: "requests",
		Name:      "duration_ms",
		Help:      "kafka client requests duration(ms).",
		Labels:    []string{"peer", "type", "name", "command"},
		Buckets:   []float64{5, 10, 25, 50, 100, 250, 500, 1000},
	})
)
