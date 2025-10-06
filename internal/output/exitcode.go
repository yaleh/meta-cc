package output

// Standard Unix exit codes
const (
	ExitSuccess   = 0 // Command succeeded with results
	ExitError     = 1 // General error (parsing, I/O, etc.)
	ExitNoResults = 2 // Command succeeded but no results found
)

// ExitCodeError is a special error type that carries an exit code
// This allows commands to return specific exit codes through Cobra's error handling
type ExitCodeError struct {
	Code    int
	Message string
}

// Error implements the error interface
func (e *ExitCodeError) Error() string {
	return e.Message
}

// NewExitCodeError creates a new exit code error
func NewExitCodeError(code int, message string) *ExitCodeError {
	return &ExitCodeError{
		Code:    code,
		Message: message,
	}
}

// DetermineExitCode determines the appropriate exit code based on results and errors
// Priority: error > no results > success
func DetermineExitCode(hasResults bool, err error) int {
	if err != nil {
		return ExitError
	}
	if !hasResults {
		return ExitNoResults
	}
	return ExitSuccess
}
