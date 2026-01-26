.PHONY: build install uninstall clean test help

# Variables
BINARY_NAME=jotr
INSTALL_PATH=$(HOME)/.local/bin
YEAR_IN_DEV=$(shell expr $(shell date +%Y) - 2025)
MONTH=$(shell date +%-m)
VERSION=$(YEAR_IN_DEV).$(MONTH).0
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X github.com/AnishShah1803/jotr/internal/version.Version=$(VERSION) -X github.com/AnishShah1803/jotr/internal/version.BuildTime=$(BUILD_TIME)"

# Default target
all: build

# Build for all platforms
build-all:
	@echo "Building for all platforms..."
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe
	@echo "✓ Built for all platforms in dist/"

# Install
install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_PATH)/..."
	@mkdir -p $(INSTALL_PATH)
	@cp $(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)
	@chmod +x $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "✓ Installed to $(INSTALL_PATH)/$(BINARY_NAME)"
	@echo ""
	@echo "Run 'jotr configure' to set up your configuration"

# Uninstall
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	@rm -f $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "✓ Uninstalled"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	rm -rf dist/
	@echo "✓ Cleaned"

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run with race detector
test-race:
	@echo "Running tests with race detector..."
	go test -race -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -cover ./...

# Generate coverage report
coverage:
	@echo "Generating coverage report..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run tests with coverage (profile)
test-coverage-profile:
	@echo "Running tests with coverage profile..."
	go test -coverprofile=coverage.out -covermode=count ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage profile generated: coverage.html"

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	@echo "✓ Formatted"

# Lint code
lint:
	@echo "Linting code..."
	golangci-lint run || echo "Install golangci-lint: brew install golangci-lint"

# Show version
version:
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"

# Development build (with debug info and dev features)
dev:
	@echo "Building development version..."
	go build -tags=dev -gcflags="all=-N -l" -o $(BINARY_NAME)-dev
	@echo "✓ Development build complete"

# Production build (without dev features)
build:
	@echo "Building $(BINARY_NAME)..."
	go build $(LDFLAGS) -o $(BINARY_NAME)
	@echo "✓ Build complete: ./$(BINARY_NAME)"

# Quick test run
run:
	@go run main.go

# Help
help:
	@echo "jotr Makefile"
	@echo ""
	@echo "Installation:"
	@echo "  make install     - Build and install to $(INSTALL_PATH)"
	@echo "  make uninstall   - Remove installed binary"
	@echo ""
	@echo "Development:"
	@echo "  make build       - Build production binary"
	@echo "  make dev         - Build development binary (includes dev mode)"
	@echo "  make clean       - Remove build artifacts"
	@echo "  make test        - Run tests"
	@echo "  make test-race  - Run tests with race detector"
	@echo "  make test-coverage - Run tests with coverage"
	@echo "  make coverage     - Generate coverage report"
	@echo "  make fmt          - Format code"
	@echo "  make lint        - Lint code"
	@echo "  make build-all   - Build for all platforms"
	@echo "  make version     - Show version info"
	@echo "  make help        - Show this help"

