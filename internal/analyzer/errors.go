package analyzer

import (
	"crypto/sha256"
	"fmt"
	"regexp"
)

// Package-level compiled regex patterns for error classification.
// Compiled once at init time, not per call.
var (
	reExitStatus       = regexp.MustCompile(`(?i)exit status \d+`)
	reCommandNotFound  = regexp.MustCompile(`(?i)command not found`)
	rePermissionDenied = regexp.MustCompile(`(?i)permission denied`)
	reFileNotFound     = regexp.MustCompile(`(?i)no such file|file not found|ENOENT|cannot find the file|cannot read`)
	reConnectionError  = regexp.MustCompile(`(?i)connection refused|connection reset|timeout|network|ECONNREFUSED`)
	reParseError       = regexp.MustCompile(`(?i)not a valid|parse error|syntax error|invalid.*json|unmarshal|unexpected token`)
	reContentTooLarge  = regexp.MustCompile(`(?i)too large|file too big|exceeds limit|content too long`)
	reAuthError        = regexp.MustCompile(`(?i)unauthorized|authentication|not authenticated|403|401`)
	reResourceNotFound = regexp.MustCompile(`(?i)not found|404|no such`)
)

// CalculateErrorSignature 计算错误签名
// 签名基于工具名和错误文本的前 100 个字符的哈希值
// 相同的错误类型会生成相同的签名，用于模式检测
func CalculateErrorSignature(toolName, errorText string) string {
	// 限制错误文本长度为前 100 个字符
	truncatedError := errorText
	if len(errorText) > 100 {
		truncatedError = errorText[:100]
	}

	// 组合工具名和错误文本
	combined := fmt.Sprintf("%s:%s", toolName, truncatedError)

	// 计算 SHA256 哈希
	hash := sha256.Sum256([]byte(combined))

	// 返回哈希的十六进制表示（前 16 个字符作为签名）
	return fmt.Sprintf("%x", hash)[:16]
}

// ClassifyErrorType returns a human-readable label for an error based on its
// tool name and error text. Classification rules are applied first-match:
//
//  1. empty error        → tool_error_no_message
//  2. exit status N      → bash_exit_code
//  3. command not found  → command_not_found
//  4. permission denied  → permission_denied
//  5. no such file, file not found, ENOENT, cannot find/read → file_not_found
//  6. connection errors  → connection_error
//  7. parse/syntax errors → parse_error
//  8. content too large  → content_too_large
//  9. auth errors        → auth_error
//  10. not found / 404   → resource_not_found
//  11. fallback          → uncategorized
func ClassifyErrorType(toolName, errorText string) string {
	if errorText == "" {
		return "tool_error_no_message"
	}
	if reExitStatus.MatchString(errorText) {
		return "bash_exit_code"
	}
	if reCommandNotFound.MatchString(errorText) {
		return "command_not_found"
	}
	if rePermissionDenied.MatchString(errorText) {
		return "permission_denied"
	}
	if reFileNotFound.MatchString(errorText) {
		return "file_not_found"
	}
	if reConnectionError.MatchString(errorText) {
		return "connection_error"
	}
	if reParseError.MatchString(errorText) {
		return "parse_error"
	}
	if reContentTooLarge.MatchString(errorText) {
		return "content_too_large"
	}
	if reAuthError.MatchString(errorText) {
		return "auth_error"
	}
	if reResourceNotFound.MatchString(errorText) {
		return "resource_not_found"
	}
	return "uncategorized"
}
