package handlers

import (
	"testing"

	"github.com/hashicorp/golang-lru/v2"

	"grout/internal/config"
	"grout/internal/render"
)

func TestReadStaticFileSecurityDirectoryTraversal(t *testing.T) {
	renderer, err := render.New()
	if err != nil {
		t.Fatalf("renderer init: %v", err)
	}
	cache, _ := lru.New[string, []byte](1)
	cfg := config.DefaultServerConfig()
	cfg.StaticDir = "/tmp/test-static"
	svc := NewService(renderer, cache, cfg)

	tests := []struct {
		name     string
		filename string
		expected string
	}{
		{
			name:     "Valid filename",
			filename: "robots.txt",
			expected: "fallback", // File doesn't exist, should use fallback
		},
		{
			name:     "Directory traversal with ../",
			filename: "../../../etc/passwd",
			expected: "fallback", // Should be blocked
		},
		{
			name:     "Directory traversal with ..",
			filename: "../../config",
			expected: "fallback", // Should be blocked
		},
		{
			name:     "Absolute path",
			filename: "/etc/passwd",
			expected: "fallback", // Should be blocked
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.readStaticFile(tt.filename, "fallback")
			if result != tt.expected {
				t.Errorf("expected %q but got %q", tt.expected, result)
			}
		})
	}
}
