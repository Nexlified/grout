package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/andybalholm/brotli"
)

func TestCompressionMiddleware_Gzip(t *testing.T) {
	tests := []struct {
		name         string
		contentType  string
		acceptEnc    string
		shouldCompr  bool
		body         string
	}{
		{
			name:        "SVG with gzip",
			contentType: "image/svg+xml",
			acceptEnc:   "gzip",
			shouldCompr: true,
			body:        `<svg xmlns="http://www.w3.org/2000/svg"><text>Test</text></svg>`,
		},
		{
			name:        "PNG with gzip",
			contentType: "image/png",
			acceptEnc:   "gzip",
			shouldCompr: false,
			body:        "fake png data",
		},
		{
			name:        "JPEG with gzip",
			contentType: "image/jpeg",
			acceptEnc:   "gzip",
			shouldCompr: false,
			body:        "fake jpeg data",
		},
		{
			name:        "WebP with gzip",
			contentType: "image/webp",
			acceptEnc:   "gzip",
			shouldCompr: false,
			body:        "fake webp data",
		},
		{
			name:        "GIF with gzip",
			contentType: "image/gif",
			acceptEnc:   "gzip",
			shouldCompr: false,
			body:        "fake gif data",
		},
		{
			name:        "HTML with gzip",
			contentType: "text/html",
			acceptEnc:   "gzip",
			shouldCompr: true,
			body:        "<html><body>Test</body></html>",
		},
		{
			name:        "JSON with gzip",
			contentType: "application/json",
			acceptEnc:   "gzip",
			shouldCompr: true,
			body:        `{"test": "data"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := CompressionMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", tt.contentType)
				w.Write([]byte(tt.body))
			}))

			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Accept-Encoding", tt.acceptEnc)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if tt.shouldCompr {
				// Should be compressed
				if enc := rec.Header().Get("Content-Encoding"); enc != "gzip" {
					t.Errorf("expected Content-Encoding gzip, got %s", enc)
				}

				// Decompress and verify
				gr, err := gzip.NewReader(rec.Body)
				if err != nil {
					t.Fatalf("failed to create gzip reader: %v", err)
				}
				defer gr.Close()

				decompressed, err := io.ReadAll(gr)
				if err != nil {
					t.Fatalf("failed to decompress: %v", err)
				}

				if string(decompressed) != tt.body {
					t.Errorf("decompressed body mismatch: got %q, want %q", string(decompressed), tt.body)
				}
			} else {
				// Should not be compressed
				if enc := rec.Header().Get("Content-Encoding"); enc == "gzip" {
					t.Errorf("expected no Content-Encoding, got gzip")
				}

				if rec.Body.String() != tt.body {
					t.Errorf("body mismatch: got %q, want %q", rec.Body.String(), tt.body)
				}
			}
		})
	}
}

func TestCompressionMiddleware_Brotli(t *testing.T) {
	tests := []struct {
		name         string
		contentType  string
		acceptEnc    string
		shouldCompr  bool
		body         string
	}{
		{
			name:        "SVG with brotli",
			contentType: "image/svg+xml",
			acceptEnc:   "br",
			shouldCompr: true,
			body:        `<svg xmlns="http://www.w3.org/2000/svg"><text>Test</text></svg>`,
		},
		{
			name:        "PNG with brotli",
			contentType: "image/png",
			acceptEnc:   "br",
			shouldCompr: false,
			body:        "fake png data",
		},
		{
			name:        "HTML with brotli",
			contentType: "text/html",
			acceptEnc:   "br",
			shouldCompr: true,
			body:        "<html><body>Test</body></html>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := CompressionMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", tt.contentType)
				w.Write([]byte(tt.body))
			}))

			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Accept-Encoding", tt.acceptEnc)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if tt.shouldCompr {
				// Should be compressed
				if enc := rec.Header().Get("Content-Encoding"); enc != "br" {
					t.Errorf("expected Content-Encoding br, got %s", enc)
				}

				// Decompress and verify
				br := brotli.NewReader(rec.Body)
				decompressed, err := io.ReadAll(br)
				if err != nil {
					t.Fatalf("failed to decompress: %v", err)
				}

				if string(decompressed) != tt.body {
					t.Errorf("decompressed body mismatch: got %q, want %q", string(decompressed), tt.body)
				}
			} else {
				// Should not be compressed
				if enc := rec.Header().Get("Content-Encoding"); enc == "br" {
					t.Errorf("expected no Content-Encoding, got br")
				}

				if rec.Body.String() != tt.body {
					t.Errorf("body mismatch: got %q, want %q", rec.Body.String(), tt.body)
				}
			}
		})
	}
}

func TestCompressionMiddleware_BrotliPreferred(t *testing.T) {
	// When both gzip and brotli are supported, brotli should be preferred
	handler := CompressionMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/svg+xml")
		w.Write([]byte(`<svg>test</svg>`))
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip, br")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if enc := rec.Header().Get("Content-Encoding"); enc != "br" {
		t.Errorf("expected brotli to be preferred, got %s", enc)
	}
}

func TestCompressionMiddleware_NoAcceptEncoding(t *testing.T) {
	// When no Accept-Encoding is present, no compression should occur
	handler := CompressionMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/svg+xml")
		w.Write([]byte(`<svg>test</svg>`))
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if enc := rec.Header().Get("Content-Encoding"); enc != "" {
		t.Errorf("expected no Content-Encoding, got %s", enc)
	}

	if rec.Body.String() != `<svg>test</svg>` {
		t.Errorf("body mismatch")
	}
}

func TestCompressionMiddleware_LargeSVG(t *testing.T) {
	// Test with a larger SVG to ensure proper compression
	largeSVG := strings.Repeat(`<circle cx="50" cy="50" r="40" />`, 100)
	fullSVG := `<svg xmlns="http://www.w3.org/2000/svg">` + largeSVG + `</svg>`

	handler := CompressionMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/svg+xml")
		w.Write([]byte(fullSVG))
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Verify compression
	if enc := rec.Header().Get("Content-Encoding"); enc != "gzip" {
		t.Fatalf("expected gzip encoding, got %s", enc)
	}

	// Compressed size should be significantly smaller
	compressedSize := rec.Body.Len()
	if compressedSize >= len(fullSVG) {
		t.Errorf("compressed size (%d) should be less than original (%d)", compressedSize, len(fullSVG))
	}

	// Verify decompression
	gr, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatalf("failed to create gzip reader: %v", err)
	}
	defer gr.Close()

	decompressed, err := io.ReadAll(gr)
	if err != nil {
		t.Fatalf("failed to decompress: %v", err)
	}

	if string(decompressed) != fullSVG {
		t.Errorf("decompressed content doesn't match original")
	}

	t.Logf("Compression ratio: %.2f%% (original: %d bytes, compressed: %d bytes)",
		float64(compressedSize)/float64(len(fullSVG))*100, len(fullSVG), compressedSize)
}

func TestShouldCompress(t *testing.T) {
	tests := []struct {
		contentType string
		expected    bool
	}{
		{"image/svg+xml", true},
		{"text/html", true},
		{"text/plain", true},
		{"text/css", true},
		{"text/javascript", true},
		{"application/javascript", true},
		{"application/json", true},
		{"application/xml", true},
		{"image/png", false},
		{"image/jpeg", false},
		{"image/gif", false},
		{"image/webp", false},
		{"application/octet-stream", false},
	}

	for _, tt := range tests {
		t.Run(tt.contentType, func(t *testing.T) {
			result := shouldCompress(tt.contentType)
			if result != tt.expected {
				t.Errorf("shouldCompress(%q) = %v, want %v", tt.contentType, result, tt.expected)
			}
		})
	}
}

func BenchmarkCompressionMiddleware_SVG_Gzip(b *testing.B) {
	svgContent := `<svg xmlns="http://www.w3.org/2000/svg" width="256" height="256">
		<rect width="256" height="256" fill="#3498db"/>
		<text x="128" y="128" font-size="64" text-anchor="middle" fill="#ffffff">AB</text>
	</svg>`

	handler := CompressionMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/svg+xml")
		w.Write([]byte(svgContent))
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}

func BenchmarkCompressionMiddleware_SVG_Brotli(b *testing.B) {
	svgContent := `<svg xmlns="http://www.w3.org/2000/svg" width="256" height="256">
		<rect width="256" height="256" fill="#3498db"/>
		<text x="128" y="128" font-size="64" text-anchor="middle" fill="#ffffff">AB</text>
	</svg>`

	handler := CompressionMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/svg+xml")
		w.Write([]byte(svgContent))
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "br")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}

func BenchmarkCompressionMiddleware_PNG_NoCompression(b *testing.B) {
	pngContent := bytes.Repeat([]byte("fake png data"), 100)

	handler := CompressionMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Write(pngContent)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}
