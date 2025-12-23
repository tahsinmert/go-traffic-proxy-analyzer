.PHONY: help build test lint clean install-tools docker-build docker-run run coverage

# Variables
BINARY_NAME=rcig
MAIN_PATH=./cmd/proxy
BUILD_DIR=./bin
DOCKER_IMAGE=rcig:latest
COVERAGE_FILE=coverage.out

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GOVET=$(GOCMD) vet
GOFMT=gofmt

# Build flags
VERSION?=dev
COMMIT?=$(shell git rev-parse --short HEAD)
BUILD_TIME?=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildTime=$(BUILD_TIME)"

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

run: ## Run the application
	$(GOCMD) run $(MAIN_PATH)

test: ## Run tests
	@echo "Running tests..."
	$(GOTEST) -v -race -timeout 30s ./...

coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	$(GOTEST) -v -race -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	@echo "Coverage report generated: $(COVERAGE_FILE)"
	$(GOCMD) tool cover -html=$(COVERAGE_FILE) -o coverage.html
	@echo "HTML coverage report: coverage.html"

lint: ## Run linters
	@echo "Running linters..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not found, run 'make install-tools' first" && exit 1)
	golangci-lint run --timeout 5m ./...

fmt: ## Format code
	@echo "Formatting code..."
	$(GOFMT) -s -w .
	@which goimports > /dev/null && goimports -w . || echo "goimports not found, skipping"

vet: ## Run go vet
	@echo "Running go vet..."
	$(GOVET) ./...

tidy: ## Tidy go modules
	@echo "Tidying go modules..."
	$(GOMOD) tidy

install-tools: ## Install development tools
	@echo "Installing development tools..."
	@which golangci-lint > /dev/null || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@which goimports > /dev/null || go install golang.org/x/tools/cmd/goimports@latest
	@which gosec > /dev/null || go install github.com/securego/gosec/v2/cmd/gosec@latest
	@echo "Tools installed successfully"

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE) .
	@echo "Docker image built: $(DOCKER_IMAGE)"

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker run --rm -p 8080:8080 \
		-e RCIG_TARGET_URL=http://host.docker.internal:9000 \
		$(DOCKER_IMAGE)

docker-compose-up: ## Start docker-compose environment
	docker-compose up -d

docker-compose-down: ## Stop docker-compose environment
	docker-compose down

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -f $(COVERAGE_FILE) coverage.html
	@rm -f *.test
	@rm -f *.out
	@echo "Clean complete"

security: ## Run security checks
	@echo "Running security checks..."
	@which gosec > /dev/null || (echo "gosec not found, run 'make install-tools' first" && exit 1)
	gosec -fmt=json -out=gosec-report.json ./...
	@echo "Security report generated: gosec-report.json"

all: clean fmt vet lint test build ## Run all checks and build

.DEFAULT_GOAL := help
