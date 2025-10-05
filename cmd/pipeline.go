package cmd

import (
	"fmt"

	"github.com/yale/meta-cc/internal/locator"
	"github.com/yale/meta-cc/internal/parser"
)

// GlobalOptions contains global CLI flags shared across commands
type GlobalOptions struct {
	SessionID   string
	ProjectPath string
	SessionOnly bool
}

// LoadOptions controls session loading behavior
type LoadOptions struct {
	AutoDetect bool
	Validate   bool
}

// SessionPipeline encapsulates session data processing flow
type SessionPipeline struct {
	opts      GlobalOptions
	session   string
	entries   []parser.SessionEntry
	turnIndex map[string]int
}

// NewSessionPipeline creates a new pipeline instance
func NewSessionPipeline(opts GlobalOptions) *SessionPipeline {
	return &SessionPipeline{
		opts:      opts,
		turnIndex: make(map[string]int),
	}
}

// Load locates and loads the session JSONL file
func (p *SessionPipeline) Load(loadOpts LoadOptions) error {
	// 1. Locate session file
	loc := locator.NewSessionLocator()

	sessionPath, err := loc.Locate(locator.LocateOptions{
		SessionID:   p.opts.SessionID,
		ProjectPath: p.opts.ProjectPath,
		SessionOnly: p.opts.SessionOnly,
	})
	if err != nil {
		return fmt.Errorf("session location failed: %w", err)
	}

	p.session = sessionPath

	// 2. Parse JSONL
	sessionParser := parser.NewSessionParser(sessionPath)
	p.entries, err = sessionParser.ParseEntries()
	if err != nil {
		return fmt.Errorf("JSONL parsing failed: %w", err)
	}

	return nil
}

// ExtractToolCalls extracts all tool calls from entries
func (p *SessionPipeline) ExtractToolCalls() []parser.ToolCall {
	return parser.ExtractToolCalls(p.entries)
}

// BuildTurnIndex creates turn_id → turn_sequence mapping
func (p *SessionPipeline) BuildTurnIndex() map[string]int {
	if len(p.turnIndex) > 0 {
		return p.turnIndex // cached
	}

	for i, entry := range p.entries {
		p.turnIndex[entry.UUID] = i
	}

	return p.turnIndex
}

// SessionPath returns the loaded session file path
func (p *SessionPipeline) SessionPath() string {
	return p.session
}

// EntryCount returns the number of entries loaded
func (p *SessionPipeline) EntryCount() int {
	return len(p.entries)
}

// GetEntries returns all parsed entries (for advanced use cases)
func (p *SessionPipeline) GetEntries() []parser.SessionEntry {
	return p.entries
}
