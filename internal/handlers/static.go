package handlers

import (
	_ "embed"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

//go:embed web/index.html
var homePageTemplate string

//go:embed web/play.html
var playPageTemplate string

//go:embed web/favicon.png
var faviconData []byte

//go:embed web/robots.txt
var fallbackRobotsTxt string

//go:embed web/sitemap.xml
var fallbackSitemapXml string

func (s *Service) handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		s.handle404(w, r)
		return
	}

	// Replace {{DOMAIN}} placeholder with actual configured domain
	html := strings.ReplaceAll(homePageTemplate, "{{DOMAIN}}", s.cfg.Domain)

	setSecurityHeaders(w)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(html))
	if err != nil {
		return
	}
}

func (s *Service) handlePlay(w http.ResponseWriter, r *http.Request) {
	// Replace {{DOMAIN}} placeholder with actual configured domain
	html := strings.ReplaceAll(playPageTemplate, "{{DOMAIN}}", s.cfg.Domain)

	setSecurityHeaders(w)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(html))
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

func (s *Service) handleRobotsTxt(w http.ResponseWriter, r *http.Request) {
	// Try to read from static directory first
	content := s.readStaticFile("robots.txt", fallbackRobotsTxt)

	// Replace {{DOMAIN}} placeholder with actual configured domain
	content = strings.ReplaceAll(content, "{{DOMAIN}}", s.cfg.Domain)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(content))
	if err != nil {
		return
	}
}

func (s *Service) handleSitemapXml(w http.ResponseWriter, r *http.Request) {
	// Try to read from static directory first
	content := s.readStaticFile("sitemap.xml", fallbackSitemapXml)

	// Replace {{DOMAIN}} placeholder with actual configured domain
	content = strings.ReplaceAll(content, "{{DOMAIN}}", s.cfg.Domain)

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(content))
	if err != nil {
		return
	}
}

// readStaticFile attempts to read a file from the static directory.
// If the file doesn't exist or can't be read, it returns the fallback content.
// The function validates that the resolved path is within the static directory to prevent directory traversal attacks.
func (s *Service) readStaticFile(filename string, fallback string) string {
	// Clean the filename to prevent directory traversal
	cleanFilename := filepath.Clean(filename)

	// Prevent directory traversal by rejecting paths that start with ".." or are absolute
	if strings.HasPrefix(cleanFilename, "..") || filepath.IsAbs(cleanFilename) {
		return fallback
	}

	// Construct the full path
	filePath := filepath.Join(s.cfg.StaticDir, cleanFilename)

	// Resolve absolute paths and verify the file is within the static directory
	absStaticDir, err := filepath.Abs(s.cfg.StaticDir)
	if err != nil {
		return fallback
	}

	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		return fallback
	}

	// Ensure the static directory ends with a path separator for proper prefix checking
	if !strings.HasSuffix(absStaticDir, string(filepath.Separator)) {
		absStaticDir += string(filepath.Separator)
	}

	// Ensure the resolved path is within the static directory (must be a file, not the directory itself)
	if !strings.HasPrefix(absFilePath, absStaticDir) {
		return fallback
	}

	data, err := os.ReadFile(absFilePath)
	if err != nil {
		// File doesn't exist or can't be read, use fallback
		return fallback
	}

	return string(data)
}
