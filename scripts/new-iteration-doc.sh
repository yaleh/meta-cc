#!/usr/bin/env bash
# new-iteration-doc.sh: bootstrap a BAIME iteration document from template

set -euo pipefail

if [ $# -lt 2 ]; then
  echo "Usage: $0 <iteration-number> <title> [duration]" >&2
  exit 1
fi

ITER_NUM=$1
TITLE=$2
DURATION=${3:-"X.X hours"}

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
REPO_ROOT=$(cd "${SCRIPT_DIR}/.." && pwd)
TEMPLATE="${REPO_ROOT}/.claude/skills/code-refactoring/templates/iteration-template.md"
TARGET_DIR="${REPO_ROOT}/.claude/skills/code-refactoring/iterations"
TARGET_FILE="${TARGET_DIR}/iteration-${ITER_NUM}.md"

if [ ! -f "${TEMPLATE}" ]; then
  echo "Template not found: ${TEMPLATE}" >&2
  exit 1
fi

if [ -f "${TARGET_FILE}" ]; then
  echo "Iteration document already exists: ${TARGET_FILE}" >&2
  exit 1
fi

mkdir -p "${TARGET_DIR}"
DATE=$(date +%Y-%m-%d)
STATUS="Planned"
NEXT_FOCUS="TBD"
V_INSTANCE="N/A"
V_META="N/A"

escape_sed() {
  printf '%s' "$1" | sed -e 's/[\&/]/\\&/g'
}

TITLE_ESCAPED=$(escape_sed "${TITLE}")
DURATION_ESCAPED=$(escape_sed "${DURATION}")
STATUS_ESCAPED=$(escape_sed "${STATUS}")
NEXT_FOCUS_ESCAPED=$(escape_sed "${NEXT_FOCUS}")
V_INSTANCE_ESCAPED=$(escape_sed "${V_INSTANCE}")
V_META_ESCAPED=$(escape_sed "${V_META}")

sed -e "s/{{NUM}}/${ITER_NUM}/g" \
    -e "s/{{TITLE}}/${TITLE_ESCAPED}/g" \
    -e "s/{{DATE}}/${DATE}/g" \
    -e "s/{{DURATION}}/${DURATION_ESCAPED}/g" \
    -e "s/{{STATUS}}/${STATUS_ESCAPED}/g" \
    -e "s/{{NEXT_FOCUS}}/${NEXT_FOCUS_ESCAPED}/g" \
    -e "s/{{V_INSTANCE}}/${V_INSTANCE_ESCAPED}/g" \
    -e "s/{{V_META}}/${V_META_ESCAPED}/g" \
    "${TEMPLATE}" > "${TARGET_FILE}"

printf "Created iteration document: %s\n" "${TARGET_FILE}"
