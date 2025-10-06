package output

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
)

func TestOutputError_JSONL(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := errors.New("test error")
	exitErr := OutputError(err, ErrInvalidArgument, "jsonl")

	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	os.Stdout = oldStdout

	output := buf.String()

	// Verify JSON structure
	var errOutput ErrorOutput
	if jsonErr := json.Unmarshal([]byte(output), &errOutput); jsonErr != nil {
		t.Fatalf("output is not valid JSON: %v\nOutput: %s", jsonErr, output)
	}

	if errOutput.Error != "test error" {
		t.Errorf("expected error='test error', got '%s'", errOutput.Error)
	}

	if errOutput.Code != ErrInvalidArgument {
		t.Errorf("expected code='INVALID_ARGUMENT', got '%s'", errOutput.Code)
	}

	// Verify helpful message is included
	if errOutput.Message == "" {
		t.Error("expected helpful message, got empty string")
	}

	// Verify exit code
	if exitCodeErr, ok := exitErr.(*ExitCodeError); ok {
		if exitCodeErr.Code != ExitError {
			t.Errorf("expected exit code %d, got %d", ExitError, exitCodeErr.Code)
		}
	} else {
		t.Error("expected ExitCodeError")
	}
}

func TestOutputError_TSV(t *testing.T) {
	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	err := errors.New("test error")
	OutputError(err, ErrInvalidArgument, "tsv")

	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	os.Stderr = oldStderr

	output := buf.String()

	// Verify stderr output
	if !strings.Contains(output, "Error: test error") {
		t.Errorf("expected stderr to contain error message, got: %s", output)
	}

	if !strings.Contains(output, "INVALID_ARGUMENT") {
		t.Errorf("expected stderr to contain error code, got: %s", output)
	}

	// Verify suggestion is included
	if !strings.Contains(output, "Suggestion:") {
		t.Errorf("expected stderr to contain suggestion, got: %s", output)
	}
}

func TestOutputError_SessionNotFound(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := errors.New("session not found")
	OutputError(err, ErrSessionNotFound, "jsonl")

	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	os.Stdout = oldStdout

	output := buf.String()

	var errOutput ErrorOutput
	if jsonErr := json.Unmarshal([]byte(output), &errOutput); jsonErr != nil {
		t.Fatalf("output is not valid JSON: %v", jsonErr)
	}

	// Verify suggestion for session not found
	if !strings.Contains(errOutput.Message, "--session") || !strings.Contains(errOutput.Message, "--project") {
		t.Errorf("expected session not found suggestion, got: %s", errOutput.Message)
	}
}

func TestOutputError_NoResults(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := errors.New("no matching results")
	exitErr := OutputError(err, ErrNoResults, "jsonl")

	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	os.Stdout = oldStdout

	// Verify exit code is 2 for no results
	if exitCodeErr, ok := exitErr.(*ExitCodeError); ok {
		if exitCodeErr.Code != ExitNoResults {
			t.Errorf("expected exit code %d for no results, got %d", ExitNoResults, exitCodeErr.Code)
		}
	} else {
		t.Error("expected ExitCodeError")
	}
}

func TestWarnNoResults_JSONL(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	exitErr := WarnNoResults("jsonl")

	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	os.Stdout = oldStdout

	output := strings.TrimSpace(buf.String())

	// Verify empty array output
	if output != "[]" {
		t.Errorf("expected '[]', got '%s'", output)
	}

	// Verify exit code
	if exitCodeErr, ok := exitErr.(*ExitCodeError); ok {
		if exitCodeErr.Code != ExitNoResults {
			t.Errorf("expected exit code %d, got %d", ExitNoResults, exitCodeErr.Code)
		}
	} else {
		t.Error("expected ExitCodeError")
	}
}

func TestWarnNoResults_TSV(t *testing.T) {
	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	// Capture stdout to ensure it's empty
	oldStdout := os.Stdout
	rStdout, wStdout, _ := os.Pipe()
	os.Stdout = wStdout

	WarnNoResults("tsv")

	wStdout.Close()
	var bufStdout bytes.Buffer
	bufStdout.ReadFrom(rStdout)
	os.Stdout = oldStdout

	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	os.Stderr = oldStderr

	stderrOutput := buf.String()
	stdoutOutput := bufStdout.String()

	// Verify stderr warning
	if !strings.Contains(stderrOutput, "Warning: No results found") {
		t.Errorf("expected warning message in stderr, got: %s", stderrOutput)
	}

	// Verify stdout is empty for TSV
	if strings.TrimSpace(stdoutOutput) != "" {
		t.Errorf("expected empty stdout for TSV, got: %s", stdoutOutput)
	}
}

func TestOutputError_DefaultFormat(t *testing.T) {
	// Capture stderr for default format
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	err := errors.New("test error")
	OutputError(err, ErrInvalidArgument, "unknown")

	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	os.Stderr = oldStderr

	output := buf.String()

	// Verify fallback to stderr
	if !strings.Contains(output, "Error: test error") {
		t.Errorf("expected stderr fallback, got: %s", output)
	}
}

func TestErrorCode_Values(t *testing.T) {
	// Test that error codes have expected values
	tests := []struct {
		code     ErrorCode
		expected string
	}{
		{ErrInvalidArgument, "INVALID_ARGUMENT"},
		{ErrSessionNotFound, "SESSION_NOT_FOUND"},
		{ErrParseError, "PARSE_ERROR"},
		{ErrFilterError, "FILTER_ERROR"},
		{ErrNoResults, "NO_RESULTS"},
		{ErrInternalError, "INTERNAL_ERROR"},
	}

	for _, tt := range tests {
		if string(tt.code) != tt.expected {
			t.Errorf("expected error code %s, got %s", tt.expected, tt.code)
		}
	}
}
