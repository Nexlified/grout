package handlers

import (
	"crypto/md5"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/hashicorp/golang-lru/v2"

	"grout/internal/config"
	"grout/internal/content"
	"grout/internal/render"
)

//go:embed web/error4xx.html
var error4xxTemplate string

//go:embed web/error5xx.html
var error5xxTemplate string

// Service bundles dependencies required by HTTP handlers.
type Service struct {
	renderer       *render.Renderer
	cache          *lru.Cache[string, []byte]
	cfg            config.ServerConfig
	contentManager *content.Manager
}

// NewService wires the handler dependencies.
func NewService(renderer *render.Renderer, cache *lru.Cache[string, []byte], cfg config.ServerConfig) *Service {
	contentManager, err := content.NewManager()
	if err != nil {
		// Content manager is optional - quotes/jokes will be unavailable but service will still work
		contentManager = nil
	}
	return &Service{renderer: renderer, cache: cache, cfg: cfg, contentManager: contentManager}
}

// RegisterRoutes attaches handlers to the provided mux.
func (s *Service) RegisterRoutes(mux *http.ServeMux, rateLimiter interface{}) {
	// Type-safe way to handle optional rate limiter
	var applyRateLimit func(http.Handler) http.Handler

	// Check if rate limiter is provided and has the right type
	if rl, ok := rateLimiter.(interface {
		Middleware(http.Handler) http.Handler
	}); ok {
		applyRateLimit = rl.Middleware
	} else {
		// No rate limiting - pass through
		applyRateLimit = func(h http.Handler) http.Handler { return h }
	}

	mux.HandleFunc("/", s.handleHome)
	mux.HandleFunc("/play", s.handlePlay)
	// Apply rate limiting to image generation endpoints
	mux.Handle("/avatar/", applyRateLimit(http.HandlerFunc(s.handleAvatar)))
	mux.Handle("/placeholder/", applyRateLimit(http.HandlerFunc(s.handlePlaceholder)))
	// No rate limiting for health, favicon, robots.txt, sitemap.xml
	mux.HandleFunc("GET /health", s.HandleHealth)
	mux.HandleFunc("GET /favicon.ico", s.handleFavicon)
	mux.HandleFunc("GET /robots.txt", s.handleRobotsTxt)
	mux.HandleFunc("GET /sitemap.xml", s.handleSitemapXml)
}

var placeholderRegex = regexp.MustCompile(`^(\d+)x(\d+)$`)

// formatExtensions maps file extensions to image formats
var formatExtensions = map[string]render.ImageFormat{
	".png":  render.FormatPNG,
	".jpg":  render.FormatJPG,
	".jpeg": render.FormatJPEG,
	".gif":  render.FormatGIF,
	".webp": render.FormatWebP,
	".svg":  render.FormatSVG,
}

// extractFormat extracts the image format from a filename, returning the format and the name without extension
func extractFormat(filename string) (render.ImageFormat, string) {
	// Check for known extensions
	for ext, format := range formatExtensions {
		if strings.HasSuffix(filename, ext) {
			return format, strings.TrimSuffix(filename, ext)
		}
	}

	// Default to SVG if no extension found
	return render.FormatSVG, filename
}

// getContentType returns the MIME type for the given format
func getContentType(format render.ImageFormat) string {
	switch format {
	case render.FormatPNG:
		return "image/png"
	case render.FormatJPG, render.FormatJPEG:
		return "image/jpeg"
	case render.FormatGIF:
		return "image/gif"
	case render.FormatWebP:
		return "image/webp"
	case render.FormatSVG:
		return "image/svg+xml"
	default:
		return "image/svg+xml"
	}
}

func (s *Service) serveImage(w http.ResponseWriter, r *http.Request, cacheKey string, format render.ImageFormat, generator func() ([]byte, error)) {
	etag := fmt.Sprintf("\"%x\"", md5.Sum([]byte(cacheKey)))

	w.Header().Set("Content-Type", getContentType(format))
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.Header().Set("ETag", etag)

	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	if imgData, ok := s.cache.Get(cacheKey); ok {
		w.Header().Set("X-Cache", "HIT")
		_, _ = w.Write(imgData)
		return
	}

	imgData, err := generator()
	if err != nil {
		// Clear headers set earlier since we're serving HTML now
		w.Header().Del("Content-Type")
		w.Header().Del("Cache-Control")
		w.Header().Del("ETag")
		s.serveErrorPage(w, http.StatusInternalServerError, "Failed to generate image. Please try again later or contact support if the problem persists.")
		return
	}

	s.cache.Add(cacheKey, imgData)
	w.Header().Set("X-Cache", "MISS")
	_, _ = w.Write(imgData)
}

// setSecurityHeaders applies security headers to HTML responses
func setSecurityHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; script-src 'self' 'unsafe-inline'")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
}

func (s *Service) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"version": "1.0.0",
	})
	if err != nil {
		return
	}
}

// serveErrorPage renders an error page with the given status code and message
func (s *Service) serveErrorPage(w http.ResponseWriter, statusCode int, message string) {
	var template string
	var statusText string

	// Determine which template to use based on status code
	if statusCode >= 400 && statusCode < 500 {
		template = error4xxTemplate
	} else {
		template = error5xxTemplate
	}

	// Get standard status text
	statusText = http.StatusText(statusCode)
	if statusText == "" {
		statusText = "Error"
	}

	// Replace placeholders
	html := strings.ReplaceAll(template, "{{STATUS_CODE}}", fmt.Sprintf("%d", statusCode))
	html = strings.ReplaceAll(html, "{{STATUS_TEXT}}", statusText)
	html = strings.ReplaceAll(html, "{{ERROR_MESSAGE}}", message)

	setSecurityHeaders(w)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(statusCode)
	_, err := w.Write([]byte(html))
	if err != nil {
		return
	}
}

// handle404 handles all 404 Not Found errors with a custom error page
func (s *Service) handle404(w http.ResponseWriter, r *http.Request) {
	message := "The page you're looking for doesn't exist. It might have been moved or deleted."
	s.serveErrorPage(w, http.StatusNotFound, message)
}
