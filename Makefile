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
DIST_DIR := dist
CAPABILITIES_DIR := capabilities
CAPABILITIES_ARCHIVE := capabilities-latest.tar.gz
BINARY_NAME := meta-cc
MCP_BINARY_NAME := meta-cc-mcp
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

.PHONY: all build build-cli build-mcp test test-all test-coverage clean clean-capabilities install cross-compile bundle-release bundle-capabilities test-capability-package lint fmt vet help sync-plugin-files dev

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

bundle-capabilities:
	@echo "Creating capability package: $(CAPABILITIES_ARCHIVE)..."
	@if [ ! -d "$(CAPABILITIES_DIR)/commands" ]; then \
		echo "ERROR: $(CAPABILITIES_DIR)/commands/ directory not found"; \
		exit 1; \
	fi
	@mkdir -p $(BUILD_DIR)
	@tar -czf $(BUILD_DIR)/$(CAPABILITIES_ARCHIVE) -C $(CAPABILITIES_DIR) commands agents 2>/dev/null || \
		tar -czf $(BUILD_DIR)/$(CAPABILITIES_ARCHIVE) -C $(CAPABILITIES_DIR) commands
	@echo "✓ Package created: $(BUILD_DIR)/$(CAPABILITIES_ARCHIVE)"
	@echo "  Size: $$(du -h $(BUILD_DIR)/$(CAPABILITIES_ARCHIVE) | cut -f1)"
	@echo "  Files: $$(tar -tzf $(BUILD_DIR)/$(CAPABILITIES_ARCHIVE) | wc -l)"

clean-capabilities:
	@echo "Cleaning capability packages..."
	@rm -f $(BUILD_DIR)/$(CAPABILITIES_ARCHIVE)

test-capability-package:
	@bash tests/integration/test-capability-package.sh

clean: clean-capabilities
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(MCP_BINARY_NAME)
	rm -rf $(BUILD_DIR)
	rm -rf $(DIST_DIR)
	rm -f coverage.out coverage.html

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
	@echo "Preparing plugin files for release packaging..."
	@mkdir -p $(DIST_DIR)/commands $(DIST_DIR)/agents
	@echo "  Copying entry point from .claude/commands/..."
	@cp .claude/commands/meta.md $(DIST_DIR)/commands/
	@echo "  Copying capabilities from $(CAPABILITIES_DIR)/commands/..."
	@cp $(CAPABILITIES_DIR)/commands/*.md $(DIST_DIR)/commands/ 2>/dev/null || true
	@echo "  Copying agents from .claude/agents/..."
	@cp .claude/agents/*.md $(DIST_DIR)/agents/ 2>/dev/null || true
	@echo "  Copying agents from $(CAPABILITIES_DIR)/agents/..."
	@cp $(CAPABILITIES_DIR)/agents/*.md $(DIST_DIR)/agents/ 2>/dev/null || true
	@echo "✓ Plugin files synced to $(DIST_DIR)/"
	@CMD_COUNT=$$(find $(DIST_DIR)/commands -name "*.md" 2>/dev/null | wc -l); \
	AGENT_COUNT=$$(find $(DIST_DIR)/agents -name "*.md" 2>/dev/null | wc -l); \
	echo "✓ Total: $$CMD_COUNT command files, $$AGENT_COUNT agent files"

dev: build
	@echo "Development build complete"
	@echo "Plugin files in .claude/ are ready for immediate use in Claude Code"
	@echo ""
	@echo "For local capability development, set:"
	@echo "  export META_CC_CAPABILITY_SOURCES=\"$(CAPABILITIES_DIR)/commands\""

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
		cp -r $(DIST_DIR)/commands/* $$BUNDLE_DIR/commands/; \
		cp -r $(DIST_DIR)/agents/* $$BUNDLE_DIR/agents/; \
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

# Quality gates (added in Bootstrap-008 Iteration 3)
install-pre-commit:
	@echo "Installing pre-commit hooks..."
	@bash scripts/install-pre-commit.sh

test-coverage-check:
	@echo "Checking test coverage meets 80% threshold..."
	@$(GOTEST) -coverprofile=coverage.out ./... > /dev/null 2>&1
	@COVERAGE=$$(go tool cover -func=coverage.out | tail -1 | awk '{print $$3}' | sed 's/%//'); \
	if [ "$$(echo "$$COVERAGE < 80" | bc)" -eq 1 ]; then \
		echo "FAIL: Coverage $$COVERAGE% is below 80% target"; \
		exit 1; \
	else \
		echo "PASS: Coverage $$COVERAGE% meets 80% target"; \
	fi

lint-fix:
	@echo "Running golangci-lint with auto-fix..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --fix ./...; \
	else \
		echo "golangci-lint not found. Install with:"; \
		echo "  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi

security:
	@echo "Running security scan with gosec..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec not found. Install with:"; \
		echo "  go install github.com/securego/gosec/v2/cmd/gosec@latest"; \
		echo "Skipping security scan..."; \
	fi

help:
	@echo "Available targets:"
	@echo "  make build                   - Build both meta-cc and meta-cc-mcp"
	@echo "  make build-cli               - Build meta-cc CLI only"
	@echo "  make build-mcp               - Build meta-cc-mcp MCP server only"
	@echo "  make dev                     - Development build (use .claude/ for immediate testing)"
	@echo "  make test                    - Run tests (short mode, skips slow E2E tests)"
	@echo "  make test-all                - Run all tests (including slow E2E tests ~30s)"
	@echo "  make test-coverage           - Run tests with coverage report (includes E2E tests)"
	@echo "  make test-coverage-check     - Check test coverage meets 80% threshold"
	@echo "  make test-capability-package - Test capability package creation and extraction"
	@echo "  make lint                    - Run static analysis (fmt + vet + golangci-lint)"
	@echo "  make lint-fix                - Run golangci-lint with auto-fix"
	@echo "  make fmt                     - Format code with gofmt"
	@echo "  make vet                     - Run go vet"
	@echo "  make security                - Run security scan with gosec"
	@echo "  make install-pre-commit      - Install pre-commit framework hooks"
	@echo "  make clean                   - Remove build artifacts ($(BUILD_DIR)/, $(DIST_DIR)/)"
	@echo "  make clean-capabilities      - Remove capability packages only"
	@echo "  make install                 - Install to GOPATH/bin"
	@echo "  make cross-compile           - Build for all platforms"
	@echo "  make bundle-capabilities     - Create capability package (.tar.gz)"
	@echo "  make sync-plugin-files       - Prepare plugin files in $(DIST_DIR)/ for packaging"
	@echo "  make bundle-release          - Create release bundles (auto-syncs first, requires VERSION=vX.Y.Z)"
	@echo "  make deps                    - Download and tidy dependencies"
	@echo "  make help                    - Show this help message"
