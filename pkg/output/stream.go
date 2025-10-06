package output

import (
	"encoding/json"
	"io"
)

// StreamWriter writes data as JSON Lines (JSONL) format.
// Each record is written as a single line of compact JSON.
type StreamWriter struct {
	writer io.Writer
}

// NewStreamWriter creates a new JSONL stream writer.
func NewStreamWriter(w io.Writer) *StreamWriter {
	return &StreamWriter{writer: w}
}

// WriteRecord writes a single record as a JSON line.
// The record is marshaled to compact JSON and written with a trailing newline.
func (sw *StreamWriter) WriteRecord(record interface{}) error {
	// Marshal to compact JSON
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}

	// Write JSON line
	_, err = sw.writer.Write(data)
	if err != nil {
		return err
	}

	// Write newline
	_, err = sw.writer.Write([]byte("\n"))
	return err
}
