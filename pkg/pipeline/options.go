package pipeline

// GlobalOptions contains global CLI flags used across commands
type GlobalOptions struct {
	SessionID   string // Explicit session UUID from --session flag
	ProjectPath string // Explicit project path from --project flag
	SessionOnly bool   // Force session-only mode (opt-out of project-level default)
}

// LoadOptions controls session loading behavior
type LoadOptions struct {
	AutoDetect bool // Whether to auto-detect session from current directory
	Validate   bool // Whether to validate session file after loading
}
