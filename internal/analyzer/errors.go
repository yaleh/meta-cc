package analyzer

import (
	"crypto/sha256"
	"fmt"
	"regexp"
)

// Package-level compiled regex patterns for error classification.
// Compiled once at init time, not per call.
var (
	// reCommandTimeout matches shell/harness timeout signals, including the
	// exit-code-143 (SIGTERM) convention used when Claude Code kills a command
	// at its timeout. Checked BEFORE the generic exit-code rule so "Exit code
	// 143\nCommand timed out..." is labeled command_timeout, not bash_exit_code.
	reCommandTimeout = regexp.MustCompile(`(?i)timed out|timeout after|ETIMEDOUT|deadline exceeded|exit code 143|exit status 143`)
	// reExitStatus matches both the Go os/exec format ("exit status N") and the
	// Claude Code Bash harness format ("Exit code N" as the first line of every
	// failed Bash call).
	reExitStatus       = regexp.MustCompile(`(?i)(?:exit status|exit code) \d+`)
	reCommandNotFound  = regexp.MustCompile(`(?i)command not found`)
	rePermissionDenied = regexp.MustCompile(`(?i)permission denied`)
	reFileNotFound     = regexp.MustCompile(`(?i)no such file|file not found|ENOENT|cannot find the file|cannot read`)
	// reConnectionError matches genuine network failure signals only. Bare
	// "timeout"/"network" words are deliberately excluded so incidental body
	// text (e.g. a config dump mentioning "timeoutMs") cannot trigger it.
	reConnectionError  = regexp.MustCompile(`(?i)connection refused|connection reset|ECONNREFUSED|ECONNRESET|network(?: is)? unreachable|dial tcp`)
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
//  1. empty error          → tool_error_no_message
//  2. timeout signals      → command_timeout
//     (timed out | timeout after | ETIMEDOUT | deadline exceeded |
//     exit code 143 | exit status 143)
//  3. exit status N | exit code N → bash_exit_code
//     (both Go os/exec's "exit status N" and the Claude Code harness's
//     "Exit code N" first-line format)
//  4. command not found    → command_not_found
//  5. permission denied    → permission_denied
//  6. no such file, file not found, ENOENT, cannot find/read → file_not_found
//  7. network failures     → connection_error
//     (connection refused/reset, ECONNREFUSED, ECONNRESET, network
//     unreachable, dial tcp — genuine signals only)
//  8. parse/syntax errors  → parse_error
//  9. content too large    → content_too_large
//  10. auth errors         → auth_error
//  11. not found / 404     → resource_not_found
//  12. fallback            → uncategorized
//
// Precedence note: command_timeout (rule 2) is deliberately checked BEFORE the
// generic exit-code rule (rule 3) and connection_error (rule 7). Claude Code
// emits "Exit code 143\nCommand timed out after Ns..." for killed/timed-out
// commands; without this ordering the generic exit-code rule would claim it as
// bash_exit_code and the old loose timeout|network connection rule mislabeled
// it (and incidental body text like "timeoutMs") as connection_error. The
// timeout rule also claims exit code/status 143 explicitly, so a bare
// "Exit code 143" with no body text is still a command_timeout.
func ClassifyErrorType(toolName, errorText string) string {
	if errorText == "" {
		return "tool_error_no_message"
	}
	if reCommandTimeout.MatchString(errorText) {
		return "command_timeout"
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
