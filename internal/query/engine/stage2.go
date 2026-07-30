package engine

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/itchyny/gojq"

	"github.com/yaleh/meta-cc/internal/parser"
	querycommon "github.com/yaleh/meta-cc/internal/query/common"
	"github.com/yaleh/meta-cc/internal/session"
)

const (
	// preflightSampleTotal bounds how many representative records the Stage 2
	// preflight executes the caller's jq against before full-corpus execution.
	preflightSampleTotal = 32
	// preflightSamplePerType bounds how many records of each distinct `type`
	// value are included, so the sample stays diverse across record shapes
	// (user string content vs assistant array content) rather than being
	// dominated by one type.
	preflightSamplePerType = 8
	// maxSkipWarnings bounds how many per-file skip reasons are surfaced in
	// diagnostics, so a fully corrupt corpus cannot flood the response.
	maxSkipWarnings = 5
)

// Stage2Query represents a Stage 2 query request.
type Stage2Query struct {
	Files     []string // Absolute file paths to query
	Filter    string   // jq filter expression (required)
	Sort      string   // jq sort expression (optional)
	Transform string   // jq transform expression (optional)
	Limit     int      // Maximum number of results (0 = no limit)
}

// Stage2Result represents the result of a Stage 2 query.
type Stage2Result struct {
	Results     []interface{}    `json:"results"`
	Metadata    QueryMetadata    `json:"metadata"`
	Diagnostics QueryDiagnostics `json:"diagnostics"`
	Warnings    []string         `json:"warnings,omitempty"`
}

// QueryMetadata contains metadata about the query execution.
type QueryMetadata struct {
	ExecutionTimeMs     int64 `json:"execution_time_ms"`
	FilesProcessed      int   `json:"files_processed"`
	TotalRecordsScanned int   `json:"total_records_scanned"`
	ResultsReturned     int   `json:"results_returned"`
	Truncated           bool  `json:"truncated"`
}

// QueryDiagnostics is the uniform, bounded query-accounting envelope emitted
// alongside query results (DIR-079 / ADR-009). It lets callers distinguish an
// empty result from an incomplete or degraded search without source
// inspection: requested vs effective backend, files considered/loaded/skipped,
// records scanned, matches returned, and whether execution was degraded.
type QueryDiagnostics struct {
	Backend           string   `json:"backend"`
	Provider          string   `json:"provider"`
	ProviderEffective string   `json:"provider_effective"`
	FilesConsidered   int      `json:"files_considered"`
	FilesLoaded       int      `json:"files_loaded"`
	FilesSkipped      int      `json:"files_skipped"`
	RecordsScanned    int      `json:"records_scanned"`
	MatchesReturned   int      `json:"matches_returned"`
	Truncated         bool     `json:"truncated"`
	Degraded          bool     `json:"degraded"`
	SkipWarnings      []string `json:"skip_warnings,omitempty"`
}

// ExecuteStage2Query executes a Stage 2 query on selected files.
func ExecuteStage2Query(query *Stage2Query) (*Stage2Result, error) {
	start := time.Now()

	if len(query.Files) == 0 {
		return nil, fmt.Errorf("files parameter cannot be empty")
	}
	if query.Filter == "" {
		return nil, fmt.Errorf("filter parameter is required")
	}

	jqExpr := buildJQExpression(query.Filter, query.Sort, query.Transform)

	compiled, err := gojq.Parse(jqExpr)
	if err != nil {
		return nil, fmt.Errorf("invalid jq expression '%s': %w", jqExpr, err)
	}

	// Bounded preflight (DIR-079 / ADR-009): run the caller's jq over a small,
	// type-diverse sample of records BEFORE full-corpus execution so common
	// type/path mismatches (e.g. test() applied to an object) fail fast with an
	// actionable diagnostic instead of a low-level error deep in the scan. The
	// caller's jq is never rewritten; valid queries pass through untouched.
	if perr := preflightTypeCheck(query.Files, compiled, jqExpr); perr != nil {
		return nil, perr
	}

	results, metadata, diag, err := streamFilesWithJQ(query.Files, compiled, query.Limit)
	if err != nil {
		return nil, err
	}

	metadata.ExecutionTimeMs = time.Since(start).Milliseconds()

	result := &Stage2Result{
		Results:     results,
		Metadata:    *metadata,
		Diagnostics: *diag,
	}

	if query.Transform != "" && querycommon.AllNullOrEmpty(results) {
		result.Warnings = append(result.Warnings, "transform produced all-null fields: check your field paths (e.g. .timestamp not .Timestamp). Use inspect_session_files(include_samples=true) to see actual field names.")
	}

	return result, nil
}

func buildJQExpression(filter, sort, transform string) string {
	if sort != "" {
		var parts []string

		if filter != "" {
			parts = append(parts, fmt.Sprintf("[.[] | %s]", filter))
		} else {
			parts = append(parts, "[.[]]")
		}

		parts = append(parts, sort)
		parts = append(parts, ".[]")

		if transform != "" {
			parts = append(parts, transform)
		}

		return strings.Join(parts, " | ")
	}

	parts := []string{".[]"}

	if filter != "" {
		parts = append(parts, filter)
	}

	if transform != "" {
		parts = append(parts, transform)
	}

	return strings.Join(parts, " | ")
}

func streamFilesWithJQ(files []string, compiled *gojq.Query, limit int) ([]interface{}, *QueryMetadata, *QueryDiagnostics, error) {
	metadata := &QueryMetadata{}
	diag := &QueryDiagnostics{
		Backend:           "stage2_jq",
		Provider:          "explicit",
		ProviderEffective: "explicit",
		FilesConsidered:   len(files),
	}
	var results []interface{}

	for _, file := range files {
		records, err := readJSONLFile(file)
		if err != nil {
			// Degrade gracefully (DIR-079 / ADR-009): skip the unreadable file
			// with a bounded warning instead of aborting the whole corpus scan.
			diag.FilesSkipped++
			diag.Degraded = true
			if len(diag.SkipWarnings) < maxSkipWarnings {
				diag.SkipWarnings = append(diag.SkipWarnings,
					fmt.Sprintf("skipped unreadable file %s: %v", file, err))
			}
			continue
		}

		diag.FilesLoaded++
		metadata.FilesProcessed++
		metadata.TotalRecordsScanned += len(records)
		diag.RecordsScanned += len(records)

		iter := compiled.Run(records)
		for {
			if limit > 0 && metadata.ResultsReturned >= limit {
				metadata.Truncated = true
				diag.Truncated = true
				diag.MatchesReturned = len(results)
				return results, metadata, diag, nil
			}

			value, ok := iter.Next()
			if !ok {
				break
			}

			if err, ok := value.(error); ok {
				return nil, nil, nil, fmt.Errorf("jq execution error: %w", err)
			}

			results = append(results, value)
			metadata.ResultsReturned++

			if limit > 0 && metadata.ResultsReturned >= limit {
				metadata.Truncated = true
				diag.Truncated = true
				diag.MatchesReturned = len(results)
				return results, metadata, diag, nil
			}
		}
	}

	if diag.FilesLoaded == 0 {
		reason := ""
		if len(diag.SkipWarnings) > 0 {
			reason = ": " + diag.SkipWarnings[0]
		}
		return nil, nil, nil, fmt.Errorf("no session files could be loaded (%d considered, %d skipped)%s",
			diag.FilesConsidered, diag.FilesSkipped, reason)
	}

	diag.MatchesReturned = len(results)
	return results, metadata, diag, nil
}

// jqTypeErrorRe extracts the observed input type from a gojq type error, e.g.
// `test("x"; null) cannot be applied to: object ({...})` -> "object".
var jqTypeErrorRe = regexp.MustCompile(`cannot be applied to:\s*([a-z]+)`)

// jqFuncNameRe extracts the leading function name from a gojq error message.
var jqFuncNameRe = regexp.MustCompile(`^([a-zA-Z_][a-zA-Z0-9_]*)`)

// preflightTypeCheck runs the compiled jq over a bounded, type-diverse sample
// of records and, if it surfaces a jq type error, returns an actionable
// diagnostic BEFORE full-corpus execution (DIR-079 / ADR-009). Non-type errors
// and unreadable inputs are left to full execution so valid-query behavior is
// unchanged. The caller's jq is never rewritten.
func preflightTypeCheck(files []string, compiled *gojq.Query, expr string) error {
	sample := collectPreflightSample(files, preflightSampleTotal, preflightSamplePerType)
	if len(sample) == 0 {
		return nil
	}
	iter := compiled.Run(sample)
	for {
		value, ok := iter.Next()
		if !ok {
			break
		}
		if runErr, isErr := value.(error); isErr {
			if diag := classifyJQTypeError(runErr, expr); diag != "" {
				return fmt.Errorf("%s", diag)
			}
			// Not a recognized type error: let full execution report it.
			return nil
		}
	}
	return nil
}

// classifyJQTypeError builds an actionable diagnostic from a jq type error,
// naming the observed input type and suggesting at least one correction (a
// coercion or a valid field path). It returns "" when the error is not a
// recognized type error, so callers can fall through to existing behavior.
func classifyJQTypeError(err error, expr string) string {
	msg := err.Error()
	m := jqTypeErrorRe.FindStringSubmatch(msg)
	if m == nil {
		return ""
	}
	observed := m[1]
	funcName := ""
	if fm := jqFuncNameRe.FindStringSubmatch(msg); fm != nil {
		funcName = fm[1]
	}

	var b strings.Builder
	b.WriteString("stage 2 preflight: jq type error detected on representative input, before full-corpus execution.\n")
	fmt.Fprintf(&b, "  expression: %s\n", expr)
	fmt.Fprintf(&b, "  observed input type: %s\n", observed)
	fmt.Fprintf(&b, "  detail: %s\n", msg)
	b.WriteString("  corrections:\n")
	if needsStringInput(funcName) && observed != "string" {
		if funcName != "" {
			fmt.Fprintf(&b, "    - `%s` expects a string but received a value of type %q; select a string field path (e.g. .message.content for user text, or .message.content[].text for assistant content blocks).\n", funcName, observed)
			fmt.Fprintf(&b, "    - or coerce explicitly: (.<field> | tostring) | %s(...).\n", funcName)
		}
	} else {
		fmt.Fprintf(&b, "    - the function received a value of type %q; select a field path whose type matches what the function expects.\n", observed)
	}
	b.WriteString("    - use inspect_session_files(include_samples=true) to discover the actual field names and value types.")
	return b.String()
}

// needsStringInput reports whether a jq builtin requires string input (and so
// fails on object/array values), for targeted preflight corrections.
func needsStringInput(funcName string) bool {
	switch funcName {
	case "test", "match", "capture", "split", "ltrimstr", "rtrimstr":
		return true
	}
	return false
}

// collectPreflightSample reads a bounded, type-diverse sample of normalized
// records across the given files. It caps both the total sample size and the
// per-`type` count so one dominant record type cannot crowd out the shapes
// (e.g. assistant array content) most likely to trigger a jq type error.
// Unreadable files are skipped silently; full execution reports them.
func collectPreflightSample(files []string, maxTotal, maxPerType int) []interface{} {
	var sample []interface{}
	perType := make(map[string]int)
	for _, file := range files {
		if len(sample) >= maxTotal {
			break
		}
		records, err := readJSONLFileBounded(file, maxTotal)
		if err != nil {
			continue
		}
		for _, rec := range records {
			if len(sample) >= maxTotal {
				break
			}
			key := recordTypeKey(rec)
			if perType[key] >= maxPerType {
				continue
			}
			perType[key]++
			sample = append(sample, rec)
		}
	}
	return sample
}

// recordTypeKey returns the record's `type` field for sample diversity, or ""
// when absent/non-string.
func recordTypeKey(rec interface{}) string {
	if m, ok := rec.(map[string]interface{}); ok {
		if t, ok := m["type"].(string); ok {
			return t
		}
	}
	return ""
}

// readJSONLFileBounded reads and normalizes up to max records from a JSONL
// file, stopping once the cap is reached so the Stage 2 preflight stays
// bounded even for very large session files.
func readJSONLFileBounded(filepath string, max int) ([]interface{}, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	r := bufio.NewReader(file)
	rawMessages, err := parser.ReadAllFiltered(r, parser.StrategyDefault)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	records := make([]interface{}, 0, max)
	normalizer := session.NewNormalizer()
	for _, raw := range rawMessages {
		if len(records) >= max {
			break
		}
		line := strings.TrimSpace(string(raw))
		if line == "" {
			continue
		}
		var record map[string]interface{}
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			continue
		}
		for _, normalized := range normalizer.NormalizeRecord(record) {
			records = append(records, normalized)
			if len(records) >= max {
				break
			}
		}
	}
	return records, nil
}

func readJSONLFile(filepath string) ([]interface{}, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	r := bufio.NewReader(file)
	rawMessages, err := parser.ReadAllFiltered(r, parser.StrategyDefault)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	lineNum := 0
	records := make([]interface{}, 0, len(rawMessages))
	normalizer := session.NewNormalizer()
	for _, raw := range rawMessages {
		lineNum++
		line := strings.TrimSpace(string(raw))
		if line == "" {
			continue
		}
		var record map[string]interface{}
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			return nil, fmt.Errorf("invalid JSON at line %d: %w", lineNum, err)
		}
		for _, normalized := range normalizer.NormalizeRecord(record) {
			records = append(records, normalized)
		}
	}

	return records, nil
}
