package output

import (
	"fmt"
	"testing"
)

func TestDetermineExitCode(t *testing.T) {
	tests := []struct {
		name       string
		hasResults bool
		err        error
		expected   int
	}{
		{
			name:       "success with results",
			hasResults: true,
			err:        nil,
			expected:   ExitSuccess,
		},
		{
			name:       "success without results",
			hasResults: false,
			err:        nil,
			expected:   ExitNoResults,
		},
		{
			name:       "error",
			hasResults: false,
			err:        fmt.Errorf("some error"),
			expected:   ExitError,
		},
		{
			name:       "error with results (still error)",
			hasResults: true,
			err:        fmt.Errorf("some error"),
			expected:   ExitError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetermineExitCode(tt.hasResults, tt.err)
			if result != tt.expected {
				t.Errorf("Expected exit code %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestExitCodeError(t *testing.T) {
	err := NewExitCodeError(ExitNoResults, "No results found")

	if err.Code != ExitNoResults {
		t.Errorf("Expected code %d, got %d", ExitNoResults, err.Code)
	}

	if err.Error() != "No results found" {
		t.Errorf("Expected message 'No results found', got '%s'", err.Error())
	}
}

func TestExitCodeError_ErrorInterface(t *testing.T) {
	// Verify ExitCodeError implements error interface
	var _ error = &ExitCodeError{}

	err := NewExitCodeError(ExitError, "test error")
	errString := err.Error()

	if errString != "test error" {
		t.Errorf("Expected 'test error', got '%s'", errString)
	}
}

func TestExitCodeConstants(t *testing.T) {
	// Verify exit code constants match Unix conventions
	if ExitSuccess != 0 {
		t.Errorf("ExitSuccess should be 0, got %d", ExitSuccess)
	}

	if ExitError != 1 {
		t.Errorf("ExitError should be 1, got %d", ExitError)
	}

	if ExitNoResults != 2 {
		t.Errorf("ExitNoResults should be 2, got %d", ExitNoResults)
	}
}

func TestDetermineExitCode_ErrorPriority(t *testing.T) {
	// Errors should always take priority over results
	result := DetermineExitCode(true, fmt.Errorf("error"))
	if result != ExitError {
		t.Errorf("Error should take priority over hasResults, expected %d, got %d", ExitError, result)
	}
}

func TestNewExitCodeError_DifferentCodes(t *testing.T) {
	tests := []struct {
		code    int
		message string
	}{
		{ExitSuccess, "success message"},
		{ExitError, "error message"},
		{ExitNoResults, "no results message"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("code_%d", tt.code), func(t *testing.T) {
			err := NewExitCodeError(tt.code, tt.message)

			if err.Code != tt.code {
				t.Errorf("Expected code %d, got %d", tt.code, err.Code)
			}

			if err.Message != tt.message {
				t.Errorf("Expected message '%s', got '%s'", tt.message, err.Message)
			}
		})
	}
}
