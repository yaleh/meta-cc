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
