package builder

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"runtime"
	"sync"
)

// Builder handles the compilation of the Go project.
type Builder struct {
	cmdStr string
	cmd    *exec.Cmd
	mu     sync.Mutex
}

// NewBuilder creates a new Builder with the given shell command.
func NewBuilder(cmdStr string) *Builder {
	return &Builder{
		cmdStr: cmdStr,
	}
}

// Build executes the build command. It can be cancelled via the context.
func (b *Builder) Build(ctx context.Context) error {
	b.mu.Lock()
	if b.cmd != nil {
		b.mu.Unlock()
		return fmt.Errorf("build already in progress")
	}

	// Execute via shell to handle complex commands and variable expansions
	if runtime.GOOS == "windows" {
		b.cmd = exec.CommandContext(ctx, "powershell", "-Command", b.cmdStr)
	} else {
		b.cmd = exec.CommandContext(ctx, "sh", "-c", b.cmdStr)
	}
	
	// Create a process group so we can kill all subprocesses if cancelled
	// Removed SysProcAttr for Windows compatibility

	stdout, err := b.cmd.StdoutPipe()
	if err != nil {
		b.mu.Unlock()
		return err
	}
	stderr, err := b.cmd.StderrPipe()
	if err != nil {
		b.mu.Unlock()
		return err
	}

	b.mu.Unlock()

	// Stream logs in real-time
	var wg sync.WaitGroup
	wg.Add(2)
	go streamLogs("build", stdout, &wg, slog.LevelInfo)
	go streamLogs("build", stderr, &wg, slog.LevelError)

	err = b.cmd.Start()
	if err != nil {
		b.clearCmd()
		return err
	}

	wg.Wait()
	err = b.cmd.Wait()
	b.clearCmd()

	if err != nil {
		return fmt.Errorf("build failed: %w", err)
	}
	return nil
}

// Cancel attempts to terminate the ongoing build process group.
func (b *Builder) Cancel() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.cmd != nil && b.cmd.Process != nil {
		// Send SIGKILL to the entire process group
		pid := b.cmd.Process.Pid
		if runtime.GOOS == "windows" {
			_ = exec.Command("taskkill", "/T", "/F", "/PID", fmt.Sprint(pid)).Run()
		} else {
			_ = b.cmd.Process.Kill()
		}
		slog.Debug("Cancelled ongoing build", "pid", pid)
	}
}

func (b *Builder) clearCmd() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.cmd = nil
}

// streamLogs reads from 'r' line by line and logs it with the given level and prefix.
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
