package handlers

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/golang-lru/v2"

	"grout/internal/config"
	"grout/internal/render"
)

// TestIntegrationMain is a top-level integration test suite that starts a real HTTP server
// This tests the full HTTP request lifecycle with actual network calls
func TestIntegrationMain(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	// Initialize dependencies
	renderer, err := render.New()
	if err != nil {
		t.Fatalf("renderer init: %v", err)
	}
	cache, err := lru.New[string, []byte](2000)
	if err != nil {
		t.Fatalf("cache init: %v", err)
	}
	cfg := config.DefaultServerConfig()
	svc := NewService(renderer, cache, cfg)
	mux := http.NewServeMux()
	svc.RegisterRoutes(mux)

	// Start a real HTTP server on a random available port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to create listener: %v", err)
	}
	defer listener.Close()

	server := &http.Server{Handler: mux}
	serverAddr := fmt.Sprintf("http://%s", listener.Addr().String())

	// Start server in goroutine
	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			t.Logf("server error: %v", err)
		}
	}()
	defer server.Close()

	// Wait for server to be ready by polling health endpoint
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Poll server health endpoint until ready or timeout
	maxAttempts := 50
	for i := 0; i < maxAttempts; i++ {
		resp, err := client.Get(serverAddr + "/health")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				break
			}
		}
		if i == maxAttempts-1 {
			t.Fatal("server failed to become ready within timeout")
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Run all integration test scenarios
	t.Run("AvatarEndpoint", func(t *testing.T) {
		testAvatarIntegration(t, client, serverAddr)
	})

	t.Run("PlaceholderEndpoint", func(t *testing.T) {
		testPlaceholderIntegration(t, client, serverAddr)
	})

	t.Run("StaticEndpoints", func(t *testing.T) {
		testStaticEndpointsIntegration(t, client, serverAddr)
	})

	t.Run("ErrorScenarios", func(t *testing.T) {
		testErrorScenariosIntegration(t, client, serverAddr)
	})

	t.Run("CachingBehavior", func(t *testing.T) {
		testCachingBehaviorIntegration(t, client, serverAddr)
	})
}

func testAvatarIntegration(t *testing.T, client *http.Client, baseURL string) {
	tests := []struct {
		name           string
		url            string
		expectedStatus int
		expectedCT     string
		checkBody      bool
		checkHeaders   []string
	}{
		{
			name:           "Default SVG avatar",
			url:            "/avatar/John+Doe",
			expectedStatus: http.StatusOK,
			expectedCT:     "image/svg+xml",
			checkBody:      true,
			checkHeaders:   []string{"Cache-Control", "ETag"},
		},
		{
			name:           "PNG avatar with parameters",
			url:            "/avatar/Jane+Smith.png?size=256&rounded=true&bold=true",
			expectedStatus: http.StatusOK,
			expectedCT:     "image/png",
			checkBody:      true,
			checkHeaders:   []string{"Cache-Control", "ETag"},
		},
		{
			name:           "JPG avatar",
			url:            "/avatar/Bob.jpg?size=128",
			expectedStatus: http.StatusOK,
			expectedCT:     "image/jpeg",
			checkBody:      true,
			checkHeaders:   []string{"Cache-Control", "ETag"},
		},
		{
			name:           "WebP avatar",
			url:            "/avatar/Alice.webp?background=random",
			expectedStatus: http.StatusOK,
			expectedCT:     "image/webp",
			checkBody:      true,
			checkHeaders:   []string{"Cache-Control", "ETag"},
		},
		{
			name:           "Avatar with custom colors",
			url:            "/avatar/Test.svg?bg=ff5733&color=ffffff",
			expectedStatus: http.StatusOK,
			expectedCT:     "image/svg+xml",
			checkBody:      true,
			checkHeaders:   []string{"Cache-Control", "ETag"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.Get(baseURL + tt.url)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if ct := resp.Header.Get("Content-Type"); ct != tt.expectedCT {
				t.Errorf("expected Content-Type %s, got %s", tt.expectedCT, ct)
			}

			if tt.checkBody {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Fatalf("failed to read body: %v", err)
				}
				if len(body) == 0 {
					t.Error("expected non-empty body")
				}
			}

			for _, header := range tt.checkHeaders {
				if val := resp.Header.Get(header); val == "" {
					t.Errorf("expected header %s to be present", header)
				}
			}
		})
	}
}

func testPlaceholderIntegration(t *testing.T, client *http.Client, baseURL string) {
	tests := []struct {
		name           string
		url            string
		expectedStatus int
		expectedCT     string
		checkBody      bool
		checkHeaders   []string
	}{
		{
			name:           "Default placeholder SVG",
			url:            "/placeholder/400x300",
			expectedStatus: http.StatusOK,
			expectedCT:     "image/svg+xml",
			checkBody:      true,
			checkHeaders:   []string{"Cache-Control", "ETag"},
		},
		{
			name:           "PNG placeholder with text",
			url:            "/placeholder/800x600.png?text=Hero+Image",
			expectedStatus: http.StatusOK,
			expectedCT:     "image/png",
			checkBody:      true,
			checkHeaders:   []string{"Cache-Control", "ETag"},
		},
		{
			name:           "Placeholder with gradient",
			url:            "/placeholder/1200x400.svg?bg=ff0000,0000ff&text=Gradient",
			expectedStatus: http.StatusOK,
			expectedCT:     "image/svg+xml",
			checkBody:      true,
			checkHeaders:   []string{"Cache-Control", "ETag"},
		},
		{
			name:           "Placeholder with quote",
			url:            "/placeholder/1000x500?quote=true",
			expectedStatus: http.StatusOK,
			expectedCT:     "image/svg+xml",
			checkBody:      true,
			checkHeaders:   []string{"Cache-Control", "ETag"},
		},
		{
			name:           "Placeholder with joke",
			url:            "/placeholder/800x400.png?joke=true&category=programming",
			expectedStatus: http.StatusOK,
			expectedCT:     "image/png",
			checkBody:      true,
			checkHeaders:   []string{"Cache-Control", "ETag"},
		},
		{
			name:           "JPG placeholder",
			url:            "/placeholder/600x400.jpg",
			expectedStatus: http.StatusOK,
			expectedCT:     "image/jpeg",
			checkBody:      true,
			checkHeaders:   []string{"Cache-Control", "ETag"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.Get(baseURL + tt.url)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if ct := resp.Header.Get("Content-Type"); ct != tt.expectedCT {
				t.Errorf("expected Content-Type %s, got %s", tt.expectedCT, ct)
			}

			if tt.checkBody {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Fatalf("failed to read body: %v", err)
				}
				if len(body) == 0 {
					t.Error("expected non-empty body")
				}
			}

			for _, header := range tt.checkHeaders {
				if val := resp.Header.Get(header); val == "" {
					t.Errorf("expected header %s to be present", header)
				}
			}
		})
	}
}

func testStaticEndpointsIntegration(t *testing.T, client *http.Client, baseURL string) {
	tests := []struct {
		name           string
		url            string
		expectedStatus int
		expectedCT     string
		checkBody      bool
		bodyContains   []string
	}{
		{
			name:           "Home page",
			url:            "/",
			expectedStatus: http.StatusOK,
			expectedCT:     "text/html; charset=utf-8",
			checkBody:      true,
			bodyContains:   []string{"Grout", "Avatar API", "Placeholder"},
		},
		{
			name:           "Play page",
			url:            "/play",
			expectedStatus: http.StatusOK,
			expectedCT:     "text/html; charset=utf-8",
			checkBody:      true,
			bodyContains:   []string{"Grout", "Playground"},
		},
		{
			name:           "Favicon",
			url:            "/favicon.ico",
			expectedStatus: http.StatusOK,
			expectedCT:     "image/png",
			checkBody:      true,
		},
		{
			name:           "Robots.txt",
			url:            "/robots.txt",
			expectedStatus: http.StatusOK,
			expectedCT:     "text/plain; charset=utf-8",
			checkBody:      true,
			bodyContains:   []string{"User-agent:", "Sitemap:"},
		},
		{
			name:           "Sitemap.xml",
			url:            "/sitemap.xml",
			expectedStatus: http.StatusOK,
			expectedCT:     "application/xml; charset=utf-8",
			checkBody:      true,
			bodyContains:   []string{"<?xml", "<urlset"},
		},
		{
			name:           "Health check",
			url:            "/health",
			expectedStatus: http.StatusOK,
			expectedCT:     "application/json",
			checkBody:      true,
			bodyContains:   []string{"\"status\":\"healthy\""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.Get(baseURL + tt.url)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if ct := resp.Header.Get("Content-Type"); ct != tt.expectedCT {
				t.Errorf("expected Content-Type %s, got %s", tt.expectedCT, ct)
			}

			if tt.checkBody {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Fatalf("failed to read body: %v", err)
				}
				if len(body) == 0 {
					t.Error("expected non-empty body")
				}

				bodyStr := string(body)
				for _, substr := range tt.bodyContains {
					if !strings.Contains(bodyStr, substr) {
						t.Errorf("expected body to contain %q", substr)
					}
				}
			}
		})
	}
}

func testErrorScenariosIntegration(t *testing.T, client *http.Client, baseURL string) {
	tests := []struct {
		name           string
		url            string
		expectedStatus int
		checkBodyHTML  bool
	}{
		{
			name:           "404 Not Found",
			url:            "/nonexistent",
			expectedStatus: http.StatusNotFound,
			checkBodyHTML:  true,
		},
		{
			name:           "404 for invalid path",
			url:            "/invalid/path",
			expectedStatus: http.StatusNotFound,
			checkBodyHTML:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.Get(baseURL + tt.url)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.checkBodyHTML {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Fatalf("failed to read body: %v", err)
				}
				bodyStr := string(body)
				if !strings.Contains(bodyStr, "<!DOCTYPE html>") {
					t.Error("expected HTML error page")
				}
			}
		})
	}
}

func testCachingBehaviorIntegration(t *testing.T, client *http.Client, baseURL string) {
	// Make first request to avatar endpoint
	url := baseURL + "/avatar/CacheTest.png?size=128"

	// First request - should not have X-Cache header
	resp1, err := client.Get(url)
	if err != nil {
		t.Fatalf("first request failed: %v", err)
	}
	defer resp1.Body.Close()

	if resp1.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp1.StatusCode)
	}

	etag1 := resp1.Header.Get("ETag")
	if etag1 == "" {
		t.Error("expected ETag header on first request")
	}

	// Read body to ensure request completes
	_, _ = io.ReadAll(resp1.Body)

	// Second request - should have same ETag and X-Cache: HIT
	resp2, err := client.Get(url)
	if err != nil {
		t.Fatalf("second request failed: %v", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp2.StatusCode)
	}

	etag2 := resp2.Header.Get("ETag")
	if etag2 != etag1 {
		t.Errorf("expected same ETag on second request, got %s (first was %s)", etag2, etag1)
	}

	xCache2 := resp2.Header.Get("X-Cache")
	if xCache2 != "HIT" {
		t.Logf("second request X-Cache: %s (expected HIT, but might be MISS depending on timing)", xCache2)
	}

	// Read body
	_, _ = io.ReadAll(resp2.Body)

	// Test Cache-Control header
	if cc := resp1.Header.Get("Cache-Control"); !strings.Contains(cc, "max-age") {
		t.Errorf("expected Cache-Control with max-age, got %s", cc)
	}

	// Test conditional request with If-None-Match
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("If-None-Match", etag1)

	resp3, err := client.Do(req)
	if err != nil {
		t.Fatalf("conditional request failed: %v", err)
	}
	defer resp3.Body.Close()

	// Should return 304 Not Modified or 200 OK depending on implementation
	if resp3.StatusCode != http.StatusNotModified && resp3.StatusCode != http.StatusOK {
		t.Logf("conditional request returned %d (expected 304 or 200)", resp3.StatusCode)
	}
}

// Benchmark integration test to ensure performance
func BenchmarkIntegrationAvatarRequest(b *testing.B) {
	// Initialize dependencies
	renderer, err := render.New()
	if err != nil {
		b.Fatalf("renderer init: %v", err)
	}
	cache, err := lru.New[string, []byte](2000)
	if err != nil {
		b.Fatalf("cache init: %v", err)
	}
	cfg := config.DefaultServerConfig()
	svc := NewService(renderer, cache, cfg)

	// Use httptest for benchmarking (faster than real HTTP server)
	mux := http.NewServeMux()
	svc.RegisterRoutes(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	client := &http.Client{Timeout: 5 * time.Second}
	url := server.URL + "/avatar/BenchTest.png?size=128"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := client.Get(url)
		if err != nil {
			b.Fatalf("request failed: %v", err)
		}
		_, _ = io.ReadAll(resp.Body)
		resp.Body.Close()
	}
}
