package query

import (
	"fmt"

	"github.com/yaleh/meta-cc/internal/analyzer"
	"github.com/yaleh/meta-cc/internal/filter"
	"github.com/yaleh/meta-cc/internal/parser"
	"github.com/yaleh/meta-cc/internal/stats"
)

// BuildSessionStats returns aggregated session statistics using analyzer package.
func BuildSessionStats(entries []parser.SessionEntry, toolCalls []parser.ToolCall) analyzer.SessionStats {
	return analyzer.CalculateStats(entries, toolCalls)
}

// AnalyzeTimeSeries applies optional expression filtering and returns time series points.
func AnalyzeTimeSeries(toolCalls []parser.ToolCall, metric, interval, filterExpr string) ([]stats.TimeSeriesPoint, error) {
	filtered := toolCalls

	if filterExpr != "" {
		expr, err := filter.ParseExpression(filterExpr)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrFilterInvalid, err)
		}

		var evalCalls []parser.ToolCall
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
				evalCalls = append(evalCalls, tc)
			}
		}

		filtered = evalCalls
	}

	cfg := stats.TimeSeriesConfig{Metric: metric, Interval: interval}
	return stats.AnalyzeTimeSeries(filtered, cfg)
}
