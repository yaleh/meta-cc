package rawfiles

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yaleh/meta-cc/internal/config"
	"github.com/yaleh/meta-cc/internal/conversation"
)

// TestParseProviderFilter_OmittedFollowsHostDefault proves the raw-file
// selection layer resolves an empty provider name through the same process
// host default as the MCP handlers (DIR-073), never a hard-coded claude.
func TestParseProviderFilter_OmittedFollowsHostDefault(t *testing.T) {
	tests := []struct {
		host string
		want []conversation.ProviderID
	}{
		{"codex", []conversation.ProviderID{conversation.ProviderCodex}},
		{"claude", []conversation.ProviderID{conversation.ProviderClaude}},
	}
	for _, tt := range tests {
		restore := config.SwapProcessDefault(tt.host)
		got, err := ParseProviderFilter("")
		require.NoError(t, err)
		require.Equal(t, tt.want, got, "host=%s", tt.host)
		restore()
	}
}

// TestParseProviderFilter_ExplicitUnchanged proves explicit values never
// consult the host default.
func TestParseProviderFilter_ExplicitUnchanged(t *testing.T) {
	restore := config.SwapProcessDefault("codex")
	defer restore()

	got, err := ParseProviderFilter("claude")
	require.NoError(t, err)
	require.Equal(t, []conversation.ProviderID{conversation.ProviderClaude}, got)

	got, err = ParseProviderFilter("all")
	require.NoError(t, err)
	require.Equal(t, []conversation.ProviderID{conversation.ProviderClaude, conversation.ProviderCodex}, got)

	_, err = ParseProviderFilter("openai")
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid provider")
}
