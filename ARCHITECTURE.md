# Grout Architecture

This document provides a comprehensive overview of the Grout architecture, design decisions, and implementation details.

## Table of Contents

- [Overview](#overview)
- [Architecture Diagram](#architecture-diagram)
- [Core Components](#core-components)
- [Request Flow](#request-flow)
- [Caching Strategy](#caching-strategy)
- [Image Rendering Pipeline](#image-rendering-pipeline)
- [Configuration Management](#configuration-management)
- [Design Decisions](#design-decisions)
- [Performance Considerations](#performance-considerations)
- [Security Considerations](#security-considerations)

## Overview

Grout is a high-performance HTTP service written in Go that generates images on-demand. It's designed to be:

- **Fast**: In-memory LRU caching with ETag support
- **Lightweight**: Small binary (~15MB Docker image)
- **Scalable**: Stateless design suitable for horizontal scaling
- **Simple**: No external dependencies or databases required

### Key Features

- Avatar generation with user initials
- Placeholder image generation
- Multiple image format support (SVG, PNG, JPG, WebP, GIF)
- Configurable colors and styling
- Quote and joke integration
- Gradient backgrounds
- Text wrapping and dynamic font sizing

## Architecture Diagram

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │ HTTP Request
       ▼
┌─────────────────────────────────────────┐
│           HTTP Server (:8080)           │
│  ┌───────────────────────────────────┐  │
│  │      Router (ServeMux)            │  │
│  │  - /avatar/                       │  │
│  │  - /placeholder/                  │  │
│  │  - /                              │  │
│  │  - /robots.txt                    │  │
│  │  - /sitemap.xml                   │  │
│  └──────────────┬────────────────────┘  │
└─────────────────┼────────────────────────┘
                  │
       ┌──────────┴──────────┐
       ▼                     ▼
┌─────────────┐      ┌──────────────┐
│  Handlers   │      │  LRU Cache   │
│  Package    │◄────►│   (2000)     │
└──────┬──────┘      └──────────────┘
       │
       │ Generate Image
       ▼
┌─────────────┐
│   Render    │
│   Package   │
│  ┌────────┐ │
│  │ gg lib │ │
│  └────────┘ │
│  ┌────────┐ │
│  │ Fonts  │ │
│  └────────┘ │
└──────┬──────┘
       │
       ▼ Image bytes
┌─────────────┐
│   Response  │
│  + ETag     │
│  + Headers  │
└─────────────┘
```

## Core Components

### 1. cmd/grout/main.go

**Responsibility**: Application entry point and server initialization

- Loads configuration from environment variables and CLI flags
- Initializes the HTTP server
- Registers routes and handlers
- Manages graceful shutdown

**Key Functions**:
- `main()`: Initializes and starts the server
- Route registration using `handlers.RegisterRoutes()`

### 2. internal/config/config.go

**Responsibility**: Configuration management

**Constants**:
- `DefaultAddr`: Default bind address (`:8080`)
- `DefaultCacheSize`: Default LRU cache size (2000)
- `DefaultDomain`: Default domain for examples
- `DefaultStaticDir`: Default static files directory

**Struct**: `ServerConfig`
- `Addr`: HTTP bind address
- `CacheSize`: LRU cache capacity
- `Domain`: Public domain name
- `StaticDir`: Static files location

**Functions**:
- `LoadServerConfig()`: Parses environment variables and CLI flags

### 3. internal/handlers/handlers.go

**Responsibility**: HTTP request handling and response generation

**Struct**: `Service`
- `cache`: LRU cache for rendered images
- `domain`: Domain for URL generation
- `staticDir`: Location of static files

**Key Methods**:
- `ServeAvatar()`: Handles `/avatar/` requests
- `ServePlaceholder()`: Handles `/placeholder/` requests
- `ServeHome()`: Serves the homepage with API examples
- `ServeRobotsTxt()`: Serves robots.txt
- `ServeSitemap()`: Serves sitemap.xml
- `serveImage()`: Common image serving logic with caching and ETag support

**Route Registration**:
```go
func RegisterRoutes(mux *http.ServeMux, s *Service) {
    mux.HandleFunc("/avatar/", s.ServeAvatar)
    mux.HandleFunc("/placeholder/", s.ServePlaceholder)
    mux.HandleFunc("/robots.txt", s.ServeRobotsTxt)
    mux.HandleFunc("/sitemap.xml", s.ServeSitemap)
    mux.HandleFunc("/", s.ServeHome)
}
```

### 4. internal/render/render.go

**Responsibility**: Image generation and rendering logic

**Key Functions**:

- `DrawAvatar()`: Creates circular or square avatars with initials
  - Extracts initials from names
  - Handles color generation and contrast
  - Supports bold fonts and rounded shapes

- `DrawPlaceholder()`: Creates rectangular placeholder images
  - Supports custom text or quotes/jokes
  - Handles gradients and solid backgrounds
  - Dynamic font sizing and text wrapping

- `ParseHexColor()`: Converts hex strings to color.Color
  - Supports 3-digit and 6-digit hex codes
  - Fallback to gray for invalid input

- `GenerateColorHash()`: Creates deterministic colors from strings
  - Uses MD5 hash for consistency
  - Returns vibrant, saturated colors

- `GetContrastColor()`: Determines readable text color
  - Calculates luminance
  - Returns black or white for optimal contrast

- `ExtractInitials()`: Extracts initials from names
  - Supports single and multiple words
  - Handles special characters

**Format Support**:
- SVG (default, vector graphics)
- PNG (raster, lossless)
- JPG/JPEG (raster, lossy)
- WebP (modern, efficient)
- GIF (legacy support)

### 5. internal/utils/parse.go

**Responsibility**: Utility functions for parsing and validation

**Key Functions**:
- `ParseIntOrDefault()`: Safely parses integers with fallback
- `ParseBoolOrDefault()`: Safely parses booleans with fallback

## Request Flow

### Avatar Generation Flow

```
1. Client Request
   GET /avatar/John+Doe.png?size=256&rounded=true&bg=random

2. Handler (ServeAvatar)
   ├─ Extract name from path or query
   ├─ Parse query parameters (size, bg, color, rounded, bold)
   ├─ Parse image format from extension
   └─ Build cache key

3. Cache Check
   ├─ Check ETag (If-None-Match header)
   ├─ If match → Return 304 Not Modified
   └─ If miss → Continue to generation

4. Image Generation
   ├─ Extract initials from name
   ├─ Generate/parse colors
   ├─ Create image context (gg.Context)
   ├─ Draw background (circle or square)
   ├─ Calculate font size
   ├─ Draw text (initials)
   └─ Encode to requested format

5. Response
   ├─ Set Content-Type header
   ├─ Set Cache-Control header (max-age=31536000)
   ├─ Set ETag header
   ├─ Set X-Cache header (HIT or MISS)
   └─ Send image bytes
```

### Placeholder Generation Flow

```
1. Client Request
   GET /placeholder/800x400.png?quote=true&category=inspirational

2. Handler (ServePlaceholder)
   ├─ Extract dimensions from path or query
   ├─ Parse query parameters (text, quote, joke, category, bg, color)
   ├─ Parse image format from extension
   └─ Build cache key

3. Cache Check
   ├─ Check ETag (If-None-Match header)
   ├─ If match → Return 304 Not Modified
   └─ If miss → Continue to generation

4. Content Selection
   ├─ If quote=true → Select random quote
   ├─ If joke=true → Select random joke
   ├─ If text provided → Use custom text
   └─ Default → Use "{width} x {height}"

5. Image Generation
   ├─ Parse/generate colors
   ├─ Create image context
   ├─ Draw background (solid or gradient)
   ├─ Calculate font size based on text length
   ├─ Wrap text if needed
   ├─ Draw text (centered, multi-line)
   └─ Encode to requested format

6. Response
   ├─ Set headers (same as avatar)
   └─ Send image bytes
```

## Caching Strategy

### LRU Cache Implementation

Grout uses an in-memory LRU (Least Recently Used) cache from `github.com/hashicorp/golang-lru/v2`.

**Configuration**:
- Default size: 2000 entries
- Configurable via `CACHE_SIZE` env var or `-cache-size` flag

**Cache Key Format**:

Avatar:
```
Avatar:{name}:{size}:{rounded}:{bold}:{bg}:{fg}:{format}
```

Placeholder:
```
PH:{width}:{height}:{bg}:{fg}:{text}:{format}
```

**Benefits**:
- Reduces CPU usage for repeated requests
- Prevents memory exhaustion (fixed size)
- Fast O(1) lookups
- Thread-safe operations

### ETag Support

ETags are generated from cache keys using MD5 hashing:

```go
hash := md5.Sum([]byte(cacheKey))
etag := hex.EncodeToString(hash[:])
```

**Behavior**:
- Client sends `If-None-Match: "etag-value"`
- Server compares with current ETag
- If match → 304 Not Modified (no body)
- If no match → 200 OK with image

**Benefits**:
- Reduces bandwidth usage
- Faster response times
- Better CDN compatibility

### Cache Headers

```
Cache-Control: public, max-age=31536000, immutable
ETag: "md5-hash-of-cache-key"
X-Cache: HIT (or MISS)
```

**Rationale**:
- `public`: Allows CDN caching
- `max-age=31536000`: 1 year (effectively forever)
- `immutable`: Prevents revalidation
- Images are content-addressed (parameters in URL)

## Image Rendering Pipeline

### Font Handling

Grout embeds Go fonts to avoid external dependencies:

```go
import "golang.org/x/image/font/gofont/goregular"
import "golang.org/x/image/font/gofont/gobold"
```

**Font Selection**:
- Default: Go Regular
- Bold: Go Bold (when `bold=true`)

**Font Size Calculation**:

Avatar:
```go
if len(initials) <= 2 {
    fontSize = minDim * 0.5  // 50% of smallest dimension
} else {
    fontSize = minDim * 0.15  // 15% for longer text
}
fontSize = max(fontSize, 12)  // Minimum 12px
```

Placeholder:
```go
baseSize := 48  // Start large
maxWidth := width * 0.8  // 80% of image width
maxHeight := height * 0.8

// Reduce size until text fits
while textWidth > maxWidth || textHeight > maxHeight {
    fontSize--
    if fontSize < 16 {  // Minimum 16px
        break
    }
}
```

### Color Processing

**Hex Parsing**:
```go
func ParseHexColor(s string) color.Color {
    if len(s) == 3 {
        // Expand: "f0e" → "ff00ee"
    } else if len(s) == 6 {
        // Parse directly
    } else {
        // Fallback to gray
        return color.RGBA{128, 128, 128, 255}
    }
}
```

**Contrast Calculation**:
```go
func GetContrastColor(bg color.Color) color.Color {
    r, g, b, _ := bg.RGBA()
    
    // Convert to 0-1 range
    rf := float64(r) / 65535.0
    gf := float64(g) / 65535.0
    bf := float64(b) / 65535.0
    
    // Calculate luminance (relative)
    lum := 0.299*rf + 0.587*gf + 0.114*bf
    
    if lum > 0.5 {
        return color.Black
    }
    return color.White
}
```

### Gradient Support

Placeholder images support linear gradients:

```go
// Parse: "ff0000,0000ff" → Red to Blue
colors := strings.Split(bgParam, ",")
if len(colors) == 2 {
    color1 := ParseHexColor(colors[0])
    color2 := ParseHexColor(colors[1])
    
    // Create gradient (top to bottom)
    for y := 0; y < height; y++ {
        ratio := float64(y) / float64(height)
        // Interpolate between color1 and color2
    }
}
```

### Text Wrapping

For quotes and jokes, text is automatically wrapped:

```go
func WrapText(text string, maxWidth float64, ctx *gg.Context) []string {
    words := strings.Fields(text)
    lines := []string{}
    currentLine := ""
    
    for _, word := range words {
        testLine := currentLine + " " + word
        w, _ := ctx.MeasureString(testLine)
        
        if w > maxWidth {
            lines = append(lines, currentLine)
            currentLine = word
        } else {
            currentLine = testLine
        }
    }
    
    if currentLine != "" {
        lines = append(lines, currentLine)
    }
    
    return lines
}
```

## Configuration Management

### Priority Order

1. **Defaults** (in code)
2. **Environment Variables**
3. **CLI Flags** (highest priority)

### Configuration Loading

```go
func LoadServerConfig() (*ServerConfig, error) {
    // Start with defaults
    cfg := &ServerConfig{
        Addr:      DefaultAddr,
        CacheSize: DefaultCacheSize,
        Domain:    DefaultDomain,
        StaticDir: DefaultStaticDir,
    }
    
    // Override with environment variables
    if addr := os.Getenv("ADDR"); addr != "" {
        cfg.Addr = addr
    }
    
    // Override with CLI flags
    flag.StringVar(&cfg.Addr, "addr", cfg.Addr, "HTTP bind address")
    flag.Parse()
    
    return cfg, nil
}
```

### Docker Configuration

Environment variables are the primary configuration method for Docker:

```yaml
environment:
  ADDR: ":8080"
  CACHE_SIZE: "2000"
  DOMAIN: "grout.example.com"
  STATIC_DIR: "/app/static"
```

## Design Decisions

### Why Go?

- **Performance**: Fast execution, efficient memory usage
- **Concurrency**: Native goroutines for handling multiple requests
- **Simplicity**: Single binary deployment, no runtime dependencies
- **Ecosystem**: Excellent libraries for image processing and HTTP servers

### Why LRU Cache?

- **Predictable Memory**: Fixed size prevents unbounded growth
- **Effective**: Most recently used items are kept (temporal locality)
- **Simple**: No complex invalidation logic needed
- **Thread-Safe**: Built-in synchronization

### Why Embedded Fonts?

- **Portability**: No external font files needed
- **Consistency**: Same appearance across deployments
- **Simplicity**: Reduces configuration complexity
- **Security**: No file system access required

### Why Multiple Image Formats?

- **Flexibility**: Different use cases prefer different formats
- **Compatibility**: SVG for web, PNG for apps, JPG for photos
- **Modern Web**: WebP for better compression
- **Legacy Support**: GIF for older systems

### Why Content-Addressed URLs?

- **Immutability**: Same parameters = same image
- **Caching**: Enables aggressive cache headers
- **CDN-Friendly**: Perfect for edge caching
- **Simple**: No database or state management

## Performance Considerations

### Memory Management

**Cache Size**: Default 2000 entries
- Average image: ~50KB
- Total cache: ~100MB
- Configurable for different workloads

**Image Generation**: Memory scales with image size
- 128x128: ~16KB per image
- 1024x1024: ~1MB per image
- LRU eviction prevents memory exhaustion

### CPU Usage

**Rendering**: CPU-intensive operations
- Font rendering: Moderate CPU
- Image encoding: Varies by format (PNG > JPG > WebP > SVG)
- Caching reduces repeated work

**Optimization**:
- Cache frequently requested images
- Use appropriate image sizes
- Consider rate limiting for public deployments

### Concurrency

**Go HTTP Server**: Handles concurrent requests automatically
- Each request in a separate goroutine
- LRU cache is thread-safe
- No blocking operations in request path

**Scalability**:
- Stateless design: Easy horizontal scaling
- No database: Eliminates bottleneck
- Share-nothing architecture: Linear scalability

### Network

**Response Size**:
- SVG: 1-5KB (vector, smallest)
- PNG: 10-100KB (lossless, medium)
- JPG: 5-50KB (lossy, small)
- WebP: 5-30KB (modern, efficient)

**CDN Integration**:
- Immutable cache headers
- ETag support for revalidation
- X-Cache header for monitoring

## Security Considerations

### Input Validation

**Size Limits**: All dimensions validated
```go
if size <= 0 || size > 4096 {
    size = DefaultSize  // Fallback to 128
}
```

**Color Parsing**: Invalid colors fallback to safe defaults
```go
if !isValidHex(colorStr) {
    return color.RGBA{128, 128, 128, 255}  // Gray
}
```

**Text Sanitization**: All text is properly escaped
- No HTML injection in SVG
- No special characters in file generation

### Resource Limits

**Cache Size**: Prevents memory exhaustion
- Fixed maximum entries
- Automatic LRU eviction

**Image Dimensions**: Reasonable limits
- Maximum suggested: 4096x4096
- Larger sizes fallback to defaults

### HTTP Security

**Security Headers**: HTML responses include comprehensive security headers
```go
Content-Security-Policy: default-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; script-src 'self' 'unsafe-inline'
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
```

These headers are automatically applied to:
- Home page (`/`)
- Play page (`/play`)
- Error pages (4xx, 5xx)

Image endpoints (`/avatar/`, `/placeholder/`) do not include these headers as they serve binary content.

**No User Data**: Stateless, no user data stored
- No cookies
- No sessions
- No personal information

### Docker Security

**Non-Root User**: Container runs as unprivileged user
```dockerfile
USER nobody:nobody
```

**Static Binary**: No runtime dependencies
```dockerfile
CGO_ENABLED=0 go build
```

**Minimal Image**: Small attack surface
```dockerfile
FROM scratch
COPY --from=builder /app/grout /grout
```

## Future Enhancements

Potential areas for expansion:

1. **Features**:
   - Additional shapes (triangles, hexagons)
   - Custom font uploads
   - Animation support (animated GIFs)
   - SVG filters and effects

2. **Performance**:
   - Distributed caching (Redis)
   - Pre-warming cache
   - HTTP/2 support
   - Compression middleware

3. **Monitoring**:
   - Prometheus metrics
   - Health check endpoint
   - Structured logging
   - Distributed tracing

4. **API**:
   - REST API versioning
   - GraphQL endpoint
   - Batch generation
   - Webhook notifications

---

For more information, see:
- [README.md](README.md) - API documentation
- [CONTRIBUTING.md](CONTRIBUTING.md) - Contributing guidelines
- [SECURITY.md](SECURITY.md) - Security policy
