// Package appserver implements a minimal client for the `codex app-server`
// stdio protocol: a newline-delimited JSON-RPC-ish stream where requests and
// responses carry an "id" and notifications omit it. Only the read-only
// stable surface needed by DIR-029 is modeled — `initialize`, `thread/list`,
// and `thread/read` — based on the schema Codex CLI 0.145 emits via
// `codex app-server generate-json-schema` and an empirical handshake against
// a scratch CODEX_HOME (see docs/reference/codex-app-server.md). Everything
// else (thread/start, approvals, mutation methods, experimental APIs) is
// deliberately out of scope: this package never issues a request that could
// create, resume, mutate, archive, or delete a Codex thread.
package appserver

import "encoding/json"

// Envelope is one newline-delimited frame exchanged with `codex app-server`.
// A request/response carries a non-empty ID; a server notification carries
// Method+Params but no ID.
type Envelope struct {
	ID     json.RawMessage `json:"id,omitempty"`
	Method string          `json:"method,omitempty"`
	Params json.RawMessage `json:"params,omitempty"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  *RPCError       `json:"error,omitempty"`
}

// IsNotification reports whether this envelope is a server-to-client
// notification (no id) rather than a response to a pending request.
func (e Envelope) IsNotification() bool {
	return len(e.ID) == 0 && e.Method != ""
}

// RPCError is the error shape `codex app-server` returns for a failed
// request.
type RPCError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

func (e *RPCError) Error() string { return e.Message }

// ClientInfo identifies this client during the initialize handshake.
type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// InitializeParams is the request payload for the "initialize" method,
// which must be the first request on a new connection.
type InitializeParams struct {
	ClientInfo ClientInfo `json:"clientInfo"`
}

// InitializeResult is the server's handshake response.
type InitializeResult struct {
	CodexHome      string `json:"codexHome"`
	PlatformFamily string `json:"platformFamily"`
	PlatformOS     string `json:"platformOs"`
	UserAgent      string `json:"userAgent"`
}

// ThreadListParams is the request payload for "thread/list". Fields are
// pointers/explicit slices (rather than Go zero values) so callers can
// distinguish "omitted" (server default) from "explicitly set to the zero
// value" — this matters because the server's own unqualified default for
// SourceKinds is "interactive sources only", not "all sources" (see
// docs/reference/codex-app-server.md's "Default-filter pitfall" note).
type ThreadListParams struct {
	SourceKinds    []string `json:"sourceKinds,omitempty"`
	Archived       *bool    `json:"archived,omitempty"`
	Cursor         string   `json:"cursor,omitempty"`
	CWD            []string `json:"cwd,omitempty"`
	Limit          *uint32  `json:"limit,omitempty"`
	ModelProviders []string `json:"modelProviders"`
	UseStateDBOnly bool     `json:"useStateDbOnly,omitempty"`
}

// ThreadListResult is the response payload for "thread/list".
type ThreadListResult struct {
	Data            []Thread `json:"data"`
	NextCursor      *string  `json:"nextCursor"`
	BackwardsCursor *string  `json:"backwardsCursor"`
}

// ThreadReadParams is the request payload for "thread/read".
type ThreadReadParams struct {
	ThreadID     string `json:"threadId"`
	IncludeTurns bool   `json:"includeTurns"`
}

// ThreadReadResult is the response payload for "thread/read".
type ThreadReadResult struct {
	Thread Thread `json:"thread"`
}

// Thread mirrors the subset of the app-server "Thread" object DIR-029 maps
// into conversation.Session. Turns is only populated when a request set
// IncludeTurns (thread/read) or on thread/resume|rollback|fork — thread/list
// always returns it empty.
type Thread struct {
	ID             string          `json:"id"`
	SessionID      string          `json:"sessionId"`
	CWD            string          `json:"cwd"`
	CreatedAt      int64           `json:"createdAt"`
	UpdatedAt      int64           `json:"updatedAt"`
	ModelProvider  string          `json:"modelProvider"`
	CLIVersion     string          `json:"cliVersion"`
	Name           *string         `json:"name"`
	Preview        string          `json:"preview"`
	Source         json.RawMessage `json:"source"`
	ParentThreadID *string         `json:"parentThreadId"`
	Turns          []Turn          `json:"turns"`
}

// Turn mirrors the app-server "Turn" object.
type Turn struct {
	ID          string       `json:"id"`
	Status      string       `json:"status"`
	StartedAt   *int64       `json:"startedAt"`
	CompletedAt *int64       `json:"completedAt"`
	Items       []ThreadItem `json:"items"`
}

// ThreadItem is a raw, not-yet-typed app-server thread item. The wire shape
// is a `oneOf` tagged by "type" with ~15 variants (userMessage,
// agentMessage, commandExecution, fileChange, mcpToolCall, webSearch,
// reasoning, contextCompaction, ...); rather than modeling every variant as
// a distinct Go type, this package keeps the envelope generic and lets
// map.go extract only the fields it recognizes, preserving the rest via
// Raw for conversation.NewRawItem. See docs/reference/codex-app-server.md
// for which item types currently get dedicated mapping.
type ThreadItem struct {
	Type string          `json:"type"`
	Raw  json.RawMessage `json:"-"`
}

// UnmarshalJSON captures both the discriminant Type and the full raw
// payload, so map.go can decode type-specific fields itself without a
// second round-trip through the wire.
func (t *ThreadItem) UnmarshalJSON(data []byte) error {
	var head struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &head); err != nil {
		return err
	}
	t.Type = head.Type
	t.Raw = append(json.RawMessage(nil), data...)
	return nil
}
