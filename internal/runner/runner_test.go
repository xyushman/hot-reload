//go:build !windows

package runner

import (
	"testing"
	"time"
)

func TestRunner_Stop(t *testing.T) {
	r := NewRunner("sleep 10")
	
	r.Start()
	
	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)
	
	r.Stop()
	
	// Wait a moment for process to exit
	time.Sleep(100 * time.Millisecond)
	
	r.mu.Lock()
	cancel := r.cancel
	r.mu.Unlock()
	
	if cancel != nil {
		t.Errorf("Expected cancel to be nil after Stop")
	}
}
