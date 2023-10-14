package tsmap

import (
	"sync"
)

type TSMap[K comparable, V any] interface {
	Get(key K) (V, bool)
	Set(key K, value V)
	Delete(key K)
	Len() int
	Range(fn func(key K, value V), exception ...func(key K, value V) bool)
}

type tsMap[K comparable, V any] struct {
	sync.Mutex
	mp map[K]V
}

func (s *tsMap[K, V]) Get(key K) (V, bool) {
	s.Lock()
	defer s.Unlock()
	v, ok := s.mp[key]
	return v, ok
}

func (s *tsMap[K, V]) Set(key K, value V) {
	s.Lock()
	defer s.Unlock()
	s.mp[key] = value
}

func (s *tsMap[K, V]) Delete(key K) {
	s.Lock()
	defer s.Unlock()
	delete(s.mp, key)
}

func (s *tsMap[K, V]) Len() int {
	s.Lock()
	defer s.Unlock()
	return len(s.mp)
}

func (s *tsMap[K, V]) Range(fn func(key K, value V), exception ...func(key K, value V) bool) {
	s.Lock()
	defer s.Unlock()
	except := s.defaultException
	if len(exception) > 0 {
		except = exception[0]
	}
	for key, value := range s.mp {
		if !except(key, value) {
			fn(key, value)
		}
	}
}

func New[K comparable, V any]() TSMap[K, V] {
	return &tsMap[K, V]{
		mp: make(map[K]V),
	}
}

func (s *tsMap[K, V]) defaultException(key K, value V) bool {
	return false
}
