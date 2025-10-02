package analyzer

import (
	"crypto/sha256"
	"fmt"
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
