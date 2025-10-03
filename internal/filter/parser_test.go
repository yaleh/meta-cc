package filter

import (
	"testing"
)

// Test ParseExpression with various operators
func TestParseExpression(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		// Comparison operators
		{name: "simple equality", input: "tool='Bash'", wantErr: false},
		{name: "inequality", input: "status!='error'", wantErr: false},
		{name: "greater than", input: "duration>100", wantErr: false},
		{name: "less than or equal", input: "duration<=200", wantErr: false},

		// Boolean operators
		{name: "AND expression", input: "tool='Bash' AND status='error'", wantErr: false},
		{name: "OR expression", input: "status='error' OR status='success'", wantErr: false},
		{name: "NOT expression", input: "NOT status='success'", wantErr: false},

		// Complex expressions
		{name: "nested AND/OR", input: "tool='Bash' AND (status='error' OR status='success')", wantErr: false},
		{name: "multiple AND", input: "tool='Bash' AND status='error' AND duration>100", wantErr: false},

		// Set operators
		{name: "IN operator", input: "tool IN ('Bash', 'Edit', 'Write')", wantErr: false},
		{name: "NOT IN operator", input: "status NOT IN ('success')", wantErr: false},

		// Range operators
		{name: "BETWEEN operator", input: "duration BETWEEN 500 AND 2000", wantErr: false},

		// Pattern matching
		{name: "LIKE operator", input: "tool LIKE 'meta%'", wantErr: false},
		{name: "REGEXP operator", input: "error REGEXP 'permission.*denied'", wantErr: false},

		// Edge cases
		{name: "whitespace handling", input: "  tool = 'Bash'  ", wantErr: false},
		{name: "empty expression", input: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseExpression(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseExpression() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Test ParseExpression and Evaluate together
func TestParseAndEvaluate(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		record   map[string]interface{}
		expected bool
	}{
		{
			name:     "simple comparison",
			expr:     "tool='Bash'",
			record:   map[string]interface{}{"tool": "Bash"},
			expected: true,
		},
		{
			name:     "AND operator",
			expr:     "tool='Bash' AND status='error'",
			record:   map[string]interface{}{"tool": "Bash", "status": "error"},
			expected: true,
		},
		{
			name:     "OR operator",
			expr:     "tool='Edit' OR tool='Bash'",
			record:   map[string]interface{}{"tool": "Bash"},
			expected: true,
		},
		{
			name:     "NOT operator",
			expr:     "NOT status='success'",
			record:   map[string]interface{}{"status": "error"},
			expected: true,
		},
		{
			name:     "IN operator",
			expr:     "tool IN ('Bash', 'Edit')",
			record:   map[string]interface{}{"tool": "Bash"},
			expected: true,
		},
		{
			name:     "BETWEEN operator",
			expr:     "duration BETWEEN 100 AND 200",
			record:   map[string]interface{}{"duration": 150},
			expected: true,
		},
		{
			name:     "LIKE operator",
			expr:     "tool LIKE 'meta%'",
			record:   map[string]interface{}{"tool": "meta-coach"},
			expected: true,
		},
		{
			name:     "REGEXP operator",
			expr:     "error REGEXP 'permission.*denied'",
			record:   map[string]interface{}{"error": "permission access denied"},
			expected: true,
		},
		{
			name:     "complex nested expression",
			expr:     "(tool='Bash' OR tool='Edit') AND status='error'",
			record:   map[string]interface{}{"tool": "Bash", "status": "error"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := ParseExpression(tt.expr)
			if err != nil {
				t.Fatalf("ParseExpression() error = %v", err)
			}

			result, err := expr.Evaluate(tt.record)
			if err != nil {
				t.Fatalf("Evaluate() error = %v", err)
			}

			if result != tt.expected {
				t.Errorf("Evaluate() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Test error cases
func TestParseExpressionErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "unclosed quote", input: "tool='Bash"},
		{name: "unclosed parenthesis", input: "(tool='Bash'"},
		{name: "invalid operator", input: "tool ~= 'Bash'"},
		{name: "missing value", input: "tool="},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseExpression(tt.input)
			if err == nil {
				t.Errorf("ParseExpression() expected error, got nil")
			}
		})
	}
}
