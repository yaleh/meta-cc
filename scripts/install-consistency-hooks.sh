#!/bin/bash
#
# Install API consistency pre-commit hook
#
# This script copies the pre-commit hook to .git/hooks/ and makes it executable.
#

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
HOOKS_DIR="$PROJECT_ROOT/.git/hooks"
HOOK_FILE="$HOOKS_DIR/pre-commit"
SAMPLE_FILE="$SCRIPT_DIR/pre-commit.sample"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo ""
echo "==================================================="
echo "Installing API Consistency Pre-Commit Hook"
echo "==================================================="
echo ""

# Check if .git exists
if [ ! -d "$PROJECT_ROOT/.git" ]; then
    echo -e "${RED}Error: Not a git repository${NC}"
    echo "Run this script from within a git repository"
    exit 1
fi

# Check if validate-api binary exists
if [ ! -f "$PROJECT_ROOT/validate-api" ]; then
    echo -e "${YELLOW}Warning: validate-api binary not found${NC}"
    echo "Building validate-api..."
    (cd "$PROJECT_ROOT" && go build -o validate-api ./cmd/validate-api)
fi

# Create hooks directory if it doesn't exist
mkdir -p "$HOOKS_DIR"

# Check if pre-commit hook already exists
if [ -f "$HOOK_FILE" ]; then
    echo -e "${YELLOW}Warning: Pre-commit hook already exists${NC}"
    echo "Backing up existing hook to pre-commit.backup"
    mv "$HOOK_FILE" "$HOOK_FILE.backup"
fi

# Copy hook from sample
if [ ! -f "$SAMPLE_FILE" ]; then
    echo -e "${RED}Error: Hook sample not found at $SAMPLE_FILE${NC}"
    exit 1
fi

cp "$SAMPLE_FILE" "$HOOK_FILE"

# Make hook executable
chmod +x "$HOOK_FILE"

echo -e "${GREEN}✓ Pre-commit hook installed successfully${NC}"
echo ""
echo "Hook location: $HOOK_FILE"
echo ""

# Test hook
echo "Testing hook installation..."
if bash "$HOOK_FILE"; then
    echo ""
    echo -e "${GREEN}✓ Hook test successful${NC}"
    echo ""
else
    EXIT_CODE=$?
    if [ $EXIT_CODE -eq 1 ]; then
        echo ""
        echo -e "${YELLOW}⚠ Hook test detected violations (expected if tools.go has issues)${NC}"
        echo ""
    else
        echo ""
        echo -e "${RED}✗ Hook test failed with unexpected error${NC}"
        echo "Please check hook installation"
        exit 1
    fi
fi

echo "==================================================="
echo "Installation Complete"
echo "==================================================="
echo ""
echo "The pre-commit hook will now run before each commit."
echo "It validates API consistency if cmd/mcp-server/tools.go changes."
echo ""
echo "To bypass the hook (not recommended):"
echo "  git commit --no-verify"
echo ""
echo "To disable the hook:"
echo "  mv .git/hooks/pre-commit .git/hooks/pre-commit.disabled"
echo ""
