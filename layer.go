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
	access    *list.List[*Item]
	inventory map[string]*list.Element[*Item]
	lock      sync.RWMutex
}

// Item within the caching layer
// TODO: this does not need to be exported
type Item struct {
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
		access:    list.NewList[*Item](),
		inventory: make(map[string]*list.Element[*Item], 0),
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
			l.Evict()
		}
	}
}

func (l *Layer) deleteLast() {
	l.lock.Lock()
	defer l.lock.Unlock()

	last := l.access.Back()
	if last == nil {
		// nothing in cache
		return
	}
	if err := l.cache.Delete(last.Value.name); err != nil {
		return
	}
	l.count.Add(-1)
	l.size.Add(-last.Value.size)
	delete(l.inventory, last.Value.name)
	l.access.Remove(last)
}

// Evict instructs all Evictionstrategies to remove items that need to be evicted.
// Returns the number of items that were evicted.
func (l *Layer) Evict() (count int) {
	for _, e := range l.evictions {
		for {
			if e.check(l) {
				count++
				l.deleteLast()
			} else {
				break
			}
		}
	}
	return
}

// Delete an items from the underlying cache. The exact error
// depends on the underlying cache implementation.
func (l *Layer) Delete(name string) error {
	if err := l.cache.Delete(name); err != nil {
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
func (l *Layer) Exists(name string) bool {
	return l.cache.Exists(name)
}

func (l *Layer) accessed(name string, size int64) {
	l.lock.Lock()
	e, ok := l.inventory[name]
	if !ok {
		i := &Item{
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
func (l *Layer) Get(name string) ([]byte, error) {
	content, err := l.cache.Get(name)
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
func (l *Layer) Put(name string, content []byte) error {
	if err := l.cache.Put(name, content); err != nil {
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
		l.Evict()
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
