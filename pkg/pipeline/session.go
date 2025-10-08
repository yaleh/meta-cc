package pipeline

import (
	"fmt"

	"github.com/yaleh/meta-cc/internal/locator"
	"github.com/yaleh/meta-cc/internal/parser"
)

// SessionPipeline encapsulates session data processing
// Provides a unified interface for locating, loading, and extracting session data
type SessionPipeline struct {
	opts      GlobalOptions         // Global configuration options
	session   string                // Path to loaded session file
	entries   []parser.SessionEntry // Parsed session entries
	turnIndex map[string]int        // UUID → turn sequence mapping (cached)
}

// NewSessionPipeline creates a new pipeline instance
func NewSessionPipeline(opts GlobalOptions) *SessionPipeline {
	return &SessionPipeline{
		opts:      opts,
		turnIndex: make(map[string]int),
	}
}

// Load locates and loads the session JSONL file
// Supports three location methods (in priority order):
//  1. Explicit --session UUID
//  2. Explicit --project path
//  3. Auto-detect from current directory (when AutoDetect is true)
func (p *SessionPipeline) Load(loadOpts LoadOptions) error {
	// Step 1: Locate session file
	loc := locator.NewSessionLocator()

	sessionPath, err := loc.Locate(locator.LocateOptions{
		SessionID:   p.opts.SessionID,
		ProjectPath: p.opts.ProjectPath,
		SessionOnly: p.opts.SessionOnly,
	})

	if err != nil {
		// If auto-detect is enabled and location failed, try auto-detect
		if loadOpts.AutoDetect && p.opts.SessionID == "" && p.opts.ProjectPath == "" {
			// Auto-detect is already the default behavior in locator.Locate
			// when no explicit options are provided
			return fmt.Errorf("session location failed: %w", err)
		}
		return fmt.Errorf("session location failed: %w", err)
	}

	p.session = sessionPath

	// Step 2: Parse JSONL
	sessionParser := parser.NewSessionParser(sessionPath)
	p.entries, err = sessionParser.ParseEntries()
	if err != nil {
		return fmt.Errorf("JSONL parsing failed: %w", err)
	}

	// Step 3: Validate if requested
	if loadOpts.Validate {
		if len(p.entries) == 0 {
			return fmt.Errorf("session file is empty or contains no valid entries")
		}
	}

	return nil
}

// ExtractToolCalls extracts all tool calls from loaded entries
// Returns an empty slice if no entries are loaded
func (p *SessionPipeline) ExtractToolCalls() ([]parser.ToolCall, error) {
	if len(p.entries) == 0 {
		return []parser.ToolCall{}, nil
	}

	toolCalls := parser.ExtractToolCalls(p.entries)
	return toolCalls, nil
}

// BuildTurnIndex creates UUID → turn_sequence mapping
// Results are cached for subsequent calls (idempotent)
func (p *SessionPipeline) BuildTurnIndex() map[string]int {
	// Return cached index if already built
	if len(p.turnIndex) > 0 {
		return p.turnIndex
	}

	// Build index from entries
	for i, entry := range p.entries {
		p.turnIndex[entry.UUID] = i
	}

	return p.turnIndex
}

// SessionPath returns the loaded session file path
// Returns empty string if no session has been loaded
func (p *SessionPipeline) SessionPath() string {
	return p.session
}

// EntryCount returns the number of entries loaded
// Returns 0 if no session has been loaded
func (p *SessionPipeline) EntryCount() int {
	return len(p.entries)
}

// Entries returns the raw session entries
// Useful for advanced processing that needs direct access to entries
func (p *SessionPipeline) Entries() []parser.SessionEntry {
	return p.entries
}
