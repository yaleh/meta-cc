package output

import (
	"encoding/json"
	"fmt"
	"os"
)

// ErrorCode represents standard error codes
type ErrorCode string

const (
	ErrInvalidArgument ErrorCode = "INVALID_ARGUMENT"
	ErrSessionNotFound ErrorCode = "SESSION_NOT_FOUND"
	ErrParseError      ErrorCode = "PARSE_ERROR"
	ErrFilterError     ErrorCode = "FILTER_ERROR"
	ErrNoResults       ErrorCode = "NO_RESULTS"
	ErrInternalError   ErrorCode = "INTERNAL_ERROR"
)

// ErrorOutput represents a structured error output
type ErrorOutput struct {
	Error   string    `json:"error"`
	Code    ErrorCode `json:"code"`
	Message string    `json:"message,omitempty"` // Additional context/suggestions
}

// OutputError outputs an error in the appropriate format
// Returns an ExitCodeError for proper exit code handling
func OutputError(err error, code ErrorCode, format string) error {
	errOutput := ErrorOutput{
		Error: err.Error(),
		Code:  code,
	}

	// Add helpful suggestions based on error code
	switch code {
	case ErrSessionNotFound:
		errOutput.Message = "Try specifying --session or --project flags"
	case ErrInvalidArgument:
		errOutput.Message = "Check command syntax with --help"
	case ErrFilterError:
		errOutput.Message = "Verify filter syntax (e.g., tool=Bash status=error)"
	}

	switch format {
	case "jsonl":
		// Output error as JSON object to stdout (valid JSONL)
		data, marshalErr := json.Marshal(errOutput)
		if marshalErr != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to marshal error: %v\n", marshalErr)
			return NewExitCodeError(ExitError, err.Error())
		}
		fmt.Println(string(data))

	case "tsv":
		// Output error message to stderr for TSV format
		if errOutput.Message != "" {
			fmt.Fprintf(os.Stderr, "Error: %s (code: %s)\nSuggestion: %s\n", err.Error(), code, errOutput.Message)
		} else {
			fmt.Fprintf(os.Stderr, "Error: %s (code: %s)\n", err.Error(), code)
		}

	default:
		// Fallback to stderr
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
	}

	// Determine exit code based on error type
	exitCode := ExitError
	if code == ErrNoResults {
		exitCode = ExitNoResults
	}

	return NewExitCodeError(exitCode, err.Error())
}

// WarnNoResults outputs a warning for no results (exit code 2)
// This is not an error, but informational
func WarnNoResults(format string) error {
	switch format {
	case "jsonl":
		// JSONL format: no results = no output (empty, no lines)
		// This conforms to JSONL spec: each line is a JSON object
		// No lines = no results
		// Warning message goes to stderr
		fmt.Fprintf(os.Stderr, "Warning: No results found\n")

	case "tsv":
		// Output warning to stderr, nothing to stdout
		fmt.Fprintf(os.Stderr, "Warning: No results found\n")

	default:
		fmt.Fprintf(os.Stderr, "Warning: No results found\n")
	}

	// Exit code 2 indicates no results (not an error, but informational)
	return NewExitCodeError(ExitNoResults, "no results found")
}
