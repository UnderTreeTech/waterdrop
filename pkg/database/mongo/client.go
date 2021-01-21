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
	"fmt"
	"sync"
	"time"

	"github.com/qiniu/qmgo"
)

type Config struct {
	DSN    string
	DBName string
	Addr   string

	MaxPoolSize uint64
	MinPoolSize uint64

	SlowQueryDuration time.Duration
}

type DB struct {
	*qmgo.Database
	*Config
}

var collections = sync.Map{}

// Open return database instance handler
func Open(config *Config) (*DB, func() error) {
	cli, close := client(config)
	dbHandler := cli.Database(config.DBName)
	db := &DB{
		Database: dbHandler,
		Config:   config,
	}
	return db, close
}

// GetCollection return collection handler
func (d *DB) GetCollection(name string) *Collection {
	if coll, ok := collections.Load(name); ok {
		return coll.(*Collection)
	}

	collection := &Collection{
		conn:   d.Collection(name),
		config: d.Config,
		name:   name,
	}
	collections.Store(name, collection)

	return collection
}

// client return mongodb connection instance handler
func client(config *Config) (*qmgo.Client, func() error) {
	if config.MaxPoolSize < config.MinPoolSize {
		panic(fmt.Sprintf("MaxPoolSize must greater and equal than MinPoolSize"))
	}

	cli, err := qmgo.NewClient(context.Background(),
		&qmgo.Config{
			Uri:         config.DSN,
			MaxPoolSize: &config.MaxPoolSize,
			MinPoolSize: &config.MinPoolSize,
		})
	if err != nil {
		panic(fmt.Sprintf("open mongo client fail, err is %s", err.Error()))
	}

	close := func() error {
		return cli.Close(context.Background())
	}

	return cli, close
}

// slowLog check one db operation is slow or not
func slowLog(start time.Time, slowQueryDuration time.Duration) (slow bool, elapse time.Duration) {
	du := time.Since(start)
	if du > slowQueryDuration {
		return true, du
	}

	return false, 0
}
