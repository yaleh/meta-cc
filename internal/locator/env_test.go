package locator

import (
	"path/filepath"
	"testing"
)

func TestDefaultSessionRoots(t *testing.T) {
	t.Setenv(projectsRootEnv, "")
	t.Setenv(codexHomeEnv, "")

	home := t.TempDir()
	t.Setenv("HOME", home)

	roots := DefaultSessionRoots()
	if len(roots) != 2 {
		t.Fatalf("expected 2 default roots, got %d", len(roots))
	}

	if roots[0].Host != HostClaudeCode {
		t.Fatalf("expected first root host %q, got %q", HostClaudeCode, roots[0].Host)
	}
	if got, want := roots[0].Path, filepath.Join(home, ".claude", "projects"); got != want {
		t.Fatalf("Claude Code root = %q, want %q", got, want)
	}
	if !roots[0].ProjectHashed {
		t.Fatal("Claude Code root should use project-hash directories")
	}

	if roots[1].Host != HostCodex {
		t.Fatalf("expected second root host %q, got %q", HostCodex, roots[1].Host)
	}
	if got, want := roots[1].Path, filepath.Join(home, ".codex", "sessions"); got != want {
		t.Fatalf("Codex root = %q, want %q", got, want)
	}
	if roots[1].ProjectHashed {
		t.Fatal("Codex sessions root should not be treated as Claude project-hash directories")
	}
}

func TestDefaultSessionRootsWithCodexHome(t *testing.T) {
	t.Setenv(projectsRootEnv, "")

	home := t.TempDir()
	codexHome := filepath.Join(t.TempDir(), "codex-home")
	t.Setenv("HOME", home)
	t.Setenv(codexHomeEnv, codexHome)

	roots := DefaultSessionRoots()
	if len(roots) != 2 {
		t.Fatalf("expected 2 default roots, got %d", len(roots))
	}

	if got, want := roots[1].Path, filepath.Join(codexHome, "sessions"); got != want {
		t.Fatalf("Codex root with CODEX_HOME = %q, want %q", got, want)
	}
}

func TestDefaultSessionRootsWithMetaCCCodexRootPrecedence(t *testing.T) {
	home := t.TempDir()
	codexHome := filepath.Join(t.TempDir(), "codex-home")
	override := filepath.Join(t.TempDir(), "override")
	t.Setenv("HOME", home)
	t.Setenv(codexHomeEnv, codexHome)
	t.Setenv(codexRootEnv, override)

	roots := DefaultSessionRoots()
	if got, want := roots[1].Path, filepath.Join(override, "sessions"); got != want {
		t.Fatalf("Codex transcript root = %q, want canonical override %q", got, want)
	}
}

func TestNewSessionLocatorUsesOverrideFirst(t *testing.T) {
	home := t.TempDir()
	override := filepath.Join(t.TempDir(), "custom-projects")
	t.Setenv("HOME", home)
	t.Setenv(codexHomeEnv, "")
	t.Setenv(projectsRootEnv, override)

	locator := NewSessionLocator()
	if got := locator.projectsRoot; got != override {
		t.Fatalf("projectsRoot = %q, want override %q", got, override)
	}

	roots := locator.TranscriptRoots()
	if len(roots) != 3 {
		t.Fatalf("expected override plus defaults, got %d roots", len(roots))
	}
	if roots[0].Host != HostOverride {
		t.Fatalf("expected first root host %q, got %q", HostOverride, roots[0].Host)
	}
	if roots[0].Path != override {
		t.Fatalf("override root = %q, want %q", roots[0].Path, override)
	}
	if !roots[0].ProjectHashed {
		t.Fatal("override root should preserve Claude project-hash behavior")
	}
}
