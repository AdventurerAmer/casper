package main

import (
	"log"
	"sync"
)

type Store struct {
	Mu     sync.RWMutex
	M      map[string][]byte
	Logger *WALogger
}

func NewStore() (*Store, error) {
	logger, err := NewWALogger()
	if err != nil {
		return nil, err
	}
	m := make(map[string][]byte)
	for _, e := range logger.ReadEntires() {
		log.Printf("%s\n", e)
		switch e.Op {
		case LogOpPut:
			m[e.Key] = e.Value
		case LogOpDelete:
			delete(m, e.Key)
		}
	}
	go logger.WriteLoop()
	return &Store{
		M:      m,
		Logger: logger,
	}, nil
}

func (s *Store) Put(k string, v []byte) {
	s.Logger.Put(k, v)
	s.Mu.Lock()
	defer s.Mu.Unlock()
	s.M[k] = v
}

func (s *Store) Get(k string) ([]byte, bool) {
	s.Mu.RLock()
	defer s.Mu.RUnlock()
	v, ok := s.M[k]
	return v, ok
}

func (s *Store) Delete(k string) {
	s.Logger.Delete(k)
	s.Mu.Lock()
	defer s.Mu.Unlock()
	delete(s.M, k)
}
