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

package mongo

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/UnderTreeTech/waterdrop/pkg/log"

	"github.com/UnderTreeTech/waterdrop/pkg/utils/xstring"
)

var (
	db   *DB
	coll = "user"
	ctx  = context.Background()
)

type UserInfo struct {
	ID       int64  `bson:"_id"`
	Name     string `bson:"name"`
	Age      uint16 `bson:"age"`
	Weight   uint32 `bson:"weight"`
	Sex      uint8  `bson:"sex"`
	Address  string `bson:"address"`
	Email    string `bson:"email"`
	Mobile   string `bson:"mobile"`
	Account  string `bson:"account"`
	Password string `bson:"password"`
	Height   uint16 `bson:"height"`
}

func TestMain(m *testing.M) {
	defer log.New(nil).Sync()
	cfg := &Config{
		Addr:              "127.0.0.1:27017",
		DBName:            "test",
		DSN:               "mongodb://root:123456@127.0.0.1:27017/?connect=direct",
		MaxPoolSize:       20,
		MinPoolSize:       10,
		SlowQueryDuration: time.Millisecond * 500,
	}
	db = Open(cfg)
	defer db.Close()

	code := m.Run()
	os.Exit(code)
}

func TestCollection(t *testing.T) {
	id := int64(211023454441556)
	// single insert
	user := &UserInfo{
		ID:       id,
		Name:     xstring.RandomString(32),
		Age:      25,
		Weight:   70,
		Sex:      1,
		Address:  xstring.RandomString(32),
		Email:    xstring.RandomString(32),
		Mobile:   xstring.RandomString(32),
		Account:  xstring.RandomString(32),
		Password: xstring.RandomString(32),
		Height:   175,
	}

	result, err := db.GetCollection(coll).Insert(ctx, user)
	assert.Nil(t, err)
	assert.EqualValues(t, result.InsertedID, user.ID)

	// batch insert
	users := make([]*UserInfo, 0)
	for i := 1; i <= 10; i++ {
		doc := &UserInfo{
			ID:       id + int64(i),
			Name:     xstring.RandomString(32),
			Age:      25,
			Weight:   70,
			Sex:      1,
			Address:  xstring.RandomString(32),
			Email:    xstring.RandomString(32),
			Mobile:   xstring.RandomString(32),
			Account:  xstring.RandomString(32),
			Password: xstring.RandomString(32),
			Height:   175,
		}
		users = append(users, doc)
	}

	// batch insert
	reply, err := db.GetCollection(coll).BatchInsert(ctx, users)
	assert.Nil(t, err)
	assert.EqualValues(t, len(reply.InsertedIDs), 10)

	// find one
	var userinfo *UserInfo
	err = db.GetCollection(coll).Find(ctx, M{"name": user.Name}).One(&userinfo)
	assert.Nil(t, err)
	assert.EqualValues(t, user, userinfo)

	// find all
	allUsers := make([]*UserInfo, 0)
	err = db.GetCollection(coll).Find(ctx, M{}).Sort("name").Limit(5).All(&allUsers)
	assert.Nil(t, err)
	assert.EqualValues(t, len(allUsers), 5)

	// count
	count, err := db.GetCollection(coll).Find(ctx, M{}).Count()
	assert.Nil(t, err)
	assert.Greater(t, count, int64(0))

	// update
	uret, err := db.GetCollection(coll).UpdateAll(ctx,
		M{
			"_id": M{
				"$in": []int64{211023454441556, 211023454441557},
			},
		},
		M{
			"$set": M{
				"email": xstring.RandomString(32),
			},
		},
	)
	assert.Nil(t, err)
	assert.EqualValues(t, uret.MatchedCount, uret.ModifiedCount)

	err = db.GetCollection(coll).UpdateOne(ctx,
		M{
			"_id": 211023454441556,
		},
		M{
			"$set": M{
				"email": xstring.RandomString(32),
			},
		},
	)
	assert.Nil(t, err)

	err = db.GetCollection(coll).UpdateId(
		ctx,
		211023454441556,
		M{
			"$set": M{
				"email": xstring.RandomString(32),
			},
		},
	)
	assert.Nil(t, err)

	// bulk
	insert := &UserInfo{
		ID:       id + int64(100),
		Name:     xstring.RandomString(32),
		Age:      25,
		Weight:   70,
		Sex:      1,
		Address:  xstring.RandomString(32),
		Email:    xstring.RandomString(32),
		Mobile:   xstring.RandomString(32),
		Account:  xstring.RandomString(32),
		Password: xstring.RandomString(32),
		Height:   175,
	}
	bret, err := db.GetCollection(coll).Bulk(ctx).
		InsertOne(insert).
		UpdateOne(M{"_id": 211023454441557}, M{"$set": M{"email": xstring.RandomString(32)}}).
		Remove(M{"_id": 211023454441558}).
		Run()
	assert.Nil(t, err)
	assert.EqualValues(t, bret.ModifiedCount, int64(1))
	assert.EqualValues(t, bret.InsertedCount, int64(1))
	assert.EqualValues(t, bret.InsertedCount, int64(1))

	// delete
	err = db.GetCollection(coll).Remove(ctx, M{"_id": 211023454441556})
	assert.Nil(t, err)

	err = db.GetCollection(coll).RemoveId(ctx, 211023454441557)
	assert.Nil(t, err)

	deleteAll, err := db.GetCollection(coll).RemoveAll(ctx, M{})
	assert.Nil(t, err)
	assert.Greater(t, deleteAll.DeletedCount, int64(0))
}
