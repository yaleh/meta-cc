#!/bin/bash
# validate-meta-cc.sh - Unix-based validation script for meta-cc
# Validates meta-cc output by processing raw JSONL files with traditional Unix tools
#
# Usage:
#   ./validate-meta-cc.sh <command> <path> [args...]
#
# Where <path> can be:
#   - A directory (project scope): Process all .jsonl files in directory
#   - A file (session scope): Process single .jsonl file
#
# Commands:
#   stats <path>                   - Calculate session statistics
#   errors <path> [window]         - Analyze error patterns
#   query-tools <path> [filter]    - Query tool calls
#   query-messages <path> [pattern]- Search user messages
#   timeline <path> [limit]        - Generate timeline view

set -euo pipefail

VERSION="1.0.0"
SCRIPT_NAME=$(basename "$0")

# Color output helpers
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
error() {
    echo -e "${RED}❌ Error: $1${NC}" >&2
    exit 1
}

warn() {
    echo -e "${YELLOW}⚠️  Warning: $1${NC}" >&2
}

info() {
    echo -e "${BLUE}ℹ️  $1${NC}" >&2
}

success() {
    echo -e "${GREEN}✅ $1${NC}" >&2
}

# Check if required tools are available
check_dependencies() {
    local missing_tools=()

    for tool in jq grep sed awk bc; do
        if ! command -v "$tool" &> /dev/null; then
            missing_tools+=("$tool")
        fi
    done

    if [ ${#missing_tools[@]} -gt 0 ]; then
        error "Missing required tools: ${missing_tools[*]}"
    fi
}

# Determine scope (session vs project) and get JSONL files
get_jsonl_files() {
    local path="$1"

    if [ ! -e "$path" ]; then
        error "Path does not exist: $path"
    fi

    if [ -f "$path" ]; then
        # Session scope: single file
        if [[ "$path" != *.jsonl ]]; then
            error "File must be a .jsonl file: $path"
        fi
        echo "$path"
    elif [ -d "$path" ]; then
        # Project scope: get latest .jsonl file by modification time
        local files=$(ls -t "$path"/*.jsonl 2>/dev/null | head -1)
        if [ -z "$files" ]; then
            error "No .jsonl files found in directory: $path"
        fi
        echo "$files"
    else
        error "Invalid path type: $path"
    fi
}

# Show usage
usage() {
    cat <<EOF
Usage: $SCRIPT_NAME <command> <path> [args...]

Validate meta-cc output using traditional Unix tools.

ARGUMENTS:
  <path>        Path to session file (.jsonl) or project directory

COMMANDS:
  stats <path>
      Calculate session statistics

  errors <path> [window]
      Analyze error patterns
      Args: window - Number of recent turns to analyze (default: 20)

  query-tools <path> [filter] [limit]
      Query tool calls
      Args: filter - Tool name or status (default: all)
            limit - Max results (default: 20)

  query-messages <path> [pattern] [limit]
      Search user messages with regex
      Args: pattern - Regex pattern (default: .*)
            limit - Max results (default: 10)

  timeline <path> [limit]
      Generate timeline view
      Args: limit - Number of recent turns (default: 50)

EXAMPLES:
  # Session scope (single file)
  $SCRIPT_NAME stats ~/.claude/projects/abc-def/session-123.jsonl

  # Project scope (directory - uses latest session)
  $SCRIPT_NAME stats ~/.claude/projects/abc-def
  $SCRIPT_NAME errors . 30
  $SCRIPT_NAME query-tools . Bash 10

OPTIONS:
  -h, --help    Show this help message
  -v, --version Show version

EOF
}

# Main command dispatcher
main() {
    # Check dependencies first
    check_dependencies

    # Parse arguments
    if [ $# -eq 0 ]; then
        usage
        exit 0
    fi

    case "$1" in
        -h|--help)
            usage
            exit 0
            ;;
        -v|--version)
            echo "$SCRIPT_NAME version $VERSION"
            exit 0
            ;;
        stats)
            shift
            cmd_stats "$@"
            ;;
        errors)
            shift
            cmd_errors "$@"
            ;;
        query-tools)
            shift
            cmd_query_tools "$@"
            ;;
        query-messages)
            shift
            cmd_query_messages "$@"
            ;;
        timeline)
            shift
            cmd_timeline "$@"
            ;;
        *)
            error "Unknown command: $1. Use --help for usage."
            ;;
    esac
}

# Command implementations

# stats - Calculate session statistics
cmd_stats() {
    if [ $# -lt 1 ]; then
        error "Usage: $SCRIPT_NAME stats <path>"
    fi

    local path="$1"
    local jsonl_file=$(get_jsonl_files "$path")

    info "Processing: $jsonl_file"

    # Extract all message entries (filter by type: "user" or "assistant")
    local entries=$(jq -c 'select(.type == "user" or .type == "assistant")' "$jsonl_file")

    # Count turns
    local total_turns=$(echo "$entries" | wc -l)
    local user_turns=$(echo "$entries" | jq -s '[.[] | select(.type == "user")] | length')
    local assistant_turns=$(echo "$entries" | jq -s '[.[] | select(.type == "assistant")] | length')

    # Calculate duration (first to last timestamp)
    local first_ts=$(echo "$entries" | head -1 | jq -r '.timestamp // empty')
    local last_ts=$(echo "$entries" | tail -1 | jq -r '.timestamp // empty')
    local duration_seconds=0

    if [ -n "$first_ts" ] && [ -n "$last_ts" ]; then
        local first_epoch=$(date -d "$first_ts" +%s 2>/dev/null || echo 0)
        local last_epoch=$(date -d "$last_ts" +%s 2>/dev/null || echo 0)
        duration_seconds=$((last_epoch - first_epoch))
    fi

    # Extract tool calls from assistant message content blocks
    local tool_calls=$(echo "$entries" | jq -c '
        select(.type == "assistant") |
        .message.content[]? |
        select(.type == "tool_use" or .type == "tool_result") |
        {
            type: .type,
            name: (.name // .tool_name // "unknown"),
            id: .id,
            status: (if .is_error == true then "error" elif .error then "error" else "success" end)
        }
    ')

    local tool_call_count=$(echo "$tool_calls" | jq -s '[.[] | select(.type == "tool_use")] | length')
    local error_count=$(echo "$tool_calls" | jq -s '[.[] | select(.status == "error")] | length')
    local error_rate=0

    if [ "$tool_call_count" -gt 0 ]; then
        error_rate=$(echo "scale=2; ($error_count * 100) / $tool_call_count" | bc)
    fi

    # Calculate tool frequency
    local tool_frequency=$(echo "$tool_calls" | jq -s '
        [.[] | select(.type == "tool_use")] |
        group_by(.name) |
        map({
            key: .[0].name,
            value: length
        }) |
        sort_by(.value) |
        reverse
    ')

    # Calculate top tools with percentages
    local top_tools=$(echo "$tool_calls" | jq -s --argjson total "$tool_call_count" '
        [.[] | select(.type == "tool_use")] |
        group_by(.name) |
        map({
            Name: .[0].name,
            Count: length,
            Percentage: (if $total > 0 then ((length / $total) * 100 | floor) else 0 end)
        }) |
        sort_by(.Count) |
        reverse
    ')

    # Output JSONL format
    jq -n \
        --argjson turn_count "$total_turns" \
        --argjson user_turn_count "$user_turns" \
        --argjson assistant_turn_count "$assistant_turns" \
        --argjson tool_call_count "$tool_call_count" \
        --argjson error_count "$error_count" \
        --argjson error_rate "$error_rate" \
        --argjson duration_seconds "$duration_seconds" \
        --argjson tool_frequency "$tool_frequency" \
        --argjson top_tools "$top_tools" \
        '{
            TurnCount: $turn_count,
            UserTurnCount: $user_turn_count,
            AssistantTurnCount: $assistant_turn_count,
            ToolCallCount: $tool_call_count,
            ErrorCount: $error_count,
            ErrorRate: $error_rate,
            DurationSeconds: $duration_seconds,
            ToolFrequency: ($tool_frequency | from_entries),
            TopTools: $top_tools
        }'
}

cmd_errors() {
    error "Command 'errors' not implemented yet"
}

cmd_query_tools() {
    if [ $# -lt 1 ]; then
        error "Usage: $SCRIPT_NAME query-tools <path> [filter] [limit]"
    fi

    local path="$1"
    local filter="${2:-}"
    local limit="${3:-20}"
    local jsonl_file=$(get_jsonl_files "$path")

    info "Processing: $jsonl_file"
    info "Filter: ${filter:-all}, Limit: $limit"

    # Step 1: Extract all tool_use entries with their context
    local tool_use_map=$(jq -c '
        select(.type == "assistant") |
        .uuid as $uuid |
        .timestamp as $ts |
        .message.content[]? |
        select(.type == "tool_use") |
        {
            id: .id,
            UUID: $uuid,
            Timestamp: $ts,
            ToolName: .name,
            Input: .input
        }
    ' "$jsonl_file")

    # Step 2: Extract all tool_result entries
    local tool_result_map=$(jq -c '
        select(.type == "user") |
        .message.content[]? |
        select(.type == "tool_result") |
        {
            tool_use_id: .tool_use_id,
            Output: (.content // ""),
            Error: (if .is_error == true then (.content // "error") else "" end),
            Status: (if .is_error == true then "error" else "" end)
        }
    ' "$jsonl_file")

    # Step 3: Merge tool_use and tool_result by ID
    # This is complex in bash, so we'll use jq to do the join
    local merged=$(jq -n \
        --argjson tool_uses "$(echo "$tool_use_map" | jq -s '.')" \
        --argjson tool_results "$(echo "$tool_result_map" | jq -s '.')" \
        '
        # Create lookup map for tool_results by tool_use_id
        ($tool_results | map({key: .tool_use_id, value: .}) | from_entries) as $result_map |
        # Iterate over tool_uses and merge with results
        $tool_uses | map(
            . as $use |
            ($result_map[$use.id] // {Output: "", Error: "", Status: ""}) as $result |
            {
                UUID: $use.UUID,
                ToolName: $use.ToolName,
                Input: $use.Input,
                Output: $result.Output,
                Status: $result.Status,
                Error: $result.Error,
                Timestamp: $use.Timestamp
            }
        )
    ')

    # Apply filter if provided
    if [ -n "$filter" ]; then
        merged=$(echo "$merged" | jq -c --arg filter "$filter" '
            map(select(.ToolName == $filter)) | .[]
        ')
    else
        merged=$(echo "$merged" | jq -c '.[]')
    fi

    # Apply limit
    local result=$(echo "$merged" | head -n "$limit")

    # Output as JSONL (one object per line)
    echo "$result"
}

cmd_query_messages() {
    if [ $# -lt 1 ]; then
        error "Usage: $SCRIPT_NAME query-messages <path> [pattern] [limit]"
    fi

    local path="$1"
    local pattern="${2:-.*}"
    local limit="${3:-10}"
    local jsonl_file=$(get_jsonl_files "$path")

    info "Processing: $jsonl_file"
    info "Pattern: $pattern, Limit: $limit"

    # Step 1: Build turn index (only count user/assistant messages)
    local turn_index=$(jq -c '
        select(.type == "user" or .type == "assistant") |
        .uuid
    ' "$jsonl_file" | \
    awk '{print NR "\t" $0}' | \
    jq -R -r 'split("\t") | {turn: (.[0] | tonumber), uuid: (.[1] | fromjson)}' | \
    jq -s 'map({key: .uuid, value: .turn}) | from_entries')

    # Step 2: Extract user messages, concatenate text blocks, filter by pattern
    local messages=$(jq -c \
        --arg pattern "$pattern" \
        --argjson turn_idx "$turn_index" \
        '
        select(.type == "user") |
        .uuid as $uuid |
        .timestamp as $ts |
        (
            if (.message.content | type) == "string" then
                # content is string - use directly
                .message.content
            else
                # content is array - concatenate text blocks
                [.message.content[]? | select(.type == "text") | .text] | join("")
            end
        ) as $content |
        select($content != "" and ($content | test($pattern))) |
        {
            turn_sequence: $turn_idx[$uuid],
            uuid: $uuid,
            timestamp: $ts,
            content: $content
        }
        ' "$jsonl_file")

    # Apply limit and output as JSONL
    echo "$messages" | head -n "$limit"
}

cmd_timeline() {
    error "Command 'timeline' not implemented yet"
}

# Run main
main "$@"
