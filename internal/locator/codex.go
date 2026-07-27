package locator

import (
	"os"
	"path/filepath"
)

const codexRootEnv = "META_CC_CODEX_ROOT"

type CodexLocator struct {
	codexRoot string
}

func NewCodexLocator() *CodexLocator {
	root := os.Getenv(codexRootEnv)
	if root == "" {
		if homeDir, err := os.UserHomeDir(); err == nil {
			root = filepath.Join(homeDir, ".codex")
		}
	} else {
		root = filepath.Clean(root)
	}

	return &CodexLocator{codexRoot: root}
}

// Root returns the resolved Codex home directory (META_CC_CODEX_ROOT
// override, or ~/.codex by default). Callers that need to point an external
// `codex` process at the same state meta-cc is configured to read (e.g. the
// app-server backend) should set CODEX_HOME to this value rather than
// re-deriving it, so the two agree even when META_CC_CODEX_ROOT is set.
func (l *CodexLocator) Root() string {
	return l.codexRoot
}

func (l *CodexLocator) SQLiteDB() string {
	return filepath.Join(l.codexRoot, "state_5.sqlite")
}

func (l *CodexLocator) SessionsRoot() string {
	return filepath.Join(l.codexRoot, "sessions")
}

func (l *CodexLocator) HistoryFile() string {
	return filepath.Join(l.codexRoot, "history.jsonl")
}
