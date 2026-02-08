.PHONY: all build clean test test-npm test-python test-go install-deps lint release

# Version from go.mod or default
VERSION ?= 0.1.0
BINARY_NAME := clawdpl
GO_MODULE := github.com/moltyverse/clawdpl

# Build directories
DIST_DIR := dist
BIN_DIR := bin

# Go build flags
LDFLAGS := -ldflags "-s -w -X $(GO_MODULE)/cmd.Version=$(VERSION) -X $(GO_MODULE)/cmd.Commit=$(shell git rev-parse --short HEAD 2>/dev/null || echo 'dev')"

# Platforms for cross-compilation
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64

all: build

# Build the Go binary for current platform
build:
	go build $(LDFLAGS) -o $(BINARY_NAME) .

# Build for all platforms
build-all: clean
	@mkdir -p $(DIST_DIR)
	@for platform in $(PLATFORMS); do \
		GOOS=$${platform%/*} GOARCH=$${platform#*/} \
		EXT=$$([ "$${platform%/*}" = "windows" ] && echo ".exe" || echo "") \
		go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)_$(VERSION)_$${platform%/*}_$${platform#*/}/$(BINARY_NAME)$${EXT} . && \
		echo "Built for $${platform}"; \
	done

# Create release archives
release: build-all
	@cd $(DIST_DIR) && for dir in */; do \
		name=$${dir%/}; \
		tar -czf $${name}.tar.gz -C $${name} .; \
		echo "Created $${name}.tar.gz"; \
	done

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME) $(BINARY_NAME).exe
	rm -rf $(DIST_DIR) $(BIN_DIR)
	rm -rf npm/bin/$(BINARY_NAME) npm/bin/$(BINARY_NAME).exe
	rm -rf python/clawdpl/$(BINARY_NAME) python/clawdpl/$(BINARY_NAME).exe

# Run Go tests
test-go:
	go test -v ./...

# Test npm package locally
test-npm: build
	@echo "Testing npm package..."
	@cp $(BINARY_NAME) npm/bin/
	@cd npm && npm test
	@rm -f npm/bin/$(BINARY_NAME)
	@echo "npm package test passed!"

# Test Python package locally
test-python: build
	@echo "Testing Python package..."
	@cp $(BINARY_NAME) python/clawdpl/
	@cd python && python -m clawdpl.cli --version
	@rm -f python/clawdpl/$(BINARY_NAME)
	@echo "Python package test passed!"

# Test npx execution (requires npm link or local install)
test-npx: build
	@echo "Testing npx execution..."
	@cp $(BINARY_NAME) npm/bin/
	@cd npm && npm pack && npx ./clawdpl-$(VERSION).tgz --version
	@rm -f npm/clawdpl-$(VERSION).tgz npm/bin/$(BINARY_NAME)
	@echo "npx test passed!"

# Test pipx execution
test-pipx: build
	@echo "Testing pipx execution..."
	@cp $(BINARY_NAME) python/clawdpl/
	@cd python && pipx run --spec . clawdpl --version
	@rm -f python/clawdpl/$(BINARY_NAME)
	@echo "pipx test passed!"

# Run all tests
test: test-go test-npm test-python
	@echo "All tests passed!"

# Install development dependencies
install-deps:
	go mod download
	@echo "Go dependencies installed"

# Lint Go code
lint:
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not installed"; exit 1; }
	golangci-lint run

# Format Go code
fmt:
	go fmt ./...

# Development: build and copy to wrappers for local testing
dev: build
	@cp $(BINARY_NAME) npm/bin/
	@cp $(BINARY_NAME) python/clawdpl/
	@echo "Binary copied to npm/bin/ and python/clawdpl/ for local testing"

# Clean development binaries from wrappers
dev-clean:
	@rm -f npm/bin/$(BINARY_NAME) npm/bin/$(BINARY_NAME).exe
	@rm -f python/clawdpl/$(BINARY_NAME) python/clawdpl/$(BINARY_NAME).exe
	@echo "Development binaries cleaned from wrappers"

# Show help
help:
	@echo "Available targets:"
	@echo "  build       - Build the Go binary for current platform"
	@echo "  build-all   - Build for all supported platforms"
	@echo "  release     - Create release archives for all platforms"
	@echo "  clean       - Remove build artifacts"
	@echo "  test        - Run all tests"
	@echo "  test-go     - Run Go tests"
	@echo "  test-npm    - Test npm package locally"
	@echo "  test-python - Test Python package locally"
	@echo "  test-npx    - Test npx execution"
	@echo "  test-pipx   - Test pipx execution"
	@echo "  dev         - Build and copy binary to wrappers for local testing"
	@echo "  dev-clean   - Remove development binaries from wrappers"
	@echo "  lint        - Run linter"
	@echo "  fmt         - Format Go code"
	@echo "  help        - Show this help"
