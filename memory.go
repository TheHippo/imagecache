package imagecache

import (
	"errors"
	"sync"
)

type Memory struct {
	data map[string][]byte
	lock sync.RWMutex
}

var notInMemory = errors.New("does not exist in memory cache")

// compile-time check
var _ Storer = &Memory{}
var _ Cacher = &Memory{}

func NewMemory() *Memory {
	return &Memory{
		data: make(map[string][]byte),
	}
}

func (m *Memory) Put(name string, content []byte) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.data[name] = content
	return nil
}

func (m *Memory) Get(name string) ([]byte, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	content, exists := m.data[name]
	if exists {
		return content, nil
	}
	return nil, notInMemory
}

func (m *Memory) Exists(name string) bool {
	m.lock.RLock()
	defer m.lock.RUnlock()
	_, exists := m.data[name]
	return exists
}

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
