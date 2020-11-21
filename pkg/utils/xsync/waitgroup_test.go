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

package xsync

import (
	"sync/atomic"
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestWrap(t *testing.T) {
	wg := &WaitGroupWrapper{}

	var result int32
	for i := 0; i <= 100; i++ {
		j := i
		wg.Wrap(func() {
			atomic.AddInt32(&result, int32(j))
		})
	}
	wg.Wait()
	assert.Equal(t, int32(5050), result)
}
