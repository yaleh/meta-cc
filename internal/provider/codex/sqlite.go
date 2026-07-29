package codex

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
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

func compatibleThreadsSchema(ctx context.Context, dsn string) error {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return err
	}
	defer db.Close()
	if err := ensureThreadsTable(ctx, db); err != nil {
		return err
	}
	rows, err := db.QueryContext(ctx, "PRAGMA table_info(threads)")
	if err != nil {
		return err
	}
	defer rows.Close()
	columns := map[string]bool{}
	for rows.Next() {
		var cid int
		var name, typ string
		var notNull, pk int
		var defaultValue interface{}
		if err := rows.Scan(&cid, &name, &typ, &notNull, &defaultValue, &pk); err != nil {
			return err
		}
		columns[name] = true
	}
	for _, required := range []string{"id", "cwd", "title", "source"} {
		if !columns[required] {
			return &SQLiteSchemaError{Message: "missing column: " + required}
		}
	}
	return rows.Err()
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

// extractTime converts a threads-table timestamp column into a time.Time,
// returning the zero time for any value that is absent or not a valid epoch
// so the created_at_ms -> created_at fallback chain (in scanSession) can
// engage on IsZero(). DIR-063: it must NEVER manufacture a bogus 1970
// instant from a present-but-unparseable value — a TEXT timestamp is parsed
// as RFC3339, an epoch is only accepted as a fully-numeric string or a
// positive integer, and zero/negative/empty/non-numeric all map to the zero
// time. (A real session never lived at the 1970 epoch, so treating int64(0)
// as "absent" rather than a valid instant is always correct here.)
func extractTime(value interface{}) time.Time {
	switch v := value.(type) {
	case int64:
		switch {
		case v > 1_000_000_000_000:
			return time.UnixMilli(v)
		case v <= 0:
			return time.Time{}
		default:
			return time.Unix(v, 0)
		}
	case []byte:
		return extractTime(stringValue(v))
	case string:
		// TEXT timestamp: try RFC3339 first, then a fully-numeric epoch
		// string. strconv.ParseInt parses the WHOLE string, so a date like
		// "2026-07-28T10:00:00Z" is rejected outright instead of being read
		// as the leading integer 2026 (the old fmt.Sscan bug).
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			return t
		}
		iv, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return time.Time{}
		}
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
