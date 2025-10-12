#!/bin/bash
# Pre-commit hook: Auto-update plugin manifest before commits

set -e

echo "🔍 Pre-commit: Checking plugin manifest..."

# Update plugin.json with current files
bash scripts/update-plugin-manifest.sh

# Check if plugin.json was modified
if ! git diff --quiet .claude-plugin/plugin.json; then
    echo "✓ plugin.json updated with latest file list"
    git add .claude-plugin/plugin.json
    echo "  (auto-staged changes)"
fi

echo "✓ Pre-commit checks passed"
