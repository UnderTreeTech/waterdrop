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

package dao

import (
	"fmt"

	"github.com/UnderTreeTech/waterdrop/pkg/database/redis"

	"github.com/UnderTreeTech/waterdrop/pkg/log"

	"github.com/UnderTreeTech/waterdrop/pkg/conf"
	"github.com/UnderTreeTech/waterdrop/pkg/database/sql"
)

var d *dao

type dao struct {
	db    *sql.DB
	redis *redis.Redis
}

func NewDao() *dao {
	db := NewMySQL()
	r := NewRedis()

	ds := &dao{
		db:    db,
		redis: r,
	}
	d = ds
	return ds
}

func (d *dao) Close() {
	d.db.Close()
	d.redis.Close()
}

func NewMySQL() *sql.DB {
	config := &sql.Config{}
	if err := conf.Unmarshal("mysql", config); err != nil {
		panic(fmt.Sprintf("unmarshal mysql config fail,err msg %s", err.Error()))
	}
	log.Infof("db config", log.Any("config", config))
	db := sql.NewMySQL(config)

	return db
}

func NewRedis() *redis.Redis {
	config := &redis.Config{}
	if err := conf.Unmarshal("redis", config); err != nil {
		panic(fmt.Sprintf("unmarshal redis config fail,err msg %s", err.Error()))
	}
	log.Infof("redis config", log.Any("config", config))

	redis := redis.New(config)
	return redis
}

func GetDao() *dao {
	return d
}

func GetRedis() *redis.Redis {
	return d.redis
}
