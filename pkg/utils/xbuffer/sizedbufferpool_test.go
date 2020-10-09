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

package xbuffer

import (
	"bytes"
	"testing"
)

// TestSizedBufferPool checks that over-sized buffers are released and that new
// buffers are created in their place.
func TestSizedBufferPool(t *testing.T) {
	size := 4
	capacity := 1024

	bufPool := NewSizedBufferPool(size, capacity)

	b := bufPool.Get()

	// Check the cap before we use the buffer.
	if cap(b.Bytes()) != capacity {
		t.Fatalf("buffer capacity incorrect: got %v want %v", cap(b.Bytes()),
			capacity)
	}

	// Grow the buffer beyond our capacity and return it to the pool
	b.Grow(capacity * 3)
	bufPool.Put(b)

	// Add some additional buffers to fill up the pool.
	for i := 0; i < size; i++ {
		bufPool.Put(bytes.NewBuffer(make([]byte, 0, bufPool.alloc*2)))
	}

	// Check that oversized buffers are being replaced.
	if len(bufPool.pool) < size {
		t.Fatalf("buffer pool too small: got %v want %v", len(bufPool.pool), size)
	}

	// Close the channel so we can iterate over it.
	close(bufPool.pool)

	// Check that there are buffers of the correct capacity in the pool.
	for buffer := range bufPool.pool {
		if cap(buffer.Bytes()) != bufPool.alloc {
			t.Fatalf("returned buffers wrong capacity: got %v want %v",
				cap(buffer.Bytes()), capacity)
		}
	}
}
