# GitHub Copilot Instructions for Grout

## Project Overview

Grout is a high-performance HTTP service written in Go that generates PNG images on-demand:
- **Avatar Generation**: Creates circular or square avatars with user initials
- **Placeholder Images**: Generates rectangular placeholder images with custom dimensions and text
- **Caching**: Uses in-memory LRU cache (2000 entries) with ETag support for optimal performance

## Architecture

### Project Structure
```
go-avatars/
├── cmd/avata-go/          # Application entry point
│   └── main.go            # Server initialization and routing setup
├── internal/
│   ├── config/            # Configuration and environment handling
│   │   └── config.go      # Server config, defaults, env/flag parsing
│   ├── handlers/          # HTTP request handlers
│   │   ├── handlers.go    # Avatar, placeholder routes and caching logic
│   │   └── handlers_test.go
│   ├── render/            # Image rendering logic
│   │   ├── render.go      # Drawing, fonts, color parsing, initials extraction
│   │   └── render_test.go
│   └── utils/             # Shared utilities
│       ├── parse.go       # Parsing helpers (int conversion, color hashing)
│       └── parse_test.go
├── .github/workflows/     # CI/CD automation
│   ├── test.yml           # PR testing, linting, build checks
│   └── release.yml        # Binary builds and Docker publishing
├── Dockerfile             # Multi-stage production build
├── docker-compose.yml     # Local development setup
└── README.md              # API documentation
```

### Key Dependencies
- `github.com/fogleman/gg` - 2D graphics rendering
- `github.com/golang/freetype` - TrueType font parsing
- `github.com/hashicorp/golang-lru/v2` - LRU cache implementation
- `golang.org/x/image/font/gofont` - Embedded Go fonts (regular & bold)

## Coding Standards

### Go Best Practices
- **Go Version**: 1.24+
- **Code Style**: Follow standard Go formatting (`gofmt`)
- **Error Handling**: Always handle errors explicitly, wrap with context using `fmt.Errorf`
- **Naming**: Use idiomatic Go naming (camelCase for unexported, PascalCase for exported)
- **Testing**: Write table-driven tests with meaningful names

### Package Organization
- `internal/` packages for encapsulation (not importable by external projects)
- Dependency injection pattern for testability (see `handlers.Service`)
- Separate concerns: config, rendering, HTTP handling, utilities

### HTTP Handlers
- Use query parameters for configuration (size, colors, flags)
- Support both path-based and query-based inputs
- Always set proper headers: `Content-Type`, `Cache-Control`, `ETag`, `X-Cache`
- Return `400` for bad requests, `500` for internal errors
- Use `serveImage` helper for consistent caching and ETag handling

### Image Rendering
- Default size: 128px
- Support hex colors (6-digit or 3-digit shorthand, both without `#` prefix)
- Auto-contrast: Use `GetContrastColor` to determine readable text color
- Font sizing: 50% of min dimension for 1-2 chars, 15% for longer text (min 12px)
- Embed fonts (no external dependencies)

### Configuration
- Priority: defaults → environment variables → CLI flags
- Use `config.LoadServerConfig()` for runtime configuration
- Environment variables: `ADDR`, `CACHE_SIZE`
- CLI flags: `-addr`, `-cache-size`

### Testing Requirements
- Unit tests for all business logic (utils, render helpers, handlers)
- Table-driven tests for multiple scenarios
- Race detection enabled in CI (`go test -race`)
- Coverage reporting (target: 80%+)
- HTTP tests using `httptest.NewRecorder()` and `httptest.NewRequest()`

### Docker & Deployment
- Multi-stage builds for minimal image size (~15MB final)
- `CGO_ENABLED=0` for static binaries
- Support multiple platforms: linux/amd64, linux/arm64
- Graceful configuration via environment variables

## API Endpoints

### `/avatar/` - Avatar Generation
- **Input**: Name (path or query), size, background, color, rounded, bold
- **Output**: PNG image with initials
- **Caching**: Key format `Avatar:{name}:{size}:{rounded}:{bold}:{bg}:{fg}`

### `/placeholder/` - Placeholder Images
- **Input**: Width, height (path `{w}x{h}` or query), text, bg, color
- **Output**: PNG image with dimensions text
- **Caching**: Key format `PH:{width}:{height}:{bg}:{fg}:{text}`

## Common Patterns

### Adding a New Endpoint
1. Add handler method to `handlers.Service`
2. Register route in `RegisterRoutes()`
3. Parse query parameters using `utils.ParseIntOrDefault()`
4. Build cache key with all parameters
5. Use `serveImage()` helper with generator function
6. Add tests in `handlers_test.go`

### Adding Configuration
1. Add constant to `internal/config/config.go`
2. Add field to `ServerConfig` struct if runtime-configurable
3. Add environment variable parsing in `LoadServerConfig()`
4. Add CLI flag if needed
5. Document in README.md

### Color Handling
- Parse with `render.ParseHexColor()` (handles 3/6 digit, fallback to gray)
- Generate random with `render.GenerateColorHash()` (deterministic MD5-based)
- Auto-contrast with `render.GetContrastColor()` (luminance calculation)

## CI/CD Workflow

### On Pull Request (test.yml)
- Run `go test -race -coverprofile=coverage.out ./...`
- Run `golangci-lint` for code quality
- Verify `go fmt` compliance
- Run `go vet` for common mistakes
- Build for all platforms to ensure compatibility

### On Release (release.yml)
- Build binaries: linux, darwin, windows (amd64, arm64)
- Create archives (`.tar.gz` for Unix, `.zip` for Windows)
- Upload to GitHub release assets
- Build and push multi-arch Docker images to Docker Hub

## Security & Performance

- No external file I/O (uses embedded fonts)
- Input validation: reject negative/zero sizes, handle empty strings
- LRU cache prevents memory exhaustion (configurable size)
- Static builds with `CGO_ENABLED=0` for portability
- Immutable cache headers for CDN optimization
- ETag support for conditional requests (304 responses)

## Future Enhancements to Consider

When extending this codebase:
- Add more shape options (triangle, hexagon)
- Support custom font uploads
- Implement rate limiting
- Add metrics/observability (Prometheus)
- Support custom color palettes
- Add gradient backgrounds
- Implement request logging middleware

