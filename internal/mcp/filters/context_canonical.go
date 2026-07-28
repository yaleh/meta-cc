package filters

import (
	"fmt"
)

// SessionRecordLoader loads the full, ordered, canonical (provider-normalized)
// record stream for one session ID, in the same projection the original query
// used to produce rawData. Implementations are expected to route through the
// provider/session abstraction (e.g. providerrecords.BuildForSession) rather
// than scanning a provider-blind directory, so every supported provider and
// backend (Codex rollout, Codex app-server, Claude) is covered uniformly.
type SessionRecordLoader func(sessionID string) ([]interface{}, error)

// canonicalKey identifies one normalized record within one session using the
// (session_id, seq) pair emitted by internal/provider/records.Normalize
// (DIR-036) rather than a Claude-only uuid. seq is a 0-based position in the
// session's normalized record stream, assigned once per Normalize() call, so
// it is stable across reloads of the same session and never depends on
// timestamp uniqueness.
type canonicalKey struct {
	sessionID string
	seq       int
}

// recordSessionID extracts a record's session ID, accepting both the
// camelCase ("sessionId") and snake_case ("session_id") forms used across the
// codebase.
func recordSessionID(obj map[string]interface{}) string {
	if sid, ok := obj["sessionId"].(string); ok && sid != "" {
		return sid
	}
	if sid, ok := obj["session_id"].(string); ok && sid != "" {
		return sid
	}
	return ""
}

// recordSeq extracts a record's canonical "seq" field. jq/gojq processing
// (internal/mcp/executor's runProviderJQ) can hand values back as either Go
// int (untouched passthrough) or float64 (post round-trip), so both are
// accepted. ok is false when no usable seq is present.
func recordSeq(obj map[string]interface{}) (int, bool) {
	switch v := obj["seq"].(type) {
	case int:
		return v, true
	case int64:
		return int(v), true
	case float64:
		return int(v), true
	default:
		return 0, false
	}
}

// ExpandContextTurnsCanonical is the provider-neutral counterpart to
// ExpandContextTurns (DIR-036). Rather than assuming a Claude uuid and
// rescanning a Claude JSONL project directory, it:
//
//  1. Groups rawData's already-matched records by session ID.
//  2. For each distinct session, reloads that session's full canonical
//     record stream via loadSession (routed through the provider/session
//     abstraction by the caller).
//  3. Locates each matched record inside that reloaded stream by
//     (session_id, seq) identity, expands a bounded +/-N window, merges
//     overlapping windows without duplicates, and marks matches
//     context:false / added context context:true.
//
// Contract (DIR-036): a session that loadSession cannot supply context for
// must never cause its already-matched records to vanish. Those matches are
// retained verbatim (context:false, no added context) and a warning
// describing the failure is returned alongside the data — never a silently
// empty result for a session that had real matches.
func ExpandContextTurnsCanonical(rawData []interface{}, n int, loadSession SessionRecordLoader, excludeCompactSummaries bool) ([]interface{}, []string, error) {
	if n <= 0 || len(rawData) == 0 {
		return rawData, nil, nil
	}

	// 1. Group matched records by session, preserving first-seen session order.
	var sessionOrder []string
	sessionSeen := make(map[string]bool)
	matchedKeys := make(map[canonicalKey]bool)
	// Records that cannot be matched to a (session_id, seq) identity at all
	// (malformed input) are retained verbatim rather than dropped.
	var unidentifiable []interface{}

	for _, entry := range rawData {
		obj, ok := entry.(map[string]interface{})
		if !ok {
			unidentifiable = append(unidentifiable, entry)
			continue
		}
		sid := recordSessionID(obj)
		seq, hasSeq := recordSeq(obj)
		if sid == "" || !hasSeq {
			unidentifiable = append(unidentifiable, entry)
			continue
		}
		if !sessionSeen[sid] {
			sessionOrder = append(sessionOrder, sid)
			sessionSeen[sid] = true
		}
		matchedKeys[canonicalKey{sid, seq}] = true
	}

	var warnings []string
	seenKeys := make(map[canonicalKey]bool)
	var result []interface{}

	for _, sid := range sessionOrder {
		turns, err := loadSession(sid)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf(
				"context_turns: failed to load session %q for context expansion, original matches retained without added context: %v",
				sid, err))
			// Retain the original matched records for this session verbatim,
			// stamped context:false, rather than losing them.
			for _, entry := range rawData {
				obj, ok := entry.(map[string]interface{})
				if !ok {
					continue
				}
				if recordSessionID(obj) != sid {
					continue
				}
				seq, hasSeq := recordSeq(obj)
				if !hasSeq {
					continue
				}
				k := canonicalKey{sid, seq}
				if seenKeys[k] {
					continue
				}
				seenKeys[k] = true
				newObj := make(map[string]interface{}, len(obj)+1)
				for kk, vv := range obj {
					newObj[kk] = vv
				}
				newObj["context"] = false
				result = append(result, newObj)
			}
			continue
		}

		// Filter compact summaries (Claude-shaped records reachable via
		// provider=all) before building the seq->index map, mirroring
		// ExpandContextTurns's semantics.
		filtered := make([]map[string]interface{}, 0, len(turns))
		for _, t := range turns {
			m, ok := t.(map[string]interface{})
			if !ok {
				continue
			}
			if excludeCompactSummaries {
				if isCompact, _ := m["isCompactSummary"].(bool); isCompact {
					continue
				}
			}
			filtered = append(filtered, m)
		}

		seqToIdx := make(map[int]int, len(filtered))
		for i, m := range filtered {
			if seq, ok := recordSeq(m); ok {
				seqToIdx[seq] = i
			}
		}

		windowSet := make(map[int]bool)
		for _, entry := range rawData {
			obj, ok := entry.(map[string]interface{})
			if !ok {
				continue
			}
			if recordSessionID(obj) != sid {
				continue
			}
			seq, hasSeq := recordSeq(obj)
			if !hasSeq {
				continue
			}
			idx, exists := seqToIdx[seq]
			if !exists {
				// The matched record's identity isn't present in the reloaded
				// stream (e.g. it was excluded as a compact summary, or the
				// reloaded session no longer contains it). Retain the match
				// itself rather than silently dropping it.
				k := canonicalKey{sid, seq}
				if !seenKeys[k] {
					seenKeys[k] = true
					newObj := make(map[string]interface{}, len(obj)+1)
					for kk, vv := range obj {
						newObj[kk] = vv
					}
					newObj["context"] = false
					result = append(result, newObj)
				}
				continue
			}
			lo := idx - n
			if lo < 0 {
				lo = 0
			}
			hi := idx + n
			if hi >= len(filtered) {
				hi = len(filtered) - 1
			}
			for i := lo; i <= hi; i++ {
				windowSet[i] = true
			}
		}

		for i := 0; i < len(filtered); i++ {
			if !windowSet[i] {
				continue
			}
			m := filtered[i]
			seq, hasSeq := recordSeq(m)
			if !hasSeq {
				continue
			}
			k := canonicalKey{sid, seq}
			if seenKeys[k] {
				continue
			}
			seenKeys[k] = true

			newObj := make(map[string]interface{}, len(m)+1)
			for kk, vv := range m {
				newObj[kk] = vv
			}
			newObj["context"] = !matchedKeys[k]
			result = append(result, newObj)
		}
	}

	result = append(result, unidentifiable...)

	return result, warnings, nil
}
