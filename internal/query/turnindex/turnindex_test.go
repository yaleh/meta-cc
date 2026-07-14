package turnindex_test

import (
	"testing"
	"time"

	"github.com/yaleh/meta-cc/internal/parser"
	"github.com/yaleh/meta-cc/internal/query/turnindex"
	"github.com/yaleh/meta-cc/internal/types"
)

func TestBuildTurnIndex(t *testing.T) {
	entries := []types.SessionEntry{
		{UUID: "uuid-1", Type: "user", Message: &parser.Message{Role: "user"}},
		{UUID: "uuid-2", Type: "assistant", Message: &parser.Message{Role: "assistant"}},
		{UUID: "uuid-3", Type: "file-history-snapshot"}, // Not a message
		{UUID: "uuid-4", Type: "user", Message: &parser.Message{Role: "user"}},
	}

	index := turnindex.BuildTurnIndex(entries)

	expected := map[string]int{
		"uuid-1": 0,
		"uuid-2": 1,
		"uuid-4": 2,
	}

	if len(index) != len(expected) {
		t.Errorf("Index length = %d, want %d", len(index), len(expected))
	}

	for uuid, expectedTurn := range expected {
		if turn, ok := index[uuid]; !ok {
			t.Errorf("Missing UUID %s in index", uuid)
		} else if turn != expectedTurn {
			t.Errorf("UUID %s: turn = %d, want %d", uuid, turn, expectedTurn)
		}
	}

	// uuid-3 should not be in the index (not a message)
	if _, exists := index["uuid-3"]; exists {
		t.Error("uuid-3 should not be in index (not a message)")
	}
}

func TestBuildTurnIndex_Empty(t *testing.T) {
	index := turnindex.BuildTurnIndex(nil)
	if len(index) != 0 {
		t.Errorf("Expected empty index for nil entries, got %d", len(index))
	}
}

func TestGetToolCallTimestamp(t *testing.T) {
	now := time.Date(2025, 10, 2, 10, 0, 0, 0, time.UTC)
	ts := now.Format(time.RFC3339Nano)

	entries := []types.SessionEntry{
		{UUID: "uuid-1", Type: "user", Timestamp: ts},
		{UUID: "uuid-2", Type: "assistant", Timestamp: ""},
	}

	got := turnindex.GetToolCallTimestamp(entries, "uuid-1")
	if got != now.Unix() {
		t.Errorf("GetToolCallTimestamp() = %d, want %d", got, now.Unix())
	}

	// Non-existent UUID returns 0
	got = turnindex.GetToolCallTimestamp(entries, "uuid-missing")
	if got != 0 {
		t.Errorf("GetToolCallTimestamp() for missing UUID = %d, want 0", got)
	}

	// Empty timestamp returns 0
	got = turnindex.GetToolCallTimestamp(entries, "uuid-2")
	if got != 0 {
		t.Errorf("GetToolCallTimestamp() for empty timestamp = %d, want 0", got)
	}
}
