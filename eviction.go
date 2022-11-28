package imagecache

import (
	"time"
)

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

type LastAccessEviction struct {
	dur time.Duration
}

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

type MaxCacheSizeEviction struct {
	maxSize int64
}

func NewMaxCacheSizeEviction(size int64) *MaxCacheSizeEviction {
	return &MaxCacheSizeEviction{
		maxSize: size,
	}
}

func (mse *MaxCacheSizeEviction) check(c *Layer) bool {
	size := c.size.Load()
	return size > mse.maxSize
}

type MaxItemsEviction struct {
	n int
}

func NewMaxItemsEviction(number int) *MaxItemsEviction {
	return &MaxItemsEviction{
		n: number,
	}
}

func (mie *MaxItemsEviction) check(c *Layer) bool {
	count := c.count.Load()
	return count > int32(mie.n)
}
