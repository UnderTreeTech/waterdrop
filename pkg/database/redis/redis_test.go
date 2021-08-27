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
	"os"
	"testing"

	"github.com/UnderTreeTech/waterdrop/pkg/log"

	"github.com/stretchr/testify/assert"
)

var (
	r   *Redis
	ctx = context.Background()
)

// TestMain test case entry point
func TestMain(m *testing.M) {
	defer log.New(nil).Sync()

	cfg := &Config{
		Addr:       []string{"127.0.0.1:6379"},
		DeployMode: "node",
	}

	rdb, err := New(cfg)
	if err != nil {
		os.Exit(1)
	}
	r = rdb
	defer r.client.Close()

	code := m.Run()
	os.Exit(code)
}

func TestGetSet(t *testing.T) {
	err := r.Set(ctx, "hello", "world")
	assert.Nil(t, err)
	val, err := r.Get(ctx, "hello")
	assert.Nil(t, err)
	assert.Equal(t, "world", val)
	err = r.MSet(ctx, map[string]string{
		"framework": "waterdrop",
		"language":  "golang",
		"db":        "redis",
	})
	assert.Nil(t, err)
	vals, err := r.MGet(ctx, []string{"hello", "db", "language", "framework", "github"}...)
	assert.Nil(t, err)
	assert.ElementsMatch(t, vals, []string{"waterdrop", "golang", "redis", "world", ""})
	_, err = r.Del(ctx, "framework", "language", "db", "hello")
	assert.Nil(t, err)
}

func TestHash(t *testing.T) {
	key := "hash"
	err := r.HSet(ctx, key, "hello", "world")
	assert.Nil(t, err)
	err = r.HMSet(ctx, key, map[string]string{
		"framework": "waterdrop",
		"language":  "golang",
		"db":        "redis",
	})
	assert.Nil(t, err)
	val, err := r.HGet(ctx, key, "language")
	assert.Nil(t, err)
	assert.Equal(t, val, "golang")
	mvals, err := r.HMGet(ctx, key, "hello", "db")
	assert.Nil(t, err)
	assert.ElementsMatch(t, []string{"world", "redis"}, mvals)
	vals, err := r.HGetAll(ctx, key)
	assert.Nil(t, err)
	assert.EqualValues(t, map[string]string{
		"hello":     "world",
		"framework": "waterdrop",
		"language":  "golang",
		"db":        "redis",
	}, vals)
	num, err := r.HDel(ctx, key, "hello")
	assert.Nil(t, err)
	assert.EqualValues(t, num, 1)
	size, err := r.HLen(ctx, key)
	assert.Nil(t, err)
	assert.EqualValues(t, size, 3)
	keys, err := r.HKeys(ctx, key)
	assert.Nil(t, err)
	assert.ElementsMatch(t, []string{"framework", "language", "db"}, keys)
	exist, err := r.HExists(ctx, key, "hello")
	assert.Nil(t, err)
	assert.False(t, exist)
	incr, err := r.HIncrBy(ctx, key, "incr", 6)
	assert.Nil(t, err)
	assert.EqualValues(t, 6, incr)
	_, err = r.Del(ctx, key)
	assert.Nil(t, err)
}

func TestList(t *testing.T) {
	key := "list"
	num, err := r.LPush(ctx, key, "hello", "waterdrop", "redis")
	assert.Nil(t, err)
	assert.EqualValues(t, num, 3)
	num, err = r.RPush(ctx, key, "golang", "db", "world")
	assert.Nil(t, err)
	assert.EqualValues(t, num, 6)
	size, err := r.LLen(ctx, key)
	assert.Nil(t, err)
	assert.EqualValues(t, size, 6)
	val, err := r.LIndex(ctx, key, 1)
	assert.Nil(t, err)
	assert.EqualValues(t, "waterdrop", val)
	val, err = r.LIndex(ctx, key, 7)
	assert.NotNil(t, err)
	assert.EqualValues(t, val, "")
	val, err = r.LPop(ctx, key)
	assert.Nil(t, err)
	assert.Equal(t, val, "redis")
	val, err = r.RPop(ctx, key)
	assert.Nil(t, err)
	assert.Equal(t, val, "world")
	num, err = r.LRem(ctx, key, 0, "hello")
	assert.Nil(t, err)
	assert.EqualValues(t, num, 1)
	vals, err := r.LRange(ctx, key, 0, -1)
	assert.Nil(t, err)
	assert.ElementsMatch(t, []string{"waterdrop", "golang", "db"}, vals)
	_, err = r.Del(ctx, key)
	assert.Nil(t, err)
}

func TestSet(t *testing.T) {
	s1 := "set1"
	s2 := "set2"
	num1, err := r.SAdd(ctx, s1, "hello", "world", "go", "waterdrop")
	assert.Nil(t, err)
	assert.EqualValues(t, num1, 4)
	num2, err := r.SAdd(ctx, s2, "hello", "framework", "github", "go")
	assert.Nil(t, err)
	assert.EqualValues(t, num2, 4)
	vals, err := r.SIntersect(ctx, s1, s2)
	assert.Nil(t, err)
	assert.ElementsMatch(t, vals, []string{"hello", "go"})
	num, err := r.SIntersectStore(ctx, "intersect", s1, s2)
	assert.Nil(t, err)
	assert.EqualValues(t, num, 2)
	vals, err = r.SDiff(ctx, s1, s2)
	assert.Nil(t, err)
	assert.ElementsMatch(t, vals, []string{"world", "waterdrop"})
	num, err = r.SDiffStore(ctx, "diff", s1, s2)
	assert.Nil(t, err)
	assert.EqualValues(t, num, 2)
	vals, err = r.SUnion(ctx, s1, s2)
	assert.Nil(t, err)
	assert.ElementsMatch(t, vals, []string{"hello", "world", "go", "waterdrop", "framework", "github"})
	num, err = r.SUnionStore(ctx, "union", s1, s2)
	assert.Nil(t, err)
	assert.EqualValues(t, num, 6)
	_, err = r.Del(ctx, "diff", "intersect", "union")
	assert.Nil(t, err)
	vals, err = r.SMembers(ctx, s1)
	assert.Nil(t, err)
	assert.ElementsMatch(t, vals, []string{"hello", "world", "go", "waterdrop"})
	exist, err := r.SIsMember(ctx, s1, "hello")
	assert.Nil(t, err)
	assert.True(t, exist)
	_, err = r.SPop(ctx, s1)
	assert.Nil(t, err)
	_, err = r.SPopN(ctx, s1, 2)
	assert.Nil(t, err)
	size, err := r.SCard(ctx, s1)
	assert.Nil(t, err)
	assert.EqualValues(t, 1, size)
	_, err = r.SRandMember(ctx, s2)
	assert.Nil(t, err)
	_, err = r.SRandMemberN(ctx, s2, 2)
	assert.Nil(t, err)
	num, err = r.SRem(ctx, s2, "github")
	assert.Nil(t, err)
	assert.EqualValues(t, num, 1)
	_, err = r.Del(ctx, s1, s2)
	assert.Nil(t, err)
}

func TestZSet(t *testing.T) {
	z1 := "zset1"
	p1 := []*Pair{
		{
			Member: "hello",
			Score:  1,
		},
		{
			Member: "world",
			Score:  2,
		},
		{
			Member: "waterdrop",
			Score:  3,
		},
		{
			Member: "go",
			Score:  4,
		},
		{
			Member: "redis",
			Score:  5,
		},
	}
	num, err := r.ZAdd(ctx, z1, p1...)
	assert.Nil(t, err)
	assert.EqualValues(t, num, 5)
	num, err = r.ZCount(ctx, z1, -1, 8)
	assert.Nil(t, err)
	assert.EqualValues(t, num, 5)
	num, err = r.ZIncrBy(ctx, z1, "redis", -1)
	assert.Nil(t, err)
	assert.EqualValues(t, num, 4)
	vals, err := r.ZPopMaxN(ctx, z1, 2)
	assert.Nil(t, err)
	assert.ElementsMatch(t, vals, []*Pair{
		{
			Member: "go",
			Score:  4,
		},
		{
			Member: "redis",
			Score:  4,
		},
	})
	vals, err = r.ZPopMin(ctx, z1, 1)
	assert.Nil(t, err)
	assert.ElementsMatch(t, vals, []*Pair{
		{
			Member: "hello",
			Score:  1,
		},
	})
	svals, err := r.ZRange(ctx, z1, 0, -1)
	assert.Nil(t, err)
	assert.ElementsMatch(t, svals, []string{"world", "waterdrop"})
	svals, err = r.ZRangeByScore(ctx, z1, 0, 5)
	assert.Nil(t, err)
	assert.ElementsMatch(t, svals, []string{"world", "waterdrop"})
	vals, err = r.ZRangeWithScores(ctx, z1, 0, 5)
	assert.Nil(t, err)
	assert.ElementsMatch(t, vals, []*Pair{
		{
			Member: "world",
			Score:  2,
		},
		{
			Member: "waterdrop",
			Score:  3,
		},
	})
	num, err = r.ZRank(ctx, z1, "world")
	assert.Nil(t, err)
	assert.Zero(t, num)
	svals, err = r.ZRevRange(ctx, z1, 0, -1)
	assert.Nil(t, err)
	assert.EqualValues(t, svals, []string{"waterdrop", "world"})
	svals, err = r.ZRevRangeByScore(ctx, z1, &ZRangeBy{Min: "0", Max: "5"})
	assert.Nil(t, err)
	assert.EqualValues(t, svals, []string{"waterdrop", "world"})
	vals, err = r.ZRevRangeWithScores(ctx, z1, 0, -1)
	assert.Nil(t, err)
	assert.EqualValues(t, vals, []*Pair{
		{
			Member: "waterdrop",
			Score:  3,
		},
		{
			Member: "world",
			Score:  2,
		},
	})
	vals, err = r.ZRevRangeByScoreWithScores(ctx, z1, &ZRangeBy{Min: "0", Max: "5"})
	assert.Nil(t, err)
	assert.EqualValues(t, vals, []*Pair{
		{
			Member: "waterdrop",
			Score:  3,
		},
		{
			Member: "world",
			Score:  2,
		},
	})
	num, err = r.ZRevRank(ctx, z1, "waterdrop")
	assert.Nil(t, err)
	assert.Zero(t, num)
	num, err = r.ZScore(ctx, z1, "waterdorp")
	assert.NotNil(t, err)
	assert.Zero(t, num)
	num, err = r.ZScore(ctx, z1, "world")
	assert.Nil(t, err)
	assert.EqualValues(t, num, 2)

	_, err = r.Del(ctx, z1)
	assert.Nil(t, err)
}

func TestPipeline(t *testing.T) {
	kvs := map[string]string{
		"language":  "golang",
		"db":        "redis",
		"framework": "waterdrop",
	}
	err := r.MSet(ctx, map[string]string{
		"framework": "waterdrop",
		"language":  "golang",
		"db":        "redis",
	})
	assert.Nil(t, err)
	cmds, err := r.Pipelined(ctx, func(pipe Pipeliner) error {
		pipe.Get(ctx, "language")
		pipe.Get(ctx, "db")
		pipe.Get(ctx, "framework")
		return nil
	})
	assert.Nil(t, err)
	for _, cmd := range cmds {
		if c, ok := cmd.(*StringCmd); ok {
			val, err := c.Result()
			assert.Nil(t, err)
			assert.Equal(t, val, kvs[c.Args()[1].(string)])
		}
	}

	cmds, err = r.TxPipelined(ctx, func(pipe Pipeliner) error {
		pipe.Get(ctx, "language")
		pipe.Get(ctx, "db")
		pipe.Get(ctx, "framework")
		return nil
	})
	assert.Nil(t, err)
	for _, cmd := range cmds {
		if c, ok := cmd.(*StringCmd); ok {
			val, err := c.Result()
			assert.Nil(t, err)
			assert.Equal(t, val, kvs[c.Args()[1].(string)])
		}
	}
	_, err = r.Del(ctx, "framework", "language", "db", "hello")
	assert.Nil(t, err)
}

func TestOtherCommands(t *testing.T) {
	err := r.Set(ctx, "hello", "world")
	assert.Nil(t, err)
	num, err := r.Del(ctx, "hello")
	assert.Nil(t, err)
	assert.Equal(t, num, int64(1))
	exists, err := r.Exists(ctx, "hello")
	assert.Nil(t, err)
	assert.False(t, exists)

	val, err := r.IncrBy(ctx, "incr", 1)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), val)
	_, err = r.Del(ctx, "incr")
	assert.Nil(t, err)

	err = r.SetEx(ctx, "setex", "", 10)
	assert.Nil(t, err)
	err = r.Expire(ctx, "setex", 10)
	assert.Nil(t, err)
	expire, err := r.TTL(ctx, "setex")
	assert.Nil(t, err)
	assert.NotZero(t, expire)
	_, err = r.Del(ctx, "setex")
	assert.Nil(t, err)

	locked, err := r.SetNxEx(ctx, "setnx", "", 10000)
	assert.Nil(t, err)
	assert.True(t, locked)
	locked, err = r.SetNxEx(ctx, "setnx", "", 10000)
	assert.Nil(t, err)
	assert.False(t, locked)
	_, err = r.Del(ctx, "setnx")
	assert.Nil(t, err)
}
