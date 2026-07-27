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
	ID         string          `json:"id"`
	Provider   ProviderID      `json:"provider"`
	Title      string          `json:"title,omitempty"`
	CWD        string          `json:"cwd"`
	Model      string          `json:"model,omitempty"`
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
