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
# Unix/Linux
go build -o hotreload ./cmd/hotreload

# Windows
go build -o hotreload.exe ./cmd/hotreload
```

## Usage

Provide the root path, the build command, and the execute command.

```bash
# Unix/Linux
./hotreload --root ./testserver --build "go build -o ./testserver/server ./testserver/main.go" --exec "./testserver/server"

# Windows (run from the hotreload directory)
.\hotreload.exe --root .\testserver --build "go build -o .\testserver\server.exe .\testserver\main.go" --exec ".\testserver\server.exe"
```

## Architecture Map
- **Watcher (`internal/watcher`)**: Wraps `fsnotify` into a recursive watcher. Handles dynamic directories. Returns a channel of relevant change events.
- **Filter (`internal/watcher/filter.go`)**: Simple rule engine discarding noisy paths.
- **Debounce (`internal/watcher/debounce.go`)**: Collapses multiple events within `300ms` into a single trigger.
- **Builder (`internal/builder`)**: Handles executing the build command using `sh -c` (or `cmd /c` on Windows). Uses context cancellation and process termination (e.g., `taskkill` on Windows) to abort builds immediately. Streams outputs via `slog`.
- **Runner (`internal/runner`)**: Controls the runtime server. Monitors for crashes (`wait()` < 2 seconds) and applies backoff. Exposes a `Stop()` method mapping to graceful takedowns or process termination.

## Known Limitations
- The ignored files logic is currently hard-coded. Expanding this tool to honor `.gitignore` naturally requires building a matcher package.
