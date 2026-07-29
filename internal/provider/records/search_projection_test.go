package records

import (
	"encoding/json"
	"testing"

	"github.com/yaleh/meta-cc/internal/conversation"
	"github.com/yaleh/meta-cc/internal/provider/projection"
)

func TestNormalizeMessageContentMatchesCanonicalSearchProjection(t *testing.T) {
	turn := conversation.Turn{
		ID:            "turn",
		UserText:      "hello",
		AssistantText: "answer",
		ToolCalls: []conversation.ToolCall{{
			ID: "call-1", Name: "lookup", Input: json.RawMessage(`{"city":"Zürich"}`), Output: "failed", IsError: true,
		}},
	}
	session := conversation.Session{ID: "session", Provider: conversation.ProviderCodex, Model: "model"}

	records := Normalize(session, []conversation.Turn{turn})
	got := make([]string, 0, len(records))
	for _, record := range records {
		message := record["message"].(map[string]interface{})
		text, err := projection.Tostring(message["content"])
		if err != nil {
			t.Fatal(err)
		}
		got = append(got, text)
	}
	want, err := projection.SearchStrings(turn)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(want) {
		t.Fatalf("Normalize emitted %d contents, projection emitted %d: %#v vs %#v", len(got), len(want), got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("content %d drifted:\nNormalize: %q\nprojection: %q", i, got[i], want[i])
		}
	}
}
