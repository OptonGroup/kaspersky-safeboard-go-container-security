package httpserver

import (
	"context"
	"log"
	"net/http"
	"time"
)

// Server wraps the HTTP server and provides start/stop helpers.
type Server struct {
	httpServer *http.Server
}

// NewHandler constructs the HTTP handler (ServeMux) used by the server.
func NewHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/enqueue", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		// Stub: accept request without processing
		w.WriteHeader(http.StatusAccepted)
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
