package analyzer

import (
	"testing"
)

func TestCalculateErrorSignature_SameError(t *testing.T) {
	// 相同错误应生成相同签名
	toolName := "Bash"
	errorText := "command not found: xyz"

	sig1 := CalculateErrorSignature(toolName, errorText)
	sig2 := CalculateErrorSignature(toolName, errorText)

	if sig1 != sig2 {
		t.Errorf("Expected same signature for identical errors, got %s and %s", sig1, sig2)
	}

	if sig1 == "" {
		t.Error("Expected non-empty signature")
	}
}

func TestCalculateErrorSignature_DifferentErrors(t *testing.T) {
	// 不同错误应生成不同签名
	toolName := "Bash"
	error1 := "command not found: xyz"
	error2 := "permission denied"

	sig1 := CalculateErrorSignature(toolName, error1)
	sig2 := CalculateErrorSignature(toolName, error2)

	if sig1 == sig2 {
		t.Errorf("Expected different signatures for different errors, got same: %s", sig1)
	}
}

func TestCalculateErrorSignature_DifferentTools(t *testing.T) {
	// 不同工具的相同错误文本应生成不同签名
	errorText := "file not found"

	sig1 := CalculateErrorSignature("Read", errorText)
	sig2 := CalculateErrorSignature("Write", errorText)

	if sig1 == sig2 {
		t.Errorf("Expected different signatures for different tools, got same: %s", sig1)
	}
}

func TestCalculateErrorSignature_LongErrorText(t *testing.T) {
	// 测试长错误文本（仅取前 100 个字符）
	toolName := "Bash"
	longError := ""
	for i := 0; i < 200; i++ {
		longError += "a"
	}

	sig := CalculateErrorSignature(toolName, longError)

	if sig == "" {
		t.Error("Expected non-empty signature for long error text")
	}

	// 验证截断：前100字符和完整文本应该产生相同签名（因为只取前100）
	sig2 := CalculateErrorSignature(toolName, longError[:100])
	if sig != sig2 {
		t.Errorf("Expected same signature for truncated text, got %s and %s", sig, sig2)
	}
}

func TestCalculateErrorSignature_EmptyError(t *testing.T) {
	// 空错误文本应生成签名（基于工具名）
	toolName := "Read"
	errorText := ""

	sig := CalculateErrorSignature(toolName, errorText)

	if sig == "" {
		t.Error("Expected non-empty signature even for empty error text")
	}
}

func TestCalculateErrorSignature_SignatureFormat(t *testing.T) {
	// 验证签名格式：16个字符的十六进制字符串
	toolName := "Bash"
	errorText := "test error"

	sig := CalculateErrorSignature(toolName, errorText)

	if len(sig) != 16 {
		t.Errorf("Expected signature length of 16 characters, got %d", len(sig))
	}

	// 验证是十六进制字符串
	for _, c := range sig {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("Expected hexadecimal string, got character '%c' in signature %s", c, sig)
		}
	}
}

func TestClassifyErrorType_EmptyError(t *testing.T) {
	result := ClassifyErrorType("Bash", "")
	if result != "tool_error_no_message" {
		t.Errorf("Expected 'tool_error_no_message' for empty error, got %q", result)
	}
}

func TestClassifyErrorType_BashExitCode(t *testing.T) {
	cases := []string{
		"exit status 1",
		"exit status 127",
		"exit status 0",
	}
	for _, tc := range cases {
		result := ClassifyErrorType("Bash", tc)
		if result != "bash_exit_code" {
			t.Errorf("Expected 'bash_exit_code' for %q, got %q", tc, result)
		}
	}
}

func TestClassifyErrorType_CommandNotFound(t *testing.T) {
	cases := []string{
		"command not found: xyz",
		"bash: npm: command not found",
	}
	for _, tc := range cases {
		result := ClassifyErrorType("Bash", tc)
		if result != "command_not_found" {
			t.Errorf("Expected 'command_not_found' for %q, got %q", tc, result)
		}
	}
}

func TestClassifyErrorType_PermissionDenied(t *testing.T) {
	cases := []string{
		"permission denied",
		"Permission denied: /etc/shadow",
	}
	for _, tc := range cases {
		result := ClassifyErrorType("Bash", tc)
		if result != "permission_denied" {
			t.Errorf("Expected 'permission_denied' for %q, got %q", tc, result)
		}
	}
}

func TestClassifyErrorType_FileNotFound(t *testing.T) {
	cases := []string{
		"no such file or directory",
		"file not found",
		"ENOENT: no such file or directory",
		"cannot find the file",
		"cannot read the file",
	}
	for _, tc := range cases {
		result := ClassifyErrorType("Read", tc)
		if result != "file_not_found" {
			t.Errorf("Expected 'file_not_found' for %q, got %q", tc, result)
		}
	}
}

func TestClassifyErrorType_ConnectionError(t *testing.T) {
	cases := []string{
		"connection refused",
		"connection reset by peer",
		"connection timeout",
		"network is unreachable",
		"ECONNREFUSED",
	}
	for _, tc := range cases {
		result := ClassifyErrorType("Read", tc)
		if result != "connection_error" {
			t.Errorf("Expected 'connection_error' for %q, got %q", tc, result)
		}
	}
}

func TestClassifyErrorType_ParseError(t *testing.T) {
	cases := []string{
		"not a valid input",
		"parse error: unexpected EOF",
		"syntax error near line 5",
		"invalid json: unexpected end",
		"unmarshal error: invalid character",
		"unexpected token at position 42",
	}
	for _, tc := range cases {
		result := ClassifyErrorType("Bash", tc)
		if result != "parse_error" {
			t.Errorf("Expected 'parse_error' for %q, got %q", tc, result)
		}
	}
}

func TestClassifyErrorType_ContentTooLarge(t *testing.T) {
	cases := []string{
		"file too large: exceeds maximum size",
		"content too long for processing",
		"file too big to read",
		"exceeds limit of 100MB",
	}
	for _, tc := range cases {
		result := ClassifyErrorType("Read", tc)
		if result != "content_too_large" {
			t.Errorf("Expected 'content_too_large' for %q, got %q", tc, result)
		}
	}
}

func TestClassifyErrorType_AuthError(t *testing.T) {
	cases := []string{
		"unauthorized access",
		"authentication failed",
		"not authenticated",
		"HTTP 403 Forbidden",
		"401 Unauthorized",
	}
	for _, tc := range cases {
		result := ClassifyErrorType("Read", tc)
		if result != "auth_error" {
			t.Errorf("Expected 'auth_error' for %q, got %q", tc, result)
		}
	}
}

func TestClassifyErrorType_ResourceNotFound(t *testing.T) {
	cases := []string{
		"not found",
		"404 page not found",
		"no such resource",
	}
	for _, tc := range cases {
		result := ClassifyErrorType("Read", tc)
		if result != "resource_not_found" {
			t.Errorf("Expected 'resource_not_found' for %q, got %q", tc, result)
		}
	}
}

func TestClassifyErrorType_Uncategorized(t *testing.T) {
	cases := []string{
		"some random error message",
		"signal: killed",
		"out of memory",
	}
	for _, tc := range cases {
		result := ClassifyErrorType("Bash", tc)
		if result != "uncategorized" {
			t.Errorf("Expected 'uncategorized' for %q, got %q", tc, result)
		}
	}
}

func TestClassifyErrorType_FirstMatchWins(t *testing.T) {
	// "no such file" should match file_not_found (rule 5), not resource_not_found (rule 10)
	result := ClassifyErrorType("Bash", "no such file or directory")
	if result != "file_not_found" {
		t.Errorf("Expected 'file_not_found' due to first-match ordering, got %q", result)
	}

	// "file not found" should match file_not_found (rule 5), not resource_not_found (rule 10)
	result = ClassifyErrorType("Bash", "file not found: /tmp/missing")
	if result != "file_not_found" {
		t.Errorf("Expected 'file_not_found' due to first-match ordering, got %q", result)
	}
}
