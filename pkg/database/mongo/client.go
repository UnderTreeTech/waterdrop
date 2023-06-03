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
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/UnderTreeTech/waterdrop/pkg/breaker"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/qiniu/qmgo"
)

// Config MongoDB DSN configs
type Config struct {
	DSN    string
	DBName string
	Addr   string

	MaxPoolSize uint64
	MinPoolSize uint64

	SlowQueryDuration time.Duration
}

// DB encapsulation of qmgo client and database
type DB struct {
	client      *qmgo.Client
	db          *qmgo.Database
	config      *Config
	close       func() error
	brk         *breaker.BreakerGroup
	collections sync.Map
}

type (
	// M is an alias of bson.M
	M = bson.M
	// A is an alias of bson.A
	A = bson.A
	// D is an alias of bson.D
	D = bson.D
	// E is an alias of bson.E
	E = bson.E
	// ObjectID is an alias of primitive.ObjectID
	ObjectID = primitive.ObjectID
)

var (
	// ErrNoSuchDocuments return if no document found
	ErrNoSuchDocuments = qmgo.ErrNoSuchDocuments

	// NilObjectID is the zero value for ObjectID
	NilObjectID = primitive.NilObjectID

	// ErrInvalidHex indicates that a hex string cannot be converted to an ObjectID
	ErrInvalidHex = primitive.ErrInvalidHex
)

// Open return database instance handler
func Open(config *Config) *DB {
	cli, close := client(config)
	dbHandler := cli.Database(config.DBName)
	db := &DB{
		client:      cli,
		db:          dbHandler,
		config:      config,
		close:       close,
		brk:         breaker.NewBreakerGroup(),
		collections: sync.Map{},
	}
	return db
}

// GetCollection return collection handler
func (d *DB) GetCollection(name string) *Collection {
	if coll, ok := d.collections.Load(name); ok {
		return coll.(*Collection)
	}

	collection := &Collection{
		conn:   d.db.Collection(name),
		config: d.config,
		name:   name,
		brk:    d.brk,
	}
	d.collections.Store(name, collection)
	return collection
}

// Close close the db connection
func (d *DB) Close() error {
	return d.close()
}

// Ping ping mongo to keepalive
func (d *DB) Ping() error {
	return d.client.Ping(2)
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

// IsErrNoDocuments check if err is no documents,
// simply call if err == ErrNoSuchDocuments or if err == mongo.ErrNoDocuments
func IsErrNoDocuments(err error) bool {
	if err == ErrNoSuchDocuments {
		return true
	}
	return false
}

// IsDup check if err is mongo E11000 (duplicate err)
func IsDup(err error) bool {
	return err != nil && strings.Contains(err.Error(), "E11000")
}

// accept check mongo op success or not
func accept(err error) bool {
	return err == nil || IsErrNoDocuments(err) || IsDup(err)
}

// NewObjectID generates a new ObjectID
func NewObjectID() ObjectID {
	return primitive.NewObjectID()
}

// ObjectIDFromHex creates a new ObjectID from a hex string
// It returns an error if the hex string is not a valid ObjectID
func ObjectIDFromHex(s string) (ObjectID, error) {
	if len(s) != 24 {
		return NilObjectID, ErrInvalidHex
	}

	b, err := hex.DecodeString(s)
	if err != nil {
		return NilObjectID, err
	}

	var oid [12]byte
	copy(oid[:], b[:])

	return oid, nil
}
