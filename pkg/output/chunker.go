package output

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/yale/meta-cc/internal/parser"
)

// ChunkMetadata contains metadata for a single chunk
type ChunkMetadata struct {
	Index       int    `json:"index"`
	File        string `json:"file"`
	Records     int    `json:"records"`
	SizeBytes   int64  `json:"size_bytes"`
	TotalChunks int    `json:"total_chunks"`
}

// ChunkManifest contains metadata for all chunks
type ChunkManifest struct {
	TotalRecords int             `json:"total_records"`
	ChunkSize    int             `json:"chunk_size"`
	NumChunks    int             `json:"num_chunks"`
	Chunks       []ChunkMetadata `json:"chunks"`
}

// SplitIntoChunks splits a slice of ToolCalls into chunks of specified size
func SplitIntoChunks(tools []parser.ToolCall, chunkSize int) [][]parser.ToolCall {
	if chunkSize <= 0 {
		return [][]parser.ToolCall{tools}
	}

	var chunks [][]parser.ToolCall
	for i := 0; i < len(tools); i += chunkSize {
		end := i + chunkSize
		if end > len(tools) {
			end = len(tools)
		}
		chunks = append(chunks, tools[i:end])
	}

	return chunks
}

// WriteChunk writes a chunk to a file in the specified format
func WriteChunk(chunk []parser.ToolCall, format, outputPath string) error {
	var data []byte
	var err error

	switch format {
	case "json":
		data, err = json.MarshalIndent(chunk, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}

	case "md", "markdown":
		content, err := FormatMarkdown(chunk)
		if err != nil {
			return fmt.Errorf("failed to format Markdown: %w", err)
		}
		data = []byte(content)

	case "csv":
		content, err := FormatCSV(chunk)
		if err != nil {
			return fmt.Errorf("failed to format CSV: %w", err)
		}
		data = []byte(content)

	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write chunk file: %w", err)
	}

	return nil
}

// GenerateManifest generates a manifest file with chunk metadata
func GenerateManifest(metadata []ChunkMetadata, manifestPath string) error {
	if len(metadata) == 0 {
		return fmt.Errorf("no chunk metadata provided")
	}

	// Calculate totals
	totalRecords := 0
	for _, meta := range metadata {
		totalRecords += meta.Records
	}

	// Infer chunk size from first chunk (or use Records if only one chunk)
	chunkSize := metadata[0].Records
	if len(metadata) > 1 && metadata[1].Records > chunkSize {
		chunkSize = metadata[1].Records
	}

	manifest := ChunkManifest{
		TotalRecords: totalRecords,
		ChunkSize:    chunkSize,
		NumChunks:    len(metadata),
		Chunks:       metadata,
	}

	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(manifestPath), 0755); err != nil {
		return fmt.Errorf("failed to create manifest directory: %w", err)
	}

	if err := os.WriteFile(manifestPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write manifest file: %w", err)
	}

	return nil
}

// ChunkToolCalls splits ToolCalls into chunks and writes them to files
// Returns metadata for all chunks created
func ChunkToolCalls(tools []parser.ToolCall, chunkSize int, outputDir, format string) ([]ChunkMetadata, error) {
	if chunkSize <= 0 {
		return nil, fmt.Errorf("chunk size must be > 0")
	}

	chunks := SplitIntoChunks(tools, chunkSize)
	totalChunks := len(chunks)
	metadata := make([]ChunkMetadata, 0, totalChunks)

	// Get file extension based on format
	ext := getExtension(format)

	for i, chunk := range chunks {
		// Generate filename with 4-digit padding: chunk_0001.json
		filename := fmt.Sprintf("chunk_%04d.%s", i+1, ext)
		outputPath := filepath.Join(outputDir, filename)

		// Write chunk file
		if err := WriteChunk(chunk, format, outputPath); err != nil {
			return nil, fmt.Errorf("failed to write chunk %d: %w", i, err)
		}

		// Get file size
		fileInfo, err := os.Stat(outputPath)
		if err != nil {
			return nil, fmt.Errorf("failed to stat chunk file %d: %w", i, err)
		}

		metadata = append(metadata, ChunkMetadata{
			Index:       i,
			File:        outputPath,
			Records:     len(chunk),
			SizeBytes:   fileInfo.Size(),
			TotalChunks: totalChunks,
		})
	}

	// Generate manifest file
	manifestPath := filepath.Join(outputDir, "manifest.json")
	if err := GenerateManifest(metadata, manifestPath); err != nil {
		return nil, fmt.Errorf("failed to generate manifest: %w", err)
	}

	return metadata, nil
}

// getExtension returns the file extension for a given format
func getExtension(format string) string {
	switch format {
	case "json":
		return "json"
	case "md", "markdown":
		return "md"
	case "csv":
		return "csv"
	default:
		return "txt"
	}
}
