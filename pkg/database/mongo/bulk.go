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
	"github.com/UnderTreeTech/waterdrop/pkg/stats/metric"

	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"

	"github.com/opentracing/opentracing-go"
	"github.com/qiniu/qmgo"
)

// Bulk is context for batching operations to be sent to database in a single
// bulk write.
//
// Bulk is not safe for concurrent use.
//
// Notes:
//
// Individual operations inside a bulk do not trigger middlewares or hooks
// at present.
//
// Different from original mgo, the qmgo implementation of Bulk does not emulate
// bulk operations individually on old versions of MongoDB servers that do not
// natively support bulk operations.
//
// Only operations supported by the official driver are exposed, that is why
// InsertMany is missing from the methods.
type Bulk struct {
	bulk   *qmgo.Bulk
	config *Config
	span   opentracing.Span
	ctx    context.Context
	brk    *breaker.BreakerGroup
}

// SetOrdered marks the bulk as ordered or unordered.
//
// If ordered, writes does not continue after one individual write fails.
// Default is ordered.
func (b *Bulk) SetOrdered(ordered bool) *Bulk {
	b.bulk = b.bulk.SetOrdered(ordered)
	return b
}

// InsertOne queues an InsertOne operation for bulk execution.
func (b *Bulk) InsertOne(doc interface{}) *Bulk {
	b.bulk = b.bulk.InsertOne(doc)
	return b
}

// Remove queues a Remove operation for bulk execution.
func (b *Bulk) Remove(filter interface{}) *Bulk {
	b.bulk = b.bulk.Remove(filter)
	return b
}

// RemoveId queues a RemoveId operation for bulk execution.
func (b *Bulk) RemoveId(id interface{}) *Bulk {
	b.bulk = b.bulk.RemoveId(id)
	return b
}

// RemoveAll queues a RemoveAll operation for bulk execution.
func (b *Bulk) RemoveAll(filter interface{}) *Bulk {
	b.bulk = b.bulk.RemoveAll(filter)
	return b
}

// Upsert queues an Upsert operation for bulk execution.
// The replacement should be document without operator
func (b *Bulk) Upsert(filter interface{}, replacement interface{}) *Bulk {
	b.bulk = b.bulk.Upsert(filter, replacement)
	return b
}

// UpsertId queues an UpsertId operation for bulk execution.
// The replacement should be document without operator
func (b *Bulk) UpsertId(id interface{}, replacement interface{}) *Bulk {
	b.bulk = b.bulk.UpsertId(id, replacement)
	return b
}

// UpdateOne queues an UpdateOne operation for bulk execution.
// The update should contain operator
func (b *Bulk) UpdateOne(filter interface{}, update interface{}) *Bulk {
	b.bulk = b.bulk.UpdateOne(filter, update)
	return b
}

// UpdateId queues an UpdateId operation for bulk execution.
// The update should contain operator
func (b *Bulk) UpdateId(id interface{}, update interface{}) *Bulk {
	b.bulk = b.bulk.UpdateId(id, update)
	return b
}

// UpdateAll queues an UpdateAll operation for bulk execution.
// The update should contain operator
func (b *Bulk) UpdateAll(filter interface{}, update interface{}) *Bulk {
	b.bulk = b.bulk.UpdateAll(filter, update)
	return b
}

// Run executes the collected operations in a single bulk operation.
//
// A successful call resets the Bulk. If an error is returned, the internal
// queue of operations is unchanged, containing both successful and failed
// operations.
func (b *Bulk) Run() (result *qmgo.BulkResult, err error) {
	err = b.brk.Do(b.config.Addr, func() error {
		now := time.Now()
		defer b.span.Finish()

		result, err = b.bulk.Run(b.ctx)

		if ok, elapse := slowLog(now, b.config.SlowQueryDuration); ok {
			ext.Error.Set(b.span, true)
			b.span.LogFields(log.String("event", "slow_query"), log.Int64("elapse", int64(elapse)))
		}
		metric.MongoClientReqDuration.Observe(time.Since(now).Seconds(), b.config.DBName, b.config.Addr, "bulk")
		return err
	}, accept)

	return
}
