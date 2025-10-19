#!/bin/bash
# validate-path.sh - Validate file paths before operations
# Purpose: Prevent "File Not Found" errors by validating paths before use
# Target: 18.7% of errors (250 errors)
# Expected speedup: 5-10x (instant validation vs manual retry)

set -euo pipefail

# Usage information
usage() {
    cat <<EOF
Usage: validate-path.sh [OPTIONS] PATH [PATH...]

Validate file paths before operations to prevent "File Not Found" errors.

OPTIONS:
    -h, --help          Show this help message
    -c, --create        Create missing directories (not files)
    -s, --suggest       Suggest similar paths if not found
    -v, --verbose       Verbose output
    -q, --quiet         Suppress warnings (exit code only)

EXIT CODES:
    0   All paths exist
    1   One or more paths do not exist
    2   Invalid arguments

EXAMPLES:
    # Validate single file
    validate-path.sh /path/to/file.txt

    # Validate multiple files
    validate-path.sh file1.txt file2.txt dir/file3.txt

    # Create missing directories
    validate-path.sh --create /path/to/new/directory

    # Get suggestions for mistyped paths
    validate-path.sh --suggest /path/to/fiel.txt

    # Use in scripts
    if validate-path.sh --quiet /path/to/file; then
        cat /path/to/file
    else
        echo "File does not exist"
    fi
EOF
    exit 0
}

# Default options
CREATE_DIRS=0
SUGGEST=0
VERBOSE=0
QUIET=0
PATHS=()

# Parse arguments
while [[ $# -gt 0 ]]; do
    case "$1" in
        -h|--help)
            usage
            ;;
        -c|--create)
            CREATE_DIRS=1
            shift
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
        -*)
            echo "Error: Unknown option: $1" >&2
            echo "Use --help for usage information" >&2
            exit 2
            ;;
        *)
            PATHS+=("$1")
            shift
            ;;
    esac
done

# Check if paths provided
if [[ ${#PATHS[@]} -eq 0 ]]; then
    echo "Error: No paths provided" >&2
    echo "Use --help for usage information" >&2
    exit 2
fi

# Function to suggest similar paths
suggest_similar() {
    local target="$1"
    local dirname=$(dirname "$target")
    local basename=$(basename "$target")

    # Check if directory exists
    if [[ ! -d "$dirname" ]]; then
        [[ $QUIET -eq 0 ]] && echo "  Directory does not exist: $dirname" >&2
        return
    fi

    # Find similar files (case-insensitive, fuzzy match)
    local suggestions=$(find "$dirname" -maxdepth 1 -iname "*${basename}*" 2>/dev/null | head -5)

    if [[ -n "$suggestions" ]]; then
        [[ $QUIET -eq 0 ]] && echo "  Did you mean:" >&2
        while IFS= read -r suggestion; do
            [[ $QUIET -eq 0 ]] && echo "    $suggestion" >&2
        done <<< "$suggestions"
    else
        [[ $QUIET -eq 0 ]] && echo "  No similar paths found" >&2
    fi
}

# Main validation loop
EXIT_CODE=0

for path in "${PATHS[@]}"; do
    # Expand ~ and environment variables
    eval path_expanded="$path"

    if [[ -e "$path_expanded" ]]; then
        # Path exists
        if [[ $VERBOSE -eq 1 ]]; then
            echo "✓ $path (exists)"
        fi
    else
        # Path does not exist
        EXIT_CODE=1

        if [[ $QUIET -eq 0 ]]; then
            echo "✗ $path (does not exist)" >&2
        fi

        # Try to create directory if requested
        if [[ $CREATE_DIRS -eq 1 ]]; then
            # Determine if this is a directory or file path
            if [[ "$path_expanded" == */ ]] || [[ ! "$path_expanded" =~ \. ]]; then
                # Looks like a directory
                if [[ $QUIET -eq 0 ]]; then
                    echo "  Creating directory: $path_expanded" >&2
                fi
                mkdir -p "$path_expanded"
                if [[ $? -eq 0 ]]; then
                    EXIT_CODE=0
                    [[ $QUIET -eq 0 ]] && echo "  ✓ Directory created" >&2
                fi
            else
                # Looks like a file - create parent directory
                local parent_dir=$(dirname "$path_expanded")
                if [[ ! -d "$parent_dir" ]]; then
                    if [[ $QUIET -eq 0 ]]; then
                        echo "  Creating parent directory: $parent_dir" >&2
                    fi
                    mkdir -p "$parent_dir"
                    [[ $? -eq 0 ]] && [[ $QUIET -eq 0 ]] && echo "  ✓ Parent directory created" >&2
                fi
            fi
        fi

        # Suggest similar paths if requested
        if [[ $SUGGEST -eq 1 ]]; then
            suggest_similar "$path_expanded"
        fi
    fi
done

exit $EXIT_CODE
