package config

import (
	"os"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("WORKERS", "")
	t.Setenv("QUEUE_SIZE", "")
	cfg := Load()
	if cfg.Workers != DefaultWorkers {
		t.Fatalf("expected default workers %d, got %d", DefaultWorkers, cfg.Workers)
	}
	if cfg.QueueSize != DefaultQueueSize {
		t.Fatalf("expected default queue size %d, got %d", DefaultQueueSize, cfg.QueueSize)
	}
}

func TestLoadOverridesValid(t *testing.T) {
	t.Setenv("WORKERS", "8")
	t.Setenv("QUEUE_SIZE", "128")
	cfg := Load()
	if cfg.Workers != 8 {
		t.Fatalf("expected workers 8, got %d", cfg.Workers)
	}
	if cfg.QueueSize != 128 {
		t.Fatalf("expected queue size 128, got %d", cfg.QueueSize)
	}
}

func TestLoadInvalidValuesFallback(t *testing.T) {
	t.Setenv("WORKERS", "0")
	t.Setenv("QUEUE_SIZE", "-1")
	cfg := Load()
	if cfg.Workers != DefaultWorkers {
		t.Fatalf("expected default workers %d on invalid, got %d", DefaultWorkers, cfg.Workers)
	}
	if cfg.QueueSize != DefaultQueueSize {
		t.Fatalf("expected default queue size %d on invalid, got %d", DefaultQueueSize, cfg.QueueSize)
	}
}

// Ensure os.Setenv from outside is not required; t.Setenv cleans up.
var _ = os.Environ
