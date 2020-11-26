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
	"errors"
	"sync/atomic"
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestDo(t *testing.T) {
	sc := &SingleCall{}
	v, err := sc.Do("key", func() (interface{}, error) {
		return "inflight", nil
	})

	assert.Equal(t, err, nil)
	assert.Equal(t, v.(string), "inflight")
}

func TestDoErr(t *testing.T) {
	sc := &SingleCall{}
	someErr := errors.New("some error")
	v, err := sc.Do("key", func() (interface{}, error) {
		return nil, someErr
	})
	assert.Equal(t, nil, v)
	assert.Equal(t, err.Error(), "some error")
}

func TestDoDupSuppress(t *testing.T) {
	sc := &SingleCall{}
	var calls int32
	fn := func() (interface{}, error) {
		atomic.AddInt32(&calls, 1)
		return "inflight", nil
	}

	wg := &WaitGroupWrapper{}
	for i := 0; i < 10; i++ {
		wg.Wrap(func() {
			v, err := sc.Do("key", fn)
			if err != nil {
				t.Errorf("Do error: %v", err)
			}
			if v.(string) != "inflight" {
				t.Errorf("got %q; want %q", v, "bar")
			}
		})
	}
	wg.Wait()
	assert.NotEqual(t, 10, calls)
}
