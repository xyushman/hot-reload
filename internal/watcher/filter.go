package watcher

import (
	"path/filepath"
	"strings"
)

// IsIgnored returns true if the path matches predefined ignored patterns.
func IsIgnored(path string) bool {
	base := filepath.Base(path)

	// Ignore common directories
	if base == ".git" || base == "node_modules" || base == "vendor" || base == "tmp" || base == "bin" || base == "dist" {
		return true
	}

	// Ignore hidden files and directories except .env (if needed)
	// For this tool, we ignore all hidden files unless they end with .go
	if strings.HasPrefix(base, ".") {
		if !strings.HasSuffix(base, ".go") {
			return true
		}
	}

	// Ignore temporary editor files
	if strings.HasSuffix(base, "~") ||
		strings.HasSuffix(base, ".swp") ||
		strings.HasSuffix(base, ".tmp") ||
		strings.HasPrefix(base, ".#") ||
		base == "4913" || // Vim temp file
		strings.HasSuffix(base, ".swo") ||
		strings.HasSuffix(base, ".swn") {
		return true
	}

	// Ignore build artifacts
	artifacts := []string{".exe", ".out", ".test", ".o", ".a", ".so", ".dll"}
	for _, ext := range artifacts {
		if strings.HasSuffix(base, ext) {
			return true
		}
	}

	return false
}
