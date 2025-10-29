#!/usr/bin/env bash
# count-artifacts.sh - Count lines in skill artifacts

set -euo pipefail

SKILL_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo "=== Artifact Line Count Report ==="
echo ""

total_lines=0

# SKILL.md
if [[ -f "$SKILL_DIR/SKILL.md" ]]; then
    lines=$(wc -l < "$SKILL_DIR/SKILL.md")
    total_lines=$((total_lines + lines))
    echo "SKILL.md: $lines lines"
    if [[ $lines -gt 40 ]]; then
        echo "  ⚠️  WARNING: Exceeds 40-line target ($(($lines - 40)) over)"
    else
        echo "  ✅ Within 40-line target"
    fi
    echo ""
fi

# Examples
echo "Examples:"
for file in "$SKILL_DIR"/examples/*.md; do
    if [[ -f "$file" ]]; then
        lines=$(wc -l < "$file")
        total_lines=$((total_lines + lines))
        basename=$(basename "$file")
        echo "  $basename: $lines lines"
        if [[ $lines -gt 150 ]]; then
            echo "    ⚠️  WARNING: Exceeds 150-line target ($(($lines - 150)) over)"
        else
            echo "    ✅ Within 150-line target"
        fi
    fi
done
echo ""

# Templates
echo "Templates:"
for file in "$SKILL_DIR"/templates/*.md; do
    if [[ -f "$file" ]]; then
        lines=$(wc -l < "$file")
        total_lines=$((total_lines + lines))
        basename=$(basename "$file")
        echo "  $basename: $lines lines"
    fi
done
echo ""

# Reference
echo "Reference:"
for file in "$SKILL_DIR"/reference/*.md; do
    if [[ -f "$file" ]]; then
        lines=$(wc -l < "$file")
        total_lines=$((total_lines + lines))
        basename=$(basename "$file")
        echo "  $basename: $lines lines"
    fi
done
echo ""

# Case Studies
echo "Case Studies:"
for file in "$SKILL_DIR"/reference/case-studies/*.md; do
    if [[ -f "$file" ]]; then
        lines=$(wc -l < "$file")
        total_lines=$((total_lines + lines))
        basename=$(basename "$file")
        echo "  $basename: $lines lines"
    fi
done
echo ""

echo "=== Summary ==="
echo "Total lines: $total_lines"
echo ""

# Compactness validation
compact_lines=0
if [[ -f "$SKILL_DIR/SKILL.md" ]]; then
    compact_lines=$((compact_lines + $(wc -l < "$SKILL_DIR/SKILL.md")))
fi
for file in "$SKILL_DIR"/examples/*.md; do
    if [[ -f "$file" ]]; then
        compact_lines=$((compact_lines + $(wc -l < "$file")))
    fi
done

echo "Compactness check (SKILL.md + examples):"
echo "  Total: $compact_lines lines"
if [[ $compact_lines -le 190 ]]; then
    echo "  ✅ Meets compactness target (≤190 lines for SKILL.md ≤40 + examples ≤150)"
else
    echo "  ⚠️  Exceeds compactness target ($((compact_lines - 190)) lines over)"
fi
