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

package es

import (
	"context"
	"io"
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/UnderTreeTech/waterdrop/pkg/log"

	"github.com/stretchr/testify/assert"
)

var (
	cli *Client
	ctx = context.Background()
	// Create a new index.
	mapping = `
	{
		"settings":{
		"number_of_shards":1,
		"number_of_replicas":0
		},
		"mappings":{
			"properties":{
					"user_id":{
						"type":"long"
					},
					"name":{
						"type":"text"
					}
			}
		}
	}
	`
)

type User struct {
	UserId int64  `json:"user_id,omitempty"`
	Name   string `json:"name,omitempty"`
}

func TestMain(m *testing.M) {
	defer log.New(nil).Sync()

	cfg := &Config{
		URLs: []string{"http://localhost:9200"},
	}
	cli = NewClient(cfg)

	os.Exit(m.Run())
}

func TestEs(t *testing.T) {
	alive, err := cli.Ping(ctx)
	assert.Nil(t, err)
	assert.True(t, alive)

	exist, err := cli.ExistIndex(ctx, "test")
	assert.Nil(t, err)
	if exist {
		err = cli.DeleteIndex(ctx, "test")
		assert.Nil(t, err)
	}

	err = cli.CreateIndex(ctx, "test", mapping)
	assert.Nil(t, err)

	u1 := &User{UserId: 1, Name: "john"}
	_id, err := cli.CreateDoc(ctx, "test", u1, strconv.Itoa(int(u1.UserId)))
	assert.Nil(t, err)
	u2 := &User{}
	hit, err := cli.GetDoc(ctx, "test", _id, u2)
	assert.Nil(t, err)
	assert.True(t, hit)

	users := make([]interface{}, 0, 20)
	uids := make([]string, 0, 20)
	now := time.Now().Nanosecond()
	for i := 0; i < 20; i++ {
		u := &User{}
		u.UserId = int64(now + i)
		u.Name = "waterdrop " + strconv.Itoa(i)
		users = append(users, u)
		uids = append(uids, strconv.Itoa(int(u.UserId)))
	}
	ids, err := cli.CreateDocs(ctx, "test", users, uids...)
	assert.Nil(t, err)
	assert.EqualValues(t, len(ids), 20)

	err = cli.DeleteDoc(ctx, "test", _id)
	assert.Nil(t, err)

	num, err := cli.DeleteDocs(ctx, "test", ids[:10])
	assert.Nil(t, err)
	assert.EqualValues(t, num, 10)

	err = cli.UpdateDoc(ctx, "test", ids[11], &User{Name: "hello world"})
	assert.Nil(t, err)
	u3 := &User{}
	hit, err = cli.GetDoc(ctx, "test", ids[11], u3)
	assert.Nil(t, err)
	assert.True(t, hit)
	assert.Equal(t, u3.Name, "hello world")
	num, err = cli.UpdateDocs(ctx, "test", []string{ids[10]}, []interface{}{&User{UserId: 1}})
	assert.Nil(t, err)
	assert.EqualValues(t, num, 1)

	ids = append(ids, "1111")
	ret, err := cli.GetDocs(ctx, "test", ids)
	assert.Nil(t, err)
	assert.NotEqualValues(t, len(ret), len(ids))

	err = cli.Refresh(ctx, "test")
	assert.Nil(t, err)

	num, err = cli.Count(ctx, "test", nil)
	assert.Nil(t, err)
	assert.EqualValues(t, num, 10)

	q := NewBoolQuery().Should(NewMatchQuery("name", "waterdrop"))
	result, err := cli.NewSearch("test").Query(q).Do(ctx)
	us := make([]*User, 0, result.TotalHits())
	if result.TotalHits() > 0 {
		hits := result.Each(reflect.TypeOf(u3))
		for _, hit := range hits {
			if u, ok := hit.(*User); ok {
				us = append(us, u)
			}
		}
	}
	assert.Nil(t, err)
	assert.EqualValues(t, len(us), 9)

	scroll := cli.NewScroll("test").Size(1).Query(q)
	defer scroll.Clear(ctx)
	page := 0
	for {
		_, err = scroll.Do(ctx)
		if err == io.EOF {
			// io.EOF as error means there are no more search results.
			break
		}
		assert.Nil(t, err)
		page++
	}
	assert.EqualValues(t, page, 9)
}
