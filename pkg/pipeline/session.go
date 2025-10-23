package pipeline

import (
	"fmt"

	"github.com/yaleh/meta-cc/internal/locator"
	"github.com/yaleh/meta-cc/internal/parser"
)

// SessionPipeline encapsulates session data processing flow.
// It abstracts the common pattern: locate → load → extract → process.
type SessionPipeline struct {
	opts      GlobalOptions         // Pipeline configuration
	session   string                // Loaded session identifier/path
	entries   []parser.SessionEntry // Parsed session entries
	turnIndex map[string]int        // Cached UUID → turn index mapping
}

// NewSessionPipeline creates a new pipeline instance.
func NewSessionPipeline(opts GlobalOptions) *SessionPipeline {
	return &SessionPipeline{
		opts:      opts,
		turnIndex: make(map[string]int),
	}
}

// Load locates and loads session JSONL entries according to the configured options.
// Supports both session-level and project-level loading.
func (p *SessionPipeline) Load(loadOpts LoadOptions) error {
	loc := locator.NewSessionLocator()

	// Reset cached state on each load attempt
	p.entries = nil
	p.session = ""
	p.turnIndex = make(map[string]int)

	shouldLoadAllSessions := p.opts.ProjectPath != "" && !p.opts.SessionOnly && p.opts.SessionID == ""

	if shouldLoadAllSessions {
		sessionPaths, err := loc.AllSessionsFromProject(p.opts.ProjectPath)
		if err != nil {
			return fmt.Errorf("failed to locate project sessions: %w", err)
		}

		var allEntries []parser.SessionEntry
		for _, sessionPath := range sessionPaths {
			sessionParser := parser.NewSessionParser(sessionPath)
			entries, parseErr := sessionParser.ParseEntries()
			if parseErr != nil {
				return fmt.Errorf("JSONL parsing failed for %s: %w", sessionPath, parseErr)
			}
			allEntries = append(allEntries, entries...)
		}

		p.entries = allEntries
		p.session = fmt.Sprintf("<project:%s (%d sessions)>", p.opts.ProjectPath, len(sessionPaths))
		return nil
	}

	sessionPath, err := loc.Locate(locator.LocateOptions{
		SessionID:   p.opts.SessionID,
		ProjectPath: p.opts.ProjectPath,
		SessionOnly: p.opts.SessionOnly,
	})
	if err != nil {
		return fmt.Errorf("session location failed: %w", err)
	}

	sessionParser := parser.NewSessionParser(sessionPath)
	entries, err := sessionParser.ParseEntries()
	if err != nil {
		return fmt.Errorf("JSONL parsing failed: %w", err)
	}

	if loadOpts.Validate && len(entries) == 0 {
		return fmt.Errorf("session file is empty or contains no valid entries")
	}

	p.entries = entries
	p.session = sessionPath

	return nil
}

// ExtractToolCalls extracts all tool calls from the currently loaded entries.
func (p *SessionPipeline) ExtractToolCalls() []parser.ToolCall {
	if len(p.entries) == 0 {
		return []parser.ToolCall{}
	}
	return parser.ExtractToolCalls(p.entries)
}

// BuildTurnIndex creates (and caches) a UUID → turn sequence mapping.
func (p *SessionPipeline) BuildTurnIndex() map[string]int {
	if len(p.turnIndex) > 0 {
		return p.turnIndex
	}

	for i, entry := range p.entries {
		p.turnIndex[entry.UUID] = i
	}

	return p.turnIndex
}

// SessionPath returns the identifier/path of the loaded session(s).
func (p *SessionPipeline) SessionPath() string {
	return p.session
}

// EntryCount returns the number of parsed entries currently loaded.
func (p *SessionPipeline) EntryCount() int {
	return len(p.entries)
}

// Entries returns the parsed session entries.
func (p *SessionPipeline) Entries() []parser.SessionEntry {
	return p.entries
}
