#!/usr/bin/env bash
#
# Test Generator
# Generates test scaffolds from function signatures using documented patterns
#
# Usage: generate-test.sh [OPTIONS] FUNCTION_NAME
#
# Part of Bootstrap-002 Test Strategy Development Experiment

set -euo pipefail

# Default options
PATTERN="table-driven"
PACKAGE=""
OUTPUT_FILE=""
APPEND=false
DRY_RUN=false
SCENARIOS=3

usage() {
    cat <<EOF
Usage: $(basename "$0") [OPTIONS] FUNCTION_NAME

Generate test scaffolds from function signatures using documented patterns.

OPTIONS:
    --pattern PATTERN    Test pattern (default: table-driven)
                        unit, table-driven, error-path, cli-command, global-flag
    --package PACKAGE    Package name (default: infer from current dir)
    --output FILE        Output file (default: <package>_test.go)
    --append             Append to existing file instead of creating new
    --scenarios N        Number of test scenarios (default: 3)
    --dry-run            Print to stdout instead of writing file
    -h, --help           Show this help message

EXAMPLES:
    $(basename "$0") ParseQuery --pattern table-driven
    $(basename "$0") ValidateInput --pattern error-path --scenarios 4
    $(basename "$0") Execute --pattern cli-command --package cmd

PATTERNS:
    unit             Simple unit test (Pattern 1)
    table-driven     Table-driven test with multiple scenarios (Pattern 2)
    error-path       Error path testing with validation (Pattern 4)
    cli-command      CLI command test with flag parsing (Pattern 7)
    global-flag      Global flag test pattern (Pattern 8)

EOF
    exit 0
}

# Parse command line arguments
FUNCTION_NAME=""
while [[ $# -gt 0 ]]; do
    case $1 in
        --pattern)
            PATTERN="$2"
            shift 2
            ;;
        --package)
            PACKAGE="$2"
            shift 2
            ;;
        --output)
            OUTPUT_FILE="$2"
            shift 2
            ;;
        --append)
            APPEND=true
            shift
            ;;
        --scenarios)
            SCENARIOS="$2"
            shift 2
            ;;
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        -h|--help)
            usage
            ;;
        *)
            if [[ -z "$FUNCTION_NAME" ]]; then
                FUNCTION_NAME="$1"
            else
                echo "Error: Unknown option $1" >&2
                usage
            fi
            shift
            ;;
    esac
done

# Validate arguments
if [[ -z "$FUNCTION_NAME" ]]; then
    echo "Error: FUNCTION_NAME required" >&2
    usage
fi

# Infer package if not specified
if [[ -z "$PACKAGE" ]]; then
    if [[ -f "go.mod" ]]; then
        PACKAGE=$(head -1 go.mod | awk '{print $2}' | xargs basename)
    else
        PACKAGE=$(basename "$PWD")
    fi
fi

# Set output file if not specified
if [[ -z "$OUTPUT_FILE" ]]; then
    OUTPUT_FILE="${PACKAGE}_test.go"
fi

# Generate test based on pattern
generate_unit_test() {
    cat <<EOF
func Test${FUNCTION_NAME}(t *testing.T) {
    // Setup
    // TODO: Create test input

    // Execute
    result, err := ${FUNCTION_NAME}(/* TODO: add arguments */)

    // Assert
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    // TODO: Add assertions
    _ = result // Remove when assertions added
}
EOF
}

generate_table_driven_test() {
    cat <<EOF
func Test${FUNCTION_NAME}(t *testing.T) {
    tests := []struct {
        name     string
        // TODO: Add input fields
        expected interface{} // TODO: Change to actual type
        wantErr  bool
    }{
EOF

    for i in $(seq 1 $SCENARIOS); do
        local comma=""
        [[ $i -lt $SCENARIOS ]] && comma=","
        cat <<EOF
        {
            name:     "scenario ${i}",
            // TODO: Add test data
            expected: nil, // TODO: Add expected value
            wantErr:  false,
        }${comma}
EOF
    done

    cat <<EOF
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Execute
            result, err := ${FUNCTION_NAME}(/* TODO: add arguments */)

            // Assert error
            if (err != nil) != tt.wantErr {
                t.Errorf("${FUNCTION_NAME}() error = %v, wantErr %v", err, tt.wantErr)
                return
            }

            // Assert result
            if !tt.wantErr {
                // TODO: Add result comparison
                _ = result // Remove when comparison added
            }
        })
    }
}
EOF
}

generate_error_path_test() {
    cat <<EOF
func Test${FUNCTION_NAME}_ErrorCases(t *testing.T) {
    tests := []struct {
        name    string
        // TODO: Add input fields
        wantErr bool
        errMsg  string
    }{
EOF

    # Generate error scenarios
    local scenarios=("nil input" "empty input" "invalid format" "out of range")
    for i in $(seq 0 $((SCENARIOS - 1))); do
        local scenario="${scenarios[$i]}"
        [[ -z "$scenario" ]] && scenario="error scenario $((i + 1))"
        local comma=""
        [[ $i -lt $((SCENARIOS - 1)) ]] && comma=","

        cat <<EOF
        {
            name:    "${scenario}",
            // TODO: Add input for this error case
            wantErr: true,
            errMsg:  "${scenario}",
        }${comma}
EOF
    done

    cat <<EOF
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Execute
            _, err := ${FUNCTION_NAME}(/* TODO: add arguments */)

            // Assert error
            if (err != nil) != tt.wantErr {
                t.Errorf("${FUNCTION_NAME}() error = %v, wantErr %v", err, tt.wantErr)
                return
            }

            // Assert error message
            if tt.wantErr && !strings.Contains(err.Error(), tt.errMsg) {
                t.Errorf("expected error containing '%s', got '%s'", tt.errMsg, err.Error())
            }
        })
    }
}
EOF
}

generate_cli_command_test() {
    cat <<EOF
func Test${FUNCTION_NAME}_Command(t *testing.T) {
    tests := []struct {
        name    string
        args    []string
        wantErr bool
    }{
EOF

    for i in $(seq 1 $SCENARIOS); do
        local comma=""
        [[ $i -lt $SCENARIOS ]] && comma=","
        cat <<EOF
        {
            name:    "scenario ${i}",
            args:    []string{/* TODO: add command args/flags */},
            wantErr: false,
        }${comma}
EOF
    done

    cat <<EOF
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup: Create command
            cmd := &cobra.Command{
                Use: "${FUNCTION_NAME}",
                RunE: func(cmd *cobra.Command, args []string) error {
                    // TODO: Add command logic or call ${FUNCTION_NAME}
                    return nil
                },
            }

            // TODO: Add flags
            // cmd.Flags().StringP("flag", "f", "default", "description")

            // Setup: Set arguments
            cmd.SetArgs(tt.args)

            // Setup: Capture output
            var buf bytes.Buffer
            cmd.SetOut(&buf)
            cmd.SetErr(&buf)

            // Execute
            err := cmd.Execute()

            // Assert
            if (err != nil) != tt.wantErr {
                t.Errorf("command failed: %v\nOutput: %s", err, buf.String())
            }

            // TODO: Verify output
            _ = buf.String()
        })
    }
}
EOF
}

generate_global_flag_test() {
    cat <<EOF
func Test${FUNCTION_NAME}_GlobalFlags(t *testing.T) {
    tests := []struct {
        name     string
        args     []string
        // TODO: Add expected option fields
        wantErr  bool
    }{
EOF

    local flags=("no flags" "flag1 set" "flag2 set" "all flags")
    for i in $(seq 0 $((SCENARIOS - 1))); do
        local scenario="${flags[$i]}"
        [[ -z "$scenario" ]] && scenario="scenario $((i + 1))"
        local comma=""
        [[ $i -lt $((SCENARIOS - 1)) ]] && comma=","

        cat <<EOF
        {
            name:    "${scenario}",
            args:    []string{/* TODO: add flags */},
            // TODO: Add expected values
            wantErr: false,
        }${comma}
EOF
    done

    cat <<EOF
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup: Reset global flags
            // TODO: Add reset logic if needed
            // resetGlobalFlags()

            // Setup: Create root command
            rootCmd := &cobra.Command{Use: "root"}

            // TODO: Add global flags
            // rootCmd.PersistentFlags().String("global-flag", "", "description")

            // Setup: Set args
            rootCmd.SetArgs(tt.args)

            // Execute: Parse flags
            if err := rootCmd.ParseFlags(tt.args); (err != nil) != tt.wantErr {
                t.Fatalf("ParseFlags() error = %v, wantErr %v", err, tt.wantErr)
            }

            // TODO: Get and verify global options
            // opts := getGlobalOptions()
            // Verify opts match expected values
        })
    }
}
EOF
}

# Generate test based on pattern
generate_test() {
    case "$PATTERN" in
        unit)
            generate_unit_test
            ;;
        table-driven)
            generate_table_driven_test
            ;;
        error-path)
            generate_error_path_test
            ;;
        cli-command)
            generate_cli_command_test
            ;;
        global-flag)
            generate_global_flag_test
            ;;
        *)
            echo "Error: Unknown pattern '$PATTERN'" >&2
            echo "Valid patterns: unit, table-driven, error-path, cli-command, global-flag" >&2
            exit 1
            ;;
    esac
}

# Generate test header if creating new file
generate_header() {
    if ! $APPEND || [[ ! -f "$OUTPUT_FILE" ]]; then
        cat <<EOF
package ${PACKAGE}

import (
    "testing"
EOF

        # Add imports based on pattern
        case "$PATTERN" in
            error-path)
                echo '    "strings"'
                ;;
            cli-command|global-flag)
                echo '    "bytes"'
                echo '    "github.com/spf13/cobra"'
                ;;
        esac

        cat <<EOF
)

EOF
    fi
}

# Main execution
main() {
    local output=""

    # Generate header
    output+=$(generate_header)

    # Generate test
    output+=$(generate_test)

    # Output
    if $DRY_RUN; then
        echo "$output"
    else
        if $APPEND && [[ -f "$OUTPUT_FILE" ]]; then
            echo "$output" >> "$OUTPUT_FILE"
            echo "Appended test to: $OUTPUT_FILE"
        else
            echo "$output" > "$OUTPUT_FILE"
            echo "Created test file: $OUTPUT_FILE"
        fi

        # Format with gofmt if available
        if command -v gofmt &> /dev/null; then
            gofmt -w "$OUTPUT_FILE"
            echo "Formatted with gofmt"
        fi

        # Show summary
        echo ""
        echo "Summary:"
        echo "  Function: ${FUNCTION_NAME}"
        echo "  Pattern: ${PATTERN}"
        echo "  Scenarios: ${SCENARIOS}"
        echo "  Output: ${OUTPUT_FILE}"
        echo ""
        echo "Next steps:"
        echo "  1. Fill in TODO items in generated test"
        echo "  2. Run: go test -v ./${PACKAGE}/..."
        echo "  3. Verify coverage: go test -cover ./${PACKAGE}/..."
    fi
}

main
