package conversation

import "testing"

// TestChildrenOf proves the DIR-032 lineage helper returns only sessions
// whose ParentThreadID matches, preserves order, and never touches
// project/cwd boundaries itself (that's the caller's job — see
// internal/mcp/executor/query_sessions_handler.go).
func TestChildrenOf(t *testing.T) {
	sessions := []Session{
		{ID: "root", ParentThreadID: ""},
		{ID: "child-1", ParentThreadID: "root"},
		{ID: "unrelated", ParentThreadID: "other"},
		{ID: "child-2", ParentThreadID: "root"},
	}

	got := ChildrenOf(sessions, "root")
	if len(got) != 2 || got[0].ID != "child-1" || got[1].ID != "child-2" {
		t.Fatalf("unexpected children: %#v", got)
	}

	if got := ChildrenOf(sessions, ""); got != nil {
		t.Fatalf("empty parentID should return nil, got %#v", got)
	}

	if got := ChildrenOf(sessions, "no-such-parent"); len(got) != 0 {
		t.Fatalf("expected no children, got %#v", got)
	}
}
