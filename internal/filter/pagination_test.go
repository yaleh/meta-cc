package filter

import (
	"math"
	"testing"

	"github.com/yaleh/meta-cc/internal/types"
)

// generateTestToolCalls creates test ToolCall data
func generateTestToolCalls(count int) []types.ToolCall {
	calls := make([]types.ToolCall, count)

	for i := 0; i < count; i++ {
		calls[i] = types.ToolCall{
			UUID:     string(rune('A' + (i % 26))),
			ToolName: "TestTool",
			Status:   "success",
		}
	}

	return calls
}

func TestApplyPagination(t *testing.T) {
	tools := generateTestToolCalls(100)

	tests := []struct {
		name     string
		config   PaginationConfig
		expected int
	}{
		{
			name:     "no pagination",
			config:   PaginationConfig{Limit: 0, Offset: 0},
			expected: 100,
		},
		{
			name:     "limit 50",
			config:   PaginationConfig{Limit: 50, Offset: 0},
			expected: 50,
		},
		{
			name:     "offset 50, limit 30",
			config:   PaginationConfig{Limit: 30, Offset: 50},
			expected: 30,
		},
		{
			name:     "offset beyond end",
			config:   PaginationConfig{Limit: 10, Offset: 120},
			expected: 0,
		},
		{
			name:     "limit exceeds remaining",
			config:   PaginationConfig{Limit: 100, Offset: 90},
			expected: 10,
		},
		{
			name:     "negative limit (treat as no limit)",
			config:   PaginationConfig{Limit: -1, Offset: 0},
			expected: 100,
		},
		{
			name:     "negative offset (treat as 0)",
			config:   PaginationConfig{Limit: 10, Offset: -5},
			expected: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ApplyPagination(tools, tt.config)
			if len(result) != tt.expected {
				t.Errorf("expected %d records, got %d", tt.expected, len(result))
			}
		})
	}
}

func TestCalculateMetadata(t *testing.T) {
	tests := []struct {
		name         string
		totalRecords int
		config       PaginationConfig
		expected     PaginationMetadata
	}{
		{
			name:         "first page",
			totalRecords: 100,
			config:       PaginationConfig{Limit: 50, Offset: 0},
			expected: PaginationMetadata{
				TotalRecords:    100,
				ReturnedRecords: 50,
				Offset:          0,
				Limit:           50,
				HasMore:         true,
			},
		},
		{
			name:         "last page (partial)",
			totalRecords: 100,
			config:       PaginationConfig{Limit: 50, Offset: 90},
			expected: PaginationMetadata{
				TotalRecords:    100,
				ReturnedRecords: 10,
				Offset:          90,
				Limit:           50,
				HasMore:         false,
			},
		},
		{
			name:         "no limit",
			totalRecords: 100,
			config:       PaginationConfig{Limit: 0, Offset: 0},
			expected: PaginationMetadata{
				TotalRecords:    100,
				ReturnedRecords: 100,
				Offset:          0,
				Limit:           0,
				HasMore:         false,
			},
		},
		{
			name:         "offset beyond end",
			totalRecords: 100,
			config:       PaginationConfig{Limit: 10, Offset: 150},
			expected: PaginationMetadata{
				TotalRecords:    100,
				ReturnedRecords: 0,
				Offset:          150,
				Limit:           10,
				HasMore:         false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata := CalculateMetadata(tt.totalRecords, tt.config)

			if metadata.TotalRecords != tt.expected.TotalRecords {
				t.Errorf("TotalRecords: expected %d, got %d", tt.expected.TotalRecords, metadata.TotalRecords)
			}
			if metadata.ReturnedRecords != tt.expected.ReturnedRecords {
				t.Errorf("ReturnedRecords: expected %d, got %d", tt.expected.ReturnedRecords, metadata.ReturnedRecords)
			}
			if metadata.Offset != tt.expected.Offset {
				t.Errorf("Offset: expected %d, got %d", tt.expected.Offset, metadata.Offset)
			}
			if metadata.Limit != tt.expected.Limit {
				t.Errorf("Limit: expected %d, got %d", tt.expected.Limit, metadata.Limit)
			}
			if metadata.HasMore != tt.expected.HasMore {
				t.Errorf("HasMore: expected %v, got %v", tt.expected.HasMore, metadata.HasMore)
			}
		})
	}
}

func TestApplyPaginationToInterfaces(t *testing.T) {
	data := func(n int) []interface{} {
		s := make([]interface{}, n)
		for i := 0; i < n; i++ {
			s[i] = map[string]interface{}{"idx": i}
		}
		return s
	}

	tests := []struct {
		name         string
		data         []interface{}
		offset       int
		pageSize     int
		wantLen      int
		wantTotal    int
		wantReturned int
		wantHasMore  bool
	}{
		{
			name: "no pagination (pageSize=0)",
			data: data(100), offset: 0, pageSize: 0,
			wantLen: 100, wantTotal: 100, wantReturned: 100, wantHasMore: false,
		},
		{
			name: "first page",
			data: data(100), offset: 0, pageSize: 25,
			wantLen: 25, wantTotal: 100, wantReturned: 25, wantHasMore: true,
		},
		{
			name: "middle page",
			data: data(100), offset: 50, pageSize: 25,
			wantLen: 25, wantTotal: 100, wantReturned: 25, wantHasMore: true,
		},
		{
			name: "last page (partial)",
			data: data(100), offset: 90, pageSize: 25,
			wantLen: 10, wantTotal: 100, wantReturned: 10, wantHasMore: false,
		},
		{
			name: "exact last page",
			data: data(100), offset: 75, pageSize: 25,
			wantLen: 25, wantTotal: 100, wantReturned: 25, wantHasMore: false,
		},
		{
			name: "offset beyond end",
			data: data(100), offset: 150, pageSize: 25,
			wantLen: 0, wantTotal: 100, wantReturned: 0, wantHasMore: false,
		},
		{
			name: "negative offset (treated as 0)",
			data: data(100), offset: -5, pageSize: 25,
			wantLen: 25, wantTotal: 100, wantReturned: 25, wantHasMore: true,
		},
		{
			name: "empty data",
			data: data(0), offset: 0, pageSize: 25,
			wantLen: 0, wantTotal: 0, wantReturned: 0, wantHasMore: false,
		},
		{
			name: "offset only (no limit)",
			data: data(100), offset: 30, pageSize: 0,
			wantLen: 70, wantTotal: 100, wantReturned: 70, wantHasMore: false,
		},
		{
			name: "smaller than page",
			data: data(5), offset: 0, pageSize: 25,
			wantLen: 5, wantTotal: 5, wantReturned: 5, wantHasMore: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, meta := ApplyPaginationToInterfaces(tt.data, tt.offset, tt.pageSize)
			if len(result) != tt.wantLen {
				t.Errorf("result length: expected %d, got %d", tt.wantLen, len(result))
			}
			if meta.TotalRecords != tt.wantTotal {
				t.Errorf("TotalRecords: expected %d, got %d", tt.wantTotal, meta.TotalRecords)
			}
			if meta.ReturnedRecords != tt.wantReturned {
				t.Errorf("ReturnedRecords: expected %d, got %d", tt.wantReturned, meta.ReturnedRecords)
			}
			if meta.Offset != tt.offset {
				t.Errorf("Offset: expected %d, got %d", tt.offset, meta.Offset)
			}
			if meta.HasMore != tt.wantHasMore {
				t.Errorf("HasMore: expected %v, got %v", tt.wantHasMore, meta.HasMore)
			}
		})
	}
}

func TestPaginationEdgeCases(t *testing.T) {
	t.Run("empty slice", func(t *testing.T) {
		tools := []types.ToolCall{}
		config := PaginationConfig{Limit: 10, Offset: 0}
		result := ApplyPagination(tools, config)

		if len(result) != 0 {
			t.Errorf("expected empty result, got %d records", len(result))
		}
	})

	t.Run("exact page boundary", func(t *testing.T) {
		tools := generateTestToolCalls(100)
		config := PaginationConfig{Limit: 50, Offset: 50}
		result := ApplyPagination(tools, config)

		if len(result) != 50 {
			t.Errorf("expected 50 records, got %d", len(result))
		}
	})
}

// interfaceData builds an n-element []interface{} for pagination tests.
func interfaceData(n int) []interface{} {
	s := make([]interface{}, n)
	for i := 0; i < n; i++ {
		s[i] = map[string]interface{}{"idx": i}
	}
	return s
}

// TestApplyPaginationToInterfaces_OverflowNoPanic is the DIR-060 regression
// test for the exact reported trigger: offset=2000 with a page_size of
// 2^63-1024 (exactly representable as a JSON float64, so it survives the
// client→server round trip). Before the fix, `start + pageSize` overflowed a
// signed int negative and `data[start:end]` panicked with "slice bounds out
// of range", crashing the MCP server. The page must instead clamp to the
// records actually available beyond the offset.
func TestApplyPaginationToInterfaces_OverflowNoPanic(t *testing.T) {
	const total = 3000
	data := interfaceData(total)

	const offset = 2000
	const pageSize = int64(9223372036854774784) // 2^63 - 1024 (math.MaxInt64 - 1023)

	var result []interface{}
	var meta PaginationMetadata
	assertNoPanic(t, func() {
		result, meta = ApplyPaginationToInterfaces(data, offset, int(pageSize))
	})

	// The page is clamped to the records available beyond the offset.
	if want := total - offset; len(result) != want {
		t.Errorf("result length: expected %d (clamped to available records), got %d", want, len(result))
	}
	if meta.TotalRecords != total {
		t.Errorf("TotalRecords: expected %d, got %d", total, meta.TotalRecords)
	}
	if want := total - offset; meta.ReturnedRecords != want {
		t.Errorf("ReturnedRecords: expected %d, got %d", want, meta.ReturnedRecords)
	}
	if meta.HasMore {
		t.Errorf("HasMore: expected false (page reaches the end), got true")
	}
}

// TestApplyPagination_OverflowNoPanic is the DIR-060 regression test for the
// ToolCall variant: an overflow-inducing Limit must not panic and must clamp
// to the records available beyond the offset.
func TestApplyPagination_OverflowNoPanic(t *testing.T) {
	const total = 3000
	tools := generateTestToolCalls(total)

	var result []types.ToolCall
	assertNoPanic(t, func() {
		result = ApplyPagination(tools, PaginationConfig{Offset: 2000, Limit: math.MaxInt64})
	})

	if want := total - 2000; len(result) != want {
		t.Errorf("result length: expected %d (clamped), got %d", want, len(result))
	}
}

// TestPaginationOverflow_Clamping exercises every overflow site
// (ApplyPagination, ApplyPaginationToInterfaces, CalculateMetadata) with
// pageSize/offset values at and above math.MaxInt64/2 and with negative
// values, asserting no panic and a well-defined clamp in each case.
func TestPaginationOverflow_Clamping(t *testing.T) {
	const total = 100
	data := interfaceData(total)
	tools := generateTestToolCalls(total)

	huge := []int{
		math.MaxInt64 / 2,
		math.MaxInt64 - 1023, // the reported float64-representable trigger
		math.MaxInt64,
	}
	negative := []int{-1, -1024, math.MinInt64}

	// Huge pageSize with a modest offset: page clamps to [offset, total).
	for _, ps := range huge {
		t.Run("huge page size", func(t *testing.T) {
			assertNoPanic(t, func() {
				result, meta := ApplyPaginationToInterfaces(data, 20, ps)
				if want := total - 20; len(result) != want {
					t.Errorf("pageSize=%d: len expected %d, got %d", ps, want, len(result))
				}
				if meta.HasMore {
					t.Errorf("pageSize=%d: HasMore expected false, got true", ps)
				}
			})
			assertNoPanic(t, func() {
				result := ApplyPagination(tools, PaginationConfig{Offset: 20, Limit: ps})
				if want := total - 20; len(result) != want {
					t.Errorf("Limit=%d: len expected %d, got %d", ps, want, len(result))
				}
			})
		})
	}

	// Huge offset (beyond total): clamps to an empty page, never panics.
	for _, off := range huge {
		t.Run("huge offset", func(t *testing.T) {
			assertNoPanic(t, func() {
				result, meta := ApplyPaginationToInterfaces(data, off, 10)
				if len(result) != 0 {
					t.Errorf("offset=%d: expected empty page, got %d records", off, len(result))
				}
				if meta.ReturnedRecords != 0 {
					t.Errorf("offset=%d: ReturnedRecords expected 0, got %d", off, meta.ReturnedRecords)
				}
				if meta.HasMore {
					t.Errorf("offset=%d: HasMore expected false, got true", off)
				}
			})
			assertNoPanic(t, func() {
				result := ApplyPagination(tools, PaginationConfig{Offset: off, Limit: 10})
				if len(result) != 0 {
					t.Errorf("offset=%d: expected empty page, got %d records", off, len(result))
				}
			})
		})
	}

	// Huge offset AND huge pageSize simultaneously (worst-case overflow).
	t.Run("huge offset and huge page size", func(t *testing.T) {
		assertNoPanic(t, func() {
			result, _ := ApplyPaginationToInterfaces(data, math.MaxInt64, math.MaxInt64)
			if len(result) != 0 {
				t.Errorf("expected empty page, got %d records", len(result))
			}
		})
	})

	// Negative values: offset clamps to 0, negative pageSize means "no limit".
	for _, off := range negative {
		t.Run("negative offset", func(t *testing.T) {
			assertNoPanic(t, func() {
				result, _ := ApplyPaginationToInterfaces(data, off, 10)
				if len(result) != 10 {
					t.Errorf("offset=%d: expected 10 records (offset clamped to 0), got %d", off, len(result))
				}
			})
		})
	}
	for _, ps := range negative {
		t.Run("negative page size", func(t *testing.T) {
			assertNoPanic(t, func() {
				result, _ := ApplyPaginationToInterfaces(data, 0, ps)
				if len(result) != total {
					t.Errorf("pageSize=%d: expected all %d records (no limit), got %d", ps, total, len(result))
				}
			})
		})
	}
}

// TestCalculateMetadata_OverflowNoPanic guards the has_more computation, which
// previously computed `Offset+Limit` and overflowed (corrupting has_more).
func TestCalculateMetadata_OverflowNoPanic(t *testing.T) {
	const total = 3000

	assertNoPanic(t, func() {
		meta := CalculateMetadata(total, PaginationConfig{Offset: 2000, Limit: math.MaxInt64})
		// A page that reaches the end has no more records, regardless of how
		// large the requested limit is.
		if meta.HasMore {
			t.Errorf("HasMore: expected false for clamped-to-end page, got true")
		}
		if want := total - 2000; meta.ReturnedRecords != want {
			t.Errorf("ReturnedRecords: expected %d, got %d", want, meta.ReturnedRecords)
		}
	})

	// A genuinely-paged request still reports has_more correctly.
	assertNoPanic(t, func() {
		meta := CalculateMetadata(total, PaginationConfig{Offset: 0, Limit: 100})
		if !meta.HasMore {
			t.Errorf("HasMore: expected true for a first page of 100/%d, got false", total)
		}
	})
}

// assertNoPanic runs fn and fails the test if it panics.
func assertNoPanic(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("unexpected panic: %v", r)
		}
	}()
	fn()
}
