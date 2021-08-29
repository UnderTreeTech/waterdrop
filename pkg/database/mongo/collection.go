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
	"time"

	"github.com/UnderTreeTech/waterdrop/pkg/breaker"

	"github.com/opentracing/opentracing-go/log"

	"github.com/qiniu/qmgo"

	"github.com/UnderTreeTech/waterdrop/pkg/trace"
	"github.com/opentracing/opentracing-go/ext"
)

type Collection struct {
	conn   *qmgo.Collection
	config *Config
	name   string
	brk    *breaker.BreakerGroup
}

// Aggregate executes an aggregate command against the collection and returns a Aggregate to get resulting documents.
func (c *Collection) Aggregate(ctx context.Context, pipeline interface{}) *Aggregate {
	span, _ := trace.StartSpanFromContext(ctx, "aggregate")
	ext.PeerAddress.Set(span, c.config.Addr)
	ext.DBType.Set(span, "mongo")
	ext.SpanKind.Set(span, ext.SpanKindRPCClientEnum)
	ext.DBInstance.Set(span, c.config.DBName)
	span = span.SetTag("db.collection", c.name)

	return &Aggregate{
		ai:     c.conn.Aggregate(ctx, pipeline),
		ctx:    ctx,
		config: c.config,
		span:   span,
		brk:    c.brk,
	}
}

// Bulk returns a new context for preparing bulk execution of operations.
func (c *Collection) Bulk(ctx context.Context) *Bulk {
	bulk := c.conn.Bulk()
	span, ctx := trace.StartSpanFromContext(ctx, "bulk")
	ext.PeerAddress.Set(span, c.config.Addr)
	ext.DBType.Set(span, "mongo")
	ext.SpanKind.Set(span, ext.SpanKindRPCClientEnum)
	ext.DBInstance.Set(span, c.config.DBName)
	span = span.SetTag("db.collection", c.name)

	return &Bulk{
		bulk:   bulk,
		config: c.config,
		span:   span,
		ctx:    ctx,
		brk:    c.brk,
	}
}

// Find find by condition filter，return Query
func (c *Collection) Find(ctx context.Context, filter interface{}) *Query {
	span, ctx := trace.StartSpanFromContext(ctx, "query")
	ext.PeerAddress.Set(span, c.config.Addr)
	ext.DBType.Set(span, "mongo")
	ext.SpanKind.Set(span, ext.SpanKindRPCClientEnum)
	ext.DBInstance.Set(span, c.config.DBName)
	span = span.SetTag("db.collection", c.name)

	query := c.conn.Find(ctx, filter)

	return &Query{
		qi:     query,
		config: c.config,
		span:   span,
		brk:    c.brk,
	}
}

// InsertOne insert one document into the collection
// Reference: https://docs.mongodb.com/manual/reference/command/insert/
func (c *Collection) Insert(ctx context.Context, doc interface{}) (result *qmgo.InsertOneResult, err error) {
	err = c.brk.Do(c.config.Addr, func() error {
		now := time.Now()
		span, ctx := trace.StartSpanFromContext(ctx, "insert")
		ext.PeerAddress.Set(span, c.config.Addr)
		ext.DBType.Set(span, "mongo")
		ext.SpanKind.Set(span, ext.SpanKindRPCClientEnum)
		ext.DBInstance.Set(span, c.config.DBName)
		span = span.SetTag("db.collection", c.name)
		defer span.Finish()

		result, err = c.conn.InsertOne(ctx, doc)
		if ok, elapse := slowLog(now, c.config.SlowQueryDuration); ok {
			ext.Error.Set(span, true)
			span.LogFields(log.String("event", "slow_query"), log.Int64("elapse", int64(elapse)))
		}
		return err
	}, accept)
	return
}

// BatchInsert executes an insert command to insert multiple documents into the collection.
// Reference: https://docs.mongodb.com/manual/reference/command/insert/
func (c *Collection) BatchInsert(ctx context.Context, docs interface{}) (result *qmgo.InsertManyResult, err error) {
	err = c.brk.Do(c.config.Addr, func() error {
		now := time.Now()
		span, ctx := trace.StartSpanFromContext(ctx, "batch_insert")
		ext.PeerAddress.Set(span, c.config.Addr)
		ext.DBType.Set(span, "mongo")
		ext.SpanKind.Set(span, ext.SpanKindRPCClientEnum)
		ext.DBInstance.Set(span, c.config.DBName)
		span = span.SetTag("db.collection", c.name)
		defer span.Finish()

		result, err = c.conn.InsertMany(ctx, docs)
		if ok, elapse := slowLog(now, c.config.SlowQueryDuration); ok {
			ext.Error.Set(span, true)
			span.LogFields(log.String("event", "slow_query"), log.Int64("elapse", int64(elapse)))
		}
		return err
	}, accept)
	return
}

// Upsert updates one documents if filter match, inserts one document if filter is not match, Error when the filter is invalid
// The replacement parameter must be a document that will be used to replace the selected document. It cannot be nil
// and cannot contain any update operators
// Reference: https://docs.mongodb.com/manual/reference/operator/update/
// If replacement has "_id" field and the document is exist, please initial it with existing id(even with Qmgo default field feature).
// Otherwise "the (immutable) field '_id' altered" error happens.
func (c *Collection) Upsert(ctx context.Context, filter interface{}, replacement interface{}) (result *qmgo.UpdateResult, err error) {
	err = c.brk.Do(c.config.Addr, func() error {
		now := time.Now()
		span, ctx := trace.StartSpanFromContext(ctx, "upsert")
		ext.PeerAddress.Set(span, c.config.Addr)
		ext.DBType.Set(span, "mongo")
		ext.SpanKind.Set(span, ext.SpanKindRPCClientEnum)
		ext.DBInstance.Set(span, c.config.DBName)
		span = span.SetTag("db.collection", c.name)
		defer span.Finish()

		result, err = c.conn.Upsert(ctx, filter, replacement)
		if ok, elapse := slowLog(now, c.config.SlowQueryDuration); ok {
			ext.Error.Set(span, true)
			span.LogFields(log.String("event", "slow_query"), log.Int64("elapse", int64(elapse)))
		}
		return err
	}, accept)
	return
}

// UpsertId updates one documents if id match, inserts one document if id is not match and the id will inject into the document
// The replacement parameter must be a document that will be used to replace the selected document. It cannot be nil
// and cannot contain any update operators
// Reference: https://docs.mongodb.com/manual/reference/operator/update/
func (c *Collection) UpsertId(ctx context.Context, id interface{}, replacement interface{}) (result *qmgo.UpdateResult, err error) {
	err = c.brk.Do(c.config.Addr, func() error {
		now := time.Now()
		span, ctx := trace.StartSpanFromContext(ctx, "upsert_id")
		ext.PeerAddress.Set(span, c.config.Addr)
		ext.DBType.Set(span, "mongo")
		ext.SpanKind.Set(span, ext.SpanKindRPCClientEnum)
		ext.DBInstance.Set(span, c.config.DBName)
		span = span.SetTag("db.collection", c.name)
		defer span.Finish()

		result, err = c.conn.UpsertId(ctx, id, replacement)
		if ok, elapse := slowLog(now, c.config.SlowQueryDuration); ok {
			ext.Error.Set(span, true)
			span.LogFields(log.String("event", "slow_query"), log.Int64("elapse", int64(elapse)))
		}
		return err
	}, accept)
	return
}

// UpdateOne executes an update command to update at most one document in the collection.
// Reference: https://docs.mongodb.com/manual/reference/operator/update/
func (c *Collection) UpdateOne(ctx context.Context, filter interface{}, update interface{}) (err error) {
	err = c.brk.Do(c.config.Addr, func() error {
		now := time.Now()
		span, ctx := trace.StartSpanFromContext(ctx, "update_one")
		ext.PeerAddress.Set(span, c.config.Addr)
		ext.DBType.Set(span, "mongo")
		ext.SpanKind.Set(span, ext.SpanKindRPCClientEnum)
		ext.DBInstance.Set(span, c.config.DBName)

		span = span.SetTag("db.collection", c.name)
		defer span.Finish()

		err = c.conn.UpdateOne(ctx, filter, update)
		if ok, elapse := slowLog(now, c.config.SlowQueryDuration); ok {
			ext.Error.Set(span, true)
			span.LogFields(log.String("event", "slow_query"), log.Int64("elapse", int64(elapse)))
		}
		return err
	}, accept)
	return
}

// UpdateId executes an update command to update at most one document in the collection.
// Reference: https://docs.mongodb.com/manual/reference/operator/update/
func (c *Collection) UpdateId(ctx context.Context, id interface{}, update interface{}) (err error) {
	err = c.brk.Do(c.config.Addr, func() error {
		now := time.Now()
		span, ctx := trace.StartSpanFromContext(ctx, "update_id")
		ext.PeerAddress.Set(span, c.config.Addr)
		ext.DBType.Set(span, "mongo")
		ext.SpanKind.Set(span, ext.SpanKindRPCClientEnum)
		ext.DBInstance.Set(span, c.config.DBName)
		span = span.SetTag("db.collection", c.name)
		defer span.Finish()

		err = c.conn.UpdateId(ctx, id, update)
		if ok, elapse := slowLog(now, c.config.SlowQueryDuration); ok {
			ext.Error.Set(span, true)
			span.LogFields(log.String("event", "slow_query"), log.Int64("elapse", int64(elapse)))
		}
		return err
	}, accept)
	return
}

// UpdateAll executes an update command to update documents in the collection.
// The matchedCount is 0 in UpdateResult if no document updated
// Reference: https://docs.mongodb.com/manual/reference/operator/update/
func (c *Collection) UpdateAll(ctx context.Context, filter interface{}, update interface{}) (result *qmgo.UpdateResult, err error) {
	err = c.brk.Do(c.config.Addr, func() error {
		now := time.Now()
		span, ctx := trace.StartSpanFromContext(ctx, "update_all")
		ext.PeerAddress.Set(span, c.config.Addr)
		ext.DBType.Set(span, "mongo")
		ext.SpanKind.Set(span, ext.SpanKindRPCClientEnum)
		ext.DBInstance.Set(span, c.config.DBName)
		span = span.SetTag("db.collection", c.name)
		defer span.Finish()

		result, err = c.conn.UpdateAll(ctx, filter, update)
		if ok, elapse := slowLog(now, c.config.SlowQueryDuration); ok {
			ext.Error.Set(span, true)
			span.LogFields(log.String("event", "slow_query"), log.Int64("elapse", int64(elapse)))
		}
		return err
	}, accept)
	return
}

// ReplaceOne executes an update command to update at most one document in the collection.
// If UpdateHook in opts is set, hook works on it, otherwise hook try the doc as hook
// Expect type of the doc is the define of user's document
func (c *Collection) ReplaceOne(ctx context.Context, filter interface{}, doc interface{}) (err error) {
	err = c.brk.Do(c.config.Addr, func() error {
		now := time.Now()
		span, ctx := trace.StartSpanFromContext(ctx, "replace_one")
		ext.PeerAddress.Set(span, c.config.Addr)
		ext.DBType.Set(span, "mongo")
		ext.SpanKind.Set(span, ext.SpanKindRPCClientEnum)
		ext.DBInstance.Set(span, c.config.DBName)
		span = span.SetTag("db.collection", c.name)
		defer span.Finish()

		err = c.conn.ReplaceOne(ctx, filter, doc)
		if ok, elapse := slowLog(now, c.config.SlowQueryDuration); ok {
			ext.Error.Set(span, true)
			span.LogFields(log.String("event", "slow_query"), log.Int64("elapse", int64(elapse)))
		}
		return err
	}, accept)
	return
}

// Remove executes a delete command to delete at most one document from the collection.
// if filter is bson.M{}，DeleteOne will delete one document in collection
// Reference: https://docs.mongodb.com/manual/reference/command/delete/
func (c *Collection) Remove(ctx context.Context, filter interface{}) (err error) {
	err = c.brk.Do(c.config.Addr, func() error {
		now := time.Now()
		span, ctx := trace.StartSpanFromContext(ctx, "remove")
		ext.PeerAddress.Set(span, c.config.Addr)
		ext.DBType.Set(span, "mongo")
		ext.SpanKind.Set(span, ext.SpanKindRPCClientEnum)
		ext.DBInstance.Set(span, c.config.DBName)
		span = span.SetTag("db.collection", c.name)
		defer span.Finish()

		err = c.conn.Remove(ctx, filter)
		if ok, elapse := slowLog(now, c.config.SlowQueryDuration); ok {
			ext.Error.Set(span, true)
			span.LogFields(log.String("event", "slow_query"), log.Int64("elapse", int64(elapse)))
		}
		return err
	}, accept)
	return
}

// RemoveId executes a delete command to delete at most one document from the collection.
func (c *Collection) RemoveId(ctx context.Context, id interface{}) (err error) {
	err = c.brk.Do(c.config.Addr, func() error {
		now := time.Now()
		span, ctx := trace.StartSpanFromContext(ctx, "remove_id")
		ext.PeerAddress.Set(span, c.config.Addr)
		ext.DBType.Set(span, "mongo")
		ext.SpanKind.Set(span, ext.SpanKindRPCClientEnum)
		ext.DBInstance.Set(span, c.config.DBName)
		span = span.SetTag("db.collection", c.name)
		defer span.Finish()

		err = c.conn.RemoveId(ctx, id)
		if ok, elapse := slowLog(now, c.config.SlowQueryDuration); ok {
			ext.Error.Set(span, true)
			span.LogFields(log.String("event", "slow_query"), log.Int64("elapse", int64(elapse)))
		}
		return err
	}, accept)
	return
}

// RemoveAll executes a delete command to delete documents from the collection.
// If filter is bson.M{}，all ducuments in Collection will be deleted
// Reference: https://docs.mongodb.com/manual/reference/command/delete/
func (c *Collection) RemoveAll(ctx context.Context, filter interface{}) (result *qmgo.DeleteResult, err error) {
	err = c.brk.Do(c.config.Addr, func() error {
		now := time.Now()
		span, ctx := trace.StartSpanFromContext(ctx, "remove_all")
		ext.PeerAddress.Set(span, c.config.Addr)
		ext.DBType.Set(span, "mongo")
		ext.SpanKind.Set(span, ext.SpanKindRPCClientEnum)
		ext.DBInstance.Set(span, c.config.DBName)
		span = span.SetTag("db.collection", c.name)
		defer span.Finish()

		result, err = c.conn.RemoveAll(ctx, filter)
		if ok, elapse := slowLog(now, c.config.SlowQueryDuration); ok {
			ext.Error.Set(span, true)
			span.LogFields(log.String("event", "slow_query"), log.Int64("elapse", int64(elapse)))
		}
		return err
	}, accept)
	return
}
