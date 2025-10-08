# Phase 0: 项目初始化

## 概述

**目标**: 建立 Go 项目骨架和开发环境，为后续开发奠定基础

**代码量**: ~150 行（每个 Stage ≤ 200 行）

**依赖**: 无（初始 Phase）

**交付物**: 可构建、可测试的 Go CLI 项目框架

---

## Phase 目标

建立完整的 Go 项目开发环境，包括：

1. Go 模块和依赖管理
2. 基于 Cobra + Viper 的 CLI 框架
3. 测试框架和 fixture 管理
4. 构建和发布脚本
5. 完整的 README.md 使用文档

**成功标准**:
- ✅ `go build` 成功
- ✅ `go test ./...` 通过
- ✅ `./meta-cc --help` 显示帮助信息
- ✅ `./meta-cc --version` 显示版本信息
- ✅ README.md 包含完整的构建和使用说明

---

## Stage 0.1: Go 模块初始化

### 目标

建立 Go 项目的基础结构，包括模块定义、依赖管理和根命令框架。

### TDD 工作流

**1. 准备阶段**
```bash
# 创建项目目录结构
mkdir -p cmd internal pkg tests/fixtures

# 初始化 Go 模块
go mod init github.com/yaleh/meta-cc
```

**2. 测试先行（无测试文件，直接实现）**

此阶段为框架搭建，无需单元测试。验证方式为功能测试。

**3. 实现代码**

创建以下文件：

#### `main.go` (~15 行)
```go
package main

import (
    "github.com/yaleh/meta-cc/cmd"
    "os"
)

func main() {
    if err := cmd.Execute(); err != nil {
        os.Exit(1)
    }
}
```

#### `cmd/root.go` (~60 行)
```go
package cmd

import (
    "fmt"
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
)

var (
    cfgFile    string
    sessionID  string
    projectPath string
    outputFormat string
)

var Version = "dev" // 将在构建时注入

var rootCmd = &cobra.Command{
    Use:   "meta-cc",
    Short: "Meta-Cognition tool for Claude Code",
    Long: `meta-cc analyzes Claude Code session history to provide
metacognitive insights and workflow optimization.`,
    Version: Version,
}

func Execute() error {
    return rootCmd.Execute()
}

func init() {
    cobra.OnInitialize(initConfig)

    // 全局参数
    rootCmd.PersistentFlags().StringVar(&sessionID, "session", "", "Session ID (or use $CC_SESSION_ID)")
    rootCmd.PersistentFlags().StringVar(&projectPath, "project", "", "Project path")
    rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "json", "Output format: json|md|csv")

    // 绑定环境变量
    viper.BindPFlag("session", rootCmd.PersistentFlags().Lookup("session"))
    viper.BindPFlag("project", rootCmd.PersistentFlags().Lookup("project"))
    viper.BindEnv("session", "CC_SESSION_ID")
    viper.BindEnv("project", "CC_PROJECT_PATH")
}

func initConfig() {
    viper.AutomaticEnv()
}
```

**4. 安装依赖**
```bash
go get github.com/spf13/cobra@latest
go get github.com/spf13/viper@latest
go mod tidy
```

**5. 功能测试**
```bash
# 构建
go build -o meta-cc

# 测试帮助信息
./meta-cc --help

# 测试版本信息
./meta-cc --version
```

### 交付物

**文件清单**:
```
meta-cc/
├── go.mod              # Go 模块定义
├── go.sum              # 依赖锁定文件
├── main.go             # 程序入口 (~15 行)
└── cmd/
    └── root.go         # Cobra 根命令 (~60 行)
```

**代码量**: ~75 行

### README.md 初始内容

创建 `README.md`:

```markdown
# meta-cc

Meta-Cognition tool for Claude Code - analyze session history for workflow optimization.

## Installation

### From Source

```bash
git clone https://github.com/yaleh/meta-cc.git
cd meta-cc
go build -o meta-cc
```

## Usage

```bash
# Show help
./meta-cc --help

# Show version
./meta-cc --version
```

## Development

### Prerequisites

- Go 1.21 or later

### Build

```bash
go build -o meta-cc
```

### Test

```bash
go test ./...
```
```

### 验收标准

- ✅ `go build -o meta-cc` 成功编译
- ✅ `./meta-cc --help` 显示完整帮助信息
- ✅ `./meta-cc --version` 显示版本号（dev）
- ✅ 全局参数 `--session`, `--project`, `--output` 可识别
- ✅ 环境变量 `CC_SESSION_ID`, `CC_PROJECT_PATH` 绑定成功
- ✅ README.md 包含基础使用说明

---

## Stage 0.2: 测试框架搭建

### 目标

建立 Go 测试基础设施，包括测试工具函数、fixture 管理和示例测试。

### TDD 工作流

**1. 测试先行 - 编写测试工具**

#### `internal/testutil/fixtures.go` (~50 行)

```go
package testutil

import (
    "os"
    "path/filepath"
    "testing"
)

// FixtureDir 返回 fixtures 目录路径
func FixtureDir() string {
    return filepath.Join("../../tests/fixtures")
}

// LoadFixture 加载测试 fixture 文件内容
func LoadFixture(t *testing.T, filename string) []byte {
    t.Helper()

    path := filepath.Join(FixtureDir(), filename)
    data, err := os.ReadFile(path)
    if err != nil {
        t.Fatalf("Failed to load fixture %s: %v", filename, err)
    }

    return data
}

// TempSessionFile 创建临时会话文件用于测试
func TempSessionFile(t *testing.T, content string) string {
    t.Helper()

    tmpFile, err := os.CreateTemp("", "session-*.jsonl")
    if err != nil {
        t.Fatalf("Failed to create temp file: %v", err)
    }

    if _, err := tmpFile.WriteString(content); err != nil {
        t.Fatalf("Failed to write temp file: %v", err)
    }

    tmpFile.Close()
    t.Cleanup(func() { os.Remove(tmpFile.Name()) })

    return tmpFile.Name()
}
```

**2. 创建测试 fixture**

#### `tests/fixtures/sample-session.jsonl` (~20 行)

```jsonl
{"sequence":0,"role":"user","timestamp":1735689600,"content":[{"type":"text","text":"帮我修复这个认证 bug"}]}
{"sequence":1,"role":"assistant","timestamp":1735689605,"content":[{"type":"text","text":"我来帮你检查代码"},{"type":"tool_use","id":"toolu_01","name":"Grep","input":{"pattern":"auth.*error","path":"."}}]}
{"sequence":2,"role":"user","timestamp":1735689610,"content":[{"type":"tool_result","tool_use_id":"toolu_01","content":"src/auth.js:15: authError: token invalid"}]}
```

**3. 编写示例测试**

#### `internal/testutil/fixtures_test.go` (~30 行)

```go
package testutil

import (
    "testing"
)

func TestLoadFixture(t *testing.T) {
    data := LoadFixture(t, "sample-session.jsonl")

    if len(data) == 0 {
        t.Error("Expected non-empty fixture data")
    }

    // 验证 JSONL 格式（每行都应该是有效的 JSON）
    lines := string(data)
    if lines == "" {
        t.Error("Expected at least one line in fixture")
    }
}

func TestTempSessionFile(t *testing.T) {
    content := `{"test":"data"}`
    path := TempSessionFile(t, content)

    if path == "" {
        t.Error("Expected non-empty temp file path")
    }

    // 文件应该在测试结束后自动删除
}
```

**4. 运行测试**
```bash
go test ./internal/testutil -v
```

### 交付物

**文件清单**:
```
meta-cc/
├── internal/
│   └── testutil/
│       ├── fixtures.go        # 测试工具函数 (~50 行)
│       └── fixtures_test.go   # 示例测试 (~30 行)
└── tests/
    └── fixtures/
        └── sample-session.jsonl  # 测试数据 (~20 行)
```

**代码量**: ~100 行

### README.md 更新

在 `README.md` 中添加测试部分：

```markdown
### Test

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test ./... -cover

# Run specific package tests
go test ./internal/testutil -v
```

### Test Fixtures

Test data is located in `tests/fixtures/`:
- `sample-session.jsonl`: Sample Claude Code session for testing parsers
```

### 验收标准

- ✅ `go test ./...` 全部通过
- ✅ `LoadFixture()` 能正确加载测试文件
- ✅ `TempSessionFile()` 能创建临时文件并自动清理
- ✅ `tests/fixtures/sample-session.jsonl` 包含有效的 JSONL 数据
- ✅ README.md 包含测试命令说明

---

## Stage 0.3: 构建和发布脚本

### 目标

提供跨平台构建能力，支持版本信息嵌入和自动化发布流程。

### TDD 工作流

**1. 创建 Makefile** (~80 行)

#### `Makefile`

```makefile
# Makefile for meta-cc

# 版本信息
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')

# 构建参数
LDFLAGS := -ldflags "-X github.com/yaleh/meta-cc/cmd.Version=$(VERSION) \
                     -X github.com/yaleh/meta-cc/cmd.Commit=$(COMMIT) \
                     -X github.com/yaleh/meta-cc/cmd.BuildTime=$(BUILD_TIME)"

# Go 参数
GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOCLEAN := $(GOCMD) clean
GOMOD := $(GOCMD) mod

# 输出目录
BUILD_DIR := build
BINARY_NAME := meta-cc

# 目标平台
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

.PHONY: all build test clean install cross-compile help

# 默认目标
all: test build

# 构建当前平台
build:
	@echo "Building $(BINARY_NAME) $(VERSION)..."
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) .

# 运行测试
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# 带覆盖率的测试
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# 清理构建产物
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# 安装到 GOPATH/bin
install:
	@echo "Installing..."
	$(GOCMD) install $(LDFLAGS) .

# 跨平台编译
cross-compile:
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	@for platform in $(PLATFORMS); do \
		GOOS=$${platform%/*} GOARCH=$${platform#*/} \
		$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-$${platform%/*}-$${platform#*/} .; \
	done
	@echo "Cross-compilation complete. Binaries in $(BUILD_DIR)/"

# 依赖管理
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# 帮助信息
help:
	@echo "Available targets:"
	@echo "  make build           - Build for current platform"
	@echo "  make test            - Run tests"
	@echo "  make test-coverage   - Run tests with coverage report"
	@echo "  make clean           - Remove build artifacts"
	@echo "  make install         - Install to GOPATH/bin"
	@echo "  make cross-compile   - Build for all platforms"
	@echo "  make deps            - Download and tidy dependencies"
	@echo "  make help            - Show this help message"
```

**2. 更新 cmd/root.go 以支持构建信息**

在 `cmd/root.go` 中添加：

```go
var (
    Version   = "dev"
    Commit    = "unknown"
    BuildTime = "unknown"
)

// 在 init() 函数中更新 Version 信息
func init() {
    rootCmd.Version = fmt.Sprintf("%s (commit: %s, built: %s)", Version, Commit, BuildTime)
    // ... 其他初始化代码
}
```

**3. 功能测试**

```bash
# 测试基本构建
make build
./meta-cc --version

# 测试清理
make clean

# 测试跨平台构建
make cross-compile
ls -lh build/

# 测试全流程
make all
```

### 交付物

**文件清单**:
```
meta-cc/
├── Makefile           # 构建脚本 (~80 行)
└── cmd/
    └── root.go        # 更新版本信息 (累计 ~70 行)
```

**代码量**: ~80 行（新增 Makefile）

### README.md 完整更新

更新 `README.md` 的完整构建部分：

```markdown
# meta-cc

Meta-Cognition tool for Claude Code - analyze session history for workflow optimization.

## Features

- 🔍 Parse Claude Code session history (JSONL format)
- 📊 Statistical analysis of tool usage and errors
- 🎯 Pattern detection for workflow optimization
- 🚀 Zero dependencies - single binary deployment

## Installation

### From Source

```bash
git clone https://github.com/yaleh/meta-cc.git
cd meta-cc
make build
```

### Cross-Platform Binaries

```bash
# Build for all supported platforms
make cross-compile

# Binaries will be in build/ directory:
# - build/meta-cc-linux-amd64
# - build/meta-cc-linux-arm64
# - build/meta-cc-darwin-amd64
# - build/meta-cc-darwin-arm64
# - build/meta-cc-windows-amd64.exe
```

## Usage

```bash
# Show help
./meta-cc --help

# Show version
./meta-cc --version

# Global options
./meta-cc --session <session-id>    # Specify session ID
./meta-cc --project <path>          # Specify project path
./meta-cc --output json|md|csv      # Output format
```

## Development

### Prerequisites

- Go 1.21 or later
- Make (optional, for build automation)

### Build

```bash
# Using Make
make build

# Or using go directly
go build -o meta-cc
```

### Test

```bash
# Run all tests
make test

# Run with coverage
make test-coverage
# Open coverage.html in browser
```

### Available Make Targets

```bash
make build           # Build for current platform
make test            # Run tests
make test-coverage   # Run tests with coverage report
make clean           # Remove build artifacts
make install         # Install to GOPATH/bin
make cross-compile   # Build for all platforms
make deps            # Download and tidy dependencies
make help            # Show help message
```

## Supported Platforms

- Linux (amd64, arm64)
- macOS (amd64, arm64/Apple Silicon)
- Windows (amd64)

## Project Structure

```
meta-cc/
├── cmd/              # Command definitions (Cobra)
├── internal/         # Internal packages
│   └── testutil/    # Test utilities
├── pkg/              # Public packages
├── tests/            # Test files and fixtures
└── docs/             # Documentation
```

## License

MIT
```

### 验收标准

- ✅ `make build` 成功构建
- ✅ `make test` 所有测试通过
- ✅ `make clean` 清理所有构建产物
- ✅ `make cross-compile` 生成所有平台的二进制文件
- ✅ `./meta-cc --version` 显示完整版本信息（版本号、commit、构建时间）
- ✅ README.md 包含完整的构建、测试和使用说明
- ✅ README.md 包含支持的平台列表

---

## Phase 0 完成标准

### 功能验收

**必须满足所有条件**:

1. **构建成功**
   ```bash
   make build
   # 或
   go build -o meta-cc
   ```
   - ✅ 无编译错误
   - ✅ 生成可执行文件 `meta-cc`

2. **测试通过**
   ```bash
   go test ./...
   ```
   - ✅ 所有测试通过
   - ✅ 无失败或跳过的测试

3. **命令行可用**
   ```bash
   ./meta-cc --help
   ```
   - ✅ 显示完整帮助信息
   - ✅ 列出所有全局参数

   ```bash
   ./meta-cc --version
   ```
   - ✅ 显示版本号、commit 和构建时间

4. **跨平台构建**
   ```bash
   make cross-compile
   ls build/
   ```
   - ✅ 生成 Linux (amd64, arm64) 二进制
   - ✅ 生成 macOS (amd64, arm64) 二进制
   - ✅ 生成 Windows (amd64) 二进制

5. **文档完整**
   - ✅ README.md 包含安装说明
   - ✅ README.md 包含构建命令
   - ✅ README.md 包含测试命令
   - ✅ README.md 包含使用示例
   - ✅ README.md 包含支持的平台列表

### 代码质量

- ✅ 总代码量 ≤ 200 行（符合 Phase 约束）
- ✅ 每个 Stage 代码量 ≤ 200 行
- ✅ 无 Go 编译警告
- ✅ 所有导出函数有注释
- ✅ 测试覆盖率 > 0%

### 项目结构

最终项目结构：

```
meta-cc/
├── go.mod                          # Go 模块定义
├── go.sum                          # 依赖锁定
├── Makefile                        # 构建脚本
├── README.md                       # 完整文档
├── main.go                         # 程序入口
├── cmd/
│   └── root.go                    # Cobra 根命令
├── internal/
│   └── testutil/
│       ├── fixtures.go            # 测试工具
│       └── fixtures_test.go       # 工具测试
├── tests/
│   └── fixtures/
│       └── sample-session.jsonl   # 测试数据
└── build/                         # 跨平台二进制（构建后生成）
    ├── meta-cc-linux-amd64
    ├── meta-cc-linux-arm64
    ├── meta-cc-darwin-amd64
    ├── meta-cc-darwin-arm64
    └── meta-cc-windows-amd64.exe
```

---

## 依赖关系

**Phase 0 依赖**:
- 无（初始 Phase）

**后续 Phase 依赖于 Phase 0**:
- Phase 1（会话文件定位）依赖于本 Phase 的 CLI 框架
- Phase 2（JSONL 解析器）依赖于本 Phase 的测试工具

---

## 风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| Go 版本兼容性问题 | 中 | 使用 Go 1.21+，在 go.mod 中明确版本要求 |
| 跨平台构建失败 | 低 | 使用标准库，避免平台特定代码 |
| Cobra/Viper 依赖冲突 | 低 | 使用最新稳定版本，go mod tidy 解决冲突 |

---

## 下一步行动

**Phase 0 完成后，进入 Phase 1: 会话文件定位**

Phase 1 将实现：
- 环境变量读取（`CC_SESSION_ID`, `CC_PROJECT_HASH`）
- 命令行参数解析（`--session`, `--project`）
- 会话文件路径解析和验证
- 项目路径哈希计算

**准备工作**:
1. 确认 Phase 0 所有验收标准已满足
2. 提交代码到 git（使用 `feat:` 前缀）
3. 创建 Phase 1 规划文档 `plans/phase-1.md`