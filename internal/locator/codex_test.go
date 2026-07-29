package locator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCodexLocatorEnvOverride(t *testing.T) {
	t.Setenv(codexRootEnv, "/tmp/codex")
	loc := NewCodexLocator()
	if got := loc.SQLiteDB(); got != "/tmp/codex/state_5.sqlite" {
		t.Fatalf("SQLiteDB() = %s", got)
	}
}

func TestCodexLocatorCODEXHomeAndPrecedence(t *testing.T) {
	codexHome := filepath.Join(t.TempDir(), "codex-home")
	override := filepath.Join(t.TempDir(), "override")
	t.Setenv(codexRootEnv, "")
	t.Setenv(codexHomeEnv, codexHome)
	if got := NewCodexLocator().Root(); got != codexHome {
		t.Fatalf("Root() with CODEX_HOME = %q, want %q", got, codexHome)
	}
	t.Setenv(codexRootEnv, override)
	if got := NewCodexLocator().Root(); got != override {
		t.Fatalf("Root() precedence = %q, want META_CC_CODEX_ROOT %q", got, override)
	}
}

func TestCodexLocatorSelectsHighestStateDB(t *testing.T) {
	root := t.TempDir()
	t.Setenv(codexRootEnv, root)
	t.Setenv(codexHomeEnv, "")
	for _, name := range []string{"state_5.sqlite", "state_6.sqlite", "state_x.sqlite"} {
		if err := os.WriteFile(filepath.Join(root, name), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if got, want := NewCodexLocator().SQLiteDB(), filepath.Join(root, "state_6.sqlite"); got != want {
		t.Fatalf("SQLiteDB() = %q, want %q", got, want)
	}
}

func TestCodexLocatorDefaultPaths(t *testing.T) {
	t.Setenv(codexRootEnv, "")
	t.Setenv(codexHomeEnv, "")
	home := t.TempDir()
	t.Setenv("HOME", home)
	loc := NewCodexLocator()

	if got := loc.SessionsRoot(); got != filepath.Join(home, ".codex", "sessions") {
		t.Fatalf("SessionsRoot() = %s", got)
	}
	if got := loc.HistoryFile(); got != filepath.Join(home, ".codex", "history.jsonl") {
		t.Fatalf("HistoryFile() = %s", got)
	}
}
