package runner

import (
	"runtime"
	"sync"
	"testing"
	"time"
)

// getSleepCommand returns a sleep command depending on OS
func getSleepCommand() string {
	if runtime.GOOS == "windows" {
		return "timeout /T 10 >nul"
	}
	return "sleep 10"
}

// getEchoCommand returns a simple command that exits immediately
func getEchoCommand() string {
	if runtime.GOOS == "windows" {
		return "echo hello"
	}
	return "echo 'hello'"
}

func TestRunner_Stop(t *testing.T) {
	r := NewRunner(getSleepCommand())

	r.Start()

	// Wait until cancel is set (process started)
	for i := 0; i < 10; i++ {
		r.mu.Lock()
		started := r.cancel != nil
		r.mu.Unlock()
		if started {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	r.Stop()

	// Wait until cancel is cleared
	for i := 0; i < 10; i++ {
		r.mu.Lock()
		c := r.cancel
		r.mu.Unlock()
		if c == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.cancel != nil {
		t.Errorf("Expected cancel to be nil after Stop, got %v", r.cancel)
	}
}

func TestRunner_RunLoopDoesNotCrashImmediately(t *testing.T) {
	// Command that exits immediately
	r := NewRunner(getEchoCommand())

	r.Start()

	// Wait a moment for runLoop to start and exit
	time.Sleep(100 * time.Millisecond)

	r.mu.Lock()
	c := r.cancel
	r.mu.Unlock()

	if c == nil {
		t.Errorf("Expected cancel to still be set after quick exit")
	}

	r.Stop()
}

func TestRunner_ConcurrentStartStop(t *testing.T) {
	r := NewRunner(getSleepCommand())
	var wg sync.WaitGroup

	// Concurrently start and stop
	for i := 0; i < 5; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			r.Start()
		}()
		go func() {
			defer wg.Done()
			r.Stop()
		}()
	}

	wg.Wait()

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.cancel != nil {
		t.Errorf("Expected cancel to be nil after Stop, got %v", r.cancel)
	}
}
