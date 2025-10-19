package validation

import (
	"testing"
)

func TestValidateParameterOrdering(t *testing.T) {
	tests := []struct {
		name         string
		tool         Tool
		expectStatus string
	}{
		{
			name: "correct ordering with all tiers",
			tool: Tool{
				Name: "query_tools",
				InputSchema: InputSchema{
					Properties: map[string]Property{
						"pattern":  {Type: "string", Description: "Pattern to match (required)"},
						"tool":     {Type: "string", Description: "Filter by tool name"},
						"min_time": {Type: "number", Description: "Minimum time"},
						"limit":    {Type: "number", Description: "Max results"},
					},
					Required: []string{"pattern"},
				},
			},
			expectStatus: "PASS",
		},
		{
			name: "tool with no parameters",
			tool: Tool{
				Name: "simple_tool",
				InputSchema: InputSchema{
					Properties: map[string]Property{},
				},
			},
			expectStatus: "PASS",
		},
		{
			name: "tool with only required parameters",
			tool: Tool{
				Name: "required_only",
				InputSchema: InputSchema{
					Properties: map[string]Property{
						"name": {Type: "string", Description: "Name (required)"},
						"path": {Type: "string", Description: "Path (required)"},
					},
					Required: []string{"name", "path"},
				},
			},
			expectStatus: "PASS",
		},
		{
			name: "tool with filtering parameters",
			tool: Tool{
				Name: "filter_tool",
				InputSchema: InputSchema{
					Properties: map[string]Property{
						"status":   {Type: "string", Description: "Filter by status"},
						"category": {Type: "string", Description: "Filter by category"},
					},
				},
			},
			expectStatus: "PASS",
		},
		{
			name: "tool with range parameters",
			tool: Tool{
				Name: "range_tool",
				InputSchema: InputSchema{
					Properties: map[string]Property{
						"min_value": {Type: "number", Description: "Minimum value"},
						"max_value": {Type: "number", Description: "Maximum value"},
						"threshold": {Type: "number", Description: "Threshold"},
					},
				},
			},
			expectStatus: "PASS",
		},
		{
			name: "tool with output control parameters",
			tool: Tool{
				Name: "output_tool",
				InputSchema: InputSchema{
					Properties: map[string]Property{
						"limit":  {Type: "number", Description: "Max results"},
						"offset": {Type: "number", Description: "Offset for pagination"},
					},
				},
			},
			expectStatus: "PASS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateParameterOrdering(tt.tool)

			if result.Status != tt.expectStatus {
				t.Errorf("Status = %v, want %v", result.Status, tt.expectStatus)
			}
		})
	}
}

func TestCategorizeParameters(t *testing.T) {
	tool := Tool{
		Name: "test_tool",
		InputSchema: InputSchema{
			Properties: map[string]Property{
				"pattern":   {Type: "string"},
				"tool":      {Type: "string"},
				"min_value": {Type: "number"},
				"limit":     {Type: "number"},
			},
			Required: []string{"pattern"},
		},
	}

	tiers := categorizeParameters(tool)

	// Check Tier 1 (Required)
	if len(tiers[1]) != 1 || tiers[1][0] != "pattern" {
		t.Errorf("Tier 1 = %v, want [pattern]", tiers[1])
	}

	// Check that other params are categorized
	if len(tiers[2]) == 0 && len(tiers[3]) == 0 && len(tiers[4]) == 0 {
		t.Error("All non-required params should be categorized")
	}
}

func TestIsRequired(t *testing.T) {
	required := []string{"name", "path"}

	tests := []struct {
		param string
		want  bool
	}{
		{"name", true},
		{"path", true},
		{"optional", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.param, func(t *testing.T) {
			got := isRequired(tt.param, required)
			if got != tt.want {
				t.Errorf("isRequired(%q) = %v, want %v", tt.param, got, tt.want)
			}
		})
	}
}

func TestIsFilteringParam(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"tool", true},
		{"status", true},
		{"pattern", true},
		{"filter", true},
		{"where", true},
		{"include_builtin", true},
		{"pattern_target", true},
		{"unrelated_param", false},
		{"limit", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isFilteringParam(tt.name)
			if got != tt.want {
				t.Errorf("isFilteringParam(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestIsRangeParam(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"min_value", true},
		{"max_value", true},
		{"start_time", true},
		{"end_time", true},
		{"threshold", true},
		{"window", true},
		{"limit", false},
		{"pattern", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isRangeParam(tt.name)
			if got != tt.want {
				t.Errorf("isRangeParam(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestIsOutputParam(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"limit", true},
		{"offset", true},
		{"page", true},
		{"cursor", true},
		{"content_summary", true},
		{"pattern", false},
		{"min_value", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isOutputParam(tt.name)
			if got != tt.want {
				t.Errorf("isOutputParam(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}
