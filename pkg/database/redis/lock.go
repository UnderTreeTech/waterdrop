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
	"time"

	"github.com/UnderTreeTech/waterdrop/pkg/log"
	"github.com/UnderTreeTech/waterdrop/pkg/utils/xstring"
)

var (
	ErrUnLockGet = errors.New("lock has already unlocked or expired")
)

const (
	_lockTime   = 1000                  // px milliseconds
	_retry      = 2                     // retry times
	_retryDelay = 15 * time.Millisecond // retry interval, milliseconds
)

// lock lock resource
// ttl ms
// retry 重试次数
// retryDelay us
// 当前lock超时ttl为1s，重试2次，每15ms try-lock一次
// 请勿必只对关键业务操作路径进行加锁，缩小加锁操作范围
// 例如库存的select-for-update，保证锁操作能在15ms内完成
func (r *Redis) lock(ctx context.Context, key string, ttl int, retry int, retryDelay time.Duration) (locked bool, lockValue string, err error) {
	if retry <= 0 {
		retry = 1
	}

	//in case other client unlock current lock
	lockValue = "locked:" + xstring.RandomString(12)

	for retryTimes := 0; retryTimes < retry; retryTimes++ {
		locked, err = r.SetNxEx(ctx, key, lockValue, ttl)
		if err != nil {
			log.Error(ctx, "redis lock failed", log.String("key", key), log.String("error", err.Error()))
			break
		}

		if locked {
			break
		}
		time.Sleep(retryDelay)
	}

	return
}

// unLock unlock resource
func (r *Redis) unLock(ctx context.Context, key string, lockValue string) (err error) {
	reply, err := r.Get(ctx, key)
	if err != nil {
		return
	}

	if reply != lockValue {
		err = ErrUnLockGet
		log.Error(ctx, "unlock fail",
			log.String("key", key),
			log.String("value", lockValue),
			log.String("error", err.Error()),
		)
		return
	}

	_, err = r.Del(ctx, key)
	return
}

// forceUnLock force unlock resource
func (r *Redis) forceUnLock(ctx context.Context, key string) (err error) {
	_, err = r.Del(ctx, key)
	return
}

// Lock lock resource
func (r *Redis) Lock(ctx context.Context, key string, expireTime ...int) (locked bool, lockValue string, err error) {
	lockTime := _lockTime
	if len(expireTime) > 0 {
		lockTime = expireTime[0]
	}

	locked, lockValue, err = r.lock(ctx, key, lockTime, _retry, _retryDelay)
	return
}

// UnLock unlock resource
func (r *Redis) UnLock(ctx context.Context, key, lockValue string) error {
	return r.unLock(ctx, key, lockValue)
}

// ForceUnLock force unlock resource
func (r *Redis) ForceUnLock(ctx context.Context, key string) error {
	return r.forceUnLock(ctx, key)
}
