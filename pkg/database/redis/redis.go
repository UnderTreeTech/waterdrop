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

package redis

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/UnderTreeTech/waterdrop/pkg/stats/metric"

	"github.com/go-redis/redis/v8"

	"github.com/UnderTreeTech/waterdrop/pkg/breaker"

	"github.com/UnderTreeTech/waterdrop/pkg/log"

	"github.com/UnderTreeTech/waterdrop/pkg/trace"
	"github.com/opentracing/opentracing-go/ext"
)

const (
	defaultSlowQueryTime = time.Millisecond * 100
	defaultMinIdleConns  = 10
)

var (
	// reference refer to current redis config
	reference *Config
	// addrErr redis address error
	addrErr = errors.New("you must assign at least one redis address")
	// nodeErr node mode error
	nodeErr = errors.New("node mode only support one address")
	// modeErr redis deploy mode error
	modeErr = errors.New("unsupported redis mode")
)

// Config redis configs
type Config struct {
	// DBName metric name
	DBName string
	// DBIndex redis db index
	DBIndex int
	// Addr redis address
	Addr []string
	// dbAddr metric addr
	dbAddr string
	// Password auth password
	Password string
	// DeployMode deploy mode
	DeployMode string
	// MasterName master name for sentinel mode
	MasterName string
	// MinIdleConns min idle connections
	MinIdleConns int
	// DialTimeout dial redis timeout
	DialTimeout time.Duration
	// ReadTimeout read time out
	ReadTimeout time.Duration
	// WriteTimeout write time out
	WriteTimeout time.Duration
	// SlowOpTimeout slow query threshold
	SlowOpTimeout time.Duration
}

// Redis redis instance
type Redis struct {
	client   redis.UniversalClient
	config   *Config
	breakers *breaker.BreakerGroup
}

type (
	// Pipeliner is an alias of redis.Pipeliner
	Pipeliner = redis.Pipeliner
	// IntCmd is an alias of redis.IntCmd
	IntCmd = redis.IntCmd
	// FloatCmd is an alias of redis.FloatCmd
	FloatCmd = redis.FloatCmd
	// StringCmd is an alias of redis.StringCmd
	StringCmd = redis.StringCmd
	// StringStringMapCmd is an alias of redis.StringStringMapCmd
	StringStringMapCmd = redis.StringStringMapCmd
	// StringSliceCmd is an alias of redis.StringSliceCmd
	StringSliceCmd = redis.StringSliceCmd
	// IntSliceCmd is an alias of redis.IntSliceCmd
	IntSliceCmd = redis.IntSliceCmd
)

// New returns a redis instance according deploy mode. There are three deploy mode.
// node: standalone
// sentinel: master-slave failover sentinel
// cluster: cluster mode
func New(cfg *Config) (rdb *Redis, err error) {
	if len(cfg.Addr) <= 0 {
		return nil, addrErr
	}

	if cfg.SlowOpTimeout <= 0 {
		cfg.SlowOpTimeout = defaultSlowQueryTime
	}

	if cfg.MinIdleConns <= 0 {
		cfg.MinIdleConns = defaultMinIdleConns
	}

	if cfg.DBName == "" {
		cfg.DBName = "default"
	}

	cfg.dbAddr = strings.Join(cfg.Addr, "")
	opts := &redis.UniversalOptions{}
	opts.DB = cfg.DBIndex
	opts.Addrs = cfg.Addr
	opts.Password = cfg.Password
	opts.MinIdleConns = cfg.MinIdleConns
	opts.DialTimeout = cfg.DialTimeout
	opts.ReadTimeout = cfg.ReadTimeout
	opts.WriteTimeout = cfg.WriteTimeout

	switch cfg.DeployMode {
	case "node":
		if len(cfg.Addr) > 1 {
			return nil, nodeErr
		}
	case "sentinel":
		opts.MasterName = cfg.MasterName
	case "cluster":
	default:
		return nil, modeErr
	}

	reference = cfg
	uc := redis.NewUniversalClient(opts)
	uc.AddHook(redisHook{})
	rdb = &Redis{
		client:   uc,
		config:   cfg,
		breakers: breaker.NewBreakerGroup(),
	}
	return
}

// accept check request success or not
func accept(err error) bool {
	return err == nil || err == redis.Nil
}

type (
	// timeKey time context key
	timeKey struct{}
	// redisHook to hack in trace and metric stats
	redisHook struct{}
)

// BeforeProcess pre handler before process
func (r redisHook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	// init span
	span, ctx := trace.StartSpanFromContext(ctx, cmd.Name())
	span = span.SetTag("db.index", reference.DBIndex)
	ext.Component.Set(span, "redis")
	ext.SpanKind.Set(span, ext.SpanKindRPCClientEnum)
	ext.PeerAddress.Set(span, reference.dbAddr)
	ext.DBInstance.Set(span, reference.DBName)

	// record current time
	start := time.Now()
	ctx = context.WithValue(ctx, timeKey{}, start)
	return ctx, nil
}

// AfterProcess post handler after process
func (r redisHook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	// check if it's a slow query
	start := ctx.Value(timeKey{}).(time.Time)
	elapse := time.Since(start)
	if elapse > reference.SlowOpTimeout {
		log.Warn(ctx, "slow-redis-query", log.String("statement", cmd.String()), log.Duration("op_time", elapse))
	}

	// finish span
	if span := trace.SpanFromContext(ctx); span != nil {
		span.Finish()
	}

	// metric error and hit/miss
	if cmd.Err() != nil {
		metric.RedisMissCounter.Inc(reference.DBName, reference.dbAddr, cmd.Name())
		if cmd.Err() != redis.Nil {
			metric.RedisClientErrCounter.Inc(reference.DBName, reference.dbAddr, cmd.Name(), cmd.Err().Error())
		}
	} else {
		metric.RedisHitCounter.Inc(reference.DBName, reference.dbAddr, cmd.Name())
	}
	metric.RedisClientReqDuration.Observe(elapse.Seconds(), reference.DBName, reference.dbAddr, cmd.Name())
	return nil
}

// BeforeProcessPipeline pre handler before process pipeline
func (r redisHook) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	// init span
	span, ctx := trace.StartSpanFromContext(ctx, "pipeline")
	span = span.SetTag("db.index", reference.DBIndex)
	ext.Component.Set(span, "redis")
	ext.SpanKind.Set(span, ext.SpanKindRPCClientEnum)
	ext.PeerAddress.Set(span, reference.dbAddr)
	ext.DBInstance.Set(span, reference.DBName)

	// record current time
	start := time.Now()
	ctx = context.WithValue(ctx, timeKey{}, start)
	return ctx, nil
}

// AfterProcessPipeline post handler after process pipeline
func (r redisHook) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	// check if it's a slow query
	start := ctx.Value(timeKey{}).(time.Time)
	elapse := time.Since(start)
	queryName := "pipeline"
	if elapse > reference.SlowOpTimeout {
		log.Warn(ctx, "slow-redis-query", log.String("statement", queryName), log.Duration("op_time", elapse))
	}

	// finish span
	if span := trace.SpanFromContext(ctx); span != nil {
		span.Finish()
	}

	// metric pipeline
	for _, cmd := range cmds {
		if cmd.Err() != nil && cmd.Err() != redis.Nil {
			metric.RedisClientErrCounter.Inc(reference.DBName, reference.dbAddr, cmd.Name(), cmd.Err().Error())
			break
		}
	}
	metric.RedisClientReqDuration.Observe(elapse.Seconds(), reference.DBName, reference.dbAddr, queryName)
	return nil
}
