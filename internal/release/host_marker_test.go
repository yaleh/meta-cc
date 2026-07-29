package release

import (
	"path/filepath"
	"testing"

	"github.com/yaleh/meta-cc/internal/config"
)

// mcpManifestFile models the two shapes a shipped MCP manifest can take:
// the Claude plugin form ({"meta-cc": {...}}) and the Codex/general form
// ({"mcpServers": {"meta-cc": {...}}}).
type mcpManifestFile struct {
	Direct  mcpServerEntry            `json:"meta-cc"`
	Servers map[string]mcpServerEntry `json:"mcpServers"`
}

type mcpServerEntry struct {
	Command string            `json:"command"`
	Env     map[string]string `json:"env"`
}

func (m mcpManifestFile) server(t *testing.T, path string) mcpServerEntry {
	t.Helper()
	if m.Direct.Command != "" {
		return m.Direct
	}
	entry, ok := m.Servers["meta-cc"]
	if !ok || entry.Command == "" {
		t.Fatalf("%s: no meta-cc server entry found", path)
	}
	return entry
}

// TestPackagedManifestsInjectHostMarker asserts the DIR-073 contract that
// both packaged plugin manifests launch the same MCP binary with a validated
// META_CC_HOST marker — claude for the Claude Code plugin, codex for the
// Codex plugin — so the server can resolve an omitted `provider` to the
// host that launched it. A drift here silently revives the "Codex session
// sees only Claude history" defect.
func TestPackagedManifestsInjectHostMarker(t *testing.T) {
	root := repoRoot(t)

	tests := []struct {
		manifest string
		wantHost string
	}{
		{filepath.Join("plugin-src", ".mcp.json"), config.HostClaude},
		{filepath.Join("plugin-src", ".codex-mcp.json"), config.HostCodex},
	}

	for _, tt := range tests {
		t.Run(tt.manifest, func(t *testing.T) {
			path := filepath.Join(root, tt.manifest)
			var mf mcpManifestFile
			readJSON(t, path, &mf)
			entry := mf.server(t, path)

			got := entry.Env[config.HostEnv]
			if got != tt.wantHost {
				t.Fatalf("%s: env[%s] = %q, want %q — without this marker the server "+
					"cannot tell Claude Code from Codex and an omitted provider cannot "+
					"follow the launched host", tt.manifest, config.HostEnv, got, tt.wantHost)
			}
		})
	}
}
