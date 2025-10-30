#!/bin/bash

# Shell helper for query_user_messages tool
# Usage: query_user_messages.sh [options] <session_directory>
# Options:
#   -p, --pattern PATTERN    Regex pattern to match in message content
#   -l, --limit LIMIT        Maximum number of results to return
#   -s, --scope SCOPE        Scope: project or session (default: project)

set -e

# Default values
PATTERN=""
LIMIT=0
SCOPE="project"
SESSION_DIR=""

# Function to display usage
usage() {
    echo "Usage: $0 [options] <session_directory>"
    echo "Options:"
    echo "  -p, --pattern PATTERN    Regex pattern to match in message content"
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
            SESSION_DIR="$1"
            shift
            ;;
    esac
done

# Check if session directory is provided
if [ -z "$SESSION_DIR" ]; then
    echo "Error: Session directory is required"
    usage
fi

# Check if session directory exists
if [ ! -d "$SESSION_DIR" ]; then
    echo "Error: Session directory does not exist: $SESSION_DIR"
    exit 1
fi

# Build the jq filter
FILTER='select(.type == "user")'
if [ -n "$PATTERN" ]; then
    FILTER="$FILTER | select(.message.content | test(\"$PATTERN\"))"
fi

# Find JSONL files
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
