# Build Quality Gates - 使用指南

本指南介绍如何使用 meta-cc 的构建质量门控系统，这是通过 BAIME 方法优化的结果。

## 🎯 目标

- **减少 CI 失败率**: 从 40% 降至 < 10%
- **加快错误发现**: 从 5-10 分钟降至 < 60 秒
- **减少迭代次数**: 从 3-4 次降至 < 1.5 次
- **改善开发体验**: 从 😫 提升到 😊

## 📊 检查层级

```
┌─────────────────────────────────────┐
│   make ci (CI 级别)                 │
│   - P0 + P1 完整检查                │
│   - 完整测试套件                    │
│   - 用于 GitHub Actions             │
└─────────────────────────────────────┘
           ↑
┌─────────────────────────────────────┐
│   make all (本地完整验证)           │
│   - P0 检查 + 快速测试              │
│   - 用于重要提交前                  │
└─────────────────────────────────────┘
           ↑
┌─────────────────────────────────────┐
│   make pre-commit (提交前检查)      │
│   - P0 关键检查                     │
│   - 快速测试 (-short)               │
│   - **推荐提交前运行**              │
└─────────────────────────────────────┘
           ↑
┌─────────────────────────────────────┐
│   make dev (开发迭代)               │
│   - 格式化 + 构建                   │
│   - 最快，用于日常开发              │
└─────────────────────────────────────┘
```

## 🚀 快速开始

### 日常开发流程

```bash
# 1. 修改代码
vim cmd/root.go

# 2. 快速构建测试
make dev

# 3. 提交前完整检查
make pre-commit

# 4. 提交 (如果通过)
git add .
git commit -m "feat: add new feature"
```

### 提交前 checklist

运行 `make pre-commit` 会自动执行：

- ✅ [1/6] 检查临时文件 (~0.5s)
- ✅ [2/6] 验证 test fixtures (~0.3s)
- ✅ [3/6] 检查依赖完整性 (~2.5s)
- ✅ [4/6] 验证 import 格式 (~0.2s)
- ✅ [5/6] 运行 linting (~5s)
- ✅ [6/6] 运行快速测试 (~10s)

**总耗时**: ~20 秒

## 🔍 P0 检查详解

### 1. 临时文件检查 (`check-temp-files`)

**目的**: 防止提交调试脚本和临时文件

**检查内容**:
- 根目录的 .go 文件 (除了 main.go)
- test_*.go, debug_*.go, tmp_*.go 等模式
- 编辑器临时文件 (*~, *.swp)
- 未 gitignore 的二进制文件

**错误示例**:
```bash
❌ ERROR: Temporary test/debug scripts found:
  - ./test_parser.go
  - ./debug_analyzer.go

Action: Delete these temporary files before committing
```

**修复方法**:
```bash
# 删除临时文件
rm test_*.go debug_*.go

# 或移动到合适的包
mv test_parser.go internal/parser/parser_integration_test.go
```

### 2. Fixture 完整性检查 (`check-fixtures`)

**目的**: 确保测试引用的 fixture 文件存在

**检查内容**:
- 扫描所有 `*_test.go` 文件
- 查找 `LoadFixture()` 调用
- 验证 fixture 文件存在于 `tests/fixtures/`

**错误示例**:
```bash
❌ Missing: sample-session.jsonl
   Referenced in:
     cmd/parse_test.go:24
     cmd/analyze_test.go:64
```

**修复方法**:
```bash
# 方法 1: 创建缺失的 fixture
mkdir -p tests/fixtures
echo '{"test":"data"}' > tests/fixtures/sample-session.jsonl

# 方法 2: 使用动态 fixture
testutil.TempSessionFile(t, []parser.SessionEntry{...})

# 方法 3: 删除引用该 fixture 的测试代码
```

### 3. 依赖完整性检查 (`check-deps`)

**目的**: 确保 go.mod 和 go.sum 同步且有效

**检查内容**:
- go.mod 和 go.sum 存在性
- 运行 `go mod verify` (验证 checksums)
- 运行 `go mod tidy` 并检查 go.sum 是否变化
- 检测未使用的依赖

**错误示例**:
```bash
❌ ERROR: go.sum is out of sync

Action required:
  1. Run: go mod tidy
  2. Review changes: git diff go.sum
  3. Commit updated go.sum
```

**修复方法**:
```bash
# 修复依赖问题
go mod tidy
go mod verify

# 查看变更
git diff go.sum

# 提交修复
git add go.sum
git commit -m "chore: update go.sum"
```

### 4. Import 格式检查 (`check-imports`)

**目的**: 确保 import 语句格式正确，无未使用的 import

**检查内容**:
- 运行 `goimports -l` 检查格式
- 检测未使用的 import

**错误示例**:
```bash
❌ ERROR: Files with incorrect imports:
  - cmd/root.go
  - internal/analyzer/patterns.go

Run 'make fix-imports' to auto-fix
```

**修复方法**:
```bash
# 自动修复 (推荐)
make fix-imports

# 或手动运行
goimports -w .

# 检查修复结果
git diff
```

## 🛠️ Make 目标参考

### 开发目标

| 目标 | 用途 | 耗时 | 何时使用 |
|-----|------|------|---------|
| `make dev` | 快速构建 | ~5s | 日常开发迭代 |
| `make build` | 完整构建 | ~10s | 验证构建成功 |
| `make test` | 快速测试 | ~15s | 测试单个更改 |
| `make fmt` | 代码格式化 | ~2s | 提交前清理 |

### 质量门控目标

| 目标 | 用途 | 耗时 | 何时使用 |
|-----|------|------|---------|
| `make check-workspace` | P0 检查 | ~3s | 提交前验证工作区 |
| `make check-temp-files` | 临时文件 | ~0.5s | 调试脚本清理后 |
| `make check-fixtures` | Fixture | ~0.3s | 添加测试后 |
| `make check-deps` | 依赖 | ~2.5s | 更新依赖后 |
| `make check-imports` | Import | ~0.2s | 添加 import 后 |

### 综合目标

| 目标 | 用途 | 耗时 | 何时使用 |
|-----|------|------|---------|
| `make pre-commit` | 提交前检查 | ~20s | **每次提交前** |
| `make all` | 完整验证 | ~30s | 重要提交前 |
| `make ci` | CI 级别 | ~60s | 模拟 CI 环境 |

## 🔧 常见问题

### Q1: `make pre-commit` 失败，如何快速修复？

**A**: 按照错误信息的建议操作：

```bash
# 1. 查看完整错误
make pre-commit 2>&1 | less

# 2. 针对性修复
make fix-imports           # 如果是 import 错误
rm test_*.go               # 如果是临时文件
go mod tidy                # 如果是依赖问题

# 3. 重新检查
make pre-commit
```

### Q2: 如何跳过某个检查？

**A**: 不推荐跳过，但可以单独运行：

```bash
# 跳过 lint (不推荐)
make check-workspace check-imports test build

# 或临时修改 Makefile
```

### Q3: 检查脚本误报怎么办？

**A**: 可以自定义排除规则：

```bash
# 例如排除特定文件
# 编辑 scripts/check-temp-files.sh
TEMP_SCRIPTS=$(find . -type f \( \
    -name "test_*.go" \
\) ! -path "*/my_special_test.go" ...)
```

### Q4: 如何在 CI 中使用？

**A**: 在 `.github/workflows/ci.yml` 中：

```yaml
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'

      - name: Run quality gates
        run: make ci
```

### Q5: 本地和 CI 检查不一致？

**A**: 检查工具版本：

```bash
# 检查 golangci-lint 版本
golangci-lint version

# 安装指定版本 (与 CI 保持一致)
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v2.12.2

# 或使用 asdf
asdf install golangci-lint 2.12.2
asdf local golangci-lint 2.12.2
```

## 📈 性能优化技巧

### 1. 使用缓存

```bash
# Go test 自动缓存
go test -short ./...  # 使用缓存

# 清除缓存 (如需)
go clean -testcache
```

### 2. 并行运行独立检查

```makefile
# Makefile 中使用 & 并行
check-parallel:
	@make check-temp-files & \
	make check-fixtures & \
	make check-imports & \
	wait
```

### 3. 跳过慢速测试

```bash
# 开发时使用 -short
go test -short ./...

# CI 时运行完整测试
go test ./...
```

## 🎓 最佳实践

### 1. 提交前总是运行 `make pre-commit`

```bash
# 设置 git alias
git config alias.pc '!make pre-commit && git commit'

# 使用
git add .
git pc -m "feat: add new feature"
```

### 2. 使用 pre-commit hooks (可选)

```bash
# 安装 pre-commit 框架
pip install pre-commit

# 配置 .pre-commit-config.yaml
# (参考 .pre-commit-config.yaml.example)

# 安装 hooks
pre-commit install
```

### 3. 定期更新依赖

```bash
# 每周运行
go get -u ./...
go mod tidy
make all  # 验证更新
```

### 4. 保持工作区清洁

```bash
# 定期清理
make clean
git clean -fdx  # 删除所有未跟踪文件 (谨慎)
```

## 📚 相关文档

- [BAIME 实验文档](../experiments/build-quality-gates/)
- [Testing Strategy Skill](../../.claude/skills/testing-strategy/)
- [CI/CD Optimization Skill](../../.claude/skills/ci-cd-optimization/)

## 🤝 贡献

发现问题或有改进建议？

1. 查看实验文档了解设计原理
2. 提 Issue 或 PR
3. 或直接修改检查脚本 (scripts/check-*.sh)

---

**最后更新**: 2025-10-20
**维护者**: meta-cc team
