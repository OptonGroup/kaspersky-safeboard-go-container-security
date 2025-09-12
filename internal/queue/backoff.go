package queue

import (
	"math/rand"
	"time"
)

const (
	BackoffBase = 200 * time.Millisecond
	JitterMax   = 100 * time.Millisecond
)

// BackoffDelay calculates exponential backoff with jitter.
// delay = base * 2^attempt + jitter(0..jitterMax)
func BackoffDelay(base time.Duration, attempt int, jitterMax time.Duration, rng *rand.Rand) time.Duration {
	if attempt < 0 {
		attempt = 0
	}
	if base <= 0 {
		base = BackoffBase
	}
	if jitterMax < 0 {
		jitterMax = 0
	}
	// Compute exponential component with overflow guard
	maxShift := 30
	if attempt > maxShift {
		attempt = maxShift
	}
	exp := base * time.Duration(1<<attempt)
	var jitter time.Duration
	if jitterMax > 0 && rng != nil {
		jitter = time.Duration(rng.Int63n(int64(jitterMax) + 1))
	}
	return exp + jitter
}
