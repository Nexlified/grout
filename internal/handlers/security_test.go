package handlers

import (
	"os"
	"path/filepath"
	"strings"
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
			name:     "Absolute path Unix",
			filename: "/etc/passwd",
			expected: "fallback", // Should be blocked
		},
		{
			name:     "Absolute path Windows",
			filename: "C:\\Windows\\System32\\config",
			expected: "fallback", // Should be blocked
		},
		{
			name:     "Empty filename",
			filename: "",
			expected: "fallback", // Should be blocked or use fallback
		},
		{
			name:     "Multiple path separators",
			filename: "//robots.txt",
			expected: "fallback", // Should be blocked (absolute path)
		},
		{
			name:     "Path with backslash Windows style",
			filename: "..\\..\\config",
			expected: "fallback", // Should be blocked after Clean()
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

func TestReadStaticFileSuccess(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "grout-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files
	testFiles := map[string]string{
		"robots.txt":      "User-agent: *\nDisallow: /",
		"sitemap.xml":     "<?xml version=\"1.0\"?><urlset></urlset>",
		"custom.txt":      "custom content",
		"with-domain.txt": "Site: {{DOMAIN}}",
		"subdir/file.txt": "nested file",
	}

	for path, content := range testFiles {
		fullPath := filepath.Join(tmpDir, path)
		// Create subdirectories if needed
		if dir := filepath.Dir(fullPath); dir != tmpDir {
			if err := os.MkdirAll(dir, 0755); err != nil {
				t.Fatalf("failed to create subdir: %v", err)
			}
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write test file %s: %v", path, err)
		}
	}

	renderer, err := render.New()
	if err != nil {
		t.Fatalf("renderer init: %v", err)
	}
	cache, _ := lru.New[string, []byte](1)
	cfg := config.DefaultServerConfig()
	cfg.StaticDir = tmpDir
	svc := NewService(renderer, cache, cfg)

	tests := []struct {
		name     string
		filename string
		expected string
	}{
		{
			name:     "Read robots.txt",
			filename: "robots.txt",
			expected: "User-agent: *\nDisallow: /",
		},
		{
			name:     "Read sitemap.xml",
			filename: "sitemap.xml",
			expected: "<?xml version=\"1.0\"?><urlset></urlset>",
		},
		{
			name:     "Read custom file",
			filename: "custom.txt",
			expected: "custom content",
		},
		{
			name:     "Read file with {{DOMAIN}} placeholder",
			filename: "with-domain.txt",
			expected: "Site: {{DOMAIN}}", // Note: Template replacement happens in handlers, not readStaticFile
		},
		{
			name:     "Read nested file",
			filename: "subdir/file.txt",
			expected: "nested file",
		},
		{
			name:     "Non-existent file uses fallback",
			filename: "nonexistent.txt",
			expected: "fallback content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.readStaticFile(tt.filename, "fallback content")
			if result != tt.expected {
				t.Errorf("expected %q but got %q", tt.expected, result)
			}
		})
	}
}

func TestReadStaticFileTemplateReplacement(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "grout-test-template-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a robots.txt with {{DOMAIN}} placeholder
	robotsContent := "User-agent: *\nSitemap: https://{{DOMAIN}}/sitemap.xml"
	if err := os.WriteFile(filepath.Join(tmpDir, "robots.txt"), []byte(robotsContent), 0644); err != nil {
		t.Fatalf("failed to write robots.txt: %v", err)
	}

	renderer, err := render.New()
	if err != nil {
		t.Fatalf("renderer init: %v", err)
	}
	cache, _ := lru.New[string, []byte](1)
	cfg := config.DefaultServerConfig()
	cfg.StaticDir = tmpDir
	cfg.Domain = "example.com"
	svc := NewService(renderer, cache, cfg)

	// Read the file through readStaticFile
	result := svc.readStaticFile("robots.txt", "fallback")

	// The readStaticFile function doesn't do template replacement
	// That happens in the handler functions (handleRobotsTxt, handleSitemapXml)
	// So we expect the raw content with {{DOMAIN}} still present
	if !strings.Contains(result, "{{DOMAIN}}") {
		t.Errorf("expected result to contain {{DOMAIN}} placeholder, got: %q", result)
	}

	// Verify template replacement would work (simulating what the handler does)
	replaced := strings.ReplaceAll(result, "{{DOMAIN}}", cfg.Domain)
	expectedAfterReplacement := "User-agent: *\nSitemap: https://example.com/sitemap.xml"
	if replaced != expectedAfterReplacement {
		t.Errorf("expected %q after replacement but got %q", expectedAfterReplacement, replaced)
	}
}
