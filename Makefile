.PHONY: all build build-debug clean test test-npm test-python test-go install-deps lint release

# Version from environment or default
VERSION ?= dev
BINARY_NAME := clawdepl
GO_MODULE := github.com/clawdepl/clawdepl

# Build directories
DIST_DIR := dist
BIN_DIR := bin

# Build metadata
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
COMMIT_FULL := $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Endpoint URLs (configurable via environment for different environments)
CONVEX_ENDPOINT ?= https://colorless-gull-839.convex.site
PROVISIONER_ENDPOINT ?= https://clawdepl-provisioner-production.up.railway.app
AUTH_ENDPOINT ?= https://clawdepl-auth.vercel.app

# ldflags for build-time injection
LDFLAGS := -ldflags "-s -w \
	-X $(GO_MODULE)/internal/buildinfo.Version=$(VERSION) \
	-X $(GO_MODULE)/internal/buildinfo.Commit=$(COMMIT) \
	-X $(GO_MODULE)/internal/buildinfo.CommitFull=$(COMMIT_FULL) \
	-X $(GO_MODULE)/internal/buildinfo.Date=$(DATE) \
	-X $(GO_MODULE)/internal/buildinfo.ConvexEndpoint=$(CONVEX_ENDPOINT) \
	-X $(GO_MODULE)/internal/buildinfo.ProvisionerEndpoint=$(PROVISIONER_ENDPOINT) \
	-X $(GO_MODULE)/internal/buildinfo.AuthEndpoint=$(AUTH_ENDPOINT)"

# Platforms for cross-compilation
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64

all: build

# Build the Go binary for current platform (production)
build:
	go build $(LDFLAGS) -o $(BINARY_NAME) .

# Build with debug flags enabled (includes --unsafe-endpoint and other debug flags)
build-debug:
	go build -tags debug $(LDFLAGS) -o $(BINARY_NAME) .

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
	rm -rf python/clawdepl/$(BINARY_NAME) python/clawdepl/$(BINARY_NAME).exe

# Run Go tests
test-go:
	go test -v ./...

# Run Go tests with debug build
test-go-debug:
	go test -tags debug -v ./...

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
	@cp $(BINARY_NAME) python/clawdepl/
	@cd python && python -m clawdepl.cli --version
	@rm -f python/clawdepl/$(BINARY_NAME)
	@echo "Python package test passed!"

# Test npx execution (requires npm link or local install)
test-npx: build
	@echo "Testing npx execution..."
	@cp $(BINARY_NAME) npm/bin/
	@cd npm && npm pack && npx ./clawdepl-$(VERSION).tgz --version
	@rm -f npm/clawdepl-$(VERSION).tgz npm/bin/$(BINARY_NAME)
	@echo "npx test passed!"

# Test pipx execution
test-pipx: build
	@echo "Testing pipx execution..."
	@cp $(BINARY_NAME) python/clawdepl/
	@cd python && pipx run --spec . clawdepl --version
	@rm -f python/clawdepl/$(BINARY_NAME)
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

# Development: build and copy to wrappers for local testing (production build)
dev: build
	@cp $(BINARY_NAME) npm/bin/
	@cp $(BINARY_NAME) python/clawdepl/
	@echo "Binary copied to npm/bin/ and python/clawdepl/ for local testing"

# Development: build with debug flags and copy to wrappers
dev-debug: build-debug
	@cp $(BINARY_NAME) npm/bin/
	@cp $(BINARY_NAME) python/clawdepl/
	@echo "Debug binary copied to npm/bin/ and python/clawdepl/ for local testing"

# Clean development binaries from wrappers
dev-clean:
	@rm -f npm/bin/$(BINARY_NAME) npm/bin/$(BINARY_NAME).exe
	@rm -f python/clawdepl/$(BINARY_NAME) python/clawdepl/$(BINARY_NAME).exe
	@echo "Development binaries cleaned from wrappers"

# Show help
help:
	@echo "Available targets:"
	@echo "  build        - Build the Go binary for current platform (production)"
	@echo "  build-debug  - Build with debug flags enabled (includes --unsafe-endpoint)"
	@echo "  build-all    - Build for all supported platforms"
	@echo "  release      - Create release archives for all platforms"
	@echo "  clean        - Remove build artifacts"
	@echo "  test         - Run all tests"
	@echo "  test-go      - Run Go tests (production)"
	@echo "  test-go-debug- Run Go tests (debug build)"
	@echo "  test-npm     - Test npm package locally"
	@echo "  test-python  - Test Python package locally"
	@echo "  test-npx     - Test npx execution"
	@echo "  test-pipx    - Test pipx execution"
	@echo "  dev          - Build and copy binary to wrappers for local testing"
	@echo "  dev-debug    - Build debug binary and copy to wrappers"
	@echo "  dev-clean    - Remove development binaries from wrappers"
	@echo "  lint         - Run linter"
	@echo "  fmt          - Format Go code"
	@echo "  help         - Show this help"
