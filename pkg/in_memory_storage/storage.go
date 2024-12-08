package inmemorystorage

import (
	"maps"
	"sync"
)

type Storage[K comparable, V any] struct {
	m   map[K]V
	mux sync.RWMutex
}

// New creates a new instance of Storage.
func New[K comparable, V any]() *Storage[K, V] {
	return &Storage[K, V]{
		m:   make(map[K]V),
		mux: sync.RWMutex{},
	}
}

// Set sets the value for the given key.
func (s *Storage[K, V]) Set(key K, value V) {
	s.mux.Lock()
	s.m[key] = value
	s.mux.Unlock()
}

// SetBatch sets multiple key-value pairs at once.
func (s *Storage[K, V]) SetBatch(m map[K]V) {
	s.mux.Lock()
	for key, value := range m {
		s.m[key] = value
	}
	s.mux.Unlock()
}

// Get returns the value for the given key.
//
//nolint:ireturn,nolintlint
func (s *Storage[K, V]) Get(key K) (V, bool) {
	s.mux.RLock()
	value, ok := s.m[key]
	s.mux.RUnlock()

	return value, ok
}

// Delete deletes the value for the given key.
func (s *Storage[K, V]) Delete(key K) {
	s.mux.Lock()
	delete(s.m, key)
	s.mux.Unlock()
}

// Keys returns all keys in the storage.
func (s *Storage[K, V]) Keys() []K {
	s.mux.RLock()
	defer s.mux.RUnlock()

	keys := make([]K, 0, len(s.m))
	for key := range s.m {
		keys = append(keys, key)
	}

	return keys
}

// Values returns all values in the storage.
func (s *Storage[K, V]) Values() []V {
	s.mux.RLock()
	defer s.mux.RUnlock()

	values := make([]V, 0, len(s.m))
	for _, value := range s.m {
		values = append(values, value)
	}

	return values
}

// Len returns the number of items in the storage.
func (s *Storage[K, V]) Len() int {
	s.mux.RLock()
	length := len(s.m)
	s.mux.RUnlock()

	return length
}

// Clear removes all items from the storage.
func (s *Storage[K, V]) Clear() {
	s.mux.Lock()
	s.m = make(map[K]V)
	s.mux.Unlock()
}

// Replace replaces the storage with the given map.
func (s *Storage[K, V]) Replace(m map[K]V) {
	s.mux.Lock()
	s.m = m
	s.mux.Unlock()
}

// Contains checks if the storage contains the given key.
func (s *Storage[K, V]) Contains(key K) bool {
	s.mux.RLock()
	_, ok := s.m[key]
	s.mux.RUnlock()

	return ok
}

// Copy returns a copy of the storage.
func (s *Storage[K, V]) Copy() *Storage[K, V] {
	return &Storage[K, V]{m: s.CopyMap(), mux: sync.RWMutex{}}
}

// CopyMap returns a copy of the storage as a map.
func (s *Storage[K, V]) CopyMap() map[K]V {
	s.mux.RLock()
	defer s.mux.RUnlock()

	m := make(map[K]V, len(s.m))
	maps.Copy(m, s.m)

	return m
}
