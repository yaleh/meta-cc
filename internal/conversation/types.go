package conversation

import (
	"encoding/json"
	"time"
)

type ProviderID string

const (
	ProviderClaude ProviderID = "claude"
	ProviderCodex  ProviderID = "codex"
)

type Session struct {
	ID       string     `json:"id"`
	Provider ProviderID `json:"provider"`
	Title    string     `json:"title,omitempty"`
	CWD      string     `json:"cwd"`
	Model    string     `json:"model,omitempty"`

	// Metadata fields below (DIR-030) describe a session/thread WITHOUT
	// requiring its turns to be loaded — populated on a best-effort basis
	// per provider/backend (see internal/provider/codex/appserver/map.go
	// and internal/provider/codex/sqlite.go for Codex; the Claude provider
	// leaves most of them at their zero value since Claude sessions don't
	// carry this metadata). Zero values (ModelProvider == "", Archived ==
	// false, ...) mean "unknown/not applicable", not "filtered out" —
	// SessionFilter treats an unset filter dimension as "no constraint".
	ModelProvider  string `json:"model_provider,omitempty"`
	SourceKind     string `json:"source_kind,omitempty"`
	Status         string `json:"status,omitempty"` // "active" or "archived"; derived from Archived
	Archived       bool   `json:"archived,omitempty"`
	ParentThreadID string `json:"parent_thread_id,omitempty"`
	// Lineage records what is known about ParentThreadID's provenance
	// (DIR-032): a session with ParentThreadID == "" is only a confirmed
	// root when Lineage == LineageStatusRoot; LineageStatusUnknown means
	// spawn metadata simply wasn't available, so callers must not treat it
	// as a root. Left at LineageStatusUnspecified by adapters that haven't
	// been updated to populate it (treated the same as "unknown" by
	// lineage-aware callers).
	Lineage    LineageStatus `json:"lineage,omitempty"`
	IsSubagent bool          `json:"is_subagent,omitempty"`
	UpdatedAt  time.Time     `json:"updated_at,omitempty"`

	CreatedAt  time.Time       `json:"created_at"`
	TokenUsage TokenUsage      `json:"token_usage"`
	Turns      []Turn          `json:"turns,omitempty"`
	Extensions json.RawMessage `json:"extensions,omitempty"`
}

// Turn is the canonical, loss-minimizing representation of one exchange
// within a Thread (see DIR-028). UserText/AssistantText/ToolCalls remain
// for backward compatibility: they are a flattened compatibility
// projection that provider adapters derive from Items (not an independent
// parse), preserving the existing MCP query contract while Items carries
// the richer, order- and phase-preserving event stream.
type Turn struct {
	ID            string          `json:"id"`
	Status        TurnStatus      `json:"status,omitempty"`
	UserText      string          `json:"user_text,omitempty"`
	AssistantText string          `json:"assistant_text,omitempty"`
	ToolCalls     []ToolCall      `json:"tool_calls,omitempty"`
	Items         []Item          `json:"items,omitempty"`
	TokenUsage    TokenUsage      `json:"token_usage,omitempty"`
	Timestamp     time.Time       `json:"timestamp"`
	Extensions    json.RawMessage `json:"extensions,omitempty"`

	// Completeness reports whether this Turn is fully materialized content
	// or a placeholder (DIR-032). See HistoryCompleteness.IsFull — a caller
	// must check this before treating UserText/AssistantText/Items as
	// complete when it is anything other than full/unspecified.
	Completeness HistoryCompleteness `json:"completeness,omitempty"`
}

type ToolCall struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Input     json.RawMessage `json:"input"`
	Output    string          `json:"output,omitempty"`
	IsError   bool            `json:"is_error,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
}

// TokenUsage is the canonical, provider-neutral token accounting for a Turn
// or a Session (DIR-071). It distinguishes the five categories a source may
// report — input, cached input, output, reasoning output, and an opaque
// provider aggregate — instead of collapsing them into a single bucket.
//
// Field semantics (see docs/reference/jsonl-schema.md "Token Usage Model"):
//
//   - InputTokens / OutputTokens / CacheTokens: the precise per-category
//     counts. For a Turn these are accumulated across the turn's model calls
//     (DIR-065); for a Session they are a cumulative total.
//   - ReasoningOutputTokens: output tokens the model spent on reasoning
//     (chain-of-thought), reported separately by Codex as
//     reasoning_output_tokens. Preserved so reasoning-inclusive totals can be
//     reconciled; zero when the source does not report it.
//   - AggregateTokens / AggregateSource: an OPAQUE provider-reported total
//     whose composition (input+cached+output+reasoning) is defined by the
//     source, NOT by meta-cc. It is never an input-token count. Populated from
//     the Codex SQLite threads.tokens_used column; AggregateSource names that
//     provenance so the value is explicitly attributable rather than silently
//     reinterpreted as InputTokens. Zero/"" when no such aggregate exists.
type TokenUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	CacheTokens  int `json:"cache_tokens,omitempty"`

	ReasoningOutputTokens int `json:"reasoning_output_tokens,omitempty"`

	AggregateTokens int    `json:"aggregate_tokens,omitempty"`
	AggregateSource string `json:"aggregate_source,omitempty"`
}

// Aggregate provenance values for TokenUsage.AggregateSource (DIR-071). Each
// names the exact backend + field an opaque AggregateTokens total came from,
// so callers can reconcile it against the precise per-category counts instead
// of mistaking it for input usage.
const (
	// AggregateSourceCodexSQLite marks an aggregate read from the Codex
	// threads.tokens_used SQLite column — a provider-maintained running
	// total, not a per-category input count.
	AggregateSourceCodexSQLite = "codex_sqlite:threads.tokens_used"
)

// HasAny reports whether any token category (including reasoning output and
// an opaque aggregate) carries a non-zero value. It is the single, shared
// "does this usage carry anything?" predicate (DIR-071): every caller that
// decides whether to emit or retain token usage must use it so a turn that
// reports ONLY reasoning tokens, or a session that reports ONLY an opaque
// aggregate, is never mistaken for an empty usage.
func (u TokenUsage) HasAny() bool {
	return u.InputTokens != 0 ||
		u.OutputTokens != 0 ||
		u.CacheTokens != 0 ||
		u.ReasoningOutputTokens != 0 ||
		u.AggregateTokens != 0
}
