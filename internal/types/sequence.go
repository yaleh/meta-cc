package types

// SequencePattern represents a repeated tool call sequence
// This is the unified structure used by both query and analyzer packages
type SequencePattern struct {
	Pattern        string               `json:"pattern"`
	Length         int                  `json:"length,omitempty"` // Number of tools in pattern (optional)
	Count          int                  `json:"count"`            // Number of occurrences
	Occurrences    []SequenceOccurrence `json:"occurrences"`
	TimeSpanMin    int                  `json:"time_span_minutes"`
	SuccessRate    float64              `json:"success_rate,omitempty"`         // Success rate of sequence (optional)
	AvgDurationMin float64              `json:"avg_duration_minutes,omitempty"` // Average duration (optional)
	Context        string               `json:"context,omitempty"`              // Additional context (optional)
}

// SequenceOccurrence represents a single occurrence of a sequence
type SequenceOccurrence struct {
	StartTurn int              `json:"start_turn"`
	EndTurn   int              `json:"end_turn"`
	Tools     []ToolInSequence `json:"tools,omitempty"` // Detailed tool info (optional)
}

// ToolInSequence represents a tool call within a sequence
type ToolInSequence struct {
	Turn    int    `json:"turn"`
	Tool    string `json:"tool"`
	File    string `json:"file,omitempty"`
	Command string `json:"command,omitempty"`
}
