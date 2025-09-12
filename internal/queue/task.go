package queue

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"time"
)

type TaskStatus string

const (
	StatusQueued  TaskStatus = "queued"
	StatusRunning TaskStatus = "running"
	StatusDone    TaskStatus = "done"
	StatusFailed  TaskStatus = "failed"
)

type Task struct {
	ID         string          `json:"id"`
	Payload    json.RawMessage `json:"payload"`
	MaxRetries int             `json:"maxRetries"`
	Attempt    int             `json:"attempt"`
	Status     TaskStatus      `json:"status"`
	CreatedAt  time.Time       `json:"createdAt"`
	UpdatedAt  time.Time       `json:"updatedAt"`
}

// NewTask constructs a new queued task with generated ID and timestamps.
func NewTask(payload json.RawMessage, maxRetries int) Task {
	if maxRetries < 0 {
		maxRetries = 0
	}
	id := generateID()
	now := time.Now().UTC()
	return Task{
		ID:         id,
		Payload:    payload,
		MaxRetries: maxRetries,
		Attempt:    0,
		Status:     StatusQueued,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

func generateID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
