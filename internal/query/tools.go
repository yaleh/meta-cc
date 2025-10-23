package query

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/yaleh/meta-cc/internal/filter"
	"github.com/yaleh/meta-cc/internal/parser"
	pipelinepkg "github.com/yaleh/meta-cc/pkg/pipeline"
)

// Sentinel errors for consistent error handling by callers.
var (
	ErrSessionLoad   = errors.New("query: session load failed")
	ErrFilterInvalid = errors.New("query: invalid filter")
)

// RunToolsQuery loads tool calls using the session pipeline, applies filters, sorting,
// and pagination according to the provided options, and returns the resulting slice.
func RunToolsQuery(opts ToolsQueryOptions) ([]parser.ToolCall, error) {
	pipe := pipelinepkg.NewSessionPipeline(opts.Pipeline)
	if err := pipe.Load(pipelinepkg.LoadOptions{AutoDetect: true}); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSessionLoad, err)
	}

	calls := pipe.ExtractToolCalls()

	filtered, err := applyToolFilters(calls, opts)
	if err != nil {
		return nil, err
	}

	sortToolCalls(filtered, opts.SortBy, opts.Reverse)

	return applyToolPagination(filtered, opts.Limit, opts.Offset), nil
}

func applyToolFilters(toolCalls []parser.ToolCall, opts ToolsQueryOptions) ([]parser.ToolCall, error) {
	filtered := toolCalls
	var err error

	if opts.Expression != "" {
		filtered, err = applyExpressionFilter(filtered, opts.Expression)
		if err != nil {
			return nil, err
		}
	}

	if opts.Where != "" {
		if isAdvancedWhere(opts.Where) {
			normalized := normalizeAdvancedWhere(opts.Where)
			filtered, err = applyExpressionFilter(filtered, normalized)
		} else {
			filtered, err = applySimpleWhere(filtered, opts.Where)
		}
		if err != nil {
			return nil, err
		}
	}

	return applyFlagFilters(filtered, opts.Status, opts.Tool), nil
}

func applyExpressionFilter(toolCalls []parser.ToolCall, expression string) ([]parser.ToolCall, error) {
	if expression == "" {
		return toolCalls, nil
	}

	expr, err := filter.ParseExpression(expression)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFilterInvalid, err)
	}

	var filtered []parser.ToolCall
	for _, tc := range toolCalls {
		record := map[string]interface{}{
			"tool":   tc.ToolName,
			"status": tc.Status,
			"uuid":   tc.UUID,
			"error":  tc.Error,
		}

		match, evalErr := expr.Evaluate(record)
		if evalErr != nil {
			return nil, fmt.Errorf("%w: %v", ErrFilterInvalid, evalErr)
		}

		if match {
			filtered = append(filtered, tc)
		}
	}

	return filtered, nil
}

func applySimpleWhere(toolCalls []parser.ToolCall, where string) ([]parser.ToolCall, error) {
	result, err := filter.ApplyWhere(toolCalls, where, "tool_calls")
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFilterInvalid, err)
	}
	return result.([]parser.ToolCall), nil
}

func applyFlagFilters(toolCalls []parser.ToolCall, status, tool string) []parser.ToolCall {
	var result []parser.ToolCall

	for _, tc := range toolCalls {
		if !matchesStatus(tc, status) {
			continue
		}
		if tool != "" && tc.ToolName != tool {
			continue
		}
		result = append(result, tc)
	}

	return result
}

func matchesStatus(tc parser.ToolCall, status string) bool {
	if status == "" {
		return true
	}

	switch status {
	case "error":
		return tc.Status == "error" || tc.Error != ""
	case "success":
		return tc.Status != "error" && tc.Error == ""
	default:
		return true
	}
}

func sortToolCalls(toolCalls []parser.ToolCall, sortBy string, reverse bool) {
	if sortBy == "" {
		// Default sort by timestamp to maintain deterministic order
		sort.SliceStable(toolCalls, func(i, j int) bool {
			if reverse {
				return toolCalls[i].Timestamp > toolCalls[j].Timestamp
			}
			return toolCalls[i].Timestamp < toolCalls[j].Timestamp
		})
		return
	}

	sort.SliceStable(toolCalls, func(i, j int) bool {
		var less bool

		switch sortBy {
		case "timestamp":
			less = toolCalls[i].Timestamp < toolCalls[j].Timestamp
		case "tool":
			less = toolCalls[i].ToolName < toolCalls[j].ToolName
		case "status":
			less = toolCalls[i].Status < toolCalls[j].Status
		case "uuid":
			less = toolCalls[i].UUID < toolCalls[j].UUID
		default:
			less = toolCalls[i].Timestamp < toolCalls[j].Timestamp
		}

		if reverse {
			return !less
		}
		return less
	})
}

func applyToolPagination(toolCalls []parser.ToolCall, limit, offset int) []parser.ToolCall {
	config := filter.PaginationConfig{Limit: limit, Offset: offset}
	return filter.ApplyPagination(toolCalls, config)
}

func isAdvancedWhere(where string) bool {
	lower := strings.ToLower(where)
	if strings.Contains(lower, " like ") || strings.Contains(lower, " between ") || strings.Contains(lower, " in ") {
		return true
	}
	if strings.Contains(lower, " and ") || strings.Contains(lower, " or ") {
		return true
	}
	if strings.ContainsAny(where, "%'_") {
		return true
	}
	if strings.Contains(where, ">") || strings.Contains(where, "<") {
		return true
	}
	return false
}

func normalizeAdvancedWhere(where string) string {
	replacer := strings.NewReplacer("=", " = ", ">", " > ", "<", " < ")
	normalized := replacer.Replace(where)
	return strings.Join(strings.Fields(normalized), " ")
}
