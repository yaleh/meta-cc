package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// generateTestJSONL creates a temporary JSONL file with n entries
func generateTestJSONL(t *testing.T, n int) string {
	t.Helper()

	tmpDir := t.TempDir()
	jsonlFile := filepath.Join(tmpDir, "test-session.jsonl")

	f, err := os.Create(jsonlFile)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)

	for i := 0; i < n; i++ {
		sessionID := fmt.Sprintf("session-%d", i%10)
		gitBranch := "main"
		if i%5 == 0 {
			gitBranch = "feature/branch"
		}

		// User message
		entry := map[string]interface{}{
			"type":       "user",
			"uuid":       fmt.Sprintf("user-%d", i),
			"timestamp":  fmt.Sprintf("2025-10-23T%02d:%02d:00Z", i/60, i%60),
			"sessionId":  sessionID,
			"parentUuid": fmt.Sprintf("parent-%d", i-1),
			"gitBranch":  gitBranch,
			"message": map[string]interface{}{
				"role": "user",
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": fmt.Sprintf("User message %d", i),
					},
				},
			},
		}
		if err := enc.Encode(entry); err != nil {
			t.Fatal(err)
		}

		// Assistant message with tool use
		toolName := "Read"
		if i%3 == 0 {
			toolName = "Edit"
		} else if i%5 == 0 {
			toolName = "Write"
		}

		entry = map[string]interface{}{
			"type":       "assistant",
			"uuid":       fmt.Sprintf("assistant-%d", i),
			"timestamp":  fmt.Sprintf("2025-10-23T%02d:%02d:05Z", i/60, i%60),
			"sessionId":  sessionID,
			"parentUuid": fmt.Sprintf("user-%d", i),
			"gitBranch":  gitBranch,
			"message": map[string]interface{}{
				"role": "assistant",
				"content": []map[string]interface{}{
					{
						"type": "tool_use",
						"id":   fmt.Sprintf("tool-%d", i),
						"name": toolName,
						"input": map[string]interface{}{
							"file_path": fmt.Sprintf("/test/file%d.txt", i),
						},
					},
				},
			},
		}
		if err := enc.Encode(entry); err != nil {
			t.Fatal(err)
		}
	}

	return tmpDir
}

// BenchmarkQueryExecution benchmarks full query execution
func BenchmarkQueryExecution(b *testing.B) {
	sizes := []int{100, 1000, 10000}

	for _, size := range sizes {
		tmpDir := generateTestJSONL(&testing.T{}, size)
		defer os.RemoveAll(tmpDir)

		executor := NewQueryExecutor(tmpDir)

		b.Run(fmt.Sprintf("size_%d_select_all", size), func(b *testing.B) {
			code, err := executor.compileExpression(".[]")
			if err != nil {
				b.Fatal(err)
			}

			files, _ := getJSONLFiles(tmpDir)
			ctx := context.Background()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				results := executor.streamFiles(ctx, files, code, 0)
				count := 0
				for range results {
					count++
				}
			}
		})

		b.Run(fmt.Sprintf("size_%d_filter_type", size), func(b *testing.B) {
			code, err := executor.compileExpression(`.[] | select(.type == "user")`)
			if err != nil {
				b.Fatal(err)
			}

			files, _ := getJSONLFiles(tmpDir)
			ctx := context.Background()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				results := executor.streamFiles(ctx, files, code, 0)
				count := 0
				for range results {
					count++
				}
			}
		})

		b.Run(fmt.Sprintf("size_%d_filter_and_project", size), func(b *testing.B) {
			code, err := executor.compileExpression(`.[] | select(.type == "user") | {timestamp, uuid}`)
			if err != nil {
				b.Fatal(err)
			}

			files, _ := getJSONLFiles(tmpDir)
			ctx := context.Background()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				results := executor.streamFiles(ctx, files, code, 0)
				count := 0
				for range results {
					count++
				}
			}
		})
	}
}

// BenchmarkExpressionCompilation benchmarks jq expression compilation
func BenchmarkExpressionCompilation(b *testing.B) {
	executor := NewQueryExecutor("")

	expressions := []string{
		".[]",
		`.[] | select(.type == "user")`,
		`.[] | select(.type == "user") | {timestamp, uuid}`,
		`.[] | group_by(.type) | map({type: .[0].type, count: length})`,
		`.[] | select(.timestamp | fromdateiso8601 > (now - 3600))`,
	}

	for _, expr := range expressions {
		b.Run(fmt.Sprintf("expr_%s", expr[:20]), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := executor.compileExpression(expr)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkStreamProcessing benchmarks streaming vs batch processing
func BenchmarkStreamProcessing(b *testing.B) {
	size := 1000
	tmpDir := generateTestJSONL(&testing.T{}, size)
	defer os.RemoveAll(tmpDir)

	executor := NewQueryExecutor(tmpDir)
	code, _ := executor.compileExpression(".[]")
	files, _ := getJSONLFiles(tmpDir)
	ctx := context.Background()

	b.Run("streaming", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			results := executor.streamFiles(ctx, files, code, 0)
			count := 0
			for range results {
				count++
			}
		}
	})

	b.Run("streaming_with_limit_10", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			results := executor.streamFiles(ctx, files, code, 10)
			count := 0
			for range results {
				count++
			}
		}
	})

	b.Run("streaming_with_limit_100", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			results := executor.streamFiles(ctx, files, code, 100)
			count := 0
			for range results {
				count++
			}
		}
	})
}

// BenchmarkHybridOutput benchmarks hybrid output mode decision
func BenchmarkHybridOutput(b *testing.B) {
	// Small result (< 8KB)
	smallResult := make([]map[string]interface{}, 10)
	for i := range smallResult {
		smallResult[i] = map[string]interface{}{
			"timestamp": "2025-10-23T10:00:00Z",
			"uuid":      fmt.Sprintf("uuid-%d", i),
			"type":      "user",
		}
	}

	// Large result (> 8KB)
	largeResult := make([]map[string]interface{}, 1000)
	for i := range largeResult {
		largeResult[i] = map[string]interface{}{
			"timestamp": "2025-10-23T10:00:00Z",
			"uuid":      fmt.Sprintf("uuid-%d", i),
			"type":      "user",
			"message": map[string]interface{}{
				"role": "user",
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": "This is a longer user message that will contribute to the overall size of the result set. We need to ensure that the result exceeds the inline threshold to test the hybrid output mode properly.",
					},
				},
			},
		}
	}

	b.Run("small_result_inline", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			data, _ := json.Marshal(smallResult)
			_ = len(data) < 8192 // Inline threshold check
		}
	})

	b.Run("large_result_file_ref", func(b *testing.B) {
		tmpDir := b.TempDir()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			data, _ := json.Marshal(largeResult)
			if len(data) >= 8192 {
				// Simulate file write
				tmpFile := filepath.Join(tmpDir, fmt.Sprintf("output-%d.jsonl", i))
				_ = os.WriteFile(tmpFile, data, 0644)
			}
		}
	})
}

// BenchmarkConcurrentQueries benchmarks concurrent query execution
func BenchmarkConcurrentQueries(b *testing.B) {
	size := 1000
	tmpDir := generateTestJSONL(&testing.T{}, size)
	defer os.RemoveAll(tmpDir)

	executor := NewQueryExecutor(tmpDir)
	code, _ := executor.compileExpression(".[]")
	files, _ := getJSONLFiles(tmpDir)

	b.Run("sequential", func(b *testing.B) {
		ctx := context.Background()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			results := executor.streamFiles(ctx, files, code, 0)
			count := 0
			for range results {
				count++
			}
		}
	})

	b.Run("concurrent_4", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			ctx := context.Background()
			for pb.Next() {
				results := executor.streamFiles(ctx, files, code, 0)
				count := 0
				for range results {
					count++
				}
			}
		})
	})
}

// BenchmarkMemoryUsage benchmarks memory usage for different result sizes
func BenchmarkMemoryUsage(b *testing.B) {
	sizes := []int{100, 1000, 10000}

	for _, size := range sizes {
		tmpDir := generateTestJSONL(&testing.T{}, size)
		defer os.RemoveAll(tmpDir)

		executor := NewQueryExecutor(tmpDir)
		code, _ := executor.compileExpression(".[]")
		files, _ := getJSONLFiles(tmpDir)

		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			ctx := context.Background()
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				results := executor.streamFiles(ctx, files, code, 0)
				// Collect all results to measure memory
				count := 0
				for range results {
					count++
				}
			}
		})
	}
}
