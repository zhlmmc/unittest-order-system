package concurrent

import (
	"sync/atomic"
)

// Counter represents an atomic counter
type Counter struct {
	value int64
}

// NewCounter creates a new Counter
func NewCounter(initial int64) *Counter {
	return &Counter{value: initial}
}

// Increment atomically increments the counter by 1
func (c *Counter) Increment() int64 {
	return atomic.AddInt64(&c.value, 1)
}

// Decrement atomically decrements the counter by 1
func (c *Counter) Decrement() int64 {
	return atomic.AddInt64(&c.value, -1)
}

// Add atomically adds delta to the counter
func (c *Counter) Add(delta int64) int64 {
	return atomic.AddInt64(&c.value, delta)
}

// Value returns the current value of the counter
func (c *Counter) Value() int64 {
	return atomic.LoadInt64(&c.value)
}

// Reset atomically sets the counter to zero
func (c *Counter) Reset() {
	atomic.StoreInt64(&c.value, 0)
}

// CompareAndSwap atomically swaps the counter if the current value equals old
func (c *Counter) CompareAndSwap(old, new int64) bool {
	return atomic.CompareAndSwapInt64(&c.value, old, new)
}
