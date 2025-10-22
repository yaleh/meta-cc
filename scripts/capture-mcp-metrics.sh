#!/usr/bin/env bash
# capture-mcp-metrics.sh: collect refactoring metrics (cyclomatic complexity and coverage) for cmd/mcp-server

set -euo pipefail

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
REPO_ROOT=$(cd "${SCRIPT_DIR}/.." && pwd)
OUTPUT_DIR=${1:-"${REPO_ROOT}/build/methodology"}
GOCACHE_DIR="${REPO_ROOT}/.gocache"
PACKAGE="./cmd/mcp-server"

mkdir -p "${OUTPUT_DIR}"
mkdir -p "${GOCACHE_DIR}"

TIMESTAMP=$(date --iso-8601=seconds)
GOCYCLO_FILE="${OUTPUT_DIR}/gocyclo-mcp-${TIMESTAMP}.txt"
COVER_PROFILE="${OUTPUT_DIR}/cmd-mcp-cover.out"
COVER_SUMMARY="${OUTPUT_DIR}/coverage-mcp-${TIMESTAMP}.txt"

# Capture cyclomatic complexity snapshot (sorted descending for quick hotspot review)
{
  echo "# gocyclo report for ${PACKAGE} (${TIMESTAMP})"
  gocyclo "${REPO_ROOT}/cmd/mcp-server" | sort -nr
} > "${GOCYCLO_FILE}"

# Run coverage and produce summary (uses local GOCACHE to avoid sandbox issues)
GOCACHE="${GOCACHE_DIR}" go test "${PACKAGE}" -coverprofile "${COVER_PROFILE}"
GOCACHE="${GOCACHE_DIR}" go tool cover -func "${COVER_PROFILE}" > "${COVER_SUMMARY}"

printf "Cyclomatic complexity written to %s\n" "${GOCYCLO_FILE}"
printf "Coverage summary written to %s\n" "${COVER_SUMMARY}"
