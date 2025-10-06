package cmd

import (
	"fmt"

	"github.com/yale/meta-cc/internal/locator"
	"github.com/yale/meta-cc/internal/parser"
)

// GlobalOptions contains global CLI flags shared across commands.
// These options control session location and scope for all meta-cc commands.
type GlobalOptions struct {
	SessionID   string // Specific session ID to analyze (e.g., "abc-123")
	ProjectPath string // Project path for project-level analysis (e.g., ".")
	SessionOnly bool   // Force session-only mode (disable project-level default)
}

// LoadOptions controls session loading behavior.
// These options configure how the pipeline loads and validates session data.
type LoadOptions struct {
	AutoDetect bool // Enable automatic session detection
	Validate   bool // Enable validation during loading
}

// SessionPipeline encapsulates the session data processing flow.
// It provides a unified interface for locating, loading, and processing
// Claude Code session history across single or multiple sessions.
//
// Pipeline workflow:
//  1. NewSessionPipeline(opts) - Create pipeline with options
//  2. Load(loadOpts) - Locate and parse JSONL file(s)
//  3. Extract data via GetEntries(), ExtractToolCalls(), BuildTurnIndex()
//
// The pipeline supports both session-level and project-level modes:
//   - Session-level: Load single session (latest or specified by ID)
//   - Project-level: Load all sessions from project directory
type SessionPipeline struct {
	opts      GlobalOptions         // Configuration options
	session   string                // Loaded session identifier or path
	entries   []parser.SessionEntry // Parsed session entries
	turnIndex map[string]int        // UUID -> turn sequence index (cached)
}

// NewSessionPipeline creates a new pipeline instance with the given options.
// The pipeline is initially empty; call Load() to populate it with session data.
func NewSessionPipeline(opts GlobalOptions) *SessionPipeline {
	return &SessionPipeline{
		opts:      opts,
		turnIndex: make(map[string]int),
	}
}

// Load locates and loads the session JSONL file(s) into the pipeline.
//
// Loading behavior depends on GlobalOptions:
//   - Project-level mode (ProjectPath set, SessionOnly=false):
//     Loads ALL sessions from the project directory for cross-session analysis
//   - Session-level mode (SessionOnly=true or SessionID specified):
//     Loads only the specified session or latest session from project
//
// The method uses SessionLocator to find session files, then parses them
// using SessionParser. All parsed entries are merged into p.entries.
//
// Returns error if session location fails or JSONL parsing fails.
func (p *SessionPipeline) Load(loadOpts LoadOptions) error {
	loc := locator.NewSessionLocator()

	// Determine if we should load multiple sessions (project-level mode)
	shouldLoadAllSessions := p.opts.ProjectPath != "" && !p.opts.SessionOnly && p.opts.SessionID == ""

	if shouldLoadAllSessions {
		// Project-level mode: load ALL sessions from project
		sessionPaths, err := loc.AllSessionsFromProject(p.opts.ProjectPath)
		if err != nil {
			return fmt.Errorf("failed to locate project sessions: %w", err)
		}

		// Parse and merge all sessions
		var allEntries []parser.SessionEntry
		for _, sessionPath := range sessionPaths {
			sessionParser := parser.NewSessionParser(sessionPath)
			entries, err := sessionParser.ParseEntries()
			if err != nil {
				return fmt.Errorf("JSONL parsing failed for %s: %w", sessionPath, err)
			}
			allEntries = append(allEntries, entries...)
		}

		p.entries = allEntries
		p.session = fmt.Sprintf("<project:%s (%d sessions)>", p.opts.ProjectPath, len(sessionPaths))

	} else {
		// Session-level mode: load single session (latest or specified)
		sessionPath, err := loc.Locate(locator.LocateOptions{
			SessionID:   p.opts.SessionID,
			ProjectPath: p.opts.ProjectPath,
			SessionOnly: p.opts.SessionOnly,
		})
		if err != nil {
			return fmt.Errorf("session location failed: %w", err)
		}

		p.session = sessionPath

		// Parse JSONL
		sessionParser := parser.NewSessionParser(sessionPath)
		p.entries, err = sessionParser.ParseEntries()
		if err != nil {
			return fmt.Errorf("JSONL parsing failed: %w", err)
		}
	}

	return nil
}

// ExtractToolCalls extracts all tool calls from loaded session entries.
// This method pairs tool_use and tool_result blocks to create complete ToolCall records
// containing both input parameters and execution results.
//
// Returns empty slice if no tool calls are found in the session(s).
func (p *SessionPipeline) ExtractToolCalls() []parser.ToolCall {
	return parser.ExtractToolCalls(p.entries)
}

// BuildTurnIndex creates a UUID -> turn_sequence mapping for all entries.
// The index maps each entry's UUID to its position in the conversation timeline.
// This is useful for temporal analysis and context retrieval.
//
// The index is cached after first call for performance.
// Returns the mapping where keys are UUIDs and values are 0-based indices.
func (p *SessionPipeline) BuildTurnIndex() map[string]int {
	if len(p.turnIndex) > 0 {
		return p.turnIndex // cached
	}

	for i, entry := range p.entries {
		p.turnIndex[entry.UUID] = i
	}

	return p.turnIndex
}

// SessionPath returns the loaded session identifier or file path.
// In session-level mode, returns the absolute path to the JSONL file.
// In project-level mode, returns a descriptive string like "<project:path (N sessions)>".
func (p *SessionPipeline) SessionPath() string {
	return p.session
}

// EntryCount returns the total number of entries loaded from session(s).
// In project-level mode, this is the sum of entries across all sessions.
func (p *SessionPipeline) EntryCount() int {
	return len(p.entries)
}

// GetEntries returns all parsed session entries.
// Entries are in chronological order based on the session file(s).
// This is the primary accessor for raw session data.
func (p *SessionPipeline) GetEntries() []parser.SessionEntry {
	return p.entries
}
