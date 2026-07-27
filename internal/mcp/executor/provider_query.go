package executor

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/yaleh/meta-cc/internal/locator"
	mcquery "github.com/yaleh/meta-cc/internal/mcp/query"
	"github.com/yaleh/meta-cc/internal/provider/rawfiles"
	providerrecords "github.com/yaleh/meta-cc/internal/provider/records"
)

func (e *ToolExecutor) ExecuteQueryForProvider(providerName, scope, jqFilter string, limit int, workingDir string, includeSubagents ...bool) (mcquery.QueryResult, error) {
	return e.ExecuteQueryWithTimeRangeForProvider(providerName, scope, jqFilter, limit, workingDir, mcquery.ParsedTimeRange{}, includeSubagents...)
}

func (e *ToolExecutor) ExecuteQueryWithTimeRangeForProvider(providerName, scope, jqFilter string, limit int, workingDir string, tr mcquery.ParsedTimeRange, includeSubagents ...bool) (mcquery.QueryResult, error) {
	if providerName == "" || providerName == "claude" {
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
// locator.FromSessionID for the default/claude path (a direct file lookup,
// no directory scan), or providerrecords.BuildForSession for codex/all
// (GetSession + LoadTurns for that one ID, never ListSessions).
func (e *ToolExecutor) ExecuteQueryForSession(providerName, sessionID, jqFilter string, limit int, workingDir string, tr mcquery.ParsedTimeRange) (mcquery.QueryResult, error) {
	if sessionID == "" {
		return mcquery.QueryResult{}, fmt.Errorf("session_id must not be empty")
	}

	if providerName == "" || providerName == "claude" {
		file, err := locator.NewSessionLocator().FromSessionID(sessionID)
		if err != nil {
			return mcquery.QueryResult{}, fmt.Errorf("session_id %q not found: %w", sessionID, err)
		}

		// DIR-030 cwd-boundary fix: locator.FromSessionID searches every
		// project-hash directory on disk for a matching {session_id}.jsonl
		// and returns whatever it finds, with no comparison against the
		// caller's working_dir — a cross-project leak (an unrelated
		// project's session content readable via any working_dir once its
		// session_id is known). Mirror providerrecords.BuildForSession's
		// cwd-boundary check (used by the codex/all path below) by
		// resolving the caller's working_dir to the same project-hash
		// directory name Claude Code itself uses (locator.PathToHash) and
		// requiring the resolved session file to actually live under it.
		// A workingDir of "" is a no-op here, matching
		// FilterSessionsForScope/BuildForSession's existing behavior when
		// no project scope was requested.
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
		if expectedHash := locator.PathToHash(projectPath); expectedHash != "" {
			actualHash := filepath.Base(filepath.Dir(file))
			if actualHash != expectedHash {
				return mcquery.QueryResult{}, fmt.Errorf("session %q not found for the requested provider(s) within project %q", sessionID, projectPath)
			}
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
