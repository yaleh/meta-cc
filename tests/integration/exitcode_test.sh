#!/bin/bash
# Test exit code behavior

set +e  # Don't exit on command failure

echo "=== Exit Code Test ==="

# Build the binary first
echo "Building meta-cc..."
cd /home/yale/work/meta-cc || exit 1
go build -o meta-cc-test ./main.go
if [ $? -ne 0 ]; then
    echo "✗ Failed to build meta-cc"
    exit 1
fi
echo "✓ Build successful"

# Use the test binary
METACC="./meta-cc-test"

# Test 1: Success with results (exit 0)
echo ""
echo "[1/3] Testing success with results..."
$METACC query tools --limit 5 >/dev/null 2>&1
EXIT_CODE=$?

if [ $EXIT_CODE -ne 0 ]; then
    echo "✗ Expected exit 0, got $EXIT_CODE"
    rm -f meta-cc-test
    exit 1
fi
echo "✓ Exit 0 for success with results"

# Test 2: Success without results (exit 2)
echo ""
echo "[2/3] Testing success without results..."
$METACC query tools --where "tool='NonExistentToolXYZ123'" >/dev/null 2>&1
EXIT_CODE=$?

if [ $EXIT_CODE -ne 2 ]; then
    echo "✗ Expected exit 2, got $EXIT_CODE"
    rm -f meta-cc-test
    exit 1
fi
echo "✓ Exit 2 for no results"

# Test 3: Error (exit 1)
echo ""
echo "[3/3] Testing error..."
$METACC query tools --where "invalid AND syntax AND" >/dev/null 2>&1
EXIT_CODE=$?

if [ $EXIT_CODE -ne 1 ]; then
    echo "✗ Expected exit 1, got $EXIT_CODE"
    rm -f meta-cc-test
    exit 1
fi
echo "✓ Exit 1 for error"

# Test 4: Conditional logic based on exit codes
echo ""
echo "[4/4] Testing conditional logic..."

# Test successful query
if $METACC query tools --limit 1 >/dev/null 2>&1; then
    echo "✓ Conditional: if statement works for exit 0"
else
    EXIT_CODE=$?
    if [ $EXIT_CODE -eq 2 ]; then
        echo "✓ Conditional: exit 2 handled (no results)"
    else
        echo "✗ Unexpected exit code in conditional: $EXIT_CODE"
        rm -f meta-cc-test
        exit 1
    fi
fi

# Test no results query
$METACC query tools --where "tool='NonExistent'" >/dev/null 2>&1
EXIT_CODE=$?
if [ $EXIT_CODE -eq 2 ]; then
    echo "✓ Conditional: exit code 2 can be checked with [ \$? -eq 2 ]"
else
    echo "✗ Expected exit 2 in no-results check, got $EXIT_CODE"
    rm -f meta-cc-test
    exit 1
fi

# Clean up
rm -f meta-cc-test

echo ""
echo "=== All Exit Code Tests Passed ✅ ==="
