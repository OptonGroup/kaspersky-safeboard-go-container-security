package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/optongroup/kaspersky-safeboard-go-container-security/internal/config"
	httpserver "github.com/optongroup/kaspersky-safeboard-go-container-security/internal/http"
	q "github.com/optongroup/kaspersky-safeboard-go-container-security/internal/queue"
)

func main() {
	cfg := config.Load()
	_ = cfg // will be used in next steps

	// Initialize queue and store
	store := q.NewStore()
	queueCh := make(chan q.Task, cfg.QueueSize)
	var accepting atomic.Bool
	accepting.Store(true)

	handler := httpserver.NewHandlerWithDeps(store, queueCh, &accepting)
	srv := httpserver.NewWithHandler(":8080", handler)

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start HTTP server
	srv.Start()

	// Handle OS signals for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	<-sigCh
	// Begin shutdown
	log.Println("Shutting down...")
	cancel()

	// Allow some time for graceful stop of HTTP server
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 5*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP shutdown error: %v", err)
	}

	wg.Wait()
	log.Println("Stopped")
}
