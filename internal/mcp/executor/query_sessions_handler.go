package executor

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yaleh/meta-cc/internal/conversation"
	"github.com/yaleh/meta-cc/internal/locator"
	mcquery "github.com/yaleh/meta-cc/internal/mcp/query"
	claudeprovider "github.com/yaleh/meta-cc/internal/provider/claude"
	codexprovider "github.com/yaleh/meta-cc/internal/provider/codex"
	"github.com/yaleh/meta-cc/internal/provider/rawfiles"
	providerrecords "github.com/yaleh/meta-cc/internal/provider/records"
)

// query_sessions_handler.go implements the DIR-030 metadata-first session
// discovery tool: it lists session/thread metadata (id, cwd, title,
// status, source kind, model, archived, parent thread, timestamps)
// WITHOUT loading turn/rollout content, and supports the filter
// dimensions in conversation.SessionFilter. It is a distinct tool from
// scope="session" on the existing content/signal tools (which means "the
// most recently modified session") — see docs/guides/mcp-query-tools.md.

func init() {
	registerQueryHandler("query_sessions", func(e *ToolExecutor, scope string, args map[string]interface{}) (mcquery.QueryResult, error) {
		return handleQuerySessions(e, scope, args)
	})
}

// codexOnlyFilterArgs names the query_sessions arguments that only apply
// to Codex session metadata (Claude sessions never populate these fields).
// Using one of these with an explicit/default provider="claude" can never
// match anything, so it fails closed with an actionable error instead of
// silently returning zero results (which would look identical to "no
// sessions matched", masking the real mistake).
var codexOnlyFilterArgs = []string{"source_kind", "model_provider", "parent_thread_id", "archived", "status", "ancestors_of"}

// maxLineageDepth bounds ancestor-chain traversal (see handleAncestorsOf):
// a pathological/cyclical parent_thread_id chain must terminate rather than
// looping or recursing without limit.
const maxLineageDepth = 32

// newCodexProvider constructs the Codex provider handleQuerySessions and
// handleAncestorsOf use. It's a package-level var (not a hardcoded call at
// each use site) so this package's own tests can substitute a Provider
// backed by a fake app-server (see codexprovider.NewProviderForAppServerTest
// and DIR-039's TestQuerySessions_Codex_LimitToleratesMidPaginationFailure),
// exercising the real FetchSessionsBounded partial-result-plus-warning
// wiring without spawning a subprocess.
var newCodexProvider = func() *codexprovider.Provider {
	return codexprovider.NewProvider(locator.NewCodexLocator())
}

func handleQuerySessions(_ *ToolExecutor, scope string, args map[string]interface{}) (mcquery.QueryResult, error) {
	providerName := providerParam(args)
	workingDir := GetStringParam(args, "working_dir", "")

	providers, err := rawfiles.ParseProviderFilter(providerName)
	if err != nil {
		return mcquery.QueryResult{}, err
	}

	if providerName == "claude" {
		for _, argName := range codexOnlyFilterArgs {
			if _, set := args[argName]; set {
				return mcquery.QueryResult{}, fmt.Errorf(
					"filter %q is only supported for provider=\"codex\" or provider=\"all\" (Claude sessions don't carry this metadata); "+
						"pass provider explicitly or drop the filter", argName)
			}
		}
	}

	filter, err := buildSessionFilterFromArgs(args)
	if err != nil {
		return mcquery.QueryResult{}, err
	}

	projectPath := workingDir
	if projectPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return mcquery.QueryResult{}, err
		}
		projectPath = cwd
	}
	if abs, err := filepath.Abs(projectPath); err == nil {
		projectPath = abs
	}
	// The project/cwd boundary: an explicit "cwd" filter narrows further,
	// but scope always stays bounded to projectPath — query_sessions never
	// lists sessions across every project on disk (matching every other
	// query/analysis tool's default project scope).
	boundaryCWD := projectPath
	if filter.CWD != "" {
		boundaryCWD = filter.CWD
	} else {
		filter.CWD = projectPath
	}

	ctx := context.Background()

	if ancestorsOf := GetStringParam(args, "ancestors_of", ""); ancestorsOf != "" {
		return handleAncestorsOf(ctx, boundaryCWD, ancestorsOf)
	}
	var (
		merged   []conversation.Session
		warnings []string
	)

	limit := GetIntParam(args, "limit", 0)

	for _, pid := range providers {
		switch pid {
		case conversation.ProviderCodex:
			p := newCodexProvider()
			if !p.IsAvailable(ctx) {
				warnings = append(warnings, "provider codex unavailable")
				continue
			}
			var sessions []conversation.Session
			var err error
			// DIR-034: use the DIR-032 cursor-based ListSessionsPage (via
			// FetchSessionsBounded) instead of a full-corpus
			// ListSessionsFiltered crawl, but only when it's actually safe
			// to bound: providerName=="codex" alone (not merged with other
			// providers' unbounded results), a real limit was requested, no
			// exact session_id lookup is in play (already O(1) via
			// ListSessionsFiltered), scope != "session" (that scope needs
			// the true most-recently-modified session across the WHOLE
			// corpus, which a partial fetch cannot guarantee), and
			// filter.Archived is already resolved to a single boolean
			// (ListSessionsPage fetches one archived-state pass per call —
			// see FetchSessionsBounded's doc). Anything else keeps the
			// existing unbounded behavior unchanged.
			if providerName == "codex" && limit > 0 && scope != "session" &&
				filter.SessionID == "" && filter.Archived != nil {
				sessions, err = p.FetchSessionsBounded(ctx, filter, limit)
			} else {
				sessions, err = p.ListSessionsFiltered(ctx, filter)
			}
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("provider codex: %v", err))
				continue
			}
			warnings = append(warnings, p.Warnings()...)
			sessions = reduceForScope(sessions, scope, boundaryCWD, conversation.ProviderCodex)
			merged = append(merged, sessions...)

		case conversation.ProviderClaude:
			p := claudeprovider.NewProvider(locator.NewSessionLocator(), boundaryCWD)
			if !p.IsAvailable(ctx) {
				warnings = append(warnings, "provider claude unavailable")
				continue
			}
			var sessions []conversation.Session
			if filter.SessionID != "" {
				session, err := p.GetSession(ctx, filter.SessionID)
				if err != nil {
					warnings = append(warnings, fmt.Sprintf("provider claude: session %s not found: %v", filter.SessionID, err))
					continue
				}
				sessions = []conversation.Session{session}
			} else {
				sessions, err = p.ListSessions(ctx)
				if err != nil {
					return mcquery.QueryResult{}, fmt.Errorf("provider claude: %w", err)
				}
			}
			claudeFilter := filter
			claudeFilter.SessionID = "" // already resolved via GetSession above
			sessions = conversation.ApplyFilter(sessions, claudeFilter)
			sessions = reduceForScope(sessions, scope, boundaryCWD, conversation.ProviderClaude)
			merged = append(merged, sessions...)
		}
	}

	conversation.SortSessionsDeterministic(merged)

	if limit > 0 && len(merged) > limit {
		merged = merged[:limit]
	}

	entries := make([]interface{}, 0, len(merged))
	for _, s := range merged {
		entries = append(entries, sessionToEntry(s))
	}

	return mcquery.QueryResult{Entries: entries, Warnings: warnings}, nil
}

// reduceForScope applies scope=="session" (most recent) via the same
// providerrecords.FilterSessionsForScope helper the content-query path
// uses, so query_sessions and query_session_content agree on what
// "session scope" means (backward-compatible: "most recently modified
// session", NOT the same thing as an exact session_id — see
// docs/guides/mcp-query-tools.md). scope=="project" (default) is a no-op
// here since these sessions are already cwd-boundary-filtered.
func reduceForScope(sessions []conversation.Session, scope, projectPath string, providerID conversation.ProviderID) []conversation.Session {
	return providerrecords.FilterSessionsForScope(sessions, scope, projectPath, providerID)
}

// handleAncestorsOf is the DIR-032 lineage-traversal query: given a Codex
// thread ID, it walks ParentThreadID upward (immediate parent, then
// grandparent, ...) and returns that chain, each entry annotated with its
// "lineage" classification (conversation.LineageStatus). It is a distinct
// code path from the normal listing above (not just another SessionFilter
// dimension) because it performs repeated single-ID lookups rather than one
// list-and-filter pass.
//
// Boundary enforcement (DIR-030 precedent): every session in the chain,
// including the starting one, must resolve to boundaryCWD — the same
// project/cwd scope every other query_sessions path enforces. A lookup
// landing outside that boundary, a lookup failure, a confirmed root, an
// explicit LineageStatusUnknown, a cycle, or the maxLineageDepth bound all
// stop traversal and return whatever chain was already resolved (never a
// partial chain silently presented as complete) plus a warning explaining
// why it stopped short.
func handleAncestorsOf(ctx context.Context, boundaryCWD, sessionID string) (mcquery.QueryResult, error) {
	p := newCodexProvider()
	if !p.IsAvailable(ctx) {
		return mcquery.QueryResult{}, fmt.Errorf("provider codex unavailable")
	}

	start, err := p.GetSession(ctx, sessionID)
	if err != nil {
		return mcquery.QueryResult{}, fmt.Errorf("ancestors_of %q: %w", sessionID, err)
	}
	// DIR-030 cwd-boundary precedent: a session_id lookup that resolves
	// outside the caller's project scope must be rejected exactly like the
	// exact-ID fast paths in provider_query.go/query_sessions_handler.go —
	// this is a NEW lineage-traversal lookup path, so it gets its own
	// explicit boundary check rather than inheriting one implicitly.
	if start.CWD != boundaryCWD {
		return mcquery.QueryResult{}, fmt.Errorf("session %q not found for the requested provider(s) within project %q", sessionID, boundaryCWD)
	}

	var (
		chain     []conversation.Session
		warnings  []string
		truncated bool
	)

	if start.Lineage == conversation.LineageStatusUnknown {
		warnings = append(warnings, fmt.Sprintf("lineage for %q is unknown: spawn metadata was not available, so its ancestry cannot be determined", sessionID))
		truncated = true
	}

	current := start.ParentThreadID
	seen := map[string]bool{sessionID: true}
	for i := 0; current != "" && !truncated && i < maxLineageDepth; i++ {
		if seen[current] {
			warnings = append(warnings, fmt.Sprintf("cycle detected in ancestor chain at %q; stopping traversal", current))
			truncated = true
			break
		}
		seen[current] = true

		ancestor, err := p.GetSession(ctx, current)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("ancestor lookup for %q failed: %v", current, err))
			truncated = true
			break
		}
		if ancestor.CWD != boundaryCWD {
			warnings = append(warnings, fmt.Sprintf("ancestor %q lies outside the project boundary %q; stopping traversal rather than crossing it", current, boundaryCWD))
			truncated = true
			break
		}

		chain = append(chain, ancestor)

		if ancestor.Lineage == conversation.LineageStatusUnknown {
			warnings = append(warnings, fmt.Sprintf("lineage for %q is unknown beyond this point", current))
			truncated = true
			break
		}
		current = ancestor.ParentThreadID
		if current == "" && i == maxLineageDepth-1 {
			// Loop about to exit on depth alone right as we also ran out of
			// parent — no separate warning needed, this is a genuine root.
			break
		}
	}
	if current != "" && !truncated {
		warnings = append(warnings, fmt.Sprintf("ancestor chain depth limit (%d) reached; traversal stopped", maxLineageDepth))
		truncated = true
	}

	entries := make([]interface{}, 0, len(chain))
	for _, s := range chain {
		entries = append(entries, sessionToEntry(s))
	}
	if truncated && len(entries) > 0 {
		entries[len(entries)-1].(map[string]interface{})["lineage_truncated"] = true
	}

	return mcquery.QueryResult{Entries: entries, Warnings: warnings}, nil
}

func sessionToEntry(s conversation.Session) map[string]interface{} {
	entry := map[string]interface{}{
		"session_id": s.ID,
		"provider":   string(s.Provider),
		"cwd":        s.CWD,
		"title":      s.Title,
		"model":      s.Model,
		"status":     s.Status,
		"archived":   s.Archived,
		"created_at": s.CreatedAt.Format(time.RFC3339),
	}
	if s.ModelProvider != "" {
		entry["model_provider"] = s.ModelProvider
	}
	if s.SourceKind != "" {
		entry["source_kind"] = s.SourceKind
	}
	if s.ParentThreadID != "" {
		entry["parent_thread_id"] = s.ParentThreadID
	}
	if s.Lineage != "" {
		entry["lineage"] = string(s.Lineage)
	}
	if s.IsSubagent {
		entry["is_subagent"] = true
	}
	if !s.UpdatedAt.IsZero() {
		entry["updated_at"] = s.UpdatedAt.Format(time.RFC3339)
	}
	// DIR-071: surface an opaque provider aggregate (e.g. Codex SQLite
	// threads.tokens_used) as explicitly-labeled fields with provenance, never
	// as input_tokens — so a session-level total is attributable and can be
	// reconciled against per-turn counts instead of being mistaken for input
	// usage. Absent when the source reports no aggregate.
	if s.TokenUsage.AggregateTokens != 0 {
		entry["aggregate_tokens"] = s.TokenUsage.AggregateTokens
		if s.TokenUsage.AggregateSource != "" {
			entry["aggregate_source"] = s.TokenUsage.AggregateSource
		}
	}
	return entry
}

// buildSessionFilterFromArgs parses and validates the query_sessions
// filter arguments into a conversation.SessionFilter, failing closed
// (actionable error, not a silently-ignored/ambiguous filter) on
// unparseable time values or a status/archived contradiction.
func buildSessionFilterFromArgs(args map[string]interface{}) (conversation.SessionFilter, error) {
	filter := conversation.SessionFilter{
		SessionID:        GetStringParam(args, "session_id", ""),
		CWD:              GetStringParam(args, "cwd", ""),
		Archived:         GetBoolPtrParam(args, "archived"),
		SourceKinds:      GetStringSliceParam(args, "source_kind"),
		ModelProvider:    GetStringParam(args, "model_provider", ""),
		Model:            GetStringParam(args, "model", ""),
		Status:           GetStringParam(args, "status", ""),
		ParentThreadID:   GetStringParam(args, "parent_thread_id", ""),
		TitleContains:    GetStringParam(args, "title_contains", ""),
		IncludeSubagents: GetBoolParam(args, "include_subagents", true),
	}

	if filter.Status != "" && filter.Status != "active" && filter.Status != "archived" {
		return filter, fmt.Errorf(`invalid status %q: must be "active" or "archived"`, filter.Status)
	}
	if filter.Status != "" && filter.Archived != nil {
		wantArchived := filter.Status == "archived"
		if *filter.Archived != wantArchived {
			return filter, fmt.Errorf("conflicting filters: status=%q implies archived=%v, but archived=%v was also given", filter.Status, wantArchived, *filter.Archived)
		}
	}

	// DIR-032: "archived sessions are discoverable only when requested" —
	// prior to this, an omitted archived/status filter meant "no
	// constraint" (both active and archived sessions returned), which is
	// the opposite of the Contract. Neither dimension explicitly set now
	// defaults to active-only; pass archived=true or status="archived" to
	// see archived sessions (either alone or via provider="all"/"codex").
	if filter.Archived == nil && filter.Status == "" {
		defaultArchived := false
		filter.Archived = &defaultArchived
	}

	if len(filter.SourceKinds) > 0 {
		for _, sk := range filter.SourceKinds {
			if !containsFoldSlice(codexprovider.KnownSourceKinds, sk) {
				return filter, fmt.Errorf("invalid source_kind %q: must be one of %s", sk, strings.Join(codexprovider.KnownSourceKinds, ", "))
			}
		}
	}

	var err error
	if filter.CreatedSince, err = parseOptionalRFC3339(args, "created_since"); err != nil {
		return filter, err
	}
	if filter.CreatedUntil, err = parseOptionalRFC3339(args, "created_until"); err != nil {
		return filter, err
	}
	if filter.UpdatedSince, err = parseOptionalRFC3339(args, "updated_since"); err != nil {
		return filter, err
	}
	if filter.UpdatedUntil, err = parseOptionalRFC3339(args, "updated_until"); err != nil {
		return filter, err
	}

	return filter, nil
}

func parseOptionalRFC3339(args map[string]interface{}, key string) (*time.Time, error) {
	raw := GetStringParam(args, key, "")
	if raw == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil, fmt.Errorf("invalid %s value %q: must be RFC3339 (e.g. 2026-03-07T00:00:00Z)", key, raw)
	}
	return &t, nil
}

func containsFoldSlice(list []string, v string) bool {
	for _, item := range list {
		if strings.EqualFold(item, v) {
			return true
		}
	}
	return false
}
