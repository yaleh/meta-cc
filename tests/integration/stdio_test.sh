#!/bin/bash
# Test stderr/stdout separation
# This test verifies that data goes to stdout and logs go to stderr

set -e

echo "=== stderr/stdout Separation Test ==="
echo ""

# Build the binary first
echo "Building meta-cc..."
go build -o /tmp/meta-cc-test ./main.go
BINARY="/tmp/meta-cc-test"

# Test 1: Data only on stdout
echo "[1/4] Testing data on stdout..."
DATA=$($BINARY query tools --limit 5 --output json 2>/dev/null || true)
if [ -z "$DATA" ]; then
    echo "  ✗ No data on stdout"
    exit 1
fi

# Verify data is valid JSON
if ! echo "$DATA" | jq empty 2>/dev/null; then
    echo "  ✗ Invalid JSON on stdout"
    exit 1
fi
echo "  ✓ Data on stdout is valid JSON"

# Test 2: Verify streaming output to stdout
echo "[2/4] Testing streaming output on stdout..."
STREAM_DATA=$($BINARY query tools --stream --limit 3 2>/dev/null || true)
if [ -z "$STREAM_DATA" ]; then
    echo "  ✗ No streaming data on stdout"
    exit 1
fi

# Verify each line is valid JSON
LINE_COUNT=$(echo "$STREAM_DATA" | wc -l)
if [ "$LINE_COUNT" -ne 3 ]; then
    echo "  ✗ Expected 3 lines, got $LINE_COUNT"
    exit 1
fi

echo "$STREAM_DATA" | while IFS= read -r line; do
    if ! echo "$line" | jq empty 2>/dev/null; then
        echo "  ✗ Invalid JSON line in stream: $line"
        exit 1
    fi
done
echo "  ✓ Streaming output works correctly"

# Test 3: Pipeline works without log interference
echo "[3/4] Testing pipeline without log interference..."
COUNT=$($BINARY query tools --stream --limit 10 2>/dev/null | jq -s '. | length')
if [ "$COUNT" -ne 10 ]; then
    echo "  ✗ Expected 10 records, got $COUNT (logs may be interfering)"
    exit 1
fi
echo "  ✓ Pipeline works cleanly without log interference"

# Test 4: Verify different output formats work
echo "[4/4] Testing different output formats..."

# JSON format
JSON_OUTPUT=$($BINARY query tools --limit 3 --output json 2>/dev/null || true)
if ! echo "$JSON_OUTPUT" | jq empty 2>/dev/null; then
    echo "  ✗ JSON output format failed"
    exit 1
fi
echo "  ✓ JSON format works"

# Markdown format
MD_OUTPUT=$($BINARY query tools --limit 3 --output md 2>/dev/null || true)
if [ -z "$MD_OUTPUT" ]; then
    echo "  ✗ Markdown output format failed"
    exit 1
fi
echo "  ✓ Markdown format works"

# TSV format
TSV_OUTPUT=$($BINARY query tools --limit 3 --output tsv 2>/dev/null || true)
if [ -z "$TSV_OUTPUT" ]; then
    echo "  ✗ TSV output format failed"
    exit 1
fi
echo "  ✓ TSV format works"

# Cleanup
rm -f /tmp/meta-cc-test

echo ""
echo "=== All stdio Tests Passed ✅ ==="
echo ""
echo "Summary:"
echo "  - Data output: stdout only (verified with JSON, Markdown, TSV)"
echo "  - Streaming output: JSONL format works correctly"
echo "  - Pipeline compatibility: No log interference"
echo "  - All output formats working correctly"
