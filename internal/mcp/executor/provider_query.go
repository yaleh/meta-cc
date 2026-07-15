package executor

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/yaleh/meta-cc/internal/conversation"
	"github.com/yaleh/meta-cc/internal/locator"
	mcquery "github.com/yaleh/meta-cc/internal/mcp/query"
	providerpkg "github.com/yaleh/meta-cc/internal/provider"
	claudeprovider "github.com/yaleh/meta-cc/internal/provider/claude"
	codexprovider "github.com/yaleh/meta-cc/internal/provider/codex"
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

	registry := providerpkg.NewRegistry(
		claudeprovider.NewProvider(locator.NewSessionLocator(), projectPath),
		codexprovider.NewProvider(locator.NewCodexLocator()),
	)

	filters, err := parseProviderFilter(providerName)
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

func parseProviderFilter(providerName string) ([]conversation.ProviderID, error) {
	switch providerName {
	case "claude":
		return []conversation.ProviderID{conversation.ProviderClaude}, nil
	case "codex":
		return []conversation.ProviderID{conversation.ProviderCodex}, nil
	case "all":
		return []conversation.ProviderID{conversation.ProviderClaude, conversation.ProviderCodex}, nil
	default:
		return nil, fmt.Errorf("invalid provider %q: must be \"claude\", \"codex\", or \"all\"", providerName)
	}
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
