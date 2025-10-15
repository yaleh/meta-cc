package validation

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
	Tool     string                 `json:"tool"`
	Check    string                 `json:"check"`
	Status   string                 `json:"status"` // "PASS", "FAIL", "WARN"
	Message  string                 `json:"message"`
	Severity string                 `json:"severity"` // "ERROR", "WARNING", "INFO"
	Details  map[string]interface{} `json:"details,omitempty"`
}

// Report aggregates all validation results
type Report struct {
	TotalTools int      `json:"total_tools"`
	ChecksRun  int      `json:"checks_run"`
	Passed     int      `json:"passed"`
	Failed     int      `json:"failed"`
	Warnings   int      `json:"warnings"`
	Results    []Result `json:"results"`
	Summary    string   `json:"summary"`
}

// NewPassResult creates a passing result
func NewPassResult(tool, check string) Result {
	return Result{
		Tool:     tool,
		Check:    check,
		Status:   "PASS",
		Severity: "INFO",
	}
}

// NewFailResult creates a failing result
func NewFailResult(tool, check, message string, details map[string]interface{}) Result {
	return Result{
		Tool:     tool,
		Check:    check,
		Status:   "FAIL",
		Message:  message,
		Severity: "ERROR",
		Details:  details,
	}
}

// NewWarnResult creates a warning result
func NewWarnResult(tool, check, message string) Result {
	return Result{
		Tool:     tool,
		Check:    check,
		Status:   "WARN",
		Message:  message,
		Severity: "WARNING",
	}
}
