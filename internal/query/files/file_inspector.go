package files

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/yaleh/meta-cc/internal/parser"
	providerpkg "github.com/yaleh/meta-cc/internal/provider"
	"github.com/yaleh/meta-cc/internal/types"
)

// RecordSample represents a sample record from a session file
type RecordSample struct {
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
	Preview   string `json:"preview"`
}

// FileMetadata contains metadata about a session file, including its health.
//
// The health fields (Entries, Parseable, Empty, Error) are what make a corrupt
// file discoverable at inspection time rather than at crash time: before them,
// a truncated file reported {"line_count": 1, "record_types": {}} — a shape
// indistinguishable from a legitimately thin session, with nothing saying the
// file was broken.
type FileMetadata struct {
	Path        string          `json:"path"`
	SizeBytes   int64           `json:"size_bytes"`
	LineCount   int             `json:"line_count"`
	Entries     int             `json:"entries"`
	Parseable   bool            `json:"parseable"`
	Empty       bool            `json:"empty"`
	RecordTypes map[string]int  `json:"record_types"`
	TimeRange   types.TimeRange `json:"time_range"`
	Samples     []RecordSample  `json:"samples,omitempty"`
	Error       string          `json:"error,omitempty"`
}

// InspectionSummary provides aggregate information about inspected files
type InspectionSummary struct {
	TotalFiles     int   `json:"total_files"`
	TotalSizeBytes int64 `json:"total_size_bytes"`
	TotalRecords   int   `json:"total_records"`
}

// InspectionResult is the result of inspecting session files. MalformedFiles is
// empty for a healthy corpus and never nil.
type InspectionResult struct {
	Files          []FileMetadata              `json:"files"`
	Summary        InspectionSummary           `json:"summary"`
	MalformedFiles []providerpkg.MalformedFile `json:"malformed_files"`
}

// InspectFiles inspects one or more session files and returns metadata.
//
// It does not fail on a bad file: an unreadable or empty or truncated entry is
// reported as structured data in that file's FileMetadata (and named in
// MalformedFiles), while the other files are inspected normally. Returning an
// error for one file would erase every other file in the batch, which is the
// whole-batch-abort failure this replaces.
func InspectFiles(files []string, includeSamples bool) *InspectionResult {
	result := &InspectionResult{
		Files:          make([]FileMetadata, 0, len(files)),
		Summary:        InspectionSummary{TotalFiles: len(files)},
		MalformedFiles: make([]providerpkg.MalformedFile, 0),
	}

	health := make([]providerpkg.FileHealth, 0, len(files))
	for _, filePath := range files {
		// One probe decides health for every discovery surface; the line scan
		// below only adds the descriptive detail (record types, time range,
		// samples) that inspection exists to provide.
		h := providerpkg.ProbeFileHealth(filePath, providerpkg.KindClaudeSession)
		health = append(health, h)

		metadata := FileMetadata{
			Path:        filePath,
			SizeBytes:   h.Bytes,
			Entries:     h.Entries,
			Parseable:   h.Parseable,
			Empty:       h.Empty,
			Error:       h.Error,
			RecordTypes: make(map[string]int),
		}
		if h.Readable() {
			if err := scanSessionFile(&metadata, includeSamples); err != nil {
				// Raced away between probe and scan (or an I/O fault): record
				// it on this file and keep the rest of the batch.
				metadata.Parseable = false
				metadata.Error = err.Error()
			}
		}

		result.Files = append(result.Files, metadata)
		result.Summary.TotalSizeBytes += metadata.SizeBytes
		result.Summary.TotalRecords += metadata.LineCount
	}

	result.MalformedFiles = providerpkg.MalformedFilesFromHealth(health)
	return result
}

// scanSessionFile fills in the descriptive fields of an already-probed
// FileMetadata. It is separate from the health probe on purpose: health has one
// rule shared by every discovery tool, while this is inspection detail.
func scanSessionFile(metadata *FileMetadata, includeSamples bool) error {
	path := metadata.Path

	// Open and read file
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	r := bufio.NewReader(file)
	lines := make([]string, 0)
	var minTime, maxTime time.Time

	// Process each line using streaming reader (handles lines of any size)
	for {
		lineBytes, skipped, readErr := parser.ReadLineFiltered(r, parser.StrategyDefault)
		if skipped {
			if readErr == io.EOF {
				break
			}
			continue
		}
		if len(lineBytes) > 0 {
			// Trim trailing newline
			line := string(lineBytes)
			if len(line) > 0 && line[len(line)-1] == '\n' {
				line = line[:len(line)-1]
			}
			if line != "" {
				metadata.LineCount++
				lines = append(lines, line)

				// Parse record type (only count valid JSON with type field)
				recordType := parseRecordType(line)
				if recordType != "unknown" {
					metadata.RecordTypes[recordType]++
				}

				// Extract timestamp for time range
				var record map[string]interface{}
				if err := json.Unmarshal([]byte(line), &record); err == nil {
					if ts, ok := record["timestamp"].(string); ok {
						if t, err := time.Parse(time.RFC3339, ts); err == nil {
							if minTime.IsZero() || t.Before(minTime) {
								minTime = t
							}
							if maxTime.IsZero() || t.After(maxTime) {
								maxTime = t
							}
						}
					}
				}
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return fmt.Errorf("error reading file: %w", readErr)
		}
	}

	// Set time range
	if !minTime.IsZero() {
		metadata.TimeRange.Start = minTime.Format(time.RFC3339)
		metadata.TimeRange.End = maxTime.Format(time.RFC3339)
	}

	// Collect samples if requested
	if includeSamples && len(lines) > 0 {
		metadata.Samples = collectSamples(lines, metadata.RecordTypes)
	}

	return nil
}

// parseRecordType extracts the record type from a JSONL line
func parseRecordType(line string) string {
	var record map[string]interface{}
	if err := json.Unmarshal([]byte(line), &record); err != nil {
		return "unknown"
	}

	if recordType, ok := record["type"].(string); ok {
		return recordType
	}

	return "unknown"
}

// collectSamples collects 1-2 sample records per type
func collectSamples(lines []string, recordTypes map[string]int) []RecordSample {
	samples := make([]RecordSample, 0)
	samplesPerType := make(map[string]int)
	maxSamplesPerType := 2

	for _, line := range lines {
		var record map[string]interface{}
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			continue
		}

		recordType, ok := record["type"].(string)
		if !ok {
			recordType = "unknown"
		}

		// Skip if we already have enough samples for this type
		if samplesPerType[recordType] >= maxSamplesPerType {
			continue
		}

		timestamp := ""
		if ts, ok := record["timestamp"].(string); ok {
			timestamp = ts
		}

		// Create preview (first 100 chars of JSON including ellipsis)
		preview := line
		if len(preview) > 100 {
			preview = preview[:97] + "..."
		}

		samples = append(samples, RecordSample{
			Type:      recordType,
			Timestamp: timestamp,
			Preview:   preview,
		})

		samplesPerType[recordType]++
	}

	return samples
}
