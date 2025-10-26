#!/usr/bin/env bash
# capture-cli-metrics.sh: collect cyclomatic complexity and coverage for CLI commands

set -euo pipefail

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
REPO_ROOT=$(cd "${SCRIPT_DIR}/.." && pwd)
OUTPUT_DIR=${1:-"${REPO_ROOT}/build/methodology"}
GOCACHE_DIR="${REPO_ROOT}/.gocache"
PACKAGE="./cmd"

mkdir -p "${OUTPUT_DIR}"
mkdir -p "${GOCACHE_DIR}"

TIMESTAMP=$(date --iso-8601=seconds)
GOCYCLO_FILE="${OUTPUT_DIR}/gocyclo-cli-${TIMESTAMP}.txt"
COVER_PROFILE="${OUTPUT_DIR}/cmd-cli-cover.out"
COVER_SUMMARY="${OUTPUT_DIR}/coverage-cli-${TIMESTAMP}.txt"

# Use short mode to avoid integration dependencies until fixtures are synthesized
{
  echo "# gocyclo report for ${PACKAGE} (${TIMESTAMP})"
  gocyclo "${REPO_ROOT}/cmd" | sort -nr
} > "${GOCYCLO_FILE}"

CLI_HOME="${REPO_ROOT}/.tmp/cli-home"
mkdir -p "${CLI_HOME}/.claude/projects"

GOCACHE="${GOCACHE_DIR}" HOME="${CLI_HOME}" META_CC_PROJECTS_ROOT="${CLI_HOME}/.claude/projects" \
  go test -short -coverprofile "${COVER_PROFILE}" "${PACKAGE}"
GOCACHE="${GOCACHE_DIR}" go tool cover -func "${COVER_PROFILE}" > "${COVER_SUMMARY}"

printf "Cyclomatic complexity written to %s\n" "${GOCYCLO_FILE}"
printf "Coverage summary written to %s\n" "${COVER_SUMMARY}"
