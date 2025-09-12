package queue

import (
	"sync"
	"time"
)

// Store is an in-memory storage for tasks guarded by RWMutex.
type Store struct {
	mu      sync.RWMutex
	tasks   map[string]Task
	metrics Metrics
}

func NewStore() *Store {
	return &Store{tasks: make(map[string]Task)}
}

// Save creates or updates a task in storage and refreshes UpdatedAt.
func (s *Store) Save(t Task) Task {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.tasks[t.ID]; !exists {
		// new task entering as queued
		s.incrementMetric(StatusQueued, 1)
	}
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
	if t.Status != status {
		s.incrementMetric(t.Status, -1)
		s.incrementMetric(status, 1)
	}
	t.Status = status
	t.Attempt = attempt
	t.UpdatedAt = time.Now().UTC()
	s.tasks[id] = t
	return t, true
}

// Metrics holds counters per status.
type Metrics struct {
	Queued  uint64
	Running uint64
	Done    uint64
	Failed  uint64
}

// GetMetrics returns a copy of current metrics snapshot.
func (s *Store) GetMetrics() Metrics {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.metrics
}

func (s *Store) incrementMetric(status TaskStatus, delta int) {
	switch status {
	case StatusQueued:
		s.metrics.Queued = uint64(int64(s.metrics.Queued) + int64(delta))
	case StatusRunning:
		s.metrics.Running = uint64(int64(s.metrics.Running) + int64(delta))
	case StatusDone:
		s.metrics.Done = uint64(int64(s.metrics.Done) + int64(delta))
	case StatusFailed:
		s.metrics.Failed = uint64(int64(s.metrics.Failed) + int64(delta))
	}
}
