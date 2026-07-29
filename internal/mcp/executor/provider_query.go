package executor

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/yaleh/meta-cc/internal/config"
	"github.com/yaleh/meta-cc/internal/locator"
	mcquery "github.com/yaleh/meta-cc/internal/mcp/query"
	"github.com/yaleh/meta-cc/internal/provider/rawfiles"
	providerrecords "github.com/yaleh/meta-cc/internal/provider/records"
)

func (e *ToolExecutor) ExecuteQueryForProvider(providerName, scope, jqFilter string, limit int, workingDir string, includeSubagents ...bool) (mcquery.QueryResult, error) {
	return e.ExecuteQueryWithTimeRangeForProvider(providerName, scope, jqFilter, limit, workingDir, mcquery.ParsedTimeRange{}, includeSubagents...)
}

// resolveProviderDefault maps an empty provider name (omitted argument from
// a direct library caller) onto the process host default (DIR-073), so this
// file's routing has one source of truth instead of a hard-coded "claude".
func resolveProviderDefault(providerName string) string {
	if providerName == "" {
		return config.OmittedProviderDefault()
	}
	return providerName
}

func (e *ToolExecutor) ExecuteQueryWithTimeRangeForProvider(providerName, scope, jqFilter string, limit int, workingDir string, tr mcquery.ParsedTimeRange, includeSubagents ...bool) (mcquery.QueryResult, error) {
	providerName = resolveProviderDefault(providerName)
	if providerName == "claude" {
		return e.ExecuteQueryWithTimeRange(scope, jqFilter, limit, workingDir, tr, includeSubagents...)
	}

	projectPath := workingDir
	if projectPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return mcquery.QueryResult{}, err
		}
		projectPath = cwd
	}
	projectPath, _ = filepath.Abs(projectPath)

	registry := rawfiles.NewRegistry(projectPath)

	filters, err := rawfiles.ParseProviderFilter(providerName)
	if err != nil {
		return mcquery.QueryResult{}, err
	}
	records, warnings, err := providerrecords.Build(context.Background(), registry, filters, scope, projectPath)
	if err != nil {
		return mcquery.QueryResult{}, err
	}
	results, err := runProviderJQ(records, jqFilter, limit, tr)
	if err != nil {
		return mcquery.QueryResult{}, err
	}
	return mcquery.QueryResult{Entries: results, Warnings: warnings}, nil
}

// ExecuteQueryForSession is the DIR-030 exact-session_id fast path shared
// by every consolidated content/signal/file-activity query handler (see
// dispatchProviderQuery): unlike ExecuteQueryForProvider/
// ExecuteQueryWithTimeRangeForProvider (which list every session in scope
// and filter afterward), this reads only the one requested session —
// locator.FromSessionID for the claude-resolved path (a direct file lookup,
// no directory scan), or providerrecords.BuildForSession for codex/all
// (GetSession + LoadTurns for that one ID, never ListSessions).
func (e *ToolExecutor) ExecuteQueryForSession(providerName, sessionID, jqFilter string, limit int, workingDir string, tr mcquery.ParsedTimeRange) (mcquery.QueryResult, error) {
	if sessionID == "" {
		return mcquery.QueryResult{}, fmt.Errorf("session_id must not be empty")
	}

	providerName = resolveProviderDefault(providerName)
	if providerName == "claude" {
		// DIR-033: locator.FromSessionID (the raw, unscoped primitive) searches
		// every project-hash directory on disk for a matching
		// {session_id}.jsonl and returns whatever it finds, with no comparison
		// against the caller's working_dir — a cross-project leak (an
		// unrelated project's session content readable via any working_dir
		// once its session_id is known). FromSessionIDScoped crystallizes the
		// cwd-boundary check this bug class needs (found and separately fixed
		// here, in provider.go, and in analysis/service.go before being
		// unified into this one helper) so every caller enforces it
		// identically. A workingDir of "" is a no-op here, matching
		// FilterSessionsForScope/BuildForSession's existing behavior when no
		// project scope was requested.
		projectPath := workingDir
		if projectPath == "" {
			cwd, err := os.Getwd()
			if err != nil {
				return mcquery.QueryResult{}, err
			}
			projectPath = cwd
		}
		if abs, err := filepath.Abs(projectPath); err == nil {
			projectPath = abs
		}
		file, err := locator.NewSessionLocator().FromSessionIDScoped(sessionID, projectPath)
		if err != nil {
			return mcquery.QueryResult{}, fmt.Errorf("session %q not found for the requested provider(s) within project %q: %w", sessionID, projectPath, err)
		}

		executor := mcquery.NewQueryExecutor("")
		code, err := executor.CompileExpression(jqFilter)
		if err != nil {
			return mcquery.QueryResult{}, fmt.Errorf("invalid jq expression: %w", err)
		}
		return executor.StreamFilesWithTimeRange(context.Background(), []string{file}, code, limit, tr), nil
	}

	projectPath := workingDir
	if projectPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return mcquery.QueryResult{}, err
		}
		projectPath = cwd
	}
	projectPath, _ = filepath.Abs(projectPath)

	registry := rawfiles.NewRegistry(projectPath)
	filters, err := rawfiles.ParseProviderFilter(providerName)
	if err != nil {
		return mcquery.QueryResult{}, err
	}
	records, warnings, err := providerrecords.BuildForSession(context.Background(), registry, filters, sessionID, projectPath)
	if err != nil {
		return mcquery.QueryResult{}, err
	}
	results, err := runProviderJQ(records, jqFilter, limit, tr)
	if err != nil {
		return mcquery.QueryResult{}, err
	}
	return mcquery.QueryResult{Entries: results, Warnings: warnings}, nil
}

func (e *ToolExecutor) executeIndexedContent(providerName, scope, literal, jqFilter string, limit int, workingDir string, tr mcquery.ParsedTimeRange) (mcquery.QueryResult, bool, error) {
	providerName = resolveProviderDefault(providerName)
	if providerName == "claude" || literal == "" {
		return mcquery.QueryResult{}, false, nil
	}
	projectPath := workingDir
	if projectPath == "" {
		var err error
		projectPath, err = os.Getwd()
		if err != nil {
			return mcquery.QueryResult{}, false, err
		}
	}
	projectPath, _ = filepath.Abs(projectPath)
	filters, err := rawfiles.ParseProviderFilter(providerName)
	if err != nil {
		return mcquery.QueryResult{}, false, err
	}
	records, warnings, used := providerrecords.BuildIndexedContent(context.Background(), rawfiles.NewRegistry(projectPath), filters, scope, projectPath, literal)
	if !used {
		return mcquery.QueryResult{Warnings: warnings}, false, nil
	}
	results, err := runProviderJQ(records, jqFilter, limit, tr)
	return mcquery.QueryResult{Entries: results, Warnings: warnings}, true, err
}

// dispatchIndexedContent is the query_session_content-only opt-in. The
// handler supplies a literal that is only a candidate selector; the original
// jq and pipeline filters still run over freshly hydrated canonical records.
func (e *ToolExecutor) dispatchIndexedContent(providerName, scope, literal, jqFilter string, limit int, workingDir, sessionID string, tr mcquery.ParsedTimeRange, includeSubagents bool) (mcquery.QueryResult, error) {
	var indexWarnings []string
	if sessionID == "" {
		result, used, err := e.executeIndexedContent(providerName, scope, literal, jqFilter, limit, workingDir, tr)
		if err != nil || used {
			return result, err
		}
		indexWarnings = result.Warnings
	}
	result, err := e.dispatchProviderQuery(providerName, scope, jqFilter, limit, workingDir, sessionID, tr, includeSubagents)
	result.Warnings = append(indexWarnings, result.Warnings...)
	return result, err
}

// dispatchProviderQuery is the single opt-in point every consolidated
// query handler uses for the DIR-030 session_id fast path: a non-empty
// sessionID routes to ExecuteQueryForSession (exact thread, no listing);
// an empty sessionID preserves the pre-existing ExecuteQueryWithTimeRangeForProvider
// behavior (including scope=="session" meaning "most recent session", NOT
// an exact ID — see docs/guides/mcp-query-tools.md for that distinction).
func (e *ToolExecutor) dispatchProviderQuery(providerName, scope, jqFilter string, limit int, workingDir, sessionID string, tr mcquery.ParsedTimeRange, includeSubagents bool) (mcquery.QueryResult, error) {
	if sessionID != "" {
		return e.ExecuteQueryForSession(providerName, sessionID, jqFilter, limit, workingDir, tr)
	}
	return e.ExecuteQueryWithTimeRangeForProvider(providerName, scope, jqFilter, limit, workingDir, tr, includeSubagents)
}

func runProviderJQ(records []map[string]interface{}, jqFilter string, limit int, tr mcquery.ParsedTimeRange) ([]interface{}, error) {
	executor := mcquery.NewQueryExecutor("")
	code, err := executor.CompileExpression(jqFilter)
	if err != nil {
		return nil, fmt.Errorf("invalid jq expression: %w", err)
	}

	var out []interface{}
	for _, record := range records {
		if !inTimeRange(record["timestamp"], tr) {
			continue
		}
		iter := code.Run(record)
		for {
			value, ok := iter.Next()
			if !ok {
				break
			}
			if _, isErr := value.(error); isErr {
				continue
			}
			out = append(out, value)
			if limit > 0 && len(out) >= limit {
				return out[:limit], nil
			}
		}
	}
	return out, nil
}

func inTimeRange(raw interface{}, tr mcquery.ParsedTimeRange) bool {
	if tr.Since == nil && tr.Until == nil {
		return true
	}
	ts, ok := raw.(string)
	if !ok {
		return true
	}
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return true
	}
	if tr.Since != nil && t.Before(*tr.Since) {
		return false
	}
	if tr.Until != nil && !t.Before(*tr.Until) {
		return false
	}
	return true
}
