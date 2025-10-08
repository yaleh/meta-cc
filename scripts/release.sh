#!/bin/bash
set -e

# Usage: ./scripts/release.sh v1.0.0

VERSION=$1

if [ -z "$VERSION" ]; then
    echo "Error: Version required"
    echo "Usage: ./scripts/release.sh v1.0.0"
    exit 1
fi

# Validate version format
if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9]+)?$ ]]; then
    echo "Error: Invalid version format. Use v1.0.0 or v1.0.0-beta"
    exit 1
fi

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

echo "=== Release $VERSION ==="
echo ""

# Run full test suite
echo "Running tests..."
make all
echo "✓ Tests passed"
echo ""

# Prompt for CHANGELOG update
echo "Please update CHANGELOG.md with release notes for $VERSION"
echo "Press Enter when ready to continue, or Ctrl+C to abort..."
read

# Verify CHANGELOG was updated
if ! grep -q "$VERSION" CHANGELOG.md; then
    echo "Warning: $VERSION not found in CHANGELOG.md"
    echo "Continue anyway? (y/N)"
    read -r response
    if [[ ! "$response" =~ ^[Yy]$ ]]; then
        echo "Aborted"
        exit 1
    fi
fi

# Create tag
echo "Creating tag $VERSION..."
git tag -a "$VERSION" -m "Release $VERSION"
echo "✓ Tag created"
echo ""

# Push tag
echo "Pushing tag to remote..."
git push origin "$VERSION"
echo "✓ Tag pushed"
echo ""

echo "=== Release $VERSION Complete ==="
echo ""
echo "GitHub Actions will now:"
echo "  1. Build cross-platform binaries"
echo "  2. Create GitHub Release"
echo "  3. Upload binaries"
echo ""
echo "Monitor progress: https://github.com/yaleh/meta-cc/actions"
