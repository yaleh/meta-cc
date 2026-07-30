package analyzer

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/yaleh/meta-cc/internal/types"
)

var markerPattern = regexp.MustCompile(`\b(TODO|FIXME|HACK|XXX)\b`)

// MarkerCount holds the count for a single debt marker label.
type MarkerCount struct {
	Label string `json:"label"`
	Count int    `json:"count"`
}

// Provenance identifies which scan bucket a FileDebt entry came from, so
// merged output is auditable (DIR-055).
type Provenance string

const (
	// ProvenanceSession marks entries counted from session-transcript tool
	// output text (GetTechDebt).
	ProvenanceSession Provenance = "session"
	// ProvenanceSource marks entries counted from a source-dir walk
	// (ScanSourceDir).
	ProvenanceSource Provenance = "source"
	// ProvenanceBoth marks entries whose path appeared in both the session
	// and source buckets (MergeTechDebtResults).
	ProvenanceBoth Provenance = "both"
)

// FileDebt holds per-file marker count for hotspot ranking. Provenance
// records which bucket produced the entry; it is omitted from JSON when
// unset (e.g. hand-built fixtures).
type FileDebt struct {
	File        string     `json:"file"`
	MarkerCount int        `json:"marker_count"`
	Provenance  Provenance `json:"provenance,omitempty"`
}

// TechDebtResult is the output of GetTechDebt. Provenance follows ADR-007:
// DataSource records the dominant provenance ("measured" — Markers and
// HotspotFiles are scanned directly from tool output text), and
// EstimatedFields lists the heuristic-derived fields by JSON name. OpenIssues
// is heuristic (error call with no subsequent success for the same tool), so
// session-scan results carry EstimatedFields ["open_issues"].
type TechDebtResult struct {
	Markers         []MarkerCount `json:"markers"`
	HotspotFiles    []FileDebt    `json:"hotspot_files"`
	OpenIssues      int           `json:"open_issues"`
	DataSource      DataSource    `json:"data_source"`
	EstimatedFields []string      `json:"estimated_fields,omitempty"`
	// Warnings names any session files skipped during load (DIR-018).
	Warnings []string `json:"warnings,omitempty"`
}

// scannerToolNames is the set of tool names whose Output we scan for markers.
var scannerToolNames = map[string]bool{
	"Read":  true,
	"Edit":  true,
	"Write": true,
	"Bash":  true,
}

// langSyntax describes the cheap lexical cues markerScanner uses for one
// file extension. It is NOT a language parser — just enough state to keep
// markers inside string/regex literals out of the count while admitting
// markers inside comments.
type langSyntax struct {
	lineComment  string // token that starts a comment running to end-of-line ("//" or "#"; "" = none)
	blockComment bool   // supports /* ... */ block comments that may span lines
	rawString    bool   // supports backtick raw strings that may span lines
	singleQuote  bool   // treats ' as a string delimiter (false for Rust, where ' starts lifetimes)
}

// extSyntax maps scanned file extensions to their comment/lexical syntax.
// All extensions are covered; an extension missing here would fall back to
// counting markers anywhere on the line (legacy behavior).
var extSyntax = map[string]langSyntax{
	".go":   {lineComment: "//", blockComment: true, rawString: true, singleQuote: true},
	".ts":   {lineComment: "//", blockComment: true, rawString: true, singleQuote: true},
	".js":   {lineComment: "//", blockComment: true, rawString: true, singleQuote: true},
	".tsx":  {lineComment: "//", blockComment: true, rawString: true, singleQuote: true},
	".jsx":  {lineComment: "//", blockComment: true, rawString: true, singleQuote: true},
	".java": {lineComment: "//", blockComment: true, singleQuote: true},
	".rs":   {lineComment: "//", blockComment: true},
	".c":    {lineComment: "//", blockComment: true, singleQuote: true},
	".h":    {lineComment: "//", blockComment: true, singleQuote: true},
	".cpp":  {lineComment: "//", blockComment: true, singleQuote: true},
	".hpp":  {lineComment: "//", blockComment: true, singleQuote: true},
	".py":   {lineComment: "#", singleQuote: true},
	".sh":   {lineComment: "#", singleQuote: true},
	".yaml": {lineComment: "#", singleQuote: true},
	".yml":  {lineComment: "#", singleQuote: true},
	".toml": {lineComment: "#", singleQuote: true},
}

// knownCodeExtensions is the set of file extensions scanned during a
// source-dir walk, derived from extSyntax so every scanned extension has a
// known comment syntax. NOTE (DIR-055): .md and .json were removed — prose
// and data files are documentation, not code debt, and previously dominated
// the hotspot ranking (e.g. tasks/*.md task prose, JSON fixtures).
var knownCodeExtensions = func() map[string]bool {
	m := make(map[string]bool, len(extSyntax))
	for ext := range extSyntax {
		m[ext] = true
	}
	return m
}()

// hiddenDirNames is the set of directory names skipped during source-dir walk.
var hiddenDirNames = map[string]bool{
	".git": true, "node_modules": true, "vendor": true, ".venv": true, "__pycache__": true,
}

// markerScanner is a line-oriented heuristic that decides whether debt
// markers on a source line appear in a comment context. It does NOT parse
// the language; per extension it tracks // or # line comments, /* */ block
// comments (spanning lines), backtick raw strings (spanning lines), and
// "..." / '...' quoted spans (within a line, with backslash escapes).
// Markers inside quoted spans — string literals, regex literals, tool
// description strings, test fixtures — are excluded; markers inside
// comments are counted. It does NOT model nested or exotic syntax (Java
// text blocks, Python triple-quoted strings, shell heredocs); where such
// constructs matter it conservatively under-counts rather than reporting
// literal text as debt.
type markerScanner struct {
	syntax         langSyntax
	inBlockComment bool // carried across lines
	inRawString    bool // carried across lines
}

var (
	blockCommentStart = []byte("/*")
	blockCommentEnd   = []byte("*/")
)

// newMarkerScanner returns a scanner for the file extension ext. Unknown
// extensions get a zero syntax, which counts markers anywhere (legacy
// fallback); unreachable in practice since knownCodeExtensions is derived
// from extSyntax.
func newMarkerScanner(ext string) *markerScanner {
	return &markerScanner{syntax: extSyntax[ext]}
}

// countLine returns the debt-marker labels found in comment context on line.
// Scanner state (open block comment / raw string) carries across calls.
func (m *markerScanner) countLine(line []byte) []string {
	var found []string
	s := m.syntax
	for i := 0; i < len(line); {
		switch {
		case m.inBlockComment:
			// Marker text inside a block comment counts; scan up to */.
			if j := bytes.Index(line[i:], blockCommentEnd); j >= 0 {
				found = appendMarkerLabels(found, line[i:i+j])
				i += j + len(blockCommentEnd)
				m.inBlockComment = false
			} else {
				return appendMarkerLabels(found, line[i:])
			}
		case m.inRawString:
			// Marker text inside a raw string never counts; skip to closing `.
			if j := bytes.IndexByte(line[i:], '`'); j >= 0 {
				i += j + 1
				m.inRawString = false
			} else {
				return found
			}
		case s.blockComment && bytes.HasPrefix(line[i:], blockCommentStart):
			m.inBlockComment = true
			i += len(blockCommentStart)
		case s.rawString && line[i] == '`':
			m.inRawString = true
			i++
		case s.lineComment != "" && bytes.HasPrefix(line[i:], []byte(s.lineComment)):
			// Rest of line is a comment: count markers from here.
			return appendMarkerLabels(found, line[i:])
		case line[i] == '"':
			i = skipQuotedSpan(line, i)
		case s.singleQuote && line[i] == '\'':
			i = skipQuotedSpan(line, i)
		default:
			i++
		}
	}
	return found
}

// skipQuotedSpan consumes a quoted span starting at the opening quote at
// line[i], honoring backslash escapes. Returns the index after the closing
// quote, or len(line) if the span is unterminated on this line (quoted
// spans are not carried across lines: terminating early is the
// conservative direction for unterminated multi-line strings).
func skipQuotedSpan(line []byte, i int) int {
	quote := line[i]
	for j := i + 1; j < len(line); j++ {
		if line[j] == '\\' {
			j++ // skip escaped char
			continue
		}
		if line[j] == quote {
			return j + 1
		}
	}
	return len(line)
}

// appendMarkerLabels appends the debt-marker labels matched in segment.
func appendMarkerLabels(found []string, segment []byte) []string {
	for _, m := range markerPattern.FindAll(segment, -1) {
		found = append(found, string(m))
	}
	return found
}

// buildTechDebtResult crystallizes the shared build+sort block previously
// duplicated across GetTechDebt, ScanSourceDir, and MergeTechDebtResults
// (DIR-055): markers sorted by count descending, hotspot files sorted by
// marker count descending then path ascending.
func buildTechDebtResult(labelCounts map[string]int, fileDebt map[string]FileDebt, openIssues int, source DataSource, estimatedFields []string) *TechDebtResult {
	var markers []MarkerCount
	for label, count := range labelCounts {
		markers = append(markers, MarkerCount{Label: label, Count: count})
	}
	sort.Slice(markers, func(i, j int) bool {
		return markers[i].Count > markers[j].Count
	})

	var hotspots []FileDebt
	for _, fd := range fileDebt {
		hotspots = append(hotspots, fd)
	}
	sort.Slice(hotspots, func(i, j int) bool {
		if hotspots[i].MarkerCount != hotspots[j].MarkerCount {
			return hotspots[i].MarkerCount > hotspots[j].MarkerCount
		}
		return hotspots[i].File < hotspots[j].File
	})

	return &TechDebtResult{
		Markers:         markers,
		HotspotFiles:    hotspots,
		OpenIssues:      openIssues,
		DataSource:      source,
		EstimatedFields: estimatedFields,
	}
}

// unionEstimatedFields returns the deduplicated union of two ADR-007
// estimated_fields lists, preserving first-seen order. Returns nil when the
// union is empty so the JSON field stays omitted.
func unionEstimatedFields(a, b []string) []string {
	seen := make(map[string]bool, len(a)+len(b))
	var out []string
	for _, f := range a {
		if !seen[f] {
			seen[f] = true
			out = append(out, f)
		}
	}
	for _, f := range b {
		if !seen[f] {
			seen[f] = true
			out = append(out, f)
		}
	}
	return out
}

// GetTechDebt scans toolCalls for debt markers (the four labels matched by
// markerPattern) in outputs and detects unresolved errors (tool calls with
// status "error" that have no subsequent success call with the same tool
// name). Hotspot entries carry provenance "session".
func GetTechDebt(entries []types.SessionEntry, toolCalls []types.ToolCall) (*TechDebtResult, error) {
	labelCounts := make(map[string]int)
	fileDebt := make(map[string]FileDebt)

	for _, tc := range toolCalls {
		if !scannerToolNames[tc.ToolName] {
			continue
		}
		matches := markerPattern.FindAllString(tc.Output, -1)
		if len(matches) == 0 {
			continue
		}
		for _, m := range matches {
			labelCounts[m]++
		}
		if fp := getFilePath(tc.Input); fp != "" {
			fd := fileDebt[fp]
			fd.File = fp
			fd.MarkerCount += len(matches)
			fd.Provenance = ProvenanceSession
			fileDebt[fp] = fd
		}
	}

	// Detect open issues: error calls with no subsequent success for same tool
	openIssues := 0
	for i, tc := range toolCalls {
		if tc.Status != "error" {
			continue
		}
		fixed := false
		for j := i + 1; j < len(toolCalls); j++ {
			if toolCalls[j].ToolName == tc.ToolName && toolCalls[j].Status == "success" {
				fixed = true
				break
			}
		}
		if !fixed {
			openIssues++
		}
	}

	// OpenIssues is heuristic (ADR-007): the session scan always applies it, so
	// its field provenance is "estimated" regardless of the resulting count.
	return buildTechDebtResult(labelCounts, fileDebt, openIssues, DataSourceMeasured, []string{"open_issues"}), nil
}

// ScanSourceDir walks sourceDir recursively, scanning known code files for
// debt markers in comment context (see markerScanner). Hidden directories
// (.git, node_modules, etc.) are skipped; documentation and data files
// (.md, .json) are not scanned. Scanning stops after maxFiles (safety cap).
// The returned result keeps dominant DataSourceMeasured provenance but lists
// "markers" and "hotspot_files" as estimated (ADR-007): markerScanner is a
// language-aware lexical heuristic, not a parser, so its comment-context
// classification can under-count unsupported syntax. Hotspot entries carry
// provenance "source".
func ScanSourceDir(sourceDir string) (*TechDebtResult, error) {
	labelCounts := make(map[string]int)
	fileDebt := make(map[string]FileDebt)
	filesScanned := 0
	const maxFiles = 10000

	err := filepath.Walk(sourceDir, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip unreadable entries
		}
		if info.IsDir() {
			base := info.Name()
			if hiddenDirNames[base] {
				return filepath.SkipDir
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(p))
		if !knownCodeExtensions[ext] {
			return nil
		}
		filesScanned++
		if filesScanned > maxFiles {
			return nil // silently stop scanning
		}

		f, err := os.Open(p)
		if err != nil {
			return nil
		}

		// NOTE(DIR-038): this hand-rolled bufio.NewReader + ReadBytes('\n')
		// loop predates and duplicates the shape now crystallized in
		// parser.ReadLineBounded (internal/parser/bounded_reader.go).
		// Migrating it is tracked as optional/best-effort follow-up, not
		// done here, to keep this task's diff scoped to closing the
		// check-no-scanner violation in
		// internal/provider/codex/appserver/client.go.
		scanner := newMarkerScanner(ext)
		reader := bufio.NewReader(f)
		lineCount := 0
		fileTotal := 0
		for {
			line, readErr := reader.ReadBytes('\n')
			if len(line) > 0 {
				lineCount++
				// Cap lines per file to avoid pathological files
				if lineCount > 20000 {
					break
				}
				if matches := scanner.countLine(line); len(matches) > 0 {
					for _, m := range matches {
						labelCounts[m]++
					}
					fileTotal += len(matches)
				}
			}
			if readErr != nil {
				break
			}
		}
		f.Close()
		if fileTotal > 0 {
			fileDebt[p] = FileDebt{File: p, MarkerCount: fileTotal, Provenance: ProvenanceSource}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return buildTechDebtResult(labelCounts, fileDebt, 0, DataSourceMeasured, []string{"markers", "hotspot_files"}), nil
}

// combineProvenance merges the provenance of two entries for the same path:
// equal values stay, an empty side yields to the other, and two distinct
// non-empty values combine into "both".
func combineProvenance(a, b Provenance) Provenance {
	switch {
	case a == b:
		return a
	case a == "":
		return b
	case b == "":
		return a
	default:
		return ProvenanceBoth
	}
}

// MergeTechDebtResults merges two TechDebtResults into one. Label marker
// counts sum (they are aggregate totals), but per-file hotspot counts take
// the MAX per path, not the sum: a file both Read during a session and
// source-scanned carries the same on-disk markers in both buckets, so
// summing double-counts it (DIR-055). Each merged hotspot entry records
// provenance ("session", "source", or "both"). OpenIssues uses the max of
// both inputs. The DataSource is set to the provided combinedSource, and the
// EstimatedFields lists of both inputs are unioned (ADR-007): a merged result
// that carries a session-scan OpenIssues value inherits "open_issues".
func MergeTechDebtResults(a, b *TechDebtResult, combinedSource DataSource) *TechDebtResult {
	labelCounts := make(map[string]int)
	fileDebt := make(map[string]FileDebt)

	for _, m := range a.Markers {
		labelCounts[m.Label] += m.Count
	}
	for _, m := range b.Markers {
		labelCounts[m.Label] += m.Count
	}

	for _, f := range a.HotspotFiles {
		fileDebt[f.File] = f
	}
	for _, f := range b.HotspotFiles {
		existing, ok := fileDebt[f.File]
		if !ok {
			fileDebt[f.File] = f
			continue
		}
		if f.MarkerCount > existing.MarkerCount {
			existing.MarkerCount = f.MarkerCount
		}
		existing.Provenance = combineProvenance(existing.Provenance, f.Provenance)
		fileDebt[f.File] = existing
	}

	openIssues := a.OpenIssues
	if b.OpenIssues > openIssues {
		openIssues = b.OpenIssues
	}

	return buildTechDebtResult(labelCounts, fileDebt, openIssues, combinedSource, unionEstimatedFields(a.EstimatedFields, b.EstimatedFields))
}

// TechDebtStats holds aggregate-only tech debt output: marker counts (bounded
// to the four known marker labels) and a hotspot file *count*, omitting the
// full HotspotFiles path list — which can grow to one entry per matched file
// across an entire scanned source tree — and any other per-item detail
// (DIR-042). Provenance is preserved per ADR-007: DataSource and
// EstimatedFields pass through from the source result.
type TechDebtStats struct {
	MarkerCount      int           `json:"marker_count"`
	Markers          []MarkerCount `json:"markers"`
	HotspotFileCount int           `json:"hotspot_file_count"`
	OpenIssueCount   int           `json:"open_issue_count"`
	DataSource       DataSource    `json:"data_source"`
	EstimatedFields  []string      `json:"estimated_fields,omitempty"`
	// Warnings names any session files skipped during load (DIR-018).
	Warnings []string `json:"warnings,omitempty"`
}

// TechDebtResultStats converts an already-computed TechDebtResult (which may
// already be a session+source_dir merge, per MergeTechDebtResults) into an
// aggregate-only stats view with no per-file hotspot list.
func TechDebtResultStats(result *TechDebtResult) *TechDebtStats {
	total := 0
	for _, m := range result.Markers {
		total += m.Count
	}
	return &TechDebtStats{
		MarkerCount:      total,
		Markers:          result.Markers,
		HotspotFileCount: len(result.HotspotFiles),
		OpenIssueCount:   result.OpenIssues,
		DataSource:       result.DataSource,
		EstimatedFields:  techDebtStatsEstimatedFields(result.EstimatedFields),
		Warnings:         result.Warnings,
	}
}

func techDebtStatsEstimatedFields(fields []string) []string {
	mapped := make([]string, 0, len(fields))
	for _, field := range fields {
		switch field {
		case "markers":
			mapped = append(mapped, "markers", "marker_count")
		case "hotspot_files":
			mapped = append(mapped, "hotspot_file_count")
		case "open_issues":
			mapped = append(mapped, "open_issue_count")
		}
	}
	return mapped
}

// getFilePath extracts file path from tool input map.
func getFilePath(input map[string]interface{}) string {
	for _, key := range []string{"file_path", "path"} {
		if v, ok := input[key]; ok {
			if s, ok := v.(string); ok && s != "" {
				return s
			}
		}
	}
	return ""
}
