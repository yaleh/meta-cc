// Package analysis provides a service facade that encapsulates the full
// pipeline of: locate session files → parse → run analyzer functions.
// cmd/mcp-server uses this package instead of importing internal/parser
// and internal/analyzer directly.
package analysis

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/yaleh/meta-cc/internal/analyzer"
	"github.com/yaleh/meta-cc/internal/locator"
	"github.com/yaleh/meta-cc/internal/parser"
	"github.com/yaleh/meta-cc/internal/provider/rawfiles"
	providerrecords "github.com/yaleh/meta-cc/internal/provider/records"
	"github.com/yaleh/meta-cc/internal/types"
)

// Analyzers holds the injected analyzer interfaces used by Service.
// Zero-value fields are replaced with DefaultAnalyzer instances at construction time.
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
func (s *Service) loadData(args map[string]interface{}) ([]types.SessionEntry, []types.ToolCall, error) {
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
	// DIR-030: session_id, when set, is an exact-thread selector distinct
	// from scope="session" ("most recent session") — it takes precedence
	// over scope entirely and reads only the one requested session.
	sessionID := stringArg(args, "session_id")

	if providerName != "" && providerName != "claude" {
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
			return nil, nil, err
		}
		files = []string{sessionFile}
	case scope == "session":
		sessionFile, err := loc.FromProjectPath(workingDir)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to locate session: %w", err)
		}
		files = []string{sessionFile}
	default:
		var err error
		files, err = loc.AllSessionsFromProject(workingDir)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to locate project sessions: %w", err)
		}
	}

	var allEntries []types.SessionEntry
	for _, f := range files {
		p := parser.NewSessionParser(f)
		entries, err := p.ParseEntries()
		if err != nil {
			continue // skip malformed files
		}
		allEntries = append(allEntries, entries...)
	}

	toolCalls := types.ExtractToolCalls(allEntries)
	return allEntries, toolCalls, nil
}

func (s *Service) loadProviderData(scope, workingDir, providerName, sessionID string) ([]types.SessionEntry, []types.ToolCall, error) {
	projectPath, err := filepath.Abs(workingDir)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to resolve project path: %w", err)
	}
	filters, err := rawfiles.ParseProviderFilter(providerName)
	if err != nil {
		return nil, nil, err
	}
	registry := rawfiles.NewRegistry(projectPath)

	var records []map[string]interface{}
	if sessionID != "" {
		// DIR-030 exact-session fast path: GetSession/LoadTurns for this
		// one ID only, never ListSessions across the whole project.
		records, _, err = providerrecords.BuildForSession(context.Background(), registry, filters, sessionID, projectPath)
	} else {
		records, _, err = providerrecords.Build(context.Background(), registry, filters, scope, projectPath)
	}
	if err != nil {
		return nil, nil, err
	}
	entries, err := entriesFromRecords(records)
	if err != nil {
		return nil, nil, err
	}
	return entries, types.ExtractToolCalls(entries), nil
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

func marshalResult(v interface{}) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}
	return string(data), nil
}

// AnalyzeBugs implements the analyze_bugs MCP tool.
// When stats_only is set, short-circuits to an aggregate-only pattern-count
// summary (analyzer.AnalyzeBugsStats) with no per-pattern Examples text,
// mirroring GetTimeline's own stats_only short-circuit (DIR-042).
func (s *Service) AnalyzeBugs(args map[string]interface{}) (string, error) {
	entries, toolCalls, err := s.loadData(args)
	if err != nil {
		return "", fmt.Errorf("failed to load session data: %w", err)
	}
	if boolArg(args, "stats_only") {
		stats, err := analyzer.AnalyzeBugsStats(entries, toolCalls)
		if err != nil {
			return "", fmt.Errorf("analyze bugs failed: %w", err)
		}
		return marshalResult(stats)
	}
	result, err := s.analyzers.BugAnalyzer.AnalyzeBugs(entries, toolCalls, intArg(args, "limit"))
	if err != nil {
		return "", fmt.Errorf("analyze bugs failed: %w", err)
	}
	return marshalResult(result)
}

// AnalyzeErrors implements the analyze_errors MCP tool.
// When stats_only is set, short-circuits to an aggregate-only per-tool/
// per-type count summary (analyzer.AnalyzeErrorsStats) with no examples
// text, mirroring GetTimeline's own stats_only short-circuit (DIR-042).
func (s *Service) AnalyzeErrors(args map[string]interface{}) (string, error) {
	entries, toolCalls, err := s.loadData(args)
	if err != nil {
		return "", fmt.Errorf("failed to load session data: %w", err)
	}
	if boolArg(args, "stats_only") {
		stats, err := analyzer.AnalyzeErrorsStats(entries, toolCalls)
		if err != nil {
			return "", fmt.Errorf("failed to analyze errors: %w", err)
		}
		return marshalResult(stats)
	}
	result, err := s.analyzers.ErrorAnalyzer.AnalyzeErrors(entries, toolCalls, intArg(args, "limit"))
	if err != nil {
		return "", fmt.Errorf("failed to analyze errors: %w", err)
	}
	return marshalResult(result)
}

// QualityScan implements the quality_scan MCP tool.
// QualityScan's result is already aggregate-only (four scored dimensions,
// no per-item example text); the stats_only short-circuit
// (analyzer.QualityScanStatsOnly) exists so this method still honors the
// documented stats_only contract explicitly rather than silently ignoring it
// (DIR-042).
func (s *Service) QualityScan(args map[string]interface{}) (string, error) {
	entries, toolCalls, err := s.loadData(args)
	if err != nil {
		return "", fmt.Errorf("failed to load session data: %w", err)
	}
	if boolArg(args, "stats_only") {
		stats, err := analyzer.QualityScanStatsOnly(entries, toolCalls)
		if err != nil {
			return "", fmt.Errorf("quality scan failed: %w", err)
		}
		return marshalResult(stats)
	}
	result, err := s.analyzers.QualityScanner.QualityScan(entries, toolCalls)
	if err != nil {
		return "", fmt.Errorf("quality scan failed: %w", err)
	}
	return marshalResult(result)
}

// GetWorkPatterns implements the get_work_patterns MCP tool.
// GetWorkPatterns's result is already aggregate-only (tool counts, a fixed
// 24-slot hourly histogram, and two scalar counters, no per-item example
// text); the stats_only short-circuit (analyzer.GetWorkPatternsStatsOnly)
// exists so this method still honors the documented stats_only contract
// explicitly rather than silently ignoring it (DIR-042).
func (s *Service) GetWorkPatterns(args map[string]interface{}) (string, error) {
	entries, toolCalls, err := s.loadData(args)
	if err != nil {
		return "", fmt.Errorf("failed to load session data: %w", err)
	}
	if boolArg(args, "stats_only") {
		stats, err := analyzer.GetWorkPatternsStatsOnly(entries, toolCalls)
		if err != nil {
			return "", fmt.Errorf("get work patterns failed: %w", err)
		}
		return marshalResult(stats)
	}
	result, err := s.analyzers.WorkPatterns.GetWorkPatterns(entries, toolCalls)
	if err != nil {
		return "", fmt.Errorf("get work patterns failed: %w", err)
	}
	return marshalResult(result)
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
	entries, _, err := s.loadData(args)
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
		return marshalResult(stats)
	}

	// When no time clipping and entry count exceeds threshold, default to stats mode
	// to prevent the 737K+ character context truncation observed in large projects.
	if since == "" && until == "" && len(entries) > timelineAutoStatsThreshold {
		stats := analyzer.GetTimelineStats(entries)
		return marshalResult(stats)
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
	return marshalResult(result)
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
	entries, toolCalls, err := s.loadData(args)
	if err != nil {
		return "", fmt.Errorf("failed to load session data: %w", err)
	}
	result, err := s.analyzers.TechDebt.GetTechDebt(entries, toolCalls)
	if err != nil {
		return "", fmt.Errorf("get tech debt failed: %w", err)
	}

	sourceDir := stringArg(args, "source_dir")
	if sourceDir != "" {
		if srcResult, scanErr := s.analyzers.TechDebt.ScanSourceDir(sourceDir); scanErr == nil {
			result = analyzer.MergeTechDebtResults(result, srcResult, analyzer.DataSourceMeasured)
		}
		// Degrade gracefully on scan error: keep the session-only result.
	}

	if boolArg(args, "stats_only") {
		return marshalResult(analyzer.TechDebtResultStats(result))
	}

	return marshalResult(result)
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

	entries, _, err := s.loadData(args)
	if err != nil {
		// When no session files are found, return an empty result immediately
		// rather than propagating the error. This prevents hangs in git worktrees,
		// CI environments, and new clones that have no Claude session data.
		if strings.Contains(err.Error(), "failed to locate project sessions") {
			result := analyzer.BuildEditSequences(nil, files, includeContent, limitPerFile)
			return marshalResult(result)
		}
		return "", fmt.Errorf("failed to load session data: %w", err)
	}

	result := analyzer.BuildEditSequences(entries, files, includeContent, limitPerFile)
	return marshalResult(result)
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
