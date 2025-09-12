package config

import (
	"os"
	"strconv"
)

// Default configuration values
const (
	DefaultWorkers   = 4
	DefaultQueueSize = 64
)

// Config holds application configuration loaded from environment variables.
type Config struct {
	Workers   int
	QueueSize int
}

// Load reads configuration from environment with defaults and minimal validation.
func Load() Config {
	cfg := Config{
		Workers:   DefaultWorkers,
		QueueSize: DefaultQueueSize,
	}

	if v := os.Getenv("WORKERS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.Workers = n
		}
	}
	if v := os.Getenv("QUEUE_SIZE"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.QueueSize = n
		}
	}

	return cfg
}
