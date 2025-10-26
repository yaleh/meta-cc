#!/bin/bash
# Install git hooks for meta-cc development

set -e

HOOKS_DIR=".githooks"
GIT_HOOKS_DIR=".git/hooks"

echo "=== Installing Git Hooks ==="
echo ""

# Check if .githooks directory exists
if [ ! -d "$HOOKS_DIR" ]; then
    echo "Error: $HOOKS_DIR directory not found"
    exit 1
fi

# Install each hook
for hook in "$HOOKS_DIR"/*; do
    if [ -f "$hook" ]; then
        hook_name=$(basename "$hook")

        # Skip if it's a .sample file
        if [[ "$hook_name" == *.sample ]]; then
            continue
        fi

        echo "Installing $hook_name..."

        # Copy hook to .git/hooks/
        cp "$hook" "$GIT_HOOKS_DIR/$hook_name"
        chmod +x "$GIT_HOOKS_DIR/$hook_name"

        echo "✓ $hook_name installed"
    fi
done

echo ""
echo "=== Git Hooks Installed ==="
echo ""
echo "Active hooks:"
ls -1 "$GIT_HOOKS_DIR" | grep -v ".sample" || echo "  (none)"
echo ""
echo "To disable auto version bump:"
echo "  rm .git/hooks/pre-commit"
echo ""
echo "To reinstall hooks:"
echo "  ./scripts/install-hooks.sh"
