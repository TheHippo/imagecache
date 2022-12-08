package imagecache

import (
	"errors"
	"sync"
)

// Memory can be used as a [Storer] or [Cacher].
type Memory struct {
	data map[string][]byte
	lock sync.RWMutex
}

// ErrNotInMemory is returned if an item does not exists in [Memory]
var ErrNotInMemory = errors.New("does not exist in memory cache")

// compile-time check
var _ Storer = &Memory{}
var _ Cacher = &Memory{}

// NewMemory creates a new in-memory [Storer] or [Cacher]
func NewMemory() *Memory {
	return &Memory{
		data: make(map[string][]byte),
	}
}

// Put an item into Memory. It can't return an error.
func (m *Memory) Put(name string, content []byte) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.data[name] = content
	return nil
}

// Get an item from Memory. If the item does not exists it return
// [ErrNotInMemory] as the error.
func (m *Memory) Get(name string) ([]byte, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	content, exists := m.data[name]
	if exists {
		return content, nil
	}
	return nil, ErrNotInMemory
}

// Checks if an item exists
func (m *Memory) Exists(name string) bool {
	m.lock.RLock()
	defer m.lock.RUnlock()
	_, exists := m.data[name]
	return exists
}

// Delete an item from Memory. It does not return an error ever,
// if the item does not exist nothing else happens.
func (m *Memory) Delete(name string) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	delete(m.data, name)
	return nil
}

// func (m *Memory) Stats() (count int32, size int64) {
// 	m.lock.RLock()
// 	defer m.lock.RUnlock()
// 	for _, v := range m.data {
// 		count++
// 		size += int64(len(v))
// 	}
// 	return
// }
