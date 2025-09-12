package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	httpserver "github.com/optongroup/kaspersky-safeboard-go-container-security/internal/http"
	q "github.com/optongroup/kaspersky-safeboard-go-container-security/internal/queue"
)

func newTestHandler(queueSize int, accepting bool, store *q.Store) (http.Handler, chan q.Task, *atomic.Bool) {
	ch := make(chan q.Task, queueSize)
	var acc atomic.Bool
	acc.Store(accepting)
	if store == nil {
		store = q.NewStore()
	}
	h := httpserver.NewHandlerWithDeps(store, ch, &acc)
	return h, ch, &acc
}

func TestEnqueue_ValidAccepted(t *testing.T) {
	store := q.NewStore()
	h, ch, _ := newTestHandler(1, true, store)
	body := map[string]any{"payload": json.RawMessage(`{"a":1}`), "max_retries": 2}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/enqueue", bytes.NewReader(b))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", rr.Code)
	}
	select {
	case <-ch:
	default:
		t.Fatal("task was not enqueued")
	}
}

func TestEnqueue_QueueFull_503(t *testing.T) {
	h, ch, _ := newTestHandler(1, true, nil)
	// Fill the queue
	ch <- q.Task{ID: "x"}
	req := httptest.NewRequest(http.MethodPost, "/enqueue", bytes.NewReader([]byte(`{"payload": {}}`)))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 when queue full, got %d", rr.Code)
	}
}

func TestEnqueue_AcceptingFalse_503(t *testing.T) {
	_, _, acc := newTestHandler(1, false, nil)
	h, _, _ := newTestHandler(1, false, nil)
	if acc.Load() != false {
		t.Fatal("accepting must be false")
	}
	req := httptest.NewRequest(http.MethodPost, "/enqueue", bytes.NewReader([]byte(`{"payload": {}}`)))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 when accepting=false, got %d", rr.Code)
	}
}

func TestEnqueue_InvalidJSON_400(t *testing.T) {
	h, _, _ := newTestHandler(1, true, nil)
	req := httptest.NewRequest(http.MethodPost, "/enqueue", bytes.NewReader([]byte(`{`)))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid json, got %d", rr.Code)
	}
}
