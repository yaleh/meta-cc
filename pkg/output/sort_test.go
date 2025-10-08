package output

import (
	"testing"
	"time"

	"github.com/yaleh/meta-cc/internal/parser"
)

// Helper function to create test ToolCall with specific timestamp
func createToolCall(uuid string, timestamp time.Time, toolName string) parser.ToolCall {
	return parser.ToolCall{
		UUID:      uuid,
		Timestamp: timestamp.Format(time.RFC3339Nano),
		ToolName:  toolName,
		Status:    "success",
	}
}

// TestSortByTimestamp_BasicOrdering tests basic timestamp sorting
func TestSortByTimestamp_BasicOrdering(t *testing.T) {
	now := time.Now()
	tools := []parser.ToolCall{
		createToolCall("c", now.Add(2*time.Hour), "Edit"),
		createToolCall("a", now, "Bash"),
		createToolCall("b", now.Add(1*time.Hour), "Read"),
	}

	SortByTimestamp(tools)

	// Verify sorted order (a, b, c by time)
	if tools[0].UUID != "a" {
		t.Errorf("Expected first UUID 'a', got '%s'", tools[0].UUID)
	}
	if tools[1].UUID != "b" {
		t.Errorf("Expected second UUID 'b', got '%s'", tools[1].UUID)
	}
	if tools[2].UUID != "c" {
		t.Errorf("Expected third UUID 'c', got '%s'", tools[2].UUID)
	}
}

// TestSortByTimestamp_Idempotency tests that sorting twice produces same result
func TestSortByTimestamp_Idempotency(t *testing.T) {
	now := time.Now()
	tools := []parser.ToolCall{
		createToolCall("d", now.Add(3*time.Hour), "Write"),
		createToolCall("a", now, "Bash"),
		createToolCall("c", now.Add(2*time.Hour), "Edit"),
		createToolCall("b", now.Add(1*time.Hour), "Read"),
	}

	// Sort first time
	SortByTimestamp(tools)
	firstResult := make([]string, len(tools))
	for i, tc := range tools {
		firstResult[i] = tc.UUID
	}

	// Sort second time
	SortByTimestamp(tools)
	secondResult := make([]string, len(tools))
	for i, tc := range tools {
		secondResult[i] = tc.UUID
	}

	// Verify identical results
	for i := range firstResult {
		if firstResult[i] != secondResult[i] {
			t.Errorf("Idempotency check failed at index %d: first=%s, second=%s",
				i, firstResult[i], secondResult[i])
		}
	}
}

// TestSortByTimestamp_Stability tests stable sort behavior
func TestSortByTimestamp_Stability(t *testing.T) {
	now := time.Now()
	sameTime := now.Format(time.RFC3339Nano)

	// Three tools with same timestamp but different UUIDs
	tools := []parser.ToolCall{
		{UUID: "third", Timestamp: sameTime, ToolName: "Bash"},
		{UUID: "first", Timestamp: sameTime, ToolName: "Read"},
		{UUID: "second", Timestamp: sameTime, ToolName: "Edit"},
	}

	SortByTimestamp(tools)

	// Verify relative order is preserved (stable sort)
	// Original order: third, first, second
	if tools[0].UUID != "third" || tools[1].UUID != "first" || tools[2].UUID != "second" {
		t.Errorf("Stable sort failed: got order [%s, %s, %s], expected [third, first, second]",
			tools[0].UUID, tools[1].UUID, tools[2].UUID)
	}
}

// TestSortByTimestamp_EmptySlice tests handling of empty slice
func TestSortByTimestamp_EmptySlice(t *testing.T) {
	tools := []parser.ToolCall{}

	// Should not panic
	SortByTimestamp(tools)

	if len(tools) != 0 {
		t.Error("Empty slice should remain empty after sorting")
	}
}

// TestSortByTimestamp_SingleElement tests handling of single element
func TestSortByTimestamp_SingleElement(t *testing.T) {
	tools := []parser.ToolCall{
		createToolCall("only", time.Now(), "Bash"),
	}

	SortByTimestamp(tools)

	if len(tools) != 1 || tools[0].UUID != "only" {
		t.Error("Single element slice should remain unchanged")
	}
}

// TestSortByUUID_BasicOrdering tests basic UUID sorting
func TestSortByUUID_BasicOrdering(t *testing.T) {
	now := time.Now()
	tools := []parser.ToolCall{
		createToolCall("charlie", now, "Edit"),
		createToolCall("alice", now, "Bash"),
		createToolCall("bob", now, "Read"),
	}

	SortByUUID(tools)

	// Verify lexicographic order
	if tools[0].UUID != "alice" {
		t.Errorf("Expected first UUID 'alice', got '%s'", tools[0].UUID)
	}
	if tools[1].UUID != "bob" {
		t.Errorf("Expected second UUID 'bob', got '%s'", tools[1].UUID)
	}
	if tools[2].UUID != "charlie" {
		t.Errorf("Expected third UUID 'charlie', got '%s'", tools[2].UUID)
	}
}

// TestSortByUUID_Idempotency tests UUID sorting idempotency
func TestSortByUUID_Idempotency(t *testing.T) {
	now := time.Now()
	tools := []parser.ToolCall{
		createToolCall("zulu", now, "Write"),
		createToolCall("alpha", now, "Bash"),
		createToolCall("bravo", now, "Read"),
	}

	// Sort first time
	SortByUUID(tools)
	firstOrder := []string{tools[0].UUID, tools[1].UUID, tools[2].UUID}

	// Sort second time
	SortByUUID(tools)
	secondOrder := []string{tools[0].UUID, tools[1].UUID, tools[2].UUID}

	// Verify identical
	for i := range firstOrder {
		if firstOrder[i] != secondOrder[i] {
			t.Errorf("UUID sort idempotency failed at index %d", i)
		}
	}
}

// TestDefaultSort_UsesTimestamp tests that DefaultSort uses timestamp
func TestDefaultSort_UsesTimestamp(t *testing.T) {
	now := time.Now()
	tools := []parser.ToolCall{
		createToolCall("c", now.Add(2*time.Hour), "Edit"),
		createToolCall("a", now, "Bash"),
		createToolCall("b", now.Add(1*time.Hour), "Read"),
	}

	DefaultSort(tools)

	// Should be sorted by timestamp (a, b, c)
	if tools[0].UUID != "a" || tools[1].UUID != "b" || tools[2].UUID != "c" {
		t.Errorf("DefaultSort should sort by timestamp, got [%s, %s, %s]",
			tools[0].UUID, tools[1].UUID, tools[2].UUID)
	}
}

// TestSortByTimestamp_LargeDataset tests performance with larger dataset
func TestSortByTimestamp_LargeDataset(t *testing.T) {
	const size = 1000
	tools := make([]parser.ToolCall, size)

	baseTime := time.Now()
	// Create tools in reverse chronological order
	for i := 0; i < size; i++ {
		tools[i] = createToolCall(
			string(rune('a'+i%26))+string(rune('0'+i)),
			baseTime.Add(time.Duration(size-i)*time.Second),
			"Bash",
		)
	}

	SortByTimestamp(tools)

	// Verify sorted order
	for i := 0; i < size-1; i++ {
		if tools[i].Timestamp > tools[i+1].Timestamp {
			t.Errorf("Large dataset not properly sorted at index %d", i)
			break
		}
	}
}

// TestSortByTimestamp_Determinism tests same input produces same output
func TestSortByTimestamp_Determinism(t *testing.T) {
	now := time.Now()

	// Create two identical slices
	createSlice := func() []parser.ToolCall {
		return []parser.ToolCall{
			createToolCall("e", now.Add(4*time.Hour), "Write"),
			createToolCall("c", now.Add(2*time.Hour), "Edit"),
			createToolCall("a", now, "Bash"),
			createToolCall("d", now.Add(3*time.Hour), "Read"),
			createToolCall("b", now.Add(1*time.Hour), "Grep"),
		}
	}

	tools1 := createSlice()
	tools2 := createSlice()

	// Sort both
	SortByTimestamp(tools1)
	SortByTimestamp(tools2)

	// Verify identical output
	if len(tools1) != len(tools2) {
		t.Fatal("Slices have different lengths after sorting")
	}

	for i := range tools1 {
		if tools1[i].UUID != tools2[i].UUID {
			t.Errorf("Determinism check failed at index %d: tools1=%s, tools2=%s",
				i, tools1[i].UUID, tools2[i].UUID)
		}
		if tools1[i].Timestamp != tools2[i].Timestamp {
			t.Errorf("Timestamp mismatch at index %d", i)
		}
	}
}

// BenchmarkSortByTimestamp_100 benchmarks sorting 100 tools
func BenchmarkSortByTimestamp_100(b *testing.B) {
	benchmarkSortByTimestamp(b, 100)
}

// BenchmarkSortByTimestamp_1000 benchmarks sorting 1000 tools
func BenchmarkSortByTimestamp_1000(b *testing.B) {
	benchmarkSortByTimestamp(b, 1000)
}

// BenchmarkSortByTimestamp_10000 benchmarks sorting 10000 tools
func BenchmarkSortByTimestamp_10000(b *testing.B) {
	benchmarkSortByTimestamp(b, 10000)
}

func benchmarkSortByTimestamp(b *testing.B, size int) {
	baseTime := time.Now()
	tools := make([]parser.ToolCall, size)

	// Create tools in random order
	for i := 0; i < size; i++ {
		tools[i] = createToolCall(
			string(rune('a'+i%26))+string(rune('0'+i)),
			baseTime.Add(time.Duration(size-i)*time.Millisecond),
			"Bash",
		)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create a copy to sort (to avoid sorting already-sorted data)
		testTools := make([]parser.ToolCall, len(tools))
		copy(testTools, tools)
		SortByTimestamp(testTools)
	}
}

// TestSortByTimestamp_ErrorEntries tests sorting error entries
func TestSortByTimestamp_ErrorEntries(t *testing.T) {
	now := time.Now()
	errors := []ErrorEntry{
		{UUID: "c", Timestamp: now.Add(2 * time.Hour).Format(time.RFC3339Nano), ToolName: "Edit"},
		{UUID: "a", Timestamp: now.Format(time.RFC3339Nano), ToolName: "Bash"},
		{UUID: "b", Timestamp: now.Add(1 * time.Hour).Format(time.RFC3339Nano), ToolName: "Read"},
	}

	SortByTimestamp(errors)

	// Verify sorted order (a, b, c by time)
	if errors[0].UUID != "a" {
		t.Errorf("Expected first UUID 'a', got '%s'", errors[0].UUID)
	}
	if errors[1].UUID != "b" {
		t.Errorf("Expected second UUID 'b', got '%s'", errors[1].UUID)
	}
	if errors[2].UUID != "c" {
		t.Errorf("Expected third UUID 'c', got '%s'", errors[2].UUID)
	}
}

// TestSortByTimestamp_ErrorEntries_Stability tests stable sort for error entries
func TestSortByTimestamp_ErrorEntries_Stability(t *testing.T) {
	now := time.Now()
	sameTime := now.Format(time.RFC3339Nano)

	errors := []ErrorEntry{
		{UUID: "third", Timestamp: sameTime, ToolName: "Bash"},
		{UUID: "first", Timestamp: sameTime, ToolName: "Read"},
		{UUID: "second", Timestamp: sameTime, ToolName: "Edit"},
	}

	SortByTimestamp(errors)

	// Verify relative order is preserved (stable sort)
	if errors[0].UUID != "third" || errors[1].UUID != "first" || errors[2].UUID != "second" {
		t.Errorf("Stable sort failed for errors: got order [%s, %s, %s]",
			errors[0].UUID, errors[1].UUID, errors[2].UUID)
	}
}

// TestSortByTimestamp_ErrorEntries_Idempotency tests idempotency for error entries
func TestSortByTimestamp_ErrorEntries_Idempotency(t *testing.T) {
	now := time.Now()
	errors := []ErrorEntry{
		{UUID: "d", Timestamp: now.Add(3 * time.Hour).Format(time.RFC3339Nano)},
		{UUID: "a", Timestamp: now.Format(time.RFC3339Nano)},
		{UUID: "c", Timestamp: now.Add(2 * time.Hour).Format(time.RFC3339Nano)},
		{UUID: "b", Timestamp: now.Add(1 * time.Hour).Format(time.RFC3339Nano)},
	}

	// Sort first time
	SortByTimestamp(errors)
	firstResult := make([]string, len(errors))
	for i, e := range errors {
		firstResult[i] = e.UUID
	}

	// Sort second time
	SortByTimestamp(errors)
	secondResult := make([]string, len(errors))
	for i, e := range errors {
		secondResult[i] = e.UUID
	}

	// Verify identical results
	for i := range firstResult {
		if firstResult[i] != secondResult[i] {
			t.Errorf("Idempotency check failed for errors at index %d: first=%s, second=%s",
				i, firstResult[i], secondResult[i])
		}
	}
}
