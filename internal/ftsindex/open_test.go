package ftsindex

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureSchema_CreatesFTS5Table(t *testing.T) {
	db, _ := openTestDB(t)
	var name string
	err := db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='items_fts'`).Scan(&name)
	if err != nil {
		t.Fatalf("items_fts table not created: %v", err)
	}
	if v := storedSchemaVersion(context.Background(), db); v != SchemaVersion {
		t.Fatalf("stored schema version = %d, want %d", v, SchemaVersion)
	}
}

func TestOpen_CorruptFileDegradesGracefully(t *testing.T) {
	path := filepath.Join(t.TempDir(), "fts.db")
	if err := os.WriteFile(path, []byte("this is not a sqlite file, just garbage bytes"), 0o644); err != nil {
		t.Fatalf("write garbage file: %v", err)
	}

	db, degraded, err := Open(context.Background(), path)
	if err != nil {
		t.Fatalf("Open on corrupt file returned a hard error instead of degrading: %v", err)
	}
	defer db.Close()
	if !degraded {
		t.Fatalf("Open on corrupt file did not report degraded=true")
	}

	// The recovered index must be immediately usable (healthy, empty schema).
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM sessions`).Scan(&count); err != nil {
		t.Fatalf("recovered index is not queryable: %v", err)
	}
	if count != 0 {
		t.Fatalf("recovered index sessions count = %d, want 0", count)
	}

	hits, err := Search(context.Background(), db, "anything", SearchFilter{CWD: "/proj"}, 10)
	if err != nil {
		t.Fatalf("Search on recovered index failed: %v", err)
	}
	if len(hits) != 0 {
		t.Fatalf("Search on freshly recovered (empty) index returned %d hits, want 0", len(hits))
	}
}

func TestOpen_ReopeningHealthyIndexIsNotDegraded(t *testing.T) {
	path := filepath.Join(t.TempDir(), "fts.db")
	db1, degraded, err := Open(context.Background(), path)
	if err != nil || degraded {
		t.Fatalf("first Open: degraded=%v err=%v", degraded, err)
	}
	db1.Close()

	db2, degraded, err := Open(context.Background(), path)
	if err != nil {
		t.Fatalf("second Open: %v", err)
	}
	defer db2.Close()
	if degraded {
		t.Fatalf("reopening a healthy index reported degraded=true")
	}
}
