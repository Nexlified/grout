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
	"grout/internal/render"
	"grout/internal/utils"
)

//go:embed web/index.html
var homePageHTML []byte

//go:embed web/favicon.png
var faviconData []byte

// Service bundles dependencies required by HTTP handlers.
type Service struct {
	renderer *render.Renderer
	cache    *lru.Cache[string, []byte]
	cfg      config.ServerConfig
}

// NewService wires the handler dependencies.
func NewService(renderer *render.Renderer, cache *lru.Cache[string, []byte], cfg config.ServerConfig) *Service {
	return &Service{renderer: renderer, cache: cache, cfg: cfg}
}

// RegisterRoutes attaches handlers to the provided mux.
func (s *Service) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", s.handleHome)
	mux.HandleFunc("/avatar/", s.handleAvatar)
	mux.HandleFunc("/placeholder/", s.handlePlaceholder)
	mux.HandleFunc("GET /health", s.HandleHealth)
	mux.HandleFunc("GET /favicon.ico", s.handleFavicon)
}

var placeholderRegex = regexp.MustCompile(`^(\d+)x(\d+)$`)

// formatExtensions maps file extensions to image formats
var formatExtensions = map[string]render.ImageFormat{
	".png":  render.FormatPNG,
	".jpg":  render.FormatJPG,
	".jpeg": render.FormatJPEG,
	".gif":  render.FormatGIF,
	".webp": render.FormatWebP,
}

// extractFormat extracts the image format from a filename, returning the format and the name without extension
func extractFormat(filename string) (render.ImageFormat, string) {
	// Check for known extensions
	for ext, format := range formatExtensions {
		if strings.HasSuffix(filename, ext) {
			return format, strings.TrimSuffix(filename, ext)
		}
	}

	// Default to WebP if no extension found
	return render.FormatWebP, filename
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
	default:
		return "image/webp"
	}
}

func (s *Service) handleAvatar(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	format := render.FormatWebP // Default to WebP

	if strings.HasPrefix(r.URL.Path, "/avatar/") {
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) > 2 && parts[2] != "" {
			format, name = extractFormat(parts[2])
		}
	}
	if name == "" {
		name = "John Doe"
	}

	size := utils.ParseIntOrDefault(r.URL.Query().Get("size"), config.DefaultSize)
	rounded := r.URL.Query().Get("rounded") == "true"
	bold := r.URL.Query().Get("bold") == "true"

	bgHex := r.URL.Query().Get("background")
	if bgHex == "" {
		bgHex = config.DefaultAvatarBg
	}
	if strings.EqualFold(bgHex, "random") {
		bgHex = render.GenerateColorHash(name)
	}

	fgHex := r.URL.Query().Get("color")
	if fgHex == "" {
		fgHex = render.GetContrastColor(bgHex)
	}

	key := fmt.Sprintf("Avatar:%s:%d:%t:%t:%s:%s:%s", name, size, rounded, bold, bgHex, fgHex, format)
	s.serveImage(w, r, key, format, func() ([]byte, error) {
		return s.renderer.DrawImageWithFormat(size, size, bgHex, fgHex, render.GetInitials(name), rounded, bold, format)
	})
}

func (s *Service) handlePlaceholder(w http.ResponseWriter, r *http.Request) {
	width, height := config.DefaultSize, config.DefaultSize
	pathMetric := strings.TrimPrefix(r.URL.Path, "/placeholder/")

	// Extract format from path
	format, pathMetric := extractFormat(pathMetric)

	if matches := placeholderRegex.FindStringSubmatch(pathMetric); len(matches) == 3 {
		width = utils.ParseIntOrDefault(matches[1], config.DefaultSize)
		height = utils.ParseIntOrDefault(matches[2], config.DefaultSize)
	} else {
		width = utils.ParseIntOrDefault(r.URL.Query().Get("w"), config.DefaultSize)
		height = utils.ParseIntOrDefault(r.URL.Query().Get("h"), config.DefaultSize)
	}

	text := r.URL.Query().Get("text")
	if text == "" {
		text = fmt.Sprintf("%d x %d", width, height)
	}

	bgHex := r.URL.Query().Get("bg")
	if bgHex == "" {
		bgHex = config.DefaultBgColor
	}
	fgHex := r.URL.Query().Get("color")
	if fgHex == "" {
		fgHex = render.GetContrastColor(bgHex)
	}

	key := fmt.Sprintf("PH:%d:%d:%s:%s:%s:%s", width, height, bgHex, fgHex, text, format)
	s.serveImage(w, r, key, format, func() ([]byte, error) {
		return s.renderer.DrawImageWithFormat(width, height, bgHex, fgHex, text, false, true, format)
	})
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
		http.Error(w, "Failed to generate image", http.StatusInternalServerError)
		return
	}

	s.cache.Add(cacheKey, imgData)
	w.Header().Set("X-Cache", "MISS")
	_, _ = w.Write(imgData)
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

func (s *Service) handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write(homePageHTML)
	if err != nil {
		return
	}
}

func (s *Service) handleFavicon(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write(faviconData)
	if err != nil {
		return
	}
}
