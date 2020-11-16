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

package xcollection

import (
	"testing"
	"time"

	"github.com/go-playground/assert/v2"
)

func TestRollingWindow(t *testing.T) {
	bucketSize := 10
	bucketDuration := time.Millisecond * 100
	window := NewRollingWindow(bucketSize, bucketDuration)

	for i := 0; i < bucketSize; i++ {
		window.Add(1)
		time.Sleep(bucketDuration)
	}

	var total int64
	var success float64
	window.Reduce(func(bucket *Bucket) {
		success += bucket.Sum
		total += bucket.Count
	})

	assert.Equal(t, int64(success), total)
}
