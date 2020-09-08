package redis

import (
	"context"
	"errors"
	"time"

	"github.com/gomodule/redigo/redis"

	"github.com/UnderTreeTech/waterdrop/pkg/log"

	"github.com/UnderTreeTech/waterdrop/pkg/utils/xstring"
)

var (
	ErrUnLockGet = errors.New("lock has already unlocked or expired")
)

const (
	_lockTime   = 1000                  //px milliseconds
	_retry      = 2                     // retry times
	_retryDelay = 15 * time.Millisecond // retry interval, milliseconds
)

//ttl ms
//retry 重试次数
//retryDelay us
//当前lock超时ttl为1s，重试2次，每15ms try-lock一次
//请勿必只对关键业务操作路径进行加锁，缩小加锁操作范围
//例如库存的select-for-update，保证锁操作能在15ms内完成
func (r *Redis) lock(ctx context.Context, key string, ttl int, retry int, retryDelay time.Duration) (err error, locked bool, lockValue string) {
	if retry <= 0 {
		retry = 1
	}

	//in case other client unlock current lock
	lockValue = "locked:" + xstring.RandomString(12)

	for retryTimes := 0; retryTimes < retry; retryTimes++ {
		var res interface{}
		res, err = r.Do(ctx, "SET", key, lockValue, "PX", ttl, "NX")
		if err != nil {
			log.Error(ctx, "redis lock failed", log.String("key", key), log.String("error", err.Error()))
			break
		}

		if res != nil {
			locked = true
			break
		}

		time.Sleep(retryDelay)
	}

	return
}

func (r *Redis) unLock(ctx context.Context, key string, lockValue string) (err error) {
	res, err := redis.String(r.Do(ctx, "GET", key))
	if err != nil {
		return
	}

	if res != lockValue {
		err = ErrUnLockGet
		log.Error(ctx, "unlock fail",
			log.String("key", key),
			log.String("value", lockValue),
			log.String("error", err.Error()),
		)
		return
	}

	_, err = r.Do(ctx, "DEL", key)

	return
}

func (r *Redis) forceUnLock(ctx context.Context, key string) (err error) {
	_, err = r.Do(ctx, "DEL", key)

	return

}

func (r *Redis) Lock(ctx context.Context, key string) (err error, gotLock bool, lockValue string) {
	return r.lock(ctx, key, _lockTime, _retry, _retryDelay)
}

func (r *Redis) UnLock(ctx context.Context, key, lockValue string) error {
	return r.unLock(ctx, key, lockValue)
}

func (r *Redis) ForceUnLock(ctx context.Context, key string) error {
	return r.forceUnLock(ctx, key)
}
