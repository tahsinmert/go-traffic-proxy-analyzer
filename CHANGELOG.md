# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Comprehensive test suite with unit tests for all packages
- Integration tests for full proxy flow
- Benchmarks for critical paths
- Structured logging using zerolog
- Context propagation for request tracing
- Input validation for all configuration values
- Prometheus metrics exporter at `/metrics` endpoint
- Multiple alert sinks (Slack, Email, Webhook)
- API key authentication for admin endpoints
- Rate limiting per endpoint
- Circuit breaker pattern for backend protection
- Retry logic with exponential backoff
- Request/response body inspection (configurable)
- Custom header manipulation
- Request timeout configuration per endpoint
- Response caching with TTL
- YAML configuration file support
- Configuration hot-reload for non-critical settings
- OpenTelemetry distributed tracing integration
- Enhanced logging with request IDs and correlation IDs
- Percentile latencies (p50, p95, p99)
- Response size tracking
- Active connections gauge
- Backend health metrics
- Multi-stage Dockerfile for optimal image size
- Docker Compose setup for local development
- GitHub Actions CI/CD workflows (test, lint, build, release, security)
- Makefile with standard development targets
- .editorconfig for consistent code formatting
- .golangci.yml for comprehensive linting
- Dependabot configuration for dependency updates
- HTTP/2 support
- gzip/brotli compression support
- pprof profiling endpoints
- TLS/HTTPS support with configurable certificates
- Security headers (HSTS, CSP, X-Frame-Options, etc.)
- CORS configuration options
- Request size limits
- OpenAPI/Swagger specification
- Architecture documentation with diagrams
- Deployment guide
- Configuration reference
- Development setup guide
- API reference documentation
- Example configurations
- Issue and PR templates
- CODEOWNERS file

### Changed
- Replaced standard `log` package with structured logging
- Improved error handling with proper error wrapping
- Enhanced code organization with interfaces and dependency injection
- Optimized connection pooling for backend connections
- Improved metrics storage with efficient data structures
- Updated README with badges, quick start, and comprehensive documentation
- Translated all Turkish strings to English

### Fixed
- Various error handling improvements
- Memory leaks in metrics collection
- Race conditions in analyzer

### Security
- Added security scanning with gosec
- Added dependency scanning with Trivy
- Implemented request sanitization to prevent injection attacks
- Added secrets management documentation

## [0.1.0] - Initial Release

### Added
- Basic reverse proxy functionality
- Real-time HTTP metrics collection
- Health check endpoint
- Drift/anomaly detection
- Console alert sink
- Configuration via environment variables

[Unreleased]: https://github.com/tahsinmert/go-traffic-proxy-analyzer/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/tahsinmert/go-traffic-proxy-analyzer/releases/tag/v0.1.0
