#!/bin/bash
# Pre-commit hook: Validate marketplace.json schema
# Ensures marketplace.json has valid JSON structure and required fields

set -e

if jq -e '.plugins[0].version' .claude-plugin/marketplace.json >/dev/null 2>&1; then
    exit 0
else
    echo "ERROR: Invalid marketplace.json structure"
    echo "Required field .plugins[0].version is missing or invalid"
    exit 1
fi
