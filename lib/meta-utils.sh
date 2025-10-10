#!/bin/bash
# meta-utils.sh - Shared utilities for meta-* slash commands
# Phase 14 - Standardized error detection, exit codes, and common functions

# Exit codes (Phase 11 convention)
readonly EXIT_SUCCESS=0
readonly EXIT_ERROR=1
readonly EXIT_NO_RESULTS=2

# Check if meta-cc is installed
check_meta_cc_installed() {
    if ! command -v meta-cc &> /dev/null; then
        echo "❌ 错误：meta-cc 未安装或不在 PATH 中" >&2
        echo "" >&2
        echo "请安装 meta-cc：" >&2
        echo "  1. 下载或构建 meta-cc 二进制文件" >&2
        echo "  2. 将其放置在 PATH 中（如 /usr/local/bin/meta-cc）" >&2
        echo "  3. 确保可执行权限：chmod +x /usr/local/bin/meta-cc" >&2
        echo "" >&2
        echo "详情参见：https://github.com/yaleh/meta-cc" >&2
        exit $EXIT_ERROR
    fi
}

# Convert JSONL to JSON array for jq processing
jsonl_to_json() {
    local jsonl="$1"
    echo "$jsonl" | jq -s '.'
}

# Check if a tool call is an error
# Usage: is_error "$status" "$error_field"
# Returns: 0 (true) if error, 1 (false) if success
is_error() {
    local status="$1"
    local error="$2"

    [ "$status" = "error" ] || [ -n "$error" ]
}

# Calculate error statistics from tool data (JSON array)
# Usage: calculate_error_stats "$tools_json"
# Output: JSON object with {total, errors, error_rate}
calculate_error_stats() {
    local data="$1"

    echo "$data" | jq '{
        total: length,
        errors: [.[] | select(.Status == "error" or (.Error | length) > 0)] | length,
        error_rate: (
            if length > 0 then
                (([.[] | select(.Status == "error" or (.Error | length) > 0)] | length) / length * 100 | floor)
            else
                0
            end
        )
    }'
}

# Format tool distribution from tool data (JSON array)
# Usage: format_tool_distribution "$tools_json" [limit]
# Output: Markdown list of top N tools
format_tool_distribution() {
    local data="$1"
    local limit="${2:-5}"

    echo "$data" | jq -r --argjson limit "$limit" '
        [.[] | .ToolName] |
        group_by(.) |
        map({tool: .[0], count: length}) |
        sort_by(.count) |
        reverse |
        .[:$limit] |
        .[] |
        "- \(.tool): \(.count) 次"
    '
}

# Handle standard exit codes based on meta-cc command output
# Usage: handle_exit_code $? "operation_name"
handle_exit_code() {
    local code=$1
    local operation="${2:-query}"

    case $code in
        0)
            return $EXIT_SUCCESS
            ;;
        1)
            echo "❌ $operation 执行失败" >&2
            exit $EXIT_ERROR
            ;;
        2)
            echo "ℹ️  未找到结果" >&2
            exit $EXIT_NO_RESULTS
            ;;
        *)
            echo "❌ 未知错误 (exit code: $code)" >&2
            exit $EXIT_ERROR
            ;;
    esac
}
