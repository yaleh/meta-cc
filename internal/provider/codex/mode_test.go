package codex

import "testing"

func TestParseModeDefaultsToAuto(t *testing.T) {
	for _, raw := range []string{"", "auto"} {
		mode, err := parseMode(raw)
		if err != nil {
			t.Fatalf("parseMode(%q): unexpected error: %v", raw, err)
		}
		if mode != ModeAuto {
			t.Fatalf("parseMode(%q) = %q, want %q", raw, mode, ModeAuto)
		}
	}
}

func TestParseModeExplicitValues(t *testing.T) {
	cases := map[string]Mode{
		"app_server": ModeAppServer,
		"files":      ModeFiles,
	}
	for raw, want := range cases {
		mode, err := parseMode(raw)
		if err != nil {
			t.Fatalf("parseMode(%q): unexpected error: %v", raw, err)
		}
		if mode != want {
			t.Fatalf("parseMode(%q) = %q, want %q", raw, mode, want)
		}
	}
}

func TestParseModeRejectsInvalidValue(t *testing.T) {
	if _, err := parseMode("bogus"); err == nil {
		t.Fatalf("expected error for invalid mode")
	}
}

func TestResolveModeReadsEnv(t *testing.T) {
	t.Setenv(backendModeEnvVar, "files")
	mode, err := ResolveMode()
	if err != nil {
		t.Fatalf("ResolveMode: %v", err)
	}
	if mode != ModeFiles {
		t.Fatalf("ResolveMode = %q, want %q", mode, ModeFiles)
	}
}
