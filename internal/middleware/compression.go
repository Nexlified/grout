package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/andybalholm/brotli"
)

// gzipWriterPool pools gzip writers for reuse
var gzipWriterPool = sync.Pool{
	New: func() interface{} {
		w, _ := gzip.NewWriterLevel(nil, gzip.BestSpeed)
		return w
	},
}

// brotliWriterPool pools brotli writers for reuse
var brotliWriterPool = sync.Pool{
	New: func() interface{} {
		return brotli.NewWriterLevel(nil, brotli.DefaultCompression)
	},
}

// shouldCompress determines if the content type should be compressed
func shouldCompress(contentType string) bool {
	// Only compress SVG and text-based content
	// Skip already-compressed formats: PNG, JPEG, GIF, WebP
	compressible := []string{
		"image/svg+xml",
		"text/html",
		"text/plain",
		"text/css",
		"text/javascript",
		"application/javascript",
		"application/json",
		"application/xml",
	}

	for _, ct := range compressible {
		if strings.Contains(contentType, ct) {
			return true
		}
	}
	return false
}

// compressionResponseWriter wraps http.ResponseWriter to compress the response
type compressionResponseWriter struct {
	http.ResponseWriter
	writer          io.WriteCloser
	encoding        string
	headerWritten   bool
	compressionUsed bool
}

func (w *compressionResponseWriter) WriteHeader(statusCode int) {
	if w.headerWritten {
		return
	}
	w.headerWritten = true
	
	// Check if we should compress based on content type
	contentType := w.Header().Get("Content-Type")
	if shouldCompress(contentType) {
		w.compressionUsed = true
		w.Header().Set("Content-Encoding", w.encoding)
		w.Header().Del("Content-Length") // Remove content-length as it will change
		
		// Create the appropriate compressor
		if w.encoding == "br" {
			bw := brotliWriterPool.Get().(*brotli.Writer)
			bw.Reset(w.ResponseWriter)
			w.writer = &brotliWriterWrapper{Writer: bw}
		} else if w.encoding == "gzip" {
			gw := gzipWriterPool.Get().(*gzip.Writer)
			gw.Reset(w.ResponseWriter)
			w.writer = &gzipWriterWrapper{Writer: gw}
		}
	}
	
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *compressionResponseWriter) Write(b []byte) (int, error) {
	if !w.headerWritten {
		w.WriteHeader(http.StatusOK)
	}
	
	if w.compressionUsed && w.writer != nil {
		return w.writer.Write(b)
	}
	
	return w.ResponseWriter.Write(b)
}

// Close flushes and closes the compressor if used
func (w *compressionResponseWriter) Close() error {
	if w.writer != nil {
		return w.writer.Close()
	}
	return nil
}

// gzipWriterWrapper wraps gzip.Writer to return it to the pool
type gzipWriterWrapper struct {
	*gzip.Writer
}

func (w *gzipWriterWrapper) Close() error {
	err := w.Writer.Close()
	gzipWriterPool.Put(w.Writer)
	return err
}

// brotliWriterWrapper wraps brotli.Writer to return it to the pool
type brotliWriterWrapper struct {
	*brotli.Writer
}

func (w *brotliWriterWrapper) Close() error {
	err := w.Writer.Close()
	brotliWriterPool.Put(w.Writer)
	return err
}

// CompressionMiddleware creates middleware that compresses responses based on Accept-Encoding
func CompressionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check Accept-Encoding header
		acceptEncoding := r.Header.Get("Accept-Encoding")
		
		// Determine which compression to use
		var encoding string
		supportsBrotli := strings.Contains(acceptEncoding, "br")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")

		// Prefer brotli over gzip if both are supported
		if supportsBrotli {
			encoding = "br"
		} else if supportsGzip {
			encoding = "gzip"
		} else {
			// No compression support
			next.ServeHTTP(w, r)
			return
		}

		// Create compression wrapper
		cw := &compressionResponseWriter{
			ResponseWriter: w,
			encoding:       encoding,
		}
		defer cw.Close()

		next.ServeHTTP(cw, r)
	})
}
