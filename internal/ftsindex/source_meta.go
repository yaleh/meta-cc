package ftsindex

import (
	"encoding/json"
	"os"

	"github.com/yaleh/meta-cc/internal/conversation"
)

// DefaultSourceMeta resolves SourceMeta for a session using the same
// Extensions-embedded file path each provider adapter already records
// (claudeprovider.FilePath / codexprovider.RolloutPath — DIR-028/030), then
// os.Stat's it for size/mtime. When no local file path is available (e.g. a
// Codex app-server-only thread with no rollout file, or the stat fails),
// it falls back to session.UpdatedAt as the invalidation fingerprint
// instead — see SourceMeta.unchanged.
func DefaultSourceMeta(session conversation.Session) SourceMeta {
	var ext struct {
		Path        string `json:"path"`
		RolloutPath string `json:"rollout_path"`
	}
	_ = json.Unmarshal(session.Extensions, &ext)
	path := ext.Path
	if session.Provider == conversation.ProviderCodex {
		path = ext.RolloutPath
	}

	if path != "" {
		if info, err := os.Stat(path); err == nil {
			return SourceMeta{Path: path, Size: info.Size(), ModTime: info.ModTime(), UpdatedAt: session.UpdatedAt}
		}
	}
	return SourceMeta{UpdatedAt: session.UpdatedAt}
}
