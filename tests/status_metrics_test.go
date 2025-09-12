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

func TestStatusAndMetrics(t *testing.T) {
	store := q.NewStore()
	ch := make(chan q.Task, 2)
	var accepting atomic.Bool
	accepting.Store(true)
	h := httpserver.NewHandlerWithDeps(store, ch, &accepting)

	// enqueue one task
	req := httptest.NewRequest(http.MethodPost, "/enqueue", bytes.NewReader([]byte(`{"payload": {}}`)))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", rr.Code)
	}
	var resp struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil || resp.ID == "" {
		t.Fatalf("invalid enqueue response: %v id=%q", err, resp.ID)
	}

	// check status endpoint
	reqS := httptest.NewRequest(http.MethodGet, "/status/"+resp.ID, nil)
	rrS := httptest.NewRecorder()
	h.ServeHTTP(rrS, reqS)
	if rrS.Code != http.StatusOK {
		t.Fatalf("expected 200 for status, got %d", rrS.Code)
	}

	// check 404
	req404 := httptest.NewRequest(http.MethodGet, "/status/nonexistent", nil)
	rr404 := httptest.NewRecorder()
	h.ServeHTTP(rr404, req404)
	if rr404.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown id, got %d", rr404.Code)
	}

	// metrics should have at least queued >= 1
	reqM := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rrM := httptest.NewRecorder()
	h.ServeHTTP(rrM, reqM)
	if rrM.Code != http.StatusOK {
		t.Fatalf("expected 200 for metrics, got %d", rrM.Code)
	}
	var m struct{ Queued uint64 }
	if err := json.Unmarshal(rrM.Body.Bytes(), &m); err != nil {
		t.Fatalf("invalid metrics json: %v", err)
	}
	if m.Queued < 1 {
		t.Fatalf("expected queued >=1, got %d", m.Queued)
	}
}
