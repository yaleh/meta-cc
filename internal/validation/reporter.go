package validation

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Reporter formats and outputs validation results
type Reporter struct {
	quiet      bool
	jsonOutput bool
	writer     io.Writer
}

// NewReporter creates a new reporter
func NewReporter(quiet, jsonOutput bool) *Reporter {
	return &Reporter{
		quiet:      quiet,
		jsonOutput: jsonOutput,
		writer:     os.Stdout,
	}
}

// Print outputs the validation report
func (r *Reporter) Print(report *Report) {
	if r.jsonOutput {
		r.printJSON(report)
	} else {
		r.printTerminal(report)
	}
}

func (r *Reporter) printJSON(report *Report) {
	encoder := json.NewEncoder(r.writer)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(report)
}

func (r *Reporter) printTerminal(report *Report) {
	if r.quiet && report.Failed == 0 {
		return
	}

	// Header
	fmt.Fprintln(r.writer, "API Consistency Validation")
	fmt.Fprintln(r.writer, "==========================")
	fmt.Fprintln(r.writer, "")

	fmt.Fprintf(r.writer, "Found %d tools\n", report.TotalTools)
	fmt.Fprintln(r.writer, "")

	// Checks run
	fmt.Fprintln(r.writer, "Running checks (MVP mode):")
	fmt.Fprintln(r.writer, "  ✓ Naming pattern validation")
	fmt.Fprintln(r.writer, "  ✓ Parameter ordering validation")
	fmt.Fprintln(r.writer, "  ✓ Description format validation")
	fmt.Fprintln(r.writer, "")

	// Results
	fmt.Fprintln(r.writer, "Results:")
	fmt.Fprintln(r.writer, "--------")
	fmt.Fprintln(r.writer, "")

	// Group results by tool
	toolResults := groupResultsByTool(report.Results)

	for toolName, results := range toolResults {
		hasFailures := false
		for _, result := range results {
			if result.Status == "FAIL" || result.Status == "WARN" {
				hasFailures = true
				break
			}
		}

		if hasFailures {
			for _, result := range results {
				if result.Status == "FAIL" {
					r.printFailure(result)
				} else if result.Status == "WARN" {
					r.printWarning(result)
				}
			}
		} else if !r.quiet {
			fmt.Fprintf(r.writer, "✓ %s: All checks passed\n", toolName)
		}
	}

	fmt.Fprintln(r.writer, "")

	// Summary
	fmt.Fprintln(r.writer, "Summary:")
	fmt.Fprintln(r.writer, "--------")
	fmt.Fprintf(r.writer, "Total tools:     %d\n", report.TotalTools)
	fmt.Fprintf(r.writer, "Checks run:      %d\n", report.ChecksRun)
	fmt.Fprintf(r.writer, "Passed:          %d\n", report.Passed)
	fmt.Fprintf(r.writer, "Failed:          %d\n", report.Failed)

	if report.Warnings > 0 {
		fmt.Fprintf(r.writer, "Warnings:        %d\n", report.Warnings)
	}

	fmt.Fprintln(r.writer, "")

	// Overall status
	if report.Failed > 0 {
		fmt.Fprintf(r.writer, "Overall Status: FAILED (%d violations found)\n", report.Failed)
	} else if report.Warnings > 0 {
		fmt.Fprintf(r.writer, "Overall Status: PASSED (with %d warnings)\n", report.Warnings)
	} else {
		fmt.Fprintln(r.writer, "Overall Status: PASSED")
	}

	fmt.Fprintln(r.writer, "")
}

func (r *Reporter) printFailure(result Result) {
	fmt.Fprintf(r.writer, "✗ %s: %s\n", result.Tool, result.Message)

	if result.Details != nil {
		if suggestion, ok := result.Details["suggestion"].(string); ok {
			fmt.Fprintf(r.writer, "  Suggestion: %s\n", suggestion)
		}

		if expected, ok := result.Details["expected"]; ok {
			fmt.Fprintf(r.writer, "  Expected: %v\n", expected)
		}

		if actual, ok := result.Details["actual"]; ok {
			fmt.Fprintf(r.writer, "  Actual:   %v\n", actual)
		}

		if tiers, ok := result.Details["tiers"].(string); ok {
			fmt.Fprintln(r.writer, "  Tier-based ordering:")
			for _, line := range splitLines(tiers) {
				fmt.Fprintf(r.writer, "    %s\n", line)
			}
		}

		if ref, ok := result.Details["reference"].(string); ok {
			fmt.Fprintf(r.writer, "  Reference: %s\n", ref)
		}
	}

	fmt.Fprintf(r.writer, "  Severity: %s\n", result.Severity)
	fmt.Fprintln(r.writer, "")
}

func (r *Reporter) printWarning(result Result) {
	fmt.Fprintf(r.writer, "⚠ %s: %s\n", result.Tool, result.Message)
	fmt.Fprintf(r.writer, "  Severity: %s\n", result.Severity)
	fmt.Fprintln(r.writer, "")
}

func groupResultsByTool(results []Result) map[string][]Result {
	grouped := make(map[string][]Result)

	for _, result := range results {
		grouped[result.Tool] = append(grouped[result.Tool], result)
	}

	return grouped
}

func splitLines(s string) []string {
	var lines []string
	for _, line := range []string{s} {
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}
