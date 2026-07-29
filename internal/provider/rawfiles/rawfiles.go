// Package rawfiles provides a provider-aware raw-file selection service shared
// by the MCP convenience query pipeline (internal/mcp/executor) and the Stage 1
// discovery tools (internal/mcp/query: get_session_directory, get_session_metadata).
//
// Unlike internal/provider/records (which normalizes Claude and Codex sessions
// into one common jq-queryable schema for the convenience query tools), this
// package resolves the *raw* on-disk files backing each provider's sessions —
// the shape that Stage 1 discovery tools hand back to Claude for stage2 jq
// queries. Claude and Codex raw files never share a schema, so callers must
// keep per-provider results separate.
package rawfiles

import (
	"context"
	"fmt"

	"github.com/yaleh/meta-cc/internal/config"
	"github.com/yaleh/meta-cc/internal/conversation"
	"github.com/yaleh/meta-cc/internal/locator"
	providerpkg "github.com/yaleh/meta-cc/internal/provider"
	claudeprovider "github.com/yaleh/meta-cc/internal/provider/claude"
	codexprovider "github.com/yaleh/meta-cc/internal/provider/codex"
	providerrecords "github.com/yaleh/meta-cc/internal/provider/records"
)

// File describes a single raw session-log file selected for a specific
// provider, as opposed to a normalized record.
type File struct {
	Path      string                  `json:"path"`
	Provider  conversation.ProviderID `json:"provider"`
	SessionID string                  `json:"session_id"`
}

// NewRegistry builds the standard Claude+Codex provider registry rooted at
// projectPath. This is the single construction point shared by the
// convenience query pipeline and the Stage 1 discovery tools so that both
// describe the same underlying corpus for a given provider.
func NewRegistry(projectPath string) *providerpkg.Registry {
	return providerpkg.NewRegistry(
		claudeprovider.NewProvider(locator.NewSessionLocator(), projectPath),
		codexprovider.NewProvider(locator.NewCodexLocator()),
	)
}

// ParseProviderFilter parses a "provider" tool argument into a filter list of
// conversation.ProviderID. An empty string resolves to the process host
// default (DIR-073: config.OmittedProviderDefault — claude for a Claude Code
// launched server, codex for a Codex launched one), never a hard-coded
// provider. Invalid provider names return an actionable error rather than
// silently falling back to Claude.
func ParseProviderFilter(providerName string) ([]conversation.ProviderID, error) {
	if providerName == "" {
		providerName = config.OmittedProviderDefault()
	}
	switch providerName {
	case "claude":
		return []conversation.ProviderID{conversation.ProviderClaude}, nil
	case "codex":
		return []conversation.ProviderID{conversation.ProviderCodex}, nil
	case "all":
		return []conversation.ProviderID{conversation.ProviderClaude, conversation.ProviderCodex}, nil
	default:
		return nil, fmt.Errorf("invalid provider %q: must be \"claude\", \"codex\", or \"all\"", providerName)
	}
}

// SelectCodexFiles resolves the raw Codex rollout files available for the
// given scope/project. It fails closed — returning an actionable error —
// when the Codex provider is unavailable (no session state found) or when no
// Codex sessions match the requested scope, rather than silently returning
// an empty result that could be mistaken for "Codex has no data".
func SelectCodexFiles(ctx context.Context, registry *providerpkg.Registry, scope, projectPath string) ([]File, error) {
	providers := registry.Providers([]conversation.ProviderID{conversation.ProviderCodex})
	if len(providers) == 0 {
		return nil, fmt.Errorf("codex provider not registered")
	}
	p := providers[0]

	if !p.IsAvailable(ctx) {
		return nil, fmt.Errorf("codex provider unavailable: no Codex session state found at %s (set META_CC_CODEX_ROOT or CODEX_HOME to override)", locator.NewCodexLocator().SQLiteDB())
	}

	sessions, err := p.ListSessions(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list codex sessions: %w", err)
	}

	sessions = providerrecords.FilterSessionsForScope(sessions, scope, projectPath, conversation.ProviderCodex)
	if len(sessions) == 0 {
		return nil, fmt.Errorf("no codex sessions found for project %q (scope=%s)", projectPath, scope)
	}

	files := make([]File, 0, len(sessions))
	for _, session := range sessions {
		path, pathErr := codexprovider.RolloutPath(session)
		if pathErr != nil {
			return nil, fmt.Errorf("failed to resolve rollout path for codex session %s: %w", session.ID, pathErr)
		}
		files = append(files, File{Path: path, Provider: conversation.ProviderCodex, SessionID: session.ID})
	}
	return files, nil
}
