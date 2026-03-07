package watcher

import (
	"testing"
)

func TestIsIgnored(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		// Go files
		{"main.go", false},
		{"pkg/utils/util.go", false},

		// Ignored Directories
		{".git", true},
		{"node_modules", true},
		{"vendor", true},
		{"tmp", true},

		// Hidden files
		{".env", true}, // By our simple rule we ignore all non-.go hidden files

		// Editor temp files
		{"main.go~", true},
		{".main.go.swp", true},
		{"4913", true}, // Vim

		// Build artifacts
		{"server.exe", true},
		{"server.out", true},
		{"server.test", true},
	}

	for _, tt := range tests {
		actual := IsIgnored(tt.path)
		if actual != tt.expected {
			t.Errorf("IsIgnored(%q) = %v; want %v", tt.path, actual, tt.expected)
		}
	}
}
