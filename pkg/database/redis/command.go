/*
 *
 * Copyright 2021 waterdrop authors.
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
	"strconv"
	"time"

	"github.com/spf13/cast"

	"github.com/go-redis/redis/v8"
)

// Close closes the client, releasing any open resources
func (r *Redis) Close(ctx context.Context) (err error) {
	err = r.client.Close()
	return
}

// Ping ping is used to test if a connection is still alive, or to measure latency
func (r *Redis) Ping(ctx context.Context) (value string, err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		return r.client.Ping(ctx).Err()
	}, accept)
	return
}

// Get get the value of key. If the key does not exist the special value nil is returned
func (r *Redis) Get(ctx context.Context, key string) (value string, err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		reply, rerr := r.client.Get(ctx, key).Result()
		if rerr != nil && rerr == redis.Nil {
			return nil
		}
		value = reply
		return rerr
	}, accept)
	return
}

// MGet the values of all specified keys
func (r *Redis) MGet(ctx context.Context, keys ...string) (value []string, err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		reply, rerr := r.client.MGet(ctx, keys...).Result()
		if rerr != nil {
			return rerr
		}
		value = cast.ToStringSlice(reply)
		return nil
	}, accept)
	return
}

// Exists returns if key exists
func (r *Redis) Exists(ctx context.Context, key string) (value bool, err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		num, err := r.client.Exists(ctx, key).Result()
		value = num == 1
		return err
	}, accept)
	return
}

// IncrBy increments the number stored at key by increment
func (r *Redis) IncrBy(ctx context.Context, key string, increment int64) (value int64, err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		total, err := r.client.IncrBy(ctx, key, increment).Result()
		value = total
		return err
	}, accept)
	return
}

// Expire set a seconds timeout on key
func (r *Redis) Expire(ctx context.Context, key string, seconds int) (err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		return r.client.Expire(ctx, key, time.Duration(seconds)*time.Second).Err()
	}, accept)
	return
}

// TTL returns the remaining time to live of a key that has a timeout
func (r *Redis) TTL(ctx context.Context, key string) (value time.Duration, err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		reply, rerr := r.client.TTL(ctx, key).Result()
		value = reply
		return rerr
	}, accept)
	return
}

// Del removes the specified keys, a key is ignored if it does not exist
func (r *Redis) Del(ctx context.Context, keys ...string) (num int64, err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		cmd := r.client.Del(ctx, keys...)
		num = cmd.Val()
		return cmd.Err()
	}, accept)
	return
}

// Set key to hold the string value
func (r *Redis) Set(ctx context.Context, key string, value string) (err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		return r.client.Set(ctx, key, value, 0).Err()
	}, accept)
	return
}

// MSet sets the given keys to their respective values
func (r *Redis) MSet(ctx context.Context, kvs map[string]string) (err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		return r.client.MSet(ctx, kvs).Err()
	}, accept)
	return
}

// SetEx set key to hold the string value and set key to timeout after a given number of seconds
func (r *Redis) SetEx(ctx context.Context, key string, value string, seconds int) (err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		return r.client.SetEX(ctx, key, value, time.Duration(seconds)*time.Second).Err()
	}, accept)
	return
}

// SetNxEx set key to hold string value if key does not exist and set key to timeout after a given number of milliseconds
func (r *Redis) SetNxEx(ctx context.Context, key string, value string, milliseconds int) (locked bool, err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		reply, rerr := r.client.SetNX(ctx, key, value, time.Duration(milliseconds)*time.Millisecond).Result()
		locked = reply
		return rerr
	}, accept)
	return
}

// HGetAll returns all fields and values of the hash stored at key
func (r *Redis) HGetAll(ctx context.Context, key string) (value map[string]string, err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		reply, rerr := r.client.HGetAll(ctx, key).Result()
		value = reply
		return rerr
	}, accept)
	return
}

// HGet returns the value associated with field in the hash stored at key
func (r *Redis) HGet(ctx context.Context, key string, field string) (value string, err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		reply, rerr := r.client.HGet(ctx, key, field).Result()
		if rerr != nil && rerr == redis.Nil {
			return nil
		}
		value = reply
		return rerr
	}, accept)
	return
}

// HMGet returns the values associated with the specified fields in the hash stored at key
func (r *Redis) HMGet(ctx context.Context, key string, fields ...string) (value []string, err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		reply, rerr := r.client.HMGet(ctx, key, fields...).Result()
		value = cast.ToStringSlice(reply)
		return rerr
	}, accept)
	return
}

// HKeys returns all field names in the hash stored at key
func (r *Redis) HKeys(ctx context.Context, key string) (value []string, err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		reply, rerr := r.client.HKeys(ctx, key).Result()
		value = reply
		return rerr
	}, accept)
	return
}

// HLen returns the number of fields contained in the hash stored at key
func (r *Redis) HLen(ctx context.Context, key string) (value int64, err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		total, rerr := r.client.HLen(ctx, key).Result()
		value = total
		return rerr
	}, accept)
	return
}

// HExists returns if field is an existing field in the hash stored at key
func (r *Redis) HExists(ctx context.Context, key string, field string) (value bool, err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		reply, rerr := r.client.HExists(ctx, key, field).Result()
		value = reply
		return rerr
	}, accept)
	return
}

// HDel removes the specified fields from the hash stored at key
func (r *Redis) HDel(ctx context.Context, key string, fields ...string) (value int64, err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		reply, rerr := r.client.HDel(ctx, key, fields...).Result()
		value = reply
		return rerr
	}, accept)
	return
}

// HIncrBy increments the number stored at field in the hash stored at key by increment
func (r *Redis) HIncrBy(ctx context.Context, key string, field string, increment int64) (value int64, err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		reply, rerr := r.client.HIncrBy(ctx, key, field, increment).Result()
		if rerr != nil {
			return rerr
		}
		value = reply
		return nil
	}, accept)
	return
}

// HSet sets field in the hash stored at key to value
func (r *Redis) HSet(ctx context.Context, key string, field string, value string) (err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		_, rerr := r.client.HSet(ctx, key, field, value).Result()
		return rerr
	}, accept)
	return
}

// HMSet sets the specified fields to their respective values in the hash stored at key
func (r *Redis) HMSet(ctx context.Context, key string, kvs map[string]string) (err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		_, rerr := r.client.HMSet(ctx, key, kvs).Result()
		return rerr
	}, accept)
	return
}

// LIndex returns the element at index index in the list stored at key
func (r *Redis) LIndex(ctx context.Context, key string, index int64) (value string, err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		reply, rerr := r.client.LIndex(ctx, key, index).Result()
		value = reply
		return rerr
	}, accept)
	return
}

// LLen returns the length of the list stored at key
func (r *Redis) LLen(ctx context.Context, key string) (value int64, err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		reply, rerr := r.client.LLen(ctx, key).Result()
		value = reply
		return rerr
	}, accept)
	return
}

// LPop removes and returns the first elements of the list stored at key
func (r *Redis) LPop(ctx context.Context, key string) (value string, err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		reply, rerr := r.client.LPop(ctx, key).Result()
		value = reply
		return rerr
	}, accept)
	return
}

// LPush insert all the specified values at the head of the list stored at key
func (r *Redis) LPush(ctx context.Context, key string, values ...interface{}) (value int64, err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		reply, rerr := r.client.LPush(ctx, key, values...).Result()
		value = reply
		return rerr
	}, accept)
	return
}

// LRange returns the specified elements of the list stored at key
func (r *Redis) LRange(ctx context.Context, key string, start int64, stop int64) (value []string, err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		reply, rerr := r.client.LRange(ctx, key, start, stop).Result()
		value = reply
		return rerr
	}, accept)
	return
}

// LRem removes the first count occurrences of elements equal to element from the list stored at key
func (r *Redis) LRem(ctx context.Context, key string, count int64, val string) (value int64, err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		reply, rerr := r.client.LRem(ctx, key, count, val).Result()
		value = reply
		return rerr
	}, accept)
	return
}

// RPop removes and returns the last elements of the list stored at key
func (r *Redis) RPop(ctx context.Context, key string) (value string, err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		reply, rerr := r.client.RPop(ctx, key).Result()
		value = reply
		return rerr
	}, accept)
	return
}

// RPush insert all the specified values at the tail of the list stored at key
func (r *Redis) RPush(ctx context.Context, key string, values ...interface{}) (value int64, err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		reply, rerr := r.client.RPush(ctx, key, values...).Result()
		value = reply
		return rerr
	}, accept)
	return
}

// SCard returns the set cardinality (number of elements) of the set stored at key
func (r *Redis) SCard(ctx context.Context, key string) (value int64, err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		reply, rerr := r.client.SCard(ctx, key).Result()
		value = reply
		return rerr
	}, accept)
	return
}

// SAdd add the specified members to the set stored at key
func (r *Redis) SAdd(ctx context.Context, key string, members ...interface{}) (value int64, err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		reply, rerr := r.client.SAdd(ctx, key, members).Result()
		value = reply
		return rerr
	}, accept)
	return
}

// SDiff returns the members of the set resulting from the difference
// between the first set and all the successive sets
func (r *Redis) SDiff(ctx context.Context, keys ...string) (value []string, err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		reply, rerr := r.client.SDiff(ctx, keys...).Result()
		value = reply
		return rerr
	}, accept)
	return
}

// SDiffStore this command is equal to SDIFF
// but instead of returning the resulting set, it is stored in destination
func (r *Redis) SDiffStore(ctx context.Context, destination string, keys ...string) (value int64, err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		reply, rerr := r.client.SDiffStore(ctx, destination, keys...).Result()
		value = reply
		return rerr
	}, accept)
	return
}

// SIntersect returns the members of the set resulting from the intersection of all the given sets
func (r *Redis) SIntersect(ctx context.Context, keys ...string) (value []string, err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		reply, rerr := r.client.SInter(ctx, keys...).Result()
		value = reply
		return rerr
	}, accept)
	return
}

// SIntersectStore this command is equal to SIntersect
// but instead of returning the resulting set, it is stored in destination
func (r *Redis) SIntersectStore(ctx context.Context, destination string, keys ...string) (value int64, err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		reply, rerr := r.client.SInterStore(ctx, destination, keys...).Result()
		value = reply
		return rerr
	}, accept)
	return
}

// SUnion returns the members of the set resulting from the union of all the given sets
func (r *Redis) SUnion(ctx context.Context, keys ...string) (value []string, err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		reply, rerr := r.client.SUnion(ctx, keys...).Result()
		value = reply
		return rerr
	}, accept)
	return
}

// SUnionStore this command is equal to SUnion
// but instead of returning the resulting set, it is stored in destination
func (r *Redis) SUnionStore(ctx context.Context, destination string, keys ...string) (value int64, err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		reply, rerr := r.client.SUnionStore(ctx, destination, keys...).Result()
		value = reply
		return rerr
	}, accept)
	return
}

// SMembers returns all the members of the set value stored at key
func (r *Redis) SMembers(ctx context.Context, key string) (value []string, err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		reply, rerr := r.client.SMembers(ctx, key).Result()
		value = reply
		return rerr
	}, accept)
	return
}

// SIsMember returns if member is a member of the set stored at key
func (r *Redis) SIsMember(ctx context.Context, key string, member interface{}) (value bool, err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		reply, rerr := r.client.SIsMember(ctx, key, member).Result()
		value = reply
		return rerr
	}, accept)
	return
}

// SPop removes and returns one random members from the set value store at key
func (r *Redis) SPop(ctx context.Context, key string) (value string, err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		reply, rerr := r.client.SPop(ctx, key).Result()
		value = reply
		return rerr
	}, accept)
	return
}

// SPopN removes and returns count random members from the set value store at key
func (r *Redis) SPopN(ctx context.Context, key string, count int64) (value []string, err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		reply, rerr := r.client.SPopN(ctx, key, count).Result()
		value = reply
		return rerr
	}, accept)
	return
}

// SRandMember return a random element from the set value stored at key
func (r *Redis) SRandMember(ctx context.Context, key string) (value string, err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		reply, rerr := r.client.SRandMember(ctx, key).Result()
		value = reply
		return rerr
	}, accept)
	return
}

// SRandMemberN return a count element from the set value stored at key
func (r *Redis) SRandMemberN(ctx context.Context, key string, count int64) (value []string, err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		reply, rerr := r.client.SRandMemberN(ctx, key, count).Result()
		value = reply
		return rerr
	}, accept)
	return
}

// SRem remove the specified members from the set stored at key
func (r *Redis) SRem(ctx context.Context, key string, members ...interface{}) (value int64, err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		reply, rerr := r.client.SRem(ctx, key, members...).Result()
		value = reply
		return rerr
	}, accept)
	return
}

// ZCard returns the sorted set number of elements
func (r *Redis) ZCard(ctx context.Context, key string) (value int64, err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		reply, rerr := r.client.ZCard(ctx, key).Result()
		value = reply
		return rerr
	}, accept)
	return
}

// ZAdd adds all the specified members with the specified scores to the sorted set
func (r *Redis) ZAdd(ctx context.Context, key string, pais ...*Pair) (value int64, err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		zs := make([]*redis.Z, 0, len(pais))
		for _, pair := range pais {
			zs = append(zs, &redis.Z{
				Score:  float64(pair.Score),
				Member: pair.Member,
			})
		}
		reply, rerr := r.client.ZAdd(ctx, key, zs...).Result()
		value = reply
		return rerr
	}, accept)
	return
}

// ZIncrBy increments the score of member in the sorted set stored at key by increment
func (r *Redis) ZIncrBy(ctx context.Context, key string, member string, increment int64) (value int64, err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		reply, rerr := r.client.ZIncrBy(ctx, key, float64(increment), member).Result()
		value = int64(reply)
		return rerr
	}, accept)
	return
}

// ZCount returns the number of elements in the sorted set with a score between min and max
func (r *Redis) ZCount(ctx context.Context, key string, min int64, max int64) (value int64, err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		reply, rerr := r.client.ZCount(
			ctx,
			key,
			strconv.FormatInt(min, 10),
			strconv.FormatInt(max, 10),
		).Result()
		value = reply
		return rerr
	}, accept)
	return
}

// ZRange returns the specified range of elements in the sorted set
// with the scores ordered from high to low
func (r *Redis) ZRange(ctx context.Context, key string, start int64, stop int64) (value []string, err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		reply, rerr := r.client.ZRange(ctx, key, start, stop).Result()
		value = reply
		return rerr
	}, accept)
	return
}

// ZRangeByScore returns all the elements in the sorted set with a score between min and max
func (r *Redis) ZRangeByScore(ctx context.Context, key string, min int64, max int64) (value []string, err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		rb := &redis.ZRangeBy{
			Min: strconv.FormatInt(min, 10),
			Max: strconv.FormatInt(max, 10),
		}
		reply, rerr := r.client.ZRangeByScore(ctx, key, rb).Result()
		value = reply
		return rerr
	}, accept)
	return
}

// ZRangeWithScores this command is similar to ZRange but with score option
func (r *Redis) ZRangeWithScores(ctx context.Context, key string, start int64, stop int64) (value []*Pair, err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		reply, rerr := r.client.ZRangeWithScores(ctx, key, start, stop).Result()
		value = r.toPairs(reply)
		return rerr
	}, accept)
	return
}

// ZRank returns the rank of member in the sorted set stored, with the scores ordered from low to high
// The rank (or index) is 0-based, which means that the member with the lowest score has rank 0
func (r *Redis) ZRank(ctx context.Context, key string, member string) (value int64, err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		reply, rerr := r.client.ZRank(ctx, key, member).Result()
		value = reply
		return rerr
	}, accept)
	return
}

// ZRem removes the specified members from the sorted set stored. Non existing members are ignored
func (r *Redis) ZRem(ctx context.Context, key string, members ...interface{}) (value int64, err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		reply, rerr := r.client.ZRem(ctx, key, members...).Result()
		value = reply
		return rerr
	}, accept)
	return
}

// ZRemRangeByRank removes all elements in the sorted set stored with rank between start and stop
func (r *Redis) ZRemRangeByRank(ctx context.Context, key string, start int64, stop int64) (value int64, err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		reply, rerr := r.client.ZRemRangeByRank(ctx, key, start, stop).Result()
		value = reply
		return rerr
	}, accept)
	return
}

// ZRemRangeByScore removes all elements in the sorted set stored with a score between min and max (inclusive)
func (r *Redis) ZRemRangeByScore(ctx context.Context, key string, min int64, max int64) (value int64, err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		reply, rerr := r.client.ZRemRangeByScore(ctx,
			key,
			strconv.FormatInt(min, 10),
			strconv.FormatInt(max, 10),
		).Result()
		value = reply
		return rerr
	}, accept)
	return
}

// ZScore returns the score of member in the sorted
func (r *Redis) ZScore(ctx context.Context, key string, member string) (value int64, err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		reply, rerr := r.client.ZScore(ctx, key, member).Result()
		value = int64(reply)
		return rerr
	}, accept)
	return
}

// ZPopMax removes and returns up to count members with the highest scores in the sorted set
// the default value for count is 1, specifying a count greater than len(set) will not produce an error
func (r *Redis) ZPopMaxN(ctx context.Context, key string, count ...int64) (value []*Pair, err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		reply, rerr := r.client.ZPopMax(ctx, key, count...).Result()
		value = r.toPairs(reply)
		return rerr
	}, accept)
	return
}

// ZPopMin removes and returns up to count members with the lowest scores in the sorted set
// the default value for count is 1, specifying a count greater than len(set) will not produce an error
func (r *Redis) ZPopMin(ctx context.Context, key string, count ...int64) (value []*Pair, err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		reply, rerr := r.client.ZPopMin(ctx, key, count...).Result()
		value = r.toPairs(reply)
		return rerr
	}, accept)
	return
}

// ZRevRange returns the specified range of elements in the sorted set stored
// with the scores ordered from high to low
func (r *Redis) ZRevRange(ctx context.Context, key string, start int64, stop int64) (value []string, err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		reply, rerr := r.client.ZRevRange(ctx, key, start, stop).Result()
		value = reply
		return rerr
	}, accept)
	return
}

// ZRevRangeWithScores this command is similar to ZRevRange but with score option
func (r *Redis) ZRevRangeWithScores(ctx context.Context, key string, start int64, stop int64) (value []*Pair, err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		reply, rerr := r.client.ZRevRangeWithScores(ctx, key, start, stop).Result()
		value = r.toPairs(reply)
		return rerr
	}, accept)
	return
}

// ZRevRangeByScore returns all the elements in the sorted set at key with a score between max and min
// including elements with score equal to max or min
func (r *Redis) ZRevRangeByScore(ctx context.Context, key string, zrb *ZRangeBy) (value []string, err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		reply, rerr := r.client.ZRevRangeByScore(ctx, key, zrb).Result()
		value = reply
		return rerr
	}, accept)
	return
}

// ZRevRangeByScoreWithScores this command is similar to ZRevRangeByScore but with score option
func (r *Redis) ZRevRangeByScoreWithScores(ctx context.Context, key string, zrb *ZRangeBy) (value []*Pair, err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		reply, rerr := r.client.ZRevRangeByScoreWithScores(ctx, key, zrb).Result()
		value = r.toPairs(reply)
		return rerr
	}, accept)
	return
}

// ZRevRank returns the rank of member in the sorted set stored at key
// with the scores ordered from high to low
func (r *Redis) ZRevRank(ctx context.Context, key string, member string) (value int64, err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		reply, rerr := r.client.ZRevRank(ctx, key, member).Result()
		value = reply
		return rerr
	}, accept)
	return
}

// Pipelined batch execute commands
func (r *Redis) Pipelined(ctx context.Context, fn func(pipe Pipeliner) error) (value []redis.Cmder, err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		cmds, rerr := r.client.Pipelined(ctx, fn)
		value = cmds
		return rerr
	}, accept)
	return
}

// TxPipelined acts like Pipeline, but wraps queued commands with MULTI/EXEC
func (r *Redis) TxPipelined(ctx context.Context, fn func(pipe Pipeliner) error) (value []redis.Cmder, err error) {
	err = r.breakers.Do(r.config.dbAddr, func() error {
		cmds, rerr := r.client.TxPipelined(ctx, fn)
		value = cmds
		return rerr
	}, accept)
	return
}

// toPairs transfer redis.Z to Pair
func (r *Redis) toPairs(zs []redis.Z) (pairs []*Pair) {
	for _, z := range zs {
		pair := &Pair{}
		pair.Score = int64(z.Score)
		if member, ok := z.Member.(string); ok {
			pair.Member = member
		}
		pairs = append(pairs, pair)
	}
	return
}
