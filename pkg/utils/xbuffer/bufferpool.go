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
	"sync"
)

type BufferPool struct {
	//alloc sets the initial capacity of new buffers to minimize calls to make()
	alloc int
	pool  *sync.Pool
}

// NewBufferPool creates a new BufferPool bounded to the given buffer size.
func NewBufferPool(alloc int) *BufferPool {
	return &BufferPool{
		alloc: alloc,
		pool: &sync.Pool{
			New: func() interface{} {
				return bytes.NewBuffer(make([]byte, 0, alloc))
			},
		},
	}
}

func (bp *BufferPool) Get() *bytes.Buffer {
	buffer := bp.pool.Get().(*bytes.Buffer)
	return buffer
}

func (bp *BufferPool) Put(buffer *bytes.Buffer) {
	if buffer.Cap() <= bp.alloc {
		buffer.Reset()
		bp.pool.Put(buffer)
	}
}
