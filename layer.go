package imagecache

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type Layer struct {
	cache     Cacher
	evictions []EvictionStrategy
	size      atomic.Int64
	count     atomic.Int32
	access    *List[*Item]
	inventory map[string]*Element[*Item]
	lock      sync.RWMutex
}

type Item struct {
	name       string
	lastAccess time.Time
	size       int64
}

type LayerStats struct {
	Count int32
	Size  int64
}

var _ Cacher = &Layer{}

func NewLayer(cache Cacher, evictions ...EvictionStrategy) *Layer {
	return &Layer{
		cache:     cache,
		evictions: evictions,
		access:    NewList[*Item](),
		inventory: make(map[string]*Element[*Item], 0),
	}
}

func (l *Layer) BackgroundEviction(ctx context.Context, dur time.Duration) {
	ticker := time.NewTicker(dur)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			l.evict()
		}
	}
}

func (l *Layer) deleteLast() {
	l.lock.Lock()
	defer l.lock.Unlock()

	last := l.access.Back()
	if err := l.cache.Delete(last.Value.name); err != nil {
		return
	}
	l.count.Add(-1)
	l.size.Add(-last.Value.size)
	delete(l.inventory, last.Value.name)
	l.access.Remove(last)
}

func (l *Layer) evict() {
	for _, e := range l.evictions {
		for {
			if e.check(l) {
				l.deleteLast()
			} else {
				break
			}
		}
	}
}

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

func (l *Layer) Get(name string) ([]byte, error) {
	content, err := l.cache.Get(name)
	if err != nil {
		return nil, err
	}
	go l.accessed(name, int64(len(content)))
	return content, nil
}

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
		l.evict()
	}(name, int64(len(content)))

	return nil
}

func (l *Layer) Stats() *LayerStats {
	return &LayerStats{
		Count: l.count.Load(),
		Size:  l.size.Load(),
	}
}
