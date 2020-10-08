package xbuffer

import "bytes"

type SizedBufferPool struct {
	pool  chan *bytes.Buffer
	size  int
	alloc int
}

// SizedBufferPool creates a new BufferPool bounded to the given size.
// size defines the number of buffers to be retained in the pool and alloc sets
// the initial capacity of new buffers to minimize calls to make().
//
// The value of alloc should seek to provide a buffer that is representative of
// most data written to the the buffer (i.e. 95th percentile) without being
// overly large (which will increase static memory consumption).
func NewSizedBufferPool(size int, alloc int) (bp *SizedBufferPool) {
	return &SizedBufferPool{
		size:  size,
		pool:  make(chan *bytes.Buffer, size),
		alloc: alloc,
	}
}

// Get gets a Buffer from the SizedBufferPool, or creates a new one if none are
// available in the pool. Buffers have a pre-allocated capacity.
func (bp *SizedBufferPool) Get() (b *bytes.Buffer) {
	select {
	case b = <-bp.pool:
	// reuse existing buffer
	default:
		// create new buffer
		b = bytes.NewBuffer(make([]byte, 0, bp.alloc))
	}
	return
}

// Put returns the given Buffer to the SizedBufferPool.
func (bp *SizedBufferPool) Put(b *bytes.Buffer) {
	if b.Cap() > bp.alloc {
		if len(bp.pool) < bp.size {
			b = bytes.NewBuffer(make([]byte, 0, bp.alloc))
		} else {
			return
		}
	} else {
		b.Reset()
	}

	select {
	case bp.pool <- b:
	default: // Discard the buffer if the pool is full.
	}
}
