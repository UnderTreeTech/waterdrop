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
