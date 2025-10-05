package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yale/meta-cc/internal/testutil"
)

func TestAnalyzeErrorsCommand_NoErrors(t *testing.T) {
	// 准备测试环境：创建无错误的会话文件
	homeDir, _ := os.UserHomeDir()
	projectHash := "-home-yale-work-test-analyze-no-errors"
	sessionID := "test-session-no-errors"

	sessionDir := filepath.Join(homeDir, ".claude", "projects", projectHash)
	os.MkdirAll(sessionDir, 0755)
	sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")

	// 使用仅包含成功工具调用的 fixture
	fixtureContent := testutil.LoadFixture(t, "sample-session.jsonl")
	os.WriteFile(sessionFile, fixtureContent, 0644)
	defer os.RemoveAll(sessionDir)

	// 设置环境变量
	os.Setenv("CC_SESSION_ID", sessionID)
	os.Setenv("CC_PROJECT_HASH", projectHash)
	defer os.Unsetenv("CC_SESSION_ID")
	defer os.Unsetenv("CC_PROJECT_HASH")

	// 捕获输出
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"analyze", "errors"})

	// 执行命令
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	output := buf.String()

	// 验证输出包含空模式数组或 "No error patterns detected"
	if !strings.Contains(output, "[]") && !strings.Contains(output, "No error patterns detected") {
		t.Errorf("Expected empty result or 'No error patterns detected', got: %s", output)
	}
}

func TestAnalyzeErrorsCommand_WithErrors(t *testing.T) {
	// 准备测试环境：创建包含重复错误的会话文件
	homeDir, _ := os.UserHomeDir()
	projectHash := "-home-yale-work-test-analyze-with-errors"
	sessionID := "test-session-with-errors"

	sessionDir := filepath.Join(homeDir, ".claude", "projects", projectHash)
	os.MkdirAll(sessionDir, 0755)
	sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")

	// 使用包含重复错误的 fixture
	fixtureContent := testutil.LoadFixture(t, "session-with-errors.jsonl")
	os.WriteFile(sessionFile, fixtureContent, 0644)
	defer os.RemoveAll(sessionDir)

	os.Setenv("CC_SESSION_ID", sessionID)
	os.Setenv("CC_PROJECT_HASH", projectHash)
	defer os.Unsetenv("CC_SESSION_ID")
	defer os.Unsetenv("CC_PROJECT_HASH")

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"analyze", "errors", "--output", "jsonl"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	output := buf.String()

	// 验证输出包含错误模式字段
	expectedFields := []string{
		"pattern_id",
		"type",
		"occurrences",
		"signature",
		"error_text",
	}

	for _, field := range expectedFields {
		if !strings.Contains(output, field) {
			t.Errorf("Expected output to contain '%s', got: %s", field, output)
		}
	}
}

func TestAnalyzeErrorsCommand_WithWindow(t *testing.T) {
	// 测试 --window 参数
	homeDir, _ := os.UserHomeDir()
	projectHash := "-home-yale-work-test-analyze-window"
	sessionID := "test-session-window"

	sessionDir := filepath.Join(homeDir, ".claude", "projects", projectHash)
	os.MkdirAll(sessionDir, 0755)
	sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")

	fixtureContent := testutil.LoadFixture(t, "session-with-errors.jsonl")
	os.WriteFile(sessionFile, fixtureContent, 0644)
	defer os.RemoveAll(sessionDir)

	os.Setenv("CC_SESSION_ID", sessionID)
	os.Setenv("CC_PROJECT_HASH", projectHash)
	defer os.Unsetenv("CC_SESSION_ID")
	defer os.Unsetenv("CC_PROJECT_HASH")

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"analyze", "errors", "--window", "10"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	// 命令应成功执行（结果取决于 fixture 中最后 10 个 Turn）
	output := buf.String()
	if output == "" {
		t.Error("Expected non-empty output")
	}
}

func TestAnalyzeErrorsCommand_TSVOutput(t *testing.T) {
	// 测试 TSV 输出格式
	homeDir, _ := os.UserHomeDir()
	projectHash := "-home-yale-work-test-analyze-md"
	sessionID := "test-session-md"

	sessionDir := filepath.Join(homeDir, ".claude", "projects", projectHash)
	os.MkdirAll(sessionDir, 0755)
	sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")

	fixtureContent := testutil.LoadFixture(t, "session-with-errors.jsonl")
	os.WriteFile(sessionFile, fixtureContent, 0644)
	defer os.RemoveAll(sessionDir)

	os.Setenv("CC_SESSION_ID", sessionID)
	os.Setenv("CC_PROJECT_HASH", projectHash)
	defer os.Unsetenv("CC_SESSION_ID")
	defer os.Unsetenv("CC_PROJECT_HASH")

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"analyze", "errors", "--output", "tsv"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	output := buf.String()

	// 验证 TSV/JSONL 格式 - just check we got some output
	if output == "" {
		t.Error("Expected non-empty output")
	}
}

func TestAnalyzeErrorsCommand_MissingSessionFile(t *testing.T) {
	// 清除环境变量
	os.Unsetenv("CC_SESSION_ID")
	os.Unsetenv("CC_PROJECT_HASH")

	var buf bytes.Buffer
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"analyze", "errors"})

	err := rootCmd.Execute()
	if err == nil {
		t.Error("Expected error when session file not found")
	}
}
