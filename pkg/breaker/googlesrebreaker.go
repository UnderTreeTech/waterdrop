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

package breaker

import (
	"math"
	"sync/atomic"
	"time"

	"github.com/UnderTreeTech/waterdrop/pkg/log"
	"github.com/UnderTreeTech/waterdrop/pkg/status"

	"github.com/UnderTreeTech/waterdrop/pkg/utils/xcollection"
)

type googleSreBreaker struct {
	// k google accepts multiplier K
	k float64
	// state service status
	state int32
	// rw rolling window to stat metrics
	rw *xcollection.RollingWindow
	// proba open breaker probability
	proba *Proba
	// name breaker name
	name string
}

type GoogleSreBreakerConfig struct {
	K          float64
	Window     time.Duration
	BucketSize int
	Name       string
}

// defaultGoogleSreBreakerConfig default google sre breaker config
func defaultGoogleSreBreakerConfig() *GoogleSreBreakerConfig {
	return &GoogleSreBreakerConfig{
		K:          1.5,
		Window:     10 * time.Second,
		BucketSize: 40,
	}
}

// newGoogleSreBreaker new a google sre breaker
func newGoogleSreBreaker(config *GoogleSreBreakerConfig) *googleSreBreaker {
	if config == nil {
		config = defaultGoogleSreBreakerConfig()
	}

	interval := time.Duration(int64(config.Window) / int64(config.BucketSize))
	rw := xcollection.NewRollingWindow(config.BucketSize, interval)

	breaker := &googleSreBreaker{
		k:     config.K,
		rw:    rw,
		proba: NewProba(),
		state: StateOpen,
		name:  config.Name,
	}

	return breaker
}

// Allow check the if the request can be successfully execute
func (gsb *googleSreBreaker) Allow() error {
	success, total := gsb.summary()
	googleAccepts := gsb.k * success
	dropRatio := math.Max(0, (float64(total)-googleAccepts)/float64(total+1))
	log.Debugf("breaker", log.String("name", gsb.name), log.Int64("total", total), log.Float64("success", success), log.Float64("accepts", googleAccepts), log.Float64("ratio", dropRatio))
	if dropRatio <= 0 {
		if atomic.LoadInt32(&gsb.state) == StateOpen {
			atomic.CompareAndSwapInt32(&gsb.state, StateOpen, StateClosed)
		}
		return nil
	}

	if atomic.LoadInt32(&gsb.state) == StateClosed {
		atomic.CompareAndSwapInt32(&gsb.state, StateClosed, StateOpen)
	}

	if gsb.proba.TrueOnProba(dropRatio) {
		return status.ServiceUnavailable
	}

	return nil
}

// summary summarize the buckets data
func (gsb *googleSreBreaker) summary() (success float64, total int64) {
	gsb.rw.Reduce(func(bucket *xcollection.Bucket) {
		success += bucket.Sum
		total += bucket.Count
	})

	return
}

// Accept indicate request execute successfully
func (gsb *googleSreBreaker) Accept() {
	gsb.rw.Add(1)
}

// Reject indicate request is denied
func (gsb *googleSreBreaker) Reject() {
	gsb.rw.Add(0)
}
