package filters

import (
	"fmt"
	"testing"
)

// mkRec builds a minimal provider-normalized record with the (session_id,
// seq) canonical identity DIR-036 introduced.
func mkRec(sessionID string, seq int, extra map[string]interface{}) map[string]interface{} {
	rec := map[string]interface{}{
		"session_id": sessionID,
		"sessionId":  sessionID,
		"seq":        seq,
	}
	for k, v := range extra {
		rec[k] = v
	}
	return rec
}

// staticLoader returns a SessionRecordLoader backed by an in-memory session->records map.
func staticLoader(sessions map[string][]interface{}) SessionRecordLoader {
	return func(sessionID string) ([]interface{}, error) {
		recs, ok := sessions[sessionID]
		if !ok {
			return nil, fmt.Errorf("session %q not found", sessionID)
		}
		return recs, nil
	}
}

func TestExpandContextTurnsCanonical_EmptyInput(t *testing.T) {
	result, warnings, err := ExpandContextTurnsCanonical([]interface{}{}, 2, staticLoader(nil), false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings, got %v", warnings)
	}
	if len(result) != 0 {
		t.Fatalf("expected empty result, got %d", len(result))
	}
}

func TestExpandContextTurnsCanonical_ZeroN(t *testing.T) {
	rawData := []interface{}{mkRec("s1", 2, nil)}
	result, _, err := ExpandContextTurnsCanonical(rawData, 0, staticLoader(nil), false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected passthrough of 1 record, got %d", len(result))
	}
}

func TestExpandContextTurnsCanonical_Basic(t *testing.T) {
	sessionID := "sess-1"
	var full []interface{}
	for i := 0; i < 5; i++ {
		full = append(full, mkRec(sessionID, i, nil))
	}
	rawData := []interface{}{mkRec(sessionID, 2, nil)} // matched: seq=2

	result, warnings, err := ExpandContextTurnsCanonical(rawData, 1, staticLoader(map[string][]interface{}{sessionID: full}), false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings, got %v", warnings)
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 results (seq 1,2,3), got %d: %#v", len(result), result)
	}
	for _, entry := range result {
		obj := entry.(map[string]interface{})
		seq := obj["seq"].(int)
		ctx := obj["context"].(bool)
		if seq == 2 && ctx != false {
			t.Errorf("matched seq=2 should have context=false, got %v", ctx)
		}
		if (seq == 1 || seq == 3) && ctx != true {
			t.Errorf("context seq=%d should have context=true, got %v", seq, ctx)
		}
	}
}

func TestExpandContextTurnsCanonical_WindowClampAtStartAndEnd(t *testing.T) {
	sessionID := "sess-clamp"
	var full []interface{}
	for i := 0; i < 3; i++ {
		full = append(full, mkRec(sessionID, i, nil))
	}
	loader := staticLoader(map[string][]interface{}{sessionID: full})

	// Match first record: window should clamp to [0,1].
	result, _, err := ExpandContextTurnsCanonical([]interface{}{mkRec(sessionID, 0, nil)}, 2, loader, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("expected all 3 (clamped at start), got %d", len(result))
	}

	// Match last record: window should clamp to [1,2].
	result, _, err = ExpandContextTurnsCanonical([]interface{}{mkRec(sessionID, 2, nil)}, 2, loader, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("expected all 3 (clamped at end), got %d", len(result))
	}
}

func TestExpandContextTurnsCanonical_OverlappingWindowsDeduplicated(t *testing.T) {
	sessionID := "sess-overlap"
	var full []interface{}
	for i := 0; i < 5; i++ {
		full = append(full, mkRec(sessionID, i, nil))
	}
	rawData := []interface{}{mkRec(sessionID, 1, nil), mkRec(sessionID, 3, nil)}

	result, _, err := ExpandContextTurnsCanonical(rawData, 1, staticLoader(map[string][]interface{}{sessionID: full}), false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 5 {
		t.Fatalf("expected 5 (no duplicates), got %d", len(result))
	}
	seen := map[int]int{}
	for _, entry := range result {
		obj := entry.(map[string]interface{})
		seen[obj["seq"].(int)]++
	}
	for seq, count := range seen {
		if count != 1 {
			t.Errorf("seq %d appeared %d times, expected 1", seq, count)
		}
	}
}

func TestExpandContextTurnsCanonical_NeverCrossesSessionBoundary(t *testing.T) {
	sessA, sessB := "sess-a", "sess-b"
	var fullA, fullB []interface{}
	for i := 0; i < 3; i++ {
		fullA = append(fullA, mkRec(sessA, i, nil))
	}
	for i := 0; i < 2; i++ {
		fullB = append(fullB, mkRec(sessB, i, nil))
	}
	rawData := []interface{}{mkRec(sessA, 1, nil), mkRec(sessB, 0, nil)}

	result, _, err := ExpandContextTurnsCanonical(rawData, 1, staticLoader(map[string][]interface{}{sessA: fullA, sessB: fullB}), false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// sessA: seq1 matched -> window [0,1,2] = 3
	// sessB: seq0 matched -> window [0,1] (clamped) = 2
	if len(result) != 5 {
		t.Fatalf("expected 5 results, got %d: %#v", len(result), result)
	}
	for _, entry := range result {
		obj := entry.(map[string]interface{})
		sid := obj["session_id"].(string)
		if sid != sessA && sid != sessB {
			t.Fatalf("unexpected session_id %q", sid)
		}
	}
}

func TestExpandContextTurnsCanonical_ExcludesCompactSummaries(t *testing.T) {
	sessionID := "sess-compact"
	full := []interface{}{
		mkRec(sessionID, 0, nil),
		mkRec(sessionID, 1, map[string]interface{}{"isCompactSummary": true}),
		mkRec(sessionID, 2, nil),
		mkRec(sessionID, 3, nil),
		mkRec(sessionID, 4, nil),
	}
	rawData := []interface{}{mkRec(sessionID, 2, nil)}

	result, _, err := ExpandContextTurnsCanonical(rawData, 2, staticLoader(map[string][]interface{}{sessionID: full}), true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, entry := range result {
		obj := entry.(map[string]interface{})
		if isCompact, _ := obj["isCompactSummary"].(bool); isCompact {
			t.Errorf("result must not contain compact summary entries: %#v", obj)
		}
	}
}

// TestExpandContextTurnsCanonical_LoadFailureRetainsMatches is the DIR-036
// contract check: when loadSession fails for a session that had real
// matches, those matches must be retained (context:false) and a warning
// returned — never a silent, unqualified empty result for that session.
func TestExpandContextTurnsCanonical_LoadFailureRetainsMatches(t *testing.T) {
	sessionID := "sess-broken"
	rawData := []interface{}{mkRec(sessionID, 5, map[string]interface{}{"marker": "keep-me"})}

	loader := func(string) ([]interface{}, error) {
		return nil, fmt.Errorf("simulated backend failure")
	}

	result, warnings, err := ExpandContextTurnsCanonical(rawData, 2, loader, false)
	if err != nil {
		t.Fatalf("expected no hard error (failure should be reported as a warning), got: %v", err)
	}
	if len(warnings) == 0 {
		t.Fatal("expected a warning describing the load failure")
	}
	if len(result) != 1 {
		t.Fatalf("expected the original match to be retained, got %d results: %#v", len(result), result)
	}
	obj := result[0].(map[string]interface{})
	if obj["marker"] != "keep-me" {
		t.Errorf("retained record lost its original fields: %#v", obj)
	}
	if ctx, ok := obj["context"].(bool); !ok || ctx != false {
		t.Errorf("retained match should have context=false, got %#v", obj["context"])
	}
}

// TestExpandContextTurnsCanonical_AcceptsFloat64Seq verifies that a "seq"
// value round-tripped through JSON/jq as float64 is still recognized (see
// runProviderJQ in internal/mcp/executor).
func TestExpandContextTurnsCanonical_AcceptsFloat64Seq(t *testing.T) {
	sessionID := "sess-float"
	var full []interface{}
	for i := 0; i < 3; i++ {
		full = append(full, mkRec(sessionID, i, nil))
	}
	rawData := []interface{}{
		map[string]interface{}{"session_id": sessionID, "sessionId": sessionID, "seq": float64(1)},
	}

	result, _, err := ExpandContextTurnsCanonical(rawData, 1, staticLoader(map[string][]interface{}{sessionID: full}), false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 results, got %d", len(result))
	}
}
