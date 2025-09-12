package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	httpserver "github.com/optongroup/kaspersky-safeboard-go-container-security/internal/http"
	q "github.com/optongroup/kaspersky-safeboard-go-container-security/internal/queue"
)

// Simple integration: enqueue several tasks via HTTP, run workers with deterministic seed, check final statuses present.
func TestIntegration_EnqueueAndProcess(t *testing.T) {
	store := q.NewStore()
	const N = 12
	ch := make(chan q.Task, N)
	var accepting atomic.Bool
	accepting.Store(true)
	handler := httpserver.NewHandlerWithDeps(store, ch, &accepting)

	// Start workers with deterministic seed
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	q.StartWorkers(ctx, &wg, store, ch, 3, 123)

	// Enqueue N tasks via HTTP and collect IDs
	ids := make([]string, 0, N)
	for i := 0; i < N; i++ {
		b := []byte(fmt.Sprintf(`{"id":"t-%d","payload":"px","max_retries":0}`, i))
		req := httptest.NewRequest(http.MethodPost, "/enqueue", bytes.NewReader(b))
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusAccepted {
			t.Fatalf("enqueue %d -> %d", i, rr.Code)
		}
		var resp struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil || resp.ID == "" {
			t.Fatalf("invalid enqueue response: %v id=%q", err, resp.ID)
		}
		ids = append(ids, resp.ID)
		// give workers a tiny chance to drain
		time.Sleep(10 * time.Millisecond)
	}

	// Wait until processed
	time.Sleep(1500 * time.Millisecond)
	cancel()
	close(ch)
	wg.Wait()

	// Verify final statuses for enqueued tasks
	for _, id := range ids {
		got, ok := store.Get(id)
		if !ok {
			t.Fatalf("task %s missing", id)
		}
		if got.Status != q.StatusDone && got.Status != q.StatusFailed {
			t.Fatalf("unexpected final status for %s: %s", id, got.Status)
		}
	}
}
