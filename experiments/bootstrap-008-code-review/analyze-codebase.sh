#!/bin/bash
# Analyze internal/ package codebase for Bootstrap-008 Iteration 0

cd /home/yale/work/meta-cc

echo "=== CODEBASE STRUCTURE ANALYSIS ==="
echo ""

echo "Modules in internal/:"
ls -1 internal/
echo ""

echo "=== SOURCE FILES PER MODULE ==="
for module in internal/*/; do
    name=$(basename "$module")
    src_count=$(find "$module" -maxdepth 1 -name "*.go" -not -name "*_test.go" | wc -l)
    test_count=$(find "$module" -maxdepth 1 -name "*_test.go" | wc -l)
    echo "$name: $src_count source, $test_count test"
done
echo ""

echo "=== LINES OF CODE PER MODULE ==="
for module in internal/*/; do
    name=$(basename "$module")
    src_files=$(find "$module" -maxdepth 1 -name "*.go" -not -name "*_test.go")
    if [ -n "$src_files" ]; then
        lines=$(echo "$src_files" | xargs cat | wc -l)
        echo "$name: $lines lines"
    fi
done
echo ""

echo "=== TOTAL STATISTICS ==="
total_src=$(find internal/ -name "*.go" -not -name "*_test.go" | wc -l)
total_test=$(find internal/ -name "*_test.go" | wc -l)
total_lines=$(find internal/ -name "*.go" -not -name "*_test.go" -exec cat {} \; | wc -l)
echo "Total source files: $total_src"
echo "Total test files: $total_test"
echo "Total lines of code: $total_lines"
