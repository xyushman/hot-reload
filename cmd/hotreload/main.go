package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"hotreload/internal/builder"
	"hotreload/internal/runner"
	"hotreload/internal/watcher"
)

func main() {
	rootFlag := flag.String("root", ".", "Directory to watch (including all subfolders)")
	buildFlag := flag.String("build", "", "Shell command to build the project")
	execFlag := flag.String("exec", "", "Shell command to run the built server")
	flag.Parse()

	if *buildFlag == "" || *execFlag == "" {
		slog.Error("--build and --exec flags are required")
		os.Exit(1)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo, AddSource: true}))
	slog.SetDefault(logger)

	slog.Info("Starting hotreload", "root", *rootFlag, "build", *buildFlag, "exec", *execFlag)

	w, err := watcher.NewWatcher(*rootFlag)
	if err != nil {
		slog.Error("Failed to initialize watcher", "error", err)
		os.Exit(1)
	}

	b := builder.NewBuilder(*buildFlag)
	r := runner.NewRunner(*execFlag)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		slog.Info("Received shutdown signal, stopping...")
		cancel()
	}()

	rawEvents := w.Start(ctx)
	debouncedEvents := watcher.Debounce(rawEvents, 300*time.Millisecond)

	// Initial Build
	doBuildAndRun(ctx, b, r)

	// Main event loop
	for {
		select {
		case <-ctx.Done():
			// Cleanup
			b.Cancel()
			r.Stop()
			slog.Info("Shutdown complete")
			return
		case <-debouncedEvents:
			slog.Info("File changes detected. Rebuilding...")
			doBuildAndRun(ctx, b, r)
		}
	}
}

func doBuildAndRun(ctx context.Context, b *builder.Builder, r *runner.Runner) {
	// Cancel any ongoing build
	b.Cancel()

	err := b.Build(ctx)
	if err != nil {
		// Output is already streamed to slog by b.Build
		slog.Error("Build step failed. Waiting for next file change...")
		return
	}

	// Build successful, restart the server
	slog.Info("Build successful. Restarting server...")
	r.Stop()
	r.Start()
}
