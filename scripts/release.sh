#!/bin/bash
set -e

# Usage: ./scripts/release.sh v1.0.0
#
# NOTE: This script is now OPTIONAL with dynamic version injection.
# You can also release directly with: git tag v1.0.0 && git push origin v1.0.0
#
# This script provides:
# - Test suite validation before tagging
# - CHANGELOG.md validation
# - Consistent release process

VERSION=$1
VERSION_NUM=${VERSION#v}  # Remove 'v' prefix

if [ -z "$VERSION" ]; then
    echo "Error: Version required"
    echo "Usage: ./scripts/release.sh v1.0.0"
    echo ""
    echo "Alternative: Create and push tag directly"
    echo "  git tag -a v1.0.0 -m 'Release v1.0.0'"
    echo "  git push origin v1.0.0"
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
echo "Please ensure CHANGELOG.md has been updated with release notes for $VERSION"
echo "Press Enter when ready to continue, or Ctrl+C to abort..."
read

# Verify CHANGELOG was updated
if ! grep -q "## \[$VERSION_NUM\]" CHANGELOG.md; then
    echo "Warning: Version $VERSION_NUM not found in CHANGELOG.md"
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
echo "  1. Inject version $VERSION_NUM into plugin.json and marketplace.json"
echo "  2. Build cross-platform binaries"
echo "  3. Create GitHub Release"
echo "  4. Upload binaries"
echo ""
echo "Monitor progress: https://github.com/yaleh/meta-cc/actions"
echo ""
echo "Note: Version files (plugin.json, marketplace.json) remain 'dev' in the repository."
echo "      The actual version is injected during the build process from the git tag."
