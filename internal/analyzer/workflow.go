package analyzer

import (
	"sort"
	"strings"
	"time"

	"github.com/yale/meta-cc/internal/parser"
)

// SequenceAnalysis represents tool sequence analysis results
type SequenceAnalysis struct {
	Sequences []SequencePattern `json:"sequences"`
}

// SequencePattern represents a repeated tool call sequence
type SequencePattern struct {
	Pattern     string               `json:"pattern"`
	Length      int                  `json:"length"`
	Count       int                  `json:"count"`
	Occurrences []SequenceOccurrence `json:"occurrences"`
	TimeSpanMin int                  `json:"time_span_minutes"`
}

// SequenceOccurrence represents a single occurrence of a sequence
type SequenceOccurrence struct {
	StartTurn int              `json:"start_turn"`
	EndTurn   int              `json:"end_turn"`
	Tools     []ToolInSequence `json:"tools"`
}

// ToolInSequence represents a tool call within a sequence
type ToolInSequence struct {
	Turn    int    `json:"turn"`
	Tool    string `json:"tool"`
	File    string `json:"file,omitempty"`
	Command string `json:"command,omitempty"`
}

// FileChurnAnalysis represents file churn analysis results
type FileChurnAnalysis struct {
	HighChurnFiles []FileChurnDetail `json:"high_churn_files"`
}

// FileChurnDetail represents detailed file access statistics
type FileChurnDetail struct {
	File          string `json:"file"`
	ReadCount     int    `json:"read_count"`
	EditCount     int    `json:"edit_count"`
	WriteCount    int    `json:"write_count"`
	TotalAccesses int    `json:"total_accesses"`
	TimeSpanMin   int    `json:"time_span_minutes"`
	FirstAccess   int64  `json:"first_access"`
	LastAccess    int64  `json:"last_access"`
}

// IdlePeriodAnalysis represents idle period analysis results
type IdlePeriodAnalysis struct {
	IdlePeriods []IdlePeriod `json:"idle_periods"`
}

// IdlePeriod represents a detected idle period
type IdlePeriod struct {
	StartTurn      int          `json:"start_turn"`
	EndTurn        int          `json:"end_turn"`
	DurationMin    float64      `json:"duration_minutes"`
	StartTimestamp int64        `json:"start_timestamp"`
	EndTimestamp   int64        `json:"end_timestamp"`
	ContextBefore  *TurnContext `json:"context_before,omitempty"`
	ContextAfter   *TurnContext `json:"context_after,omitempty"`
}

// TurnContext represents context around an event
type TurnContext struct {
	Turn    int    `json:"turn"`
	Role    string `json:"role,omitempty"`
	Tool    string `json:"tool,omitempty"`
	Status  string `json:"status,omitempty"`
	Preview string `json:"preview,omitempty"`
}

// DetectToolSequences detects repeated tool call sequences
func DetectToolSequences(entries []parser.SessionEntry, minLength, minOccurrences int) SequenceAnalysis {
	// Build turn index
	turnIndex := buildTurnIndex(entries)

	// Extract tool calls with turn numbers
	toolCalls := extractToolCallsWithTurns(entries, turnIndex)

	// Sort by turn
	sort.Slice(toolCalls, func(i, j int) bool {
		return toolCalls[i].turn < toolCalls[j].turn
	})

	// Find all sequences
	sequences := findAllSequences(toolCalls, minLength, minOccurrences, entries)

	return SequenceAnalysis{
		Sequences: sequences,
	}
}

// DetectFileChurn detects files with frequent access
func DetectFileChurn(entries []parser.SessionEntry, threshold int) FileChurnAnalysis {
	// Extract file access events
	fileAccess := make(map[string]*fileAccessStats)

	toolCalls := parser.ExtractToolCalls(entries)
	for _, tc := range toolCalls {
		// Extract file path
		filePath := extractFileFromToolCall(tc)
		if filePath == "" {
			continue
		}

		// Get action type
		action := getActionType(tc.ToolName)
		if action == "" {
			continue
		}

		// Get timestamp
		timestamp := getToolCallTimestamp(entries, tc.UUID)

		// Initialize or update stats
		if _, exists := fileAccess[filePath]; !exists {
			fileAccess[filePath] = &fileAccessStats{
				file:        filePath,
				firstAccess: timestamp,
				lastAccess:  timestamp,
			}
		}

		stats := fileAccess[filePath]
		stats.totalAccesses++

		switch action {
		case "Read":
			stats.readCount++
		case "Edit":
			stats.editCount++
		case "Write":
			stats.writeCount++
		}

		if timestamp < stats.firstAccess {
			stats.firstAccess = timestamp
		}
		if timestamp > stats.lastAccess {
			stats.lastAccess = timestamp
		}
	}

	// Filter by threshold and build result
	var highChurnFiles []FileChurnDetail
	for _, stats := range fileAccess {
		if stats.totalAccesses >= threshold {
			timeSpan := 0
			if stats.lastAccess > stats.firstAccess {
				timeSpan = int((stats.lastAccess - stats.firstAccess) / 60)
			}

			highChurnFiles = append(highChurnFiles, FileChurnDetail{
				File:          stats.file,
				ReadCount:     stats.readCount,
				EditCount:     stats.editCount,
				WriteCount:    stats.writeCount,
				TotalAccesses: stats.totalAccesses,
				TimeSpanMin:   timeSpan,
				FirstAccess:   stats.firstAccess,
				LastAccess:    stats.lastAccess,
			})
		}
	}

	// Sort by total accesses (descending)
	sort.Slice(highChurnFiles, func(i, j int) bool {
		return highChurnFiles[i].TotalAccesses > highChurnFiles[j].TotalAccesses
	})

	return FileChurnAnalysis{
		HighChurnFiles: highChurnFiles,
	}
}

// DetectIdlePeriods detects idle periods in the session
func DetectIdlePeriods(entries []parser.SessionEntry, thresholdMin int) IdlePeriodAnalysis {
	// Build turn index
	turnIndex := buildTurnIndex(entries)

	// Extract all entries with timestamps (both user and assistant)
	type entryWithTurn struct {
		entry parser.SessionEntry
		turn  int
	}

	var entriesWithTurns []entryWithTurn
	for _, entry := range entries {
		if turn, ok := turnIndex[entry.UUID]; ok {
			entriesWithTurns = append(entriesWithTurns, entryWithTurn{
				entry: entry,
				turn:  turn,
			})
		}
	}

	// Sort by turn
	sort.Slice(entriesWithTurns, func(i, j int) bool {
		return entriesWithTurns[i].turn < entriesWithTurns[j].turn
	})

	// Find idle periods
	var idlePeriods []IdlePeriod
	thresholdSec := float64(thresholdMin * 60)

	for i := 0; i < len(entriesWithTurns)-1; i++ {
		current := entriesWithTurns[i]
		next := entriesWithTurns[i+1]

		currentTs := parseTimestamp(current.entry.Timestamp)
		nextTs := parseTimestamp(next.entry.Timestamp)

		if currentTs == 0 || nextTs == 0 {
			continue
		}

		gapSec := float64(nextTs - currentTs)
		if gapSec >= thresholdSec {
			// Found an idle period
			period := IdlePeriod{
				StartTurn:      current.turn,
				EndTurn:        next.turn,
				DurationMin:    gapSec / 60,
				StartTimestamp: currentTs,
				EndTimestamp:   nextTs,
			}

			// Add context
			period.ContextBefore = extractTurnContext(current.entry, current.turn)
			period.ContextAfter = extractTurnContext(next.entry, next.turn)

			idlePeriods = append(idlePeriods, period)
		}
	}

	return IdlePeriodAnalysis{
		IdlePeriods: idlePeriods,
	}
}

// Helper types and functions

type fileAccessStats struct {
	file          string
	readCount     int
	editCount     int
	writeCount    int
	totalAccesses int
	firstAccess   int64
	lastAccess    int64
}

type toolCallWithTurn struct {
	toolName string
	turn     int
	uuid     string
	filePath string
	command  string
}

func buildTurnIndex(entries []parser.SessionEntry) map[string]int {
	turnIndex := make(map[string]int)
	turn := 1
	for _, entry := range entries {
		if entry.IsMessage() {
			turnIndex[entry.UUID] = turn
			turn++
		}
	}
	return turnIndex
}

func extractToolCallsWithTurns(entries []parser.SessionEntry, turnIndex map[string]int) []toolCallWithTurn {
	var result []toolCallWithTurn

	toolCalls := parser.ExtractToolCalls(entries)
	for _, tc := range toolCalls {
		if turn, ok := turnIndex[tc.UUID]; ok {
			result = append(result, toolCallWithTurn{
				toolName: tc.ToolName,
				turn:     turn,
				uuid:     tc.UUID,
				filePath: extractFileFromToolCall(tc),
				command:  extractCommandFromToolCall(tc),
			})
		}
	}

	return result
}

func findAllSequences(toolCalls []toolCallWithTurn, minLength, minOccurrences int, entries []parser.SessionEntry) []SequencePattern {
	sequenceMap := make(map[string][]SequenceOccurrence)

	// Try sequences of different lengths
	maxLen := 5
	if maxLen > len(toolCalls) {
		maxLen = len(toolCalls)
	}

	for seqLen := minLength; seqLen <= maxLen; seqLen++ {
		for i := 0; i <= len(toolCalls)-seqLen; i++ {
			// Extract sequence
			tools := make([]string, seqLen)
			for j := 0; j < seqLen; j++ {
				tools[j] = toolCalls[i+j].toolName
			}

			// Create pattern string
			pattern := strings.Join(tools, " → ")

			// Build occurrence with tool details
			var toolsInSeq []ToolInSequence
			for j := 0; j < seqLen; j++ {
				tc := toolCalls[i+j]
				toolsInSeq = append(toolsInSeq, ToolInSequence{
					Turn:    tc.turn,
					Tool:    tc.toolName,
					File:    tc.filePath,
					Command: tc.command,
				})
			}

			occurrence := SequenceOccurrence{
				StartTurn: toolCalls[i].turn,
				EndTurn:   toolCalls[i+seqLen-1].turn,
				Tools:     toolsInSeq,
			}

			sequenceMap[pattern] = append(sequenceMap[pattern], occurrence)
		}
	}

	// Filter by minimum occurrences and build result
	var result []SequencePattern
	for pattern, occurrences := range sequenceMap {
		if len(occurrences) >= minOccurrences {
			// Calculate length
			length := len(strings.Split(pattern, " → "))

			// Calculate time span
			timeSpan := calculateSequenceTimeSpan(occurrences, entries)

			result = append(result, SequencePattern{
				Pattern:     pattern,
				Length:      length,
				Count:       len(occurrences),
				Occurrences: occurrences,
				TimeSpanMin: timeSpan,
			})
		}
	}

	// Sort by count (descending), then by length (descending)
	sort.Slice(result, func(i, j int) bool {
		if result[i].Count != result[j].Count {
			return result[i].Count > result[j].Count
		}
		return result[i].Length > result[j].Length
	})

	return result
}

func calculateSequenceTimeSpan(occurrences []SequenceOccurrence, entries []parser.SessionEntry) int {
	if len(occurrences) == 0 {
		return 0
	}

	var minTs, maxTs int64

	for _, occ := range occurrences {
		// Find timestamps for turns in this occurrence
		for _, entry := range entries {
			ts := parseTimestamp(entry.Timestamp)
			if ts == 0 {
				continue
			}

			// Check if this entry is part of the occurrence
			for range occ.Tools {
				if entry.UUID != "" && ts > 0 {
					// Update min/max
					if minTs == 0 || ts < minTs {
						minTs = ts
					}
					if ts > maxTs {
						maxTs = ts
					}
					break
				}
			}
		}
	}

	if minTs == 0 || maxTs == 0 {
		return 0
	}

	return int((maxTs - minTs) / 60)
}

func extractFileFromToolCall(tc parser.ToolCall) string {
	fileParams := []string{"file_path", "notebook_path", "path"}

	for _, param := range fileParams {
		if val, ok := tc.Input[param]; ok {
			if filePath, ok := val.(string); ok && filePath != "" {
				return filePath
			}
		}
	}

	return ""
}

func extractCommandFromToolCall(tc parser.ToolCall) string {
	if tc.ToolName == "Bash" {
		if val, ok := tc.Input["command"]; ok {
			if cmd, ok := val.(string); ok {
				// Return first line only for preview
				lines := strings.Split(cmd, "\n")
				if len(lines) > 0 {
					return lines[0]
				}
			}
		}
	}
	return ""
}

func getActionType(toolName string) string {
	switch toolName {
	case "Read":
		return "Read"
	case "Edit":
		return "Edit"
	case "Write":
		return "Write"
	case "NotebookEdit":
		return "Edit"
	default:
		return ""
	}
}

func getToolCallTimestamp(entries []parser.SessionEntry, uuid string) int64 {
	for _, entry := range entries {
		if entry.UUID == uuid {
			return parseTimestamp(entry.Timestamp)
		}
	}
	return 0
}

func parseTimestamp(ts string) int64 {
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return 0
	}
	return t.Unix()
}

func extractTurnContext(entry parser.SessionEntry, turn int) *TurnContext {
	ctx := &TurnContext{
		Turn: turn,
		Role: entry.Type,
	}

	if entry.Message != nil {
		// Extract tool info
		for _, block := range entry.Message.Content {
			if block.Type == "tool_use" && block.ToolUse != nil {
				ctx.Tool = block.ToolUse.Name
			} else if block.Type == "tool_result" && block.ToolResult != nil {
				ctx.Status = block.ToolResult.Status
				if ctx.Status == "" && block.ToolResult.Error != "" {
					ctx.Status = "error"
				}
			} else if block.Type == "text" && block.Text != "" {
				// Extract preview (first 100 chars)
				preview := block.Text
				if len(preview) > 100 {
					preview = preview[:100] + "..."
				}
				ctx.Preview = preview
			}
		}
	}

	return ctx
}
