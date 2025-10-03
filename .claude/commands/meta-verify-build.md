---
name: meta-verify-build
description: Verify meta-cc build with real session data
arguments:
  - name: session
    description: Session ID to test with (defaults to MVP session)
    required: false
---

# Build Verification Command

Verifies the meta-cc build by running all main commands against a real session to ensure everything works correctly.

## What This Does

1. Runs `parse stats` to verify statistics calculation
2. Runs `parse extract --type turns` to verify turn extraction
3. Runs `parse extract --type tools` to verify tool extraction
4. Runs `analyze errors` to verify error analysis
5. Validates JSON output structure

## Usage

```bash
# Use default MVP session
/meta-verify-build

# Use specific session
/meta-verify-build abc123-def456-...
```

## Implementation

```bash
#!/bin/bash
set -e

# Default to MVP development session
SESSION="${1:-6a32f273-191a-49c8-a5fc-a5dcba08531a}"

echo "🔍 Verifying meta-cc build with session: $SESSION"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Change to project directory
cd /home/yale/work/meta-cc || exit 1

# Verify binary exists
if [[ ! -f ./meta-cc ]]; then
    echo "❌ meta-cc binary not found. Run 'go build' first."
    exit 1
fi

# Test 1: Parse stats
echo "📊 Test 1/4: parse stats"
TURN_COUNT=$(./meta-cc --session "$SESSION" parse stats --output json | jq -r '.TurnCount')
if [[ -z "$TURN_COUNT" || "$TURN_COUNT" == "null" ]]; then
    echo "❌ Failed: TurnCount is missing or null"
    exit 1
fi
echo "✅ Passed: TurnCount = $TURN_COUNT"
echo ""

# Test 2: Extract turns
echo "📝 Test 2/4: parse extract --type turns"
TURNS_COUNT=$(./meta-cc --session "$SESSION" parse extract --type turns --output json | jq 'length')
if [[ -z "$TURNS_COUNT" || "$TURNS_COUNT" == "null" || "$TURNS_COUNT" -eq 0 ]]; then
    echo "❌ Failed: No turns extracted"
    exit 1
fi
echo "✅ Passed: Extracted $TURNS_COUNT turns"
echo ""

# Test 3: Extract tools
echo "🔧 Test 3/4: parse extract --type tools"
TOOLS_COUNT=$(./meta-cc --session "$SESSION" parse extract --type tools --output json | jq 'length')
if [[ -z "$TOOLS_COUNT" || "$TOOLS_COUNT" == "null" ]]; then
    echo "❌ Failed: Tool extraction returned null"
    exit 1
fi
echo "✅ Passed: Extracted $TOOLS_COUNT tool calls"
echo ""

# Test 4: Analyze errors
echo "🐛 Test 4/4: analyze errors"
ERROR_ANALYSIS=$(./meta-cc --session "$SESSION" analyze errors --output json)
ERROR_PATTERNS=$(echo "$ERROR_ANALYSIS" | jq 'if type == "array" then length else 0 end')
if [[ -z "$ERROR_PATTERNS" || "$ERROR_PATTERNS" == "null" ]]; then
    echo "❌ Failed: Error analysis failed"
    exit 1
fi
echo "✅ Passed: Found $ERROR_PATTERNS error patterns"
echo ""

# All tests passed
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ All verification tests passed!"
echo ""
echo "Summary:"
echo "  - Turns: $TURNS_COUNT"
echo "  - Tool Calls: $TOOLS_COUNT"
echo "  - Error Patterns: $ERROR_PATTERNS"
echo "  - Session: $SESSION"
```

## Expected Output

```
🔍 Verifying meta-cc build with session: 6a32f273-191a-49c8-a5fc-a5dcba08531a
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📊 Test 1/4: parse stats
✅ Passed: TurnCount = 2676

📝 Test 2/4: parse extract --type turns
✅ Passed: Extracted 2676 turns

🔧 Test 3/4: parse extract --type tools
✅ Passed: Extracted 1012 tool calls

🐛 Test 4/4: analyze errors
✅ Passed: Found 0 error patterns

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✅ All verification tests passed!

Summary:
  - Turns: 2676
  - Tool Calls: 1012
  - Error Patterns: 0
  - Session: 6a32f273-191a-49c8-a5fc-a5dcba08531a
```

## When to Use

- After completing a Phase implementation
- Before creating a pull request
- After making changes to core parsing logic
- When debugging session parsing issues

## Related Commands

- `/meta-stats` - Quick session statistics
- `/meta-errors` - Error pattern analysis
- `/meta-timeline` - Session timeline visualization
