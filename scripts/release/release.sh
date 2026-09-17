#!/bin/bash
# Automated release script
#
# Purpose: Create and publish a new release with full validation
# Usage: ./scripts/release/release.sh <version> [--skip-checks] [--dry-run]
# Example: ./scripts/release/release.sh v2.0.3
# Example: ./scripts/release/release.sh v2.0.3 --dry-run
# Example: make release VERSION=v2.0.3 DRY_RUN=1
#
# This script:
# 0. Runs fast preconditions (clean tree / branch) — seconds, before any long phase
# 1. Runs pre-release validation checks (scripts/release/pre-release-check.sh, FAIL-CLOSED)
# 2. Updates marketplace.json version
# 3. Generates CHANGELOG entry
# 4. Commits version changes
# 5. Creates and pushes git tag
# 6. Triggers GitHub Actions release workflow
#
# DIR-100 — fail-closed release validation
# ----------------------------------------
# Two defects this script closes structurally:
#
#   (a) The validator guard used to test `scripts/pre-release-check.sh` while the
#       script lives at `scripts/release/pre-release-check.sh`. The guard was
#       therefore always false, every release silently took a warn-and-continue
#       fallback ("pre-release-check.sh not found (skipping validation)"), and
#       real validation never ran. The validator is now resolved relative to THIS
#       SCRIPT (${BASH_SOURCE[0]}), never to $PWD, and a missing validator is a
#       hard failure rather than a warning (see STEP 1 below).
#   (b) The fallback's `make all` test phase was unbounded, so the loop driver's
#       Bash watchdog killed half-finished releases at 10 minutes. That fallback is
#       gone (AC2), and the validator invocation is bounded explicitly below.
#
# RELEASE BOUND (AC4). The release test phase is bounded two ways:
#   * SHORT MODE — the validator runs `go test -short ./...` (its Check 4); the
#     release path never runs the full, slow suite. The old unbounded `make all`
#     fallback is deleted, so no unbounded test path remains here.
#   * EXPLICIT TIMEOUT — the whole validator is wrapped in `timeout` with a
#     default bound of 300 seconds, override with RELEASE_VALIDATION_TIMEOUT
#     (seconds). Measured headroom: the validator's own `go test -short ./...`
#     passes are ~23s each, so 300s is generous while still finishing a full
#     release well under the loop driver's 10-minute (600s) watchdog instead of
#     racing it.
#   The bound is required, not best-effort: if no `timeout`/`gtimeout` binary is
#   available the script refuses to run (fail-closed) rather than silently
#   reverting to an unbounded validation.

set -e

# ------------------------------------------------------------------
# DIR-100: resolve the repo root from THIS SCRIPT's own location, never from $PWD.
# Every repo-relative path below (.claude-plugin/marketplace.json, CHANGELOG.md,
# git) and the validator itself are then correct no matter which directory the
# script was invoked from. The previous validator guard was resolved against $PWD
# and silently missed the file.
# ------------------------------------------------------------------
SCRIPT_PATH="${BASH_SOURCE[0]}"
SCRIPT_DIR="$(cd "$(dirname "$SCRIPT_PATH")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# The validator is a SIBLING of this script (scripts/release/pre-release-check.sh).
# Resolving it from SCRIPT_DIR is what makes the guard true; the historical
# `scripts/pre-release-check.sh` path never existed.
VALIDATOR="$SCRIPT_DIR/pre-release-check.sh"

# Hard bound for the validator (and therefore for the release test phase). AC4.
RELEASE_VALIDATION_TIMEOUT="${RELEASE_VALIDATION_TIMEOUT:-300}"

cd "$REPO_ROOT"

VERSION=$1
VERSION_NUM=${VERSION#v}  # Remove 'v' prefix
SKIP_CHECKS=""
DRY_RUN=""

# Parse optional flags
shift || true
while [ $# -gt 0 ]; do
    case "$1" in
        --skip-checks)
            SKIP_CHECKS="--skip-checks"
            ;;
        --dry-run)
            DRY_RUN="--dry-run"
            ;;
        *)
            echo "Error: Unknown option: $1"
            echo "Usage: ./scripts/release/release.sh v1.0.0 [--skip-checks] [--dry-run]"
            exit 1
            ;;
    esac
    shift
done

if [ -z "$VERSION" ]; then
    echo "Error: Version required"
    echo "Usage: ./scripts/release/release.sh v1.0.0 [--skip-checks] [--dry-run]"
    exit 1
fi

# Validate version format
if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9]+)?$ ]]; then
    echo "Error: Invalid version format. Use v1.0.0 or v1.0.0-beta"
    exit 1
fi

# Check if jq is installed
if ! command -v jq &> /dev/null; then
    echo "Error: jq is required but not installed"
    echo "Install with: sudo apt-get install jq (Ubuntu/Debian) or brew install jq (macOS)"
    exit 1
fi

# Check the release bound is enforceable (DIR-100 AC4). This is REQUIRED, not
# best-effort: without a `timeout` binary the validator would run unbounded and
# reintroduce exactly the watchdog-killed-release failure this task fixes, so we
# fail closed here instead of silently degrading.
if command -v timeout &> /dev/null; then
    TIMEOUT_BIN="timeout"
elif command -v gtimeout &> /dev/null; then
    TIMEOUT_BIN="gtimeout"
else
    echo "Error: no 'timeout' (or 'gtimeout') command found — cannot bound the release validation phase"
    echo "Install GNU coreutils: sudo apt-get install coreutils (Ubuntu/Debian) or brew install coreutils (macOS)"
    echo "Refusing to run an unbounded release (fail-closed). See DIR-100."
    exit 1
fi

# Validate the bound itself so a typo cannot silently disable it.
if ! [[ "$RELEASE_VALIDATION_TIMEOUT" =~ ^[0-9]+$ ]] || [ "$RELEASE_VALIDATION_TIMEOUT" -eq 0 ]; then
    echo "Error: RELEASE_VALIDATION_TIMEOUT must be a positive integer number of seconds (got: '$RELEASE_VALIDATION_TIMEOUT')"
    exit 1
fi

if [ -n "$DRY_RUN" ]; then
    echo "========================================"
    echo "DRY RUN MODE - No changes will be made"
    echo "========================================"
    echo ""
fi

echo "=== Release $VERSION ==="
echo ""

# ==================================================================
# STEP 0: Fast Preconditions (DIR-100 AC3)
# ==================================================================
# These cost milliseconds and run BEFORE any long-running phase, so a release
# that is going to fail fails in seconds — the 06:26:59 event should have died
# here instead of first entering validation. Severity deliberately mirrors the
# validator's own verdict for each condition (pre-release-check.sh Check 1.1
# hard-fails a dirty tree; Check 1.2 only warns about the branch), so these
# preconditions can never disagree with the validator that runs immediately after.

if [ "$SKIP_CHECKS" != "--skip-checks" ]; then
    echo "Step 0: Fast preconditions..."
    echo ""

    # Clean tree — fatal. A release commits version files and creates a tag, so a
    # dirty tree is never releasable; the validator fails on this too (Check 1.1).
    if [ -n "$(git status --porcelain)" ]; then
        echo "❌ Error: Working directory not clean. Commit or stash changes."
        echo ""
        echo "Dirty paths:"
        git status --porcelain | sed 's/^/  /'
        echo ""
        echo "Release aborted before validation (no long-running phase was entered)."
        exit 1
    fi
    echo "  ✓ Working directory is clean"

    # Branch — advisory, matching pre-release-check.sh Check 1.2.
    BRANCH=$(git rev-parse --abbrev-ref HEAD)
    if [ "$BRANCH" = "main" ] || [ "$BRANCH" = "develop" ]; then
        echo "  ✓ On release branch ($BRANCH)"
    else
        echo "  ⚠ Releasing from branch '$BRANCH' (main or develop expected)"
    fi
    echo ""
fi

# ==================================================================
# STEP 1: Pre-Release Validation (FAIL-CLOSED, DIR-100 AC1/AC2)
# ==================================================================
# The validator is resolved from SCRIPT_DIR (a sibling of this script), never
# from $PWD. An unresolvable validator is a hard failure: skipping validation
# silently is the exact defect this task fixes.
#
# The invocation is bounded by RELEASE_VALIDATION_TIMEOUT (default 300s) so the
# validator — including its `go test -short ./...` test phase — can never exceed
# the loop driver's watchdog. See the header for the full bound rationale.

if [ "$SKIP_CHECKS" != "--skip-checks" ]; then
    echo "Step 1: Running pre-release validation..."
    echo ""

    if [ ! -f "$VALIDATOR" ]; then
        echo "❌ ERROR: pre-release validation script not found — refusing to release (fail-closed)"
        echo ""
        echo "   Expected at: $VALIDATOR"
        echo "   (resolved relative to this script: $SCRIPT_PATH — NOT to \$PWD)"
        echo ""
        echo "   Release validation must not be skipped silently. Either:"
        echo "     1. restore $VALIDATOR, or"
        echo "     2. run with --skip-checks to bypass validation explicitly (not recommended)."
        exit 1
    fi

    echo "Validator: $VALIDATOR"
    echo "Bound:     ${RELEASE_VALIDATION_TIMEOUT}s via $TIMEOUT_BIN (override: RELEASE_VALIDATION_TIMEOUT=<seconds>)"
    echo ""

    VALIDATION_RC=0
    "$TIMEOUT_BIN" "$RELEASE_VALIDATION_TIMEOUT" bash "$VALIDATOR" "$VERSION" || VALIDATION_RC=$?

    if [ "$VALIDATION_RC" -eq 0 ]; then
        echo ""
        echo "✓ Pre-release validation passed"
        echo ""
    elif [ "$VALIDATION_RC" -eq 124 ] || [ "$VALIDATION_RC" -eq 137 ]; then
        echo ""
        echo "❌ ERROR: Pre-release validation exceeded its ${RELEASE_VALIDATION_TIMEOUT}s bound and was killed"
        echo ""
        echo "   The bound exists so a release cannot outlive the loop driver's watchdog."
        echo "   Investigate the slow check, or raise it deliberately with:"
        echo "     RELEASE_VALIDATION_TIMEOUT=<seconds> make release VERSION=$VERSION"
        exit 1
    else
        echo ""
        echo "❌ Pre-release validation failed"
        echo ""
        echo "Fix the issues above or run with --skip-checks to bypass (not recommended)"
        exit 1
    fi
else
    echo "⚠️  SKIPPING PRE-RELEASE CHECKS (--skip-checks flag used)"
    echo ""
fi

# Get current branch after validation
BRANCH=$(git rev-parse --abbrev-ref HEAD)

# ==================================================================
# STEP 2: Update Version Files
# ==================================================================

echo "Step 2: Updating version files..."
echo ""

# Update marketplace.json version
CURRENT_VERSION=$(jq -r '.plugins[0].version' .claude-plugin/marketplace.json)
echo "  Current version: v$CURRENT_VERSION"
echo "  Target version:  $VERSION ($VERSION_NUM)"
echo ""

if [ -n "$DRY_RUN" ]; then
    echo "[DRY RUN] Would call bump-plugin-version.sh --version $VERSION_NUM --non-interactive"
    echo ""
else
    bash scripts/release/bump-plugin-version.sh --version "$VERSION_NUM" --non-interactive
    echo ""

    # Verify version parity
    PLUGIN_JSON="plugin-src/.claude-plugin/plugin.json"
    MARKET_VER=$(jq -r '.plugins[0].version' .claude-plugin/marketplace.json)
    PLUGIN_VER=$(jq -r '.version' "$PLUGIN_JSON" 2>/dev/null || echo "N/A")
    if [ "$MARKET_VER" != "$PLUGIN_VER" ]; then
        echo "ERROR: Version mismatch after update: marketplace=$MARKET_VER plugin=$PLUGIN_VER"
        exit 1
    fi
    echo "✓ Version parity verified: $MARKET_VER"
    echo ""
fi

# ==================================================================
# STEP 3: Generate CHANGELOG Entry
# ==================================================================

echo "Step 3: Generating CHANGELOG entry..."
echo ""

if [ -n "$DRY_RUN" ]; then
    echo "[DRY RUN] Would generate CHANGELOG entry for $VERSION"
    if [ -f "scripts/release/generate-changelog-entry.sh" ]; then
        echo "[DRY RUN] Using: scripts/release/generate-changelog-entry.sh"
    else
        echo "[DRY RUN] Warning: scripts/release/generate-changelog-entry.sh not found"
        echo "[DRY RUN] Would require manual CHANGELOG update"
    fi
    echo ""
else
    # CHANGELOG generation is MANDATORY for releases
    if [ -f "scripts/release/generate-changelog-entry.sh" ]; then
        echo "Attempting automatic CHANGELOG generation..."
        if bash scripts/release/generate-changelog-entry.sh "$VERSION"; then
            echo "✓ CHANGELOG.md updated automatically"
            echo ""
        else
            echo ""
            echo "❌ ERROR: Automatic CHANGELOG generation failed"
            echo ""
            echo "CHANGELOG.md must be updated for release $VERSION."
            echo "Please manually add a CHANGELOG entry, then press Enter to continue..."
            echo "(Or press Ctrl+C to abort and fix the issue)"
            read

            # Verify CHANGELOG was actually updated
            if ! grep -q "\[$VERSION_NUM\]" CHANGELOG.md; then
                echo ""
                echo "❌ ERROR: CHANGELOG.md does not contain entry for [$VERSION_NUM]"
                echo "Release aborted - CHANGELOG entry is required"
                exit 1
            fi
            echo "✓ CHANGELOG.md entry verified"
            echo ""
        fi
    else
        echo "❌ ERROR: scripts/release/generate-changelog-entry.sh not found"
        echo ""
        echo "CHANGELOG.md must be updated for release $VERSION."
        echo "Please manually add a CHANGELOG entry, then press Enter to continue..."
        echo "(Or press Ctrl+C to abort)"
        read

        # Verify CHANGELOG was actually updated
        if ! grep -q "\[$VERSION_NUM\]" CHANGELOG.md; then
            echo ""
            echo "❌ ERROR: CHANGELOG.md does not contain entry for [$VERSION_NUM]"
            echo "Release aborted - CHANGELOG entry is required"
            exit 1
        fi
        echo "✓ CHANGELOG.md entry verified"
        echo ""
    fi
fi

# ==================================================================
# STEP 4: Commit Version Updates
# ==================================================================

echo "Step 4: Committing version updates..."
echo ""

if [ -n "$DRY_RUN" ]; then
    echo "[DRY RUN] Would run: git add .claude-plugin/marketplace.json plugin-src/.claude-plugin/plugin.json plugin-src/.claude-plugin/marketplace.json plugin-src/.codex-plugin/plugin.json internal/version/release.json CHANGELOG.md"
    echo "[DRY RUN] Would commit with message:"
    echo "    chore: release $VERSION"
    echo ""
    echo "    Update marketplace.json, plugin.json, and CHANGELOG.md to version $VERSION_NUM."
    echo ""
else
    git add .claude-plugin/marketplace.json plugin-src/.claude-plugin/plugin.json plugin-src/.claude-plugin/marketplace.json plugin-src/.codex-plugin/plugin.json internal/version/release.json CHANGELOG.md
    git commit -m "chore: release $VERSION

Update marketplace.json, plugin.json, Codex plugin.json, internal/version/release.json, and CHANGELOG.md to version $VERSION_NUM.

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>"
    echo "✓ Version updates committed"
    echo ""
fi

# ==================================================================
# STEP 5: Create Git Tag
# ==================================================================

echo "Step 5: Creating git tag..."
echo ""

if [ -n "$DRY_RUN" ]; then
    echo "[DRY RUN] Would create annotated tag: $VERSION"
    echo "[DRY RUN] Tag message: Release $VERSION"
    echo ""
else
    git tag -a "$VERSION" -m "Release $VERSION

See CHANGELOG.md for release notes.

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>"
    echo "✓ Tag $VERSION created"
    echo ""
fi

# ==================================================================
# STEP 6: Push to Remote
# ==================================================================

echo "Step 6: Pushing to remote..."
echo ""

if [ -n "$DRY_RUN" ]; then
    echo "[DRY RUN] Would run: git push origin $BRANCH"
    echo "[DRY RUN] Would run: git push origin $VERSION"
    echo ""
else
    git push origin "$BRANCH"
    git push origin "$VERSION"
    echo "✓ Pushed commits and tag to remote"
    echo ""
fi

# ==================================================================
# RELEASE COMPLETE
# ==================================================================

if [ -n "$DRY_RUN" ]; then
    echo "========================================="
    echo "DRY RUN COMPLETE - No changes were made"
    echo "========================================="
    echo ""
    echo "To perform the actual release, run:"
    echo "  ./scripts/release/release.sh $VERSION"
    echo ""
    echo "Summary of what would happen:"
    echo "  1. Update marketplace.json: v$CURRENT_VERSION → $VERSION_NUM"
    echo "  2. Generate/update CHANGELOG.md entry"
    echo "  3. Commit changes to $BRANCH branch"
    echo "  4. Create annotated tag: $VERSION"
    echo "  5. Push commits and tag to origin"
    echo "  6. Trigger GitHub Actions release workflow"
    echo ""
else
    echo "========================================="
    echo "Release $VERSION Complete"
    echo "========================================="
    echo ""
    echo "GitHub Actions will now:"
    echo "  1. Verify marketplace.json version"
    echo "  2. Build MCP server binaries (5 platforms)"
    echo "  3. Run smoke tests"
    echo "  4. Create GitHub Release"
    echo "  5. Upload release artifacts"
    echo ""
    echo "Monitor progress:"
    echo "  https://github.com/yaleh/meta-cc/actions"
    echo ""
    echo "Expected release URL:"
    echo "  https://github.com/yaleh/meta-cc/releases/tag/$VERSION"
    echo ""
fi
