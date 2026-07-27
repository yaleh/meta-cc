package ftsindex

import (
	"os"

	"github.com/yaleh/meta-cc/internal/conversation"
	claudeprovider "github.com/yaleh/meta-cc/internal/provider/claude"
	codexprovider "github.com/yaleh/meta-cc/internal/provider/codex"
)

// DefaultSourceMeta resolves SourceMeta for a session using the same
// Extensions-embedded file path each provider adapter already records
// (claudeprovider.FilePath / codexprovider.RolloutPath — DIR-028/030), then
// os.Stat's it for size/mtime. When no local file path is available (e.g. a
// Codex app-server-only thread with no rollout file, or the stat fails),
// it falls back to session.UpdatedAt as the invalidation fingerprint
// instead — see SourceMeta.unchanged.
func DefaultSourceMeta(session conversation.Session) SourceMeta {
	var path string
	switch session.Provider {
	case conversation.ProviderClaude:
		path, _ = claudeprovider.FilePath(session)
	case conversation.ProviderCodex:
		path, _ = codexprovider.RolloutPath(session)
	}

	if path != "" {
		if info, err := os.Stat(path); err == nil {
			return SourceMeta{Path: path, Size: info.Size(), ModTime: info.ModTime(), UpdatedAt: session.UpdatedAt}
		}
	}
	return SourceMeta{UpdatedAt: session.UpdatedAt}
}
