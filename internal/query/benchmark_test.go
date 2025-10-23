package query

import (
	"fmt"
	"testing"

	"github.com/yaleh/meta-cc/internal/parser"
)

// generateTestEntries creates n test entries for benchmarking
func generateTestEntries(n int) []parser.SessionEntry {
	entries := make([]parser.SessionEntry, 0, n)

	for i := 0; i < n; i++ {
		sessionID := fmt.Sprintf("session-%d", i%10) // 10 different sessions
		gitBranch := "main"
		if i%5 == 0 {
			gitBranch = "feature/branch"
		}

		// User message
		entries = append(entries, parser.SessionEntry{
			Type:       "user",
			UUID:       fmt.Sprintf("user-%d", i),
			Timestamp:  fmt.Sprintf("2025-10-23T%02d:%02d:00Z", i/60, i%60),
			SessionID:  sessionID,
			ParentUUID: fmt.Sprintf("parent-%d", i-1),
			GitBranch:  gitBranch,
			Message: &parser.Message{
				Role: "user",
				Content: []parser.ContentBlock{
					{
						Type: "text",
						Text: fmt.Sprintf("User message %d", i),
					},
				},
			},
		})

		// Assistant message with tool use
		toolName := "Read"
		if i%3 == 0 {
			toolName = "Edit"
		} else if i%5 == 0 {
			toolName = "Write"
		}

		entries = append(entries, parser.SessionEntry{
			Type:       "assistant",
			UUID:       fmt.Sprintf("assistant-%d", i),
			Timestamp:  fmt.Sprintf("2025-10-23T%02d:%02d:05Z", i/60, i%60),
			SessionID:  sessionID,
			ParentUUID: fmt.Sprintf("user-%d", i),
			GitBranch:  gitBranch,
			Message: &parser.Message{
				Role: "assistant",
				Content: []parser.ContentBlock{
					{
						Type: "tool_use",
						ToolUse: &parser.ToolUse{
							ID:   fmt.Sprintf("tool-%d", i),
							Name: toolName,
							Input: map[string]interface{}{
								"file_path": fmt.Sprintf("/test/file%d.txt", i),
							},
						},
					},
				},
			},
		})

		// Tool result
		status := "success"
		if i%10 == 0 {
			status = "error"
		}

		entries = append(entries, parser.SessionEntry{
			Type:       "user",
			UUID:       fmt.Sprintf("tool-result-%d", i),
			Timestamp:  fmt.Sprintf("2025-10-23T%02d:%02d:10Z", i/60, i%60),
			SessionID:  sessionID,
			ParentUUID: fmt.Sprintf("assistant-%d", i),
			GitBranch:  gitBranch,
			Message: &parser.Message{
				Role: "user",
				Content: []parser.ContentBlock{
					{
						Type: "tool_result",
						ToolResult: &parser.ToolResult{
							ToolUseID: fmt.Sprintf("tool-%d", i),
							Content:   fmt.Sprintf("Result for tool %d", i),
							IsError:   status == "error",
							Status:    status,
						},
					},
				},
			},
		})
	}

	return entries
}

// BenchmarkQueryEntries benchmarks querying all entries (baseline)
func BenchmarkQueryEntries(b *testing.B) {
	sizes := []int{100, 1000, 10000}

	for _, size := range sizes {
		entries := generateTestEntries(size)

		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			params := QueryParams{
				Resource: "entries",
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := Query(entries, params)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkQueryMessages benchmarks extracting and querying messages
func BenchmarkQueryMessages(b *testing.B) {
	sizes := []int{100, 1000, 10000}

	for _, size := range sizes {
		entries := generateTestEntries(size)

		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			params := QueryParams{
				Resource: "messages",
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := Query(entries, params)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkQueryTools benchmarks extracting and querying tools
func BenchmarkQueryTools(b *testing.B) {
	sizes := []int{100, 1000, 10000}

	for _, size := range sizes {
		entries := generateTestEntries(size)

		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			params := QueryParams{
				Resource: "tools",
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := Query(entries, params)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkQueryToolsWithFilter benchmarks filtering tool results
func BenchmarkQueryToolsWithFilter(b *testing.B) {
	sizes := []int{100, 1000, 10000}

	for _, size := range sizes {
		entries := generateTestEntries(size)

		b.Run(fmt.Sprintf("size_%d_filter_name", size), func(b *testing.B) {
			params := QueryParams{
				Resource: "tools",
				Filter: FilterSpec{
					ToolName: "Read",
				},
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := Query(entries, params)
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run(fmt.Sprintf("size_%d_filter_status", size), func(b *testing.B) {
			params := QueryParams{
				Resource: "tools",
				Filter: FilterSpec{
					ToolStatus: "error",
				},
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := Query(entries, params)
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run(fmt.Sprintf("size_%d_filter_both", size), func(b *testing.B) {
			params := QueryParams{
				Resource: "tools",
				Filter: FilterSpec{
					ToolName:   "Read",
					ToolStatus: "error",
				},
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := Query(entries, params)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkQueryWithAggregate benchmarks aggregation operations
func BenchmarkQueryWithAggregate(b *testing.B) {
	sizes := []int{100, 1000, 10000}

	for _, size := range sizes {
		entries := generateTestEntries(size)

		b.Run(fmt.Sprintf("size_%d_count", size), func(b *testing.B) {
			params := QueryParams{
				Resource: "tools",
				Aggregate: AggregateSpec{
					Function: "count",
				},
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := Query(entries, params)
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run(fmt.Sprintf("size_%d_count_by_field", size), func(b *testing.B) {
			params := QueryParams{
				Resource: "tools",
				Aggregate: AggregateSpec{
					Function: "count",
					Field:    "tool_name",
				},
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := Query(entries, params)
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run(fmt.Sprintf("size_%d_group_by_status", size), func(b *testing.B) {
			params := QueryParams{
				Resource: "tools",
				Aggregate: AggregateSpec{
					Function: "count",
					Field:    "status",
				},
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := Query(entries, params)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkQueryFilterAndAggregate benchmarks combined filter + aggregate
func BenchmarkQueryFilterAndAggregate(b *testing.B) {
	sizes := []int{100, 1000, 10000}

	for _, size := range sizes {
		entries := generateTestEntries(size)

		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			params := QueryParams{
				Resource: "tools",
				Filter: FilterSpec{
					ToolStatus: "error",
				},
				Aggregate: AggregateSpec{
					Function: "count",
					Field:    "tool_name",
				},
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := Query(entries, params)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkQueryUserMessages benchmarks user message filtering
func BenchmarkQueryUserMessages(b *testing.B) {
	sizes := []int{100, 1000, 10000}

	for _, size := range sizes {
		entries := generateTestEntries(size)

		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			params := QueryParams{
				Resource: "messages",
				Filter: FilterSpec{
					Role: "user",
				},
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := Query(entries, params)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkQueryAssistantMessages benchmarks assistant message filtering
func BenchmarkQueryAssistantMessages(b *testing.B) {
	sizes := []int{100, 1000, 10000}

	for _, size := range sizes {
		entries := generateTestEntries(size)

		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			params := QueryParams{
				Resource: "messages",
				Filter: FilterSpec{
					Role: "assistant",
				},
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := Query(entries, params)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkQueryBySession benchmarks session filtering
func BenchmarkQueryBySession(b *testing.B) {
	sizes := []int{100, 1000, 10000}

	for _, size := range sizes {
		entries := generateTestEntries(size)

		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			params := QueryParams{
				Resource: "entries",
				Filter: FilterSpec{
					SessionID: "session-5",
				},
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := Query(entries, params)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkQueryByGitBranch benchmarks git branch filtering
func BenchmarkQueryByGitBranch(b *testing.B) {
	sizes := []int{100, 1000, 10000}

	for _, size := range sizes {
		entries := generateTestEntries(size)

		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			params := QueryParams{
				Resource: "entries",
				Filter: FilterSpec{
					GitBranch: "feature/branch",
				},
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := Query(entries, params)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkResourceSelection benchmarks the resource selection step
func BenchmarkResourceSelection(b *testing.B) {
	sizes := []int{100, 1000, 10000}

	for _, size := range sizes {
		entries := generateTestEntries(size)

		b.Run(fmt.Sprintf("size_%d_entries", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := SelectResource(entries, "entries")
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run(fmt.Sprintf("size_%d_messages", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := SelectResource(entries, "messages")
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run(fmt.Sprintf("size_%d_tools", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := SelectResource(entries, "tools")
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkFilterApplication benchmarks the filter application step
func BenchmarkFilterApplication(b *testing.B) {
	entries := generateTestEntries(1000)
	tools, _ := SelectResource(entries, "tools")

	b.Run("no_filter", func(b *testing.B) {
		filter := FilterSpec{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = ApplyFilter(tools, filter)
		}
	})

	b.Run("single_field_filter", func(b *testing.B) {
		filter := FilterSpec{
			ToolName: "Read",
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = ApplyFilter(tools, filter)
		}
	})

	b.Run("multi_field_filter", func(b *testing.B) {
		filter := FilterSpec{
			ToolName:   "Read",
			ToolStatus: "error",
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = ApplyFilter(tools, filter)
		}
	})
}

// BenchmarkAggregation benchmarks the aggregation step
func BenchmarkAggregation(b *testing.B) {
	entries := generateTestEntries(1000)
	tools, _ := SelectResource(entries, "tools")

	b.Run("count_all", func(b *testing.B) {
		agg := AggregateSpec{
			Function: "count",
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = ApplyAggregate(tools, agg)
		}
	})

	b.Run("count_by_field", func(b *testing.B) {
		agg := AggregateSpec{
			Function: "count",
			Field:    "tool_name",
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = ApplyAggregate(tools, agg)
		}
	})
}
