package analyzer

import (
	"testing"
	"time"

	"github.com/yale/meta-cc/internal/parser"
)

// TestDetectToolSequences tests tool sequence detection
func TestDetectToolSequences(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name           string
		entries        []parser.SessionEntry
		minLength      int
		minOccurrences int
		expectedCount  int // Number of distinct patterns expected
	}{
		{
			name: "detect repeated Read-Edit-Bash sequence",
			entries: []parser.SessionEntry{
				makeToolUseEntry("1", "Read", now.Add(0*time.Second)),
				makeToolUseEntry("2", "Edit", now.Add(1*time.Second)),
				makeToolUseEntry("3", "Bash", now.Add(2*time.Second)),
				makeToolUseEntry("4", "Read", now.Add(10*time.Second)),
				makeToolUseEntry("5", "Edit", now.Add(11*time.Second)),
				makeToolUseEntry("6", "Bash", now.Add(12*time.Second)),
				makeToolUseEntry("7", "Read", now.Add(20*time.Second)),
				makeToolUseEntry("8", "Edit", now.Add(21*time.Second)),
				makeToolUseEntry("9", "Bash", now.Add(22*time.Second)),
			},
			minLength:      3,
			minOccurrences: 3,
			expectedCount:  1, // "Read → Edit → Bash" appears 3 times
		},
		{
			name: "no sequences when below threshold",
			entries: []parser.SessionEntry{
				makeToolUseEntry("1", "Read", now.Add(0*time.Second)),
				makeToolUseEntry("2", "Edit", now.Add(1*time.Second)),
				makeToolUseEntry("3", "Read", now.Add(2*time.Second)),
				makeToolUseEntry("4", "Edit", now.Add(3*time.Second)),
			},
			minLength:      3,
			minOccurrences: 3,
			expectedCount:  0, // No sequence appears 3 times
		},
		{
			name: "multiple overlapping sequences",
			entries: []parser.SessionEntry{
				makeToolUseEntry("1", "Read", now.Add(0*time.Second)),
				makeToolUseEntry("2", "Grep", now.Add(1*time.Second)),
				makeToolUseEntry("3", "Read", now.Add(2*time.Second)),
				makeToolUseEntry("4", "Grep", now.Add(3*time.Second)),
				makeToolUseEntry("5", "Read", now.Add(4*time.Second)),
				makeToolUseEntry("6", "Grep", now.Add(5*time.Second)),
			},
			minLength:      2,
			minOccurrences: 3,
			expectedCount:  1, // "Read → Grep" appears 3 times
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectToolSequences(tt.entries, tt.minLength, tt.minOccurrences)

			if len(result.Sequences) != tt.expectedCount {
				t.Errorf("expected %d sequences, got %d", tt.expectedCount, len(result.Sequences))
			}

			// Verify sequence has correct occurrences count
			if tt.expectedCount > 0 {
				seq := result.Sequences[0]
				if seq.Count < tt.minOccurrences {
					t.Errorf("sequence count %d is below minimum %d", seq.Count, tt.minOccurrences)
				}
				if len(seq.Occurrences) != seq.Count {
					t.Errorf("occurrences array length %d doesn't match count %d", len(seq.Occurrences), seq.Count)
				}
			}
		})
	}
}

// TestDetectFileChurn tests file churn detection
func TestDetectFileChurn(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name          string
		entries       []parser.SessionEntry
		threshold     int
		expectedFiles int
	}{
		{
			name: "detect high churn file",
			entries: []parser.SessionEntry{
				makeFileAccessEntry("1", "Read", "test.js", now.Add(0*time.Second)),
				makeFileAccessEntry("2", "Edit", "test.js", now.Add(1*time.Second)),
				makeFileAccessEntry("3", "Read", "test.js", now.Add(2*time.Second)),
				makeFileAccessEntry("4", "Edit", "test.js", now.Add(3*time.Second)),
				makeFileAccessEntry("5", "Read", "test.js", now.Add(4*time.Second)),
				makeFileAccessEntry("6", "Edit", "test.js", now.Add(5*time.Second)),
				makeFileAccessEntry("7", "Read", "other.js", now.Add(6*time.Second)),
			},
			threshold:     5,
			expectedFiles: 1, // test.js has 6 accesses
		},
		{
			name: "no files above threshold",
			entries: []parser.SessionEntry{
				makeFileAccessEntry("1", "Read", "test.js", now.Add(0*time.Second)),
				makeFileAccessEntry("2", "Edit", "test.js", now.Add(1*time.Second)),
				makeFileAccessEntry("3", "Read", "other.js", now.Add(2*time.Second)),
			},
			threshold:     5,
			expectedFiles: 0,
		},
		{
			name: "multiple high churn files",
			entries: []parser.SessionEntry{
				makeFileAccessEntry("1", "Read", "a.js", now.Add(0*time.Second)),
				makeFileAccessEntry("2", "Edit", "a.js", now.Add(1*time.Second)),
				makeFileAccessEntry("3", "Read", "a.js", now.Add(2*time.Second)),
				makeFileAccessEntry("4", "Read", "b.js", now.Add(3*time.Second)),
				makeFileAccessEntry("5", "Edit", "b.js", now.Add(4*time.Second)),
				makeFileAccessEntry("6", "Read", "b.js", now.Add(5*time.Second)),
			},
			threshold:     3,
			expectedFiles: 2, // Both files have 3 accesses
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectFileChurn(tt.entries, tt.threshold)

			if len(result.HighChurnFiles) != tt.expectedFiles {
				t.Errorf("expected %d high churn files, got %d", tt.expectedFiles, len(result.HighChurnFiles))
			}

			// Verify each file has correct total
			for _, file := range result.HighChurnFiles {
				if file.TotalAccesses < tt.threshold {
					t.Errorf("file %s has %d accesses, below threshold %d", file.File, file.TotalAccesses, tt.threshold)
				}
				// Verify total equals sum of operations
				sum := file.ReadCount + file.EditCount + file.WriteCount
				if sum != file.TotalAccesses {
					t.Errorf("file %s: sum of operations (%d) != total accesses (%d)", file.File, sum, file.TotalAccesses)
				}
			}
		})
	}
}

// TestDetectIdlePeriods tests idle period detection
func TestDetectIdlePeriods(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name           string
		entries        []parser.SessionEntry
		thresholdMin   int
		expectedPeriods int
	}{
		{
			name: "detect single idle period",
			entries: []parser.SessionEntry{
				makeToolUseEntry("1", "Bash", now.Add(0*time.Second)),
				makeToolUseEntry("2", "Read", now.Add(10*time.Minute)), // 10 minute gap
			},
			thresholdMin:    5,
			expectedPeriods: 1,
		},
		{
			name: "no idle periods when below threshold",
			entries: []parser.SessionEntry{
				makeToolUseEntry("1", "Bash", now.Add(0*time.Second)),
				makeToolUseEntry("2", "Read", now.Add(2*time.Minute)), // 2 minute gap
			},
			thresholdMin:    5,
			expectedPeriods: 0,
		},
		{
			name: "multiple idle periods",
			entries: []parser.SessionEntry{
				makeToolUseEntry("1", "Bash", now.Add(0*time.Second)),
				makeToolUseEntry("2", "Read", now.Add(10*time.Minute)),  // Gap 1: 10min
				makeToolUseEntry("3", "Edit", now.Add(12*time.Minute)),  // Gap: 2min (below threshold)
				makeToolUseEntry("4", "Bash", now.Add(20*time.Minute)),  // Gap 2: 8min
			},
			thresholdMin:    5,
			expectedPeriods: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectIdlePeriods(tt.entries, tt.thresholdMin)

			if len(result.IdlePeriods) != tt.expectedPeriods {
				t.Errorf("expected %d idle periods, got %d", tt.expectedPeriods, len(result.IdlePeriods))
			}

			// Verify each idle period exceeds threshold
			for _, period := range result.IdlePeriods {
				if period.DurationMin < float64(tt.thresholdMin) {
					t.Errorf("idle period duration %.2f min is below threshold %d min", period.DurationMin, tt.thresholdMin)
				}
			}
		})
	}
}

// Helper functions to create test data

func makeToolUseEntry(uuid, toolName string, ts time.Time) parser.SessionEntry {
	return parser.SessionEntry{
		Type:      "assistant",
		UUID:      uuid,
		Timestamp: ts.Format(time.RFC3339),
		Message: &parser.Message{
			Role: "assistant",
			Content: []parser.ContentBlock{
				{
					Type: "tool_use",
					ToolUse: &parser.ToolUse{
						ID:    uuid + "-tool",
						Name:  toolName,
						Input: map[string]interface{}{},
					},
				},
			},
		},
	}
}

func makeFileAccessEntry(uuid, toolName, filePath string, ts time.Time) parser.SessionEntry {
	return parser.SessionEntry{
		Type:      "assistant",
		UUID:      uuid,
		Timestamp: ts.Format(time.RFC3339),
		Message: &parser.Message{
			Role: "assistant",
			Content: []parser.ContentBlock{
				{
					Type: "tool_use",
					ToolUse: &parser.ToolUse{
						ID:   uuid + "-tool",
						Name: toolName,
						Input: map[string]interface{}{
							"file_path": filePath,
						},
					},
				},
			},
		},
	}
}
