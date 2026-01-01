# OtherSide Paranormal Investigation Application
# Makefile for building and running the application

.PHONY: build run test clean deps lint help

# Build configuration
BINARY_NAME=otherside-server
BUILD_DIR=bin
MAIN_PATH=./cmd/server
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

# Default target
all: deps lint test build

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Build the application
build:
	@echo "Building OtherSide server..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Run the application
run:
	@echo "Starting OtherSide server..."
	go run $(MAIN_PATH)

# Run on custom port
run-port:
	@echo "Starting OtherSide server on port $(PORT)..."
	PORT=$(PORT) go run $(MAIN_PATH)

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run tests with race detection
test-race:
	@echo "Running tests with race detection..."
	go test -v -race ./...

# Run tests with coverage and race detection
test-coverage-race:
	@echo "Running tests with coverage and race detection..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run specific test package
test-pkg:
	@if [ -z "$(PKG)" ]; then \
		echo "Usage: make test-pkg PKG=package/name"; \
		exit 1; \
	fi
	@echo "Running tests for package $(PKG)..."
	go test -v ./$(PKG)

# Run tests with verbose output and short mode
test-quick:
	@echo "Running quick tests..."
	go test -v -short ./...

# Generate test coverage report (text)
test-coverage-text:
	@echo "Running tests with coverage (text report)..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	@echo "Coverage summary shown above"

# Run benchmarks
test-bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem -run=^$$ ./...

# Run integration tests (requires test database)
test-integration:
	@echo "Running integration tests..."
	go test -v -tags=integration ./...

# Run unit tests only
test-unit:
	@echo "Running unit tests..."
	go test -v -short ./... | grep -v integration

# Clean test artifacts
test-clean:
	@echo "Cleaning test artifacts..."
	rm -f coverage.out coverage.html
	rm -f test.log
	go clean -testcache

# Test with coverage and threshold checking (requires 80% minimum)
test-coverage-check:
	@echo "Running tests with coverage threshold check..."
	go test -v -coverprofile=coverage.out ./...
	@COVERAGE=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	if [ $$(echo "$$COVERAGE >= 80" | bc -l) -eq 1 ]; then \
		echo "✅ Coverage $$COVERAGE% meets minimum requirement (80%)"; \
	else \
		echo "❌ Coverage $$COVERAGE% below minimum requirement (80%)"; \
		exit 1; \
	fi

# Generate HTML coverage report and open in browser
test-coverage-html:
	@echo "Running tests with HTML coverage report..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
	@if command -v xdg-open >/dev/null 2>&1; then \
		xdg-open coverage.html; \
	elif command -v open >/dev/null 2>&1; then \
		open coverage.html; \
	else \
		echo "Open coverage.html in your browser to view the report"; \
	fi

# Lint code
lint:
	@echo "Running linters..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, running basic checks..."; \
		go vet ./...; \
		go fmt ./...; \
	fi

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	rm -rf data/otherside.db*
	rm -rf data/exports/*

# Set up development environment
dev-setup:
	@echo "Setting up development environment..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/air-verse/air@latest
	@echo "Development tools installed"

# Watch and reload during development
dev:
	@echo "Starting development server with hot reload..."
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "Air not installed, running normally..."; \
		$(MAKE) run; \
	fi

# Create data directories
init-dirs:
	@echo "Creating data directories..."
	mkdir -p data/sessions
	mkdir -p data/exports
	mkdir -p data/audio
	mkdir -p data/video
	chmod 755 data data/sessions data/exports data/audio data/video

# Database operations
db-reset:
	@echo "Resetting database..."
	rm -f data/otherside.db*
	@echo "Database reset complete"

db-backup:
	@echo "Backing up database..."
	@if [ -f data/otherside.db ]; then \
		cp data/otherside.db data/otherside-backup-$(shell date +%Y%m%d_%H%M%S).db; \
		echo "Database backed up"; \
	else \
		echo "No database found to backup"; \
	fi

# Docker operations (future)
docker-build:
	@echo "Building Docker image..."
	docker build -t otherside:$(VERSION) .

docker-run:
	@echo "Running in Docker..."
	docker run -p 8080:8080 -v $(PWD)/data:/app/data otherside:$(VERSION)

# Production deployment helpers
deploy-check:
	@echo "Running deployment checks..."
	$(MAKE) deps
	$(MAKE) lint
	$(MAKE) test
	$(MAKE) build
	@echo "Deployment checks passed ✅"

# Generate API documentation (future)
docs:
	@echo "Generating API documentation..."
	@echo "API documentation generation not yet implemented"

# Performance benchmarks
bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# Security scan
security:
	@echo "Running security checks..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec not installed, skipping security scan"; \
		echo "Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi

# Help
help:
	@echo "OtherSide Paranormal Investigation Application"
	@echo ""
	@echo "Available targets:"
	@echo "  build                 Build the application binary"
	@echo "  run                   Run the application"
	@echo "  run-port              Run on custom port (use PORT=8081 make run-port)"
	@echo ""
	@echo "Testing:"
	@echo "  test                  Run all tests"
	@echo "  test-coverage         Run tests with coverage report"
	@echo "  test-race             Run tests with race detection"
	@echo "  test-coverage-race    Run tests with coverage and race detection"
	@echo "  test-pkg PKG=name     Run tests for specific package"
	@echo "  test-quick            Run quick tests (short mode)"
	@echo "  test-coverage-text    Generate coverage text report"
	@echo "  test-bench            Run benchmarks"
	@echo "  test-integration      Run integration tests"
	@echo "  test-unit             Run unit tests only"
	@echo "  test-clean            Clean test artifacts"
	@echo "  test-coverage-check   Check coverage meets 80% threshold"
	@echo "  test-coverage-html    Generate HTML coverage and open browser"
	@echo ""
	@echo "Development:"
	@echo "  lint                  Run code linters"
	@echo "  clean                 Clean build artifacts and data"
	@echo "  deps                  Install/update dependencies"
	@echo "  dev-setup             Install development tools"
	@echo "  dev                   Run with hot reload (requires air)"
	@echo "  init-dirs             Create required data directories"
	@echo ""
	@echo "Database:"
	@echo "  db-reset              Reset database"
	@echo "  db-backup             Backup database"
	@echo ""
	@echo "Deployment:"
	@echo "  deploy-check          Run all checks for deployment"
	@echo "  bench                 Run performance benchmarks"
	@echo "  security              Run security scans"
	@echo ""
	@echo "  help                  Show this help message"
	@echo ""
	@echo "Examples:"
	@echo "  make run                    # Start server on default port 8080"
	@echo "  PORT=3000 make run-port     # Start server on port 3000"
	@echo "  make test-coverage          # Run tests and generate coverage report"
	@echo "  make test-pkg PKG=internal/domain   # Test specific package"
	@echo "  make deploy-check           # Run all pre-deployment checks"