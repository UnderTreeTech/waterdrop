package xcollection

import (
	"sync"
	"time"
)

type RollingWindow struct {
	mutex            sync.RWMutex
	win              *window
	bucketSize       int
	bucketDuration   time.Duration
	lastWindowOffset int
	lastWindowTime   int
}

func NewRollingWindow(bucketSize int, bucketDuration time.Duration) *RollingWindow {
	rw := &RollingWindow{
		bucketSize:     bucketSize,
		bucketDuration: bucketDuration,
		win:            newWindow(bucketSize),
	}

	return rw
}

func (rw *RollingWindow) Add(v float64) {
	rw.mutex.Lock()
	rw.updateWindowOffset()
	rw.win.buckets[rw.lastWindowOffset].add(v)
	rw.mutex.Unlock()
}

func (rw *RollingWindow) Reduce(fn func(*window) float64) float64 {
	rw.mutex.RLock()
	rw.updateWindowOffset()
	reduce := fn(rw.win)
	rw.mutex.RUnlock()

	return reduce
}

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

	rw.lastWindowTime = adjustedTime
	rw.lastWindowOffset = windowOffset
}

func (rw *RollingWindow) resetWindow() {
	for _, bucket := range rw.win.buckets {
		bucket.reset()
	}
}

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

type window struct {
	buckets    []*bucket
	bucketSize int
}

func newWindow(size int) *window {
	buckets := make([]*bucket, 0, size)
	for i := 0; i < size; i++ {
		buckets = append(buckets, &bucket{})
	}

	return &window{
		buckets:    buckets,
		bucketSize: size,
	}
}

type bucket struct {
	Sum   float64
	Count int64
}

func (b *bucket) add(v float64) {
	b.Sum += v
	b.Count++
}

func (b *bucket) reset() {
	b.Sum = 0
	b.Count = 0
}
