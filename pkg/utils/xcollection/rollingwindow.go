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
	"sync"
	"time"
)

// RollingWindow is a window Accumulator implementation that uses some
// duration of time to determine the content of the window
type RollingWindow struct {
	mutex            sync.RWMutex
	win              *window
	bucketSize       int
	bucketDuration   time.Duration
	lastWindowOffset int
	lastWindowTime   int
}

// NewRollingWindow manages a window with rolling time durations
// The given duration will be used to bucket data within the window
func NewRollingWindow(bucketSize int, bucketDuration time.Duration) *RollingWindow {
	rw := &RollingWindow{
		bucketSize:     bucketSize,
		bucketDuration: bucketDuration,
		win:            newWindow(bucketSize),
	}

	return rw
}

// Add a value to the window using a time bucketing strategy
func (rw *RollingWindow) Add(v float64) {
	rw.mutex.Lock()
	rw.updateWindowOffset()
	rw.win.buckets[rw.lastWindowOffset].add(v)
	rw.mutex.Unlock()
}

// Reduce the window to a single value using a reduction function
func (rw *RollingWindow) Reduce(fn func(*Bucket)) {
	rw.mutex.RLock()
	adjustedTime := int(time.Now().UnixNano() / rw.bucketDuration.Nanoseconds())
	windowOffset := adjustedTime % rw.bucketSize

	if adjustedTime-rw.lastWindowTime < rw.bucketSize && windowOffset >= rw.lastWindowOffset {
		//当时间跨越到n个时钟周期之后时，当前统计无意义,只有当同处一个时钟内，且必须是顺序索引（逆序说明时钟跑到下一个时钟去了）
		// When one or more buckets are missed we need to zero them out.
		for _, bucket := range rw.win.buckets {
			fn(bucket)
		}
	}
	rw.mutex.RUnlock()
}

// updateWindowOffset update window offset to keep consistent
func (rw *RollingWindow) updateWindowOffset() {
	adjustedTime := int(time.Now().UnixNano() / rw.bucketDuration.Nanoseconds())
	windowOffset := adjustedTime % rw.bucketSize

	// If we've waiting longer than a full window for data then we need to clear
	// the internal state completely.
	if adjustedTime-rw.lastWindowTime > rw.bucketSize {
		rw.resetWindow()
	}

	// When one or more buckets are missed we need to zero them out.
	if adjustedTime != rw.lastWindowTime && adjustedTime-rw.lastWindowTime < rw.bucketSize {
		rw.resetBucket(windowOffset)
	}

	if adjustedTime != rw.lastWindowTime {
		rw.lastWindowTime = adjustedTime
		rw.lastWindowOffset = windowOffset
	}
}

// resetWindow reset all the window
func (rw *RollingWindow) resetWindow() {
	for _, bucket := range rw.win.buckets {
		bucket.reset()
	}
}

// resetBucket reset bucket
func (rw *RollingWindow) resetBucket(offset int) {
	distance := offset - rw.lastWindowOffset
	// If the distance between current and last is negative then we've wrapped
	// around the ring. Recalculate the distance.
	if distance < 0 {
		distance = (rw.bucketSize - rw.lastWindowOffset) + offset
	}

	for counter := 1; counter <= distance; counter++ {
		offset := (counter + rw.lastWindowOffset) % rw.bucketSize
		rw.win.buckets[offset].reset()
	}
}

// window represents a bucketed set of data
type window struct {
	buckets    []*Bucket
	bucketSize int
}

// newWindow creates a Window with the given number of buckets
func newWindow(size int) *window {
	buckets := make([]*Bucket, 0, size)
	for i := 0; i < size; i++ {
		buckets = append(buckets, &Bucket{})
	}

	return &window{
		buckets:    buckets,
		bucketSize: size,
	}
}

// Bucket is a rolling window bucket
type Bucket struct {
	Sum   float64
	Count int64
}

// add a value to the bucket
func (b *Bucket) add(v float64) {
	b.Sum += v
	b.Count++
}

// reset bucket to initial status
func (b *Bucket) reset() {
	b.Sum = 0
	b.Count = 0
}
