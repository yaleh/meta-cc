package query

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	queryfiles "github.com/yaleh/meta-cc/internal/query/files"
)

func TestHandleInspectSessionFilesValidation(t *testing.T) {
	tests := []struct {
		name string
		args map[string]interface{}
		want string
	}{
		{"missing files", nil, "files parameter is required"},
		{"files not array", map[string]interface{}{"files": "x"}, "files must be an array"},
		{"non-string file", map[string]interface{}{"files": []interface{}{1}}, "file at index 0 is not a string"},
		{"empty files", map[string]interface{}{"files": []interface{}{}}, "files array cannot be empty"},
		{"samples not bool", map[string]interface{}{"files": []interface{}{"x"}, "include_samples": "yes"}, "include_samples must be a boolean"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := HandleInspectSessionFiles(context.Background(), tt.args)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.want)
		})
	}
}

func TestHandleInspectSessionFilesSuccessAndInspectionError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "session.jsonl")
	require.NoError(t, os.WriteFile(path, []byte("{\"type\":\"user\",\"timestamp\":\"2026-01-01T00:00:00Z\"}\n"), 0o600))
	got, err := HandleInspectSessionFiles(context.Background(), map[string]interface{}{
		"files":           []interface{}{path},
		"include_samples": true,
	})
	require.NoError(t, err)
	assert.NotNil(t, got)

	// DIR-098: an unreadable file is reported as structured data on the
	// response, not raised as an error. The previous assertion here
	// (require.Error + "failed to inspect files") is the fail-fast contract this
	// task replaces: with one bad entry in the list, the caller lost the
	// inspection of every other file.
	missing := filepath.Join(t.TempDir(), "missing.jsonl")
	got, err = HandleInspectSessionFiles(context.Background(), map[string]interface{}{"files": []interface{}{missing}})
	require.NoError(t, err)

	result, ok := got.(*queryfiles.InspectionResult)
	require.True(t, ok, "expected an inspection result, got %T", got)
	require.Len(t, result.MalformedFiles, 1)
	assert.Equal(t, missing, result.MalformedFiles[0].File)
	assert.NotEmpty(t, result.MalformedFiles[0].Reason, "a malformed file must carry a reason string")
	require.Len(t, result.Files, 1)
	assert.False(t, result.Files[0].Parseable)
	assert.NotEmpty(t, result.Files[0].Error)
}

func TestHandleExecuteStage2QueryValidation(t *testing.T) {
	tests := []struct {
		name string
		args map[string]interface{}
		want string
	}{
		{"missing files", nil, "files parameter is required"},
		{"files not array", map[string]interface{}{"files": "x"}, "files must be an array"},
		{"non-string file", map[string]interface{}{"files": []interface{}{1}}, "file at index 0 is not a string"},
		{"empty files", map[string]interface{}{"files": []interface{}{}}, "files array cannot be empty"},
		{"missing filter", map[string]interface{}{"files": []interface{}{"x"}}, "filter parameter is required"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := HandleExecuteStage2Query(context.Background(), tt.args)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.want)
		})
	}
}

func TestHandleExecuteStage2QueryAcceptsIntLimitAndReportsEngineError(t *testing.T) {
	_, err := HandleExecuteStage2Query(context.Background(), map[string]interface{}{
		"files":     []interface{}{filepath.Join(t.TempDir(), "missing.jsonl")},
		"filter":    ".",
		"sort":      1,
		"transform": 1,
		"limit":     2,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to execute stage 2 query")
}
