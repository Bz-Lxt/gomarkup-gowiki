package lock

import (
	"sync"
	"time"

	"gowiki/internal/pkg/timeutil"
)

type Record struct {
	ParagraphID string
	HolderID    string
	HolderName  string
	Until       time.Time
}

type Store struct {
	mu       sync.Mutex
	locks    map[string]map[string]Record // docID -> paragraphID -> record
	timeout  time.Duration
}

func New(timeout time.Duration) *Store {
	return &Store{
		locks:   map[string]map[string]Record{},
		timeout: timeout,
	}
}

func (s *Store) Acquire(docID, paragraphID, userID, name string) (Record, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.gcLocked(docID)
	m := s.doc(docID)
	now := timeutil.Now()
	if rec, ok := m[paragraphID]; ok && rec.Until.After(now) && rec.HolderID != userID {
		return rec, false
	}
	rec := Record{
		ParagraphID: paragraphID,
		HolderID:    userID,
		HolderName:  name,
		Until:       now.Add(s.timeout),
	}
	m[paragraphID] = rec
	return rec, true
}

func (s *Store) Heartbeat(docID, paragraphID, userID string) (Record, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	m := s.doc(docID)
	rec, ok := m[paragraphID]
	if !ok || rec.HolderID != userID {
		return rec, false
	}
	rec.Until = timeutil.Now().Add(s.timeout)
	m[paragraphID] = rec
	return rec, true
}

func (s *Store) Release(docID, paragraphID, userID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	m := s.doc(docID)
	if rec, ok := m[paragraphID]; ok && rec.HolderID == userID {
		delete(m, paragraphID)
	}
}

func (s *Store) ReleaseAll(docID, userID string) []Record {
	s.mu.Lock()
	defer s.mu.Unlock()
	m := s.doc(docID)
	var dropped []Record
	for id, rec := range m {
		if rec.HolderID == userID {
			dropped = append(dropped, rec)
			delete(m, id)
		}
	}
	return dropped
}

func (s *Store) List(docID string) []Record {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.gcLocked(docID)
	m := s.doc(docID)
	out := make([]Record, 0, len(m))
	for _, rec := range m {
		out = append(out, rec)
	}
	return out
}

func (s *Store) doc(docID string) map[string]Record {
	m, ok := s.locks[docID]
	if !ok {
		m = map[string]Record{}
		s.locks[docID] = m
	}
	return m
}

func (s *Store) gcLocked(docID string) {
	now := timeutil.Now()
	m := s.doc(docID)
	for id, rec := range m {
		if !rec.Until.After(now) {
			delete(m, id)
		}
	}
}
