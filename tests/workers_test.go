package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	q "github.com/optongroup/kaspersky-safeboard-go-container-security/internal/queue"
)

func TestWorkers_ProcessTasks_StatusTransitions(t *testing.T) {
	store := q.NewStore()
	ch := make(chan q.Task, 10)
	seed := int64(1)
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	q.StartWorkers(ctx, &wg, store, ch, 4, seed)

	// enqueue several tasks
	ids := make([]string, 0, 8)
	for i := 0; i < 8; i++ {
		payload := json.RawMessage([]byte(fmt.Sprintf(`{"i":%d}`, i)))
		tsk := q.NewTask(payload, 0)
		saved := store.Save(tsk)
		ids = append(ids, saved.ID)
		ch <- saved
	}

	// wait for processing to complete
	time.Sleep(1200 * time.Millisecond)
	close(ch)
	wg.Wait()
	cancel()

	// verify statuses are either done or failed
	for _, id := range ids {
		got, ok := store.Get(id)
		if !ok {
			t.Fatalf("task %s missing", id)
		}
		if got.Status != q.StatusDone && got.Status != q.StatusFailed {
			t.Fatalf("unexpected final status: %s", got.Status)
		}
	}
}
