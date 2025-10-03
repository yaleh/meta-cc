#!/bin/bash

# Test script for Stage 8.11 workflow pattern commands
echo "=== Testing Workflow Pattern Detection Commands ==="
echo ""

PROJECT=/home/yale/work/meta-cc

echo "1. Testing 'analyze sequences' command..."
./meta-cc analyze sequences --project $PROJECT --min-length 2 --min-occurrences 2 --output json > /tmp/sequences.json
if [ $? -eq 0 ]; then
    echo "✓ analyze sequences: SUCCESS"
    echo "  Found $(jq '.sequences | length' /tmp/sequences.json) sequences"
else
    echo "✗ analyze sequences: FAILED"
    exit 1
fi
echo ""

echo "2. Testing 'analyze file-churn' command..."
./meta-cc analyze file-churn --project $PROJECT --threshold 3 --output json > /tmp/file_churn.json
if [ $? -eq 0 ]; then
    echo "✓ analyze file-churn: SUCCESS"
    echo "  Found $(jq '.high_churn_files | length' /tmp/file_churn.json) high churn files"
else
    echo "✗ analyze file-churn: FAILED"
    exit 1
fi
echo ""

echo "3. Testing 'analyze idle-periods' command..."
./meta-cc analyze idle-periods --project $PROJECT --threshold 5 --output json > /tmp/idle_periods.json
if [ $? -eq 0 ]; then
    echo "✓ analyze idle-periods: SUCCESS"
    echo "  Found $(jq '.idle_periods | length' /tmp/idle_periods.json) idle periods"
else
    echo "✗ analyze idle-periods: FAILED"
    exit 1
fi
echo ""

echo "4. Testing Markdown output..."
./meta-cc analyze sequences --project $PROJECT --min-length 3 --min-occurrences 3 --output md > /tmp/sequences.md
if [ $? -eq 0 ]; then
    echo "✓ Markdown output: SUCCESS"
else
    echo "✗ Markdown output: FAILED"
    exit 1
fi
echo ""

echo "=== All Tests Passed! ==="
echo ""
echo "Output files:"
echo "  - /tmp/sequences.json"
echo "  - /tmp/file_churn.json"
echo "  - /tmp/idle_periods.json"
echo "  - /tmp/sequences.md"
