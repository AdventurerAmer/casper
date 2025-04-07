package main

import "sync"

type Store struct {
	Mu sync.RWMutex
	M  map[string]any
}

func NewStore() *Store {
	return &Store{
		M: make(map[string]any),
	}
}

func (s *Store) Put(k string, v any) {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	s.M[k] = v
}

func (s *Store) Get(k string) (any, bool) {
	s.Mu.RLock()
	defer s.Mu.RUnlock()
	v, ok := s.M[k]
	return v, ok
}

func (s *Store) Delete(k string) {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	delete(s.M, k)
}
