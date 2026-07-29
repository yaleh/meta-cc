package records

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/yaleh/meta-cc/internal/conversation"
	providerpkg "github.com/yaleh/meta-cc/internal/provider"
	"github.com/yaleh/meta-cc/internal/provider/projection"
)

func Build(ctx context.Context, registry *providerpkg.Registry, filters []conversation.ProviderID, scope, projectPath string) ([]map[string]interface{}, []string, error) {
	var (
		out      []map[string]interface{}
		warnings []string
	)

	for _, p := range registry.Providers(filters) {
		if !p.IsAvailable(ctx) {
			warnings = append(warnings, fmt.Sprintf("provider %s unavailable", p.ID()))
			continue
		}
		sessions, err := p.ListSessions(ctx)
		if err != nil {
			return nil, warnings, err
		}
		sessions = FilterSessionsForScope(sessions, scope, projectPath, p.ID())
		for _, session := range sessions {
			turns, err := p.LoadTurns(ctx, session.ID)
			if err != nil {
				// DIR-030: one corrupt/unreadable session must not erase
				// valid results from every other session a project query
				// would otherwise return — record a per-session warning
				// and keep going, rather than aborting the whole query.
				warnings = append(warnings, fmt.Sprintf("provider %s session %s: failed to load turns, skipped: %v", p.ID(), session.ID, err))
				continue
			}
			out = append(out, Normalize(session, turns)...)
		}
	}
	return out, warnings, nil
}

// BuildForSession is Build narrowed to exactly one session ID: rather than
// listing every session for the requested provider(s) and iterating them,
// it calls GetSession/LoadTurns directly for sessionID on each requested
// provider — satisfying the DIR-030 "exact session lookup reads only the
// requested thread" contract. When filters names more than one provider
// (e.g. "all"), a session not found in a given provider (or found but
// outside projectPath's cwd boundary) is recorded as a warning and that
// provider is skipped, rather than aborting the whole call — mirroring
// Build's per-session tolerance. An error is returned only when the
// session was not found (within bounds) in ANY requested provider.
func BuildForSession(ctx context.Context, registry *providerpkg.Registry, filters []conversation.ProviderID, sessionID, projectPath string) ([]map[string]interface{}, []string, error) {
	var (
		out      []map[string]interface{}
		warnings []string
		found    bool
	)

	for _, p := range registry.Providers(filters) {
		if !p.IsAvailable(ctx) {
			warnings = append(warnings, fmt.Sprintf("provider %s unavailable", p.ID()))
			continue
		}
		session, err := p.GetSession(ctx, sessionID)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("provider %s: session %s not found: %v", p.ID(), sessionID, err))
			continue
		}
		// Mirror FilterSessionsForScope's cwd-boundary check: a session_id
		// lookup must never cross project boundaries, e.g. by returning a
		// same-named session that actually belongs to a different project.
		// Unlike FilterSessionsForScope (which only needs this for Codex,
		// since Claude's ListSessions is already project-scoped by
		// construction via locator.AllSessionsFromProject), GetSession's
		// underlying lookup for EVERY provider can resolve outside
		// projectPath (Claude's findSessionFile falls back to a global,
		// unscoped locator.FromSessionID search), so this check applies to
		// all providers, not just Codex.
		if session.CWD != "" && projectPath != "" && session.CWD != projectPath {
			warnings = append(warnings, fmt.Sprintf("provider %s: session %s belongs to a different project (%s), excluded", p.ID(), sessionID, session.CWD))
			continue
		}
		turns, err := p.LoadTurns(ctx, sessionID)
		if err != nil {
			return nil, warnings, fmt.Errorf("provider %s: failed to load turns for session %s: %w", p.ID(), sessionID, err)
		}
		out = append(out, Normalize(session, turns)...)
		found = true
	}
	if !found {
		return nil, warnings, fmt.Errorf("session %q not found for the requested provider(s) within project %q", sessionID, projectPath)
	}
	return out, warnings, nil
}

func FilterSessionsForScope(sessions []conversation.Session, scope, projectPath string, providerID conversation.ProviderID) []conversation.Session {
	filtered := make([]conversation.Session, 0, len(sessions))
	for _, session := range sessions {
		// Apply the project/CWD filter for BOTH "project" and "session"
		// scope: a session-scope lookup must never cross project boundaries
		// and return another project's most-recent session just because it
		// sorts first. When projectPath or session.CWD is unset (e.g. no
		// working_dir supplied), the filter is a no-op and every session
		// stays eligible for the subsequent most-recent-first sort.
		if providerID == conversation.ProviderCodex && session.CWD != "" && projectPath != "" && session.CWD != projectPath {
			continue
		}
		filtered = append(filtered, session)
	}
	if scope != "session" || len(filtered) <= 1 {
		return filtered
	}
	slices.SortFunc(filtered, func(a, b conversation.Session) int {
		if a.CreatedAt.Before(b.CreatedAt) {
			return 1
		}
		if a.CreatedAt.After(b.CreatedAt) {
			return -1
		}
		return 0
	})
	return filtered[:1]
}

// Normalize converts a provider's ordered Turn stream into the common,
// jq-queryable record schema. Beyond the pre-existing fields, every emitted
// record also carries a canonical, provider-neutral identity (DIR-036):
//
//   - turn_id:    the originating Turn's stable ID (conversation.Turn.ID).
//   - turn_index: the Turn's 0-based position in the turns slice passed in.
//   - seq:        a 0-based, strictly increasing position of THIS record
//     within this call's output (one session's full normalized record
//     stream). Combined with session_id, (session_id, seq) is a stable
//     identity that never relies on timestamp uniqueness (multiple records
//     from one turn legitimately share a timestamp) and survives jq
//     projections that only select/filter (rather than reshape) records.
//
// This identity/order is what lets context-window expansion (see
// internal/mcp/filters.ExpandContextTurnsCanonical) locate a matched record
// inside a freshly reloaded canonical session stream without depending on
// Claude's uuid field, which normalized records never carry.
func Normalize(session conversation.Session, turns []conversation.Turn) []map[string]interface{} {
	var out []map[string]interface{}
	seq := 0
	emit := func(rec map[string]interface{}, turnIndex int, turnID string) {
		rec["turn_id"] = turnID
		rec["turn_index"] = turnIndex
		rec["seq"] = seq
		seq++
		out = append(out, rec)
	}
	for turnIndex, turn := range turns {
		ts := turn.Timestamp.Format(time.RFC3339)
		// turnStartSeq snapshots the output position before this turn emits
		// anything, so lifecycle-only turns can still produce a record below.
		turnStartSeq := seq
		contents := projection.Contents(turn)
		for _, projected := range contents {
			switch projected.Kind {
			case projection.UserMessage:
				emit(map[string]interface{}{
					"type":       "user",
					"provider":   session.Provider,
					"session_id": session.ID,
					"sessionId":  session.ID,
					"cwd":        session.CWD,
					"timestamp":  ts,
					"message": map[string]interface{}{
						"role":    "user",
						"content": projected.Value,
					},
				}, turnIndex, turn.ID)
			case projection.AssistantMessage:
				message := map[string]interface{}{
					"role":    "assistant",
					"model":   session.Model,
					"content": projected.Value,
				}
				if turn.TokenUsage.HasAny() {
					message["usage"] = usageMap(turn.TokenUsage)
				} else if session.Provider != conversation.ProviderCodex && session.TokenUsage.HasAny() {
					message["usage"] = usageMap(session.TokenUsage)
				}
				emit(map[string]interface{}{
					"type":       "assistant",
					"provider":   session.Provider,
					"session_id": session.ID,
					"sessionId":  session.ID,
					"cwd":        session.CWD,
					"timestamp":  ts,
					"message":    message,
				}, turnIndex, turn.ID)
			case projection.ToolResults:
				emit(map[string]interface{}{
					"type":       "user",
					"provider":   session.Provider,
					"session_id": session.ID,
					"sessionId":  session.ID,
					"cwd":        session.CWD,
					"timestamp":  ts,
					"message": map[string]interface{}{
						"role":    "user",
						"content": projected.Value,
					},
				}, turnIndex, turn.ID)
			}
		}
		// DIR-053: a turn whose only content is a typed lifecycle signal
		// (ItemKindSessionEnd / ItemKindCompaction) or a non-default Status
		// (e.g. TurnStatusAborted) produces none of the user/assistant/
		// tool_result records above, so it would remain invisible to every MCP
		// tool even though DIR-050's flush() now retains it as a Turn. When the
		// turn emitted nothing else (seq unchanged), emit a minimal lifecycle
		// record so the turn is observable end-to-end. Turns that already
		// surface content are left untouched (no extra record) to avoid
		// changing record counts for ordinary sessions.
		if seq == turnStartSeq {
			for _, rec := range lifecycleRecords(session, turn, ts) {
				emit(rec, turnIndex, turn.ID)
			}
		}
	}
	return out
}

// lifecycleRecords builds the minimal normalized record(s) for a turn that
// carries a lifecycle signal but no user/assistant/tool content (DIR-053). It
// mirrors the widening DIR-050 applied to flush()'s retention condition, but at
// the normalized-record layer: one record per ItemKindSessionEnd /
// ItemKindCompaction item, plus — when no such item is present — one record for
// a non-default Turn.Status (Codex's "turn_aborted" sets a status with no
// accompanying item). Each record reuses the existing record vocabulary: the
// standard identity fields (type/provider/session_id/sessionId/cwd/timestamp,
// plus turn_id/turn_index/seq added by emit) and a distinguishable type value
// drawn from the ItemKind/TurnStatus vocabulary ("session_end", "compaction",
// "turn_aborted", ...), so a jq filter such as select(.type == "session_end")
// or a role="all" conversation-flow query can surface it. No large new schema
// is introduced — just the signal-specific fields each kind already carries
// (reason for session_end, the CompactionBoundary for compaction).
func lifecycleRecords(session conversation.Session, turn conversation.Turn, ts string) []map[string]interface{} {
	base := func(typ string) map[string]interface{} {
		rec := map[string]interface{}{
			"type":       typ,
			"provider":   session.Provider,
			"session_id": session.ID,
			"sessionId":  session.ID,
			"cwd":        session.CWD,
			"timestamp":  ts,
		}
		if turn.Status != conversation.TurnStatusUnspecified {
			rec["turn_status"] = string(turn.Status)
		}
		return rec
	}

	var recs []map[string]interface{}
	itemEmitted := false
	for _, item := range turn.Items {
		switch item.Kind {
		case conversation.ItemKindSessionEnd:
			rec := base("session_end")
			if item.Text != "" {
				rec["reason"] = item.Text
			}
			recs = append(recs, rec)
			itemEmitted = true
		case conversation.ItemKindCompaction:
			rec := base("compaction")
			if item.Compaction != nil {
				boundary := map[string]interface{}{}
				if item.Compaction.Reason != "" {
					boundary["reason"] = item.Compaction.Reason
				}
				if item.Compaction.Summary != "" {
					boundary["summary"] = item.Compaction.Summary
				}
				if len(boundary) > 0 {
					rec["compaction"] = boundary
				}
			}
			recs = append(recs, rec)
			itemEmitted = true
		}
	}
	// Status-only terminal/failure signals (e.g. turn_aborted): emit a record
	// only when no lifecycle item already produced one. InProgress is parser
	// bookkeeping (not a lifecycle event), so it deliberately emits nothing.
	if !itemEmitted {
		if typ, ok := lifecycleTypeForStatus(turn.Status); ok {
			recs = append(recs, base(typ))
		}
	}
	return recs
}

// lifecycleTypeForStatus maps only explicit terminal/failure statuses. A bare
// InProgress status is not a genuine lifecycle event and must remain invisible.
func lifecycleTypeForStatus(status conversation.TurnStatus) (string, bool) {
	switch status {
	case conversation.TurnStatusAborted:
		return "turn_aborted", true
	case conversation.TurnStatusFailed:
		return "turn_failed", true
	case conversation.TurnStatusCompleted:
		return "turn_completed", true
	default:
		return "", false
	}
}

// usageMap renders a canonical TokenUsage as the normalized message.usage
// object. input_tokens / output_tokens / cache_tokens are always present
// (unchanged legacy contract); reasoning_output_tokens is added only when the
// source reported it (DIR-071), so existing records without reasoning gain no
// new zero-valued field. An opaque AggregateTokens is deliberately NOT mapped
// here: it is session-level metadata with its own provenance and must never
// be re-emitted as a per-turn input/output count.
func usageMap(usage conversation.TokenUsage) map[string]interface{} {
	m := map[string]interface{}{
		"input_tokens":  usage.InputTokens,
		"output_tokens": usage.OutputTokens,
		"cache_tokens":  usage.CacheTokens,
	}
	if usage.ReasoningOutputTokens != 0 {
		m["reasoning_output_tokens"] = usage.ReasoningOutputTokens
	}
	return m
}
