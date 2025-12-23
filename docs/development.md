# Development Guide

This guide will help you set up your development environment and contribute to RCIG.

## Prerequisites

### Required
- **Go**: Version 1.22 or later
  - Download from [golang.org](https://golang.org/dl/)
  - Verify: `go version`

- **Git**: For version control
  - Verify: `git --version`

### Recommended
- **Make**: For build automation
  - Linux/macOS: Usually pre-installed
  - Windows: Install via [Chocolatey](https://chocolatey.org/): `choco install make`

- **Docker**: For containerized development
  - Install from [docker.com](https://www.docker.com/get-started)
  - Verify: `docker --version`

- **Docker Compose**: For multi-container setup
  - Usually included with Docker Desktop
  - Verify: `docker-compose --version`

## Getting Started

### 1. Clone the Repository

```bash
git clone https://github.com/tahsinmert/go-traffic-proxy-analyzer.git
cd go-traffic-proxy-analyzer
```

### 2. Install Development Tools

```bash
make install-tools
```

This installs:
- `golangci-lint`: Comprehensive Go linter
- `goimports`: Import formatter
- `gosec`: Security checker

### 3. Install Dependencies

```bash
go mod download
go mod verify
```

### 4. Build the Project

```bash
make build
```

Binary will be created at `bin/rcig`.

## Development Workflow

### Running Locally

#### Option 1: Go Run (Quick Testing)

```bash
# Set environment variables
export RCIG_TARGET_URL=http://localhost:9000
export RCIG_LISTEN_ADDR=:8080

# Run
make run
# or
go run ./cmd/proxy
```

#### Option 2: Build and Run

```bash
make build
export RCIG_TARGET_URL=http://localhost:9000
./bin/rcig
```

#### Option 3: Docker Compose (Full Stack)

```bash
# Start all services (RCIG + example backend)
make docker-compose-up

# Stop all services
make docker-compose-down
```

### Setting Up a Test Backend

For testing, you need a backend service. Options:

#### Quick Test Server (Python)

```bash
# In a separate terminal
cd /tmp
python3 -m http.server 9000
```

#### nginx (via Docker)

```bash
docker run -d -p 9000:80 nginx:alpine
```

#### Custom Service

Point `RCIG_TARGET_URL` to any HTTP service you want to proxy.

## Testing

### Run All Tests

```bash
make test
```

### Run Tests with Coverage

```bash
make coverage
```

This generates:
- `coverage.out`: Coverage data
- `coverage.html`: HTML coverage report (open in browser)

### Run Specific Package Tests

```bash
go test -v ./internal/metrics/
go test -v ./internal/analyzer/
go test -v ./internal/alert/
go test -v ./internal/proxy/
```

### Run Tests with Race Detector

```bash
go test -race ./...
```

### Run Benchmarks

```bash
go test -bench=. -benchmem ./...
```

### Benchmark Specific Functions

```bash
go test -bench=BenchmarkRegistry_Record -benchmem ./internal/metrics/
```

## Code Quality

### Formatting

```bash
make fmt
```

This runs:
- `gofmt`: Standard Go formatter
- `goimports`: Import organizer

### Linting

```bash
make lint
```

This runs `golangci-lint` with the configuration in `.golangci.yml`.

### Vet

```bash
make vet
```

Runs `go vet` to catch suspicious constructs.

### Security Scanning

```bash
make security
```

Runs `gosec` to find security issues.

### All Quality Checks

```bash
make all
```

Runs: clean, fmt, vet, lint, test, build

## Project Structure

```
.
├── cmd/
│   └── proxy/
│       └── main.go           # Application entry point
├── internal/
│   ├── alert/
│   │   ├── alert.go          # Alert interfaces and implementations
│   │   └── alert_test.go     # Alert tests
│   ├── analyzer/
│   │   ├── analyzer.go       # Drift detection logic
│   │   └── analyzer_test.go  # Analyzer tests
│   ├── metrics/
│   │   ├── metrics.go        # Metrics registry and circular buffer
│   │   ├── http.go           # HTTP handler for metrics
│   │   ├── metrics_test.go   # Metrics tests
│   │   └── http_test.go      # HTTP handler tests
│   └── proxy/
│       ├── handler.go        # Reverse proxy handler
│       └── handler_test.go   # Proxy tests
├── config/
│   ├── rcig.default.yml      # Default configuration
│   ├── rcig.example.yml      # Example configuration with all options
│   └── prometheus.yml        # Prometheus scrape config
├── docs/
│   ├── architecture.md       # Architecture documentation
│   ├── development.md        # This file
│   ├── deployment.md         # Deployment guide
│   ├── configuration.md      # Configuration reference
│   └── api.md                # API documentation
├── .github/
│   ├── workflows/            # CI/CD workflows
│   ├── ISSUE_TEMPLATE/       # Issue templates
│   ├── pull_request_template.md
│   ├── CODEOWNERS            # Code owners
│   └── dependabot.yml        # Dependency updates
├── Dockerfile                # Multi-stage Docker build
├── docker-compose.yml        # Local development setup
├── Makefile                  # Build automation
├── go.mod                    # Go module definition
├── go.sum                    # Go module checksums
├── README.md                 # Project README
├── CHANGELOG.md              # Version history
├── CONTRIBUTING.md           # Contribution guidelines
├── LICENSE                   # MIT License
└── SECURITY.md               # Security policy
```

## Making Changes

### 1. Create a Feature Branch

```bash
git checkout -b feature/my-new-feature
```

### 2. Make Your Changes

- Write code following Go best practices
- Add tests for new functionality
- Update documentation if needed
- Ensure all tests pass
- Run linters and fix warnings

### 3. Commit Your Changes

```bash
git add .
git commit -m "Add my new feature"
```

Use conventional commit messages:
- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation changes
- `test:` Adding tests
- `refactor:` Code refactoring
- `chore:` Maintenance tasks

### 4. Push and Create Pull Request

```bash
git push origin feature/my-new-feature
```

Then create a pull request on GitHub.

## Writing Tests

### Test Structure

Use table-driven tests:

```go
func TestMyFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    int
        expected int
    }{
        {"zero", 0, 0},
        {"positive", 5, 10},
        {"negative", -5, -10},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := MyFunction(tt.input)
            if result != tt.expected {
                t.Errorf("got %d, want %d", result, tt.expected)
            }
        })
    }
}
```

### Test Helpers

Create helper functions for common setup:

```go
func setupTestRegistry(t *testing.T) *metrics.Registry {
    t.Helper()
    return metrics.NewRegistry(100)
}
```

### Test Coverage Goals

- Aim for 80%+ coverage
- Focus on testing logic, not trivial code
- Test error paths
- Test concurrent access

## Debugging

### Debug Logging

Add temporary debug prints:

```go
import "log"

log.Printf("Debug: value=%v", someValue)
```

### Using Delve Debugger

Install Delve:

```bash
go install github.com/go-delve/delve/cmd/dlv@latest
```

Debug the application:

```bash
dlv debug ./cmd/proxy -- --target-url http://localhost:9000
```

Set breakpoints:

```
(dlv) break main.main
(dlv) continue
(dlv) print cfg
```

### Profiling

CPU profiling:

```bash
go test -cpuprofile=cpu.prof -bench=. ./internal/metrics/
go tool pprof cpu.prof
```

Memory profiling:

```bash
go test -memprofile=mem.prof -bench=. ./internal/metrics/
go tool pprof mem.prof
```

## Docker Development

### Build Docker Image

```bash
make docker-build
```

### Run in Docker

```bash
make docker-run
```

### Multi-Container Setup

```bash
docker-compose up -d
```

View logs:

```bash
docker-compose logs -f rcig
```

Stop services:

```bash
docker-compose down
```

## Common Tasks

### Adding a New Package

1. Create directory under `internal/` or `pkg/`
2. Create `.go` files
3. Add package documentation
4. Create `_test.go` files
5. Update `go.mod` if needed
6. Add to documentation

### Adding a New Dependency

```bash
go get github.com/some/package@v1.2.3
go mod tidy
```

### Updating Dependencies

```bash
go get -u ./...
go mod tidy
```

### Generating Mocks (if using gomock)

```bash
mockgen -source=internal/alert/alert.go -destination=internal/alert/mock_alert.go -package=alert
```

## CI/CD

### GitHub Actions Workflows

- **test.yml**: Runs tests on push/PR
- **lint.yml**: Runs linters
- **build.yml**: Builds binaries for multiple platforms
- **release.yml**: Creates releases on tags
- **security.yml**: Runs security scans

### Running CI Locally

You can run similar checks locally:

```bash
# What CI runs
make fmt
make vet
make lint
make test
make build
```

## Best Practices

### Code Style

- Follow [Effective Go](https://golang.org/doc/effective_go)
- Use `gofmt` for formatting
- Keep functions small and focused
- Write self-documenting code
- Add comments for public APIs

### Error Handling

```go
// Good: Wrap errors with context
if err != nil {
    return fmt.Errorf("failed to parse config: %w", err)
}

// Bad: Lose context
if err != nil {
    return err
}
```

### Concurrency

- Use mutexes for shared state
- Prefer channels for communication
- Document goroutine lifecycles
- Always handle goroutine panics

### Testing

- Write tests first (TDD)
- Test one thing per test
- Use descriptive test names
- Mock external dependencies
- Test error cases

## Getting Help

- **Documentation**: Check `docs/` directory
- **Issues**: Search existing GitHub issues
- **Discussions**: Start a GitHub discussion
- **Code**: Read the source code (it's well-commented!)

## Contributing

See [CONTRIBUTING.md](../CONTRIBUTING.md) for detailed contribution guidelines.

## Resources

### Go Resources
- [Go Documentation](https://golang.org/doc/)
- [Effective Go](https://golang.org/doc/effective_go)
- [Go by Example](https://gobyexample.com/)

### Project Resources
- [Architecture](./architecture.md)
- [Configuration](./configuration.md)
- [API Reference](./api.md)
- [Deployment](./deployment.md)

### Tools
- [golangci-lint](https://golangci-lint.run/)
- [Delve Debugger](https://github.com/go-delve/delve)
- [pprof](https://golang.org/pkg/net/http/pprof/)
