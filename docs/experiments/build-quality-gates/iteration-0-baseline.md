# Iteration 0: Baseline - Build Quality Gates

**日期**: 2025-10-20
**目标**: 建立 baseline 指标并设计实验框架

## 📊 Baseline 指标 (历史数据分析)

### 1. CI 失败率

**数据源**: 最近 10 次 CI runs (main 分支)

```
❌ fix(ci): use golangci-lint-action v6 for v1 compatibility - success
❌ fix(ci): revert to golangci-lint v1.64.8 for GitHub Actions - failure
❌ fix(ci): upgrade golangci-lint-action to v8 - failure
❌ fix(ci): upgrade golangci-lint to v2.5.0 - failure
❌ fix(mcp): set default scope to 'session' for get_session_stats tool - failure
```

**失败率**: 40% (4/10)

### 2. 错误类型分布 (50个历史Bash错误样本)

| 错误类型 | 数量 | 占比 | 可在make中检测 |
|---------|------|------|---------------|
| 临时文件污染 (test_*.go) | 14 | 28% | ✅ YES |
| 未使用 Import | 5 | 10% | ✅ YES |
| Test Fixture 缺失 | 4 | 8% | ✅ YES |
| JQ/Shell 脚本错误 | 15 | 30% | ✅ YES |
| JSON 断言错误 | 5 | 10% | ⚠️  PARTIAL |
| 类型错误 | 4 | 8% | ✅ YES (lint) |
| 其他 | 3 | 6% | - |

**可提前检测**: 80% (40/50)

### 3. 错误修复时间

**平均迭代周期**:
1. 本地修改: 5-15 分钟
2. Commit + Push: 1 分钟
3. CI 运行: 5-8 分钟
4. 发现错误: 立即
5. **总计**: 11-24 分钟/迭代

**平均迭代次数**: 3-4 次
**总修复时间**: 33-96 分钟 (平均 ~60 分钟)

### 4. 当前 Makefile 覆盖率

**现有检查**:
- ✅ `make fmt`: gofmt 格式化
- ✅ `make vet`: go vet 静态分析
- ✅ `make lint`: golangci-lint (但配置宽松)
- ✅ `make test`: 单元测试
- ⚠️  `make lint-errors`: 自定义错误检查 (部分)

**缺失检查**:
- ❌ 临时文件检测
- ❌ Fixture 完整性验证
- ❌ Import 格式化验证
- ❌ Shell 脚本语法检查
- ❌ 插件文件一致性
- ❌ 依赖完整性 (go.mod/go.sum)

**覆盖率**: ~30% (只覆盖基础 lint 和测试)

## 🎯 实验设计

### Iteration 1: P0 Critical Checks

**目标**: 捕获 80% 的历史错误

**实施内容**:
1. ✅ `check-temp-files`: 临时文件检测
2. ✅ `check-fixtures`: Fixture 完整性
3. ✅ `check-imports`: Import 格式化
4. ✅ `check-deps`: go.mod/go.sum 验证
5. ✅ `check-plugin-version`: 版本一致性

**预期效果**:
- CI 失败率: 40% → 15%
- 平均迭代: 3-4 → 2
- 检测时间: 5-10分钟 → 30秒

### Iteration 2: P1 Important Checks

**目标**: 捕获剩余 10% 错误 + 提升代码质量

**实施内容**:
1. ✅ `check-debug`: Debug 语句检测
2. ✅ `check-scripts`: Shell 脚本验证
3. ✅ `check-plugin-files`: 插件文件存在性
4. ✅ `check-module-path`: 模块路径一致性
5. ✅ `test-scripts`: Bats 测试执行

**预期效果**:
- CI 失败率: 15% → 5%
- 平均迭代: 2 → 1.5
- 开发者体验: 😫 → 😊

### Iteration 3: Performance & Integration

**目标**: 优化性能 + CI 集成

**实施内容**:
1. ✅ 并行化检查 (Make 并发执行)
2. ✅ 缓存机制 (跳过未变更文件)
3. ✅ Pre-commit hooks 集成
4. ✅ CI workflow 优化

**预期效果**:
- 检查时间: 30-60秒 → 15-30秒
- CI 失败率: 5% → 2%
- 完全自动化

## 📐 评估指标

### V_instance (实例质量)

**公式**:
```
V_instance = 0.4 × (1 - CI_failure_rate)
           + 0.3 × (1 - avg_iterations/baseline_iterations)
           + 0.2 × (baseline_detection_time/actual_detection_time)
           + 0.1 × error_coverage_rate
```

**Baseline**:
```
V_instance(baseline) = 0.4 × (1 - 0.40)
                     + 0.3 × (1 - 3.5/3.5)
                     + 0.2 × (480s/480s)
                     + 0.1 × 0.30
= 0.4 × 0.6 + 0.3 × 0 + 0.2 × 1.0 + 0.1 × 0.3
= 0.24 + 0 + 0.20 + 0.03
= 0.47
```

### V_meta (方法论质量)

**公式**:
```
V_meta = 0.3 × transferability
       + 0.25 × automation_level
       + 0.25 × documentation_quality
       + 0.2 × (1 - performance_overhead/threshold)
```

**Baseline**:
```
V_meta(baseline) = 0.3 × 0.30   # 只有基础 lint
                 + 0.25 × 0.50   # 部分自动化
                 + 0.25 × 0.60   # 文档不完整
                 + 0.2 × 0.80    # 性能可接受
= 0.09 + 0.125 + 0.15 + 0.16
= 0.525
```

## 🎬 下一步

1. **创建检查脚本** (scripts/check-*.sh)
2. **更新 Makefile** (添加 check-* 目标)
3. **运行回顾式验证** (在历史错误上测试)
4. **评估 Iteration 1 效果** (V_instance, V_meta)

---

**Baseline Summary**:
- V_instance(s₀) = 0.47
- V_meta(m₀) = 0.525
- **Target**: V_instance ≥ 0.85, V_meta ≥ 0.80
- **Gap**: Δ_instance = 0.38, Δ_meta = 0.275
