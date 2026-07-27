package ftsindex

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/yaleh/meta-cc/internal/conversation"
)

// SearchFilter narrows a Search call. CWD is intentionally mandatory (see
// Search): every search is scoped to one project boundary by construction,
// the same rule DIR-030 enforces for query_sessions/query_session_content
// after the cross-project cwd-boundary bug that motivated it.
type SearchFilter struct {
	CWD      string // required: project boundary, never optional
	Provider conversation.ProviderID
	ThreadID string
	Role     string
	Kind     string
	ToolName string
}

// SearchHit is one FTS match plus enough provenance to hydrate the exact
// canonical Item/Turn from the authoritative source (Contract: "search
// returns provenance and can hydrate exact canonical records after
// candidate selection").
type SearchHit struct {
	Provider  conversation.ProviderID
	ThreadID  string
	TurnID    string
	ItemID    string
	Role      string
	Kind      string
	ToolName  string
	CWD       string
	Title     string
	Timestamp time.Time
	Snippet   string
	Truncated bool
}

// Search runs an FTS5 MATCH query against the index, always scoped to
// filter.CWD, plus any other filter dimensions supplied. It returns an
// error (rather than silently searching every project) if filter.CWD is
// empty — Search has no "all projects" mode, matching every other
// query/search entry point in this codebase.
func Search(ctx context.Context, db *sql.DB, query string, filter SearchFilter, limit int) ([]SearchHit, error) {
	if filter.CWD == "" {
		return nil, fmt.Errorf("ftsindex: Search requires filter.CWD (no cross-project search)")
	}
	if limit <= 0 {
		limit = 50
	}

	var b strings.Builder
	args := []interface{}{query}
	b.WriteString(`
		SELECT items.provider, items.thread_id, items.turn_id, items.item_id,
		       items.role, items.kind, items.tool_name, items.cwd, items.title,
		       items.ts_unix, items.truncated,
		       snippet(items_fts, 0, '[', ']', ' ... ', 10)
		FROM items_fts
		JOIN items ON items.rowid = items_fts.rowid
		WHERE items_fts MATCH ?
		  AND items.cwd = ?`)
	args = append(args, filter.CWD)

	if filter.Provider != "" {
		b.WriteString(" AND items.provider = ?")
		args = append(args, string(filter.Provider))
	}
	if filter.ThreadID != "" {
		b.WriteString(" AND items.thread_id = ?")
		args = append(args, filter.ThreadID)
	}
	if filter.Role != "" {
		b.WriteString(" AND items.role = ?")
		args = append(args, filter.Role)
	}
	if filter.Kind != "" {
		b.WriteString(" AND items.kind = ?")
		args = append(args, filter.Kind)
	}
	if filter.ToolName != "" {
		b.WriteString(" AND items.tool_name = ?")
		args = append(args, filter.ToolName)
	}
	b.WriteString(" ORDER BY items_fts.rank LIMIT ?")
	args = append(args, limit)

	rows, err := db.QueryContext(ctx, b.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("ftsindex: search query: %w", err)
	}
	defer rows.Close()

	var hits []SearchHit
	for rows.Next() {
		var (
			provider, threadID, turnID, itemIDv, role, kind, toolName, cwd, title, snippet string
			tsUnix                                                                         int64
			truncatedInt                                                                   int
		)
		if err := rows.Scan(&provider, &threadID, &turnID, &itemIDv, &role, &kind, &toolName, &cwd, &title, &tsUnix, &truncatedInt, &snippet); err != nil {
			return nil, fmt.Errorf("ftsindex: scan search row: %w", err)
		}
		hits = append(hits, SearchHit{
			Provider:  conversation.ProviderID(provider),
			ThreadID:  threadID,
			TurnID:    turnID,
			ItemID:    itemIDv,
			Role:      role,
			Kind:      kind,
			ToolName:  toolName,
			CWD:       cwd,
			Title:     title,
			Timestamp: time.Unix(tsUnix, 0).UTC(),
			Snippet:   snippet,
			Truncated: truncatedInt != 0,
		})
	}
	return hits, rows.Err()
}
