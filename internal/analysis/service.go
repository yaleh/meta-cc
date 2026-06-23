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
	"path/filepath"

	"github.com/yaleh/meta-cc/internal/analyzer"
	"github.com/yaleh/meta-cc/internal/conversation"
	"github.com/yaleh/meta-cc/internal/locator"
	"github.com/yaleh/meta-cc/internal/parser"
	providerpkg "github.com/yaleh/meta-cc/internal/provider"
	claudeprovider "github.com/yaleh/meta-cc/internal/provider/claude"
	codexprovider "github.com/yaleh/meta-cc/internal/provider/codex"
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
	if providerName != "" && providerName != "claude" {
		return s.loadProviderData(scope, workingDir, providerName)
	}

	loc := locator.NewSessionLocator()
	var files []string
	if scope == "session" {
		sessionFile, err := loc.FromProjectPath(workingDir)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to locate session: %w", err)
		}
		files = []string{sessionFile}
	} else {
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

func (s *Service) loadProviderData(scope, workingDir, providerName string) ([]types.SessionEntry, []types.ToolCall, error) {
	projectPath, err := filepath.Abs(workingDir)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to resolve project path: %w", err)
	}
	filters, err := providerFilter(providerName)
	if err != nil {
		return nil, nil, err
	}
	registry := providerpkg.NewRegistry(
		claudeprovider.NewProvider(locator.NewSessionLocator(), projectPath),
		codexprovider.NewProvider(locator.NewCodexLocator()),
	)
	records, _, err := providerrecords.Build(context.Background(), registry, filters, scope, projectPath)
	if err != nil {
		return nil, nil, err
	}
	entries, err := entriesFromRecords(records)
	if err != nil {
		return nil, nil, err
	}
	return entries, types.ExtractToolCalls(entries), nil
}

func providerFilter(providerName string) ([]conversation.ProviderID, error) {
	switch providerName {
	case "codex":
		return []conversation.ProviderID{conversation.ProviderCodex}, nil
	case "all":
		return []conversation.ProviderID{conversation.ProviderClaude, conversation.ProviderCodex}, nil
	default:
		return nil, fmt.Errorf("invalid provider %q: must be \"claude\", \"codex\", or \"all\"", providerName)
	}
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
func (s *Service) AnalyzeBugs(args map[string]interface{}) (string, error) {
	entries, toolCalls, err := s.loadData(args)
	if err != nil {
		return "", fmt.Errorf("failed to load session data: %w", err)
	}
	result, err := s.analyzers.BugAnalyzer.AnalyzeBugs(entries, toolCalls, intArg(args, "limit"))
	if err != nil {
		return "", fmt.Errorf("analyze bugs failed: %w", err)
	}
	return marshalResult(result)
}

// AnalyzeErrors implements the analyze_errors MCP tool.
func (s *Service) AnalyzeErrors(args map[string]interface{}) (string, error) {
	entries, toolCalls, err := s.loadData(args)
	if err != nil {
		return "", fmt.Errorf("failed to load session data: %w", err)
	}
	result, err := s.analyzers.ErrorAnalyzer.AnalyzeErrors(entries, toolCalls, intArg(args, "limit"))
	if err != nil {
		return "", fmt.Errorf("failed to analyze errors: %w", err)
	}
	return marshalResult(result)
}

// QualityScan implements the quality_scan MCP tool.
func (s *Service) QualityScan(args map[string]interface{}) (string, error) {
	entries, toolCalls, err := s.loadData(args)
	if err != nil {
		return "", fmt.Errorf("failed to load session data: %w", err)
	}
	result, err := s.analyzers.QualityScanner.QualityScan(entries, toolCalls)
	if err != nil {
		return "", fmt.Errorf("quality scan failed: %w", err)
	}
	return marshalResult(result)
}

// GetWorkPatterns implements the get_work_patterns MCP tool.
func (s *Service) GetWorkPatterns(args map[string]interface{}) (string, error) {
	entries, toolCalls, err := s.loadData(args)
	if err != nil {
		return "", fmt.Errorf("failed to load session data: %w", err)
	}
	result, err := s.analyzers.WorkPatterns.GetWorkPatterns(entries, toolCalls)
	if err != nil {
		return "", fmt.Errorf("get work patterns failed: %w", err)
	}
	return marshalResult(result)
}

// GetTimeline implements the get_timeline MCP tool.
func (s *Service) GetTimeline(args map[string]interface{}) (string, error) {
	entries, _, err := s.loadData(args)
	if err != nil {
		return "", fmt.Errorf("failed to load session data: %w", err)
	}

	if boolArg(args, "stats_only") {
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

// GetTechDebt implements the get_tech_debt MCP tool.
func (s *Service) GetTechDebt(args map[string]interface{}) (string, error) {
	entries, toolCalls, err := s.loadData(args)
	if err != nil {
		return "", fmt.Errorf("failed to load session data: %w", err)
	}
	result, err := s.analyzers.TechDebt.GetTechDebt(entries, toolCalls)
	if err != nil {
		return "", fmt.Errorf("get tech debt failed: %w", err)
	}
	return marshalResult(result)
}

// QueryEditSequences implements the query_edit_sequences MCP tool.
func (s *Service) QueryEditSequences(args map[string]interface{}) (string, error) {
	entries, _, err := s.loadData(args)
	if err != nil {
		return "", fmt.Errorf("failed to load session data: %w", err)
	}

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

	includeContent := boolArg(args, "include_content")
	limitPerFile := intArg(args, "limit_per_file")

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
