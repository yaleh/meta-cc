package output

import (
	"bytes"
	"testing"
)

func TestWriteData(t *testing.T) {
	var dataBuf, logBuf bytes.Buffer
	SetWriters(&dataBuf, &logBuf)
	defer ResetWriters() // Reset to defaults after test

	WriteData("data output")

	if dataBuf.String() != "data output" {
		t.Errorf("Expected 'data output', got '%s'", dataBuf.String())
	}

	if logBuf.String() != "" {
		t.Error("LogWriter should be empty")
	}
}

func TestWriteLog(t *testing.T) {
	var dataBuf, logBuf bytes.Buffer
	SetWriters(&dataBuf, &logBuf)
	defer ResetWriters()

	WriteLog("log message")

	if logBuf.String() != "log message" {
		t.Errorf("Expected 'log message', got '%s'", logBuf.String())
	}

	if dataBuf.String() != "" {
		t.Error("DataWriter should be empty")
	}
}

func TestWriteLogLine(t *testing.T) {
	var dataBuf, logBuf bytes.Buffer
	SetWriters(&dataBuf, &logBuf)
	defer ResetWriters()

	WriteLogLine("log line")

	if logBuf.String() != "log line\n" {
		t.Errorf("Expected 'log line\\n', got '%s'", logBuf.String())
	}

	if dataBuf.String() != "" {
		t.Error("DataWriter should be empty")
	}
}

func TestSetWriters(t *testing.T) {
	var customData, customLog bytes.Buffer
	SetWriters(&customData, &customLog)
	defer ResetWriters()

	if DataWriter != &customData {
		t.Error("DataWriter not set correctly")
	}

	if LogWriter != &customLog {
		t.Error("LogWriter not set correctly")
	}
}

func TestResetWriters(t *testing.T) {
	var customData, customLog bytes.Buffer
	SetWriters(&customData, &customLog)

	// Reset to defaults
	ResetWriters()

	// Verify default writers are restored (should be os.Stdout and os.Stderr)
	// We can't directly compare with os.Stdout/Stderr in tests,
	// but we can verify they're no longer the custom buffers
	if DataWriter == &customData {
		t.Error("DataWriter should be reset")
	}

	if LogWriter == &customLog {
		t.Error("LogWriter should be reset")
	}
}

func TestWriteDataWithFormat(t *testing.T) {
	var dataBuf, logBuf bytes.Buffer
	SetWriters(&dataBuf, &logBuf)
	defer ResetWriters()

	WriteData("value: %d, name: %s", 42, "test")

	expected := "value: 42, name: test"
	if dataBuf.String() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, dataBuf.String())
	}
}

func TestWriteLogWithFormat(t *testing.T) {
	var dataBuf, logBuf bytes.Buffer
	SetWriters(&dataBuf, &logBuf)
	defer ResetWriters()

	WriteLog("error: %s, code: %d", "not found", 404)

	expected := "error: not found, code: 404"
	if logBuf.String() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, logBuf.String())
	}
}

func TestWriteLogLineWithFormat(t *testing.T) {
	var dataBuf, logBuf bytes.Buffer
	SetWriters(&dataBuf, &logBuf)
	defer ResetWriters()

	WriteLogLine("Processing %s...", "file.txt")

	expected := "Processing file.txt...\n"
	if logBuf.String() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, logBuf.String())
	}
}

func TestMultipleWrites(t *testing.T) {
	var dataBuf, logBuf bytes.Buffer
	SetWriters(&dataBuf, &logBuf)
	defer ResetWriters()

	// Mix data and log writes
	WriteData("data1")
	WriteLog("log1")
	WriteData(" data2")
	WriteLogLine("log2")

	expectedData := "data1 data2"
	if dataBuf.String() != expectedData {
		t.Errorf("Expected data '%s', got '%s'", expectedData, dataBuf.String())
	}

	expectedLog := "log1log2\n"
	if logBuf.String() != expectedLog {
		t.Errorf("Expected log '%s', got '%s'", expectedLog, logBuf.String())
	}
}

func TestConcurrentWriteSafety(t *testing.T) {
	var dataBuf, logBuf bytes.Buffer
	SetWriters(&dataBuf, &logBuf)
	defer ResetWriters()

	// Note: These writes are sequential in tests, but the API should be safe for concurrent use
	// For true concurrent testing, we'd need goroutines and sync primitives
	done := make(chan bool)

	go func() {
		WriteData("concurrent data")
		done <- true
	}()

	go func() {
		WriteLog("concurrent log")
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done

	// Just verify both writes succeeded (order may vary)
	if dataBuf.Len() == 0 {
		t.Error("Data buffer should have content")
	}

	if logBuf.Len() == 0 {
		t.Error("Log buffer should have content")
	}
}
