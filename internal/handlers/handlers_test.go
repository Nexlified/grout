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
	svc.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/avatar/", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rec.Code)
	}
	// Default format is now WebP
	if ct := rec.Header().Get("Content-Type"); ct != "image/webp" {
		t.Fatalf("expected content-type image/webp got %s", ct)
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
	svc.RegisterRoutes(mux)

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
		{"No extension defaults to WebP", "/avatar/JohnDoe", "image/webp"},
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
	svc.RegisterRoutes(mux)

	tests := []struct {
		name        string
		path        string
		contentType string
	}{
		{"PNG format", "/placeholder/200x100.png", "image/png"},
		{"JPG format", "/placeholder/200x100.jpg", "image/jpeg"},
		{"GIF format", "/placeholder/200x100.gif", "image/gif"},
		{"WebP format", "/placeholder/200x100.webp", "image/webp"},
		{"No extension defaults to WebP", "/placeholder/200x100", "image/webp"},
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
	svc.RegisterRoutes(mux)

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
	svc.RegisterRoutes(mux)

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
	svc.RegisterRoutes(mux)

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
	svc.RegisterRoutes(mux)

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
	svc.RegisterRoutes(mux)

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
	svc.RegisterRoutes(mux)

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
	svc.RegisterRoutes(mux)

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
	svc.RegisterRoutes(mux)

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
svc.RegisterRoutes(mux)

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
svc.RegisterRoutes(mux)

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
