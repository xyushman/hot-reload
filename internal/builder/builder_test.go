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

	// Cancel the context after a short delay
	time.Sleep(100 * time.Millisecond)
	cancel() // propagate cancellation

	err := <-errCh
	if err == nil {
		t.Fatalf("expected error when cancelling build, got nil")
	}
}

func TestBuilder_Success(t *testing.T) {
	b := NewBuilder("echo 'hello'")

	ctx := context.Background()
	err := b.Build(ctx)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
}
