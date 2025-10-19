#!/bin/bash
# check-file-size.sh - Check file size before reading to prevent token limit errors
# Purpose: Prevent "File Content Size Exceeded" errors
# Target: 1.5% of errors (20 errors)
# Expected speedup: Instant prevention (vs failed Read + retry with offset/limit)

set -euo pipefail

# Usage information
usage() {
    cat <<EOF
Usage: check-file-size.sh [OPTIONS] FILE [FILE...]

Check file sizes before reading to prevent token limit errors.
Claude Code has a 25,000 token limit per Read operation.

OPTIONS:
    -h, --help              Show this help message
    -t, --tokens LIMIT      Token limit (default: 25000)
    -c, --chars-per-token N Characters per token estimate (default: 4)
    -s, --suggest           Suggest alternative read strategies
    -v, --verbose           Verbose output
    -q, --quiet             Suppress warnings (exit code only)
    --warn-threshold PCT    Warn if size exceeds PCT% of limit (default: 80)

EXIT CODES:
    0   All files are within token limit
    1   One or more files exceed token limit
    2   Invalid arguments

EXAMPLES:
    # Check single file
    check-file-size.sh largefile.txt

    # Check multiple files
    check-file-size.sh file1.txt file2.txt file3.txt

    # Custom token limit
    check-file-size.sh --tokens 10000 file.txt

    # Get suggestions for large files
    check-file-size.sh --suggest largefile.json

    # Use in workflow
    if check-file-size.sh --quiet file.txt; then
        # Safe to read entire file
        cat file.txt
    else
        # File too large, use alternatives
        head -n 100 file.txt
    fi

ALTERNATIVE STRATEGIES:
    For files exceeding token limit:
    - Use Read tool with offset and limit parameters
    - Use Grep tool to search for specific content
    - Use head/tail to read portions of file
    - Split large files into smaller chunks
EOF
    exit 0
}

# Default options
TOKEN_LIMIT=25000
CHARS_PER_TOKEN=4
SUGGEST=0
VERBOSE=0
QUIET=0
WARN_THRESHOLD=80
FILES=()

# Parse arguments
while [[ $# -gt 0 ]]; do
    case "$1" in
        -h|--help)
            usage
            ;;
        -t|--tokens)
            TOKEN_LIMIT="$2"
            shift 2
            ;;
        -c|--chars-per-token)
            CHARS_PER_TOKEN="$2"
            shift 2
            ;;
        -s|--suggest)
            SUGGEST=1
            shift
            ;;
        -v|--verbose)
            VERBOSE=1
            shift
            ;;
        -q|--quiet)
            QUIET=1
            shift
            ;;
        --warn-threshold)
            WARN_THRESHOLD="$2"
            shift 2
            ;;
        -*)
            echo "Error: Unknown option: $1" >&2
            echo "Use --help for usage information" >&2
            exit 2
            ;;
        *)
            FILES+=("$1")
            shift
            ;;
    esac
done

# Check if files provided
if [[ ${#FILES[@]} -eq 0 ]]; then
    echo "Error: No files provided" >&2
    echo "Use --help for usage information" >&2
    exit 2
fi

# Calculate limits
CHAR_LIMIT=$((TOKEN_LIMIT * CHARS_PER_TOKEN))
WARN_LIMIT=$((CHAR_LIMIT * WARN_THRESHOLD / 100))

# Function to suggest alternatives
suggest_alternatives() {
    local file="$1"
    local size="$2"

    [[ $QUIET -eq 0 ]] && cat <<EOF >&2

  Alternative strategies:
    1. Read with offset/limit:
       Read tool with offset=0, limit=1000 (read first 1000 lines)

    2. Search for specific content:
       Grep tool to find relevant sections

    3. Read file portions:
       head -n 1000 "$file"  # First 1000 lines
       tail -n 1000 "$file"  # Last 1000 lines

    4. Split file:
       split -l 5000 "$file" "${file}.part-"

    5. Calculate optimal offset/limit:
       Total lines: $(wc -l < "$file" 2>/dev/null || echo "unknown")
       Suggested limit per read: $((TOKEN_LIMIT * 4 / 10))  # ~40% of limit
EOF
}

# Function to format size
format_size() {
    local bytes="$1"
    if [[ $bytes -lt 1024 ]]; then
        echo "${bytes}B"
    elif [[ $bytes -lt $((1024 * 1024)) ]]; then
        echo "$((bytes / 1024))KB"
    else
        echo "$((bytes / 1024 / 1024))MB"
    fi
}

# Main checking loop
EXIT_CODE=0

for file in "${FILES[@]}"; do
    # Check if file exists
    if [[ ! -f "$file" ]]; then
        EXIT_CODE=1
        [[ $QUIET -eq 0 ]] && echo "✗ $file (does not exist)" >&2
        continue
    fi

    # Get file size in bytes
    file_size=$(stat -f%z "$file" 2>/dev/null || stat -c%s "$file" 2>/dev/null || echo 0)

    # Estimate tokens
    estimated_tokens=$((file_size / CHARS_PER_TOKEN))

    # Check against limit
    if [[ $file_size -gt $CHAR_LIMIT ]]; then
        # Exceeds limit
        EXIT_CODE=1
        if [[ $QUIET -eq 0 ]]; then
            echo "✗ $file ($(format_size $file_size), ~${estimated_tokens} tokens, EXCEEDS LIMIT)" >&2
        fi

        if [[ $SUGGEST -eq 1 ]]; then
            suggest_alternatives "$file" "$file_size"
        fi
    elif [[ $file_size -gt $WARN_LIMIT ]]; then
        # Within limit but close to threshold
        if [[ $QUIET -eq 0 ]]; then
            echo "⚠ $file ($(format_size $file_size), ~${estimated_tokens} tokens, ${WARN_THRESHOLD}%+ of limit)" >&2
        fi
    else
        # Well within limit
        if [[ $VERBOSE -eq 1 ]]; then
            echo "✓ $file ($(format_size $file_size), ~${estimated_tokens} tokens, OK)"
        fi
    fi
done

# Summary
if [[ $EXIT_CODE -eq 0 ]] && [[ $VERBOSE -eq 1 ]]; then
    echo ""
    echo "All files are within token limit (${TOKEN_LIMIT} tokens = ${CHAR_LIMIT} chars)"
elif [[ $EXIT_CODE -eq 1 ]] && [[ $QUIET -eq 0 ]]; then
    echo "" >&2
    echo "One or more files exceed token limit." >&2
    echo "Use --suggest for alternative read strategies." >&2
fi

exit $EXIT_CODE
