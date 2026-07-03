package executor

import (
	"context"
	"fmt"
)

func init() {
	registerHandler("query_edit_sequences", handleQueryEditSequences)
}

func handleQueryEditSequences(_ context.Context, e *ToolExecutor, params map[string]interface{}) (string, error) {
	// Validate required 'files' parameter
	var files []interface{}
	if raw, ok := params["files"]; ok {
		if arr, ok := raw.([]interface{}); ok {
			files = arr
		}
	}
	if len(files) == 0 {
		return "", fmt.Errorf("query_edit_sequences: 'files' parameter is required and must be a non-empty array of file paths")
	}

	return e.AnalysisSvc.QueryEditSequences(params)
}
