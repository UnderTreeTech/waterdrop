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
	"time"

	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"

	"github.com/opentracing/opentracing-go"
	"github.com/qiniu/qmgo"
)

// Query struct definition
type Query struct {
	qi     qmgo.QueryI
	config *Config
	span   opentracing.Span
}

// Sort is Used to set the sorting rules for the returned results
// Format: "age" or "+age" means to sort the age field in ascending order, "-age" means in descending order
// When multiple sort fields are passed in at the same time, they are arranged in the order in which the fields are passed in.
// For example, {"age", "-name"}, first sort by age in ascending order, then sort by name in descending order
func (q *Query) Sort(fields ...string) *Query {
	q.qi = q.qi.Sort(fields...)
	return q
}

// Select is used to determine which fields are displayed or not displayed in the returned results
// Format: bson.M{"age": 1} means that only the age field is displayed
// bson.M{"age": 0} means to display other fields except age
// When _id is not displayed and is set to 0, it will be returned to display
func (q *Query) Select(projection interface{}) *Query {
	q.qi = q.qi.Select(projection)
	return q
}

// Skip skip n records
func (q *Query) Skip(n int64) *Query {
	q.qi = q.qi.Skip(n)
	return q
}

// Hint sets the value for the Hint field.
// This should either be the index name as a string or the index specification
// as a document. The default value is nil, which means that no hint will be sent.
func (q *Query) Hint(hint interface{}) *Query {
	q.qi = q.qi.Hint(hint)
	return q
}

// Limit limits the maximum number of documents found to n
// The default value is 0, and 0  means no limit, and all matching results are returned
// When the limit value is less than 0, the negative limit is similar to the positive limit, but the cursor is closed after returning a single batch result.
// Reference https://docs.mongodb.com/manual/reference/method/cursor.limit/index.html
func (q *Query) Limit(n int64) *Query {
	q.qi = q.qi.Limit(n)
	return q
}

// One query a record that meets the filter conditions
// If the search fails, an error will be returned
func (q *Query) One(result interface{}) (err error) {
	now := time.Now()
	q.span = q.span.SetOperationName("query_one")
	defer q.span.Finish()

	err = q.qi.One(result)
	if ok, elapse := slowLog(now, q.config.SlowQueryDuration); ok {
		ext.Error.Set(q.span, true)
		q.span.LogFields(log.String("event", "slow_query"), log.Int64("elapse", int64(elapse)))
	}
	return
}

// All query multiple records that meet the filter conditions
// The static type of result must be a slice pointer
func (q *Query) All(result interface{}) (err error) {
	now := time.Now()
	q.span = q.span.SetOperationName("query_all")
	defer q.span.Finish()

	err = q.qi.All(result)
	if ok, elapse := slowLog(now, q.config.SlowQueryDuration); ok {
		ext.Error.Set(q.span, true)
		q.span.LogFields(log.String("event", "slow_query"), log.Int64("elapse", int64(elapse)))
	}
	return
}

// Count count the number of eligible entries
func (q *Query) Count() (n int64, err error) {
	return q.qi.Count()
}

// Distinct gets the unique value of the specified field in the collection and return it in the form of slice
// result should be passed a pointer to slice
// The function will verify whether the static type of the elements in the result slice is consistent with the data type obtained in mongodb
// reference https://docs.mongodb.com/manual/reference/command/distinct/
func (q *Query) Distinct(key string, result interface{}) error {
	return q.qi.Distinct(key, result)
}

// Cursor gets a Cursor object, which can be used to traverse the query result set
// After obtaining the CursorI object, you should actively call the Close interface to close the cursor
// Strongly suggest use One or All
func (q *Query) Cursor() qmgo.CursorI {
	return q.qi.Cursor()
}

// Apply runs the findAndModify command, which allows updating, replacing
// or removing a document matching a query and atomically returning either the old
// version (the default) or the new version of the document (when ReturnNew is true)
//
// The Sort and Select query methods affect the result of Apply. In case
// multiple documents match the query, Sort enables selecting which document to
// act upon by ordering it first. Select enables retrieving only a selection
// of fields of the new or old document.
//
// When Change.Replace is true, it means replace at most one document in the collection
// and the update parameter must be a document and cannot contain any update operators;
// if no objects are found and Change.Upsert is false, it will returns ErrNoDocuments.
// When Change.Remove is true, it means delete at most one document in the collection
// and returns the document as it appeared before deletion; if no objects are found,
// it will returns ErrNoDocuments.
// When both Change.Replace and Change.Remove are false，it means update at most one document
// in the collection and the update parameter must be a document containing update operators;
// if no objects are found and Change.Upsert is false, it will returns ErrNoDocuments.
//
// reference: https://docs.mongodb.com/manual/reference/command/findAndModify/
func (q *Query) Apply(change qmgo.Change, result interface{}) error {
	return q.qi.Apply(change, result)
}
