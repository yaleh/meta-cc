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

// ApplyPagination applies pagination to ToolCall slice
func ApplyPagination(tools []types.ToolCall, config PaginationConfig) []types.ToolCall {
	// Handle negative offset
	if config.Offset < 0 {
		config.Offset = 0
	}

	// Handle offset beyond length
	if config.Offset >= len(tools) {
		return []types.ToolCall{}
	}

	start := config.Offset
	end := len(tools)

	// Apply limit if specified and positive
	if config.Limit > 0 {
		end = start + config.Limit
		if end > len(tools) {
			end = len(tools)
		}
	}

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

	if normalizedOffset >= totalRecords {
		return []interface{}{}, meta
	}

	start := normalizedOffset
	end := totalRecords

	if pageSize > 0 {
		end = start + pageSize
		if end > totalRecords {
			end = totalRecords
		}
	}

	return data[start:end], meta
}

// CalculateMetadata calculates pagination metadata
func CalculateMetadata(totalRecords int, config PaginationConfig) PaginationMetadata {
	// Handle negative offset
	if config.Offset < 0 {
		config.Offset = 0
	}

	// Calculate returned records
	returned := totalRecords - config.Offset
	if config.Limit > 0 && returned > config.Limit {
		returned = config.Limit
	}
	if returned < 0 {
		returned = 0
	}

	// Calculate hasMore
	hasMore := false
	if config.Limit > 0 {
		hasMore = config.Offset+config.Limit < totalRecords
	}

	return PaginationMetadata{
		TotalRecords:    totalRecords,
		ReturnedRecords: returned,
		Offset:          config.Offset,
		Limit:           config.Limit,
		HasMore:         hasMore,
	}
}
