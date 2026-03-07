package watcher

import "testing"

func TestIsIgnored(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		// Go files (not ignored)
		{"root go file", "main.go", false},
		{"nested go file", "pkg/utils/util.go", false},
		{"hidden go file allowed", ".config.go", false},

		// Ignored directories
		{"git dir", ".git", true},
		{"node_modules dir", "node_modules", true},
		{"vendor dir", "vendor", true},
		{"tmp dir", "tmp", true},
		{"bin dir", "bin", true},
		{"dist dir", "dist", true},

		// Hidden files
		{"hidden env", ".env", true},
		{"hidden config file", ".config", true},

		// Editor temp files
		{"backup file", "main.go~", true},
		{"vim swap file", ".main.go.swp", true},
		{"vim swo file", ".main.go.swo", true},
		{"vim swn file", ".main.go.swn", true},
		{"tmp file", "file.tmp", true},
		{"lock file", ".#main.go", true},
		{"vim 4913 file", "4913", true},

		// Build artifacts
		{"windows binary", "server.exe", true},
		{"test binary", "server.test", true},
		{"build output", "server.out", true},
		{"object file", "file.o", true},
		{"static lib", "file.a", true},
		{"shared lib", "file.so", true},
		{"dll file", "file.dll", true},

		// Normal files
		{"text file", "README.md", false},
		{"config json", "config.json", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := IsIgnored(tt.path)
			if actual != tt.expected {
				t.Fatalf("IsIgnored(%q) = %v; want %v", tt.path, actual, tt.expected)
			}
		})
	}
}