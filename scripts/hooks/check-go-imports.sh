#!/usr/bin/env bash
# Check Go imports without modifying files (DIR-092)
# Replaces the upstream go-imports hook which hardcodes -w
set -e -o pipefail

if ! command -v goimports &>/dev/null; then
    echo "goimports not installed or available in the PATH" >&2
    echo "please check https://pkg.go.dev/golang.org/x/tools/cmd/goimports" >&2
    exit 1
fi

# -l = list files that need fixing (no output = all good)
# Intentionally omitting -w to be check-only
output=$(goimports -l -local=github.com/yaleh/meta-cc "$@")
[[ -z "$output" ]]
