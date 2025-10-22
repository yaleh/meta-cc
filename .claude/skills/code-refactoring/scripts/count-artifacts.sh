#!/usr/bin/env bash
set -euo pipefail

SKILL_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
cd "${SKILL_DIR}"

count_files() {
  find "$1" -type f 2>/dev/null | wc -l | tr -d ' '
}

ITERATIONS=$(count_files "iterations")
TEMPLATES=$(count_files "templates")
SCRIPTS=$(count_files "scripts")
KNOWLEDGE=$(count_files "knowledge")
REFERENCE=$(count_files "reference")
EXAMPLES=$(count_files "examples")

cat <<JSON
{
  "iterations": ${ITERATIONS},
  "templates": ${TEMPLATES},
  "scripts": ${SCRIPTS},
  "knowledge": ${KNOWLEDGE},
  "reference": ${REFERENCE},
  "examples": ${EXAMPLES}
}
JSON
