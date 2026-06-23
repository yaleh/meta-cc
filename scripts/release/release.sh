#!/bin/bash
# Automated release script
#
# Purpose: Create and publish a new release with full validation
# Usage: ./scripts/release.sh <version> [--skip-checks] [--dry-run]
# Example: ./scripts/release.sh v2.0.3
# Example: ./scripts/release.sh v2.0.3 --dry-run
#
# This script:
# 1. Runs pre-release validation checks
# 2. Updates marketplace.json version
# 3. Generates CHANGELOG entry
# 4. Commits version changes
# 5. Creates and pushes git tag
# 6. Triggers GitHub Actions release workflow

set -e

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
            echo "Usage: ./scripts/release.sh v1.0.0 [--skip-checks] [--dry-run]"
            exit 1
            ;;
    esac
    shift
done

if [ -z "$VERSION" ]; then
    echo "Error: Version required"
    echo "Usage: ./scripts/release.sh v1.0.0 [--skip-checks] [--dry-run]"
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

if [ -n "$DRY_RUN" ]; then
    echo "========================================"
    echo "DRY RUN MODE - No changes will be made"
    echo "========================================"
    echo ""
fi

echo "=== Release $VERSION ==="
echo ""

# ==================================================================
# STEP 1: Pre-Release Validation
# ==================================================================

if [ "$SKIP_CHECKS" != "--skip-checks" ]; then
    echo "Step 1: Running pre-release validation..."
    echo ""

    if [ -f "scripts/pre-release-check.sh" ]; then
        if bash scripts/pre-release-check.sh "$VERSION"; then
            echo ""
            echo "✓ Pre-release validation passed"
            echo ""
        else
            echo ""
            echo "❌ Pre-release validation failed"
            echo ""
            echo "Fix the issues above or run with --skip-checks to bypass (not recommended)"
            exit 1
        fi
    else
        echo "⚠️  Warning: pre-release-check.sh not found (skipping validation)"
        echo ""
        echo "Basic checks:"

        # Check current branch
        BRANCH=$(git rev-parse --abbrev-ref HEAD)
        if [[ "$BRANCH" != "main" && "$BRANCH" != "develop" ]]; then
            echo "Error: Must be on main or develop branch (current: $BRANCH)"
            exit 1
        fi

        # Check working directory clean
        if [ -n "$(git status --porcelain)" ]; then
            echo "Error: Working directory not clean. Commit or stash changes."
            exit 1
        fi

        # Run tests
        echo "Running tests..."
        make all
        echo "✓ Tests passed"
        echo ""
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
    echo "[DRY RUN] Would run: git add .claude-plugin/marketplace.json plugin-src/.claude-plugin/plugin.json CHANGELOG.md"
    echo "[DRY RUN] Would commit with message:"
    echo "    chore: release $VERSION"
    echo ""
    echo "    Update marketplace.json, plugin.json, and CHANGELOG.md to version $VERSION_NUM."
    echo ""
else
    git add .claude-plugin/marketplace.json plugin-src/.claude-plugin/plugin.json CHANGELOG.md
    git commit -m "chore: release $VERSION

Update marketplace.json, plugin.json, and CHANGELOG.md to version $VERSION_NUM.

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
