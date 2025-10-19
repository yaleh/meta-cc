#!/bin/bash
# check-read-before-write.sh - Check if file was read before write/edit
# Purpose: Prevent "Write Before Read" violations (Claude Code safety constraint)
# Target: 3.0% of errors (40 errors)
# Expected speedup: 10x (instant check vs manual retry)

set -euo pipefail

# Usage information
usage() {
    cat <<EOF
Usage: check-read-before-write.sh [OPTIONS] FILE [FILE...]

Check if files have been read before attempting write/edit operations.
This prevents Claude Code "Write Before Read" safety constraint violations.

OPTIONS:
    -h, --help          Show this help message
    -a, --auto-read     Automatically read files if not yet read
    -l, --log FILE      Log file tracking read operations (default: .read-log)
    -v, --verbose       Verbose output
    -q, --quiet         Suppress warnings (exit code only)
    --reset             Reset read log (clear tracking)

EXIT CODES:
    0   All files have been read (safe to write)
    1   One or more files have not been read
    2   Invalid arguments

EXAMPLES:
    # Check if file was read
    check-read-before-write.sh /path/to/file.txt

    # Auto-read if not yet read
    check-read-before-write.sh --auto-read /path/to/file.txt

    # Use in workflow
    if check-read-before-write.sh --quiet file.txt; then
        # Safe to edit
        edit file.txt
    else
        # Need to read first
        cat file.txt
    fi

NOTE:
    This is a simple demonstration tool. In practice, Claude Code tracks
    read operations internally. This tool simulates that tracking for
    illustration and can be used in automated workflows.
EOF
    exit 0
}

# Default options
AUTO_READ=0
VERBOSE=0
QUIET=0
LOG_FILE="${HOME}/.read-log"
RESET=0
FILES=()

# Parse arguments
while [[ $# -gt 0 ]]; do
    case "$1" in
        -h|--help)
            usage
            ;;
        -a|--auto-read)
            AUTO_READ=1
            shift
            ;;
        -l|--log)
            LOG_FILE="$2"
            shift 2
            ;;
        -v|--verbose)
            VERBOSE=1
            shift
            ;;
        -q|--quiet)
            QUIET=1
            shift
            ;;
        --reset)
            RESET=1
            shift
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

# Reset log if requested
if [[ $RESET -eq 1 ]]; then
    rm -f "$LOG_FILE"
    [[ $QUIET -eq 0 ]] && echo "Read log reset"
    exit 0
fi

# Check if files provided
if [[ ${#FILES[@]} -eq 0 ]]; then
    echo "Error: No files provided" >&2
    echo "Use --help for usage information" >&2
    exit 2
fi

# Create log file if it doesn't exist
touch "$LOG_FILE"

# Function to mark file as read
mark_as_read() {
    local file="$1"
    local abs_path=$(realpath "$file" 2>/dev/null || echo "$file")
    local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    # Append to log
    echo "${timestamp}|${abs_path}" >> "$LOG_FILE"

    [[ $VERBOSE -eq 1 ]] && echo "Marked as read: $abs_path"
}

# Function to check if file was read
was_read() {
    local file="$1"
    local abs_path=$(realpath "$file" 2>/dev/null || echo "$file")

    # Check if file exists in log
    if grep -q "|${abs_path}$" "$LOG_FILE"; then
        return 0
    else
        return 1
    fi
}

# Function to read file (demonstration)
read_file() {
    local file="$1"

    if [[ ! -f "$file" ]]; then
        [[ $QUIET -eq 0 ]] && echo "Error: File does not exist: $file" >&2
        return 1
    fi

    # Simulate reading (in real workflow, would use Read tool)
    [[ $VERBOSE -eq 1 ]] && echo "Reading file: $file"
    head -n 1 "$file" >/dev/null 2>&1

    # Mark as read
    mark_as_read "$file"

    return 0
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

    # Check if file was read
    if was_read "$file"; then
        # File was read - safe to write
        if [[ $VERBOSE -eq 1 ]]; then
            echo "✓ $file (read previously, safe to write)"
        fi
    else
        # File not read yet
        EXIT_CODE=1

        if [[ $QUIET -eq 0 ]]; then
            echo "✗ $file (not read yet, unsafe to write)" >&2
        fi

        # Auto-read if requested
        if [[ $AUTO_READ -eq 1 ]]; then
            if read_file "$file"; then
                EXIT_CODE=0
                [[ $QUIET -eq 0 ]] && echo "  ✓ File read automatically" >&2
            else
                [[ $QUIET -eq 0 ]] && echo "  ✗ Failed to read file" >&2
            fi
        else
            [[ $QUIET -eq 0 ]] && echo "  Hint: Read file first or use --auto-read" >&2
        fi
    fi
done

exit $EXIT_CODE
