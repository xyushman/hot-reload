package watcher

import (
	"sync"
	"testing"
	"time"
)

func TestDebounce(t *testing.T) {
	in := make(chan struct{})
	out := Debounce(in, 100*time.Millisecond)

	var mu sync.Mutex
	triggerCount := 0

	done := make(chan struct{})

	go func() {
		for {
			select {
			case <-out:
				mu.Lock()
				triggerCount++
				mu.Unlock()
			case <-done:
				return
			}
		}
	}()

	// Rapid events (should debounce into one)
	in <- struct{}{}
	in <- struct{}{}
	in <- struct{}{}

	time.Sleep(200 * time.Millisecond)

	// Second burst
	in <- struct{}{}

	time.Sleep(200 * time.Millisecond)

	close(done)

	mu.Lock()
	defer mu.Unlock()

	if triggerCount != 2 {
		t.Errorf("Expected 2 triggers, got %d", triggerCount)
	}
}