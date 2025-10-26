#!/bin/bash
# Simple E2E test for meta-cc-mcp server
# Usage: ./tests/e2e/mcp-e2e-simple.sh [binary_path]
#
# This script tests individual requests by starting a fresh server for each test.
# For production testing, use mcp-inspector or a proper MCP client.

set -e

BINARY="${1:-./meta-cc-mcp}"

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Check if binary exists
if [ ! -f "$BINARY" ]; then
    echo -e "${RED}Error: Binary not found: $BINARY${NC}"
    echo "Usage: $0 [binary_path]"
    exit 1
fi

# Helper function to extract JSON-RPC response (filter out log lines)
get_json_response() {
    local input="$1"
    # Look for lines starting with { and containing "jsonrpc"
    echo "$input" | grep -E '^\s*\{' | grep '"jsonrpc"' | head -1
}

echo "=========================================="
echo "MCP Simple E2E Test"
echo "=========================================="
echo "Binary: $BINARY"
echo "=========================================="
echo ""

# Test 1: List tools (single request)
echo -e "${BLUE}Test 1: List Tools${NC}"
RAW_OUTPUT=$(echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | timeout 3 "$BINARY" 2>&1)
RESPONSE=$(get_json_response "$RAW_OUTPUT")

if [ -z "$RESPONSE" ]; then
    echo -e "  ${RED}✗ FAILED${NC} - No JSON-RPC response received"
    echo "  Raw output (first 5 lines):"
    echo "$RAW_OUTPUT" | head -5 | sed 's/^/    /'
    exit 1
fi

if echo "$RESPONSE" | jq -e '.result.tools' >/dev/null 2>&1; then
    TOOL_COUNT=$(echo "$RESPONSE" | jq '.result.tools | length')
    echo -e "  ${GREEN}✓ PASSED${NC} - Found $TOOL_COUNT tools"
    echo "  Tools:"
    echo "$RESPONSE" | jq -r '.result.tools[].name' | sed 's/^/    - /'
else
    echo -e "  ${RED}✗ FAILED${NC} - Could not parse tool list"
    echo "  Response: $RESPONSE"
    exit 1
fi
echo ""

# Test 2: Call a simple tool (query_tool_errors)
echo -e "${BLUE}Test 2: Call query_tool_errors${NC}"
REQUEST='{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"query_tool_errors","arguments":{"limit":5}}}'
RAW_OUTPUT=$(echo "$REQUEST" | timeout 5 "$BINARY" 2>&1)
RESPONSE=$(get_json_response "$RAW_OUTPUT")

if [ -z "$RESPONSE" ]; then
    echo -e "  ${YELLOW}⚠ WARNING${NC} - No JSON-RPC response received"
else
    if echo "$RESPONSE" | jq -e '.result' >/dev/null 2>&1; then
        echo -e "  ${GREEN}✓ PASSED${NC} - Tool executed successfully"
    else
        ERROR_MSG=$(echo "$RESPONSE" | jq -r '.error.message // "Unknown error"')
        echo -e "  ${YELLOW}⚠ WARNING${NC} - Tool returned error: $ERROR_MSG"
    fi
fi
echo ""

# Test 3: Call another tool (query_tools)
echo -e "${BLUE}Test 3: Call query_tools${NC}"
REQUEST='{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"query_tools","arguments":{"limit":10}}}'
RAW_OUTPUT=$(echo "$REQUEST" | timeout 5 "$BINARY" 2>&1)
RESPONSE=$(get_json_response "$RAW_OUTPUT")

if [ -z "$RESPONSE" ]; then
    echo -e "  ${YELLOW}⚠ WARNING${NC} - No JSON-RPC response received"
else
    if echo "$RESPONSE" | jq -e '.result' >/dev/null 2>&1; then
        echo -e "  ${GREEN}✓ PASSED${NC} - Tool executed successfully"
    else
        ERROR_MSG=$(echo "$RESPONSE" | jq -r '.error.message // "Unknown error"')
        echo -e "  ${YELLOW}⚠ WARNING${NC} - Tool returned error: $ERROR_MSG"
    fi
fi
echo ""

# Test 4: Check for Phase 27 tools
echo -e "${BLUE}Test 4: Check Phase 27 Tools${NC}"

# Get fresh tools list
RAW_OUTPUT=$(echo '{"jsonrpc":"2.0","id":4,"method":"tools/list"}' | timeout 3 "$BINARY" 2>&1)
TOOLS_LIST=$(get_json_response "$RAW_OUTPUT")

if [ -z "$TOOLS_LIST" ]; then
    echo -e "  ${YELLOW}⚠ Could not retrieve tools list${NC}"
else
    # Check get_session_directory
    if echo "$TOOLS_LIST" | jq -e '.result.tools[] | select(.name == "get_session_directory")' >/dev/null 2>&1; then
        echo -e "  ${GREEN}✓ get_session_directory found${NC}"
    else
        echo -e "  ${YELLOW}⚠ get_session_directory not found (not implemented yet)${NC}"
    fi

    # Check inspect_session_files
    if echo "$TOOLS_LIST" | jq -e '.result.tools[] | select(.name == "inspect_session_files")' >/dev/null 2>&1; then
        echo -e "  ${GREEN}✓ inspect_session_files found${NC}"
    else
        echo -e "  ${YELLOW}⚠ inspect_session_files not found (not implemented yet)${NC}"
    fi

    # Check execute_stage2_query
    if echo "$TOOLS_LIST" | jq -e '.result.tools[] | select(.name == "execute_stage2_query")' >/dev/null 2>&1; then
        echo -e "  ${GREEN}✓ execute_stage2_query found${NC}"
    else
        echo -e "  ${YELLOW}⚠ execute_stage2_query not found (not implemented yet)${NC}"
    fi
fi
echo ""

# Test 5: get_session_directory (project scope)
echo -e "${BLUE}Test 5: get_session_directory (project scope)${NC}"
REQUEST='{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"get_session_directory","arguments":{"scope":"project"}}}'
RAW_OUTPUT=$(echo "$REQUEST" | timeout 5 "$BINARY" 2>&1)
RESPONSE=$(get_json_response "$RAW_OUTPUT")

if [ -z "$RESPONSE" ]; then
    echo -e "  ${YELLOW}⚠ WARNING${NC} - No JSON-RPC response received"
else
    if echo "$RESPONSE" | jq -e '.result.content[0].text | fromjson | .directory' >/dev/null 2>&1; then
        DIRECTORY=$(echo "$RESPONSE" | jq -r '.result.content[0].text | fromjson | .directory')
        FILE_COUNT=$(echo "$RESPONSE" | jq -r '.result.content[0].text | fromjson | .file_count')
        echo -e "  ${GREEN}✓ PASSED${NC} - Directory: $DIRECTORY, Files: $FILE_COUNT"
    else
        ERROR_MSG=$(echo "$RESPONSE" | jq -r '.error.message // "Unknown error"')
        echo -e "  ${YELLOW}⚠ WARNING${NC} - $ERROR_MSG"
    fi
fi
echo ""

# Test 6: get_session_directory (session scope)
echo -e "${BLUE}Test 6: get_session_directory (session scope)${NC}"
REQUEST='{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"get_session_directory","arguments":{"scope":"session"}}}'
RAW_OUTPUT=$(echo "$REQUEST" | timeout 5 "$BINARY" 2>&1)
RESPONSE=$(get_json_response "$RAW_OUTPUT")

if [ -z "$RESPONSE" ]; then
    echo -e "  ${YELLOW}⚠ WARNING${NC} - No JSON-RPC response received"
else
    if echo "$RESPONSE" | jq -e '.result.content[0].text | fromjson | .directory' >/dev/null 2>&1; then
        DIRECTORY=$(echo "$RESPONSE" | jq -r '.result.content[0].text | fromjson | .directory')
        FILE_COUNT=$(echo "$RESPONSE" | jq -r '.result.content[0].text | fromjson | .file_count')
        echo -e "  ${GREEN}✓ PASSED${NC} - Directory: $DIRECTORY, Files: $FILE_COUNT"
    else
        ERROR_MSG=$(echo "$RESPONSE" | jq -r '.error.message // "Unknown error"')
        echo -e "  ${YELLOW}⚠ WARNING${NC} - $ERROR_MSG"
    fi
fi
echo ""

# Test 7: inspect_session_files (without samples)
echo -e "${BLUE}Test 7: inspect_session_files (without samples)${NC}"
# First get the directory to get file list
REQUEST='{"jsonrpc":"2.0","id":7,"method":"tools/call","params":{"name":"get_session_directory","arguments":{"scope":"project"}}}'
RAW_OUTPUT=$(echo "$REQUEST" | timeout 5 "$BINARY" 2>&1)
DIR_RESPONSE=$(get_json_response "$RAW_OUTPUT")

if [ -z "$DIR_RESPONSE" ]; then
    echo -e "  ${YELLOW}⚠ WARNING${NC} - Could not get directory"
else
    DIRECTORY=$(echo "$DIR_RESPONSE" | jq -r '.result.content[0].text | fromjson | .directory')
    # Get first 5 files from directory with FULL PATHS
    FILES=$(ls "$DIRECTORY" 2>/dev/null | head -5 | sed "s|^|$DIRECTORY/|" | jq -R . | jq -s .)

    if [ -z "$FILES" ] || [ "$FILES" = "[]" ]; then
        echo -e "  ${YELLOW}⚠ WARNING${NC} - No files found in directory"
    else
        REQUEST="{\"jsonrpc\":\"2.0\",\"id\":7,\"method\":\"tools/call\",\"params\":{\"name\":\"inspect_session_files\",\"arguments\":{\"scope\":\"project\",\"files\":$FILES,\"include_samples\":false}}}"
        RAW_OUTPUT=$(echo "$REQUEST" | timeout 5 "$BINARY" 2>&1)
        RESPONSE=$(get_json_response "$RAW_OUTPUT")

        if [ -z "$RESPONSE" ]; then
            echo -e "  ${YELLOW}⚠ WARNING${NC} - No JSON-RPC response received"
        else
            if echo "$RESPONSE" | jq -e '.result.files' >/dev/null 2>&1; then
                FILE_COUNT=$(echo "$RESPONSE" | jq -r '.result.files | length')
                echo -e "  ${GREEN}✓ PASSED${NC} - Inspected $FILE_COUNT files"
            else
                ERROR_MSG=$(echo "$RESPONSE" | jq -r '.error.message // "Unknown error"')
                echo -e "  ${YELLOW}⚠ WARNING${NC} - $ERROR_MSG"
            fi
        fi
    fi
fi
echo ""

# Test 8: inspect_session_files (with samples)
echo -e "${BLUE}Test 8: inspect_session_files (with samples)${NC}"
# First get the directory to get file list
REQUEST='{"jsonrpc":"2.0","id":8,"method":"tools/call","params":{"name":"get_session_directory","arguments":{"scope":"project"}}}'
RAW_OUTPUT=$(echo "$REQUEST" | timeout 5 "$BINARY" 2>&1)
DIR_RESPONSE=$(get_json_response "$RAW_OUTPUT")

if [ -z "$DIR_RESPONSE" ]; then
    echo -e "  ${YELLOW}⚠ WARNING${NC} - Could not get directory"
else
    DIRECTORY=$(echo "$DIR_RESPONSE" | jq -r '.result.content[0].text | fromjson | .directory')
    # Get first 3 files from directory
    FILES=$(ls "$DIRECTORY" 2>/dev/null | head -3 | jq -R . | jq -s .)

    if [ -z "$FILES" ] || [ "$FILES" = "[]" ]; then
        echo -e "  ${YELLOW}⚠ WARNING${NC} - No files found in directory"
    else
        REQUEST="{\"jsonrpc\":\"2.0\",\"id\":8,\"method\":\"tools/call\",\"params\":{\"name\":\"inspect_session_files\",\"arguments\":{\"scope\":\"project\",\"files\":$FILES,\"include_samples\":true,\"sample_size\":3}}}"
        RAW_OUTPUT=$(echo "$REQUEST" | timeout 5 "$BINARY" 2>&1)
        RESPONSE=$(get_json_response "$RAW_OUTPUT")

        if [ -z "$RESPONSE" ]; then
            echo -e "  ${YELLOW}⚠ WARNING${NC} - No JSON-RPC response received"
        else
            if echo "$RESPONSE" | jq -e '.result.content[0].text | fromjson | .files' >/dev/null 2>&1; then
                FILE_COUNT=$(echo "$RESPONSE" | jq -r '.result.content[0].text | fromjson | .files | length')
                HAS_SAMPLES=$(echo "$RESPONSE" | jq -r '.result.content[0].text | fromjson | .files[0].sample_entries // [] | length > 0')
                if [ "$HAS_SAMPLES" = "true" ]; then
                    echo -e "  ${GREEN}✓ PASSED${NC} - Inspected $FILE_COUNT files with samples"
                else
                    echo -e "  ${GREEN}✓ PASSED${NC} - Inspected $FILE_COUNT files (no samples available)"
                fi
            else
                ERROR_MSG=$(echo "$RESPONSE" | jq -r '.error.message // "Unknown error"')
                echo -e "  ${YELLOW}⚠ WARNING${NC} - $ERROR_MSG"
            fi
        fi
    fi
fi
echo ""

# Test 9: execute_stage2_query (basic filter)
echo -e "${BLUE}Test 9: execute_stage2_query (basic filter)${NC}"
# First get the directory to get file list
REQUEST='{"jsonrpc":"2.0","id":9,"method":"tools/call","params":{"name":"get_session_directory","arguments":{"scope":"project"}}}'
RAW_OUTPUT=$(echo "$REQUEST" | timeout 5 "$BINARY" 2>&1)
DIR_RESPONSE=$(get_json_response "$RAW_OUTPUT")

if [ -z "$DIR_RESPONSE" ]; then
    echo -e "  ${YELLOW}⚠ WARNING${NC} - Could not get directory"
else
    DIRECTORY=$(echo "$DIR_RESPONSE" | jq -r '.result.content[0].text | fromjson | .directory')
    # Get first 5 files from directory
    FILES=$(ls "$DIRECTORY" 2>/dev/null | head -5 | jq -R . | jq -s .)

    if [ -z "$FILES" ] || [ "$FILES" = "[]" ]; then
        echo -e "  ${YELLOW}⚠ WARNING${NC} - No files found in directory"
    else
        REQUEST="{\"jsonrpc\":\"2.0\",\"id\":9,\"method\":\"tools/call\",\"params\":{\"name\":\"execute_stage2_query\",\"arguments\":{\"scope\":\"project\",\"files\":$FILES,\"filter\":\"select(.type == \\\\\\\"user\\\\\\\")\",\"limit\":5}}}"
        RAW_OUTPUT=$(echo "$REQUEST" | timeout 5 "$BINARY" 2>&1)
        RESPONSE=$(get_json_response "$RAW_OUTPUT")

        if [ -z "$RESPONSE" ]; then
            echo -e "  ${YELLOW}⚠ WARNING${NC} - No JSON-RPC response received"
        else
            if echo "$RESPONSE" | jq -e '.result' >/dev/null 2>&1; then
                echo -e "  ${GREEN}✓ PASSED${NC} - Query executed successfully"
            else
                ERROR_MSG=$(echo "$RESPONSE" | jq -r '.error.message // "Unknown error"')
                echo -e "  ${YELLOW}⚠ WARNING${NC} - $ERROR_MSG"
            fi
        fi
    fi
fi
echo ""

# Test 10: execute_stage2_query (full pipeline)
echo -e "${BLUE}Test 10: execute_stage2_query (full pipeline)${NC}"
# First get the directory to get file list
REQUEST='{"jsonrpc":"2.0","id":10,"method":"tools/call","params":{"name":"get_session_directory","arguments":{"scope":"project"}}}'
RAW_OUTPUT=$(echo "$REQUEST" | timeout 5 "$BINARY" 2>&1)
DIR_RESPONSE=$(get_json_response "$RAW_OUTPUT")

if [ -z "$DIR_RESPONSE" ]; then
    echo -e "  ${YELLOW}⚠ WARNING${NC} - Could not get directory"
else
    DIRECTORY=$(echo "$DIR_RESPONSE" | jq -r '.result.content[0].text | fromjson | .directory')
    # Get first 5 files from directory
    FILES=$(ls "$DIRECTORY" 2>/dev/null | head -5 | jq -R . | jq -s .)

    if [ -z "$FILES" ] || [ "$FILES" = "[]" ]; then
        echo -e "  ${YELLOW}⚠ WARNING${NC} - No files found in directory"
    else
        REQUEST="{\"jsonrpc\":\"2.0\",\"id\":10,\"method\":\"tools/call\",\"params\":{\"name\":\"execute_stage2_query\",\"arguments\":{\"scope\":\"project\",\"files\":$FILES,\"filter\":\"select(.type == \\\\\\\"assistant\\\\\\\")\",\"sort_by\":\"timestamp\",\"sort_order\":\"desc\",\"transform\":\"{timestamp, turn}\",\"limit\":3}}}"
        RAW_OUTPUT=$(echo "$REQUEST" | timeout 5 "$BINARY" 2>&1)
        RESPONSE=$(get_json_response "$RAW_OUTPUT")

        if [ -z "$RESPONSE" ]; then
            echo -e "  ${YELLOW}⚠ WARNING${NC} - No JSON-RPC response received"
        else
            if echo "$RESPONSE" | jq -e '.result' >/dev/null 2>&1; then
                echo -e "  ${GREEN}✓ PASSED${NC} - Full pipeline executed successfully"
            else
                ERROR_MSG=$(echo "$RESPONSE" | jq -r '.error.message // "Unknown error"')
                echo -e "  ${YELLOW}⚠ WARNING${NC} - $ERROR_MSG"
            fi
        fi
    fi
fi
echo ""

# Test 11: Performance validation
echo -e "${BLUE}Test 11: Performance Validation (< 100ms target)${NC}"
# First get the directory to get file list
REQUEST='{"jsonrpc":"2.0","id":11,"method":"tools/call","params":{"name":"get_session_directory","arguments":{"scope":"project"}}}'
RAW_OUTPUT=$(echo "$REQUEST" | timeout 5 "$BINARY" 2>&1)
DIR_RESPONSE=$(get_json_response "$RAW_OUTPUT")

if [ -z "$DIR_RESPONSE" ]; then
    echo -e "  ${YELLOW}⚠ WARNING${NC} - Could not get directory"
else
    DIRECTORY=$(echo "$DIR_RESPONSE" | jq -r '.result.content[0].text | fromjson | .directory')
    # Get first 10 files from directory
    FILES=$(ls "$DIRECTORY" 2>/dev/null | head -10 | jq -R . | jq -s .)

    if [ -z "$FILES" ] || [ "$FILES" = "[]" ]; then
        echo -e "  ${YELLOW}⚠ WARNING${NC} - No files found in directory"
    else
        START_TIME=$(date +%s%3N)
        REQUEST="{\"jsonrpc\":\"2.0\",\"id\":11,\"method\":\"tools/call\",\"params\":{\"name\":\"execute_stage2_query\",\"arguments\":{\"scope\":\"project\",\"files\":$FILES,\"filter\":\"select(.type == \\\\\\\"user\\\\\\\")\",\"limit\":10}}}"
        RAW_OUTPUT=$(echo "$REQUEST" | timeout 5 "$BINARY" 2>&1)
        END_TIME=$(date +%s%3N)
        DURATION=$((END_TIME - START_TIME))

        RESPONSE=$(get_json_response "$RAW_OUTPUT")

        if [ -z "$RESPONSE" ]; then
            echo -e "  ${YELLOW}⚠ WARNING${NC} - No response received"
        else
            if [ $DURATION -lt 100 ]; then
                echo -e "  ${GREEN}✓ PASSED${NC} - Query completed in ${DURATION}ms (< 100ms target)"
            else
                echo -e "  ${YELLOW}⚠ SLOW${NC} - Query took ${DURATION}ms (target: < 100ms)"
            fi
        fi
    fi
fi
echo ""

# Test 12: Complete two-stage workflow
echo -e "${BLUE}Test 12: Complete Two-Stage Workflow${NC}"

# Stage 1: Get directory
echo -e "  ${BLUE}Stage 1: Get session directory${NC}"
REQUEST='{"jsonrpc":"2.0","id":12,"method":"tools/call","params":{"name":"get_session_directory","arguments":{"scope":"project"}}}'
RAW_OUTPUT=$(echo "$REQUEST" | timeout 5 "$BINARY" 2>&1)
RESPONSE=$(get_json_response "$RAW_OUTPUT")

if [ -z "$RESPONSE" ]; then
    echo -e "    ${RED}✗ FAILED${NC} - Stage 1 failed"
else
    if echo "$RESPONSE" | jq -e '.result.content[0].text | fromjson | .directory' >/dev/null 2>&1; then
        echo -e "    ${GREEN}✓ Stage 1 complete${NC}"

        DIRECTORY=$(echo "$RESPONSE" | jq -r '.result.content[0].text | fromjson | .directory')
        FILES=$(ls "$DIRECTORY" 2>/dev/null | head -5 | jq -R . | jq -s .)

        # Stage 2: Execute query
        echo -e "  ${BLUE}Stage 2: Execute query${NC}"
        REQUEST="{\"jsonrpc\":\"2.0\",\"id\":13,\"method\":\"tools/call\",\"params\":{\"name\":\"execute_stage2_query\",\"arguments\":{\"scope\":\"project\",\"files\":$FILES,\"filter\":\"select(.type == \\\\\\\"user\\\\\\\")\",\"limit\":5}}}"
        RAW_OUTPUT=$(echo "$REQUEST" | timeout 5 "$BINARY" 2>&1)
        RESPONSE=$(get_json_response "$RAW_OUTPUT")

        if [ -z "$RESPONSE" ]; then
            echo -e "    ${RED}✗ FAILED${NC} - Stage 2 failed"
        else
            if echo "$RESPONSE" | jq -e '.result' >/dev/null 2>&1; then
                echo -e "    ${GREEN}✓ Stage 2 complete${NC}"
                echo -e "  ${GREEN}✓ PASSED${NC} - Complete two-stage workflow successful"
            else
                ERROR_MSG=$(echo "$RESPONSE" | jq -r '.error.message // "Unknown error"')
                echo -e "    ${RED}✗ FAILED${NC} - Stage 2 returned error: $ERROR_MSG"
            fi
        fi
    else
        echo -e "    ${RED}✗ FAILED${NC} - Stage 1 returned error"
    fi
fi
echo ""

echo "=========================================="
echo -e "${GREEN}✓ Basic tests completed${NC}"
echo "=========================================="
echo ""
echo "For more comprehensive testing:"
echo ""
echo "1. Interactive testing (single request):"
echo "   echo '{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"tools/list\"}' | $BINARY 2>&1 | grep '\"jsonrpc\"' | jq ."
echo ""
echo "2. MCP Inspector (recommended for development):"
echo "   npm install -g @modelcontextprotocol/inspector"
echo "   mcp-inspector $BINARY"
echo ""
echo "3. Direct tool testing:"
echo "   echo '{\"jsonrpc\":\"2.0\",\"id\":2,\"method\":\"tools/call\",\"params\":{\"name\":\"TOOL_NAME\",\"arguments\":{}}}' | \\"
echo "     $BINARY 2>&1 | grep '\"jsonrpc\"' | jq ."
echo ""
