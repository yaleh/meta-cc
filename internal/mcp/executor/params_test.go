package executor_test

import (
	"testing"

	"github.com/yaleh/meta-cc/internal/mcp/executor"
)

func TestGetStringParam(t *testing.T) {
	t.Parallel()
	args := map[string]interface{}{
		"key": "value",
	}
	if got := executor.GetStringParam(args, "key", "default"); got != "value" {
		t.Errorf("expected 'value', got %q", got)
	}
	if got := executor.GetStringParam(args, "missing", "default"); got != "default" {
		t.Errorf("expected 'default', got %q", got)
	}
}

func TestGetBoolParam(t *testing.T) {
	t.Parallel()
	args := map[string]interface{}{
		"key": true,
	}
	if got := executor.GetBoolParam(args, "key", false); !got {
		t.Error("expected true")
	}
	if got := executor.GetBoolParam(args, "missing", false); got {
		t.Error("expected false for missing key")
	}
}

func TestGetIntParam(t *testing.T) {
	t.Parallel()
	args := map[string]interface{}{
		"float_key": float64(42),
		"int_key":   10,
	}
	if got := executor.GetIntParam(args, "float_key", 0); got != 42 {
		t.Errorf("expected 42, got %d", got)
	}
	if got := executor.GetIntParam(args, "int_key", 0); got != 10 {
		t.Errorf("expected 10, got %d", got)
	}
	if got := executor.GetIntParam(args, "missing", 99); got != 99 {
		t.Errorf("expected 99, got %d", got)
	}
}

func TestGetFloatParam(t *testing.T) {
	t.Parallel()
	args := map[string]interface{}{
		"key": float64(3.14),
	}
	if got := executor.GetFloatParam(args, "key", 0.0); got != 3.14 {
		t.Errorf("expected 3.14, got %f", got)
	}
	if got := executor.GetFloatParam(args, "missing", 1.0); got != 1.0 {
		t.Errorf("expected 1.0, got %f", got)
	}
}
