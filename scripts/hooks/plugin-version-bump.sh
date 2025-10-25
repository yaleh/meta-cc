#!/bin/bash
# Pre-commit hook for meta-cc
# Auto-bumps plugin version when .claude/ files are modified

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if jq is installed
if ! command -v jq &> /dev/null; then
    echo -e "${YELLOW}Warning: jq not installed, skipping automatic version bump${NC}"
    exit 0
fi

# Get list of staged files
STAGED_FILES=$(git diff --cached --name-only --diff-filter=ACM)

# Check if .claude/commands/ or .claude/agents/ files are modified
PLUGIN_FILES_CHANGED=false
for file in $STAGED_FILES; do
    if [[ "$file" =~ ^\.claude/commands/.*\.md$ ]] || [[ "$file" =~ ^\.claude/agents/.*\.md$ ]]; then
        PLUGIN_FILES_CHANGED=true
        echo -e "${GREEN}Detected plugin file change: $file${NC}"
    fi
done

# If no plugin files changed, skip version bump
if [ "$PLUGIN_FILES_CHANGED" = false ]; then
    exit 0
fi

# Check if version files are already staged (user manually bumped version)
VERSION_FILES_STAGED=false
if echo "$STAGED_FILES" | grep -q "\.claude-plugin/plugin\.json\|\.claude-plugin/marketplace\.json"; then
    VERSION_FILES_STAGED=true
fi

if [ "$VERSION_FILES_STAGED" = true ]; then
    echo -e "${YELLOW}Version files already staged, skipping auto-bump${NC}"
    exit 0
fi

echo ""
echo -e "${GREEN}=== Auto Plugin Version Bump ===${NC}"
echo ""

# Get current version
CURRENT_VERSION=$(jq -r '.version' .claude-plugin/plugin.json)
echo "Current plugin version: $CURRENT_VERSION"

# Parse version components
IFS='.' read -r MAJOR MINOR PATCH <<< "$CURRENT_VERSION"

# Auto-increment patch version (conservative default)
PATCH=$((PATCH + 1))
NEW_VERSION="${MAJOR}.${MINOR}.${PATCH}"

echo "New plugin version: $NEW_VERSION (auto-bumped patch)"
echo ""

# Update plugin.json
jq --arg ver "$NEW_VERSION" '.version = $ver' .claude-plugin/plugin.json > .claude-plugin/plugin.json.tmp
mv .claude-plugin/plugin.json.tmp .claude-plugin/plugin.json

# Update marketplace.json
jq --arg ver "$NEW_VERSION" '.plugins[0].version = $ver' .claude-plugin/marketplace.json > .claude-plugin/marketplace.json.tmp
mv .claude-plugin/marketplace.json.tmp .claude-plugin/marketplace.json

# Stage the updated version files
git add .claude-plugin/plugin.json .claude-plugin/marketplace.json

echo -e "${GREEN}✓ Plugin version auto-bumped to $NEW_VERSION${NC}"
echo -e "${GREEN}✓ Version files staged${NC}"
echo ""
echo -e "${YELLOW}Note: Version was auto-incremented (patch).${NC}"
echo -e "${YELLOW}If you need minor/major bump, please:${NC}"
echo -e "${YELLOW}  1. Cancel this commit (Ctrl+C)${NC}"
echo -e "${YELLOW}  2. Run: ./scripts/bump-plugin-version.sh [minor|major]${NC}"
echo ""

exit 0
