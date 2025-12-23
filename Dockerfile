# Multi-stage Dockerfile for RCIG
# Stage 1: Build stage
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod ./

# Download dependencies (none currently, but support for future)
RUN go mod download || true

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.Version=docker -X main.BuildTime=$(date -u '+%Y-%m-%d_%H:%M:%S')" \
    -o rcig ./cmd/proxy

# Stage 2: Runtime stage
FROM alpine:latest

# Install CA certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1000 rcig && \
    adduser -D -u 1000 -G rcig rcig

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/rcig /app/rcig

# Copy config examples (optional)
COPY --from=builder /build/config /app/config

# Change ownership
RUN chown -R rcig:rcig /app

# Switch to non-root user
USER rcig

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/rcig/healthz || exit 1

# Run the application
ENTRYPOINT ["/app/rcig"]
