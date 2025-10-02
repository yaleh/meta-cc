#!/bin/bash
# Integration test script for meta-cc Slash Commands

set -e  # Exit on error

echo "=== meta-cc Slash Commands Integration Test ==="
echo ""

# Step 1: Check meta-cc installation
echo "[1/5] Checking meta-cc installation..."
if ! command -v meta-cc &> /dev/null; then
    echo "❌ Error: meta-cc not installed or not in PATH"
    exit 1
fi
echo "✅ meta-cc installed: $(which meta-cc)"
echo ""

# Step 2: Check Slash Command files exist
echo "[2/5] Checking Slash Command files..."
if [ ! -f ".claude/commands/meta-stats.md" ]; then
    echo "❌ Error: .claude/commands/meta-stats.md does not exist"
    exit 1
fi
if [ ! -f ".claude/commands/meta-errors.md" ]; then
    echo "❌ Error: .claude/commands/meta-errors.md does not exist"
    exit 1
fi
echo "✅ Slash Command files exist"
echo ""

# Step 3: Test meta-cc parse stats command
echo "[3/5] Testing meta-cc parse stats..."
if ! meta-cc parse stats --output md &> /dev/null; then
    echo "❌ Error: meta-cc parse stats failed"
    exit 1
fi
echo "✅ meta-cc parse stats executed successfully"
echo ""

# Step 4: Test meta-cc analyze errors command
echo "[4/5] Testing meta-cc analyze errors..."
if ! meta-cc analyze errors --window 20 --output md &> /dev/null; then
    echo "❌ Error: meta-cc analyze errors failed"
    exit 1
fi
echo "✅ meta-cc analyze errors executed successfully"
echo ""

# Step 5: Test meta-cc parse extract command (used by /meta-errors)
echo "[5/5] Testing meta-cc parse extract..."
if ! meta-cc parse extract --type tools --filter "status=error" --output json &> /dev/null; then
    echo "❌ Error: meta-cc parse extract failed"
    exit 1
fi
echo "✅ meta-cc parse extract executed successfully"
echo ""

echo "=== All tests passed ✅ ==="
echo ""
echo "Next steps:"
echo "1. Open this project in Claude Code"
echo "2. Type /meta-stats to test statistics command"
echo "3. Type /meta-errors to test error analysis command"
echo "4. Type /meta-errors 50 to test custom window parameter"
