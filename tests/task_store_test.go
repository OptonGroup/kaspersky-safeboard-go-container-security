package tests

import (
	"encoding/json"
	"testing"
	"time"

	q "github.com/optongroup/kaspersky-safeboard-go-container-security/internal/queue"
)

func TestNewTaskDefaults(t *testing.T) {
	payload := json.RawMessage(`{"k":"v"}`)
	task := q.NewTask(payload, 3)
	if task.ID == "" {
		t.Fatal("expected non-empty ID")
	}
	if string(task.Payload) != string(payload) {
		t.Fatalf("payload mismatch: %s vs %s", string(task.Payload), string(payload))
	}
	if task.MaxRetries != 3 || task.Attempt != 0 {
		t.Fatalf("unexpected retries/attempt: %d/%d", task.MaxRetries, task.Attempt)
	}
	if task.Status != q.StatusQueued {
		t.Fatalf("expected status queued, got %s", task.Status)
	}
	if task.CreatedAt.IsZero() || task.UpdatedAt.IsZero() {
		t.Fatal("timestamps must be set")
	}
}

func TestStoreSaveGetUpdate(t *testing.T) {
	s := q.NewStore()
	payload := json.RawMessage(`1`)
	task := q.NewTask(payload, 2)

	// Save new
	saved := s.Save(task)
	if saved.ID != task.ID {
		t.Fatalf("ids differ: %s vs %s", saved.ID, task.ID)
	}
	got, ok := s.Get(task.ID)
	if !ok {
		t.Fatal("expected to get saved task")
	}
	if got.ID != task.ID {
		t.Fatalf("get returned wrong task id: %s", got.ID)
	}

	// Update status
	time.Sleep(5 * time.Millisecond)
	updated, ok := s.UpdateStatus(task.ID, q.StatusRunning, 1)
	if !ok {
		t.Fatal("expected update to succeed")
	}
	if updated.Status != q.StatusRunning || updated.Attempt != 1 {
		t.Fatalf("unexpected status/attempt: %s/%d", updated.Status, updated.Attempt)
	}
	if !updated.UpdatedAt.After(updated.CreatedAt) {
		t.Fatal("UpdatedAt must be after CreatedAt")
	}
}
