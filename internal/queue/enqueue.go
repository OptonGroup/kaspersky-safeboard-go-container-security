package queue

import (
	"context"
	"time"
)

// TryEnqueueWithContext tries to enqueue task non-blockingly, retrying when the channel is full
// until context is done. Small sleep is used between attempts to avoid busy loop.
func TryEnqueueWithContext(ctx context.Context, ch chan<- Task, t Task, retrySleep time.Duration) bool {
	if retrySleep <= 0 {
		retrySleep = 10 * time.Millisecond
	}
	for {
		select {
		case <-ctx.Done():
			return false
		default:
		}
		select {
		case ch <- t:
			return true
		default:
			time.Sleep(retrySleep)
		}
	}
}
