.PHONY: build install uninstall clean test help

# Variables
BINARY_NAME=jotr
INSTALL_PATH=/usr/local/bin
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

# Default target
all: build

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	go build $(LDFLAGS) -o $(BINARY_NAME)
	@echo "✓ Build complete: ./$(BINARY_NAME)"

# Build for all platforms
build-all:
	@echo "Building for all platforms..."
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe
	@echo "✓ Built for all platforms in dist/"

# Install locally
install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_PATH)..."
	sudo cp $(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)
	sudo chmod +x $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "✓ Installed to $(INSTALL_PATH)/$(BINARY_NAME)"
	@echo ""
	@echo "Run 'jotr configure' to set up your configuration"

# Uninstall
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	sudo rm -f $(INSTALL_PATH)/$(BINARY_NAME)
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

# Development build (with debug info)
dev:
	@echo "Building development version..."
	go build -gcflags="all=-N -l" -o $(BINARY_NAME)
	@echo "✓ Development build complete"

# Quick test run
run:
	@go run main.go

# Help
help:
	@echo "jotr Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build       - Build the binary"
	@echo "  make install     - Build and install to $(INSTALL_PATH)"
	@echo "  make uninstall   - Remove installed binary"
	@echo "  make clean       - Remove build artifacts"
	@echo "  make test        - Run tests"
	@echo "  make fmt         - Format code"
	@echo "  make lint        - Lint code"
	@echo "  make build-all   - Build for all platforms"
	@echo "  make version     - Show version info"
	@echo "  make help        - Show this help"

