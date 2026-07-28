package filter

import "github.com/yaleh/meta-cc/internal/types"

// PaginationConfig defines pagination parameters
type PaginationConfig struct {
	Limit  int // 0 means no limit
	Offset int
}

// PaginationMetadata contains pagination information
type PaginationMetadata struct {
	TotalRecords    int  `json:"total_records"`
	ReturnedRecords int  `json:"returned_records"`
	Offset          int  `json:"offset"`
	Limit           int  `json:"limit"`
	HasMore         bool `json:"has_more"`
}

// clampPage computes overflow-safe half-open [start, end) slice bounds for
// paginating a collection of total records. It is the single sanitization
// point shared by ApplyPagination, ApplyPaginationToInterfaces, and
// CalculateMetadata so their bounds logic cannot diverge (DIR-060).
//
// offset is clamped to [0, total]; a non-positive pageSize means "no limit"
// (end == total). The page length is derived as min(pageSize, total-start) —
// never as start+pageSize — so computing end can never overflow a signed int:
// both start and the clamped length are ≤ total, so end = start+length ≤ total.
// This eliminates the `start + pageSize` overflow that previously produced a
// negative end and a "slice bounds out of range" panic (a server-crashing DoS
// triggerable by any client via a large page_size/offset).
func clampPage(total, offset, pageSize int) (start, end int) {
	if total <= 0 {
		return 0, 0
	}

	// Clamp offset into [0, total].
	start = offset
	if start < 0 {
		start = 0
	}
	if start > total {
		start = total
	}

	// remaining is the number of records from start to the end of the
	// collection; it is always ≥ 0 because start ≤ total by construction.
	remaining := total - start
	if pageSize <= 0 || pageSize > remaining {
		end = total
	} else {
		end = start + pageSize // safe: pageSize ≤ remaining == total-start
	}
	return start, end
}

// ApplyPagination applies pagination to ToolCall slice
func ApplyPagination(tools []types.ToolCall, config PaginationConfig) []types.ToolCall {
	start, end := clampPage(len(tools), config.Offset, config.Limit)
	return tools[start:end]
}

// ApplyPaginationToInterfaces applies pagination to an interface slice and returns
// the sliced data along with pagination metadata.
func ApplyPaginationToInterfaces(data []interface{}, offset, pageSize int) ([]interface{}, PaginationMetadata) {
	totalRecords := len(data)

	normalizedOffset := offset
	if normalizedOffset < 0 {
		normalizedOffset = 0
	}

	config := PaginationConfig{Offset: normalizedOffset, Limit: pageSize}
	meta := CalculateMetadata(totalRecords, config)
	meta.Offset = offset

	start, end := clampPage(totalRecords, normalizedOffset, pageSize)
	return data[start:end], meta
}

// CalculateMetadata calculates pagination metadata
func CalculateMetadata(totalRecords int, config PaginationConfig) PaginationMetadata {
	offset := config.Offset
	if offset < 0 {
		offset = 0
	}

	// Derive returned/hasMore from the same overflow-safe bounds used for
	// slicing. ReturnedRecords is end-start and HasMore is end<total — both
	// trivially overflow-safe since 0 ≤ start ≤ end ≤ totalRecords. This
	// replaces the old `Offset+Limit` has_more computation, which overflowed
	// a signed int and corrupted the flag for large limits (DIR-060).
	start, end := clampPage(totalRecords, offset, config.Limit)

	return PaginationMetadata{
		TotalRecords:    totalRecords,
		ReturnedRecords: end - start,
		Offset:          offset,
		Limit:           config.Limit,
		HasMore:         end < totalRecords,
	}
}
