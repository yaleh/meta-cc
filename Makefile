# Makefile for meta-cc

# DIR-035: harden PATH so every recipe's shell can find `go`/`gofmt`/`goimports`
# regardless of the invoking process's inherited PATH. This recurring infra
# failure ("go: command not found", exit 127) was previously worked around
# ad hoc in individual tasks' `extra.acceptance` strings instead of being
# fixed once, centrally, here. Do NOT strip this line as dead code — see
# tasks/DIR-035.md and the `check-path-independence` regression check below,
# which fails if this hardening is removed.
export PATH := $(PATH):/usr/local/go/bin:$(HOME)/go/bin

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')
# DIR-049: LDFLAGS_VALUE holds just the -X assignments (no "-ldflags" prefix,
# no surrounding quotes) so `print-ldflags-value` can hand it, verbatim, to a
# `go build -ldflags <value>` invocation from a test or script -- proving the
# actual Makefile-computed -X targets resolve, not a hand-copied duplicate.
# These target real package-level vars in internal/version (see
# internal/version/version.go); a previous version of this file pointed at
# the nonexistent github.com/yaleh/meta-cc/cmd package, which Go's linker
# silently no-ops on instead of erroring (see tasks/DIR-049.md).
LDFLAGS_VALUE := -X github.com/yaleh/meta-cc/internal/version.Commit=$(COMMIT) \
                 -X github.com/yaleh/meta-cc/internal/version.BuildTime=$(BUILD_TIME)
LDFLAGS := -ldflags "$(LDFLAGS_VALUE)"

# DIR-049: prints the exact -X assignments used by build/install/cross-compile,
# with no surrounding quotes, so a Go test can pass it straight to
# `go build -ldflags <value>` and prove the Makefile's actual LDFLAGS wiring
# resolves to real symbols (internal/version/ldflags_test.go).
print-ldflags-value:
	@echo $(LDFLAGS_VALUE)

GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOCLEAN := $(GOCMD) clean
GOMOD := $(GOCMD) mod
BUILD_DIR := build
DIST_DIR := dist
MCP_BINARY_NAME := meta-cc-mcp
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

# Default target when running 'make' without arguments
.DEFAULT_GOAL := all

.PHONY: all build stage test test-verbose test-all test-coverage clean install install-local install-user install-user-codex uninstall-local uninstall-user uninstall-legacy cross-compile bundle-release lint lint-errors fmt vet help sync-plugin-files dev check-workspace check-temp-files check-fixtures check-deps check-imports check-scripts check-debug check-go-quality pre-commit ci metrics-mcp check-test-quality check-formatting fix-formatting check-plugin-sync check-mod-tidy test-bats check-release-ready test-all-local pre-commit-full check-essential check-code-quality check-build-quality check-comprehensive check-commit-ready check-push-ready check-no-scanner test-e2e-mcp test-e2e-codex check-session-locator-scope check-docs check-path-independence _path-independence-probe print-ldflags-value

# ==============================================================================
# Build Quality Gates (BAIME Experiment - Iteration 1)
# ==============================================================================

# ==============================================================================
# QUALITY GATES - Unified Check Groups (Phase 28.5 Refactoring)
# ==============================================================================

# Group 1: Essential (P0) - Blocks commit
check-essential: check-temp-files check-fixtures check-deps check-session-locator-scope check-path-independence check-docs
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
	@echo "[1/3] Checking git tag exists..."
	@LATEST_TAG=$$(git describe --tags --abbrev=0 2>/dev/null || echo "none"); \
	if [ "$$LATEST_TAG" = "none" ]; then \
		echo "❌ ERROR: No git tags found"; \
		echo "Run 'git tag v0.1.0' or similar first"; \
		exit 1; \
	fi; \
	echo "✓ Latest tag: $$LATEST_TAG"
	@echo ""
	@echo "[2/3] Verifying marketplace.json version matches tag..."
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
	@echo "[3/3] Verifying release/tool-count consistency (internal/version, internal/release)..."
	@$(GOTEST) ./internal/version/... ./internal/release/...
	@echo "✓ Claude/Codex/server version and tool-count surfaces agree with internal/version/release.json"
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

# DIR-033: fails if any code outside internal/locator/ calls the raw,
# unscoped SessionLocator.FromSessionID directly instead of going through
# FromSessionIDScoped's cwd-boundary check. See scripts/checks/check-session-
# locator-scope.sh for the bug-class history this closes structurally.
check-session-locator-scope:
	@bash scripts/checks/check-session-locator-scope.sh

# DIR-078: offline documentation-contract gate. Fails if a current doc invokes a
# removed MCP tool, documents a Go prerequisite below the go.mod baseline, or a
# curated index/guide page carries a broken relative link. Historical/migration
# docs are allowlisted in internal/release/doc_contract_test.go. CI runs the
# same internal/release package via `make test`.
check-docs:
	@echo "=== Documentation Contract Check (DIR-078) ==="
	@$(GOTEST) ./internal/release/...

# DIR-035: regression check for the Makefile PATH hardening (see the
# `export PATH` line near the top of this file). Spawns a nested `make`
# invocation whose inherited PATH is stripped down to /usr/bin:/bin (the
# exact minimal-PATH failure mode documented in tasks/DIR-035.md, which
# recurred at least six times as "go: command not found", exit 127). If the
# PATH hardening line is ever removed, `go version` inside the nested make
# invocation will fail and this check will fail, catching the regression
# automatically instead of requiring another manual audit note.
check-path-independence:
	@echo "=== PATH Independence Regression Check (DIR-035) ==="
	@PROBE_LOG=$$(mktemp /tmp/meta-cc-path-independence.XXXXXX); \
	if env -i HOME="$$HOME" PATH=/usr/bin:/bin $(MAKE) --no-print-directory -C $(CURDIR) _path-independence-probe >"$$PROBE_LOG" 2>&1; then \
		echo "✓ go/gofmt reachable via Makefile PATH hardening even with a minimal inherited PATH"; \
		rm -f "$$PROBE_LOG"; \
	else \
		echo "❌ ERROR: PATH hardening regression detected"; \
		echo "   'env -i PATH=/usr/bin:/bin make _path-independence-probe' failed to find 'go'."; \
		echo "   This means the 'export PATH := \$$(PATH):/usr/local/go/bin:\$$(HOME)/go/bin'"; \
		echo "   line near the top of the Makefile was removed or broken. See tasks/DIR-035.md."; \
		echo "--- probe output ---"; \
		cat "$$PROBE_LOG" 2>/dev/null || true; \
		rm -f "$$PROBE_LOG"; \
		exit 1; \
	fi

# Internal helper target for check-path-independence. Not intended to be
# invoked directly; it is what proves the PATH hardening actually works,
# by requiring `go` to be resolvable purely via the Makefile's exported PATH.
_path-independence-probe:
	@go version >/dev/null
	@echo "✓ 'go version' succeeded inside PATH-stripped subprocess"

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
commit: check-essential check-no-scanner test
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
push: check-code-quality check-build-quality check-comprehensive test-all test-coverage-check test-bats lint build
	@echo ""
	@echo "✅ Ready to push"
	@echo ""
	@echo "All quality gates passed:"
	@echo "  ✓ Essential validation"
	@echo "  ✓ Code quality checks"
	@echo "  ✓ Build quality checks"
	@echo "  ✓ Comprehensive checks"
	@echo "  ✓ All tests passed (including E2E)"
	@echo "  ✓ Bats pipeline tests passed (incl. DIR-037 stage/Codex-registration regressions)"
	@echo "  ✓ Test coverage meets threshold (80%)"
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
	@mkdir -p bin
	$(GOBUILD) $(LDFLAGS) -o bin/$(MCP_BINARY_NAME) ./cmd/mcp-server

stage: build
	@echo "Staging binary to plugin-src/bin/..."
	@mkdir -p plugin-src/bin
	@TMP=$$(mktemp plugin-src/bin/.$(MCP_BINARY_NAME).XXXXXX) && \
	 cp bin/$(MCP_BINARY_NAME) "$$TMP" && \
	 chmod +x "$$TMP" && \
	 mv -f "$$TMP" plugin-src/bin/$(MCP_BINARY_NAME)
	@echo "✓ Staged plugin-src/bin/$(MCP_BINARY_NAME)"

install-local: stage
	@echo "Installing plugin at local scope (this project only)..."
	@PROJECT_PATH="$(shell pwd)"; \
	mkdir -p .claude; \
	jq -n \
		--arg path "$$PROJECT_PATH" \
		'{"permissions": {"allow": ["Bash(make:*)", "Bash(go test:*)"]}, "extraKnownMarketplaces": {"meta-cc-marketplace": {"source": {"source": "directory", "path": $$path}}}, "enabledPlugins": {"meta-cc@meta-cc-marketplace": true}}' \
		> .claude/settings.local.json; \
	echo "✓ Generated .claude/settings.local.json (source: $$PROJECT_PATH)"
	@if [ -f ~/.claude/settings.json ]; then \
		if ! jq -e '.enabledPlugins' ~/.claude/settings.json > /dev/null 2>&1; then \
			jq '. + {"enabledPlugins": {}}' ~/.claude/settings.json > /tmp/cc-settings-tmp.json && mv /tmp/cc-settings-tmp.json ~/.claude/settings.json; \
			echo "✓ Added enabledPlugins key to ~/.claude/settings.json (CC bug workaround)"; \
		fi; \
	fi
	@rm -rf ~/.claude/plugins/cache/meta-cc-marketplace/meta-cc/
	@echo "✓ Purged plugin cache"
	@mkdir -p ~/.claude/plugins; \
	INSTALLED=~/.claude/plugins/installed_plugins.json; \
	if [ ! -f "$$INSTALLED" ]; then \
		echo '{"version": 2, "plugins": {}}' > "$$INSTALLED"; \
	fi; \
	VERSION=$$(jq -r '.version' plugin-src/.claude-plugin/plugin.json); \
	PROJECT_PATH="$(shell pwd)"; \
	CACHE_PATH=~/.claude/plugins/cache/meta-cc-marketplace/meta-cc/$$VERSION; \
	NOW=$$(date -u +%Y-%m-%dT%H:%M:%S.000Z); \
	GIT_SHA=$$(git rev-parse --short HEAD 2>/dev/null || echo "unknown"); \
	jq --arg key "meta-cc@meta-cc-marketplace" \
		--arg ver "$$VERSION" \
		--arg path "$$PROJECT_PATH" \
		--arg cache "$$CACHE_PATH" \
		--arg now "$$NOW" \
		--arg sha "$$GIT_SHA" \
		'.plugins[$$key] = {"scope": "local", "version": $$ver, "source": {"source": "directory", "path": $$path}, "installedAt": $$now, "gitSha": $$sha}' \
		"$$INSTALLED" > /tmp/installed_plugins_tmp.json && mv /tmp/installed_plugins_tmp.json "$$INSTALLED"; \
	echo "✓ Updated ~/.claude/plugins/installed_plugins.json (version: $$VERSION, sha: $$GIT_SHA)"
	@echo ""
	@echo "✅ Local install complete. Restart Claude Code to load the plugin."

install-user: stage
	@if [ -f .claude/settings.local.json ] && jq -e '.enabledPlugins["meta-cc@meta-cc-marketplace"]' .claude/settings.local.json > /dev/null 2>&1; then \
		echo "❌ Local scope active. Run 'make uninstall-local' first."; \
		exit 1; \
	fi
	@echo "Installing plugin at user scope (~/.local/share/meta-cc)..."
	@mkdir -p ~/.local/share/meta-cc
	@rsync -a --delete plugin-src/ ~/.local/share/meta-cc/
	@echo "✓ Copied plugin-src/ to ~/.local/share/meta-cc/ (includes .claude-plugin/marketplace.json with source='.')"
	@VERSION=$$(jq -r '.version' plugin-src/.claude-plugin/plugin.json); \
	 jq --arg ver "$$VERSION" '.plugins[0].version = $$ver | .plugins[0].source = "."' ~/.local/share/meta-cc/.claude-plugin/marketplace.json > /tmp/mcc-mkt.tmp && \
	 mv /tmp/mcc-mkt.tmp ~/.local/share/meta-cc/.claude-plugin/marketplace.json; \
	 echo "✓ Stamped installed marketplace.json version: $$VERSION source: ."
	@mkdir -p ~/.claude; \
	SETTINGS=~/.claude/settings.json; \
	if [ ! -f "$$SETTINGS" ]; then \
		echo '{}' > "$$SETTINGS"; \
	fi; \
	jq '. + {"extraKnownMarketplaces": ((.extraKnownMarketplaces // {}) + {"meta-cc-marketplace": {"source": {"source": "directory", "path": (env.HOME + "/.local/share/meta-cc")}}}), "enabledPlugins": ((.enabledPlugins // {}) + {"meta-cc@meta-cc-marketplace": true})}' \
		"$$SETTINGS" > /tmp/cc-user-settings-tmp.json && mv /tmp/cc-user-settings-tmp.json "$$SETTINGS"; \
	echo "✓ Updated ~/.claude/settings.json"
	@rm -rf ~/.claude/plugins/cache/meta-cc-marketplace/meta-cc/
	@echo "✓ Purged plugin cache"
	@echo ""
	@echo "✅ User install complete. Restart Claude Code to load the plugin."
	@echo ""
	@echo "Note: this only registers the plugin for Claude Code. If you also use"
	@echo "Codex and want it pointed at this same user-scope install instead of a"
	@echo "dev-mode repo path, run 'make install-user-codex' once as well."

install-user-codex:
	@echo "Registering and refreshing Codex against the user-scope install..."
	@if [ ! -d "$(HOME)/.local/share/meta-cc" ]; then \
		echo "❌ $(HOME)/.local/share/meta-cc does not exist yet."; \
		echo "   Run 'make install-user' first, then re-run 'make install-user-codex'."; \
		exit 1; \
	fi
	@CODEX_HOME="$${CODEX_HOME:-$(HOME)/.codex}"; \
	python3 scripts/install/update-codex-marketplace-toml.py \
		"$$CODEX_HOME/config.toml" "$(HOME)/.local/share/meta-cc" local && \
	CODEX_HOME="$$CODEX_HOME" scripts/install/refresh-codex-plugin.sh \
		"$(HOME)/.local/share/meta-cc"
	@echo ""
	@echo "✅ Codex registration and cache refresh complete."

uninstall-local:
	@echo "Removing local scope plugin install..."
	@if [ -f .claude/settings.local.json ]; then \
		jq 'del(.enabledPlugins["meta-cc@meta-cc-marketplace"])' .claude/settings.local.json > /tmp/settings_local_tmp.json && mv /tmp/settings_local_tmp.json .claude/settings.local.json; \
		echo "✓ Removed meta-cc@meta-cc-marketplace from .claude/settings.local.json"; \
	else \
		echo "  (no .claude/settings.local.json found, skipping)"; \
	fi
	@rm -rf ~/.claude/plugins/cache/meta-cc-marketplace/meta-cc/
	@echo "✓ Purged plugin cache"
	@INSTALLED=~/.claude/plugins/installed_plugins.json; \
	if [ -f "$$INSTALLED" ]; then \
		jq 'del(.plugins["meta-cc@meta-cc-marketplace"])' "$$INSTALLED" > /tmp/installed_plugins_tmp.json && mv /tmp/installed_plugins_tmp.json "$$INSTALLED"; \
		echo "✓ Removed meta-cc@meta-cc-marketplace from $$INSTALLED"; \
	else \
		echo "  (no installed_plugins.json found, skipping)"; \
	fi
	@echo ""
	@echo "✅ Local uninstall complete."

uninstall-user:
	@if [ -f .claude/settings.local.json ] && jq -e '.enabledPlugins["meta-cc@meta-cc-marketplace"]' .claude/settings.local.json > /dev/null 2>&1; then \
		echo "Note: local scope also active. Run 'make uninstall-local' to remove it."; \
	fi
	@echo "Removing user scope plugin install..."
	@rm -rf ~/.local/share/meta-cc/
	@echo "✓ Removed ~/.local/share/meta-cc/"
	@SETTINGS=~/.claude/settings.json; \
	if [ -f "$$SETTINGS" ]; then \
		jq 'del(.extraKnownMarketplaces["meta-cc-marketplace"]) | del(.enabledPlugins["meta-cc@meta-cc-marketplace"])' "$$SETTINGS" > /tmp/cc-user-settings-tmp.json && mv /tmp/cc-user-settings-tmp.json "$$SETTINGS"; \
		echo "✓ Removed meta-cc entries from ~/.claude/settings.json"; \
	else \
		echo "  (no ~/.claude/settings.json found, skipping)"; \
	fi
	@rm -rf ~/.claude/plugins/cache/meta-cc-marketplace/meta-cc/
	@echo "✓ Purged plugin cache"
	@echo ""
	@echo "✅ User uninstall complete."

uninstall-legacy:
	@echo "Removing legacy meta-cc artifacts..."
	@if [ -f ~/.claude/mcp.json ]; then \
		if jq -e '.mcpServers["meta-cc"]' ~/.claude/mcp.json > /dev/null 2>&1; then \
			jq 'del(.mcpServers["meta-cc"])' ~/.claude/mcp.json > /tmp/mcp_tmp.json && mv /tmp/mcp_tmp.json ~/.claude/mcp.json; \
			echo "✓ Removed meta-cc from ~/.claude/mcp.json"; \
		else \
			echo "  (meta-cc not found in ~/.claude/mcp.json, skipping)"; \
		fi; \
	else \
		echo "  (no ~/.claude/mcp.json found, skipping)"; \
	fi
	@if [ -f ~/.local/bin/meta-cc-mcp ]; then \
		rm -f ~/.local/bin/meta-cc-mcp; \
		echo "✓ Removed ~/.local/bin/meta-cc-mcp"; \
	else \
		echo "  (no ~/.local/bin/meta-cc-mcp found, skipping)"; \
	fi
	@for cmd in prompt-find prompt-list prompt-show; do \
		if [ -f ~/.claude/commands/$${cmd}.md ]; then \
			rm -f ~/.claude/commands/$${cmd}.md; \
			echo "✓ Removed ~/.claude/commands/$${cmd}.md"; \
		else \
			echo "  (no ~/.claude/commands/$${cmd}.md found, skipping)"; \
		fi; \
	done
	@echo ""
	@echo "✅ Legacy uninstall complete."

test:
	@echo "Running tests (short mode, compact output)..."
	$(GOTEST) -short ./...

test-verbose:
	@echo "Running tests (short mode, verbose output)..."
	$(GOTEST) -short -v ./...

test-e2e-mcp: build
	@echo "Running MCP E2E tests..."
	@bash tests/e2e/mcp-e2e-simple.sh ./bin/$(MCP_BINARY_NAME)

test-e2e-codex: build
	@echo "Running Codex E2E tests..."
	@bash tests/e2e/codex-e2e.sh ./bin/$(MCP_BINARY_NAME)
	@echo ""
	@echo "Running Codex plugin-manager E2E test..."
	@GOOS=$$($(GOCMD) env GOOS); GOARCH=$$($(GOCMD) env GOARCH); \
	mkdir -p $(BUILD_DIR); \
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(MCP_BINARY_NAME)-$$GOOS-$$GOARCH ./cmd/mcp-server; \
	$(MAKE) bundle-release VERSION=v0.0.0-codex-e2e PLATFORMS="$$GOOS/$$GOARCH"; \
	bash tests/e2e/codex-plugin-manager-e2e.sh \
		"$(BUILD_DIR)/bundles/meta-cc-v0.0.0-codex-e2e-$$GOOS-$$GOARCH" \
		"./bin/$(MCP_BINARY_NAME)"

# DIR-086: test-all is the single suite-executing `go test` run in the push
# path.  It omits -short so slow/e2e tests execute (the E2E shell scripts are
# separate prerequisites), and it writes coverage.out so that test-coverage-check
# can validate the threshold without a second scan.  Do NOT add `test` back as a
# prerequisite — the short-mode run already gates `make commit` and repeating it
# here would be a second execution.
test-all: test-e2e-mcp test-e2e-codex
	@echo "Running all tests (including slow E2E tests ~30s)..."
	$(GOTEST) -v -coverpkg=./... -coverprofile=coverage.out ./...
	@echo ""
	@echo "✅ All tests passed (unit + E2E)"

test-coverage: build
	@echo "Running tests with coverage..."
	$(GOTEST) -short -v -coverpkg=./... -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

test-coverage-full: build
	@echo "Running tests with coverage (including E2E and slow tests)..."
	$(GOTEST) -v -coverpkg=./... -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

metrics-mcp:
	@echo "Capturing MCP server metrics snapshot..."
	@./scripts/ci/capture-mcp-metrics.sh
	@echo "✅ MCP metrics snapshot complete"

clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -f bin/$(MCP_BINARY_NAME)
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
	@mkdir -p $(DIST_DIR)/commands
	@echo "  Copying commands from plugin-src/commands/..."
	@cp plugin-src/commands/prompt-find.md $(DIST_DIR)/commands/ 2>/dev/null || true
	@cp plugin-src/commands/prompt-list.md $(DIST_DIR)/commands/ 2>/dev/null || true
	@cp plugin-src/commands/prompt-show.md $(DIST_DIR)/commands/ 2>/dev/null || true
	@echo "✓ Plugin files synced to $(DIST_DIR)/"
	@CMD_COUNT=$$(find $(DIST_DIR)/commands -name "*.md" 2>/dev/null | wc -l); \
	echo "✓ Total: $$CMD_COUNT command(s)"

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
		mkdir -p $$BUNDLE_DIR/bin $$BUNDLE_DIR/commands $$BUNDLE_DIR/skills $$BUNDLE_DIR/.claude-plugin $$BUNDLE_DIR/.codex-plugin $$BUNDLE_DIR/lib; \
		if [ "$${platform%/*}" = "windows" ]; then \
			cp $(BUILD_DIR)/$(MCP_BINARY_NAME)-$$PLATFORM_NAME.exe $$BUNDLE_DIR/bin/$(MCP_BINARY_NAME).exe 2>/dev/null || true; \
		else \
			cp $(BUILD_DIR)/$(MCP_BINARY_NAME)-$$PLATFORM_NAME $$BUNDLE_DIR/bin/$(MCP_BINARY_NAME) 2>/dev/null || true; \
		fi; \
		cp -r $(DIST_DIR)/commands/* $$BUNDLE_DIR/commands/; \
		cp -r plugin-src/skills/* $$BUNDLE_DIR/skills/; \
		cp -r lib/* $$BUNDLE_DIR/lib/; \
		cp -r .claude-plugin/* $$BUNDLE_DIR/.claude-plugin/; \
		cp plugin-src/.claude-plugin/plugin.json $$BUNDLE_DIR/.claude-plugin/ 2>/dev/null || true; \
		cp plugin-src/.mcp.json $$BUNDLE_DIR/ 2>/dev/null || true; \
		cp -r plugin-src/.codex-plugin/* $$BUNDLE_DIR/.codex-plugin/; \
		cp plugin-src/.codex-mcp.json $$BUNDLE_DIR/ 2>/dev/null || true; \
		jq '.commands |= map(gsub("\\./commands/"; "./commands/"))' $$BUNDLE_DIR/.claude-plugin/plugin.json > $$BUNDLE_DIR/.claude-plugin/plugin.json.tmp && mv $$BUNDLE_DIR/.claude-plugin/plugin.json.tmp $$BUNDLE_DIR/.claude-plugin/plugin.json 2>/dev/null || true; \
		jq '.plugins[0].commands |= map(gsub("\\./plugin-src/commands/"; "./commands/")) | .plugins[0].source = "."' $$BUNDLE_DIR/.claude-plugin/marketplace.json > $$BUNDLE_DIR/.claude-plugin/marketplace.json.tmp && mv $$BUNDLE_DIR/.claude-plugin/marketplace.json.tmp $$BUNDLE_DIR/.claude-plugin/marketplace.json 2>/dev/null || true; \
		cp scripts/install/install.sh $$BUNDLE_DIR/; \
		cp scripts/install/uninstall.sh $$BUNDLE_DIR/ 2>/dev/null || true; \
		cp scripts/install/install-mcp.sh $$BUNDLE_DIR/ 2>/dev/null || true; \
		cp scripts/install/install-skills.sh $$BUNDLE_DIR/ 2>/dev/null || true; \
		cp README.md $$BUNDLE_DIR/; \
		cp LICENSE $$BUNDLE_DIR/; \
		tar -czf $(BUILD_DIR)/meta-cc-bundle-$$PLATFORM_NAME.tar.gz -C $(BUILD_DIR)/bundles meta-cc-$(VERSION)-$$PLATFORM_NAME; \
	done
	@echo "Bundle creation complete. Archives in $(BUILD_DIR)/"

deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

check-no-scanner:
	@echo "Checking for raw bufio.NewScanner on line-oriented parsing paths..."
	@if grep -rn "bufio\.NewScanner" internal/ cmd/mcp-server/ --include="*.go" \
		| grep -v "main\.go" \
		| grep -v "_test\.go" \
		| grep -v "^Binary"; then \
		echo "ERROR: Found raw bufio.NewScanner usage. Use parser.ReadLineFiltered for JSONL session content, or parser.ReadLineBounded for other line-length-sensitive, non-JSONL reads (arbitrary text, subprocess framing, etc.)."; \
		exit 1; \
	fi
	@echo "OK: No raw bufio.NewScanner found."

lint: fmt vet lint-errors lint-error-handling lint-markdown check-no-scanner
	@echo "Running static analysis..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not found. Install with:"; \
		echo "  go install github.com/golangci/golangci-lint/cmd/golangci-lint@v2.12.2"; \
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

# DIR-086: test-coverage-check only validates the threshold against
# coverage.out produced by test-all (the single suite-executing run in the push
# path).  Do NOT add a `go test -coverpkg=...` line here — that would be a
# second execution.
test-coverage-check:
	@bash scripts/checks/check-coverage.sh 80

lint-fix:
	@echo "Running golangci-lint with auto-fix..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --fix ./...; \
	else \
		echo "golangci-lint not found. Install with:"; \
		echo "  go install github.com/golangci/golangci-lint/cmd/golangci-lint@v2.12.2"; \
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
	@echo "  make stage                   - Build + copy binary to plugin-src/bin/ for local install"
	@echo "  make test                    - Run tests (short mode, compact output)"
	@echo "  make test-verbose            - Run tests (short mode, verbose per-test output)"
	@echo "  make test-all                - Run full test suite + coverage profile (single pass)"
	@echo "  make test-e2e-codex          - Run Codex install/session E2E tests"
	@echo "  make test-coverage           - Run tests with coverage report"
	@echo "  make test-coverage-check     - Verify coverage threshold (\(\geq 80\%\)) from test-all profile"
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
	@echo "  make check-essential         - P0: Essential validation (temp files, fixtures, deps, PATH independence)"
	@echo "  make check-docs              - P0: Documentation contract gate (removed tools, Go baseline, nav links; DIR-078)"
	@echo "  make check-path-independence - Verify Makefile finds go/gofmt even with a minimal PATH (DIR-035)"
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
	@echo "Plugin Install/Uninstall:"
	@echo "  make install-local           - Install plugin at local scope (this project only)"
	@echo "  make install-user            - Install plugin at user scope (all projects, this machine)"
	@echo "  make install-user-codex      - Point Codex's ~/.codex/config.toml at the user-scope install (run once after install-user)"
	@echo "  make uninstall-local         - Remove local scope plugin install"
	@echo "  make uninstall-user          - Remove user scope plugin install"
	@echo "  make uninstall-legacy        - Remove old-style legacy artifacts (mcp.json, ~/.local/bin, ~/.claude/commands)"
	@echo ""
	@echo "Build & Package:"
	@echo "  make cross-compile           - Build MCP server for all platforms"
	@echo "  make sync-plugin-files       - Prepare plugin files in $(DIST_DIR)/ for packaging"
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
