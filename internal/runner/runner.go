package runner

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"runtime"
	"sync"
	"time"
)

type Runner struct {
	cmdStr string
	mu     sync.Mutex
	cancel context.CancelFunc
}

func NewRunner(cmdStr string) *Runner {
	return &Runner{
		cmdStr: cmdStr,
	}
}

func (r *Runner) Start() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.cancel != nil {
		slog.Warn("Server is already running")
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	r.cancel = cancel

	go r.runLoop(ctx)
}

func (r *Runner) Stop() {
	r.mu.Lock()
	cancel := r.cancel
	r.cancel = nil
	r.mu.Unlock()

	if cancel != nil {
		cancel()
	}
}

func (r *Runner) runLoop(ctx context.Context) {
	backoff := 500 * time.Millisecond
	maxBackoff := 5 * time.Second

	for {
		if ctx.Err() != nil {
			return
		}

		slog.Info("Starting server process...")
		var cmd *exec.Cmd
		if runtime.GOOS == "windows" {
			cmd = exec.Command("powershell", "-Command", r.cmdStr)
		} else {
			cmd = exec.Command("sh", "-c", r.cmdStr)
		}

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			slog.Error("Failed to get stdout pipe", "error", err)
			return
		}

		stderr, err := cmd.StderrPipe()
		if err != nil {
			slog.Error("Failed to get stderr pipe", "error", err)
			return
		}

		startTime := time.Now()

		if err := cmd.Start(); err != nil {
			slog.Error("Failed to start server", "error", err)
			time.Sleep(backoff)
			continue
		}

		var wg sync.WaitGroup
		wg.Add(2)
		go streamLogs("server", stdout, &wg, slog.LevelInfo)
		go streamLogs("server", stderr, &wg, slog.LevelError)

		errCh := make(chan error, 1)
		go func() {
			wg.Wait()
			errCh <- cmd.Wait()
		}()

		var exitErr error
		select {
		case exitErr = <-errCh:
			// Process exited organically
		case <-ctx.Done():
			// We need to stop it!
			killProcessGroup(cmd)
			<-errCh // wait for actual exit
			slog.Info("Server stopped manually")
			return
		}

		uptime := time.Since(startTime)
		
		if exitErr != nil {
			slog.Error("Server process crashed", "error", exitErr, "uptime", uptime)
		} else {
			slog.Warn("Server process exited cleanly unexpectedly", "uptime", uptime)
		}

		if uptime < 2*time.Second {
			slog.Warn("Crash loop detected. Applying backoff", "delay", backoff)
			time.Sleep(backoff)
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		} else {
			backoff = 500 * time.Millisecond
		}
	}
}

func killProcessGroup(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	pid := cmd.Process.Pid
	if runtime.GOOS == "windows" {
		_ = exec.Command("taskkill", "/T", "/F", "/PID", fmt.Sprint(pid)).Run()
	} else {
		_ = cmd.Process.Kill() 
	}
}

func streamLogs(prefix string, r io.Reader, wg *sync.WaitGroup, level slog.Level) {
	defer wg.Done()
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if level == slog.LevelError {
			slog.Error("output", "source", prefix, "line", line)
		} else {
			slog.Info("output", "source", prefix, "line", line)
		}
	}
}
