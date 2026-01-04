# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies for CGO and WebP
RUN apk add --no-cache gcc musl-dev libwebp-dev

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o grout ./cmd/grout

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache libwebp ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/grout .

# Expose port
EXPOSE 8080

# Set default environment variables
ENV ADDR=":8080"
ENV CACHE_SIZE="2000"

# Run the application
CMD ["./grout"]

