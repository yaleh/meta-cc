package validation

const (
	// Status constants
	StatusPass = "PASS"
	StatusFail = "FAIL"
	StatusWarn = "WARN"

	// Severity constants
	SeverityError   = "ERROR"
	SeverityWarning = "WARNING"
	SeverityInfo    = "INFO"
)

// Tool represents a parsed MCP tool definition
type Tool struct {
	Name        string
	Description string
	InputSchema InputSchema
}

// InputSchema represents the tool's input parameters
type InputSchema struct {
	Type       string
	Properties map[string]Property
	Required   []string
}

// Property represents a parameter definition
type Property struct {
	Type        string
	Description string
}

// Result represents a validation result for a specific check
type Result struct {
	Details  map[string]interface{} `json:"details,omitempty"`
	Tool     string                 `json:"tool"`
	Check    string                 `json:"check"`
	Status   string                 `json:"status"` // "PASS", "FAIL", "WARN"
	Message  string                 `json:"message"`
	Severity string                 `json:"severity"` // "ERROR", "WARNING", "INFO"
}

// Report aggregates all validation results
type Report struct {
	Results    []Result `json:"results"`
	Summary    string   `json:"summary"`
	TotalTools int      `json:"total_tools"`
	ChecksRun  int      `json:"checks_run"`
	Passed     int      `json:"passed"`
	Failed     int      `json:"failed"`
	Warnings   int      `json:"warnings"`
}

// NewPassResult creates a passing result
func NewPassResult(tool, check string) Result {
	return Result{
		Tool:     tool,
		Check:    check,
		Status:   StatusPass,
		Severity: SeverityInfo,
	}
}

// NewFailResult creates a failing result
func NewFailResult(tool, check, message string, details map[string]interface{}) Result {
	return Result{
		Tool:     tool,
		Check:    check,
		Status:   StatusFail,
		Message:  message,
		Severity: SeverityError,
		Details:  details,
	}
}

// NewWarnResult creates a warning result
func NewWarnResult(tool, check, message string) Result {
	return Result{
		Tool:     tool,
		Check:    check,
		Status:   StatusWarn,
		Message:  message,
		Severity: SeverityWarning,
	}
}
