package projection

import (
	"encoding/json"
	"testing"

	"github.com/yaleh/meta-cc/internal/conversation"
)

func TestSearchStringsIncludesCanonicalStructureAndCrossBlockSerialization(t *testing.T) {
	turn := conversation.Turn{
		AssistantText: "intro",
		ToolCalls: []conversation.ToolCall{{
			ID: "call-1", Name: "lookup", Input: json.RawMessage(`{"city":"Zürich"}`), Output: "boom", IsError: true,
		}},
	}
	got, err := SearchStrings(turn)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("got %#v", got)
	}
	for _, literal := range []string{`"type":"tool_use"`, `"id":"call-1"`, `"name":"lookup"`, `Zürich`, `"type":"text"},{"id":"call-1"`} {
		if !contains(got[0], literal) {
			t.Errorf("assistant projection %q missing %q", got[0], literal)
		}
	}
	for _, literal := range []string{`"type":"tool_result"`, `"tool_use_id":"call-1"`, `"status":"error"`, `"is_error":true`} {
		if !contains(got[1], literal) {
			t.Errorf("result projection %q missing %q", got[1], literal)
		}
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
