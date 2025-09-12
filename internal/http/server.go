package httpserver

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	q "github.com/optongroup/kaspersky-safeboard-go-container-security/internal/queue"
)

// Server wraps the HTTP server and provides start/stop helpers.
type Server struct {
	httpServer *http.Server
}

// NewHandler constructs the HTTP handler (ServeMux) used by the server.
func NewHandler() http.Handler {
	// default dependencies for backward compatibility
	store := q.NewStore()
	ch := make(chan q.Task, 1)
	var accepting atomic.Bool
	accepting.Store(true)
	return NewHandlerWithDeps(store, ch, &accepting)
}

// NewHandlerWithDeps builds handler with injected store, queue channel and accepting flag
func NewHandlerWithDeps(store *q.Store, ch chan<- q.Task, accepting *atomic.Bool) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	type enqueueRequest struct {
		ID         string `json:"id"`
		Payload    string `json:"payload"`
		MaxRetries int    `json:"max_retries"`
	}
	type enqueueResponse struct {
		ID     string       `json:"id"`
		Status q.TaskStatus `json:"status"`
	}

	mux.HandleFunc("/enqueue", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if !accepting.Load() {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		defer r.Body.Close()
		var req enqueueRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(req.ID) == "" || strings.TrimSpace(req.Payload) == "" {
			http.Error(w, "id and payload required", http.StatusBadRequest)
			return
		}
		if req.MaxRetries < 0 {
			req.MaxRetries = 0
		}
		// check duplicate id
		if _, exists := store.Get(req.ID); exists {
			http.Error(w, "duplicate id", http.StatusBadRequest)
			return
		}
		task := q.NewTaskWithID(req.ID, []byte(req.Payload), req.MaxRetries)
		select {
		case ch <- task:
			store.Save(task)
			log.Printf("enqueued task id=%s", task.ID)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusAccepted)
			_ = json.NewEncoder(w).Encode(enqueueResponse{ID: task.ID, Status: task.Status})
		default:
			w.WriteHeader(http.StatusServiceUnavailable)
		}
	})

	// GET /status/{id}
	mux.HandleFunc("/status/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		id := strings.TrimPrefix(r.URL.Path, "/status/")
		if id == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if t, ok := store.Get(id); ok {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(t)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})

	// GET /metrics (simple JSON counters)
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		m := store.GetMetrics()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(m)
	})

	return mux
}

// New creates a new HTTP server bound to addr with handlers set up.
func New(addr string) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:              addr,
			Handler:           NewHandler(),
			ReadHeaderTimeout: 5 * time.Second,
		},
	}
}

// NewWithHandler creates a server with provided handler.
func NewWithHandler(addr string, handler http.Handler) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:              addr,
			Handler:           handler,
			ReadHeaderTimeout: 5 * time.Second,
		},
	}
}

// Start launches the HTTP server in a separate goroutine.
func (s *Server) Start() {
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("http server error: %v", err)
		}
	}()
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
