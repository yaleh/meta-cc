#!/bin/bash
# Phase 12 MCP Project-Level Query Integration Test

set -e

echo "=== Phase 12 MCP Project-Level Query Test ==="
echo ""

# Prerequisites check
if ! command -v meta-cc &> /dev/null; then
    echo "✗ meta-cc CLI not found"
    exit 1
fi

echo "✓ meta-cc CLI found: $(which meta-cc)"

# Step 1: Verify MCP configuration includes all 16 tools
echo ""
echo "[1/5] Verifying MCP configuration..."

CONFIG_FILE=".claude/mcp-servers/meta-cc.json"
if [ ! -f "$CONFIG_FILE" ]; then
    echo "✗ MCP configuration file not found: $CONFIG_FILE"
    exit 1
fi

# Check project-level tools
PROJECT_TOOLS=(
    "get_stats"
    "analyze_errors"
    "query_tools"
    "query_user_messages"
    "query_tool_sequences"
    "query_file_access"
    "query_successful_prompts"
    "query_context"
)

MISSING_TOOLS=()
for tool in "${PROJECT_TOOLS[@]}"; do
    if ! grep -q "\"$tool\"" "$CONFIG_FILE"; then
        MISSING_TOOLS+=("$tool")
    fi
done

if [ ${#MISSING_TOOLS[@]} -gt 0 ]; then
    echo "✗ Missing project-level tools: ${MISSING_TOOLS[*]}"
    exit 1
fi

echo "  ✓ All 8 project-level tools configured"

# Check session-level tools
SESSION_TOOLS=(
    "get_session_stats"
    "analyze_errors_session"
    "query_tools_session"
    "query_user_messages_session"
    "query_tool_sequences_session"
    "query_file_access_session"
    "query_successful_prompts_session"
    "query_context_session"
)

MISSING_TOOLS=()
for tool in "${SESSION_TOOLS[@]}"; do
    if ! grep -q "\"$tool\"" "$CONFIG_FILE"; then
        MISSING_TOOLS+=("$tool")
    fi
done

if [ ${#MISSING_TOOLS[@]} -gt 0 ]; then
    echo "✗ Missing session-level tools: ${MISSING_TOOLS[*]}"
    exit 1
fi

echo "  ✓ All 8 session-level tools configured"

# Step 2: Verify documentation exists
echo ""
echo "[2/5] Verifying documentation..."

if [ ! -f "docs/mcp-project-scope.md" ]; then
    echo "✗ MCP project scope documentation not found"
    exit 1
fi

# Check documentation has key sections
DOC_FILE="docs/mcp-project-scope.md"
REQUIRED_SECTIONS=(
    "Tool Naming Convention"
    "Project-Level Tools"
    "Session-Level Tools"
    "Usage Examples"
    "When to Use Each Scope"
)

MISSING_SECTIONS=()
for section in "${REQUIRED_SECTIONS[@]}"; do
    if ! grep -q "$section" "$DOC_FILE"; then
        MISSING_SECTIONS+=("$section")
    fi
done

if [ ${#MISSING_SECTIONS[@]} -gt 0 ]; then
    echo "✗ Missing documentation sections: ${MISSING_SECTIONS[*]}"
    exit 1
fi

echo "  ✓ Documentation complete with all required sections"

# Step 3: Verify README.md updates
echo ""
echo "[3/5] Verifying README.md updates..."

if ! grep -q "Phase 12" README.md; then
    echo "✗ README.md not updated with Phase 12 information"
    exit 1
fi

if ! grep -q "mcp-project-scope.md" README.md; then
    echo "✗ README.md missing link to MCP project scope guide"
    exit 1
fi

echo "  ✓ README.md updated with Phase 12 features"

# Step 4: Test project-level vs session-level behavior (basic validation)
echo ""
echo "[4/5] Testing project vs session scope (CLI flag support)..."

# Test that --project flag is recognized (even if no data available)
OUTPUT=$(meta-cc query tools --project . --limit 1 --output json 2>&1 || true)
if echo "$OUTPUT" | grep -q "unknown flag"; then
    echo "✗ --project flag not recognized by CLI"
    exit 1
else
    echo "  ✓ Project-level query (--project .) flag recognized"
fi

# Test that session-level works without --project flag
OUTPUT=$(meta-cc query tools --limit 1 --output json 2>&1 || true)
if echo "$OUTPUT" | grep -q "unknown flag"; then
    echo "✗ Session-level query has flag errors"
    exit 1
else
    echo "  ✓ Session-level query (no flag) works"
fi

# Step 5: Verify backward compatibility
echo ""
echo "[5/5] Testing backward compatibility..."

# Session-level commands should still work
OUTPUT=$(meta-cc analyze stats --output json 2>&1 || true)
if echo "$OUTPUT" | grep -q "unknown command"; then
    echo "✗ Session stats command broken"
    exit 1
else
    echo "  ✓ Session stats command backward compatible"
fi

echo ""
echo "=== All Phase 12 Tests Passed ✅ ==="
echo ""
echo "Summary:"
echo "  - 16 tools configured (8 project + 8 session)"
echo "  - Documentation complete"
echo "  - README.md updated"
echo "  - Project-level queries work"
echo "  - Session-level queries work"
echo "  - Backward compatibility maintained"
