package records

import (
	"context"
	"fmt"

	"github.com/yaleh/meta-cc/internal/conversation"
	"github.com/yaleh/meta-cc/internal/ftsindex"
	providerpkg "github.com/yaleh/meta-cc/internal/provider"
)

// BuildIndexedContent incrementally refreshes the project-local index and
// returns canonical records from candidate sessions only. used=false is a
// fail-open signal: the caller must perform the authoritative direct scan.
// Candidate bodies are privacy-bounded and matched in Go with the same
// Unicode-aware case-insensitive regexp shape used by canonical gojq test().
func BuildIndexedContent(ctx context.Context, registry *providerpkg.Registry, filters []conversation.ProviderID, scope, projectPath, literal string) (records []map[string]interface{}, warnings []string, used bool) {
	if scope != "project" || projectPath == "" || ftsindex.IsDisabled() || len([]rune(literal)) < 3 {
		return nil, nil, false
	}
	providers := registry.Providers(filters)
	byID := make(map[conversation.ProviderID]providerpkg.Provider, len(providers))
	var sessions []conversation.Session
	for _, p := range providers {
		if !p.IsAvailable(ctx) {
			warnings = append(warnings, fmt.Sprintf("provider %s unavailable", p.ID()))
			continue
		}
		listed, err := p.ListSessions(ctx)
		if err != nil {
			return nil, append(warnings, "ftsindex: provider listing failed; using direct scan"), false
		}
		listed = FilterSessionsForScope(listed, scope, projectPath, p.ID())
		sessions = append(sessions, listed...)
		byID[p.ID()] = p
	}
	db, degraded, err := ftsindex.Open(ctx, ftsindex.DefaultPath(projectPath))
	if err != nil {
		return nil, append(warnings, "ftsindex: unavailable; using direct scan"), false
	}
	defer db.Close()
	load := func(ctx context.Context, session conversation.Session) ([]conversation.Turn, error) {
		return byID[session.Provider].LoadTurns(ctx, session.ID)
	}
	stats, refreshWarnings, err := ftsindex.Refresh(ctx, db, sessions, ftsindex.DefaultSourceMeta, load, ftsindex.DefaultBodyLimitBytes)
	warnings = append(warnings, refreshWarnings...)
	if err != nil || stats.SessionsFailed > 0 || degraded {
		return nil, append(warnings, "ftsindex: refresh incomplete or recovered; using direct scan"), false
	}
	for providerID := range byID {
		live := map[string]bool{}
		for _, session := range sessions {
			if session.Provider == providerID {
				live[string(providerID)+":"+session.ID] = true
			}
		}
		if _, err := ftsindex.Reconcile(ctx, db, providerID, projectPath, live); err != nil {
			return nil, append(warnings, "ftsindex: reconciliation failed; using direct scan"), false
		}
	}
	candidates, err := ftsindex.LiteralSessionCandidates(ctx, db, literal, projectPath, filters)
	if err != nil {
		return nil, append(warnings, "ftsindex: candidate search incomplete; using direct scan"), false
	}
	candidateIDs := map[string]bool{}
	for _, candidate := range candidates {
		candidateIDs[string(candidate.Provider)+":"+candidate.ThreadID] = true
	}
	for _, session := range sessions {
		if !candidateIDs[string(session.Provider)+":"+session.ID] {
			continue
		}
		turns, err := load(ctx, session)
		if err != nil {
			return nil, append(warnings, "ftsindex: candidate hydration failed; using direct scan"), false
		}
		records = append(records, Normalize(session, turns)...)
	}
	return records, warnings, true
}
