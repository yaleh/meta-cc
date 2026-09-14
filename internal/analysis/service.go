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
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/yaleh/meta-cc/internal/analyzer"
	"github.com/yaleh/meta-cc/internal/config"
	mcerrors "github.com/yaleh/meta-cc/internal/errors"
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
// The returned warnings slice carries one entry per session file that could
// not be parsed (DIR-018): previously such files were skipped with a bare
// `continue`, so analysis results silently excluded data. Each skipped file
// is also logged at WARN level. Callers must surface the warnings in their
// marshaled results so MCP responses never hide data exclusion.
func (s *Service) loadData(args map[string]interface{}) ([]types.SessionEntry, []types.ToolCall, []string, error) {
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
		// DIR-094: the provider path already accumulates per-session and
		// per-file diagnostics (providerrecords.Build / BuildForSession), but
		// this call site dropped them on the floor — so a codex/all analysis
		// silently excluded sessions while the claude path warned about the
		// same class of problem. Thread them through instead.
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

	var allEntries []types.SessionEntry
	var warnings []string
	for _, f := range files {
		p := parser.NewSessionParser(f)
		entries, err := p.ParseEntries()
		// DIR-018 (skip-and-report) extended by DIR-094 to the zero-entry
		// case: a file that parses cleanly but yields no message entries —
		// literally empty, or a metadata-only session stub — is just as
		// excluded from the result as an unparseable one, so it must be
		// warned about too. Only warning on the parse error left the
		// 8eda8f4e-style empty file invisible here. locator.ExclusionFor is
		// the shared rule the claude listing path applies to the same file,
		// so both paths reach one verdict.
		if exclusion := locator.ExclusionFor(f, len(entries), err); exclusion != nil {
			slog.Warn("skipping session file that contributed no data", "file", f, "error", exclusion.Err)
			warnings = append(warnings, exclusion.Warning())
			continue
		}
		allEntries = append(allEntries, entries...)
	}

	toolCalls := types.ExtractToolCalls(allEntries)
	return allEntries, toolCalls, warnings, nil
}

// loadProviderData is loadData's non-claude branch: it reads the corpus
// through the provider abstraction (providerrecords.Build / BuildForSession)
// rather than the raw claude JSONL parser. It returns warnings with the same
// shape and meaning as loadData's own, so callers surface them identically
// regardless of which provider answered (DIR-094).
func (s *Service) loadProviderData(scope, workingDir, providerName, sessionID string) ([]types.SessionEntry, []types.ToolCall, []string, error) {
	projectPath, err := filepath.Abs(workingDir)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to resolve project path: %w", err)
	}
	filters, err := rawfiles.ParseProviderFilter(providerName)
	if err != nil {
		return nil, nil, nil, err
	}
	registry := rawfiles.NewRegistry(projectPath)

	var (
		records  []map[string]interface{}
		warnings []string
	)
	if sessionID != "" {
		// DIR-030 exact-session fast path: GetSession/LoadTurns for this
		// one ID only, never ListSessions across the whole project.
		records, warnings, err = providerrecords.BuildForSession(context.Background(), registry, filters, sessionID, projectPath)
	} else {
		records, warnings, err = providerrecords.Build(context.Background(), registry, filters, scope, projectPath)
	}
	if err != nil {
		return nil, nil, warnings, err
	}
	// The providers' own listing diagnostics (per-file skips on the claude
	// path, backend degradation on the codex one) are a separate channel from
	// providerrecords' per-session warnings; fold both in.
	for _, p := range registry.Providers(filters) {
		if wp, ok := p.(interface{ Warnings() []string }); ok {
			warnings = append(warnings, wp.Warnings()...)
		}
	}
	entries, err := entriesFromRecords(records)
	if err != nil {
		return nil, nil, warnings, err
	}
	return entries, types.ExtractToolCalls(entries), warnings, nil
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

// intArgDefault returns args[key] when the caller supplied it as a number, and
// def otherwise. Unlike intArg it can tell "absent" from an explicit 0, which
// is what max_patterns/limit need: 0 is their documented "unlimited" sentinel,
// so it must never be silently reinterpreted as "use the default".
func intArgDefault(args map[string]interface{}, key string, def int) int {
	if v, ok := args[key]; ok && v != nil {
		if f, ok := v.(float64); ok {
			return int(f)
		}
	}
	return def
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

// Bug-analysis output shaping (DIR-096).
//
// analyze_bugs used to emit every pattern it found, each carrying up to
// `limit` raw (untruncated) error strings. On this project's 5-day corpus a
// no-argument call returned 56 patterns / 70,248 bytes -- past the 32KB inline
// threshold, so it always spilled to file_ref mode -- and each example was a
// bare string with no session or timestamp to attribute it to (a consumer
// parsing them as objects hit AttributeError).
//
// The shaping below bounds the response structurally rather than by hoping the
// corpus stays small: at most defaultBugMaxPatterns patterns, at most
// defaultBugExamplesPerPattern examples per pattern, each free-text field
// clipped to bugExampleTextLimit bytes, and the examples section as a whole
// clipped to bugExamplesByteBudget.
const (
	// defaultBugMaxPatterns caps the pattern list when max_patterns is absent.
	// An explicit max_patterns of 0 means unlimited.
	defaultBugMaxPatterns = 20

	// defaultBugExamplesPerPattern caps examples per pattern when `limit` is
	// absent. An explicit `limit` wins, and 0 there still means unlimited.
	defaultBugExamplesPerPattern = 3

	// bugExampleTextLimit clips error_text and fix_text. The longest error
	// message observed on the live corpus was 10,148 characters, so one
	// uncapped example could outweigh every other pattern combined.
	bugExampleTextLimit = 500

	// bugExamplesByteBudget bounds the examples section as a whole, so the
	// response stays inline even for a corpus whose per-field texts all sit
	// just under bugExampleTextLimit: the examples can never grow past this
	// budget, which leaves ample headroom under the 32KB inline threshold for
	// the per-pattern metadata.
	bugExamplesByteBudget = 16 * 1024
)

// bugExample is one observed occurrence of a bug pattern, attributed to the
// session and turn it happened in. Before DIR-096 an example was the bare error
// string, which carried no way to find the occurrence again.
type bugExample struct {
	SessionID string `json:"session_id"`
	Timestamp string `json:"timestamp"`
	ErrorText string `json:"error_text"`
	// FixText is the paired same-tool success output (the heuristic fix
	// described by ADR-007), omitted when no pair was found.
	FixText string `json:"fix_text,omitempty"`
	// Signature is the example's own error signature, so an example stays
	// interpretable when detached from its pattern.
	Signature string `json:"signature"`
}

// bugPatternView is analyzer.BugPattern with structured, bounded examples.
type bugPatternView struct {
	ErrorSignature string       `json:"error_signature"`
	FixCount       int          `json:"fix_count"`
	Recurrences    int          `json:"recurrences"`
	UnfixedErrors  int          `json:"unfixed_errors"`
	Examples       []bugExample `json:"examples"`
}

// bugAnalysisView is analyze_bugs' response shape.
type bugAnalysisView struct {
	Patterns      []bugPatternView `json:"patterns"`
	TotalPairs    int              `json:"total_pairs"`
	TotalErrors   int              `json:"total_errors"`
	UnfixedErrors int              `json:"unfixed_errors"`
	// TotalPatterns counts every pattern found, before max_patterns clipped the
	// list; TruncatedPatterns reports how many were dropped.
	TotalPatterns     int                 `json:"total_patterns"`
	TruncatedPatterns int                 `json:"truncated_patterns,omitempty"`
	DataSource        analyzer.DataSource `json:"data_source"`
	EstimatedFields   []string            `json:"estimated_fields,omitempty"`
	// Warnings names any session files skipped during load (DIR-018).
	Warnings []string `json:"warnings,omitempty"`
}

// AnalyzeBugs implements the analyze_bugs MCP tool.
// When stats_only is set, short-circuits to an aggregate-only pattern-count
// summary (analyzer.AnalyzeBugsStats) with no per-pattern Examples text,
// mirroring GetTimeline's own stats_only short-circuit (DIR-042) -- no example
// collection runs on that path at all (DIR-096).
func (s *Service) AnalyzeBugs(args map[string]interface{}) (string, error) {
	entries, toolCalls, warnings, err := s.loadData(args)
	if err != nil {
		return "", fmt.Errorf("failed to load session data: %w", err)
	}
	entries, toolCalls, err = applyTimeWindow(entries, toolCalls, args)
	if err != nil {
		return "", err
	}
	if boolArg(args, "stats_only") {
		stats, err := analyzer.AnalyzeBugsStats(entries, toolCalls)
		if err != nil {
			return "", fmt.Errorf("analyze bugs failed: %w", err)
		}
		stats.Warnings = warnings
		return marshalResult(stats)
	}
	// The effective per-pattern example limit is the response's, so the
	// injected analyzer is asked for exactly what will be emitted.
	perPattern := intArgDefault(args, "limit", defaultBugExamplesPerPattern)
	result, err := s.analyzers.BugAnalyzer.AnalyzeBugs(entries, toolCalls, perPattern)
	if err != nil {
		return "", fmt.Errorf("analyze bugs failed: %w", err)
	}
	result.Warnings = warnings
	return marshalResult(shapeBugAnalysis(result, entries, toolCalls, perPattern, intArgDefault(args, "max_patterns", defaultBugMaxPatterns)))
}

// shapeBugAnalysis converts the analyzer's raw result into the bounded,
// structured response documented on the shaping constants above. maxPatterns
// and perPattern are already resolved (0 = unlimited for either).
func shapeBugAnalysis(result *analyzer.BugAnalysisResult, entries []types.SessionEntry, toolCalls []types.ToolCall, perPattern, maxPatterns int) bugAnalysisView {
	// Copy before sorting: the analyzer's own slice is sorted by recurrences
	// only, so equal-recurrence patterns keep whatever order the map range that
	// built them produced. Ranking by (recurrences, fix_count, signature) makes
	// the cap below select the same patterns run-to-run.
	patterns := make([]analyzer.BugPattern, len(result.Patterns))
	copy(patterns, result.Patterns)
	sort.SliceStable(patterns, func(i, j int) bool {
		if patterns[i].Recurrences != patterns[j].Recurrences {
			return patterns[i].Recurrences > patterns[j].Recurrences
		}
		if patterns[i].FixCount != patterns[j].FixCount {
			return patterns[i].FixCount > patterns[j].FixCount
		}
		return patterns[i].ErrorSignature < patterns[j].ErrorSignature
	})

	kept := patterns
	if maxPatterns > 0 && len(kept) > maxPatterns {
		kept = kept[:maxPatterns]
	}

	examples := collectBugExamples(entries, toolCalls, perPattern)

	view := bugAnalysisView{
		Patterns:        make([]bugPatternView, 0, len(kept)),
		TotalPairs:      result.TotalPairs,
		TotalErrors:     result.TotalErrors,
		UnfixedErrors:   result.UnfixedErrors,
		TotalPatterns:   len(patterns),
		DataSource:      result.DataSource,
		EstimatedFields: result.EstimatedFields,
		Warnings:        result.Warnings,
	}
	if dropped := len(patterns) - len(kept); dropped > 0 {
		view.TruncatedPatterns = dropped
	}

	budgetUsed := 0
	for _, p := range kept {
		pv := bugPatternView{
			ErrorSignature: p.ErrorSignature,
			FixCount:       p.FixCount,
			Recurrences:    p.Recurrences,
			UnfixedErrors:  p.UnfixedErrors,
			Examples:       []bugExample{},
		}
		for _, ex := range examples[p.ErrorSignature] {
			size := bugExampleSize(ex)
			if budgetUsed+size > bugExamplesByteBudget {
				break
			}
			budgetUsed += size
			pv.Examples = append(pv.Examples, ex)
		}
		view.Patterns = append(view.Patterns, pv)
	}
	return view
}

// collectBugExamples walks toolCalls the way analyzer.AnalyzeBugs walks them
// (same signature rule, same 3-position error->success pairing, same
// consumed-success bookkeeping) and returns up to perPattern attributable
// examples per error signature. perPattern <= 0 means unlimited.
//
// The walk is repeated here rather than threaded out of the analyzer because
// analyzer.BugPattern.Examples is []string: it carries no session, timestamp or
// fix text, so attribution has to come from entries/toolCalls, which this
// package holds. Counts (fix_count, recurrences, unfixed_errors) still come
// from the analyzer -- this pass only attributes occurrences for the patterns
// the analyzer already ranked, and must not be read as a second source of
// truth for them.
func collectBugExamples(entries []types.SessionEntry, toolCalls []types.ToolCall, perPattern int) map[string][]bugExample {
	sessions := toolCallSessions(entries, toolCalls)

	out := make(map[string][]bugExample)
	consumed := make(map[int]bool)
	for i := 0; i < len(toolCalls); i++ {
		tc := toolCalls[i]
		if tc.Status != "error" {
			continue
		}
		sig := analyzer.CalculateErrorSignature(tc.ToolName, tc.Error)
		ex := bugExample{
			SessionID: sessions[i],
			Timestamp: tc.Timestamp,
			ErrorText: clipBugText(tc.Error),
			Signature: sig,
		}

		// Claim the paired success even when this occurrence will not be
		// recorded: the bookkeeping has to match the analyzer's, or a later
		// occurrence would claim a fix the analyzer credited to this one.
		for j := i + 1; j <= i+3 && j < len(toolCalls); j++ {
			candidate := toolCalls[j]
			if candidate.ToolName == tc.ToolName && candidate.Status == "success" && !consumed[j] {
				consumed[j] = true
				ex.FixText = clipBugText(candidate.Output)
				break
			}
		}

		if perPattern <= 0 || len(out[sig]) < perPattern {
			out[sig] = append(out[sig], ex)
		}
	}
	return out
}

// toolCallSessions returns, for tool call i, the session id of the entry that
// produced it. types.ExtractToolCalls emits one ToolCall per tool_use block in
// entry order, first-occurrence-wins on duplicate tool_use ids (DIR-058), so
// mirroring that walk attributes every call positionally.
//
// Positional rather than keyed on ToolCall.UUID on purpose: the provider
// normalization path (internal/provider/records) emits records carrying
// sessionId but no uuid, and those entries reach AnalyzeBugs through exactly
// the same loadData/ExtractToolCalls pipeline -- a UUID-keyed lookup would
// silently attribute nothing for every codex/all corpus.
func toolCallSessions(entries []types.SessionEntry, toolCalls []types.ToolCall) []string {
	sessions := make([]string, 0, len(toolCalls))
	emitted := make(map[string]struct{})
	for i := range entries {
		entry := entries[i]
		if entry.Message == nil {
			continue
		}
		for _, block := range entry.Message.Content {
			if block.Type != "tool_use" || block.ToolUse == nil {
				continue
			}
			if _, dup := emitted[block.ToolUse.ID]; dup {
				continue
			}
			emitted[block.ToolUse.ID] = struct{}{}
			sessions = append(sessions, entry.SessionID)
		}
	}
	if len(sessions) != len(toolCalls) {
		// The mirror walk and types.ExtractToolCalls disagree (it changed, or
		// the caller supplied a toolCalls slice built elsewhere). Attribute
		// nothing rather than attribute wrongly.
		return make([]string, len(toolCalls))
	}
	return sessions
}

// bugExampleSize estimates one example's serialized footprint, for the
// examples-section budget.
func bugExampleSize(ex bugExample) int {
	// 80 approximates the surrounding JSON scaffolding (keys, quotes, commas).
	return len(ex.SessionID) + len(ex.Timestamp) + len(ex.ErrorText) + len(ex.FixText) + len(ex.Signature) + 80
}

// clipBugText bounds one free-text example field, cutting on a rune boundary
// and marking the cut so a clipped field is never mistaken for the whole
// message.
func clipBugText(s string) string {
	if len(s) <= bugExampleTextLimit {
		return s
	}
	cut := bugExampleTextLimit
	for cut > 0 && !utf8.RuneStart(s[cut]) {
		cut--
	}
	return s[:cut] + "…"
}

// AnalyzeErrors implements the analyze_errors MCP tool.
// When stats_only is set, short-circuits to an aggregate-only per-tool/
// per-type count summary (analyzer.AnalyzeErrorsStats) with no examples
// text, mirroring GetTimeline's own stats_only short-circuit (DIR-042).
func (s *Service) AnalyzeErrors(args map[string]interface{}) (string, error) {
	entries, toolCalls, warnings, err := s.loadData(args)
	if err != nil {
		return "", fmt.Errorf("failed to load session data: %w", err)
	}
	entries, toolCalls, err = applyTimeWindow(entries, toolCalls, args)
	if err != nil {
		return "", err
	}
	if boolArg(args, "stats_only") {
		stats, err := analyzer.AnalyzeErrorsStats(entries, toolCalls)
		if err != nil {
			return "", fmt.Errorf("failed to analyze errors: %w", err)
		}
		stats.Warnings = warnings
		return marshalResult(stats)
	}
	result, err := s.analyzers.ErrorAnalyzer.AnalyzeErrors(entries, toolCalls, intArg(args, "limit"))
	if err != nil {
		return "", fmt.Errorf("failed to analyze errors: %w", err)
	}
	result.Warnings = warnings
	return marshalResult(result)
}

// QualityScan implements the quality_scan MCP tool.
// QualityScan's result is already aggregate-only (four scored dimensions,
// no per-item example text); the stats_only short-circuit
// (analyzer.QualityScanStatsOnly) exists so this method still honors the
// documented stats_only contract explicitly rather than silently ignoring it
// (DIR-042).
func (s *Service) QualityScan(args map[string]interface{}) (string, error) {
	entries, toolCalls, warnings, err := s.loadData(args)
	if err != nil {
		return "", fmt.Errorf("failed to load session data: %w", err)
	}
	entries, toolCalls, err = applyTimeWindow(entries, toolCalls, args)
	if err != nil {
		return "", err
	}
	if boolArg(args, "stats_only") {
		stats, err := analyzer.QualityScanStatsOnly(entries, toolCalls)
		if err != nil {
			return "", fmt.Errorf("quality scan failed: %w", err)
		}
		stats.Warnings = warnings
		return marshalResult(stats)
	}
	result, err := s.analyzers.QualityScanner.QualityScan(entries, toolCalls)
	if err != nil {
		return "", fmt.Errorf("quality scan failed: %w", err)
	}
	result.Warnings = warnings
	return marshalResult(result)
}

// GetWorkPatterns implements the get_work_patterns MCP tool.
// GetWorkPatterns's result is already aggregate-only (tool counts, a fixed
// 24-slot hourly histogram, and two scalar counters, no per-item example
// text); the stats_only short-circuit (analyzer.GetWorkPatternsStatsOnly)
// exists so this method still honors the documented stats_only contract
// explicitly rather than silently ignoring it (DIR-042).
func (s *Service) GetWorkPatterns(args map[string]interface{}) (string, error) {
	entries, toolCalls, warnings, err := s.loadData(args)
	if err != nil {
		return "", fmt.Errorf("failed to load session data: %w", err)
	}
	entries, toolCalls, err = applyTimeWindow(entries, toolCalls, args)
	if err != nil {
		return "", err
	}
	if boolArg(args, "stats_only") {
		stats, err := analyzer.GetWorkPatternsStatsOnly(entries, toolCalls)
		if err != nil {
			return "", fmt.Errorf("get work patterns failed: %w", err)
		}
		stats.Warnings = warnings
		return marshalResult(stats)
	}
	result, err := s.analyzers.WorkPatterns.GetWorkPatterns(entries, toolCalls)
	if err != nil {
		return "", fmt.Errorf("get work patterns failed: %w", err)
	}
	result.Warnings = warnings
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
	entries, _, warnings, err := s.loadData(args)
	if err != nil {
		return "", fmt.Errorf("failed to load session data: %w", err)
	}

	since := stringArg(args, "since")
	until := stringArg(args, "until")

	// Apply since/until time-clipping if provided. GetTimeline ignored the
	// toolCalls loadData returned even before DIR-095, so the re-derived slice
	// applyTimeWindow hands back is discarded here as it always was.
	entries, _, err = applyTimeWindow(entries, nil, args)
	if err != nil {
		return "", err
	}

	if boolArg(args, "stats_only") {
		stats := analyzer.GetTimelineStats(entries)
		stats.Warnings = warnings
		return marshalResult(stats)
	}

	// When no time clipping and entry count exceeds threshold, default to stats mode
	// to prevent the 737K+ character context truncation observed in large projects.
	if since == "" && until == "" && len(entries) > timelineAutoStatsThreshold {
		stats := analyzer.GetTimelineStats(entries)
		stats.Warnings = warnings
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
	result.Warnings = warnings
	return marshalResult(result)
}

// applyTimeWindow narrows a freshly loaded corpus to the optional RFC3339
// since/until window BEFORE any aggregation runs, so every downstream output
// -- the full result and the aggregate-only stats_only result alike --
// describes the window rather than the whole corpus (DIR-095). Semantics match
// the pre-existing since/until on get_timeline, query_session_content, and
// query_session_signals: since is inclusive, until is exclusive.
//
// A call with neither parameter set returns its input untouched, so the
// unwindowed path is exactly the pre-DIR-095 behavior.
func applyTimeWindow(entries []types.SessionEntry, toolCalls []types.ToolCall, args map[string]interface{}) ([]types.SessionEntry, []types.ToolCall, error) {
	since := stringArg(args, "since")
	until := stringArg(args, "until")
	if since == "" && until == "" {
		return entries, toolCalls, nil
	}

	filtered, err := filterEntriesByTimeRange(entries, since, until)
	if err != nil {
		return nil, nil, err
	}
	// toolCalls are re-derived from the filtered entries rather than filtered
	// independently. A ToolCall's Timestamp is its own tool_use entry's, and
	// ExtractToolCalls pairs tool_use with tool_result by ID across the slice
	// it is given, so re-deriving is precisely "extract from a corpus that
	// contained only the in-window entries". That is what makes a windowed run
	// over the full corpus equal an unwindowed run over the in-window subset,
	// and it cannot leave an out-of-window tool call behind.
	//
	// The corollary is deliberate: a tool_use inside the window whose
	// tool_result falls outside it comes back with no observed output/status,
	// exactly as it would from the subset corpus -- the completion evidence
	// lies outside the window being asked about.
	return filtered, types.ExtractToolCalls(filtered), nil
}

// filterEntriesByTimeRange filters session entries to those within the since/until range.
// Both since and until are optional ISO 8601 strings. Since is inclusive, until is exclusive.
// An unparseable bound is reported as mcerrors.ErrInvalidInput (DIR-095) so callers
// can distinguish a bad parameter from a failure to read the corpus.
func filterEntriesByTimeRange(entries []types.SessionEntry, since, until string) ([]types.SessionEntry, error) {
	var sinceTime, untilTime time.Time
	var hasSince, hasUntil bool

	if since != "" {
		t, err := time.Parse(time.RFC3339, since)
		if err != nil {
			return nil, fmt.Errorf("invalid since value %q: must be ISO 8601 / RFC3339 (e.g. 2026-01-01T00:00:00Z): %w", since, mcerrors.ErrInvalidInput)
		}
		sinceTime = t
		hasSince = true
	}
	if until != "" {
		t, err := time.Parse(time.RFC3339, until)
		if err != nil {
			return nil, fmt.Errorf("invalid until value %q: must be ISO 8601 / RFC3339 (e.g. 2026-06-01T00:00:00Z): %w", until, mcerrors.ErrInvalidInput)
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
	entries, toolCalls, warnings, err := s.loadData(args)
	if err != nil {
		return "", fmt.Errorf("failed to load session data: %w", err)
	}
	entries, toolCalls, err = applyTimeWindow(entries, toolCalls, args)
	if err != nil {
		return "", err
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
	result.Warnings = warnings
	if scanWarning != "" {
		result.Warnings = append(result.Warnings, scanWarning)
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

	entries, _, warnings, err := s.loadData(args)
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
	result.Warnings = warnings
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
