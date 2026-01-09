# GitHub Issues to Create from Technical Audit

This document lists the GitHub issues that should be created based on the Technical Audit Report. Issues are organized by priority and include all necessary details for implementation.

---

## 🔴 Critical Priority Issues

### Issue 1: Add Rate Limiting to Prevent DoS Attacks
**Labels:** `security`, `enhancement`, `priority: critical`

**Description:**
The service currently has no rate limiting, making it vulnerable to DoS attacks through unlimited requests. This can exhaust CPU/memory and cause service degradation.

**Proposed Solution:**
Implement per-IP rate limiting using `golang.org/x/time/rate`:
- 100 requests per minute per IP for `/avatar/` and `/placeholder/`
- No limit for static assets (favicon, robots.txt)
- Return HTTP 429 (Too Many Requests) when limit exceeded

**Acceptance Criteria:**
- [ ] Rate limiter middleware implemented
- [ ] Configurable limits via environment variables
- [ ] Tests for rate limiting behavior
- [ ] Documentation updated

**Reference:** TECHNICAL_AUDIT.md Section 3.1

---

### Issue 2: Add HTTP Server Timeouts
**Labels:** `security`, `enhancement`, `priority: critical`

**Description:**
The HTTP server has no timeouts configured, making it vulnerable to slowloris attacks and hanging connections.

**Proposed Solution:**
Configure server timeouts in main.go:
```go
import (
    "time"
    "net/http"
)

srv := &http.Server{
    Addr:         cfg.Addr,
    Handler:      mux,
    ReadTimeout:  5 * time.Second,
    WriteTimeout: 10 * time.Second,
    IdleTimeout:  120 * time.Second,
}
```

**Acceptance Criteria:**
- [ ] Server timeouts configured
- [ ] Configurable via environment variables (optional)
- [ ] Documentation updated

**Reference:** TECHNICAL_AUDIT.md Section 3.2

---

### Issue 3: Implement Configuration Validation at Startup
**Labels:** `bug`, `enhancement`, `priority: critical`

**Description:**
Configuration values loaded from environment/flags are not validated, allowing invalid values (negative cache sizes, malformed addresses) to cause runtime issues.

**Proposed Solution:**
Add `Validate()` method to ServerConfig:
- Check cache size is between 1 and 100,000
- Validate address format
- Verify static directory exists or can be created
- Validate domain format

**Acceptance Criteria:**
- [ ] Validation method implemented
- [ ] Called at startup before server starts
- [ ] Clear error messages for each validation failure
- [ ] Tests for validation logic

**Reference:** TECHNICAL_AUDIT.md Section 1.1

---

### Issue 4: Add Response Size Limits
**Labels:** `security`, `performance`, `priority: critical`

**Description:**
No limits on image dimensions can lead to memory exhaustion and potential DoS. Users can request 10000x10000 images.

**Proposed Solution:**
Add maximum dimension constant (4096px) and enforce in parsing:
```go
const MaxImageDimension = 4096

func ParseIntWithLimit(s string, def, max int) int {
    val := ParseIntOrDefault(s, def)
    if val > max {
        return max
    }
    return val
}
```

**Acceptance Criteria:**
- [ ] Maximum dimension enforced for width and height
- [ ] Configurable via environment variable
- [ ] Tests for limit enforcement
- [ ] Documentation updated with limits

**Reference:** TECHNICAL_AUDIT.md Section 2.2

---

## 🟡 High Priority Issues

### Issue 5: Add Observability (Metrics and Structured Logging)
**Labels:** `enhancement`, `priority: high`, `operations`

**Description:**
No metrics or structured logging makes production debugging difficult. Need visibility into:
- Request rates and status codes
- Cache hit rates
- Render durations by format
- Error rates

**Proposed Solution:**
1. Add Prometheus metrics endpoint `/metrics`
2. Implement structured logging with zerolog or zap
3. Track key metrics:
   - `grout_requests_total{endpoint, status}`
   - `grout_cache_hits_total{hit}`
   - `grout_render_duration_seconds{format}`

**Acceptance Criteria:**
- [ ] Prometheus metrics implemented
- [ ] Structured logging throughout codebase
- [ ] Metrics endpoint exposed
- [ ] Documentation with example Grafana dashboard

**Reference:** TECHNICAL_AUDIT.md Section 6.1

---

### Issue 6: Implement Graceful Shutdown
**Labels:** `enhancement`, `priority: high`, `operations`

**Description:**
Server doesn't handle SIGTERM gracefully, causing in-flight requests to fail during deployments.

**Proposed Solution:**
Implement signal handling for graceful shutdown:
- Listen for SIGTERM/SIGINT
- Stop accepting new requests
- Wait up to 30 seconds for in-flight requests to complete
- Close server

**Acceptance Criteria:**
- [ ] Graceful shutdown implemented
- [ ] Configurable timeout
- [ ] Logs shutdown progress
- [ ] Tests for shutdown behavior

**Reference:** TECHNICAL_AUDIT.md Section 6.2

---

### Issue 7: Add Gzip/Brotli Compression Middleware
**Labels:** `enhancement`, `performance`, `priority: high`

**Description:**
Responses are not compressed, wasting bandwidth. SVG responses can compress 70-80%.

**Proposed Solution:**
Add compression middleware using `github.com/NYTimes/gziphandler` or similar:
- Compress responses based on Accept-Encoding
- Support gzip and brotli
- Skip compression for already-compressed formats (PNG already compressed)

**Acceptance Criteria:**
- [ ] Compression middleware implemented
- [ ] Respects Accept-Encoding header
- [ ] Tests verify compression
- [ ] Performance benchmarks show improvement
- [ ] Documentation updated

**Estimated Impact:** 70-80% bandwidth reduction for SVG, 20-30% for PNG

**Reference:** TECHNICAL_AUDIT.md Section 2.1

---

### Issue 8: Create OpenAPI Specification
**Labels:** `documentation`, `priority: high`

**Description:**
No machine-readable API documentation makes integration harder. OpenAPI spec would enable:
- Auto-generated client libraries
- Interactive API explorer (Swagger UI)
- Better developer experience

**Proposed Solution:**
Create `docs/openapi.yaml` with complete API specification:
- All endpoints documented
- All parameters and their types
- Response schemas
- Example requests/responses

**Acceptance Criteria:**
- [ ] OpenAPI 3.0 spec created
- [ ] All endpoints documented
- [ ] Hosted Swagger UI (optional)
- [ ] Referenced in README.md

**Reference:** TECHNICAL_AUDIT.md Section 5.1

---

### Issue 9: Add Performance Benchmarks
**Labels:** `testing`, `performance`, `priority: high`

**Description:**
No benchmarks to track performance regressions over time.

**Proposed Solution:**
Add benchmark tests for critical paths:
- Avatar generation (all formats)
- Placeholder generation (all formats)
- Cache operations
- Text wrapping

**Acceptance Criteria:**
- [ ] Benchmark tests in `*_bench_test.go` files
- [ ] CI runs benchmarks and compares with main branch
- [ ] Regression alerts if performance drops >10%
- [ ] Documentation on running benchmarks

**Reference:** TECHNICAL_AUDIT.md Section 4.1

---

### Issue 10: Implement Content Negotiation (Accept Header)
**Labels:** `enhancement`, `feature`, `priority: high`

**Description:**
Support automatic format selection based on Accept header, improving browser compatibility.

**Proposed Solution:**
Parse Accept header and select best format:
```go
func selectFormat(r *http.Request, defaultFormat ImageFormat) ImageFormat {
    accept := r.Header.Get("Accept")
    if strings.Contains(accept, "image/webp") {
        return FormatWebP
    }
    return defaultFormat
}
```

**Acceptance Criteria:**
- [ ] Accept header parsing implemented
- [ ] Falls back to default if no match
- [ ] Tests for format negotiation
- [ ] Documentation with examples

**Reference:** TECHNICAL_AUDIT.md Section 7.1

---

## 🟢 Medium Priority Issues

### Issue 11: Add Custom Font Support
**Labels:** `enhancement`, `feature`, `priority: medium`

**Description:**
Users cannot use custom fonts, limiting branding options.

**Proposed Solution:**
1. Add `/fonts` endpoint to list available fonts
2. Support font parameter in avatar/placeholder
3. Allow font upload via Docker volume or configuration

**Acceptance Criteria:**
- [ ] Font loading from directory
- [ ] Font query parameter support
- [ ] Tests with custom fonts
- [ ] Documentation with examples

**Reference:** TECHNICAL_AUDIT.md Section 7.3

---

### Issue 12: Add Batch API Endpoint
**Labels:** `enhancement`, `feature`, `priority: medium`

**Description:**
No way to generate multiple images in one request, requiring multiple round trips.

**Proposed Solution:**
Add POST `/batch` endpoint:
```json
{
  "images": [
    {"type": "avatar", "name": "John", "size": 128},
    {"type": "placeholder", "width": 800, "height": 600}
  ]
}
```

**Acceptance Criteria:**
- [ ] Batch endpoint implemented
- [ ] Returns array of URLs or errors
- [ ] Respects rate limits per image
- [ ] Tests and documentation

**Reference:** TECHNICAL_AUDIT.md Section 7.2

---

### Issue 13: Add Border and Shadow Effects
**Labels:** `enhancement`, `feature`, `priority: medium`

**Description:**
No visual effects available, limiting design options.

**Proposed Solution:**
Add query parameters:
- `border=2` (width in pixels)
- `border-color=000000`
- `shadow=true`

**Acceptance Criteria:**
- [ ] Border rendering implemented
- [ ] Shadow effect implemented
- [ ] Works with rounded avatars
- [ ] Tests for all effects
- [ ] Documentation with examples

**Reference:** TECHNICAL_AUDIT.md Section 7.4

---

### Issue 14: Add Kubernetes Deployment Manifests
**Labels:** `operations`, `documentation`, `priority: medium`

**Description:**
No Kubernetes resources for easy deployment in K8s clusters.

**Proposed Solution:**
Create `deploy/kubernetes/` with:
- Deployment
- Service
- Ingress
- HorizontalPodAutoscaler
- ConfigMap

**Acceptance Criteria:**
- [ ] K8s manifests created and tested
- [ ] Helm chart (optional)
- [ ] Documentation for K8s deployment

**Reference:** TECHNICAL_AUDIT.md Section 6.5

---

### Issue 15: Add Dependabot for Dependency Updates
**Labels:** `dependencies`, `priority: medium`

**Description:**
No automated dependency updates, requiring manual monitoring.

**Proposed Solution:**
Add `.github/dependabot.yml`:
```yaml
version: 2
updates:
  - package-ecosystem: "gomod"
    directory: "/"
    schedule:
      interval: "weekly"
```

**Acceptance Criteria:**
- [ ] Dependabot configuration added
- [ ] Weekly update schedule
- [ ] Auto-merge minor/patch updates (optional)

**Reference:** TECHNICAL_AUDIT.md Section 9.2

---

### Issue 16: Refactor Large Files (render.go and handlers.go)
**Labels:** `refactoring`, `technical-debt`, `priority: medium`

**Description:**
`render.go` (540 lines) and `handlers.go` (427 lines) are becoming difficult to maintain.

**Proposed Solution:**
Split into focused files:
- `render.go` → `render.go`, `svg.go`, `raster.go`, `text.go`
- `handlers.go` → `handlers.go`, `avatar.go`, `placeholder.go`, `static.go`

**Acceptance Criteria:**
- [ ] Files split maintaining public APIs
- [ ] All tests still pass
- [ ] No behavioral changes
- [ ] Better code organization

**Reference:** TECHNICAL_AUDIT.md Section 8.1, 8.2

---

### Issue 17: Add Integration Tests
**Labels:** `testing`, `priority: medium`

**Description:**
Only unit tests exist. Need end-to-end tests with real HTTP server.

**Proposed Solution:**
Add integration tests that:
- Start real HTTP server
- Make actual HTTP requests
- Verify headers, status codes, content types
- Test error scenarios

**Acceptance Criteria:**
- [ ] Integration test suite created
- [ ] Runs in CI after unit tests
- [ ] Covers main user flows
- [ ] Fast enough for CI (<30 seconds)

**Reference:** TECHNICAL_AUDIT.md Section 4.2

---

### Issue 18: Add Security Headers
**Labels:** `security`, `enhancement`, `priority: medium`

**Description:**
Missing security headers for HTML responses.

**Proposed Solution:**
Add headers to HTML responses:
- Content-Security-Policy
- X-Content-Type-Options
- X-Frame-Options
- X-XSS-Protection

**Acceptance Criteria:**
- [ ] Security headers added to HTML endpoints
- [ ] Tests verify headers present
- [ ] Documentation updated

**Reference:** TECHNICAL_AUDIT.md Section 3.3

---

## 🔵 Low Priority Issues (Nice to Have)

### Issue 19: Add Emoji Support
**Labels:** `enhancement`, `feature`, `priority: low`

**Description:**
Support emoji in avatars instead of initials.

**Reference:** TECHNICAL_AUDIT.md Section 7.5

---

### Issue 20: Add Pattern Backgrounds
**Labels:** `enhancement`, `feature`, `priority: low`

**Description:**
Support geometric patterns (dots, stripes, checkerboard) instead of solid colors.

**Reference:** TECHNICAL_AUDIT.md Section 7.6

---

### Issue 21: Add QR Code Generation
**Labels:** `enhancement`, `feature`, `priority: low`

**Description:**
New endpoint `/qr/{data}` to generate QR codes with customizable colors and size.

**Reference:** TECHNICAL_AUDIT.md Section 7.8

---

### Issue 22: Add Image Presets System
**Labels:** `enhancement`, `feature`, `priority: low`

**Description:**
Named presets for common configurations (e.g., `?preset=social-avatar`).

**Reference:** TECHNICAL_AUDIT.md Section 7.7

---

### Issue 23: Add Animation Support
**Labels:** `enhancement`, `feature`, `priority: low`

**Description:**
Generate animated GIFs with effects (fade, pulse, rotation).

**Reference:** TECHNICAL_AUDIT.md Section 7.9

---

### Issue 24: Improve Dockerfile Security (Non-root User)
**Labels:** `security`, `operations`, `priority: low`

**Description:**
Dockerfile runs as root. Should use non-root user and distroless image.

**Reference:** TECHNICAL_AUDIT.md Section 6.4

---

## Summary Statistics

- **Total Issues:** 24
- **Critical Priority:** 4 issues
- **High Priority:** 6 issues
- **Medium Priority:** 8 issues
- **Low Priority:** 6 issues

## Implementation Recommendations

1. **Phase 1 (Critical):** Issues 1-4 (Security & Stability)
   - Expected duration: 1-2 weeks
   - Focus: Production readiness

2. **Phase 2 (High Priority):** Issues 5-10 (Operations & Performance)
   - Expected duration: 2-3 weeks
   - Focus: Observability and optimization

3. **Phase 3 (Medium Priority):** Issues 11-18 (Features & Refactoring)
   - Expected duration: 4-6 weeks
   - Focus: New features and code quality

4. **Phase 4 (Low Priority):** Issues 19-24 (Nice to Have)
   - Expected duration: Ongoing
   - Focus: Advanced features and polish

---

**Note:** These issues should be created in the GitHub repository with appropriate labels, milestones, and assignments. Each issue should link back to the relevant section in TECHNICAL_AUDIT.md for full context.
