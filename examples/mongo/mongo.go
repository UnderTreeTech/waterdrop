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

package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/UnderTreeTech/waterdrop/pkg/log"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/UnderTreeTech/waterdrop/pkg/trace/jaeger"

	"github.com/UnderTreeTech/waterdrop/pkg/database/mongo"

	"github.com/UnderTreeTech/waterdrop/pkg/utils/xstring"

	"github.com/UnderTreeTech/waterdrop/pkg/server/http/config"

	"github.com/UnderTreeTech/waterdrop/pkg/conf"

	"github.com/UnderTreeTech/waterdrop/pkg/server/http/server"
	"github.com/gin-gonic/gin"
)

var db *mongo.DB
var close func() error

func main() {
	flag.Parse()

	conf.Init()
	defer log.New(nil).Sync()
	defer jaeger.Init()()

	cfg := &mongo.Config{}
	if err := conf.Unmarshal("mongo", cfg); err != nil {
		panic(fmt.Sprintf("unmarshal mongo config fail, err msg %s", err.Error()))
	}

	db = mongo.Open(cfg)
	defer db.Close()

	srvConfig := &config.ServerConfig{}
	if err := conf.Unmarshal("server.http", srvConfig); err != nil {
		panic(fmt.Sprintf("unmarshal http server config fail, err msg %s", err.Error()))
	}
	srv := server.New(srvConfig)

	g := srv.Group("/api")
	{
		g.GET("/ping", ping)
		g.GET("/mongo", testMongo)
	}

	srv.Start()

	time.Sleep(time.Minute * 5)
	srv.Stop(context.Background())
}

func ping(c *gin.Context) {
	c.JSON(200, "ping")
}

type UserInfo struct {
	ID       int64  `bson:"id"`
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

func testMongo(c *gin.Context) {
	ctx := c.Request.Context()

	// single insert
	user := &UserInfo{
		ID:       time.Now().UnixNano(),
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

	result, err := db.GetCollection("user").Insert(ctx, user)
	if err != nil {
		fmt.Println("insert doc fail", err)
	} else {
		fmt.Println(result.InsertedID)
	}

	// batch insert
	users := make([]*UserInfo, 0)
	for i := 0; i < 10; i++ {
		doc := &UserInfo{
			ID:       time.Now().UnixNano(),
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
	result2, err := db.GetCollection("user").BatchInsert(ctx, users)
	if err != nil {
		fmt.Println("batch insert doc fail", err)
	} else {
		fmt.Println(result2.InsertedIDs)
	}

	// find one
	var userinfo *UserInfo
	err = db.GetCollection("user").Find(ctx, bson.M{"name": user.Name}).One(&userinfo)
	if err != nil {
		fmt.Println("query doc fail", err)
	} else {
		fmt.Println(userinfo)
	}

	// find all
	allUsers := make([]*UserInfo, 0)
	err = db.GetCollection("user").Find(ctx, bson.M{}).Sort("name").Limit(5).All(&allUsers)
	if err != nil {
		fmt.Println("query limit docs fail", err)
	} else {
		for _, v := range allUsers {
			fmt.Println(v.ID)
		}
	}

	// count
	count, err := db.GetCollection("user").Find(ctx, bson.M{}).Count()
	if err != nil {
		fmt.Println("count fail", err)
	} else {
		fmt.Println("count number", count)
	}

	// update
	uret, err := db.GetCollection("user").UpdateAll(ctx, bson.M{"id": bson.M{"$in": []int64{1611138298337130000, 1611138298495058000}}}, bson.M{"$set": bson.M{"email": xstring.RandomString(32)}})
	if err != nil {
		fmt.Println("update doc fail", err)
	} else {
		fmt.Println(uret.MatchedCount, uret.ModifiedCount)
	}

	// delete
	err = db.GetCollection("user").Remove(ctx, bson.M{"id": 1611138298337130000})
	if err != nil {
		fmt.Println("delete doc fail", err)
	}

	// bulk
	insert := &UserInfo{
		ID:       time.Now().UnixNano(),
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

	bret, err := db.GetCollection("user").Bulk(ctx).
		InsertOne(insert).
		UpdateOne(bson.M{"id": 1611138298495058000}, bson.M{"$set": bson.M{"email": xstring.RandomString(32)}}).
		Remove(bson.M{"id": 1611138298495401000}).
		Run()

	if err != nil {
		fmt.Println("bulk doc fail", err)
	} else {
		fmt.Println(bret.ModifiedCount, bret.InsertedCount, bret.DeletedCount)
	}

	c.JSON(200, err.Error())
}
