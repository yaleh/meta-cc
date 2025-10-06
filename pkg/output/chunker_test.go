package output

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/yale/meta-cc/internal/parser"
)

// generateToolCalls generates N mock ToolCall records for testing
func generateToolCalls(count int) []parser.ToolCall {
	tools := make([]parser.ToolCall, count)

	toolNames := []string{"Bash", "Read", "Edit", "Write", "Grep", "Glob"}
	statuses := []string{"success", "success", "success", "error"} // 75% success, 25% error

	for i := 0; i < count; i++ {
		tools[i] = parser.ToolCall{
			UUID:     fmt.Sprintf("uuid-%04d", i),
			ToolName: toolNames[i%len(toolNames)],
			Status:   statuses[i%len(statuses)],
			Input:    map[string]interface{}{"param": fmt.Sprintf("value-%d", i)},
			Output:   fmt.Sprintf("output-%d", i),
			Error:    "",
		}

		// Add error message for error status
		if tools[i].Status == "error" {
			tools[i].Error = fmt.Sprintf("error message %d", i)
		}
	}

	return tools
}

func TestSplitIntoChunks(t *testing.T) {
	tools := generateToolCalls(250) // 250 records

	tests := []struct {
		name           string
		chunkSize      int
		expectedChunks int
		expectedSizes  []int
	}{
		{
			name:           "250 records with chunk size 100",
			chunkSize:      100,
			expectedChunks: 3,
			expectedSizes:  []int{100, 100, 50},
		},
		{
			name:           "250 records with chunk size 50",
			chunkSize:      50,
			expectedChunks: 5,
			expectedSizes:  []int{50, 50, 50, 50, 50},
		},
		{
			name:           "250 records with chunk size 300",
			chunkSize:      300,
			expectedChunks: 1,
			expectedSizes:  []int{250},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunks := SplitIntoChunks(tools, tt.chunkSize)

			if len(chunks) != tt.expectedChunks {
				t.Errorf("expected %d chunks, got %d", tt.expectedChunks, len(chunks))
			}

			for i, chunk := range chunks {
				if len(chunk) != tt.expectedSizes[i] {
					t.Errorf("chunk %d: expected %d records, got %d", i, tt.expectedSizes[i], len(chunk))
				}
			}
		})
	}
}

func TestWriteChunk(t *testing.T) {
	tools := generateToolCalls(10)
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "test-chunk.jsonl")

	err := WriteChunk(tools, "jsonl", outputFile)
	if err != nil {
		t.Fatalf("WriteChunk failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Errorf("output file %s does not exist", outputFile)
	}

	// Verify content is valid JSON
	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	var parsedTools []parser.ToolCall
	if err := json.Unmarshal(data, &parsedTools); err != nil {
		t.Errorf("output is not valid JSON: %v", err)
	}

	if len(parsedTools) != 10 {
		t.Errorf("expected 10 records in output, got %d", len(parsedTools))
	}
}

func TestGenerateManifest(t *testing.T) {
	tempDir := t.TempDir()
	manifestPath := filepath.Join(tempDir, "manifest.jsonl")

	metadata := []ChunkMetadata{
		{
			Index:       0,
			File:        "chunk_0001.jsonl",
			Records:     100,
			SizeBytes:   12345,
			TotalChunks: 3,
		},
		{
			Index:       1,
			File:        "chunk_0002.jsonl",
			Records:     100,
			SizeBytes:   12340,
			TotalChunks: 3,
		},
		{
			Index:       2,
			File:        "chunk_0003.jsonl",
			Records:     50,
			SizeBytes:   6170,
			TotalChunks: 3,
		},
	}

	err := GenerateManifest(metadata, manifestPath)
	if err != nil {
		t.Fatalf("GenerateManifest failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Errorf("manifest file %s does not exist", manifestPath)
	}

	// Verify content
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("failed to read manifest: %v", err)
	}

	var manifest ChunkManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Errorf("manifest is not valid JSON: %v", err)
	}

	// Verify manifest fields
	if manifest.TotalRecords != 250 {
		t.Errorf("expected total records 250, got %d", manifest.TotalRecords)
	}
	if manifest.ChunkSize != 100 {
		t.Errorf("expected chunk size 100, got %d", manifest.ChunkSize)
	}
	if manifest.NumChunks != 3 {
		t.Errorf("expected 3 chunks, got %d", manifest.NumChunks)
	}
	if len(manifest.Chunks) != 3 {
		t.Errorf("expected 3 chunk entries, got %d", len(manifest.Chunks))
	}
}

func TestChunkToolCallsIntegration(t *testing.T) {
	// Integration test: 2000 records split into 20 chunks of 100 each
	tools := generateToolCalls(2000)
	tempDir := t.TempDir()

	metadata, err := ChunkToolCalls(tools, 100, tempDir, "jsonl")
	if err != nil {
		t.Fatalf("ChunkToolCalls failed: %v", err)
	}

	// Verify 20 chunks created
	if len(metadata) != 20 {
		t.Errorf("expected 20 chunks, got %d", len(metadata))
	}

	// Verify each chunk
	totalRecords := 0
	for i, meta := range metadata {
		// Check chunk index
		if meta.Index != i {
			t.Errorf("chunk %d has wrong index: %d", i, meta.Index)
		}

		// Check file exists
		if _, err := os.Stat(meta.File); os.IsNotExist(err) {
			t.Errorf("chunk file %s does not exist", meta.File)
		}

		// Check record count
		if meta.Records != 100 {
			t.Errorf("chunk %d: expected 100 records, got %d", i, meta.Records)
		}

		totalRecords += meta.Records

		// Verify file content
		data, err := os.ReadFile(meta.File)
		if err != nil {
			t.Errorf("failed to read chunk %d: %v", i, err)
			continue
		}

		var chunkTools []parser.ToolCall
		if err := json.Unmarshal(data, &chunkTools); err != nil {
			t.Errorf("chunk %d: invalid JSON: %v", i, err)
		}

		if len(chunkTools) != meta.Records {
			t.Errorf("chunk %d: file contains %d records but metadata says %d", i, len(chunkTools), meta.Records)
		}
	}

	// Verify total records
	if totalRecords != 2000 {
		t.Errorf("expected total 2000 records, got %d", totalRecords)
	}

	// Verify manifest exists
	manifestPath := filepath.Join(tempDir, "manifest.jsonl")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Errorf("manifest file does not exist")
	}
}

func TestChunkFileNaming(t *testing.T) {
	tools := generateToolCalls(50)
	tempDir := t.TempDir()

	metadata, err := ChunkToolCalls(tools, 10, tempDir, "jsonl")
	if err != nil {
		t.Fatalf("ChunkToolCalls failed: %v", err)
	}

	expectedNames := []string{
		"chunk_0001.jsonl",
		"chunk_0002.jsonl",
		"chunk_0003.jsonl",
		"chunk_0004.jsonl",
		"chunk_0005.jsonl",
	}

	for i, meta := range metadata {
		expectedPath := filepath.Join(tempDir, expectedNames[i])
		if meta.File != expectedPath {
			t.Errorf("chunk %d: expected file %s, got %s", i, expectedPath, meta.File)
		}
	}
}

func TestChunkDifferentFormats(t *testing.T) {
	tools := generateToolCalls(30)
	tempDir := t.TempDir()

	formats := []string{"jsonl", "tsv", "tsv"}

	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			formatDir := filepath.Join(tempDir, format)
			if err := os.MkdirAll(formatDir, 0755); err != nil {
				t.Fatalf("failed to create format dir: %v", err)
			}

			metadata, err := ChunkToolCalls(tools, 10, formatDir, format)
			if err != nil {
				t.Fatalf("ChunkToolCalls failed for format %s: %v", format, err)
			}

			if len(metadata) != 3 {
				t.Errorf("expected 3 chunks for format %s, got %d", format, len(metadata))
			}

			// Verify all files exist
			for _, meta := range metadata {
				if _, err := os.Stat(meta.File); os.IsNotExist(err) {
					t.Errorf("chunk file %s does not exist", meta.File)
				}
			}
		})
	}
}
