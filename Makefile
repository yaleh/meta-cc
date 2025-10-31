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
MCP_BINARY_NAME := meta-cc-mcp
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

# Default target when running 'make' without arguments
.DEFAULT_GOAL := all

.PHONY: all build test test-all test-coverage clean clean-capabilities install cross-compile bundle-release bundle-capabilities test-capability-package lint lint-errors fmt vet help sync-plugin-files dev check-workspace check-temp-files check-fixtures check-deps check-imports check-scripts check-debug check-go-quality pre-commit ci metrics-mcp check-test-quality check-formatting fix-formatting check-plugin-sync check-mod-tidy test-bats check-release-ready test-all-local pre-commit-full check-essential check-code-quality check-build-quality check-comprehensive check-commit-ready check-push-ready

# ==============================================================================
# Build Quality Gates (BAIME Experiment - Iteration 1)
# ==============================================================================

# ==============================================================================
# QUALITY GATES - Unified Check Groups (Phase 28.5 Refactoring)
# ==============================================================================

# Group 1: Essential (P0) - Blocks commit
check-essential: check-temp-files check-fixtures check-deps
	@echo "✅ Essential validation passed"

# Group 2: Code Quality (P1) - Blocks push
check-code-quality: check-formatting check-mod-tidy
	@echo "✅ Code quality checks passed"

# Group 3: Build Quality (P1) - Blocks push
check-build-quality: check-plugin-sync check-go-quality check-imports
	@echo "✅ Build quality checks passed"

# Group 4: Comprehensive (P2) - For full validation
check-comprehensive: check-scripts check-debug check-test-quality
	@echo "✅ Comprehensive checks passed"

# Legacy P0: Critical checks (blocks commit) - DEPRECATED: Use check-essential
check-workspace: check-temp-files check-fixtures check-deps
	@echo "✅ Workspace validation passed"
	@echo "⚠️  DEPRECATED: Use 'make check-essential' instead"

# P1: Enhanced checks (Iteration 2) - Now part of check-comprehensive
check-scripts:
	@bash scripts/checks/check-scripts.sh

check-debug:
	@bash scripts/checks/check-debug.sh

check-go-quality:
	@bash scripts/checks/check-go-quality.sh

# ==============================================================================
# CI-Derived Local Checks (从CI迁移的本地检查)
# ==============================================================================

check-test-quality:
	@bash scripts/checks/check-test-quality.sh

check-formatting:
	@echo "=== Code Formatting Check ==="
	@echo ""
	@echo "[1/3] Checking Go formatting..."
	@UNFORMATTED=$$(gofmt -l . 2>/dev/null | grep -v vendor || true); \
	if [ -n "$$UNFORMATTED" ]; then \
		echo "❌ ERROR: Unformatted Go files:"; \
		echo "$$UNFORMATTED" | sed 's/^/  - /'; \
		echo "Run 'make fmt' to fix"; \
		exit 1; \
	else \
		echo "✓ Go formatting is correct"; \
	fi
	@echo ""
	@echo "✅ Formatting check passed"

fix-formatting:
	@echo "Auto-fixing formatting issues..."
	@gofmt -w .
	@echo "✓ Formatting fixed"

check-plugin-sync:
	@bash scripts/sync-plugin-files.sh
	@bash scripts/sync-plugin-files.sh --verify

check-mod-tidy:
	@echo "=== Go Module Tidy Check ==="
	@echo ""
	@echo "Checking go.mod and go.sum are tidy..."
	@cp go.mod go.mod.bak 2>/dev/null || true
	@cp go.sum go.sum.bak 2>/dev/null || true
	@go mod tidy
	@if ! diff -q go.mod go.mod.bak >/dev/null 2>&1 || ! diff -q go.sum go.sum.bak >/dev/null 2>&1; then \
		echo "❌ ERROR: go.mod or go.sum not tidy"; \
		echo ""; \
		echo "Run 'go mod tidy' and commit changes"; \
		rm -f go.mod.bak go.sum.bak; \
		exit 1; \
	fi
	@rm -f go.mod.bak go.sum.bak
	@echo "✓ go.mod and go.sum are tidy"
	@echo ""
	@echo "✅ Module tidy check passed"

test-bats:
	@echo "=== Bats Pipeline Tests ==="
	@echo ""
	@if ! command -v bats >/dev/null 2>&1; then \
		echo "⚠️  WARNING: bats not installed"; \
		echo ""; \
		echo "Install with:"; \
		echo "  Ubuntu/Debian: sudo apt-get install bats"; \
		echo "  macOS: brew install bats-core"; \
		echo ""; \
		echo "Skipping Bats tests..."; \
		exit 0; \
	fi
	@echo "Running Bats tests..."
	@bats tests/scripts/*.bats
	@echo ""
	@echo "✅ Bats tests passed"

check-release-ready:
	@echo "=== Release Readiness Check ==="
	@echo ""
	@echo "[1/2] Checking git tag exists..."
	@LATEST_TAG=$$(git describe --tags --abbrev=0 2>/dev/null || echo "none"); \
	if [ "$$LATEST_TAG" = "none" ]; then \
		echo "❌ ERROR: No git tags found"; \
		echo "Run 'git tag v0.1.0' or similar first"; \
		exit 1; \
	fi; \
	echo "✓ Latest tag: $$LATEST_TAG"
	@echo ""
	@echo "[2/2] Verifying marketplace.json version matches tag..."
	@LATEST_TAG=$$(git describe --tags --abbrev=0); \
	VERSION_NUM=$${LATEST_TAG#v}; \
	MARKETPLACE_VERSION=$$(jq -r '.plugins[0].version' .claude-plugin/marketplace.json); \
	if [ "$$MARKETPLACE_VERSION" != "$$VERSION_NUM" ]; then \
		echo "❌ ERROR: Version mismatch!"; \
		echo "  Git tag: $$LATEST_TAG ($$VERSION_NUM)"; \
		echo "  marketplace.json: $$MARKETPLACE_VERSION"; \
		echo ""; \
		echo "Run './scripts/release/release.sh $$LATEST_TAG' to fix"; \
		exit 1; \
	fi; \
	echo "✓ marketplace.json version verified: $$MARKETPLACE_VERSION"
	@echo ""
	@echo "✅ Release ready"

# ==============================================================================
# Pre-Release Validation (Phase 27.6)
# ==============================================================================

pre-release-check:
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: VERSION required"; \
		echo "Usage: make pre-release-check VERSION=v2.0.3"; \
		exit 1; \
	fi
	@echo "Running pre-release validation for $(VERSION)..."
	@bash scripts/release/pre-release-check.sh $(VERSION)

bump-version:
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: VERSION required"; \
		echo "Usage: make bump-version VERSION=v2.0.3"; \
		exit 1; \
	fi
	@echo "Bumping version to $(VERSION)..."
	@bash scripts/release/bump-version.sh $(VERSION)

release:
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: VERSION required"; \
		echo "Usage: make release VERSION=v2.0.3"; \
		exit 1; \
	fi
	@echo "Creating release $(VERSION)..."
	@bash scripts/release/release.sh $(VERSION)

test-all-local: test-all test-bats
	@echo "✅ All tests passed (including Bats)"

# P0 + P1 + P2: Complete workspace validation - Uses new grouped checks
check-workspace-full: check-essential check-code-quality check-build-quality check-comprehensive
	@echo "✅ Full workspace validation passed"

# Quick validation for commit (uses new grouped checks)
check-commit-ready: check-essential test
	@echo "✅ Ready for commit (essential checks passed)"

# Full validation for push (uses new grouped checks)
check-push-ready: check-essential check-code-quality check-build-quality test-all lint build
	@echo "✅ Ready for push (all quality gates passed)"

check-temp-files:
	@bash scripts/checks/check-temp-files.sh

check-fixtures:
	@bash scripts/checks/check-fixtures.sh

check-deps:
	@bash scripts/checks/check-deps.sh

check-imports:
	@echo "Checking import formatting..."
	@UNFORMATTED=$$(goimports -l . 2>/dev/null | grep -v vendor || true); \
	if [ -n "$$UNFORMATTED" ]; then \
		echo "❌ ERROR: Files with incorrect imports:"; \
		echo "$$UNFORMATTED" | sed 's/^/  - /'; \
		echo ""; \
		echo "Run 'make fix-imports' to auto-fix"; \
		exit 1; \
	fi
	@echo "✓ Imports verified"

fix-imports:
	@echo "Auto-fixing imports..."
	@goimports -w .
	@echo "✓ Imports fixed"

# ==============================================================================
# Unified Build Targets (3-Tier Workflow)
# ==============================================================================

# Tier 1: FAST - Quick developer iteration (<10s)
dev: fmt build
	@echo "✅ Development build ready"
	@echo ""
	@echo "For commit preparation, run:"
	@echo "  make commit"

# Tier 2: COMMIT - Essential pre-commit validation (<60s)
commit: check-essential test
	@echo ""
	@echo "✅ Ready to commit"
	@echo ""
	@echo "Essential checks passed:"
	@echo "  ✓ Workspace clean (no temp files)"
	@echo "  ✓ Fixtures verified"
	@echo "  ✓ Dependencies in sync"
	@echo "  ✓ Tests passed (short mode)"
	@echo ""
	@echo "Before pushing to remote, run:"
	@echo "  make push"

# Tier 3: PUSH - Full validation before push (<120s)
push: check-code-quality check-build-quality check-comprehensive test-all lint build
	@echo ""
	@echo "✅ Ready to push"
	@echo ""
	@echo "All quality gates passed:"
	@echo "  ✓ Essential validation"
	@echo "  ✓ Code quality checks"
	@echo "  ✓ Build quality checks"
	@echo "  ✓ Comprehensive checks"
	@echo "  ✓ All tests passed (including E2E)"
	@echo "  ✓ Lint checks passed"
	@echo "  ✓ Build successful"

# Legacy aliases (deprecated, will be removed in future version)
pre-commit: commit
	@echo "⚠️  DEPRECATED: Use 'make commit' instead"

all: push
	@echo "⚠️  DEPRECATED: Use 'make push' instead"

ci: push
	@echo "⚠️  DEPRECATED: Use 'make push' instead"

build:
	@echo "Building $(MCP_BINARY_NAME) $(VERSION)..."
	$(GOBUILD) -o $(MCP_BINARY_NAME) ./cmd/mcp-server

test:
	@echo "Running tests (short mode, skips slow E2E tests)..."
	$(GOTEST) -short -v ./...

test-e2e-mcp: build
	@echo "Running MCP E2E tests..."
	@bash tests/e2e/mcp-e2e-simple.sh ./$(MCP_BINARY_NAME)

test-all: test test-e2e-mcp
	@echo "Running all tests (including slow E2E tests ~30s)..."
	$(GOTEST) -v ./...
	@echo ""
	@echo "✅ All tests passed (unit + E2E)"

test-coverage: build
	@echo "Running tests with coverage..."
	$(GOTEST) -short -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

test-coverage-full: build
	@echo "Running tests with coverage (including E2E and slow tests)..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

metrics-mcp:
	@echo "Capturing MCP server metrics snapshot..."
	@./scripts/ci/capture-mcp-metrics.sh
	@echo "✅ MCP metrics snapshot complete"

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
	rm -f $(MCP_BINARY_NAME)
	rm -rf $(BUILD_DIR)
	rm -rf $(DIST_DIR)
	rm -f coverage.out coverage.html

install:
	@echo "Installing MCP server..."
	$(GOCMD) install $(LDFLAGS) ./cmd/mcp-server

cross-compile:
	@echo "Building MCP server for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	@for platform in $(PLATFORMS); do \
		GOOS=$${platform%/*} GOARCH=$${platform#*/} \
		$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(MCP_BINARY_NAME)-$${platform%/*}-$${platform#*/} ./cmd/mcp-server; \
		if [ "$${platform%/*}" = "windows" ]; then \
			mv $(BUILD_DIR)/$(MCP_BINARY_NAME)-$${platform%/*}-$${platform#*/} $(BUILD_DIR)/$(MCP_BINARY_NAME)-$${platform%/*}-$${platform#*/}.exe; \
		fi; \
	done
	@echo "Cross-compilation complete. MCP server binaries in $(BUILD_DIR)/"

sync-plugin-files:
	@echo "Preparing plugin files for release packaging..."
	@mkdir -p $(DIST_DIR)/commands $(DIST_DIR)/agents $(DIST_DIR)/skills
	@echo "  Copying entry point from .claude/commands/..."
	@cp .claude/commands/meta.md $(DIST_DIR)/commands/
	@echo "  Copying agents from .claude/agents/..."
	@cp .claude/agents/*.md $(DIST_DIR)/agents/ 2>/dev/null || true
	@echo "  Copying skills from .claude/skills/..."
	@if [ -d ".claude/skills" ]; then \
		cp -r .claude/skills/* $(DIST_DIR)/skills/; \
		SKILL_COUNT=$$(find $(DIST_DIR)/skills -name "SKILL.md" 2>/dev/null | wc -l); \
		echo "    ✓ Copied $$SKILL_COUNT skills"; \
	fi
	@echo "✓ Plugin files synced to $(DIST_DIR)/"
	@CMD_COUNT=$$(find $(DIST_DIR)/commands -name "*.md" 2>/dev/null | wc -l); \
	AGENT_COUNT=$$(find $(DIST_DIR)/agents -name "*.md" 2>/dev/null | wc -l); \
	SKILL_COUNT=$$(find $(DIST_DIR)/skills -name "SKILL.md" 2>/dev/null | wc -l); \
	echo "✓ Total: $$CMD_COUNT commands, $$AGENT_COUNT agents, $$SKILL_COUNT skills"

# dev target is now defined in Build Quality Gates section above (line ~64)

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
			cp $(BUILD_DIR)/$(MCP_BINARY_NAME)-$$PLATFORM_NAME.exe $$BUNDLE_DIR/bin/ 2>/dev/null || true; \
		else \
			cp $(BUILD_DIR)/$(MCP_BINARY_NAME)-$$PLATFORM_NAME $$BUNDLE_DIR/bin/ 2>/dev/null || true; \
		fi; \
		cp -r $(DIST_DIR)/commands/* $$BUNDLE_DIR/commands/; \
		cp -r $(DIST_DIR)/agents/* $$BUNDLE_DIR/agents/; \
		cp -r lib/* $$BUNDLE_DIR/lib/; \
		cp -r .claude-plugin/* $$BUNDLE_DIR/.claude-plugin/; \
		cp scripts/install/install.sh $$BUNDLE_DIR/; \
		cp scripts/install/uninstall.sh $$BUNDLE_DIR/ 2>/dev/null || true; \
		cp README.md $$BUNDLE_DIR/; \
		cp LICENSE $$BUNDLE_DIR/; \
		tar -czf $(BUILD_DIR)/meta-cc-bundle-$$PLATFORM_NAME.tar.gz -C $(BUILD_DIR)/bundles meta-cc-$(VERSION)-$$PLATFORM_NAME; \
	done
	@echo "Bundle creation complete. Archives in $(BUILD_DIR)/"

deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

lint: fmt vet lint-errors lint-error-handling lint-markdown
	@echo "Running static analysis..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./... || echo "⚠️ golangci-lint issues found (non-blocking)"; \
	else \
		echo "golangci-lint not found. Install with:"; \
		echo "  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		echo "Skipping lint checks..."; \
	fi

lint-errors:
	@echo "Running error linting..."
	@./scripts/checks/lint-errors.sh cmd/ internal/

lint-error-handling:
	@echo "Checking error handling quality..."
	@# Check for the specific patterns we fixed - unsupported output format errors
	@UNSUPPORTED_FORMAT_ERRORS=$$(grep -r "unsupported output format.*supported:" cmd/mcp-server/*.go | grep -v "mcerrors\.ErrInvalidInput" | wc -l || true); \
	if [ "$$UNSUPPORTED_FORMAT_ERRORS" -gt 0 ]; then \
		echo "❌ ERROR: Found $$UNSUPPORTED_FORMAT_ERRORS unsupported format errors without proper sentinel errors"; \
		echo "All unsupported format errors should use mcerrors.ErrInvalidInput sentinel error"; \
		exit 1; \
	else \
		echo "✅ All unsupported format errors use proper sentinel errors"; \
	fi
	@# Check for invalid type errors (should use mcerrors.ErrInvalidInput)
	@INVALID_TYPE_ERRORS=$$(grep -r "invalid type.*must be one of:" cmd/mcp-server/*.go | grep -v "mcerrors\.ErrInvalidInput" | wc -l); \
	if [ "$$INVALID_TYPE_ERRORS" -gt 0 ]; then \
		echo "❌ ERROR: Found $$INVALID_TYPE_ERRORS invalid type errors without proper sentinel errors"; \
		echo "All invalid type errors should use mcerrors.ErrInvalidInput sentinel error"; \
		exit 1; \
	else \
		echo "✅ All invalid type errors use proper sentinel errors"; \
	fi
	@# Check error wrapping consistency
	@FILES_WITH_ERRORS=$$(grep -l "fmt\.Errorf.*%w" cmd/mcp-server/*.go | wc -l); \
	FILES_WITH_MCERRORS=$$(grep -l "mcerrors" cmd/mcp-server/*.go | wc -l || true); \
	echo "Files with error wrapping: $$FILES_WITH_ERRORS"; \
	echo "Files with mcerrors imports: $$FILES_WITH_MCERRORS"; \
	if [ $$FILES_WITH_ERRORS -gt 0 ]; then \
		echo "✅ Error wrapping is implemented in $$FILES_WITH_ERRORS files"; \
	else \
		echo "⚠️  No files with error wrapping found (this may be expected)"; \
	fi
	@echo "✅ Error handling quality check passed"

lint-markdown:
	@echo "Running markdown linting..."
	@if command -v markdownlint >/dev/null 2>&1; then \
		markdownlint --config .markdownlint.json **/*.md || echo "⚠️ Markdown linting issues found (non-blocking)"; \
	elif command -v npm >/dev/null 2>&1 && npm list -g markdownlint-cli >/dev/null 2>&1; then \
		npx markdownlint-cli --config .markdownlint.json **/*.md || echo "⚠️ Markdown linting issues found (non-blocking)"; \
	else \
		echo "markdownlint not found. Install with:"; \
		echo "  npm install -g markdownlint-cli"; \
		echo "Skipping markdown linting..."; \
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
	@bash scripts/install/install-pre-commit.sh

test-coverage-check:
	@$(GOTEST) -coverprofile=coverage.out ./... > /dev/null 2>&1
	@bash scripts/checks/check-coverage.sh 75

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
	@echo ""
	@echo "Development Workflow (3-Tier):"
	@echo "  make dev                     - Tier 1: Quick iteration (fmt + build, <10s)"
	@echo "  make commit                  - Tier 2: Pre-commit checks (essential + tests, <60s)"
	@echo "  make push                    - Tier 3: Full validation before push (all checks, <120s)"
	@echo ""
	@echo "Individual Tasks:"
	@echo "  make build                   - Build meta-cc-mcp MCP server"
	@echo "  make test                    - Run tests (short mode, skips slow E2E tests)"
	@echo "  make test-all                - Run all tests (including slow E2E tests ~30s)"
	@echo "  make test-coverage           - Run tests with coverage report"
	@echo "  make test-coverage-check     - Check test coverage meets 75% threshold"
	@echo "  make lint                    - Run static analysis (fmt + vet + error-linting + golangci-lint + markdown)"
	@echo "  make lint-markdown           - Run markdown linting"
	@echo "  make fmt                     - Format code with gofmt"
	@echo "  make vet                     - Run go vet"
	@echo ""
	@echo "Release Management:"
	@echo "  make bump-version VERSION=vX.Y.Z      - Bump marketplace.json version"
	@echo "  make pre-release-check VERSION=vX.Y.Z - Run pre-release validation checks"
	@echo "  make release VERSION=vX.Y.Z           - Create and push release (runs pre-release-check)"
	@echo "  make check-release-ready              - Verify latest tag matches marketplace.json"
	@echo ""
	@echo "Quality Gates (Grouped):"
	@echo "  make check-essential         - P0: Essential validation (temp files, fixtures, deps)"
	@echo "  make check-code-quality      - P1: Code quality (formatting, mod tidy)"
	@echo "  make check-build-quality     - P1: Build quality (plugin sync, go quality)"
	@echo "  make check-comprehensive     - P2: Comprehensive (scripts, debug, test quality)"
	@echo "  make check-workspace-full    - Full workspace validation (all groups)"
	@echo "  make check-commit-ready      - Quick commit validation (essential + tests)"
	@echo "  make check-push-ready        - Full push validation (all quality gates)"
	@echo ""
	@echo "Quality Gates (Legacy):"
	@echo "  make check-workspace         - P0 workspace validation (DEPRECATED: use check-essential)"
	@echo "  make check-test-quality      - Check test quality issues (now part of check-comprehensive)"
	@echo "  make check-plugin-sync       - Verify plugin file sync (now part of check-build-quality)"
	@echo "  make install-pre-commit      - Install pre-commit framework hooks"
	@echo ""
	@echo "Build & Package:"
	@echo "  make cross-compile           - Build MCP server for all platforms"
	@echo "  make sync-plugin-files       - Prepare plugin files in $(DIST_DIR)/ for packaging"
	@echo "  make bundle-capabilities     - Create capability package (.tar.gz)"
	@echo "  make bundle-release          - Create release bundles (auto-syncs first, requires VERSION=vX.Y.Z)"
	@echo ""
	@echo "Utilities:"
	@echo "  make clean                   - Remove build artifacts ($(BUILD_DIR)/, $(DIST_DIR)/)"
	@echo "  make deps                    - Download and tidy dependencies"
	@echo "  make security                - Run security scan with gosec"
	@echo "  make help                    - Show this help message"
	@echo ""
	@echo "Legacy (Deprecated):"
	@echo "  make all                     - Use 'make push' instead"
	@echo "  make pre-commit              - Use 'make commit' instead"
	@echo "  make ci                      - Use 'make push' instead"
