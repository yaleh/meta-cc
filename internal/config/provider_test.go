package config

import (
	"strings"
	"testing"
)

// TestLoadProviderConfig is the DIR-073 table-driven contract for the
// central default-provider resolver: META_CC_HOST (injected by the plugin
// manifests) decides which corpus an omitted `provider` argument searches,
// unset means the documented standalone fallback (claude), and anything
// invalid fails Load() visibly instead of silently reading wrong history.
func TestLoadProviderConfig(t *testing.T) {
	tests := []struct {
		name        string
		hostEnv     string
		setEnv      bool
		wantHost    string
		wantDefault string
		wantErr     string // substring of the Load() error; empty means no error
	}{
		{name: "unset falls back to claude", setEnv: false, wantHost: "", wantDefault: "claude"},
		{name: "claude host", hostEnv: "claude", setEnv: true, wantHost: "claude", wantDefault: "claude"},
		{name: "codex host", hostEnv: "codex", setEnv: true, wantHost: "codex", wantDefault: "codex"},
		{name: "host value is normalized", hostEnv: "  Codex ", setEnv: true, wantHost: "codex", wantDefault: "codex"},
		{name: "upper case host", hostEnv: "CLAUDE", setEnv: true, wantHost: "claude", wantDefault: "claude"},
		{name: "invalid host fails visibly", hostEnv: "openai", setEnv: true, wantErr: "META_CC_HOST"},
		{name: "all is not a host", hostEnv: "all", setEnv: true, wantErr: "META_CC_HOST"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				t.Setenv(HostEnv, tt.hostEnv)
			} else {
				t.Setenv(HostEnv, "")
			}
			restore := SwapProcessDefault(StandaloneHostFallback)
			defer restore()

			cfg, err := Load()
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("Load() succeeded, want error containing %q", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("Load() error %q does not mention %q (must be actionable)", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Load() unexpected error: %v", err)
			}
			if cfg.Provider.Host != tt.wantHost {
				t.Errorf("Provider.Host = %q, want %q", cfg.Provider.Host, tt.wantHost)
			}
			if cfg.Provider.Default != tt.wantDefault {
				t.Errorf("Provider.Default = %q, want %q", cfg.Provider.Default, tt.wantDefault)
			}
			// Load() publishes the validated default as the process-wide
			// source of truth every provider-aware tool reads.
			if got := OmittedProviderDefault(); got != tt.wantDefault {
				t.Errorf("OmittedProviderDefault() = %q, want %q", got, tt.wantDefault)
			}
		})
	}
}

// TestOmittedProviderDefault_StandaloneFallback covers processes that never
// call Load() (library embedding, tests): the default is deterministic.
func TestOmittedProviderDefault_StandaloneFallback(t *testing.T) {
	restore := SwapProcessDefault(StandaloneHostFallback)
	defer restore()
	if got := OmittedProviderDefault(); got != "claude" {
		t.Fatalf("standalone fallback = %q, want %q", got, "claude")
	}
}

// TestSwapProcessDefault proves the test/override seam restores exactly.
func TestSwapProcessDefault(t *testing.T) {
	restore := SwapProcessDefault("codex")
	if got := OmittedProviderDefault(); got != "codex" {
		t.Fatalf("after swap = %q, want codex", got)
	}
	restore()
	if got := OmittedProviderDefault(); got != StandaloneHostFallback {
		t.Fatalf("after restore = %q, want %q", got, StandaloneHostFallback)
	}
}
