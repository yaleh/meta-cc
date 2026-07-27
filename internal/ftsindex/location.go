package ftsindex

import (
	"os"
	"path/filepath"
)

// indexDirName/indexFileName follow the repo's existing `.meta-cc/`
// project-local derived-data convention (see .meta-cc/prompts/library for
// precedent): derived/cache data lives under `.meta-cc/`, is git-ignored by
// default, and is safe to delete/rebuild at any time.
const (
	indexDirName  = "index"
	indexFileName = "fts.db"
)

// DefaultPath returns the on-disk path of the FTS index database for a
// given project directory: "<projectPath>/.meta-cc/index/fts.db". One index
// file covers every provider for that project (rows are scoped by the
// provider+cwd columns), matching how query_sessions/query_session_content
// already scope a single call to one project boundary.
func DefaultPath(projectPath string) string {
	return filepath.Join(projectPath, ".meta-cc", indexDirName, indexFileName)
}

// EnsureDir creates the index directory (and a .gitignore inside it, so the
// derived database is never accidentally committed) for the given index
// path. Safe to call repeatedly.
func EnsureDir(indexPath string) error {
	dir := filepath.Dir(indexPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	gitignore := filepath.Join(dir, ".gitignore")
	if _, err := os.Stat(gitignore); os.IsNotExist(err) {
		content := "# Derived FTS index (DIR-031) - safe to delete, rebuilt on demand.\n*\n"
		_ = os.WriteFile(gitignore, []byte(content), 0o644)
	}
	return nil
}
