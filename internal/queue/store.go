package queue

import (
	"sync"
	"time"
)

// Store is an in-memory storage for tasks guarded by RWMutex.
type Store struct {
	mu    sync.RWMutex
	tasks map[string]Task
}

func NewStore() *Store {
	return &Store{tasks: make(map[string]Task)}
}

// Save creates or updates a task in storage and refreshes UpdatedAt.
func (s *Store) Save(t Task) Task {
	s.mu.Lock()
	defer s.mu.Unlock()
	t.UpdatedAt = time.Now().UTC()
	s.tasks[t.ID] = t
	return t
}

// Get returns a task by id.
func (s *Store) Get(id string) (Task, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tasks[id]
	return t, ok
}

// UpdateStatus sets status and attempt for a task if exists.
func (s *Store) UpdateStatus(id string, status TaskStatus, attempt int) (Task, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.tasks[id]
	if !ok {
		return Task{}, false
	}
	t.Status = status
	t.Attempt = attempt
	t.UpdatedAt = time.Now().UTC()
	s.tasks[id] = t
	return t, true
}
