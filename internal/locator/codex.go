package locator

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
)

const codexRootEnv = "META_CC_CODEX_ROOT"

type CodexLocator struct {
	codexRoot string
}

func NewCodexLocator() *CodexLocator {
	root := resolveCodexRoot()
	return &CodexLocator{codexRoot: root}
}

func resolveCodexRoot() string {
	for _, key := range []string{codexRootEnv, codexHomeEnv} {
		if root := os.Getenv(key); root != "" {
			return filepath.Clean(root)
		}
	}
	if homeDir, err := os.UserHomeDir(); err == nil {
		return filepath.Join(homeDir, ".codex")
	}
	return ""
}

// Root returns the resolved Codex home directory (META_CC_CODEX_ROOT
// override, or ~/.codex by default). Callers that need to point an external
// `codex` process at the same state meta-cc is configured to read (e.g. the
// app-server backend) should set CODEX_HOME to this value rather than
// re-deriving it, so the two agree even when META_CC_CODEX_ROOT is set.
func (l *CodexLocator) Root() string {
	return l.codexRoot
}

var stateDBPattern = regexp.MustCompile(`^state_([0-9]+)\.sqlite$`)

// SQLiteDB returns the highest-numbered state_N.sqlite present in the
// canonical Codex root. Schema compatibility is validated by the provider;
// when no candidate exists, state_5.sqlite is returned for diagnostics and
// backward-compatible path reporting.
func (l *CodexLocator) SQLiteDB() string {
	candidates := l.SQLiteDBCandidates()
	if len(candidates) > 0 {
		return candidates[0]
	}
	return filepath.Join(l.codexRoot, "state_5.sqlite")
}

// SQLiteDBCandidates returns state databases newest-first. The provider
// checks schema compatibility in this deterministic order.
func (l *CodexLocator) SQLiteDBCandidates() []string {
	entries, err := os.ReadDir(l.codexRoot)
	if err != nil {
		return nil
	}
	type candidate struct {
		version int
		path    string
	}
	var candidates []candidate
	for _, entry := range entries {
		match := stateDBPattern.FindStringSubmatch(entry.Name())
		if entry.IsDir() || match == nil {
			continue
		}
		version, err := strconv.Atoi(match[1])
		if err == nil {
			candidates = append(candidates, candidate{version, filepath.Join(l.codexRoot, entry.Name())})
		}
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].version > candidates[j].version })
	paths := make([]string, len(candidates))
	for i := range candidates {
		paths[i] = candidates[i].path
	}
	return paths
}

func (l *CodexLocator) ArchivedSessionsRoot() string {
	return filepath.Join(l.codexRoot, "archived_sessions")
}

func (l *CodexLocator) SessionsRoot() string {
	return filepath.Join(l.codexRoot, "sessions")
}

func (l *CodexLocator) HistoryFile() string {
	return filepath.Join(l.codexRoot, "history.jsonl")
}
