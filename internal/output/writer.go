package output

import (
	"fmt"
	"io"
	"os"
)

// Output destinations
var (
	// DataWriter is the writer for command output data (default: stdout)
	DataWriter io.Writer = os.Stdout

	// LogWriter is the writer for diagnostic messages (default: stderr)
	LogWriter io.Writer = os.Stderr
)

// WriteData writes data to stdout (DataWriter)
// This should be used for all command output data (JSON, CSV, TSV, etc.)
func WriteData(format string, args ...interface{}) {
	fmt.Fprintf(DataWriter, format, args...)
}

// WriteLog writes log message to stderr (LogWriter)
// This should be used for diagnostic messages, warnings, and progress indicators
func WriteLog(format string, args ...interface{}) {
	fmt.Fprintf(LogWriter, format, args...)
}

// WriteLogLine writes log message with newline to stderr (LogWriter)
// This is a convenience function for log messages that should be on their own line
func WriteLogLine(format string, args ...interface{}) {
	fmt.Fprintf(LogWriter, format+"\n", args...)
}

// SetWriters sets custom writers for data and log output
// This is primarily useful for testing
func SetWriters(dataWriter, logWriter io.Writer) {
	DataWriter = dataWriter
	LogWriter = logWriter
}

// ResetWriters resets writers to default values (os.Stdout and os.Stderr)
// This should be called after tests that use custom writers
func ResetWriters() {
	DataWriter = os.Stdout
	LogWriter = os.Stderr
}
