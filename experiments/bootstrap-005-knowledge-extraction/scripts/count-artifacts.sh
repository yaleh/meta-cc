#!/bin/bash
#
# Count extractable artifacts in a BAIME experiment.
#
# Usage:
#   ./count-artifacts.sh <experiment_dir>
#
# Example:
#   ./count-artifacts.sh experiments/bootstrap-006-api-design

set -euo pipefail

EXPERIMENT_DIR="${1:?Error: Experiment directory required}"

if [ ! -d "$EXPERIMENT_DIR" ]; then
    echo "Error: Experiment directory not found: $EXPERIMENT_DIR" >&2
    exit 1
fi

echo "Counting artifacts in: $EXPERIMENT_DIR"
echo "========================================"
echo ""

# Count patterns
PATTERNS_COUNT=0
if [ -f "$EXPERIMENT_DIR/results.md" ]; then
    PATTERNS_COUNT=$(grep -c "^### Pattern" "$EXPERIMENT_DIR/results.md" || echo "0")
fi
echo "Patterns:    $PATTERNS_COUNT"

# Count principles/lessons
PRINCIPLES_COUNT=0
if [ -f "$EXPERIMENT_DIR/results.md" ]; then
    PRINCIPLES_COUNT=$(grep -cE "^### (Lesson|Principle)" "$EXPERIMENT_DIR/results.md" || echo "0")
fi
echo "Principles:  $PRINCIPLES_COUNT"

# Count templates
TEMPLATES_COUNT=0
if [ -d "$EXPERIMENT_DIR/knowledge/templates" ]; then
    TEMPLATES_COUNT=$(find "$EXPERIMENT_DIR/knowledge/templates" -name "*.md" -type f | wc -l)
fi
echo "Templates:   $TEMPLATES_COUNT"

# Count scripts
SCRIPTS_COUNT=0
if [ -d "$EXPERIMENT_DIR/scripts" ]; then
    SCRIPTS_COUNT=$(find "$EXPERIMENT_DIR/scripts" -name "*.sh" -o -name "*.py" -type f | wc -l)
fi
echo "Scripts:     $SCRIPTS_COUNT"

# Count iterations
ITERATIONS_COUNT=0
ITERATIONS_COUNT=$(find "$EXPERIMENT_DIR" -maxdepth 1 -name "iteration-*.md" -type f | wc -l)
echo "Iterations:  $ITERATIONS_COUNT"

# Count total markdown files
MARKDOWN_COUNT=0
MARKDOWN_COUNT=$(find "$EXPERIMENT_DIR" -name "*.md" -type f | wc -l)
echo "Markdown:    $MARKDOWN_COUNT files"

# Count total lines
TOTAL_LINES=0
if command -v wc &> /dev/null; then
    TOTAL_LINES=$(find "$EXPERIMENT_DIR" -name "*.md" -type f -exec wc -l {} + | tail -n 1 | awk '{print $1}')
fi
echo "Total lines: $TOTAL_LINES"

echo ""
echo "Extraction Inventory Summary:"
echo "  - $PATTERNS_COUNT patterns to extract"
echo "  - $PRINCIPLES_COUNT principles to extract"
echo "  - $TEMPLATES_COUNT templates to copy"
echo "  - $SCRIPTS_COUNT scripts to adapt"
echo "  - $ITERATIONS_COUNT iterations for examples"

echo ""
echo "Estimated extraction time: $((PATTERNS_COUNT * 3 + PRINCIPLES_COUNT * 2 + TEMPLATES_COUNT * 5 + SCRIPTS_COUNT * 10 + 30)) minutes"
