# Makefile for meta-cc
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -ldflags "-X github.com/yaleh/meta-cc/cmd.Version=$(VERSION) \
                     -X github.com/yaleh/meta-cc/cmd.Commit=$(COMMIT) \
                     -X github.com/yaleh/meta-cc/cmd.BuildTime=$(BUILD_TIME)"

GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOCLEAN := $(GOCMD) clean
GOMOD := $(GOCMD) mod
BUILD_DIR := build
BINARY_NAME := meta-cc
MCP_BINARY_NAME := meta-cc-mcp
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

.PHONY: all build build-cli build-mcp test test-all test-coverage clean install cross-compile bundle-release lint fmt vet help sync-plugin-files dev

all: lint test build

build: build-cli build-mcp

build-cli:
	@echo "Building $(BINARY_NAME) $(VERSION)..."
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) .

build-mcp:
	@echo "Building $(MCP_BINARY_NAME) $(VERSION)..."
	$(GOBUILD) -o $(MCP_BINARY_NAME) ./cmd/mcp-server

test:
	@echo "Running tests (short mode, skips slow E2E tests)..."
	$(GOTEST) -short -v ./...

test-all:
	@echo "Running all tests (including slow E2E tests ~30s)..."
	$(GOTEST) -v ./...

test-coverage:
	@echo "Running tests with coverage (including E2E tests)..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(MCP_BINARY_NAME)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	rm -rf commands agents  # Clean synced build artifacts

install:
	@echo "Installing..."
	$(GOCMD) install $(LDFLAGS) .

cross-compile:
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	@for platform in $(PLATFORMS); do \
		GOOS=$${platform%/*} GOARCH=$${platform#*/} \
		$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-$${platform%/*}-$${platform#*/} .; \
		if [ "$${platform%/*}" = "windows" ]; then \
			mv $(BUILD_DIR)/$(BINARY_NAME)-$${platform%/*}-$${platform#*/} $(BUILD_DIR)/$(BINARY_NAME)-$${platform%/*}-$${platform#*/}.exe; \
		fi; \
	done
	@echo "Cross-compilation complete. Binaries in $(BUILD_DIR)/"

sync-plugin-files:
	@echo "Syncing plugin files from .claude/ to root..."
	@bash scripts/sync-plugin-files.sh

dev: build
	@echo "Development build complete"
	@echo "Plugin files in .claude/ are ready for immediate use in Claude Code"

bundle-release: sync-plugin-files
	@echo "Creating release bundles for all platforms..."
	@if [ -z "$(VERSION)" ] || [ "$(VERSION)" = "dev" ]; then \
		echo "ERROR: VERSION must be set (e.g., make bundle-release VERSION=v1.0.0)"; \
		exit 1; \
	fi
	@mkdir -p $(BUILD_DIR)/bundles
	@for platform in $(PLATFORMS); do \
		PLATFORM_NAME=$${platform%/*}-$${platform#*/}; \
		BUNDLE_DIR=$(BUILD_DIR)/bundles/meta-cc-$(VERSION)-$$PLATFORM_NAME; \
		mkdir -p $$BUNDLE_DIR/bin $$BUNDLE_DIR/commands $$BUNDLE_DIR/agents $$BUNDLE_DIR/.claude-plugin $$BUNDLE_DIR/lib; \
		if [ "$${platform%/*}" = "windows" ]; then \
			cp $(BUILD_DIR)/$(BINARY_NAME)-$$PLATFORM_NAME.exe $$BUNDLE_DIR/bin/ 2>/dev/null || true; \
			cp $(BUILD_DIR)/$(MCP_BINARY_NAME)-$$PLATFORM_NAME.exe $$BUNDLE_DIR/bin/ 2>/dev/null || true; \
		else \
			cp $(BUILD_DIR)/$(BINARY_NAME)-$$PLATFORM_NAME $$BUNDLE_DIR/bin/ 2>/dev/null || true; \
			cp $(BUILD_DIR)/$(MCP_BINARY_NAME)-$$PLATFORM_NAME $$BUNDLE_DIR/bin/ 2>/dev/null || true; \
		fi; \
		cp -r commands/* $$BUNDLE_DIR/commands/; \
		cp -r agents/* $$BUNDLE_DIR/agents/; \
		cp -r lib/* $$BUNDLE_DIR/lib/; \
		cp -r .claude-plugin/* $$BUNDLE_DIR/.claude-plugin/; \
		cp scripts/install.sh $$BUNDLE_DIR/; \
		cp scripts/uninstall.sh $$BUNDLE_DIR/ 2>/dev/null || true; \
		cp README.md $$BUNDLE_DIR/; \
		cp LICENSE $$BUNDLE_DIR/; \
		tar -czf $(BUILD_DIR)/meta-cc-bundle-$$PLATFORM_NAME.tar.gz -C $(BUILD_DIR)/bundles meta-cc-$(VERSION)-$$PLATFORM_NAME; \
	done
	@echo "Bundle creation complete. Archives in $(BUILD_DIR)/"

deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

lint: fmt vet
	@echo "Running static analysis..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not found. Install with:"; \
		echo "  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		echo "Skipping lint checks..."; \
	fi

fmt:
	@echo "Formatting code..."
	@gofmt -l -w .

vet:
	@echo "Running go vet..."
	@$(GOCMD) vet ./...

help:
	@echo "Available targets:"
	@echo "  make build             - Build both meta-cc and meta-cc-mcp"
	@echo "  make build-cli         - Build meta-cc CLI only"
	@echo "  make build-mcp         - Build meta-cc-mcp MCP server only"
	@echo "  make dev               - Development build (use .claude/ for immediate testing)"
	@echo "  make test              - Run tests (short mode, skips slow E2E tests)"
	@echo "  make test-all          - Run all tests (including slow E2E tests ~30s)"
	@echo "  make test-coverage     - Run tests with coverage report (includes E2E tests)"
	@echo "  make lint              - Run static analysis (fmt + vet + golangci-lint)"
	@echo "  make fmt               - Format code with gofmt"
	@echo "  make vet               - Run go vet"
	@echo "  make clean             - Remove build artifacts and synced files"
	@echo "  make install           - Install to GOPATH/bin"
	@echo "  make cross-compile     - Build for all platforms"
	@echo "  make sync-plugin-files - Sync .claude/ files to root for packaging"
	@echo "  make bundle-release    - Create release bundles (auto-syncs first, requires VERSION=vX.Y.Z)"
	@echo "  make deps              - Download and tidy dependencies"
	@echo "  make help              - Show this help message"
