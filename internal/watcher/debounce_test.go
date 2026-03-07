package watcher

import (
	"testing"
	"time"
)

func TestDebounce(t *testing.T) {
	in := make(chan struct{})
	out := Debounce(in, 100*time.Millisecond)

	triggerCount := 0

	go func() {
		for range out {
			triggerCount++
		}
	}()

	in <- struct{}{}
	in <- struct{}{}
	in <- struct{}{}

	time.Sleep(200 * time.Millisecond)

	in <- struct{}{}

	time.Sleep(200 * time.Millisecond)

	if triggerCount != 2 {
		t.Errorf("Expected 2 triggers, got %d", triggerCount)
	}
}
