package filter

import (
	"testing"
)

// Test ComparisonExpression
func TestComparisonExpression(t *testing.T) {
	tests := []struct {
		name     string
		expr     *ComparisonExpression
		record   map[string]interface{}
		expected bool
	}{
		{
			name: "string equality",
			expr: &ComparisonExpression{Field: "tool", Operator: "=", Value: "Bash"},
			record: map[string]interface{}{"tool": "Bash"},
			expected: true,
		},
		{
			name: "string inequality",
			expr: &ComparisonExpression{Field: "tool", Operator: "!=", Value: "Edit"},
			record: map[string]interface{}{"tool": "Bash"},
			expected: true,
		},
		{
			name: "numeric greater than",
			expr: &ComparisonExpression{Field: "duration", Operator: ">", Value: 100},
			record: map[string]interface{}{"duration": 150},
			expected: true,
		},
		{
			name: "numeric less than or equal",
			expr: &ComparisonExpression{Field: "duration", Operator: "<=", Value: 200},
			record: map[string]interface{}{"duration": 200},
			expected: true,
		},
		{
			name: "field missing",
			expr: &ComparisonExpression{Field: "missing", Operator: "=", Value: "test"},
			record: map[string]interface{}{"tool": "Bash"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.expr.Evaluate(tt.record)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// Test BinaryExpression (AND, OR)
func TestBinaryExpression(t *testing.T) {
	tests := []struct {
		name     string
		expr     Expression
		record   map[string]interface{}
		expected bool
	}{
		{
			name: "AND - both true",
			expr: &BinaryExpression{
				Operator: "AND",
				Left:     &ComparisonExpression{Field: "tool", Operator: "=", Value: "Bash"},
				Right:    &ComparisonExpression{Field: "status", Operator: "=", Value: "error"},
			},
			record: map[string]interface{}{"tool": "Bash", "status": "error"},
			expected: true,
		},
		{
			name: "AND - one false",
			expr: &BinaryExpression{
				Operator: "AND",
				Left:     &ComparisonExpression{Field: "tool", Operator: "=", Value: "Bash"},
				Right:    &ComparisonExpression{Field: "status", Operator: "=", Value: "success"},
			},
			record: map[string]interface{}{"tool": "Bash", "status": "error"},
			expected: false,
		},
		{
			name: "OR - one true",
			expr: &BinaryExpression{
				Operator: "OR",
				Left:     &ComparisonExpression{Field: "tool", Operator: "=", Value: "Edit"},
				Right:    &ComparisonExpression{Field: "status", Operator: "=", Value: "error"},
			},
			record: map[string]interface{}{"tool": "Bash", "status": "error"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.expr.Evaluate(tt.record)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// Test UnaryExpression (NOT)
func TestUnaryExpression(t *testing.T) {
	expr := &UnaryExpression{
		Operator: "NOT",
		Operand:  &ComparisonExpression{Field: "status", Operator: "=", Value: "success"},
	}

	record := map[string]interface{}{"status": "error"}
	result, err := expr.Evaluate(record)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result {
		t.Errorf("expected true, got false")
	}
}

// Test InExpression
func TestInExpression(t *testing.T) {
	tests := []struct {
		name     string
		expr     *InExpression
		record   map[string]interface{}
		expected bool
	}{
		{
			name: "IN - value present",
			expr: &InExpression{
				Field:  "tool",
				Values: []interface{}{"Bash", "Edit", "Write"},
				Negate: false,
			},
			record: map[string]interface{}{"tool": "Bash"},
			expected: true,
		},
		{
			name: "IN - value absent",
			expr: &InExpression{
				Field:  "tool",
				Values: []interface{}{"Edit", "Write"},
				Negate: false,
			},
			record: map[string]interface{}{"tool": "Bash"},
			expected: false,
		},
		{
			name: "NOT IN - value absent",
			expr: &InExpression{
				Field:  "tool",
				Values: []interface{}{"Edit", "Write"},
				Negate: true,
			},
			record: map[string]interface{}{"tool": "Bash"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.expr.Evaluate(tt.record)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// Test BetweenExpression
func TestBetweenExpression(t *testing.T) {
	tests := []struct {
		name     string
		expr     *BetweenExpression
		record   map[string]interface{}
		expected bool
	}{
		{
			name: "numeric in range",
			expr: &BetweenExpression{Field: "duration", Lower: 100, Upper: 200},
			record: map[string]interface{}{"duration": 150},
			expected: true,
		},
		{
			name: "numeric out of range",
			expr: &BetweenExpression{Field: "duration", Lower: 100, Upper: 200},
			record: map[string]interface{}{"duration": 250},
			expected: false,
		},
		{
			name: "numeric at boundary",
			expr: &BetweenExpression{Field: "duration", Lower: 100, Upper: 200},
			record: map[string]interface{}{"duration": 200},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.expr.Evaluate(tt.record)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// Test LikeExpression
func TestLikeExpression(t *testing.T) {
	tests := []struct {
		name     string
		expr     *LikeExpression
		record   map[string]interface{}
		expected bool
	}{
		{
			name: "starts with pattern",
			expr: &LikeExpression{Field: "tool", Pattern: "meta%"},
			record: map[string]interface{}{"tool": "meta-coach"},
			expected: true,
		},
		{
			name: "ends with pattern",
			expr: &LikeExpression{Field: "tool", Pattern: "%coach"},
			record: map[string]interface{}{"tool": "meta-coach"},
			expected: true,
		},
		{
			name: "contains pattern",
			expr: &LikeExpression{Field: "tool", Pattern: "%ta-co%"},
			record: map[string]interface{}{"tool": "meta-coach"},
			expected: true,
		},
		{
			name: "no match",
			expr: &LikeExpression{Field: "tool", Pattern: "bash%"},
			record: map[string]interface{}{"tool": "meta-coach"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.expr.Evaluate(tt.record)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// Test RegexpExpression
func TestRegexpExpression(t *testing.T) {
	tests := []struct {
		name     string
		expr     *RegexpExpression
		record   map[string]interface{}
		expected bool
		wantErr  bool
	}{
		{
			name: "simple pattern match",
			expr: &RegexpExpression{Field: "error", Pattern: "permission.*denied"},
			record: map[string]interface{}{"error": "permission access denied"},
			expected: true,
		},
		{
			name: "no match",
			expr: &RegexpExpression{Field: "error", Pattern: "timeout"},
			record: map[string]interface{}{"error": "permission denied"},
			expected: false,
		},
		{
			name: "invalid regexp",
			expr: &RegexpExpression{Field: "error", Pattern: "[invalid("},
			record: map[string]interface{}{"error": "test"},
			expected: false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.expr.Evaluate(tt.record)
			if (err != nil) != tt.wantErr {
				t.Fatalf("wantErr=%v, got err=%v", tt.wantErr, err)
			}
			if !tt.wantErr && result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
