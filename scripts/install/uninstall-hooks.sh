#!/bin/bash
# Uninstall git hooks for meta-cc

set -e

GIT_HOOKS_DIR=".git/hooks"

echo "=== Uninstalling Git Hooks ==="
echo ""

# Remove non-sample hooks
for hook in "$GIT_HOOKS_DIR"/*; do
    if [ -f "$hook" ] && [[ ! "$hook" == *.sample ]]; then
        hook_name=$(basename "$hook")
        echo "Removing $hook_name..."
        rm "$hook"
        echo "✓ $hook_name removed"
    fi
done

echo ""
echo "=== Git Hooks Uninstalled ==="
echo ""
echo "To reinstall hooks:"
echo "  ./scripts/install-hooks.sh"
