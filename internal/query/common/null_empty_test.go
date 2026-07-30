package common

import (
	"errors"
	"testing"
)

func TestAllNullOrEmpty(t *testing.T) {
	tests := []struct {
		name    string
		entries []interface{}
		want    bool
	}{
		{name: "empty slice", entries: []interface{}{}, want: false},
		{name: "nil scalar", entries: []interface{}{nil}, want: true},
		{name: "all null map", entries: []interface{}{map[string]interface{}{"a": nil, "b": nil}}, want: true},
		{name: "all empty map", entries: []interface{}{map[string]interface{}{"a": "", "b": ""}}, want: true},
		{name: "mixed empty entries", entries: []interface{}{nil, map[string]interface{}{"a": "", "b": nil}}, want: true},
		{name: "non-empty map", entries: []interface{}{map[string]interface{}{"a": nil, "b": "value"}}, want: false},
		{name: "non-map scalar", entries: []interface{}{"value"}, want: false},
		{name: "jq error value", entries: []interface{}{errors.New("jq execution error")}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AllNullOrEmpty(tt.entries); got != tt.want {
				t.Fatalf("AllNullOrEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}
