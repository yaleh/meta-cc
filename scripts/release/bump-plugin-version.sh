#!/bin/bash
set -e

# Usage: ./scripts/bump-plugin-version.sh [patch|minor|major]
#        ./scripts/bump-plugin-version.sh --version X.Y.Z [--non-interactive]
#
# This script bumps the plugin version when plugin-src/ files change.
# It should be run ONLY when:
# - plugin-src/commands/*.md changes (e.g., /meta command logic)
# - plugin-src/agents/*.md changes (e.g., new/modified subagents)
#
# It should NOT be run when:
# - capabilities/ files change (content updates, not plugin API changes)
# - CLI/MCP code changes (separate versioning)

# Parse --version and --non-interactive flags before positional args
EXPLICIT_VERSION=""
NON_INTERACTIVE=""
POSITIONAL_ARGS=()

while [[ $# -gt 0 ]]; do
    case "$1" in
        --version)
            EXPLICIT_VERSION="$2"
            shift 2
            ;;
        --non-interactive)
            NON_INTERACTIVE="1"
            shift
            ;;
        *)
            POSITIONAL_ARGS+=("$1")
            shift
            ;;
    esac
done

# Restore positional args
set -- "${POSITIONAL_ARGS[@]:-}"

BUMP_TYPE=${1:-patch}  # Default to patch version

# If --version provided, validate format; otherwise validate bump type
if [ -n "$EXPLICIT_VERSION" ]; then
    if [[ ! "$EXPLICIT_VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        echo "Error: Invalid version format '$EXPLICIT_VERSION'. Use X.Y.Z (e.g., 1.2.3)"
        exit 1
    fi
else
    # Validate bump type
    if [[ ! "$BUMP_TYPE" =~ ^(patch|minor|major)$ ]]; then
        echo "Error: Invalid bump type. Use: patch, minor, or major"
        echo "Usage: ./scripts/bump-plugin-version.sh [patch|minor|major]"
        exit 1
    fi
fi

# Check current branch (skip in non-interactive mode: release.sh manages branch checks)
if [ -z "$NON_INTERACTIVE" ]; then
    BRANCH=$(git rev-parse --abbrev-ref HEAD)
    if [[ "$BRANCH" != "main" && "$BRANCH" != "develop" ]]; then
        echo "Error: Must be on main or develop branch (current: $BRANCH)"
        exit 1
    fi
else
    BRANCH=$(git rev-parse --abbrev-ref HEAD)
fi

# Check working directory clean
# Skipped in --non-interactive mode: release.sh has already modified files; clean-dir guard is intentionally bypassed
if [ -z "$NON_INTERACTIVE" ] && [ -n "$(git status --porcelain)" ]; then
    echo "Error: Working directory not clean. Commit or stash changes first."
    exit 1
fi

# Check if jq is installed
if ! command -v jq &> /dev/null; then
    echo "Error: jq is required but not installed"
    echo "Install with: sudo apt-get install jq (Ubuntu/Debian) or brew install jq (macOS)"
    exit 1
fi

# Get current version from plugin-src/.claude-plugin/plugin.json
CURRENT=$(jq -r '.version' plugin-src/.claude-plugin/plugin.json)
echo "Current plugin version: $CURRENT"

if [ -n "$EXPLICIT_VERSION" ]; then
    # Use explicitly provided version; skip bump arithmetic
    NEW_VERSION="$EXPLICIT_VERSION"
else
    # Parse version components
    IFS='.' read -r MAJOR MINOR PATCH <<< "$CURRENT"

    # Bump version based on type
    case $BUMP_TYPE in
        major)
            MAJOR=$((MAJOR + 1))
            MINOR=0
            PATCH=0
            ;;
        minor)
            MINOR=$((MINOR + 1))
            PATCH=0
            ;;
        patch)
            PATCH=$((PATCH + 1))
            ;;
    esac

    NEW_VERSION="${MAJOR}.${MINOR}.${PATCH}"
fi

echo "New plugin version: $NEW_VERSION"
echo ""

# Confirm with user (skipped in --non-interactive mode)
if [ -z "$NON_INTERACTIVE" ]; then
    echo "This will update:"
    echo "  - plugin-src/.claude-plugin/plugin.json: $CURRENT → $NEW_VERSION"
    echo "  - .claude-plugin/marketplace.json: $CURRENT → $NEW_VERSION"
    echo "  - plugin-src/.claude-plugin/marketplace.json: $CURRENT → $NEW_VERSION"
    echo ""
    echo "Press Enter to continue, or Ctrl+C to abort..."
    read
fi

# Update plugin-src/.claude-plugin/plugin.json version
echo "Updating plugin-src/.claude-plugin/plugin.json..."
jq --arg ver "$NEW_VERSION" '.version = $ver' plugin-src/.claude-plugin/plugin.json > plugin-src/.claude-plugin/plugin.json.tmp
mv plugin-src/.claude-plugin/plugin.json.tmp plugin-src/.claude-plugin/plugin.json
echo "✓ plugin-src/.claude-plugin/plugin.json updated to $NEW_VERSION"

# Update marketplace.json version
echo "Updating marketplace.json..."
jq --arg ver "$NEW_VERSION" '.plugins[0].version = $ver' .claude-plugin/marketplace.json > .claude-plugin/marketplace.json.tmp
mv .claude-plugin/marketplace.json.tmp .claude-plugin/marketplace.json
echo "✓ marketplace.json updated to $NEW_VERSION"

# Update plugin-src/.claude-plugin/marketplace.json version
echo "Updating plugin-src/.claude-plugin/marketplace.json..."
jq --arg ver "$NEW_VERSION" '.plugins[0].version = $ver' plugin-src/.claude-plugin/marketplace.json > plugin-src/.claude-plugin/marketplace.json.tmp
mv plugin-src/.claude-plugin/marketplace.json.tmp plugin-src/.claude-plugin/marketplace.json
echo "✓ plugin-src/.claude-plugin/marketplace.json updated to $NEW_VERSION"
echo ""

# Commit changes
# Skipped in --non-interactive mode: release.sh owns the release commit
if [ -z "$NON_INTERACTIVE" ]; then
    echo "Committing version bump..."
    git add plugin-src/.claude-plugin/plugin.json .claude-plugin/marketplace.json plugin-src/.claude-plugin/marketplace.json
    git commit -m "chore: bump plugin version to $NEW_VERSION

Updated plugin.json and marketplace.json version.

This version bump reflects changes to plugin-src/ plugin structure
(commands or agents), not capabilities content updates.

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>"
    echo "✓ Version bump committed"
    echo ""
fi

echo "=== Plugin Version Bumped ==="
echo ""
echo "Next steps:"
echo "  1. Review the commit: git show HEAD"
echo "  2. Push to remote: git push origin $BRANCH"
echo ""
echo "Note: This only updates the plugin version."
echo "To create a full release (CLI + MCP + Plugin), use:"
echo "  ./scripts/release.sh v$NEW_VERSION"
