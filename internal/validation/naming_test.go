package validation

import (
	"testing"
)

func TestValidateNaming(t *testing.T) {
	tests := []struct {
		name     string
		toolName string
		wantPass bool
	}{
		{"valid query prefix", "query_tools", true},
		{"valid get prefix", "get_capability", true},
		{"valid list prefix", "list_capabilities", true},
		{"valid cleanup prefix", "cleanup_temp_files", true},
		{"invalid prefix", "retrieve_data", false}, // Should use valid prefix
		{"not snake_case", "queryTools", false},
		{"too long", "query_very_long_tool_name_that_exceeds_the_forty_character_limit", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := Tool{Name: tt.toolName}
			result := ValidateNaming(tool)

			isPass := result.Status == "PASS"
			if isPass != tt.wantPass {
				t.Errorf("ValidateNaming(%q) = %v (status=%s), want pass=%v",
					tt.toolName, result, result.Status, tt.wantPass)
			}
		})
	}
}

func TestHasValidPrefix(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"query_tools", true},
		{"get_capability", true},
		{"list_capabilities", true},
		{"cleanup_temp_files", true},
		{"retrieve_data", false},
		{"fetch_data", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasValidPrefix(tt.name); got != tt.want {
				t.Errorf("hasValidPrefix(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestIsSnakeCase(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"query_tools", true},
		{"get_capability", true},
		{"queryTools", false},  // camelCase
		{"QueryTools", false},  // PascalCase
		{"query tools", false}, // spaces
		{"query", false},       // no underscore
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isSnakeCase(tt.name); got != tt.want {
				t.Errorf("isSnakeCase(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}
