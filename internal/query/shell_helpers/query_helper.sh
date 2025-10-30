#!/bin/bash

# Generic shell helper for query tools
# Usage: query_helper.sh [options] <tool_name> <session_directory>
# Options:
#   -p, --pattern PATTERN    Regex pattern to match (for query_user_messages)
#   -t, --tool TOOL_NAME     Tool name to filter (for query_tools)
#   -k, --keyword KEYWORD    Keyword to search (for query_summaries)
#   -b, --block-type TYPE    Block type: tool_use or tool_result (for query_tool_blocks)
#   -l, --limit LIMIT        Maximum number of results to return
#   -s, --scope SCOPE        Scope: project or session (default: project)

set -e

# Default values
PATTERN=""
TOOL_NAME=""
KEYWORD=""
BLOCK_TYPE=""
LIMIT=0
SCOPE="project"
SESSION_DIR=""
TOOL_NAME_PARAM=""

# Function to display usage
usage() {
    echo "Usage: $0 [options] <tool_name> <session_directory>"
    echo "Tool names: query_user_messages, query_tool_errors, query_token_usage, query_conversation_flow,"
    echo "            query_system_errors, query_file_snapshots, query_timestamps, query_summaries,"
    echo "            query_tool_blocks, query_tools"
    echo ""
    echo "Options:"
    echo "  -p, --pattern PATTERN    Regex pattern to match (for query_user_messages)"
    echo "  -t, --tool TOOL_NAME     Tool name to filter (for query_tools)"
    echo "  -k, --keyword KEYWORD    Keyword to search (for query_summaries)"
    echo "  -b, --block-type TYPE    Block type: tool_use or tool_result (for query_tool_blocks)"
    echo "  -l, --limit LIMIT        Maximum number of results to return"
    echo "  -s, --scope SCOPE        Scope: project or session (default: project)"
    echo "  -h, --help               Display this help message"
    exit 1
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -p|--pattern)
            PATTERN="$2"
            shift 2
            ;;
        -t|--tool)
            TOOL_NAME="$2"
            shift 2
            ;;
        -k|--keyword)
            KEYWORD="$2"
            shift 2
            ;;
        -b|--block-type)
            BLOCK_TYPE="$2"
            shift 2
            ;;
        -l|--limit)
            LIMIT="$2"
            shift 2
            ;;
        -s|--scope)
            SCOPE="$2"
            shift 2
            ;;
        -h|--help)
            usage
            ;;
        -*)
            echo "Unknown option $1"
            usage
            ;;
        *)
            if [ -z "$TOOL_NAME_PARAM" ]; then
                TOOL_NAME_PARAM="$1"
            elif [ -z "$SESSION_DIR" ]; then
                SESSION_DIR="$1"
            else
                echo "Too many arguments"
                usage
            fi
            shift
            ;;
    esac
done

# Check if tool name and session directory are provided
if [ -z "$TOOL_NAME_PARAM" ] || [ -z "$SESSION_DIR" ]; then
    echo "Error: Tool name and session directory are required"
    usage
fi

# Check if session directory exists
if [ ! -d "$SESSION_DIR" ]; then
    echo "Error: Session directory does not exist: $SESSION_DIR"
    exit 1
fi

# Function to get jq filter based on tool name
get_jq_filter() {
    local tool="$1"
    local filter=""

    case "$tool" in
        query_user_messages)
            filter='select(.type == "user")'
            if [ -n "$PATTERN" ]; then
                filter="$filter | select(.message.content | test(\"$PATTERN\"))"
            fi
            ;;
        query_tool_errors)
            filter='select(.type == "user" and (.message.content | type == "array")) | select(.message.content[] | select(.type == "tool_result" and .is_error == true))'
            ;;
        query_token_usage)
            filter='select(.type == "assistant" and has("message")) | select(.message | has("usage"))'
            ;;
        query_conversation_flow)
            filter='select(.type == "user" or .type == "assistant")'
            ;;
        query_system_errors)
            filter='select(.type == "system" and .subtype == "api_error")'
            ;;
        query_file_snapshots)
            filter='select(.type == "file-history-snapshot" and has("messageId"))'
            ;;
        query_timestamps)
            filter='select(.timestamp != null)'
            ;;
        query_summaries)
            filter='select(.type == "summary")'
            if [ -n "$KEYWORD" ]; then
                filter="$filter | select(.summary | test(\"$KEYWORD\"; \"i\"))"
            fi
            ;;
        query_tool_blocks)
            if [ "$BLOCK_TYPE" = "tool_use" ]; then
                filter='select(.type == "assistant") | .message.content[] | select(.type == "tool_use")'
            elif [ "$BLOCK_TYPE" = "tool_result" ]; then
                filter='select(.type == "user" and (.message.content | type == "array")) | .message.content[] | select(.type == "tool_result")'
            else
                filter='select(.type == "assistant") | .message.content[] | select(.type == "tool_use")'
            fi
            ;;
        query_tools)
            filter='select(.type == "assistant") | select(.message.content[] | .type == "tool_use")'
            if [ -n "$TOOL_NAME" ]; then
                filter="$filter | select(.message.content[] | select(.type == \"tool_use\" and .name == \"$TOOL_NAME\"))"
            fi
            ;;
        *)
            echo "Error: Unknown tool name: $tool"
            exit 1
            ;;
    esac

    echo "$filter"
}

# Get the jq filter for the specified tool
FILTER=$(get_jq_filter "$TOOL_NAME_PARAM")

# Find JSONL files based on scope
if [ "$SCOPE" = "session" ]; then
    # For session scope, use only the most recent file
    JSONL_FILES=$(find "$SESSION_DIR" -name "*.jsonl" -type f -printf '%T@ %p\n' | sort -n | tail -1 | cut -d' ' -f2-)
else
    # For project scope, use all files
    JSONL_FILES=$(find "$SESSION_DIR" -name "*.jsonl" -type f)
fi

# Process files
COUNT=0
for file in $JSONL_FILES; do
    if [ -f "$file" ]; then
        # Process the file
        while IFS= read -r line; do
            # Check if we've reached the limit
            if [ $LIMIT -gt 0 ] && [ $COUNT -ge $LIMIT ]; then
                break 2
            fi

            # Apply the filter using jq
            echo "$line" | jq -e "$FILTER" >/dev/null 2>&1
            if [ $? -eq 0 ]; then
                echo "$line" | jq -c "$FILTER"
                COUNT=$((COUNT + 1))
            fi
        done < "$file"
    fi
done
