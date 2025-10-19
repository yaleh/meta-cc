#!/usr/bin/env bash
#
# Coverage Gap Analyzer
# Identifies functions with low coverage and suggests test strategies
#
# Usage: analyze-coverage-gaps.sh [OPTIONS] COVERAGE_FILE
#
# Part of Bootstrap-002 Test Strategy Development Experiment

set -euo pipefail

# Default options
THRESHOLD=80
TOP_N=10
CATEGORY=""
JSON_OUTPUT=false
SHOW_ESTIMATES=true

# Colors for terminal output
RED='\033[0;31m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

usage() {
    cat <<EOF
Usage: $(basename "$0") [OPTIONS] COVERAGE_FILE

Analyze test coverage gaps and suggest priorities for test development.

OPTIONS:
    --threshold PCT      Coverage threshold (default: 80)
    --top N              Show top N functions (default: 10)
    --category CAT       Filter by category (error-handling, business-logic, cli, etc.)
    --json               Output as JSON
    --no-estimates       Don't show time/coverage estimates
    -h, --help           Show this help message

EXAMPLES:
    $(basename "$0") coverage.out
    $(basename "$0") coverage.out --threshold 70 --top 5
    $(basename "$0") coverage.out --category error-handling --json

CATEGORIES:
    error-handling    Validation, error handling (P1: 80-90% target)
    business-logic    Core algorithms, transformations (P2: 75-85% target)
    cli               CLI handlers, command execution (P2: 70-80% target)
    integration       MCP handlers, I/O operations (P3: 70-80% target)
    utility           Helpers, formatters (P3: 60-70% target)
    infrastructure    Init, logging, config (P4: best effort)

EOF
    exit 0
}

# Parse command line arguments
COVERAGE_FILE=""
while [[ $# -gt 0 ]]; do
    case $1 in
        --threshold)
            THRESHOLD="$2"
            shift 2
            ;;
        --top)
            TOP_N="$2"
            shift 2
            ;;
        --category)
            CATEGORY="$2"
            shift 2
            ;;
        --json)
            JSON_OUTPUT=true
            shift
            ;;
        --no-estimates)
            SHOW_ESTIMATES=false
            shift
            ;;
        -h|--help)
            usage
            ;;
        *)
            if [[ -z "$COVERAGE_FILE" ]]; then
                COVERAGE_FILE="$1"
            else
                echo "Error: Unknown option $1" >&2
                usage
            fi
            shift
            ;;
    esac
done

# Validate arguments
if [[ -z "$COVERAGE_FILE" ]]; then
    echo "Error: COVERAGE_FILE required" >&2
    usage
fi

if [[ ! -f "$COVERAGE_FILE" ]]; then
    echo "Error: Coverage file not found: $COVERAGE_FILE" >&2
    exit 1
fi

# Check for required tools
if ! command -v go &> /dev/null; then
    echo "Error: 'go' command not found. Is Go installed?" >&2
    exit 1
fi

# Categorize function based on file path and function name
categorize_function() {
    local file="$1"
    local func="$2"
    local category=""
    local priority=5

    # P1: Error Handling (80-90% target)
    if [[ "$func" =~ ^(Validate|Handle|Check|Verify) ]] || \
       [[ "$file" =~ validation/ ]] || \
       [[ "$file" =~ errors/ ]]; then
        category="error-handling"
        priority=1
    # P2: Business Logic (75-85% target)
    elif [[ "$file" =~ query/ ]] || \
         [[ "$file" =~ analyzer/ ]] || \
         [[ "$func" =~ ^(Process|Parse|Transform|Calculate) ]]; then
        category="business-logic"
        priority=2
    # P2: CLI (70-80% target)
    elif [[ "$file" =~ cmd/ ]] && [[ ! "$file" =~ cmd/mcp-server ]]; then
        category="cli"
        priority=2
    # P3: Integration (70-80% target)
    elif [[ "$file" =~ cmd/mcp-server ]] || \
         [[ "$func" =~ ^(handle|execute|run) ]]; then
        category="integration"
        priority=3
    # P3: Utilities (60-70% target)
    elif [[ "$file" =~ util/ ]] || \
         [[ "$func" =~ ^(Format|Convert|Helper) ]]; then
        category="utility"
        priority=3
    # P4: Infrastructure (best effort)
    elif [[ "$func" =~ ^(Init|Setup|Configure) ]] || \
         [[ "$func" =~ Logger ]] || \
         [[ "$file" =~ config/ ]]; then
        category="infrastructure"
        priority=4
    else
        category="other"
        priority=5
    fi

    echo "$category:$priority"
}

# Suggest test pattern based on category and function
suggest_pattern() {
    local category="$1"
    local func="$2"

    case "$category" in
        error-handling)
            echo "Error Path Pattern (Pattern 4) + Table-Driven (Pattern 2)"
            ;;
        business-logic)
            if [[ "$func" =~ ^(Parse|Process) ]]; then
                echo "Table-Driven Pattern (Pattern 2)"
            else
                echo "Unit Test Pattern (Pattern 1)"
            fi
            ;;
        cli)
            if [[ "$func" =~ Flag|Option ]]; then
                echo "Global Flag Test Pattern (Pattern 8)"
            else
                echo "CLI Command Test Pattern (Pattern 7)"
            fi
            ;;
        integration)
            echo "Integration Test Pattern (Pattern 3) + Dependency Injection (Pattern 6)"
            ;;
        utility)
            echo "Table-Driven Pattern (Pattern 2)"
            ;;
        infrastructure)
            echo "Best effort (may not be testable)"
            ;;
        *)
            echo "Unit Test Pattern (Pattern 1)"
            ;;
    esac
}

# Calculate target coverage based on category
get_target_coverage() {
    local category="$1"

    case "$category" in
        error-handling)
            echo "85"
            ;;
        business-logic)
            echo "80"
            ;;
        cli)
            echo "75"
            ;;
        integration)
            echo "75"
            ;;
        utility)
            echo "65"
            ;;
        infrastructure)
            echo "50"
            ;;
        *)
            echo "70"
            ;;
    esac
}

# Estimate time to write test
estimate_time() {
    local category="$1"
    local coverage_gap="$2"

    # Base time in minutes
    case "$category" in
        error-handling)
            echo "15"  # Error path tests take longer (multiple scenarios)
            ;;
        business-logic)
            echo "12"  # Business logic needs careful testing
            ;;
        cli)
            echo "12"  # CLI tests need setup
            ;;
        integration)
            echo "20"  # Integration tests complex
            ;;
        utility)
            echo "8"   # Utilities usually simpler
            ;;
        infrastructure)
            echo "25"  # Hard to test, may need refactoring
            ;;
        *)
            echo "10"
            ;;
    esac
}

# Estimate coverage impact
estimate_coverage_impact() {
    local category="$1"
    local current_coverage="$2"
    local target_coverage="$3"

    # Simple estimate: gap × 0.7 (assuming we can cover 70% of gap)
    local gap=$((target_coverage - current_coverage))
    local impact=$(awk "BEGIN {print $gap * 0.7}")

    printf "%.1f" "$impact"
}

# Parse coverage and analyze
analyze_coverage() {
    local tmpfile=$(mktemp)

    # Get function-level coverage
    go tool cover -func="$COVERAGE_FILE" > "$tmpfile"

    # Get total coverage
    local total_coverage=$(tail -1 "$tmpfile" | awk '{print $NF}' | sed 's/%//')

    # Parse functions with coverage below threshold
    local -a functions=()
    local -a files=()
    local -a coverages=()
    local -a categories=()
    local -a priorities=()
    local -a patterns=()
    local -a targets=()
    local -a times=()
    local -a impacts=()

    while read -r line; do
        # Skip empty lines and total line
        [[ -z "$line" ]] && continue
        [[ "$line" =~ ^total: ]] && continue

        # Parse line: file:line:function coverage%
        # Split by tabs and spaces, handling function names with special chars
        local file=$(echo "$line" | awk -F: '{print $1}')
        local func=$(echo "$line" | awk '{for(i=2;i<NF;i++) printf "%s ",$i; print ""}' | sed 's/[[:space:]]*$//')
        local coverage=$(echo "$line" | awk '{print $NF}' | sed 's/%//')

        # Skip if no coverage data or function name
        [[ -z "$func" ]] && continue
        [[ -z "$coverage" ]] && continue

        # Skip if coverage >= threshold (handle decimal comparison safely)
        local cov_int=${coverage%.*}
        [[ -z "$cov_int" ]] && cov_int=0
        (( cov_int >= THRESHOLD )) && continue

        # Categorize
        local cat_pri=$(categorize_function "$file" "$func")
        local category=$(echo "$cat_pri" | cut -d':' -f1)
        local priority=$(echo "$cat_pri" | cut -d':' -f2)

        # Filter by category if specified
        if [[ -n "$CATEGORY" ]] && [[ "$category" != "$CATEGORY" ]]; then
            continue
        fi

        # Get pattern suggestion
        local pattern=$(suggest_pattern "$category" "$func")

        # Get target coverage
        local target=$(get_target_coverage "$category")

        # Estimate time
        local time=$(estimate_time "$category" "$((target - ${coverage%.*}))")

        # Estimate impact
        local impact=$(estimate_coverage_impact "$category" "${coverage%.*}" "$target")

        # Store
        functions+=("$func")
        files+=("$file")
        coverages+=("$coverage")
        categories+=("$category")
        priorities+=("$priority")
        patterns+=("$pattern")
        targets+=("$target")
        times+=("$time")
        impacts+=("$impact")
    done < <(tail -n +1 "$tmpfile" | grep -v "^total:")

    rm "$tmpfile"

    # Sort by priority then coverage (ascending)
    local -a indices=()
    for i in "${!functions[@]}"; do
        indices+=("$i")
    done

    # Sort indices
    IFS=$'\n' sorted_indices=($(
        for i in "${indices[@]}"; do
            printf "%d\t%d\t%.1f\n" "$i" "${priorities[$i]}" "${coverages[$i]}"
        done | sort -t$'\t' -k2,2n -k3,3n | cut -f1
    ))
    unset IFS

    # Output results
    if $JSON_OUTPUT; then
        output_json "$total_coverage"
    else
        output_text "$total_coverage"
    fi
}

output_text() {
    local total_coverage="$1"

    echo -e "${BLUE}COVERAGE GAP ANALYSIS - $(date +%Y-%m-%d)${NC}"
    echo ""
    echo "Total Coverage: ${total_coverage}%"
    echo "Target: ${THRESHOLD}.0%"
    local gap=$(awk "BEGIN {printf \"%.1f\", $THRESHOLD - $total_coverage}")
    echo "Gap: ${gap} percentage points"
    echo ""

    # Group by priority
    local current_priority=""
    local count=0

    for idx in "${sorted_indices[@]}"; do
        [[ $count -ge $TOP_N ]] && break

        local priority="${priorities[$idx]}"
        local category="${categories[$idx]}"
        local func="${functions[$idx]}"
        local file="${files[$idx]}"
        local coverage="${coverages[$idx]}"
        local target="${targets[$idx]}"
        local pattern="${patterns[$idx]}"
        local time="${times[$idx]}"
        local impact="${impacts[$idx]}"

        # Print priority header
        if [[ "$priority" != "$current_priority" ]]; then
            current_priority="$priority"
            echo ""
            case "$priority" in
                1)
                    echo -e "${RED}HIGH PRIORITY (Error Handling - 0-${THRESHOLD}% coverage):${NC}"
                    ;;
                2)
                    echo -e "${YELLOW}MEDIUM PRIORITY (Business Logic/CLI - 0-${THRESHOLD}% coverage):${NC}"
                    ;;
                3)
                    echo -e "${YELLOW}MEDIUM PRIORITY (Integration/Utilities - 0-${THRESHOLD}% coverage):${NC}"
                    ;;
                4)
                    echo -e "${GREEN}LOW PRIORITY (Infrastructure - best effort):${NC}"
                    ;;
            esac
        fi

        count=$((count + 1))

        # Print function info
        echo -e "$count. ${file}:${func} (${coverage}%) - P${priority} ${category}"
        if $SHOW_ESTIMATES; then
            echo "   Target: ${target}%, Pattern: ${pattern}"
            echo "   Est. time: ${time} min, Est. coverage impact: +${impact}% function"
        fi
    done

    # Summary
    if $SHOW_ESTIMATES; then
        echo ""
        echo -e "${BLUE}RECOMMENDED TEST PATTERNS:${NC}"

        # Get unique category-pattern pairs
        declare -A pattern_map
        for idx in "${sorted_indices[@]}"; do
            local category="${categories[$idx]}"
            local pattern="${patterns[$idx]}"
            pattern_map["$category"]="$pattern"
        done

        for category in "${!pattern_map[@]}"; do
            echo "- ${category}: ${pattern_map[$category]}"
        done

        # Estimates
        echo ""
        echo -e "${BLUE}ESTIMATED WORK:${NC}"

        # Calculate totals by priority
        local p1_count=0 p1_time=0 p1_impact=0
        local p2_count=0 p2_time=0 p2_impact=0
        local p3_count=0 p3_time=0 p3_impact=0

        local shown=0
        for idx in "${sorted_indices[@]}"; do
            [[ $shown -ge $TOP_N ]] && break
            shown=$((shown + 1))

            local priority="${priorities[$idx]}"
            local time="${times[$idx]}"
            local impact="${impacts[$idx]}"

            case "$priority" in
                1)
                    p1_count=$((p1_count + 1))
                    p1_time=$((p1_time + time))
                    p1_impact=$(awk "BEGIN {printf \"%.1f\", $p1_impact + $impact}")
                    ;;
                2)
                    p2_count=$((p2_count + 1))
                    p2_time=$((p2_time + time))
                    p2_impact=$(awk "BEGIN {printf \"%.1f\", $p2_impact + $impact}")
                    ;;
                3)
                    p3_count=$((p3_count + 1))
                    p3_time=$((p3_time + time))
                    p3_impact=$(awk "BEGIN {printf \"%.1f\", $p3_impact + $impact}")
                    ;;
            esac
        done

        if [[ $p1_count -gt 0 ]]; then
            echo "- High priority: $p1_count functions × ~$((p1_time / p1_count)) min avg = $p1_time min → est. +${p1_impact}% function coverage"
        fi
        if [[ $p2_count -gt 0 ]]; then
            echo "- Medium priority: $p2_count functions × ~$((p2_time / p2_count)) min avg = $p2_time min → est. +${p2_impact}% function coverage"
        fi
        if [[ $p3_count -gt 0 ]]; then
            echo "- Low priority: $p3_count functions × ~$((p3_time / p3_count)) min avg = $p3_time min → est. +${p3_impact}% function coverage"
        fi

        local total_time=$((p1_time + p2_time + p3_time))
        local total_impact=$(awk "BEGIN {printf \"%.1f\", $p1_impact + $p2_impact + $p3_impact}")
        if [[ $total_time -gt 0 ]]; then
            echo "- Total estimated: $total_time min ($(awk "BEGIN {printf \"%.1f\", $total_time / 60}") hours)"
        fi
    fi
}

output_json() {
    local total_coverage="$1"

    echo "{"
    echo "  \"timestamp\": \"$(date -Iseconds)\","
    echo "  \"total_coverage\": $total_coverage,"
    echo "  \"threshold\": $THRESHOLD,"
    echo "  \"gap\": $(awk "BEGIN {printf \"%.1f\", $THRESHOLD - $total_coverage}"),"
    echo "  \"functions\": ["

    local count=0
    for idx in "${sorted_indices[@]}"; do
        [[ $count -ge $TOP_N ]] && break
        [[ $count -gt 0 ]] && echo ","

        local priority="${priorities[$idx]}"
        local category="${categories[$idx]}"
        local func="${functions[$idx]}"
        local file="${files[$idx]}"
        local coverage="${coverages[$idx]}"
        local target="${targets[$idx]}"
        local pattern="${patterns[$idx]}"
        local time="${times[$idx]}"
        local impact="${impacts[$idx]}"

        echo -n "    {"
        echo -n "\"file\": \"$file\", "
        echo -n "\"function\": \"$func\", "
        echo -n "\"coverage\": $coverage, "
        echo -n "\"target\": $target, "
        echo -n "\"priority\": $priority, "
        echo -n "\"category\": \"$category\", "
        echo -n "\"pattern\": \"$pattern\", "
        echo -n "\"estimate_time_min\": $time, "
        echo -n "\"estimate_impact_pct\": $impact"
        echo -n "}"

        count=$((count + 1))
    done

    echo ""
    echo "  ]"
    echo "}"
}

# Main execution
analyze_coverage
