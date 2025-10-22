package cmd

import (
	"bytes"
	"path/filepath"
	"testing"
)

func TestValidateAPICommand(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	toolsPath := filepath.Join("..", "cmd", "mcp-server", "tools.go")
	rootCmd.SetArgs([]string{"validate", "api", "--file", toolsPath, "--quiet"})
	defer rootCmd.SetArgs(nil)

	err := rootCmd.Execute()
	if err == nil {
		t.Fatalf("expected validation failure, got success. Output: %s", buf.String())
	}

	if got := err.Error(); got == "" || !bytes.Contains([]byte(got), []byte("validation failed")) {
		t.Fatalf("unexpected error: %v", err)
	}
}
