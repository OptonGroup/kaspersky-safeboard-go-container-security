package tests

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	httpserver "github.com/optongroup/kaspersky-safeboard-go-container-security/internal/http"
	q "github.com/optongroup/kaspersky-safeboard-go-container-security/internal/queue"
)

func TestHealthz(t *testing.T) {
	handler := httpserver.NewHandler()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestMethodNotAllowed(t *testing.T) {
	handler := httpserver.NewHandler()
	tests := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/healthz"},
		{http.MethodGet, "/enqueue"},
	}
	for _, tc := range tests {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusMethodNotAllowed {
			t.Fatalf("%s %s expected 405, got %d", tc.method, tc.path, rr.Code)
		}
	}
}

func TestEnqueueStub(t *testing.T) {
	handler := httpserver.NewHandler()
	req := httptest.NewRequest(http.MethodPost, "/enqueue", bytes.NewReader([]byte(`{"id":"stub","payload":"p"}`)))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", rr.Code)
	}
}

func TestHealthz_IndependentOfQueueAndAccepting(t *testing.T) {
	h, ch, acc := newTestHandler(1, false, nil)
	if acc.Load() {
		t.Fatal("accepting must be false for this test")
	}
	// fill queue
	ch <- q.Task{ID: "x"}
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("/healthz must return 200 regardless of queue, got %d", rr.Code)
	}
}
