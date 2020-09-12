package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/UnderTreeTech/waterdrop/pkg/stats/metric"

	"github.com/UnderTreeTech/waterdrop/pkg/log"

	"github.com/UnderTreeTech/waterdrop/pkg/trace"
	"github.com/opentracing/opentracing-go/ext"

	"github.com/gomodule/redigo/redis"
)

type Config struct {
	DBName        string
	DBIndex       int
	Addr          string
	Password      string
	MaxActive     int
	MaxIdle       int
	IdleTimeout   time.Duration
	DialTimeout   time.Duration
	ReadTimeout   time.Duration
	WriteTimeout  time.Duration
	SlowOpTimeout time.Duration
}

type Redis struct {
	pool *redis.Pool
	conf *Config
}

func New(conf *Config) *Redis {
	p := &redis.Pool{
		// Other pool configuration not shown in this example.
		Dial: func() (redis.Conn, error) {
			opts := make([]redis.DialOption, 0)
			opts = append(opts,
				redis.DialReadTimeout(conf.ReadTimeout),
				redis.DialWriteTimeout(conf.WriteTimeout),
				redis.DialConnectTimeout(conf.DialTimeout),
				redis.DialDatabase(conf.DBIndex),
				redis.DialPassword(conf.Password),
			)
			conn, err := redis.Dial("tcp", conf.Addr, opts...)
			if err != nil {
				return nil, err
			}
			return conn, nil
		},
		MaxIdle:     conf.MaxIdle,
		MaxActive:   conf.MaxActive,
		IdleTimeout: conf.IdleTimeout,
	}

	return &Redis{
		pool: p,
		conf: conf,
	}
}

func (r *Redis) Do(ctx context.Context, commandName string, args ...interface{}) (interface{}, error) {
	statement := r.getStatement(commandName, args...)
	span, ctx := trace.StartSpanFromContext(ctx, "redis."+commandName)
	span = span.SetTag("db.index", r.conf.DBIndex)
	ext.Component.Set(span, "redis")
	ext.SpanKind.Set(span, ext.SpanKindRPCClientEnum)
	ext.PeerAddress.Set(span, r.conf.Addr)
	ext.DBInstance.Set(span, r.conf.DBName)
	ext.DBStatement.Set(span, statement)

	conn, err := r.pool.GetContext(ctx)
	if err != nil {
		metric.RedisClientErrCounter.Inc(r.conf.DBName, r.conf.Addr, commandName, err.Error())
		return nil, err
	}

	now := time.Now()
	defer func() {
		conn.Close()
		span.Finish()
		r.slowLog(ctx, statement, now)
	}()

	reply, err := conn.Do(commandName, args...)
	if err != nil {
		metric.RedisClientErrCounter.Inc(r.conf.DBName, r.conf.Addr, commandName, err.Error())
	}

	metric.RedisClientReqDuration.Observe(time.Since(now).Seconds(), r.conf.DBName, r.conf.Addr, commandName)

	return reply, err
}

func (r *Redis) Close() error {
	return r.pool.Close()
}

func (r *Redis) getStatement(commandName string, args ...interface{}) (res string) {
	res = commandName
	if len(args) > 0 {
		res = fmt.Sprintf("%s %v", commandName, args[0])
	}
	return
}

func (r *Redis) slowLog(ctx context.Context, statement string, now time.Time) {
	elapse := time.Since(now)
	if elapse > r.conf.SlowOpTimeout {
		log.Warn(ctx, "slow-redis-query", log.String("statement", statement), log.Duration("op_time", elapse))
	}
}
