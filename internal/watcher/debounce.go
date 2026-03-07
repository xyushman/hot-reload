package watcher

import (
	"sync"
	"time"
)

// Debounce takes a channel of struct{} (events) and emits a single struct{}
// after the idle duration has passed without any new events.
func Debounce(events <-chan struct{}, idle time.Duration) <-chan struct{} {
	trigger := make(chan struct{})

	go func() {
		var timer *time.Timer
		var mu sync.Mutex

		for range events {
			mu.Lock()
			if timer != nil {
				timer.Stop()
			}
			timer = time.AfterFunc(idle, func() {
				trigger <- struct{}{}
			})
			mu.Unlock()
		}
	}()

	return trigger
}
