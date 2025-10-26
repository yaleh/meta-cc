package query

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// RecordSample represents a sample record from a session file
type RecordSample struct {
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
	Preview   string `json:"preview"`
}

// FileMetadata contains metadata about a session file
type FileMetadata struct {
	Path        string         `json:"path"`
	SizeBytes   int64          `json:"size_bytes"`
	LineCount   int            `json:"line_count"`
	RecordTypes map[string]int `json:"record_types"`
	TimeRange   TimeRange      `json:"time_range"`
	Samples     []RecordSample `json:"samples,omitempty"`
}

// InspectionSummary provides aggregate information about inspected files
type InspectionSummary struct {
	TotalFiles     int   `json:"total_files"`
	TotalSizeBytes int64 `json:"total_size_bytes"`
	TotalRecords   int   `json:"total_records"`
}

// InspectionResult is the result of inspecting session files
type InspectionResult struct {
	Files   []FileMetadata    `json:"files"`
	Summary InspectionSummary `json:"summary"`
}

// InspectFiles inspects one or more session files and returns metadata
func InspectFiles(files []string, includeSamples bool) (*InspectionResult, error) {
	result := &InspectionResult{
		Files: make([]FileMetadata, 0, len(files)),
		Summary: InspectionSummary{
			TotalFiles: len(files),
		},
	}

	for _, filePath := range files {
		metadata, err := inspectFile(filePath, includeSamples)
		if err != nil {
			return nil, fmt.Errorf("failed to inspect file %s: %w", filePath, err)
		}

		result.Files = append(result.Files, *metadata)
		result.Summary.TotalSizeBytes += metadata.SizeBytes
		result.Summary.TotalRecords += metadata.LineCount
	}

	return result, nil
}

// inspectFile inspects a single session file and returns its metadata
func inspectFile(path string, includeSamples bool) (*FileMetadata, error) {
	// Get file info
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	metadata := &FileMetadata{
		Path:        path,
		SizeBytes:   fileInfo.Size(),
		RecordTypes: make(map[string]int),
	}

	// Open and read file
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// Increase buffer size to handle long lines (session files can have very long tool_use blocks)
	maxCapacity := 10 * 1024 * 1024 // 10MB max line length
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	lines := make([]string, 0)
	var minTime, maxTime time.Time

	// Process each line
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

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

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
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

	return metadata, nil
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
