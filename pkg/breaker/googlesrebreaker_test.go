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

package breaker

import (
	"testing"

	"github.com/UnderTreeTech/waterdrop/pkg/log"

	"github.com/stretchr/testify/assert"
)

var bg = NewBreakerGroup()

// TestBreakerAccept test breaker Accept
func TestBreakerAccept(t *testing.T) {
	defer log.New(nil).Sync()
	breaker := bg.Get("breaker")
	for i := 0; i < 100; i++ {
		breaker.Accept()
	}
	assert.Nil(t, breaker.Allow())
}

// TestBreakerReject test breaker Reject
func TestBreakerReject(t *testing.T) {
	defer log.New(nil).Sync()
	breaker := bg.Get("breaker")
	for i := 0; i < 4000; i++ {
		breaker.Reject()
	}
	err := breaker.Allow()
	assert.NotNil(t, err)
}

// TestBreakerDo test breaker Do
func TestBreakerDo(t *testing.T) {
	defer log.New(nil).Sync()
	err := bg.Do("do", func() error {
		return nil
	}, func(e error) bool {
		return e == nil
	})
	assert.Nil(t, err)
	assert.Panics(t, func() {
		bg.Do("do", func() error {
			panic("exit")
		}, func(e error) bool {
			return e == nil
		})
	})
}
