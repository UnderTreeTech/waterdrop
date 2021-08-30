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

	"github.com/opentracing/opentracing-go"

	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"github.com/qiniu/qmgo"
)

type Aggregate struct {
	ai     qmgo.AggregateI
	ctx    context.Context
	config *Config
	span   opentracing.Span
	brk    *breaker.BreakerGroup
}

// All iterates the cursor from aggregate and decodes each document into results.
func (a *Aggregate) All(results interface{}) (err error) {
	err = a.brk.Do(a.config.Addr, func() error {
		now := time.Now()
		a.span = a.span.SetOperationName("aggregate_all")
		defer a.span.Finish()

		err = a.ai.All(results)
		if ok, elapse := slowLog(now, a.config.SlowQueryDuration); ok {
			ext.Error.Set(a.span, true)
			a.span.LogFields(log.String("event", "slow_query"), log.Int64("elapse", int64(elapse)))
		}
		return err
	}, accept)
	return
}

// One iterates the cursor from aggregate and decodes current document into result.
func (a *Aggregate) One(result interface{}) (err error) {
	err = a.brk.Do(a.config.Addr, func() error {
		now := time.Now()
		a.span = a.span.SetOperationName("aggregate_one")
		defer a.span.Finish()

		err = a.ai.One(result)
		if ok, elapse := slowLog(now, a.config.SlowQueryDuration); ok {
			ext.Error.Set(a.span, true)
			a.span.LogFields(log.String("event", "slow_query"), log.Int64("elapse", int64(elapse)))
		}
		return err
	}, accept)
	return
}

// Iter return the cursor after aggregate
// In most scenario do not use Iter
func (a *Aggregate) Iter() qmgo.CursorI {
	return a.ai.Iter()
}
