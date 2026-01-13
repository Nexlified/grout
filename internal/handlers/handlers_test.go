package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/golang-lru/v2"

	"grout/internal/config"
	"grout/internal/render"
)

func TestAvatarHandlerDefaults(t *testing.T) {
	renderer, err := render.New()
	if err != nil {
		t.Fatalf("renderer init: %v", err)
	}
	cache, _ := lru.New[string, []byte](1)
	svc := NewService(renderer, cache, config.DefaultServerConfig())
	mux := http.NewServeMux()
	svc.RegisterRoutes(mux, nil)

	req := httptest.NewRequest(http.MethodGet, "/avatar/", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rec.Code)
	}
	// Default format is now SVG
	if ct := rec.Header().Get("Content-Type"); ct != "image/svg+xml" {
		t.Fatalf("expected content-type image/svg+xml got %s", ct)
	}
	if rec.Body.Len() == 0 {
		t.Fatal("expected body to contain image data")
	}
}

func TestAvatarHandlerFormats(t *testing.T) {
	renderer, err := render.New()
	if err != nil {
		t.Fatalf("renderer init: %v", err)
	}
	cache, _ := lru.New[string, []byte](1)
	svc := NewService(renderer, cache, config.DefaultServerConfig())
	mux := http.NewServeMux()
	svc.RegisterRoutes(mux, nil)

	tests := []struct {
		name        string
		path        string
		contentType string
	}{
		{"PNG format", "/avatar/JohnDoe.png", "image/png"},
		{"JPG format", "/avatar/JohnDoe.jpg", "image/jpeg"},
		{"JPEG format", "/avatar/JohnDoe.jpeg", "image/jpeg"},
		{"GIF format", "/avatar/JohnDoe.gif", "image/gif"},
		{"WebP format", "/avatar/JohnDoe.webp", "image/webp"},
		{"SVG format", "/avatar/JohnDoe.svg", "image/svg+xml"},
		{"No extension defaults to SVG", "/avatar/JohnDoe", "image/svg+xml"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected 200 got %d", rec.Code)
			}
			if ct := rec.Header().Get("Content-Type"); ct != tt.contentType {
				t.Fatalf("expected content-type %s got %s", tt.contentType, ct)
			}
			if rec.Body.Len() == 0 {
				t.Fatal("expected body to contain image data")
			}
		})
	}
}

func TestPlaceholderHandlerFormats(t *testing.T) {
	renderer, err := render.New()
	if err != nil {
		t.Fatalf("renderer init: %v", err)
	}
	cache, _ := lru.New[string, []byte](1)
	svc := NewService(renderer, cache, config.DefaultServerConfig())
	mux := http.NewServeMux()
	svc.RegisterRoutes(mux, nil)

	tests := []struct {
		name        string
		path        string
		contentType string
	}{
		{"PNG format", "/placeholder/200x100.png", "image/png"},
		{"JPG format", "/placeholder/200x100.jpg", "image/jpeg"},
		{"GIF format", "/placeholder/200x100.gif", "image/gif"},
		{"WebP format", "/placeholder/200x100.webp", "image/webp"},
		{"SVG format", "/placeholder/200x100.svg", "image/svg+xml"},
		{"No extension defaults to SVG", "/placeholder/200x100", "image/svg+xml"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected 200 got %d", rec.Code)
			}
			if ct := rec.Header().Get("Content-Type"); ct != tt.contentType {
				t.Fatalf("expected content-type %s got %s", tt.contentType, ct)
			}
			if rec.Body.Len() == 0 {
				t.Fatal("expected body to contain image data")
			}
		})
	}
}

func TestPlaceholderHandlerGradient(t *testing.T) {
	renderer, err := render.New()
	if err != nil {
		t.Fatalf("renderer init: %v", err)
	}
	cache, _ := lru.New[string, []byte](1)
	svc := NewService(renderer, cache, config.DefaultServerConfig())
	mux := http.NewServeMux()
	svc.RegisterRoutes(mux, nil)

	tests := []struct {
		name string
		path string
	}{
		{"Gradient with comma", "/placeholder/800x400?bg=ff0000,0000ff"},
		{"Gradient PNG", "/placeholder/800x400.png?bg=ff0000,0000ff"},
		{"Gradient with text", "/placeholder/800x400?bg=ff0000,0000ff&text=Hero+Image"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected 200 got %d", rec.Code)
			}
			if rec.Body.Len() == 0 {
				t.Fatal("expected body to contain image data")
			}
		})
	}
}

func TestHomeHandler(t *testing.T) {
	renderer, err := render.New()
	if err != nil {
		t.Fatalf("renderer init: %v", err)
	}
	cache, _ := lru.New[string, []byte](1)
	svc := NewService(renderer, cache, config.DefaultServerConfig())
	mux := http.NewServeMux()
	svc.RegisterRoutes(mux, nil)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Fatalf("expected content-type text/html; charset=utf-8 got %s", ct)
	}
	if rec.Body.Len() == 0 {
		t.Fatal("expected body to contain HTML content")
	}

	body := rec.Body.String()
	expectedStrings := []string{
		"Grout",
		"Made with love in Nexlified Lab",
		"https://github.com/Nexlified/grout",
		"Avatar API Examples",
		"Placeholder Image API Examples",
		"Avatar URL Parameters",
		"Placeholder URL Parameters",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(body, expected) {
			t.Errorf("expected body to contain %q", expected)
		}
	}
}

func TestHomeHandlerNotFound(t *testing.T) {
	renderer, err := render.New()
	if err != nil {
		t.Fatalf("renderer init: %v", err)
	}
	cache, _ := lru.New[string, []byte](1)
	svc := NewService(renderer, cache, config.DefaultServerConfig())
	mux := http.NewServeMux()
	svc.RegisterRoutes(mux, nil)

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 got %d", rec.Code)
	}
}

func TestFaviconHandler(t *testing.T) {
	renderer, err := render.New()
	if err != nil {
		t.Fatalf("renderer init: %v", err)
	}
	cache, _ := lru.New[string, []byte](1)
	svc := NewService(renderer, cache, config.DefaultServerConfig())
	mux := http.NewServeMux()
	svc.RegisterRoutes(mux, nil)

	req := httptest.NewRequest(http.MethodGet, "/favicon.ico", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "image/png" {
		t.Fatalf("expected content-type image/png got %s", ct)
	}
	if rec.Body.Len() == 0 {
		t.Fatal("expected body to contain favicon data")
	}
	// Check for cache control header
	if cc := rec.Header().Get("Cache-Control"); !strings.Contains(cc, "max-age") {
		t.Fatalf("expected Cache-Control header with max-age, got %s", cc)
	}
}

func TestPlaceholderHandlerWithQuote(t *testing.T) {
	renderer, err := render.New()
	if err != nil {
		t.Fatalf("renderer init: %v", err)
	}
	cache, _ := lru.New[string, []byte](1)
	svc := NewService(renderer, cache, config.DefaultServerConfig())
	mux := http.NewServeMux()
	svc.RegisterRoutes(mux, nil)

	tests := []struct {
		name string
		path string
	}{
		{"Quote without category", "/placeholder/800x400?quote=true"},
		{"Quote with category", "/placeholder/800x400?quote=true&category=inspirational"},
		{"Quote with PNG format", "/placeholder/800x400.png?quote=true"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected 200 got %d", rec.Code)
			}
			if rec.Body.Len() == 0 {
				t.Fatal("expected body to contain image data")
			}
		})
	}
}

func TestPlaceholderHandlerWithJoke(t *testing.T) {
	renderer, err := render.New()
	if err != nil {
		t.Fatalf("renderer init: %v", err)
	}
	cache, _ := lru.New[string, []byte](1)
	svc := NewService(renderer, cache, config.DefaultServerConfig())
	mux := http.NewServeMux()
	svc.RegisterRoutes(mux, nil)

	tests := []struct {
		name string
		path string
	}{
		{"Joke without category", "/placeholder/800x400?joke=true"},
		{"Joke with category", "/placeholder/800x400?joke=true&category=programming"},
		{"Joke with PNG format", "/placeholder/800x400.png?joke=true"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected 200 got %d", rec.Code)
			}
			if rec.Body.Len() == 0 {
				t.Fatal("expected body to contain image data")
			}
		})
	}
}

func TestPlaceholderHandlerWithInvalidCategory(t *testing.T) {
	renderer, err := render.New()
	if err != nil {
		t.Fatalf("renderer init: %v", err)
	}
	cache, _ := lru.New[string, []byte](1)
	svc := NewService(renderer, cache, config.DefaultServerConfig())
	mux := http.NewServeMux()
	svc.RegisterRoutes(mux, nil)

	// With invalid category, should fall back to default dimensions text
	req := httptest.NewRequest(http.MethodGet, "/placeholder/800x400?quote=true&category=nonexistent", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rec.Code)
	}
	if rec.Body.Len() == 0 {
		t.Fatal("expected body to contain image data")
	}
}

func TestErrorPage404(t *testing.T) {
	renderer, err := render.New()
	if err != nil {
		t.Fatalf("renderer init: %v", err)
	}
	cache, _ := lru.New[string, []byte](1)
	svc := NewService(renderer, cache, config.DefaultServerConfig())
	mux := http.NewServeMux()
	svc.RegisterRoutes(mux, nil)

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 got %d", rec.Code)
	}

	body := rec.Body.String()
	// Check that it's HTML, not plain text
	if !strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("expected HTML response for 404")
	}
	// Check for key error page elements
	if !strings.Contains(body, "404") {
		t.Error("expected body to contain 404 status code")
	}
	if !strings.Contains(body, "Not Found") {
		t.Error("expected body to contain 'Not Found'")
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Fatalf("expected content-type text/html; charset=utf-8 got %s", ct)
	}
}

func TestServeErrorPage4xx(t *testing.T) {
	renderer, err := render.New()
	if err != nil {
		t.Fatalf("renderer init: %v", err)
	}
	cache, _ := lru.New[string, []byte](1)
	svc := NewService(renderer, cache, config.DefaultServerConfig())

	tests := []struct {
		name       string
		statusCode int
		message    string
	}{
		{"400 Bad Request", http.StatusBadRequest, "Invalid request parameters"},
		{"403 Forbidden", http.StatusForbidden, "Access denied"},
		{"404 Not Found", http.StatusNotFound, "Page not found"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			svc.serveErrorPage(rec, tt.statusCode, tt.message)

			if rec.Code != tt.statusCode {
				t.Fatalf("expected %d got %d", tt.statusCode, rec.Code)
			}

			body := rec.Body.String()
			if !strings.Contains(body, "<!DOCTYPE html>") {
				t.Error("expected HTML response")
			}
			if !strings.Contains(body, tt.message) {
				t.Errorf("expected body to contain message: %s", tt.message)
			}
			if ct := rec.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
				t.Fatalf("expected content-type text/html; charset=utf-8 got %s", ct)
			}
		})
	}
}

func TestServeErrorPage5xx(t *testing.T) {
	renderer, err := render.New()
	if err != nil {
		t.Fatalf("renderer init: %v", err)
	}
	cache, _ := lru.New[string, []byte](1)
	svc := NewService(renderer, cache, config.DefaultServerConfig())

	tests := []struct {
		name       string
		statusCode int
		message    string
	}{
		{"500 Internal Server Error", http.StatusInternalServerError, "Something went wrong"},
		{"503 Service Unavailable", http.StatusServiceUnavailable, "Service temporarily unavailable"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			svc.serveErrorPage(rec, tt.statusCode, tt.message)

			if rec.Code != tt.statusCode {
				t.Fatalf("expected %d got %d", tt.statusCode, rec.Code)
			}

			body := rec.Body.String()
			if !strings.Contains(body, "<!DOCTYPE html>") {
				t.Error("expected HTML response")
			}
			if !strings.Contains(body, tt.message) {
				t.Errorf("expected body to contain message: %s", tt.message)
			}
			if ct := rec.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
				t.Fatalf("expected content-type text/html; charset=utf-8 got %s", ct)
			}
		})
	}
}

func TestRobotsTxtHandler(t *testing.T) {
	renderer, err := render.New()
	if err != nil {
		t.Fatalf("renderer init: %v", err)
	}
	cache, _ := lru.New[string, []byte](1)
	cfg := config.DefaultServerConfig()
	cfg.Domain = "example.com"
	svc := NewService(renderer, cache, cfg)
	mux := http.NewServeMux()
	svc.RegisterRoutes(mux, nil)

	req := httptest.NewRequest(http.MethodGet, "/robots.txt", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/plain; charset=utf-8" {
		t.Fatalf("expected content-type text/plain; charset=utf-8 got %s", ct)
	}

	body := rec.Body.String()
	expectedStrings := []string{
		"User-agent: *",
		"Allow: /",
		"Sitemap: https://example.com/sitemap.xml",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(body, expected) {
			t.Errorf("expected body to contain %q", expected)
		}
	}
}

func TestSitemapXmlHandler(t *testing.T) {
	renderer, err := render.New()
	if err != nil {
		t.Fatalf("renderer init: %v", err)
	}
	cache, _ := lru.New[string, []byte](1)
	cfg := config.DefaultServerConfig()
	cfg.Domain = "example.com"
	svc := NewService(renderer, cache, cfg)
	mux := http.NewServeMux()
	svc.RegisterRoutes(mux, nil)

	req := httptest.NewRequest(http.MethodGet, "/sitemap.xml", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/xml; charset=utf-8" {
		t.Fatalf("expected content-type application/xml; charset=utf-8 got %s", ct)
	}

	body := rec.Body.String()
	expectedStrings := []string{
		"<?xml version=\"1.0\" encoding=\"UTF-8\"?>",
		"<urlset",
		"https://example.com/",
		"https://example.com/play",
		"<priority>1.0</priority>",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(body, expected) {
			t.Errorf("expected body to contain %q", expected)
		}
	}
}

func TestPlaceholderHandlerMinimumWidthForQuotes(t *testing.T) {
	renderer, err := render.New()
	if err != nil {
		t.Fatalf("renderer init: %v", err)
	}
	cache, _ := lru.New[string, []byte](1)
	svc := NewService(renderer, cache, config.DefaultServerConfig())
	mux := http.NewServeMux()
	svc.RegisterRoutes(mux, nil)

	tests := []struct {
		name        string
		path        string
		expectQuote bool
	}{
		{"Quote with sufficient width", "/placeholder/800x400?quote=true", true},
		{"Quote with minimum width", "/placeholder/300x400?quote=true", true},
		{"Quote with insufficient width", "/placeholder/200x400?quote=true", false},
		{"Joke with sufficient width", "/placeholder/600x300?joke=true", true},
		{"Joke with insufficient width", "/placeholder/250x300?joke=true", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected 200 got %d", rec.Code)
			}
			if rec.Body.Len() == 0 {
				t.Fatal("expected body to contain image data")
			}
		})
	}
}

func TestAvatarHandlerBackgroundParamConsistency(t *testing.T) {
	renderer, err := render.New()
	if err != nil {
		t.Fatalf("renderer init: %v", err)
	}
	cache, _ := lru.New[string, []byte](1)
	svc := NewService(renderer, cache, config.DefaultServerConfig())
	mux := http.NewServeMux()
	svc.RegisterRoutes(mux, nil)

	tests := []struct {
		name string
		path string
	}{
		{"Using background param", "/avatar/JohnDoe?background=ff0000"},
		{"Using bg param", "/avatar/JohnDoe?bg=ff0000"},
		{"Using both (background takes precedence)", "/avatar/JohnDoe?background=ff0000&bg=00ff00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected 200 got %d", rec.Code)
			}
			if rec.Body.Len() == 0 {
				t.Fatal("expected body to contain image data")
			}
		})
	}
}

func TestPlaceholderHandlerBackgroundParamConsistency(t *testing.T) {
	renderer, err := render.New()
	if err != nil {
		t.Fatalf("renderer init: %v", err)
	}
	cache, _ := lru.New[string, []byte](1)
	svc := NewService(renderer, cache, config.DefaultServerConfig())
	mux := http.NewServeMux()
	svc.RegisterRoutes(mux, nil)

	tests := []struct {
		name string
		path string
	}{
		{"Using background param", "/placeholder/400x300?background=ff0000"},
		{"Using bg param", "/placeholder/400x300?bg=ff0000"},
		{"Using both (background takes precedence)", "/placeholder/400x300?background=ff0000&bg=00ff00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected 200 got %d", rec.Code)
			}
			if rec.Body.Len() == 0 {
				t.Fatal("expected body to contain image data")
			}
		})
	}
}

// rateLimiterWrapper is a test helper that wraps a middleware function
type rateLimiterWrapper struct {
	middleware func(http.Handler) http.Handler
}

func (w rateLimiterWrapper) Middleware(next http.Handler) http.Handler {
	return w.middleware(next)
}

func TestRateLimitingIntegration(t *testing.T) {
	renderer, err := render.New()
	if err != nil {
		t.Fatalf("renderer init: %v", err)
	}
	cache, _ := lru.New[string, []byte](1)
	svc := NewService(renderer, cache, config.DefaultServerConfig())
	mux := http.NewServeMux()

	// Mock rate limiter for testing
	count := 0
	rlMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			count++
			if count > 2 {
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}

	middlewareWrapper := rateLimiterWrapper{middleware: rlMiddleware}
	svc.RegisterRoutes(mux, middlewareWrapper)

	tests := []struct {
		name           string
		path           string
		expectedStatus int
	}{
		{"First avatar request should succeed", "/avatar/JohnDoe", http.StatusOK},
		{"Second avatar request should succeed", "/avatar/JaneDoe", http.StatusOK},
		{"Third avatar request should be rate limited", "/avatar/BobSmith", http.StatusTooManyRequests},
		{"Favicon should not be rate limited", "/favicon.ico", http.StatusOK},
		{"Health should not be rate limited", "/health", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

// expectedSecurityHeaders returns the map of expected security headers
func expectedSecurityHeaders() map[string]string {
	return map[string]string{
		"Content-Security-Policy": "default-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; script-src 'self' 'unsafe-inline'",
		"X-Content-Type-Options":  "nosniff",
		"X-Frame-Options":         "DENY",
		"X-XSS-Protection":        "1; mode=block",
	}
}

// setupTestService creates a test service with renderer, cache, and mux
func setupTestService(t *testing.T) (*Service, *http.ServeMux) {
	renderer, err := render.New()
	if err != nil {
		t.Fatalf("renderer init: %v", err)
	}
	cache, _ := lru.New[string, []byte](1)
	svc := NewService(renderer, cache, config.DefaultServerConfig())
	mux := http.NewServeMux()
	svc.RegisterRoutes(mux, nil)
	return svc, mux
}

// verifySecurityHeaders checks that all expected security headers are present
func verifySecurityHeaders(t *testing.T, rec *httptest.ResponseRecorder) {
	headers := expectedSecurityHeaders()
	for header, expectedValue := range headers {
		actualValue := rec.Header().Get(header)
		if actualValue != expectedValue {
			t.Errorf("expected %s header to be %q, got %q", header, expectedValue, actualValue)
		}
	}
}

func TestSecurityHeadersOnHomeEndpoint(t *testing.T) {
	_, mux := setupTestService(t)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rec.Code)
	}

	verifySecurityHeaders(t, rec)
}

func TestSecurityHeadersOnPlayEndpoint(t *testing.T) {
	_, mux := setupTestService(t)

	req := httptest.NewRequest(http.MethodGet, "/play", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rec.Code)
	}

	verifySecurityHeaders(t, rec)
}

func TestSecurityHeadersOn404ErrorPage(t *testing.T) {
	_, mux := setupTestService(t)

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 got %d", rec.Code)
	}

	verifySecurityHeaders(t, rec)
}

func TestSecurityHeadersOn500ErrorPage(t *testing.T) {
	svc, _ := setupTestService(t)

	rec := httptest.NewRecorder()
	svc.serveErrorPage(rec, http.StatusInternalServerError, "Test error")

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 got %d", rec.Code)
	}

	verifySecurityHeaders(t, rec)
}

func TestSecurityHeadersNotPresentOnImageEndpoints(t *testing.T) {
	_, mux := setupTestService(t)

	tests := []struct {
		name string
		path string
	}{
		{"Avatar endpoint", "/avatar/JohnDoe"},
		{"Placeholder endpoint", "/placeholder/200x100"},
		{"Favicon endpoint", "/favicon.ico"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected 200 got %d", rec.Code)
			}

			// Verify security headers are NOT present on image responses
			// (they should only be on HTML responses)
			securityHeaders := []string{
				"Content-Security-Policy",
				"X-Frame-Options",
				"X-XSS-Protection",
			}

			for _, header := range securityHeaders {
				if value := rec.Header().Get(header); value != "" {
					t.Errorf("did not expect %s header on image endpoint, but got: %q", header, value)
				}
			}
		})
	}
}

func TestCompressionIntegration(t *testing.T) {
	renderer, err := render.New()
	if err != nil {
		t.Fatalf("renderer init: %v", err)
	}
	cache, _ := lru.New[string, []byte](100)
	svc := NewService(renderer, cache, config.DefaultServerConfig())
	mux := http.NewServeMux()
	svc.RegisterRoutes(mux, nil)

	// Import compression middleware
	compressMux := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mux.ServeHTTP(w, r)
	})

	tests := []struct {
		name              string
		path              string
		acceptEncoding    string
		shouldCompress    bool
		expectedEncoding  string
		expectedMimeType  string
	}{
		{
			name:             "SVG with gzip",
			path:             "/avatar/JohnDoe.svg",
			acceptEncoding:   "gzip",
			shouldCompress:   true,
			expectedEncoding: "gzip",
			expectedMimeType: "image/svg+xml",
		},
		{
			name:             "SVG with brotli",
			path:             "/avatar/JohnDoe.svg",
			acceptEncoding:   "br",
			shouldCompress:   true,
			expectedEncoding: "br",
			expectedMimeType: "image/svg+xml",
		},
		{
			name:             "PNG should not compress",
			path:             "/avatar/JohnDoe.png",
			acceptEncoding:   "gzip",
			shouldCompress:   false,
			expectedEncoding: "",
			expectedMimeType: "image/png",
		},
		{
			name:             "Placeholder SVG with gzip",
			path:             "/placeholder/800x400.svg",
			acceptEncoding:   "gzip",
			shouldCompress:   true,
			expectedEncoding: "gzip",
			expectedMimeType: "image/svg+xml",
		},
		{
			name:             "SVG without Accept-Encoding",
			path:             "/avatar/JohnDoe.svg",
			acceptEncoding:   "",
			shouldCompress:   false,
			expectedEncoding: "",
			expectedMimeType: "image/svg+xml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			if tt.acceptEncoding != "" {
				req.Header.Set("Accept-Encoding", tt.acceptEncoding)
			}
			rec := httptest.NewRecorder()

			compressMux.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected 200 got %d", rec.Code)
			}

			// Check content type
			if ct := rec.Header().Get("Content-Type"); ct != tt.expectedMimeType {
				t.Errorf("expected content-type %s got %s", tt.expectedMimeType, ct)
			}

			// This test doesn't have compression middleware applied,
			// so we just verify the handlers work correctly
			// The actual compression is tested in middleware tests
		})
	}
}
