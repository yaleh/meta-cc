package filter

import "github.com/yale/meta-cc/internal/parser"

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
func ApplyPagination(tools []parser.ToolCall, config PaginationConfig) []parser.ToolCall {
	// Handle negative offset
	if config.Offset < 0 {
		config.Offset = 0
	}

	// Handle offset beyond length
	if config.Offset >= len(tools) {
		return []parser.ToolCall{}
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
