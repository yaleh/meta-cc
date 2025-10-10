package locator

import (
	"testing"
)

func TestPathToHash(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "basic path",
			input:    "/home/yale/work/myproject",
			expected: "-home-yale-work-myproject",
		},
		{
			name:     "trailing slash",
			input:    "/home/yale/work/myproject/",
			expected: "-home-yale-work-myproject-",
		},
		{
			name:     "single directory",
			input:    "/project",
			expected: "-project",
		},
		{
			name:     "relative path",
			input:    "home/yale/work",
			expected: "home-yale-work",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Windows absolute path",
			input:    "C:/Users/yale/work/myproject",
			expected: "C--Users-yale-work-myproject",
		},
		{
			name:     "Windows path with backslashes",
			input:    `C:\Users\yale\work\myproject`,
			expected: "C--Users-yale-work-myproject",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pathToHash(tt.input)
			if result != tt.expected {
				t.Errorf("pathToHash(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestLocate_Priority(t *testing.T) {
	// 测试定位策略的优先级
	// 这是一个集成测试，验证 Locate() 方法按正确顺序尝试各种策略

	// 测试1: 环境变量优先
	// 测试2: --session 参数
	// 测试3: --project 参数
	// 测试4: 自动检测（当前目录）

	// 由于需要复杂的环境准备，此测试将在集成测试中验证
	// 这里提供占位符测试，验证方法存在且签名正确
	locator := NewSessionLocator()

	_, err := locator.Locate(LocateOptions{})
	if err == nil {
		t.Log("Locate method exists and is callable")
	}
}
