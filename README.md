# Hotreload

`hotreload` is a robust command-line tool that watches a Go project folder for file changes, automatically rebuilds the project, and restarts the server efficiently. It focuses on a smooth developer experience by debouncing rapid editor events, securely managing child processes (via process groups), and avoiding rapid crash loops through exponential backoff.

## Features
- **File Watching (fsnotify)**: Recursively watches deeply nested directories out of the box. Automatically watches newly created directories and un-watches removed ones.
- **Debouncing**: Editors (like Vim, VS Code) emit multiple events sequentially (e.g. `write`, `chmod`, `rename`). Hotreload debounces them tightly into a single build trigger.
- **Process Group Termination**: Kills the *entire process tree* of both the build process and your server, leaving no orphaned binaries. Handles stubborn servers by sending `SIGTERM` followed by `SIGKILL`.
- **Crash Loop Protection**: Detects immediate crashes upon startup and applies an exponential backoff before the next automated restart attempt.
- **Intelligent Filtering**: Hard-coded heuristics to ignore common non-Go directories (`.git`, `node_modules`, `vendor`) and temp/editor/build artifacts (`*.swp`, `*.o`, `*.exe`).

## Installation

```bash
go mod tidy
go build -o hotreload ./cmd/hotreload
```

## Usage

Provide the root path, the build command, and the execute command.

```bash
./hotreload --root ./testserver --build "go build -o ./testserver/server ./testserver/main.go" --exec "./testserver/server"
```

## Architecture Map
- **Watcher (`internal/watcher`)**: Wraps `fsnotify` into a recursive watcher. Handles dynamic directories. Returns a channel of relevant change events.
- **Filter (`internal/watcher/filter.go`)**: Simple rule engine discarding noisy paths.
- **Debounce (`internal/watcher/debounce.go`)**: Collapses multiple events within `300ms` into a single trigger.
- **Builder (`internal/builder`)**: Handles executing the build command using `sh -c`. Uses context cancellation tied to process group killing to abort builds immediately. Streams outputs via `slog`.
- **Runner (`internal/runner`)**: Controls the runtime server. Monitors for crashes (`wait()` < 2 seconds) and applies backoff. Exposes a `Stop()` method mapping to `SIGTERM->SIGKILL` graceful takedowns.

## Known Limitations
- Built specifically for Unix systems implementing `syscall.SysProcAttr{Setpgid: true}`. Running this on Windows will encounter compilation/execution compatibility errors regarding process boundaries.
- The ignored files logic is currently hard-coded. Expanding this tool to honor `.gitignore` naturally requires building a matcher package.
