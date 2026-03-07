//go:build !windows

package builder

import (
	"context"
	"testing"
	"time"
)

func TestBuilder_Cancel(t *testing.T) {
	b := NewBuilder("sleep 10")
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error)
	go func() {
		errCh <- b.Build(ctx)
	}()

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	b.Cancel()

	err := <-errCh
	if err == nil {
		t.Errorf("expected error when cancelling build, got nil")
	}
}

func TestBuilder_Success(t *testing.T) {
	b := NewBuilder("echo 'hello'")
	
	ctx := context.Background()
	err := b.Build(ctx)

	if err != nil {
		t.Errorf("expected success, got %v", err)
	}
}
