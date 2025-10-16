package validation

import (
	"strings"
	"testing"
)

func TestValidateDescription(t *testing.T) {
	tests := []struct {
		name          string
		tool          Tool
		expectStatus  string // "PASS", "FAIL", "WARN"
		expectCheck   string
		expectMessage string
	}{
		{
			name: "valid description under 100 chars",
			tool: Tool{
				Name:        "test_tool",
				Description: "Query tool calls with filters. Default scope: project.",
			},
			expectStatus: "PASS",
			expectCheck:  "description",
		},
		{
			name: "missing Default scope suffix",
			tool: Tool{
				Name:        "bad_tool",
				Description: "Query tool calls with filters.",
			},
			expectStatus:  "FAIL",
			expectCheck:   "description_scope",
			expectMessage: "Description must include 'Default scope:' suffix",
		},
		{
			name: "invalid scope value",
			tool: Tool{
				Name:        "bad_tool",
				Description: "Query tool calls with filters. Default scope: global.",
			},
			expectStatus:  "FAIL",
			expectCheck:   "description_format",
			expectMessage: "Description must match template format",
		},
		{
			name: "missing period after scope",
			tool: Tool{
				Name:        "bad_tool",
				Description: "Query tool calls with filters. Default scope: project",
			},
			expectStatus:  "FAIL",
			expectCheck:   "description_format",
			expectMessage: "Description must match template format",
		},
		{
			name: "does not start with capital letter",
			tool: Tool{
				Name:        "bad_tool",
				Description: "query tool calls with filters. Default scope: project.",
			},
			expectStatus:  "FAIL",
			expectCheck:   "description_format",
			expectMessage: "Description must match template format",
		},
		{
			name: "description over 100 characters",
			tool: Tool{
				Name:        "long_tool",
				Description: "Query tool call history across project with filters including tool name status and timestamps. Default scope: project.",
			},
			expectStatus:  "WARN",
			expectCheck:   "description_length",
			expectMessage: "Description exceeds 100 characters",
		},
		{
			name: "valid description with session scope",
			tool: Tool{
				Name:        "session_tool",
				Description: "Get session statistics and metrics. Default scope: session.",
			},
			expectStatus: "PASS",
			expectCheck:  "description",
		},
		{
			name: "valid description with none scope",
			tool: Tool{
				Name:        "cleanup_tool",
				Description: "Remove old temporary MCP files. Default scope: none.",
			},
			expectStatus: "PASS",
			expectCheck:  "description",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateDescription(tt.tool)

			if result.Status != tt.expectStatus {
				t.Errorf("Status = %v, want %v", result.Status, tt.expectStatus)
			}

			if result.Check != tt.expectCheck {
				t.Errorf("Check = %v, want %v", result.Check, tt.expectCheck)
			}

			if tt.expectMessage != "" && !strings.Contains(result.Message, tt.expectMessage) {
				t.Errorf("Message = %v, expected to contain %v", result.Message, tt.expectMessage)
			}
		})
	}
}
