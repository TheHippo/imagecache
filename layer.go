package imagecache

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/TheHippo/imagecache/list"
)

// Layer represents a caching layer
type Layer struct {
	cache     Cacher
	evictions []EvictionStrategy
	size      atomic.Int64
	count     atomic.Int32
	access    *list.List[*item]
	inventory map[string]*list.Element[*item]
	lock      sync.RWMutex
}

// item within the caching layer
type item struct {
	name       string
	lastAccess time.Time
	size       int64
}

// LayerStats contains information about the cache items
// in the layer
//   - Count of items
//   - Size of items in bytes
type LayerStats struct {
	Count int32
	Size  int64
}

// compile time check
var _ Cacher = &Layer{}

// NewLayer creates a new caching layer with various eviction strategies.
// If no evicition strategy is passed the items will never be deleted.
func NewLayer(cache Cacher, evictions ...EvictionStrategy) *Layer {
	return &Layer{
		cache:     cache,
		evictions: evictions,
		access:    list.NewList[*item](),
		inventory: make(map[string]*list.Element[*item], 0),
	}
}

// BackgroundEviction enabled eviction of stale items in the background.
// Items are checked for eviction according to dur. This function blocks
// and stop it provide a Context that can be canceled.
func (l *Layer) BackgroundEviction(ctx context.Context, dur time.Duration) {
	ticker := time.NewTicker(dur)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			l.Evict(ctx)
		}
	}
}

func (l *Layer) deleteLast(ctx context.Context) {
	l.lock.Lock()
	defer l.lock.Unlock()

	last := l.access.Back()
	if last == nil {
		// nothing in cache
		return
	}
	if err := l.cache.Delete(ctx, last.Value.name); err != nil {
		return
	}
	l.count.Add(-1)
	l.size.Add(-last.Value.size)
	delete(l.inventory, last.Value.name)
	l.access.Remove(last)
}

// Evict instructs all Evictionstrategies to remove items that need to be evicted.
// Returns the number of items that were evicted.
func (l *Layer) Evict(ctx context.Context) (count int) {
	for _, e := range l.evictions {
		for {
			if e.check(l) {
				count++
				l.deleteLast(ctx)
			} else {
				break
			}
		}
	}
	return
}

// Delete an items from the underlying cache. The exact error
// depends on the underlying cache implementation.
func (l *Layer) Delete(ctx context.Context, name string) error {
	if err := l.cache.Delete(ctx, name); err != nil {
		return err
	}

	go func(name string) {
		l.lock.Lock()
		if e, ok := l.inventory[name]; ok {
			l.count.Add(-1)
			l.size.Add(-e.Value.size)
			l.access.Remove(e)
			delete(l.inventory, name)
		}
		l.lock.Unlock()
	}(name)

	return nil
}

// Exists checks if an items exists in the underlying cache.
// Does not count as an access.
func (l *Layer) Exists(ctx context.Context, name string) bool {
	return l.cache.Exists(ctx, name)
}

func (l *Layer) accessed(name string, size int64) {
	l.lock.Lock()
	e, ok := l.inventory[name]
	if !ok {
		i := &item{
			name:       name,
			lastAccess: time.Now(),
			size:       size,
		}
		l.inventory[name] = l.access.PushFront(i)
		l.count.Add(1)
		l.size.Add(size)
	} else {
		e.Value.lastAccess = time.Now()
		e.Value.size = size
		l.access.MoveToFront(e)
	}
	l.lock.Unlock()
}

// Get an item from the layer. Returns the content
// or an error if something went wrong. The behavior of
// the error depends on the underlying cache. This also
// counts as an access to the item.
func (l *Layer) Get(ctx context.Context, name string) ([]byte, error) {
	content, err := l.cache.Get(ctx, name)
	if err != nil {
		return nil, err
	}
	// this is necessary because there might be items in the
	// cache that the cache isn't aware of. (filesystem after restart)
	go l.accessed(name, int64(len(content)))
	return content, nil
}

// Put an item into the layer. Required a key and the content itself.
// Returns an error if something went wrong. The exact error depends
// on the underlying cache. If the item already exists the item is
// overwritten. This also counts as an access to the item.
func (l *Layer) Put(ctx context.Context, name string, content []byte) error {
	if err := l.cache.Put(ctx, name, content); err != nil {
		return err
	}

	go func(name string, size int64) {
		// update size, if already in cache
		l.lock.RLock()
		if e, ok := l.inventory[name]; ok {
			old := e.Value.size
			diff := size - old
			l.size.Add(int64(diff))

		}
		l.lock.RUnlock()
		l.accessed(name, size)
		l.Evict(ctx)
	}(name, int64(len(content)))

	return nil
}

// Stats returns the current state of the layer.
func (l *Layer) Stats() *LayerStats {
	return &LayerStats{
		Count: l.count.Load(),
		Size:  l.size.Load(),
	}
}
