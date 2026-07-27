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

type TokenUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	CacheTokens  int `json:"cache_tokens,omitempty"`
}
