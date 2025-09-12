package tests

import (
	"context"
	"math/rand"
	"testing"
	"time"

	q "github.com/optongroup/kaspersky-safeboard-go-container-security/internal/queue"
)

func TestBackoffGrowthAndBounds(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	d0 := q.BackoffDelay(200*time.Millisecond, 0, 100*time.Millisecond, rng)
	d1 := q.BackoffDelay(200*time.Millisecond, 1, 100*time.Millisecond, rng)
	d2 := q.BackoffDelay(200*time.Millisecond, 2, 100*time.Millisecond, rng)
	if !(d1 >= d0 && d2 >= d1) {
		t.Fatalf("expected non-decreasing backoff, got %v, %v, %v", d0, d1, d2)
	}
	if d0 < 200*time.Millisecond || d0 > 300*time.Millisecond { // base..base+jitter
		t.Fatalf("d0 out of bounds: %v", d0)
	}
}

func TestTryEnqueueWithContext_RespectCtx(t *testing.T) {
	ch := make(chan q.Task, 1)
	// fill channel
	ch <- q.Task{ID: "x"}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	ok := q.TryEnqueueWithContext(ctx, ch, q.Task{ID: "y"}, 10*time.Millisecond)
	if ok {
		t.Fatal("expected false when context timeout while queue is full")
	}
}
