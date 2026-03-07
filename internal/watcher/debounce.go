package watcher

import (
	"sync"
	"testing"
	"time"
)

func TestDebounce(t *testing.T) {
	in := make(chan struct{})
	out := Debounce(in, 100*time.Millisecond)

	var wg sync.WaitGroup
	var mu sync.Mutex
	triggerCount := 0

	wg.Add(1)
	go func() {
		defer wg.Done()
		timeout := time.After(500 * time.Millisecond)

		for {
			select {
			case <-out:
				mu.Lock()
				triggerCount++
				mu.Unlock()
			case <-timeout:
				return
			}
		}
	}()

	// Burst of events (should result in one trigger)
	in <- struct{}{}
	in <- struct{}{}
	in <- struct{}{}

	time.Sleep(200 * time.Millisecond)

	// Another event after debounce window
	in <- struct{}{}

	time.Sleep(200 * time.Millisecond)

	close(in)
	wg.Wait()

	mu.Lock()
	defer mu.Unlock()

	if triggerCount != 2 {
		t.Fatalf("expected 2 triggers, got %d", triggerCount)
	}
}
