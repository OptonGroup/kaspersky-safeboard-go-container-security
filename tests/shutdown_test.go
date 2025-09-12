package tests

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	httpserver "github.com/optongroup/kaspersky-safeboard-go-container-security/internal/http"
	q "github.com/optongroup/kaspersky-safeboard-go-container-security/internal/queue"
)

func TestGracefulShutdown_StopAcceptingAndWorkersFinish(t *testing.T) {
	store := q.NewStore()
	ch := make(chan q.Task, 2)
	var accepting atomic.Bool
	accepting.Store(true)
	handler := httpserver.NewHandlerWithDeps(store, ch, &accepting)

	// Start workers
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	q.StartWorkers(ctx, &wg, store, ch, 2, 42)

	// Enqueue one task
	req := httptest.NewRequest(http.MethodPost, "/enqueue", bytes.NewReader([]byte(`{"id":"g1","payload":"p"}`)))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", rr.Code)
	}

	// Trigger graceful: stop accepting
	accepting.Store(false)

	// New enqueues must be rejected
	req2 := httptest.NewRequest(http.MethodPost, "/enqueue", bytes.NewReader([]byte(`{"id":"g2","payload":"p"}`)))
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 after stop accepting, got %d", rr2.Code)
	}

	// Cancel workers and wait
	cancel()
	close(ch)
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("workers did not finish in time")
	}
}
