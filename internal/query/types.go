package query

// ContextQuery represents context query results
type ContextQuery struct {
	ErrorSignature string              `json:"error_signature,omitempty"`
	Occurrences    []ContextOccurrence `json:"occurrences"`
}

// ContextOccurrence represents a single occurrence with context
type ContextOccurrence struct {
	Turn          int           `json:"turn"`
	ContextBefore []TurnPreview `json:"context_before"`
	ErrorTurn     ErrorDetail   `json:"error_turn"`
	ContextAfter  []TurnPreview `json:"context_after"`
}

// TurnPreview represents a brief summary of a turn
type TurnPreview struct {
	Turn      int      `json:"turn"`
	Role      string   `json:"role"`
	Preview   string   `json:"preview,omitempty"`
	Tools     []string `json:"tools,omitempty"`
	Timestamp int64    `json:"timestamp"`
}

// ErrorDetail represents detailed error information
type ErrorDetail struct {
	Turn      int    `json:"turn"`
	Tool      string `json:"tool"`
	Command   string `json:"command,omitempty"`
	Error     string `json:"error"`
	File      string `json:"file,omitempty"`
	Timestamp int64  `json:"timestamp"`
}

// FileAccessQuery represents file access history query results
type FileAccessQuery struct {
	File          string            `json:"file"`
	TotalAccesses int               `json:"total_accesses"`
	Operations    map[string]int    `json:"operations"`
	Timeline      []FileAccessEvent `json:"timeline"`
	TimeSpanMin   int               `json:"time_span_minutes"`
}

// FileAccessEvent represents a single file access event
type FileAccessEvent struct {
	Turn      int    `json:"turn"`
	Action    string `json:"action"` // Read/Edit/Write
	Timestamp int64  `json:"timestamp"`
}

// ToolSequenceQuery represents tool sequence pattern query results
type ToolSequenceQuery struct {
	Sequences []SequencePattern `json:"sequences"`
}

// SequencePattern represents a repeated tool call sequence
type SequencePattern struct {
	Pattern      string               `json:"pattern"`
	Count        int                  `json:"count"`
	Occurrences  []SequenceOccurrence `json:"occurrences"`
	TimeSpanMin  int                  `json:"time_span_minutes"`
}

// SequenceOccurrence represents a single occurrence of a sequence
type SequenceOccurrence struct {
	StartTurn int `json:"start_turn"`
	EndTurn   int `json:"end_turn"`
}
