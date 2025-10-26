#!/bin/bash
# Automated release script
#
# Purpose: Create and publish a new release with full validation
# Usage: ./scripts/release.sh <version> [--skip-checks]
# Example: ./scripts/release.sh v2.0.3
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
SKIP_CHECKS=${2:-}

if [ -z "$VERSION" ]; then
    echo "Error: Version required"
    echo "Usage: ./scripts/release.sh v1.0.0 [--skip-checks]"
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

jq --arg ver "$VERSION_NUM" '.plugins[0].version = $ver' .claude-plugin/marketplace.json > .claude-plugin/marketplace.json.tmp
mv .claude-plugin/marketplace.json.tmp .claude-plugin/marketplace.json
echo "✓ marketplace.json updated to $VERSION_NUM"
echo ""

# ==================================================================
# STEP 3: Generate CHANGELOG Entry
# ==================================================================

echo "Step 3: Generating CHANGELOG entry..."
echo ""

if [ -f "scripts/generate-changelog-entry.sh" ]; then
    if bash scripts/generate-changelog-entry.sh "$VERSION"; then
        echo "✓ CHANGELOG.md updated automatically"
    else
        echo "⚠️  Failed to generate CHANGELOG entry automatically"
        echo ""
        echo "Would you like to edit CHANGELOG.md manually? (y/N)"
        read -r response
        if [[ "$response" =~ ^[Yy]$ ]]; then
            echo "Please update CHANGELOG.md with release notes for $VERSION"
            echo "Press Enter when ready to continue, or Ctrl+C to abort..."
            read
        else
            echo "Aborted - CHANGELOG entry required for releases"
            exit 1
        fi
    fi
else
    echo "⚠️  scripts/generate-changelog-entry.sh not found"
    echo ""
    echo "Please update CHANGELOG.md with release notes for $VERSION"
    echo "Press Enter when ready to continue, or Ctrl+C to abort..."
    read
fi
echo ""

# ==================================================================
# STEP 4: Commit Version Updates
# ==================================================================

echo "Step 4: Committing version updates..."
echo ""

git add .claude-plugin/marketplace.json CHANGELOG.md
git commit -m "chore: release $VERSION

Update marketplace.json and CHANGELOG.md to version $VERSION_NUM.

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>"
echo "✓ Version updates committed"
echo ""

# ==================================================================
# STEP 5: Create Git Tag
# ==================================================================

echo "Step 5: Creating git tag..."
echo ""

git tag -a "$VERSION" -m "Release $VERSION

See CHANGELOG.md for release notes.

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>"
echo "✓ Tag $VERSION created"
echo ""

# ==================================================================
# STEP 6: Push to Remote
# ==================================================================

echo "Step 6: Pushing to remote..."
echo ""

git push origin "$BRANCH"
git push origin "$VERSION"
echo "✓ Pushed commits and tag to remote"
echo ""

# ==================================================================
# RELEASE COMPLETE
# ==================================================================

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
