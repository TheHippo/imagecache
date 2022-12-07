package imagecache

import (
	"time"
)

// Prefixes that help to calculate cache sizes.
//
//	// 1GB
//	const MaxSize = 1 * imageCache.GB
const (
	KB = 1024
	MB = KB * 1024
	GB = MB * 1024
)

type EvictionStrategy interface {
	check(c *Layer) bool
}

// compile time checks
var _ EvictionStrategy = &LastAccessEviction{}
var _ EvictionStrategy = &MaxCacheSizeEviction{}
var _ EvictionStrategy = &MaxItemsEviction{}

// LastAccessEviction evicts items after a certain time not being accessed
type LastAccessEviction struct {
	dur time.Duration
}

// NewLastAccessEviction creates a new EvictionStrategy based on the time of
// the last access
func NewLastAccessEviction(duration time.Duration) *LastAccessEviction {
	return &LastAccessEviction{
		dur: duration,
	}
}

func (lae *LastAccessEviction) check(c *Layer) bool {
	c.lock.RLock()
	defer c.lock.RUnlock()
	last := c.access.Back()
	if last != nil {
		return time.Since(c.access.Back().Value.lastAccess) > lae.dur
	}
	// cache is empty
	return false
}

// MaxCacheSizeEviction evict items when a certain size is reached. Items
// are evicted in the order of their last access.
type MaxCacheSizeEviction struct {
	maxSize int64
}

// NewMaxCacheSizeEviction creates a new EvictionStrategy which evicts items
// by their last access when a certain size is reached.
func NewMaxCacheSizeEviction(size int64) *MaxCacheSizeEviction {
	return &MaxCacheSizeEviction{
		maxSize: size,
	}
}

func (mse *MaxCacheSizeEviction) check(c *Layer) bool {
	size := c.size.Load()
	return size > mse.maxSize
}

// MaxItemsEviction evict items when a certain number of items is reached.
// Items are evicted in the order of their last access.
type MaxItemsEviction struct {
	n int
}

// NewMaxItemsEviction create a new EvictionStrategy which evicts items
// when a certain number of items are in the layer. Items are evicted in the
// order by their last access.
func NewMaxItemsEviction(number int) *MaxItemsEviction {
	return &MaxItemsEviction{
		n: number,
	}
}

func (mie *MaxItemsEviction) check(c *Layer) bool {
	count := c.count.Load()
	return count > int32(mie.n)
}
