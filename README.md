# Grout API Guide

Grout is a small HTTP service that renders PNG avatars with user initials and rectangular placeholder images. It relies on the `github.com/fogleman/gg` drawing library and embeds Go fonts for crisp text output.

## Quick Start

### Using Docker Compose

```bash
docker compose up --build
```

### Using Go directly

```bash
go run ./cmd/avata-go
```

### Using pre-built binaries

Download the latest release for your platform from the [Releases page](https://github.com/your-username/go-avatars/releases) and run:

```bash
# Linux/macOS
./avata-go-linux-amd64

# Windows
avata-go-windows-amd64.exe
```

The server listens on `:8080` by default and exposes the routes below.

## `/avatar/` Endpoint

Generates a square avatar that displays the initials derived from the provided name.

- **Path**: `/avatar/{name}[.ext]` where `ext` can be `png`, `jpg`, `jpeg`, `gif`, or `webp`. You can also use the `name` query parameter.
- **Format**: Images are served as WebP by default when no extension is specified. Use `.png`, `.jpg`, `.jpeg`, `.gif`, or `.webp` extension to request a specific format.
- **Size**: `size` query parameter (default `128`), applied to both width and height.
- **Background Color**: `background` query parameter accepts hex (`f0e9e9`) or the literal `random` to derive a deterministic color per name.
- **Text Color**: `color` query parameter (hex, default auto-contrasted).
- **Rounded**: `rounded=true` draws a circle instead of a square.
- **Bold**: `bold=true` switches to the embedded Go Bold font.

Examples:

```bash
# Default WebP format
curl "http://localhost:8080/avatar/Jane+Doe?size=256&rounded=true&bold=true&background=random"

# PNG format
curl "http://localhost:8080/avatar/Jane+Doe.png?size=256&rounded=true&bold=true&background=random"

# JPG format
curl "http://localhost:8080/avatar/Jane+Doe.jpg?size=256"

# WebP format (explicit)
curl "http://localhost:8080/avatar/Jane+Doe.webp?size=256"
```

## `/placeholder/` Endpoint

Creates a rectangular placeholder image with custom dimensions and optional overlay text.

- **Path Form**: `/placeholder/{width}x{height}[.ext]` where `ext` can be `png`, `jpg`, `jpeg`, `gif`, or `webp`. If extension is omitted, images are served as WebP by default.
- **Format**: Images are served as WebP by default when no extension is specified. Use `.png`, `.jpg`, `.jpeg`, `.gif`, or `.webp` extension to request a specific format.
- **Dimensions**: Can also use query parameters `w` and `h` (default `128`).
- **Text**: `text` query parameter (defaults to "{width} x {height}").
- **Background Color**: `bg` query parameter (hex, default `cccccc`). Supports gradients with comma-separated colors (e.g., `ff0000,0000ff` for red to blue).
- **Text Color**: `color` query parameter (hex, default auto-contrasted).

Examples:

```bash
# Default WebP format
curl "http://localhost:8080/placeholder/800x400?text=Hero+Image&bg=222222&color=f5f5f5"

# PNG format
curl "http://localhost:8080/placeholder/800x400.png?text=Hero+Image&bg=222222&color=f5f5f5"

# JPG format
curl "http://localhost:8080/placeholder/1200x600.jpg?text=Banner"

# GIF format
curl "http://localhost:8080/placeholder/400x400.gif"

# Gradient background (red to blue)
curl "http://localhost:8080/placeholder/800x400.png?bg=ff0000,0000ff&text=Gradient"

# Gradient background (green to yellow)
curl "http://localhost:8080/placeholder/1200x600?bg=00ff00,ffff00"
```

## Response Characteristics

- Images are served as WebP by default (when no extension is specified). The `Content-Type` header is set based on the requested format: `image/webp`, `image/png`, `image/jpeg`, or `image/gif`.
- Successful responses include `Cache-Control: public, max-age=31536000, immutable` and an `ETag` keyed by the query parameters and format.
- Cached entries are stored in an in-memory LRU (`CacheSize = 2000`) to reduce rendering overhead. Cache hits expose the header `X-Cache: HIT`.

## Error Handling

If generation fails (for example due to invalid parameters), the server responds with HTTP `500` and `Failed to generate image`. Invalid dimensions fallback to safe defaults to keep the server responsive.

## Configuration

- `ADDR` env var or `-addr` flag controls the HTTP bind address (default `:8080`).
- `CACHE_SIZE` env var or `-cache-size` flag sets LRU entry count (default `2000`).

### Docker Configuration

When using Docker Compose, you can override environment variables in `docker-compose.yml`:

```yaml
environment:
  ADDR: ":3000"
  CACHE_SIZE: "5000"
```

## Building from Source

### Build binary

```bash
go build -o avata-go ./cmd/avata-go
```

### Build Docker image

```bash
docker build -t avata-go .
```

### Run Docker container

```bash
docker run -p 8080:8080 -e ADDR=":8080" avata-go
```

## CI/CD

The project includes GitHub Actions workflows that automatically:

### Test Workflow (`.github/workflows/test.yml`)
Runs on every pull request and push to main/master:
- **Tests**: Runs all unit tests with race detection and coverage reporting
- **Lint**: Runs `golangci-lint` for code quality checks
- **Format**: Verifies code is properly formatted with `go fmt`
- **Vet**: Runs `go vet` to catch common issues
- **Build Check**: Ensures the code builds successfully for all target platforms
- **Coverage**: Optionally uploads coverage to Codecov (requires `CODECOV_TOKEN` secret)

### Release Workflow (`.github/workflows/release.yml`)
Runs when you create a GitHub release:
- Builds binaries for Linux, macOS, and Windows (amd64 and arm64)
- Attaches them to GitHub releases
- Builds and pushes multi-platform Docker images to Docker Hub (if credentials are configured)

### Setup Secrets

To enable Docker Hub publishing, add these secrets to your GitHub repository:
- `DOCKER_USERNAME`: Your Docker Hub username
- `DOCKER_PASSWORD`: Your Docker Hub access token

To enable Codecov integration (optional):
- `CODECOV_TOKEN`: Your Codecov upload token

## Development Tips

- Customize the defaults by editing the constants in `internal/config/config.go`.
- Extend `DrawImage` in `internal/render/render.go` if you need additional shapes, padding, or font scaling strategies.
- Consider fronting the service with a CDN when deploying to production so the long-lived cache headers are effective.
- Run tests with `go test ./...`

