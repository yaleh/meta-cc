package analyzer

import (
	"sort"
	"time"

	"github.com/yaleh/meta-cc/internal/parser"
)

// SessionStats represents statistical information about a session
type SessionStats struct {
	TurnCount          int            // Total turn count (all message rounds)
	UserTurnCount      int            // User turn count
	AssistantTurnCount int            // Assistant turn count
	ToolCallCount      int            // Total tool calls
	ErrorCount         int            // Error tool calls count
	DurationSeconds    int64          // Session duration (seconds)
	ToolFrequency      map[string]int // Tool usage frequency (tool name -> count)
	ErrorRate          float64        // Error rate (percentage)
	TopTools           []ToolFreq     // Most frequently used tools (Top 5)
}

// ToolFreq represents tool usage frequency
type ToolFreq struct {
	Name  string // Tool name
	Count int    // Call count
}

// CalculateStats calculates session statistics
func CalculateStats(entries []parser.SessionEntry, toolCalls []parser.ToolCall) SessionStats {
	stats := SessionStats{
		ToolFrequency: make(map[string]int),
	}

	// Calculate turn counts
	stats.TurnCount = len(entries)
	for _, entry := range entries {
		if entry.Type == "user" {
			stats.UserTurnCount++
		} else if entry.Type == "assistant" {
			stats.AssistantTurnCount++
		}
	}

	// Calculate tool call statistics
	stats.ToolCallCount = len(toolCalls)
	for _, tc := range toolCalls {
		// Count errors
		if tc.Status == "error" || tc.Error != "" {
			stats.ErrorCount++
		}

		// Count tool usage frequency
		stats.ToolFrequency[tc.ToolName]++
	}

	// Calculate error rate
	if stats.ToolCallCount > 0 {
		stats.ErrorRate = float64(stats.ErrorCount) / float64(stats.ToolCallCount) * 100
	}

	// Calculate session duration
	if len(entries) >= 2 {
		firstTime, err1 := time.Parse(time.RFC3339, entries[0].Timestamp)
		lastTime, err2 := time.Parse(time.RFC3339, entries[len(entries)-1].Timestamp)

		if err1 == nil && err2 == nil {
			stats.DurationSeconds = int64(lastTime.Sub(firstTime).Seconds())
		}
	}

	// Calculate TopTools (sorted by frequency)
	stats.TopTools = calculateTopTools(stats.ToolFrequency, 5)

	return stats
}

// calculateTopTools calculates the most frequently used tools (Top N)
func calculateTopTools(frequency map[string]int, topN int) []ToolFreq {
	var tools []ToolFreq

	for name, count := range frequency {
		tools = append(tools, ToolFreq{Name: name, Count: count})
	}

	// Sort by count descending
	sort.Slice(tools, func(i, j int) bool {
		if tools[i].Count == tools[j].Count {
			// When counts are equal, sort by name alphabetically
			return tools[i].Name < tools[j].Name
		}
		return tools[i].Count > tools[j].Count
	})

	// Return top N
	if len(tools) > topN {
		tools = tools[:topN]
	}

	return tools
}
