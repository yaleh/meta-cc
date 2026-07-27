package codex

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"github.com/yaleh/meta-cc/internal/conversation"
)

func listSessionsFromDB(ctx context.Context, dsn string) ([]conversation.Session, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	if err := ensureThreadsTable(ctx, db); err != nil {
		return nil, err
	}

	rows, err := db.QueryContext(ctx, "SELECT * FROM threads ORDER BY created_at DESC, id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []conversation.Session
	for rows.Next() {
		session, err := scanSession(rows)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}
	return sessions, rows.Err()
}

// listSessionsFromDBFiltered is listSessionsFromDB extended with DIR-030
// metadata filters. It pushes a "cwd" predicate into the SQL WHERE clause
// when filter.CWD is set (the "cwd" column is part of the tested/supported
// threads schema — see scanSession's "required" columns — so this pushdown
// is always safe), then applies conversation.ApplyFilter for every other
// dimension in Go. That second pass still touches only the metadata this
// SELECT already fetched — no rollout file is opened — so this remains a
// metadata-only listing regardless of how much of the filter reached SQL.
//
// A single row that fails to scan (e.g. a NULL/type mismatch in one
// corrupted threads-table row) is skipped with a warning rather than
// aborting the whole listing, mirroring the per-session tolerance the
// files backend's turn-loading path (providerrecords.Build) applies.
func listSessionsFromDBFiltered(ctx context.Context, dsn string, filter conversation.SessionFilter) ([]conversation.Session, []string, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, nil, err
	}
	defer db.Close()

	if err := ensureThreadsTable(ctx, db); err != nil {
		return nil, nil, err
	}

	query := "SELECT * FROM threads"
	var args []interface{}
	if filter.CWD != "" {
		query += " WHERE cwd = ?"
		args = append(args, filter.CWD)
	}
	query += " ORDER BY created_at DESC, id DESC"

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var sessions []conversation.Session
	var warnings []string
	for rows.Next() {
		session, err := scanSession(rows)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("codex files backend: skipped unreadable threads row: %v", err))
			continue
		}
		sessions = append(sessions, session)
	}
	if err := rows.Err(); err != nil {
		return nil, warnings, err
	}
	return conversation.ApplyFilter(sessions, filter), warnings, nil
}

func getSessionFromDB(ctx context.Context, dsn, sessionID string) (conversation.Session, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return conversation.Session{}, err
	}
	defer db.Close()

	if err := ensureThreadsTable(ctx, db); err != nil {
		return conversation.Session{}, err
	}

	rows, err := db.QueryContext(ctx, "SELECT * FROM threads WHERE id = ? LIMIT 1", sessionID)
	if err != nil {
		return conversation.Session{}, err
	}
	defer rows.Close()
	if !rows.Next() {
		return conversation.Session{}, ErrSessionNotFound
	}

	return scanSession(rows)
}

func ensureThreadsTable(ctx context.Context, db *sql.DB) error {
	var name string
	err := db.QueryRowContext(ctx, "SELECT name FROM sqlite_master WHERE type='table' AND name='threads'").Scan(&name)
	if err == sql.ErrNoRows || name == "" {
		return &SQLiteSchemaError{Message: "missing threads table"}
	}
	return err
}

func scanSession(rows *sql.Rows) (conversation.Session, error) {
	cols, err := rows.Columns()
	if err != nil {
		return conversation.Session{}, err
	}
	values := make([]interface{}, len(cols))
	ptrs := make([]interface{}, len(cols))
	for i := range values {
		ptrs[i] = &values[i]
	}
	if err := rows.Scan(ptrs...); err != nil {
		return conversation.Session{}, err
	}

	colMap := make(map[string]interface{}, len(cols))
	for i, col := range cols {
		colMap[col] = values[i]
	}

	required := []string{"id", "cwd", "title", "source"}
	for _, key := range required {
		if _, ok := colMap[key]; !ok {
			return conversation.Session{}, &SQLiteSchemaError{Message: "missing column: " + key}
		}
	}

	createdAt := extractTime(colMap["created_at_ms"])
	if createdAt.IsZero() {
		createdAt = extractTime(colMap["created_at"])
	}
	// updated_at/updated_at_ms and archived/parent_thread_id are NOT part
	// of the tested/documented threads schema (see sqlite_test.go), so
	// these are opportunistic: present on schemas that have them, zero
	// value (colMap lookup miss -> "") on schemas that don't, exactly like
	// every other optional column this function already tolerates.
	updatedAt := extractTime(colMap["updated_at_ms"])
	if updatedAt.IsZero() {
		updatedAt = extractTime(colMap["updated_at"])
	}
	if updatedAt.IsZero() {
		updatedAt = createdAt
	}
	sourceKind := stringValue(colMap["source"])
	archived := boolValue(colMap["archived"])
	isSubagent := strings.HasPrefix(sourceKind, "subAgent")

	// Lineage (DIR-032): a threads-table schema without a parent_thread_id
	// column at all cannot report lineage, so it's Unknown rather than
	// Root — an older/incompatible schema must not be presented as "this
	// session confirmed has no parent". When the column exists, an empty
	// value for a subagent source kind is still Unknown (spawn metadata
	// was not recorded even though the column exists); every other empty
	// value is a confirmed Root.
	_, hasParentCol := colMap["parent_thread_id"]
	parentThreadID := stringValue(colMap["parent_thread_id"])
	lineage := conversation.LineageStatusUnknown
	if hasParentCol {
		switch {
		case parentThreadID != "":
			lineage = conversation.LineageStatusChild
		case isSubagent:
			lineage = conversation.LineageStatusUnknown
		default:
			lineage = conversation.LineageStatusRoot
		}
	}

	ext, _ := json.Marshal(map[string]interface{}{
		"rollout_path":   stringValue(colMap["rollout_path"]),
		"source":         sourceKind,
		"thread_source":  stringValue(colMap["thread_source"]),
		"model_provider": stringValue(colMap["model_provider"]),
	})

	status := "active"
	if archived {
		status = "archived"
	}

	return conversation.Session{
		ID:             stringValue(colMap["id"]),
		Provider:       conversation.ProviderCodex,
		Title:          stringValue(colMap["title"]),
		CWD:            stringValue(colMap["cwd"]),
		Model:          firstNonEmpty(stringValue(colMap["model"]), stringValue(colMap["model_provider"])),
		ModelProvider:  stringValue(colMap["model_provider"]),
		SourceKind:     sourceKind,
		Status:         status,
		Archived:       archived,
		ParentThreadID: parentThreadID,
		Lineage:        lineage,
		IsSubagent:     isSubagent,
		CreatedAt:      createdAt.UTC(),
		UpdatedAt:      updatedAt.UTC(),
		TokenUsage: conversation.TokenUsage{
			InputTokens: intValue(colMap["tokens_used"]),
		},
		Extensions: ext,
	}, nil
}

func extractTime(value interface{}) time.Time {
	switch v := value.(type) {
	case int64:
		if v > 1_000_000_000_000 {
			return time.UnixMilli(v)
		}
		return time.Unix(v, 0)
	case []byte:
		return extractTime(stringValue(v))
	case string:
		var iv int64
		_, _ = fmt.Sscan(v, &iv)
		return extractTime(iv)
	default:
		return time.Time{}
	}
}

func stringValue(v interface{}) string {
	switch x := v.(type) {
	case string:
		return x
	case []byte:
		return string(x)
	default:
		return ""
	}
}

func intValue(v interface{}) int {
	switch x := v.(type) {
	case int:
		return x
	case int64:
		return int(x)
	case []byte:
		var n int
		_, _ = fmt.Sscan(string(x), &n)
		return n
	case string:
		var n int
		_, _ = fmt.Sscan(x, &n)
		return n
	default:
		return 0
	}
}

// boolValue interprets a threads-table column value as a boolean. SQLite
// has no native boolean type, so drivers commonly surface it as int64(0/1)
// or occasionally a string; a missing column (nil) or any unrecognized
// value defaults to false, matching the "unknown metadata" convention used
// throughout this file rather than erroring.
func boolValue(v interface{}) bool {
	switch x := v.(type) {
	case bool:
		return x
	case int64:
		return x != 0
	case int:
		return x != 0
	case []byte:
		return boolValue(string(x))
	case string:
		return x == "1" || x == "true" || x == "t"
	default:
		return false
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
