package xbuffer

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBufferPool(t *testing.T) {
	alloc := 1024
	pool := NewBufferPool(alloc)
	pool.Put(bytes.NewBuffer(make([]byte, 0, 2*alloc)))
	assert.True(t, pool.Get().Cap() <= alloc)
}
