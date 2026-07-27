package conversation

import (
	"sort"
	"strings"
	"time"
)

// SessionFilter specifies metadata-only filters for listing session/thread
// metadata (see the MCP query_sessions tool and DIR-030) without requiring
// turn/rollout content to be loaded. Every field is optional; the zero
// value for a field means "no constraint on that dimension" — see Matches.
//
// SessionID and Cursor/Limit are pagination/exact-lookup concerns handled
// by the caller (a provider's exact-lookup fast path, or the MCP layer's
// existing offset/page_size pagination), not by Matches itself.
type SessionFilter struct {
	// SessionID, when set, is an exact session/thread ID selector. Callers
	// implementing metadata listing SHOULD treat a non-empty SessionID as a
	// fast path (GetSession instead of ListSessions) rather than listing
	// everything and filtering by ID.
	SessionID string

	CWD              string
	Archived         *bool // nil = don't filter by archived state
	SourceKinds      []string
	ModelProvider    string
	Model            string // substring match against Session.Model
	Status           string // "active" or "archived"; must agree with Archived if both are set
	ParentThreadID   string
	TitleContains    string // case-insensitive substring match against Session.Title
	CreatedSince     *time.Time
	CreatedUntil     *time.Time
	UpdatedSince     *time.Time
	UpdatedUntil     *time.Time
	IncludeSubagents bool // when false, excludes sessions with IsSubagent == true
}

// Matches reports whether s satisfies every filter dimension set on f.
// SessionID is intentionally NOT checked here (it is a fast-path selector
// handled before Matches is ever called), so a filter with SessionID set
// still applies its other dimensions (e.g. an archived filter) to the one
// session fetched via the fast path.
func (f SessionFilter) Matches(s Session) bool {
	if f.CWD != "" && s.CWD != f.CWD {
		return false
	}
	if f.Archived != nil && s.Archived != *f.Archived {
		return false
	}
	if len(f.SourceKinds) > 0 && !containsFold(f.SourceKinds, s.SourceKind) {
		return false
	}
	if f.ModelProvider != "" && !strings.EqualFold(s.ModelProvider, f.ModelProvider) {
		return false
	}
	if f.Model != "" && !strings.Contains(strings.ToLower(s.Model), strings.ToLower(f.Model)) {
		return false
	}
	if f.Status != "" && !strings.EqualFold(f.Status, statusFor(s.Archived)) {
		return false
	}
	if f.ParentThreadID != "" && s.ParentThreadID != f.ParentThreadID {
		return false
	}
	if f.TitleContains != "" && !strings.Contains(strings.ToLower(s.Title), strings.ToLower(f.TitleContains)) {
		return false
	}
	if f.CreatedSince != nil && s.CreatedAt.Before(*f.CreatedSince) {
		return false
	}
	if f.CreatedUntil != nil && !s.CreatedAt.Before(*f.CreatedUntil) {
		return false
	}
	if f.UpdatedSince != nil && s.UpdatedAt.Before(*f.UpdatedSince) {
		return false
	}
	if f.UpdatedUntil != nil && !s.UpdatedAt.Before(*f.UpdatedUntil) {
		return false
	}
	if !f.IncludeSubagents && s.IsSubagent {
		return false
	}
	return true
}

// ApplyFilter returns the subset of sessions matching f, preserving
// relative order. It is the correctness backstop used regardless of how
// much (if any) of f a given backend already pushed down into its own
// listing call — pushdown is a performance optimization, ApplyFilter is
// what guarantees the final result is actually correct.
func ApplyFilter(sessions []Session, f SessionFilter) []Session {
	out := make([]Session, 0, len(sessions))
	for _, s := range sessions {
		if f.Matches(s) {
			out = append(out, s)
		}
	}
	return out
}

// SortSessionsDeterministic sorts sessions by CreatedAt descending, then ID
// ascending as a tie-break, in place. This is the single ordering used
// across providers/backends for metadata listing so that cursor-based
// (offset) pagination over the merged, multi-provider result is stable.
func SortSessionsDeterministic(sessions []Session) {
	sort.SliceStable(sessions, func(i, j int) bool {
		a, b := sessions[i], sessions[j]
		if !a.CreatedAt.Equal(b.CreatedAt) {
			return a.CreatedAt.After(b.CreatedAt)
		}
		return a.ID < b.ID
	})
}

// ChildrenOf returns the subset of sessions whose ParentThreadID equals
// parentID, preserving relative order. It is a pure, metadata-only lineage
// query (DIR-032): callers are responsible for having already scoped
// sessions to whatever cwd/project boundary applies (see
// internal/mcp/executor/query_sessions_handler.go) before calling this —
// ChildrenOf itself performs no lookup and enforces no boundary, so it
// cannot introduce a cross-project leak on its own.
func ChildrenOf(sessions []Session, parentID string) []Session {
	if parentID == "" {
		return nil
	}
	out := make([]Session, 0)
	for _, s := range sessions {
		if s.ParentThreadID == parentID {
			out = append(out, s)
		}
	}
	return out
}

func statusFor(archived bool) string {
	if archived {
		return "archived"
	}
	return "active"
}

func containsFold(list []string, v string) bool {
	for _, item := range list {
		if strings.EqualFold(item, v) {
			return true
		}
	}
	return false
}
