package stats

import (
	"sort"

	"github.com/yale/meta-cc/internal/parser"
)

// FileStats represents file-level operation statistics
type FileStats struct {
	FilePath   string  `json:"file_path"`
	ReadCount  int     `json:"read_count"`
	EditCount  int     `json:"edit_count"`
	WriteCount int     `json:"write_count"`
	ErrorCount int     `json:"error_count"`
	TotalOps   int     `json:"total_ops"`
	ErrorRate  float64 `json:"error_rate"`
}

// AnalyzeFileStats analyzes file-level statistics from tool calls
func AnalyzeFileStats(toolCalls []parser.ToolCall) []FileStats {
	fileMap := make(map[string]*FileStats)

	for _, tc := range toolCalls {
		// Extract file path from tool input
		filePath := extractFilePath(tc)
		if filePath == "" {
			continue // Skip non-file operations
		}

		// Initialize file stats if not exists
		if _, exists := fileMap[filePath]; !exists {
			fileMap[filePath] = &FileStats{
				FilePath: filePath,
			}
		}

		stats := fileMap[filePath]

		// Count operation type
		switch tc.ToolName {
		case "Read":
			stats.ReadCount++
		case "Edit":
			stats.EditCount++
		case "Write":
			stats.WriteCount++
		case "NotebookEdit":
			stats.EditCount++ // Treat notebook edits as edits
		}

		// Count errors
		if tc.Status == "error" {
			stats.ErrorCount++
		}

		stats.TotalOps++
	}

	// Calculate error rates and convert to slice
	var results []FileStats
	for _, stats := range fileMap {
		if stats.TotalOps > 0 {
			stats.ErrorRate = float64(stats.ErrorCount) / float64(stats.TotalOps)
		}
		results = append(results, *stats)
	}

	// Default sort: by total ops descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].TotalOps > results[j].TotalOps
	})

	return results
}

// extractFilePath extracts file path from ToolCall input
func extractFilePath(tc parser.ToolCall) string {
	// Try common field names
	if filePath, ok := tc.Input["file_path"].(string); ok {
		return filePath
	}

	if filePath, ok := tc.Input["path"].(string); ok {
		return filePath
	}

	if filePath, ok := tc.Input["notebook_path"].(string); ok {
		return filePath
	}

	return ""
}

// SortFileStats sorts file statistics by specified field
func SortFileStats(stats []FileStats, sortBy string) {
	sort.Slice(stats, func(i, j int) bool {
		switch sortBy {
		case "read_count":
			return stats[i].ReadCount > stats[j].ReadCount
		case "edit_count":
			return stats[i].EditCount > stats[j].EditCount
		case "write_count":
			return stats[i].WriteCount > stats[j].WriteCount
		case "error_count":
			return stats[i].ErrorCount > stats[j].ErrorCount
		case "error_rate":
			return stats[i].ErrorRate > stats[j].ErrorRate
		default: // "total_ops"
			return stats[i].TotalOps > stats[j].TotalOps
		}
	})
}
