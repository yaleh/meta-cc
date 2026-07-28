package records

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"github.com/yaleh/meta-cc/internal/conversation"
	providerpkg "github.com/yaleh/meta-cc/internal/provider"
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
		if turn.UserText != "" {
			emit(map[string]interface{}{
				"type":       "user",
				"provider":   session.Provider,
				"session_id": session.ID,
				"sessionId":  session.ID,
				"cwd":        session.CWD,
				"timestamp":  ts,
				"message": map[string]interface{}{
					"role":    "user",
					"content": turn.UserText,
				},
			}, turnIndex, turn.ID)
		}
		if turn.AssistantText != "" || len(turn.ToolCalls) > 0 || hasUsage(turn.TokenUsage) {
			content := make([]interface{}, 0, len(turn.ToolCalls)+1)
			if turn.AssistantText != "" {
				content = append(content, map[string]interface{}{"type": "text", "text": turn.AssistantText})
			}
			for _, call := range turn.ToolCalls {
				content = append(content, map[string]interface{}{
					"type":  "tool_use",
					"id":    call.ID,
					"name":  call.Name,
					"input": toolInput(call.Input),
				})
			}
			message := map[string]interface{}{
				"role":    "assistant",
				"model":   session.Model,
				"content": content,
			}
			if hasUsage(turn.TokenUsage) {
				message["usage"] = map[string]interface{}{"input_tokens": turn.TokenUsage.InputTokens, "output_tokens": turn.TokenUsage.OutputTokens, "cache_tokens": turn.TokenUsage.CacheTokens}
			} else if session.Provider != conversation.ProviderCodex && hasUsage(session.TokenUsage) {
				message["usage"] = map[string]interface{}{"input_tokens": session.TokenUsage.InputTokens, "output_tokens": session.TokenUsage.OutputTokens, "cache_tokens": session.TokenUsage.CacheTokens}
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
		}
		var toolResults []interface{}
		for _, call := range turn.ToolCalls {
			if call.Output == "" && !call.IsError {
				continue
			}
			entry := map[string]interface{}{
				"type":        "tool_result",
				"tool_use_id": call.ID,
				"content":     call.Output,
			}
			if call.IsError {
				entry["is_error"] = true
				entry["status"] = "error"
				entry["error"] = call.Output
			} else {
				entry["is_error"] = false
				entry["status"] = "success"
			}
			toolResults = append(toolResults, entry)
		}
		if len(toolResults) > 0 {
			emit(map[string]interface{}{
				"type":       "user",
				"provider":   session.Provider,
				"session_id": session.ID,
				"sessionId":  session.ID,
				"cwd":        session.CWD,
				"timestamp":  ts,
				"message": map[string]interface{}{
					"role":    "user",
					"content": toolResults,
				},
			}, turnIndex, turn.ID)
		}
	}
	return out
}

func hasUsage(usage conversation.TokenUsage) bool {
	return usage.InputTokens != 0 || usage.OutputTokens != 0 || usage.CacheTokens != 0
}

func toolInput(raw json.RawMessage) map[string]interface{} {
	input := map[string]interface{}{}
	if len(raw) == 0 {
		return input
	}
	if err := json.Unmarshal(raw, &input); err != nil || input == nil {
		return map[string]interface{}{}
	}
	return input
}
