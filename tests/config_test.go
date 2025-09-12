package tests

import (
    cfg "github.com/optongroup/kaspersky-safeboard-go-container-security/internal/config"
    "testing"
)

func TestLoadDefaults(t *testing.T) {
    t.Setenv("WORKERS", "")
    t.Setenv("QUEUE_SIZE", "")
    c := cfg.Load()
    if c.Workers != cfg.DefaultWorkers {
        t.Fatalf("expected default workers %d, got %d", cfg.DefaultWorkers, c.Workers)
    }
    if c.QueueSize != cfg.DefaultQueueSize {
        t.Fatalf("expected default queue size %d, got %d", cfg.DefaultQueueSize, c.QueueSize)
    }
}

func TestLoadOverridesValid(t *testing.T) {
    t.Setenv("WORKERS", "8")
    t.Setenv("QUEUE_SIZE", "128")
    c := cfg.Load()
    if c.Workers != 8 {
        t.Fatalf("expected workers 8, got %d", c.Workers)
    }
    if c.QueueSize != 128 {
        t.Fatalf("expected queue size 128, got %d", c.QueueSize)
    }
}

func TestLoadInvalidValuesFallback(t *testing.T) {
    t.Setenv("WORKERS", "0")
    t.Setenv("QUEUE_SIZE", "-1")
    c := cfg.Load()
    if c.Workers != cfg.DefaultWorkers {
        t.Fatalf("expected default workers %d on invalid, got %d", cfg.DefaultWorkers, c.Workers)
    }
    if c.QueueSize != cfg.DefaultQueueSize {
        t.Fatalf("expected default queue size %d on invalid, got %d", cfg.DefaultQueueSize, c.QueueSize)
    }
}


