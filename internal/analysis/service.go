// Package analysis provides a service facade that encapsulates the full
// pipeline of: locate session files → parse → run analyzer functions.
// cmd/mcp-server uses this package instead of importing internal/parser
// and internal/analyzer directly.
package analysis

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/yaleh/meta-cc/internal/analyzer"
	"github.com/yaleh/meta-cc/internal/config"
	"github.com/yaleh/meta-cc/internal/locator"
	"github.com/yaleh/meta-cc/internal/parser"
	"github.com/yaleh/meta-cc/internal/provider/rawfiles"
	providerrecords "github.com/yaleh/meta-cc/internal/provider/records"
	"github.com/yaleh/meta-cc/internal/types"
)

// Analyzers holds the injected analyzer interfaces used by Service.
// Zero-value fields are replaced with DefaultAnalyzer instances at construction time.
//
// Contract: the injected implementations serve full (non-stats_only) analysis
// calls only. When args carry stats_only:true, five of the seven Service
// methods (AnalyzeBugs, AnalyzeErrors, QualityScan, GetWorkPatterns,
// GetTimeline) short-circuit to package-level analyzer.*Stats* aggregations
// and never invoke the injected analyzer (DIR-042, enforced by
// service_stats_only_test.go). GetTechDebt differs deliberately: it always
// calls the injected TechDebt analyzer first (the full result is required for
// the optional source_dir merge) and stats-converts that result afterwards.
// QueryEditSequences has no stats_only parameter and always runs directly.
type Analyzers struct {
	BugAnalyzer    analyzer.BugAnalyzer
	ErrorAnalyzer  analyzer.ErrorAnalyzer
	QualityScanner analyzer.QualityScanner
	WorkPatterns   analyzer.WorkPatternsAnalyzer
	Timeline       analyzer.TimelineAnalyzer
	TechDebt       analyzer.TechDebtAnalyzer
}

// Service encapsulates the analysis pipeline for MCP tool handlers.
type Service struct {
	analyzers Analyzers
}

// New creates a new Service backed by the default (real) analyzer implementations.
func New() *Service {
	return NewWithAnalyzers(Analyzers{})
}

// NewWithAnalyzers creates a new Service with the provided analyzer interfaces.
// Any nil field is replaced with the corresponding DefaultAnalyzer method.
// Note the stats_only contract documented on Analyzers: injected
// implementations are bypassed on stats_only calls (except GetTechDebt,
// which stats-converts the injected analyzer's result).
func NewWithAnalyzers(a Analyzers) *Service {
	d := &analyzer.DefaultAnalyzer{}
	if a.BugAnalyzer == nil {
		a.BugAnalyzer = d
	}
	if a.ErrorAnalyzer == nil {
		a.ErrorAnalyzer = d
	}
	if a.QualityScanner == nil {
		a.QualityScanner = d
	}
	if a.WorkPatterns == nil {
		a.WorkPatterns = d
	}
	if a.Timeline == nil {
		a.Timeline = d
	}
	if a.TechDebt == nil {
		a.TechDebt = d
	}
	return &Service{analyzers: a}
}

// loadData locates session files, parses them, and extracts tool calls.
// It supports "project" (default) and "session" scopes, and an optional
// working_dir override extracted from args.
//
// The returned SkipReport carries one entry per corpus file that was excluded
// (DIR-018, unified across every enumerating path by DIR-094): previously such
// files were skipped with a bare `continue`, so analysis results silently
// excluded data. Each skipped file is also logged at WARN level. Callers must
// surface the report in their marshaled results so MCP responses never hide
// data exclusion — see marshalResult, which attaches both the human-readable
// `warnings` and the machine-readable `skipped_files`.
//
// DIR-094 closed the second half of this gap on the Claude path: a file that
// parses cleanly but yields zero message entries (the 0-byte / metadata-only
// session stub) used to pass through silently, because ParseEntries returns no
// error for it. Such a file is now reported exactly like a parse failure.
func (s *Service) loadData(args map[string]interface{}) ([]types.SessionEntry, []types.ToolCall, *locator.SkipReport, error) {
	scope := "project"
	if v, ok := args["scope"].(string); ok && v != "" {
		scope = v
	}

	workingDir := ""
	if v, ok := args["working_dir"].(string); ok {
		workingDir = v
	}
	if workingDir == "" {
		var err error
		workingDir, err = os.Getwd()
		if err != nil {
			workingDir = "."
		}
	}

	providerName := stringArg(args, "provider")
	// DIR-073: an omitted/empty provider resolves to the process host
	// default (config.OmittedProviderDefault), never a hard-coded claude.
	if providerName == "" {
		providerName = config.OmittedProviderDefault()
	}
	// DIR-030: session_id, when set, is an exact-thread selector distinct
	// from scope="session" ("most recent session") — it takes precedence
	// over scope entirely and reads only the one requested session.
	sessionID := stringArg(args, "session_id")

	if providerName != "claude" {
		return s.loadProviderData(scope, workingDir, providerName, sessionID)
	}

	loc := locator.NewSessionLocator()
	var files []string
	switch {
	case sessionID != "":
		// DIR-033: loc.FromSessionID (the raw, unscoped primitive) searches
		// every project-hash directory on disk for a matching
		// {session_id}.jsonl and returns whatever it finds, with no
		// comparison against workingDir — a cross-project leak letting any
		// caller who knows a session_id read that session's content
		// regardless of the working_dir they claim to be scoped to. This was
		// the third independent instance of this exact bug class (after
		// ExecuteQueryForSession in internal/mcp/executor/provider_query.go
		// and findSessionFile in internal/provider/claude/provider.go).
		// FromSessionIDScoped crystallizes the boundary check (via
		// locator.PathToHash, the same directory-naming scheme Claude Code
		// itself uses) so it is enforced once, not re-derived at each call
		// site.
		sessionFile, err := loc.FromSessionIDScoped(sessionID, workingDir)
		if err != nil {
			return nil, nil, nil, err
		}
		files = []string{sessionFile}
	case scope == "session":
		sessionFile, err := loc.FromProjectPath(workingDir)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to locate session: %w", err)
		}
		files = []string{sessionFile}
	default:
		var err error
		files, err = loc.AllSessionsFromProject(workingDir)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to locate project sessions: %w", err)
		}
	}

	skips := &locator.SkipReport{}
	var allEntries []types.SessionEntry
	for _, f := range files {
		p := parser.NewSessionParser(f)
		entries, err := p.ParseEntries()
		if err != nil {
			// DIR-018: never silently exclude data — record a warning naming
			// the file and log at WARN level instead of a bare `continue`.
			slog.Warn("skipping unparseable session file", "file", f, "error", err)
			skips.Skip(f, err)
			continue
		}
		if len(entries) == 0 {
			// DIR-094: a 0-byte file or a metadata-only session stub parses
			// cleanly and yields nothing. Tolerating that silently is the
			// same data exclusion DIR-018 made visible for parse errors, so
			// it is reported through the same channel — and uses the same
			// "no message entries" vocabulary the Claude provider's
			// errNoMessageEntries sentinel uses on the ListSessions path.
			const reason = "no message entries (zero-message session stub)"
			slog.Warn("skipping session file with no message entries", "file", f)
			skips.SkipReason(f, reason)
			continue
		}
		allEntries = append(allEntries, entries...)
	}

	toolCalls := types.ExtractToolCalls(allEntries)
	return allEntries, toolCalls, skips, nil
}

func (s *Service) loadProviderData(scope, workingDir, providerName, sessionID string) ([]types.SessionEntry, []types.ToolCall, *locator.SkipReport, error) {
	projectPath, err := filepath.Abs(workingDir)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to resolve project path: %w", err)
	}
	filters, err := rawfiles.ParseProviderFilter(providerName)
	if err != nil {
		return nil, nil, nil, err
	}
	registry := rawfiles.NewRegistry(projectPath)

	skips := &locator.SkipReport{}
	var (
		records          []map[string]interface{}
		providerWarnings []string
	)
	if sessionID != "" {
		// DIR-030 exact-session fast path: GetSession/LoadTurns for this
		// one ID only, never ListSessions across the whole project.
		records, providerWarnings, err = providerrecords.BuildForSession(context.Background(), registry, filters, sessionID, projectPath)
	} else {
		records, providerWarnings, err = providerrecords.Build(context.Background(), registry, filters, scope, projectPath)
	}
	if err != nil {
		return nil, nil, nil, err
	}
	// DIR-094: providerrecords.Build already implements DIR-030's per-session
	// skip-and-report contract, but this caller discarded its warnings — so
	// on every non-Claude provider the exclusion was computed and then thrown
	// away one frame later. Adopt them verbatim (they name a session ID, not a
	// corpus file path, so they contribute to `warnings` only).
	for _, warning := range providerWarnings {
		skips.AdoptWarning(warning)
	}
	entries, err := entriesFromRecords(records)
	if err != nil {
		return nil, nil, nil, err
	}
	return entries, types.ExtractToolCalls(entries), skips, nil
}

func entriesFromRecords(records []map[string]interface{}) ([]types.SessionEntry, error) {
	entries := make([]types.SessionEntry, 0, len(records))
	for _, record := range records {
		data, err := json.Marshal(record)
		if err != nil {
			return nil, err
		}
		var entry types.SessionEntry
		if err := json.Unmarshal(data, &entry); err != nil {
			return nil, err
		}
		if entry.IsMessage() {
			entries = append(entries, entry)
		}
	}
	return entries, nil
}

func stringArg(args map[string]interface{}, key string) string {
	if v, ok := args[key].(string); ok {
		return v
	}
	return ""
}

func intArg(args map[string]interface{}, key string) int {
	if v, ok := args[key].(float64); ok {
		return int(v)
	}
	return 0
}

func boolArg(args map[string]interface{}, key string) bool {
	if v, ok := args[key].(bool); ok {
		return v
	}
	return false
}

// marshalResult serializes an analysis result and attaches the corpus-exclusion
// metadata DIR-094 requires. The human-readable channel is the `warnings` field
// the analyzer result structs already carry (DIR-018); this adds the
// machine-readable counterpart, `skipped_files`, naming each excluded corpus
// file so a response never tolerates a bad file silently.
//
// A clean corpus (skips.Empty()) adds nothing, and a value that does not
// marshal to a JSON object — e.g. a bare array or scalar — is returned exactly
// as serialized rather than losing the whole result to a cosmetic metadata
// attach.
func marshalResult(v interface{}, skips *locator.SkipReport) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}
	skippedPaths := skips.Paths()
	if len(skippedPaths) == 0 {
		return string(data), nil
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return string(data), nil
	}
	parsed["skipped_files"] = skippedPaths

	withSkips, err := json.Marshal(parsed)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result with skipped_files: %w", err)
	}
	return string(withSkips), nil
}

// AnalyzeBugs implements the analyze_bugs MCP tool.
// When stats_only is set, short-circuits to an aggregate-only pattern-count
// summary (analyzer.AnalyzeBugsStats) with no per-pattern Examples text,
// mirroring GetTimeline's own stats_only short-circuit (DIR-042).
func (s *Service) AnalyzeBugs(args map[string]interface{}) (string, error) {
	entries, toolCalls, skips, err := s.loadData(args)
	if err != nil {
		return "", fmt.Errorf("failed to load session data: %w", err)
	}
	if boolArg(args, "stats_only") {
		stats, err := analyzer.AnalyzeBugsStats(entries, toolCalls)
		if err != nil {
			return "", fmt.Errorf("analyze bugs failed: %w", err)
		}
		stats.Warnings = skips.Warnings()
		return marshalResult(stats, skips)
	}
	result, err := s.analyzers.BugAnalyzer.AnalyzeBugs(entries, toolCalls, intArg(args, "limit"))
	if err != nil {
		return "", fmt.Errorf("analyze bugs failed: %w", err)
	}
	result.Warnings = skips.Warnings()
	return marshalResult(result, skips)
}

// AnalyzeErrors implements the analyze_errors MCP tool.
// When stats_only is set, short-circuits to an aggregate-only per-tool/
// per-type count summary (analyzer.AnalyzeErrorsStats) with no examples
// text, mirroring GetTimeline's own stats_only short-circuit (DIR-042).
func (s *Service) AnalyzeErrors(args map[string]interface{}) (string, error) {
	entries, toolCalls, skips, err := s.loadData(args)
	if err != nil {
		return "", fmt.Errorf("failed to load session data: %w", err)
	}
	if boolArg(args, "stats_only") {
		stats, err := analyzer.AnalyzeErrorsStats(entries, toolCalls)
		if err != nil {
			return "", fmt.Errorf("failed to analyze errors: %w", err)
		}
		stats.Warnings = skips.Warnings()
		return marshalResult(stats, skips)
	}
	result, err := s.analyzers.ErrorAnalyzer.AnalyzeErrors(entries, toolCalls, intArg(args, "limit"))
	if err != nil {
		return "", fmt.Errorf("failed to analyze errors: %w", err)
	}
	result.Warnings = skips.Warnings()
	return marshalResult(result, skips)
}

// QualityScan implements the quality_scan MCP tool.
// QualityScan's result is already aggregate-only (four scored dimensions,
// no per-item example text); the stats_only short-circuit
// (analyzer.QualityScanStatsOnly) exists so this method still honors the
// documented stats_only contract explicitly rather than silently ignoring it
// (DIR-042).
func (s *Service) QualityScan(args map[string]interface{}) (string, error) {
	entries, toolCalls, skips, err := s.loadData(args)
	if err != nil {
		return "", fmt.Errorf("failed to load session data: %w", err)
	}
	if boolArg(args, "stats_only") {
		stats, err := analyzer.QualityScanStatsOnly(entries, toolCalls)
		if err != nil {
			return "", fmt.Errorf("quality scan failed: %w", err)
		}
		stats.Warnings = skips.Warnings()
		return marshalResult(stats, skips)
	}
	result, err := s.analyzers.QualityScanner.QualityScan(entries, toolCalls)
	if err != nil {
		return "", fmt.Errorf("quality scan failed: %w", err)
	}
	result.Warnings = skips.Warnings()
	return marshalResult(result, skips)
}

// GetWorkPatterns implements the get_work_patterns MCP tool.
// GetWorkPatterns's result is already aggregate-only (tool counts, a fixed
// 24-slot hourly histogram, and two scalar counters, no per-item example
// text); the stats_only short-circuit (analyzer.GetWorkPatternsStatsOnly)
// exists so this method still honors the documented stats_only contract
// explicitly rather than silently ignoring it (DIR-042).
func (s *Service) GetWorkPatterns(args map[string]interface{}) (string, error) {
	entries, toolCalls, skips, err := s.loadData(args)
	if err != nil {
		return "", fmt.Errorf("failed to load session data: %w", err)
	}
	if boolArg(args, "stats_only") {
		stats, err := analyzer.GetWorkPatternsStatsOnly(entries, toolCalls)
		if err != nil {
			return "", fmt.Errorf("get work patterns failed: %w", err)
		}
		stats.Warnings = skips.Warnings()
		return marshalResult(stats, skips)
	}
	result, err := s.analyzers.WorkPatterns.GetWorkPatterns(entries, toolCalls)
	if err != nil {
		return "", fmt.Errorf("get work patterns failed: %w", err)
	}
	result.Warnings = skips.Warnings()
	return marshalResult(result, skips)
}

// timelineAutoStatsThreshold is the entry count above which get_timeline switches to
// stats summary mode automatically when no since/until clipping is provided.
// This prevents context truncation in large projects (e.g., baime with 126 sessions).
const timelineAutoStatsThreshold = 1000

// GetTimeline implements the get_timeline MCP tool.
// Supports optional since/until ISO 8601 time-clipping params.
// When no since/until is set and entry count exceeds timelineAutoStatsThreshold,
// defaults to stats summary mode to prevent context overflow.
func (s *Service) GetTimeline(args map[string]interface{}) (string, error) {
	entries, _, skips, err := s.loadData(args)
	if err != nil {
		return "", fmt.Errorf("failed to load session data: %w", err)
	}

	// Apply since/until time-clipping if provided.
	since := stringArg(args, "since")
	until := stringArg(args, "until")
	if since != "" || until != "" {
		entries, err = filterEntriesByTimeRange(entries, since, until)
		if err != nil {
			return "", err
		}
	}

	if boolArg(args, "stats_only") {
		stats := analyzer.GetTimelineStats(entries)
		stats.Warnings = skips.Warnings()
		return marshalResult(stats, skips)
	}

	// When no time clipping and entry count exceeds threshold, default to stats mode
	// to prevent the 737K+ character context truncation observed in large projects.
	if since == "" && until == "" && len(entries) > timelineAutoStatsThreshold {
		stats := analyzer.GetTimelineStats(entries)
		stats.Warnings = skips.Warnings()
		return marshalResult(stats, skips)
	}

	limit := intArg(args, "limit")
	// Auto-limit large project-scope queries to prevent context overflow.
	if limit == 0 && stringArg(args, "scope") != "session" && len(entries) > 2000 {
		limit = 500
	}

	result, err := s.analyzers.Timeline.GetTimeline(entries, limit)
	if err != nil {
		return "", fmt.Errorf("get timeline failed: %w", err)
	}
	result.Warnings = skips.Warnings()
	return marshalResult(result, skips)
}

// filterEntriesByTimeRange filters session entries to those within the since/until range.
// Both since and until are optional ISO 8601 strings. Since is inclusive, until is exclusive.
func filterEntriesByTimeRange(entries []types.SessionEntry, since, until string) ([]types.SessionEntry, error) {
	var sinceTime, untilTime time.Time
	var hasSince, hasUntil bool

	if since != "" {
		t, err := time.Parse(time.RFC3339, since)
		if err != nil {
			return nil, fmt.Errorf("invalid since value %q: must be ISO 8601 / RFC3339 (e.g. 2026-01-01T00:00:00Z)", since)
		}
		sinceTime = t
		hasSince = true
	}
	if until != "" {
		t, err := time.Parse(time.RFC3339, until)
		if err != nil {
			return nil, fmt.Errorf("invalid until value %q: must be ISO 8601 / RFC3339 (e.g. 2026-06-01T00:00:00Z)", until)
		}
		untilTime = t
		hasUntil = true
	}

	filtered := make([]types.SessionEntry, 0, len(entries))
	for _, e := range entries {
		if e.Timestamp == "" {
			filtered = append(filtered, e)
			continue
		}
		// Try multiple timestamp formats as done in analyzer/timeline.go
		ts, err := parseEntryTimestamp(e.Timestamp)
		if err != nil {
			// Keep entries with unparseable timestamps
			filtered = append(filtered, e)
			continue
		}
		if hasSince && ts.Before(sinceTime) {
			continue
		}
		if hasUntil && !ts.Before(untilTime) {
			continue
		}
		filtered = append(filtered, e)
	}
	return filtered, nil
}

// parseEntryTimestamp parses an entry timestamp string into time.Time.
// Tries multiple formats to match the formats supported by the analyzer package.
func parseEntryTimestamp(ts string) (time.Time, error) {
	formats := []string{
		"2006-01-02T15:04:05.000Z",
		time.RFC3339Nano,
		time.RFC3339,
	}
	for _, f := range formats {
		if t, err := time.Parse(f, ts); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognized timestamp format: %s", ts)
}

// GetTechDebt implements the get_tech_debt MCP tool.
// When stats_only is set, short-circuits to an aggregate-only summary
// (analyzer.TechDebtResultStats): marker counts (bounded to the four known
// marker labels) plus a hotspot *file count* in place of the full
// HotspotFiles path list, which can grow to one entry per matched file
// across an entire scanned source tree. Applied after any source_dir merge
// so stats_only reflects the same combined result the full response would
// (DIR-042).
func (s *Service) GetTechDebt(args map[string]interface{}) (string, error) {
	entries, toolCalls, skips, err := s.loadData(args)
	if err != nil {
		return "", fmt.Errorf("failed to load session data: %w", err)
	}
	result, err := s.analyzers.TechDebt.GetTechDebt(entries, toolCalls)
	if err != nil {
		return "", fmt.Errorf("get tech debt failed: %w", err)
	}

	sourceDir := stringArg(args, "source_dir")
	var scanWarning string
	if sourceDir != "" {
		if srcResult, scanErr := s.analyzers.TechDebt.ScanSourceDir(sourceDir); scanErr == nil {
			result = analyzer.MergeTechDebtResults(result, srcResult, analyzer.DataSourceMeasured)
		} else {
			slog.Warn("ScanSourceDir failed, returning session-only result", "source_dir", sourceDir, "error", scanErr)
			scanWarning = fmt.Sprintf("source_dir scan failed for %s: %v", sourceDir, scanErr)
		}
	}

	// Set before the stats_only conversion so TechDebtResultStats carries the
	// warnings through to the aggregate response (DIR-018).
	result.Warnings = skips.Warnings()
	if scanWarning != "" {
		result.Warnings = append(result.Warnings, scanWarning)
	}

	if boolArg(args, "stats_only") {
		return marshalResult(analyzer.TechDebtResultStats(result), skips)
	}

	return marshalResult(result, skips)
}

// resolveFilePaths converts any relative paths in the slice to absolute paths
// using projectRoot as the base directory. Paths that are already absolute are
// returned unchanged. If projectRoot is empty, the slice is returned as-is.
func resolveFilePaths(files []string, projectRoot string) []string {
	if projectRoot == "" {
		return files
	}
	resolved := make([]string, len(files))
	for i, f := range files {
		if filepath.IsAbs(f) {
			resolved[i] = f
		} else {
			resolved[i] = filepath.Join(projectRoot, f)
		}
	}
	return resolved
}

// gitProjectRoot attempts to discover the git repository root via
// "git rev-parse --show-toplevel". Returns an empty string on any error so
// that callers can degrade gracefully.
func gitProjectRoot() string {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// QueryEditSequences implements the query_edit_sequences MCP tool.
func (s *Service) QueryEditSequences(args map[string]interface{}) (string, error) {
	// Extract files before loadData so we can build an empty result on no-session errors.
	var files []string
	if raw, ok := args["files"]; ok {
		switch v := raw.(type) {
		case []interface{}:
			for _, item := range v {
				if str, ok := item.(string); ok {
					files = append(files, str)
				}
			}
		case []string:
			files = v
		}
	}

	// Auto-resolve relative paths to absolute using the git project root.
	// Degrades gracefully: if git is unavailable, paths are used as-is.
	if len(files) > 0 {
		files = resolveFilePaths(files, gitProjectRoot())
	}

	includeContent := boolArg(args, "include_content")
	limitPerFile := intArg(args, "limit_per_file")

	entries, _, skips, err := s.loadData(args)
	if err != nil {
		// When no session files are found, return an empty result immediately
		// rather than propagating the error. This prevents hangs in git worktrees,
		// CI environments, and new clones that have no Claude session data.
		if strings.Contains(err.Error(), "failed to locate project sessions") {
			result := analyzer.BuildEditSequences(nil, files, includeContent, limitPerFile)
			return marshalResult(result, skips)
		}
		return "", fmt.Errorf("failed to load session data: %w", err)
	}

	result := analyzer.BuildEditSequences(entries, files, includeContent, limitPerFile)
	result.Warnings = skips.Warnings()
	return marshalResult(result, skips)
}

// AnalysisService is the interface implemented by *Service.
// It allows cmd/mcp-server to use a mock in tests.
type AnalysisService interface {
	AnalyzeBugs(args map[string]interface{}) (string, error)
	AnalyzeErrors(args map[string]interface{}) (string, error)
	QualityScan(args map[string]interface{}) (string, error)
	GetWorkPatterns(args map[string]interface{}) (string, error)
	GetTimeline(args map[string]interface{}) (string, error)
	GetTechDebt(args map[string]interface{}) (string, error)
	QueryEditSequences(args map[string]interface{}) (string, error)
}
