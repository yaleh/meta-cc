package query

import (
	"testing"
	"time"

	"github.com/yaleh/meta-cc/internal/parser"
)

func TestBuildFileAccessQuery(t *testing.T) {
	now := time.Now()

	entries := []parser.SessionEntry{
		{
			UUID:      "uuid-1",
			Type:      "assistant",
			Timestamp: now.Format(time.RFC3339Nano),
			Message: &parser.Message{
				Role: "assistant",
				Content: []parser.ContentBlock{
					{
						Type: "tool_use",
						ToolUse: &parser.ToolUse{
							ID:   "tool-1",
							Name: "Read",
							Input: map[string]interface{}{
								"file_path": "/path/to/test.js",
							},
						},
					},
				},
			},
		},
		{
			UUID:      "uuid-2",
			Type:      "user",
			Timestamp: now.Add(1 * time.Minute).Format(time.RFC3339Nano),
			Message: &parser.Message{
				Role: "user",
				Content: []parser.ContentBlock{
					{
						Type: "tool_result",
						ToolResult: &parser.ToolResult{
							ToolUseID: "tool-1",
							Status:    "success",
							Content:   "file contents",
						},
					},
				},
			},
		},
		{
			UUID:      "uuid-3",
			Type:      "assistant",
			Timestamp: now.Add(2 * time.Minute).Format(time.RFC3339Nano),
			Message: &parser.Message{
				Role: "assistant",
				Content: []parser.ContentBlock{
					{
						Type: "tool_use",
						ToolUse: &parser.ToolUse{
							ID:   "tool-2",
							Name: "Edit",
							Input: map[string]interface{}{
								"file_path": "/path/to/test.js",
							},
						},
					},
				},
			},
		},
		{
			UUID:      "uuid-4",
			Type:      "user",
			Timestamp: now.Add(3 * time.Minute).Format(time.RFC3339Nano),
			Message: &parser.Message{
				Role: "user",
				Content: []parser.ContentBlock{
					{
						Type: "tool_result",
						ToolResult: &parser.ToolResult{
							ToolUseID: "tool-2",
							Status:    "success",
							Content:   "edited",
						},
					},
				},
			},
		},
	}

	tests := []struct {
		name          string
		filePath      string
		wantAccesses  int
		wantReadCount int
		wantEditCount int
		wantErr       bool
	}{
		{
			name:          "full path match",
			filePath:      "/path/to/test.js",
			wantAccesses:  2,
			wantReadCount: 1,
			wantEditCount: 1,
			wantErr:       false,
		},
		{
			name:          "basename match",
			filePath:      "test.js",
			wantAccesses:  2,
			wantReadCount: 1,
			wantEditCount: 1,
			wantErr:       false,
		},
		{
			name:         "non-existent file",
			filePath:     "other.js",
			wantAccesses: 0,
			wantErr:      false,
		},
		{
			name:     "empty file path",
			filePath: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildFileAccessQuery(entries, tt.filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildFileAccessQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if got.TotalAccesses != tt.wantAccesses {
				t.Errorf("TotalAccesses = %d, want %d", got.TotalAccesses, tt.wantAccesses)
			}

			if tt.wantAccesses > 0 {
				if got.Operations["Read"] != tt.wantReadCount {
					t.Errorf("Read count = %d, want %d", got.Operations["Read"], tt.wantReadCount)
				}
				if got.Operations["Edit"] != tt.wantEditCount {
					t.Errorf("Edit count = %d, want %d", got.Operations["Edit"], tt.wantEditCount)
				}

				// Check timeline is sorted
				for i := 1; i < len(got.Timeline); i++ {
					if got.Timeline[i].Turn < got.Timeline[i-1].Turn {
						t.Error("Timeline is not sorted by turn")
						break
					}
				}

				// Check time span
				if got.TimeSpanMin <= 0 {
					t.Error("TimeSpanMin should be positive for multiple accesses")
				}
			}
		})
	}
}

func TestExtractFileFromToolCall(t *testing.T) {
	tests := []struct {
		name     string
		toolCall parser.ToolCall
		want     string
	}{
		{
			name: "file_path parameter",
			toolCall: parser.ToolCall{
				Input: map[string]interface{}{
					"file_path": "/path/to/file.js",
				},
			},
			want: "/path/to/file.js",
		},
		{
			name: "notebook_path parameter",
			toolCall: parser.ToolCall{
				Input: map[string]interface{}{
					"notebook_path": "/path/to/notebook.ipynb",
				},
			},
			want: "/path/to/notebook.ipynb",
		},
		{
			name: "no file parameter",
			toolCall: parser.ToolCall{
				Input: map[string]interface{}{
					"command": "ls",
				},
			},
			want: "",
		},
		{
			name: "empty file_path",
			toolCall: parser.ToolCall{
				Input: map[string]interface{}{
					"file_path": "",
				},
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractFileFromToolCall(tt.toolCall)
			if got != tt.want {
				t.Errorf("extractFileFromToolCall() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMatchesFile(t *testing.T) {
	tests := []struct {
		name         string
		accessedFile string
		queryFile    string
		want         bool
	}{
		{
			name:         "exact match",
			accessedFile: "/path/to/file.js",
			queryFile:    "/path/to/file.js",
			want:         true,
		},
		{
			name:         "basename match",
			accessedFile: "/path/to/file.js",
			queryFile:    "file.js",
			want:         true,
		},
		{
			name:         "different files",
			accessedFile: "/path/to/file1.js",
			queryFile:    "file2.js",
			want:         false,
		},
		{
			name:         "same basename different path",
			accessedFile: "/path1/file.js",
			queryFile:    "/path2/file.js",
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchesFile(tt.accessedFile, tt.queryFile)
			if got != tt.want {
				t.Errorf("matchesFile(%q, %q) = %v, want %v", tt.accessedFile, tt.queryFile, got, tt.want)
			}
		})
	}
}

func TestGetActionType(t *testing.T) {
	tests := []struct {
		toolName string
		want     string
	}{
		{"Read", "Read"},
		{"Edit", "Edit"},
		{"Write", "Write"},
		{"NotebookEdit", "Edit"},
		{"Bash", ""},
		{"Grep", ""},
	}

	for _, tt := range tests {
		t.Run(tt.toolName, func(t *testing.T) {
			got := getActionType(tt.toolName)
			if got != tt.want {
				t.Errorf("getActionType(%q) = %q, want %q", tt.toolName, got, tt.want)
			}
		})
	}
}

func TestCalculateTimeSpan(t *testing.T) {
	now := time.Now().Unix()

	tests := []struct {
		name     string
		timeline []FileAccessEvent
		want     int
	}{
		{
			name: "5 minute span",
			timeline: []FileAccessEvent{
				{Timestamp: now},
				{Timestamp: now + 5*60},
			},
			want: 5,
		},
		{
			name: "single event",
			timeline: []FileAccessEvent{
				{Timestamp: now},
			},
			want: 0,
		},
		{
			name:     "empty timeline",
			timeline: []FileAccessEvent{},
			want:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateTimeSpan(tt.timeline)
			if got != tt.want {
				t.Errorf("calculateTimeSpan() = %d, want %d", got, tt.want)
			}
		})
	}
}
