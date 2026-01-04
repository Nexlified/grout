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
		"Avatar Examples",
		"Placeholder Examples",
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
