package queue

import (
	"context"
	"math/rand"
	"sync"
	"time"
)

// StartWorkers launches numWorkers goroutines that consume tasks from queueCh until ctx is done.
// Each worker marks task as running, simulates processing for 100-500ms, and finishes with
// either done or failed status, with approximately 20% failure probability.
// To keep tests deterministic, pass a seed; each worker derives its own independent RNG from this seed.
func StartWorkers(ctx context.Context, wg *sync.WaitGroup, store *Store, queueCh <-chan Task, numWorkers int, seed int64) {
	if numWorkers <= 0 {
		return
	}
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		workerSeed := seed + int64(i+1)
		go func(localSeed int64) {
			defer wg.Done()
			rng := rand.New(rand.NewSource(localSeed))
			for {
				select {
				case <-ctx.Done():
					return
				case t, ok := <-queueCh:
					if !ok {
						return
					}
					// Mark running
					store.UpdateStatus(t.ID, StatusRunning, t.Attempt)
					// Simulate processing time 100-500ms
					sleepMs := 100 + rng.Intn(401) // [100,500]
					timer := time.NewTimer(time.Duration(sleepMs) * time.Millisecond)
					select {
					case <-ctx.Done():
						if !timer.Stop() {
							<-timer.C
						}
						return
					case <-timer.C:
					}

					// 20% failure probability
					if rng.Intn(100) < 20 {
						// failed without retries at this step
						store.UpdateStatus(t.ID, StatusFailed, t.Attempt)
						continue
					}
					store.UpdateStatus(t.ID, StatusDone, t.Attempt)
				}
			}
		}(workerSeed)
	}
}
