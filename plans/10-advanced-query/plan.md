# Phase 10: Advanced Query Capabilities (高级查询能力)

## 概述

**目标**: 实现高级过滤、聚合统计、时间序列分析和文件级统计，为 Claude 集成层(@meta-coach, MCP Server, Slash Commands)提供更丰富的数据维度

**代码量**: ~400 行 (Go 源代码)

**依赖**: Phase 0-9 (完整的 meta-cc 工具链 + 查询基础 + 上下文管理)

**交付物**:
- 高级过滤器：布尔表达式、范围查询、集合操作、模式匹配
- 聚合函数：`stats aggregate --group-by` 命令
- 时间序列：`stats time-series` 命令
- 文件级统计：`stats files` 命令
- 更新文档和集成测试

---

## Phase 目标

扩展 Phase 8 的基础查询能力，提供高级分析功能，解决以下需求：

### 核心需求

1. **复杂过滤表达式**：支持 SQL-like 查询语法（AND/OR/NOT, IN, BETWEEN, LIKE）
2. **统计聚合**：按字段分组计算统计指标（count, avg, error_rate, percentiles）
3. **时间趋势分析**：工具使用、错误率、会话活跃度随时间变化
4. **文件维度分析**：识别高频修改文件、错误集中的文件

### 设计原则

Phase 10 遵循 meta-cc 职责边界：

- ✅ **meta-cc 职责**: 数据过滤、数学/统计计算、模式检测（基于规则）
- ❌ **不做**: 语义分析、因果推断、建议生成（留给 Claude 集成层）
- ✅ **输出**: 高密度结构化数据（JSON/Markdown/CSV）供 Claude 理解

**示例**:
```bash
# meta-cc: 计算每个工具的错误率（纯数学）
meta-cc stats aggregate --group-by tool --metrics error_rate

# Claude (@meta-coach): 语义理解和建议
# "Bash 工具错误率 15% 偏高，可能原因：路径问题、权限不足、命令语法错误。
#  建议：1) 添加错误检查 hook  2) 使用绝对路径  3) 验证命令参数"
```

---

## 成功标准

**功能验收**:
- ✅ 所有 Stage 单元测试通过（TDD）
- ✅ 高级过滤器支持 SQL-like 语法（AND, OR, NOT, IN, BETWEEN, LIKE）
- ✅ 聚合统计准确（与手动计算结果对比，误差 < 1%）
- ✅ 时间序列数据正确分桶（hour/day/week）
- ✅ 文件统计匹配实际操作计数

**集成验收**:
- ✅ 与现有 `query tools` 命令无缝集成
- ✅ @meta-coach 使用示例更新
- ✅ MCP Server 增强（可选）
- ✅ README.md 添加"高级查询功能"章节

**代码质量**:
- ✅ 实际代码量: ~400 行（目标 350-450 行）
- ✅ 每个 Stage ≤ 120 行
- ✅ 测试覆盖率: ≥ 80%
- ✅ 查询性能: < 200ms（复杂查询，1000 turns 会话）

---

## Stage 10.1: 高级过滤引擎

### 目标

实现高级过滤表达式解析和执行，支持 SQL-like 语法。

### 实现步骤

#### 1. 定义过滤表达式语法

**支持的操作符**:

```
布尔操作符: AND, OR, NOT
比较操作符: =, !=, >, <, >=, <=
集合操作符: IN, NOT IN
范围操作符: BETWEEN ... AND ...
模式匹配: LIKE (SQL wildcard), REGEXP (正则表达式)
```

**示例表达式**:
```bash
# 布尔组合
"tool='Bash' AND status='error'"
"status='error' OR duration>1000"
"NOT (status='success')"

# 集合操作
"tool IN ('Bash', 'Edit', 'Write')"
"status NOT IN ('success')"

# 范围查询
"duration BETWEEN 500 AND 2000"
"timestamp BETWEEN '2025-10-01' AND '2025-10-03'"

# 模式匹配
"tool LIKE 'meta%'"              # 以 meta 开头
"error_text REGEXP 'permission.*denied'"
```

#### 2. 实现表达式解析器

**文件**: `internal/filter/expression.go` (新建, ~60 行)

```go
package filter

import (
    "fmt"
    "regexp"
    "strings"
)

// Expression 过滤表达式接口
type Expression interface {
    Evaluate(record map[string]interface{}) (bool, error)
}

// BinaryExpression 二元表达式 (AND, OR)
type BinaryExpression struct {
    Operator string      // "AND" 或 "OR"
    Left     Expression
    Right    Expression
}

func (e *BinaryExpression) Evaluate(record map[string]interface{}) (bool, error) {
    left, err := e.Left.Evaluate(record)
    if err != nil {
        return false, err
    }

    right, err := e.Right.Evaluate(record)
    if err != nil {
        return false, err
    }

    switch e.Operator {
    case "AND":
        return left && right, nil
    case "OR":
        return left || right, nil
    default:
        return false, fmt.Errorf("unknown operator: %s", e.Operator)
    }
}

// UnaryExpression 一元表达式 (NOT)
type UnaryExpression struct {
    Operator string      // "NOT"
    Operand  Expression
}

func (e *UnaryExpression) Evaluate(record map[string]interface{}) (bool, error) {
    result, err := e.Operand.Evaluate(record)
    if err != nil {
        return false, err
    }

    return !result, nil
}

// ComparisonExpression 比较表达式 (=, !=, >, <, etc.)
type ComparisonExpression struct {
    Field    string
    Operator string      // "=", "!=", ">", "<", ">=", "<="
    Value    interface{}
}

func (e *ComparisonExpression) Evaluate(record map[string]interface{}) (bool, error) {
    fieldValue, exists := record[e.Field]
    if !exists {
        return false, nil // 字段不存在，视为不匹配
    }

    return compareValues(fieldValue, e.Operator, e.Value)
}

// InExpression 集合成员检查 (IN, NOT IN)
type InExpression struct {
    Field  string
    Values []interface{}
    Negate bool  // true 表示 NOT IN
}

func (e *InExpression) Evaluate(record map[string]interface{}) (bool, error) {
    fieldValue, exists := record[e.Field]
    if !exists {
        return false, nil
    }

    found := false
    for _, v := range e.Values {
        if valueEquals(fieldValue, v) {
            found = true
            break
        }
    }

    if e.Negate {
        return !found, nil
    }
    return found, nil
}

// BetweenExpression 范围检查 (BETWEEN ... AND ...)
type BetweenExpression struct {
    Field string
    Lower interface{}
    Upper interface{}
}

func (e *BetweenExpression) Evaluate(record map[string]interface{}) (bool, error) {
    fieldValue, exists := record[e.Field]
    if !exists {
        return false, nil
    }

    lowerOk, _ := compareValues(fieldValue, ">=", e.Lower)
    upperOk, _ := compareValues(fieldValue, "<=", e.Upper)

    return lowerOk && upperOk, nil
}

// LikeExpression 模式匹配 (LIKE)
type LikeExpression struct {
    Field   string
    Pattern string  // SQL wildcard: % (任意字符), _ (单个字符)
}

func (e *LikeExpression) Evaluate(record map[string]interface{}) (bool, error) {
    fieldValue, exists := record[e.Field]
    if !exists {
        return false, nil
    }

    str, ok := fieldValue.(string)
    if !ok {
        return false, nil
    }

    // 将 SQL wildcard 转换为正则表达式
    pattern := strings.ReplaceAll(e.Pattern, "%", ".*")
    pattern = strings.ReplaceAll(pattern, "_", ".")
    pattern = "^" + pattern + "$"

    matched, _ := regexp.MatchString(pattern, str)
    return matched, nil
}

// RegexpExpression 正则表达式匹配
type RegexpExpression struct {
    Field   string
    Pattern string
}

func (e *RegexpExpression) Evaluate(record map[string]interface{}) (bool, error) {
    fieldValue, exists := record[e.Field]
    if !exists {
        return false, nil
    }

    str, ok := fieldValue.(string)
    if !ok {
        return false, nil
    }

    matched, err := regexp.MatchString(e.Pattern, str)
    if err != nil {
        return false, fmt.Errorf("invalid regexp: %v", err)
    }

    return matched, nil
}

// 辅助函数：值比较
func compareValues(left interface{}, operator string, right interface{}) (bool, error) {
    // 类型转换和比较逻辑
    // 支持 string, int, float64, time.Time 等类型
    // 实现略...
    return false, nil
}

// 辅助函数：值相等检查
func valueEquals(left, right interface{}) bool {
    // 实现略...
    return false
}
```

#### 3. 实现表达式解析器（字符串 → Expression 树）

**文件**: `internal/filter/parser.go` (新建, ~60 行)

```go
package filter

import (
    "fmt"
    "strings"
)

// ParseExpression 解析过滤表达式字符串
func ParseExpression(expr string) (Expression, error) {
    parser := &ExpressionParser{
        input: strings.TrimSpace(expr),
        pos:   0,
    }

    return parser.parse()
}

// ExpressionParser 表达式解析器
type ExpressionParser struct {
    input string
    pos   int
}

func (p *ExpressionParser) parse() (Expression, error) {
    return p.parseOr()
}

func (p *ExpressionParser) parseOr() (Expression, error) {
    left, err := p.parseAnd()
    if err != nil {
        return nil, err
    }

    for p.match(" OR ") {
        right, err := p.parseAnd()
        if err != nil {
            return nil, err
        }
        left = &BinaryExpression{
            Operator: "OR",
            Left:     left,
            Right:    right,
        }
    }

    return left, nil
}

func (p *ExpressionParser) parseAnd() (Expression, error) {
    left, err := p.parseUnary()
    if err != nil {
        return nil, err
    }

    for p.match(" AND ") {
        right, err := p.parseUnary()
        if err != nil {
            return nil, err
        }
        left = &BinaryExpression{
            Operator: "AND",
            Left:     left,
            Right:    right,
        }
    }

    return left, nil
}

func (p *ExpressionParser) parseUnary() (Expression, error) {
    if p.match("NOT ") {
        operand, err := p.parsePrimary()
        if err != nil {
            return nil, err
        }
        return &UnaryExpression{
            Operator: "NOT",
            Operand:  operand,
        }, nil
    }

    return p.parsePrimary()
}

func (p *ExpressionParser) parsePrimary() (Expression, error) {
    // 解析括号表达式
    if p.match("(") {
        expr, err := p.parse()
        if err != nil {
            return nil, err
        }
        if !p.match(")") {
            return nil, fmt.Errorf("missing closing parenthesis")
        }
        return expr, nil
    }

    // 解析基础表达式（字段操作符值）
    field, err := p.parseIdentifier()
    if err != nil {
        return nil, err
    }

    // 检查操作符
    if p.match(" IN ") {
        values, err := p.parseValueList()
        if err != nil {
            return nil, err
        }
        return &InExpression{
            Field:  field,
            Values: values,
            Negate: false,
        }, nil
    }

    if p.match(" NOT IN ") {
        values, err := p.parseValueList()
        if err != nil {
            return nil, err
        }
        return &InExpression{
            Field:  field,
            Values: values,
            Negate: true,
        }, nil
    }

    if p.match(" BETWEEN ") {
        lower, err := p.parseValue()
        if err != nil {
            return nil, err
        }
        if !p.match(" AND ") {
            return nil, fmt.Errorf("BETWEEN requires AND")
        }
        upper, err := p.parseValue()
        if err != nil {
            return nil, err
        }
        return &BetweenExpression{
            Field: field,
            Lower: lower,
            Upper: upper,
        }, nil
    }

    if p.match(" LIKE ") {
        pattern, err := p.parseValue()
        if err != nil {
            return nil, err
        }
        return &LikeExpression{
            Field:   field,
            Pattern: pattern.(string),
        }, nil
    }

    if p.match(" REGEXP ") {
        pattern, err := p.parseValue()
        if err != nil {
            return nil, err
        }
        return &RegexpExpression{
            Field:   field,
            Pattern: pattern.(string),
        }, nil
    }

    // 比较操作符
    operator, err := p.parseOperator()
    if err != nil {
        return nil, err
    }

    value, err := p.parseValue()
    if err != nil {
        return nil, err
    }

    return &ComparisonExpression{
        Field:    field,
        Operator: operator,
        Value:    value,
    }, nil
}

func (p *ExpressionParser) match(token string) bool {
    if strings.HasPrefix(p.input[p.pos:], token) {
        p.pos += len(token)
        return true
    }
    return false
}

func (p *ExpressionParser) parseIdentifier() (string, error) {
    // 解析字段名（字母、数字、下划线）
    // 实现略...
    return "", nil
}

func (p *ExpressionParser) parseOperator() (string, error) {
    // 解析操作符: =, !=, >, <, >=, <=
    // 实现略...
    return "", nil
}

func (p *ExpressionParser) parseValue() (interface{}, error) {
    // 解析值: 字符串（带引号）、数字、布尔值
    // 实现略...
    return nil, nil
}

func (p *ExpressionParser) parseValueList() ([]interface{}, error) {
    // 解析值列表: ('value1', 'value2', 'value3')
    // 实现略...
    return nil, nil
}
```

#### 4. 集成到查询命令

**文件**: `cmd/query_tools.go` (修改 ~20 行)

```go
// 在 query tools 命令中添加 --where 标志支持

var whereFlag string

func init() {
    queryToolsCmd.Flags().StringVar(&whereFlag, "where", "", "Advanced filter expression (e.g., \"tool='Bash' AND status='error'\")")
}

func runQueryTools(cmd *cobra.Command, args []string) error {
    // ... 现有逻辑（解析、过滤）...

    // 如果提供了 --where 表达式，使用高级过滤器
    if whereFlag != "" {
        expr, err := filter.ParseExpression(whereFlag)
        if err != nil {
            return fmt.Errorf("invalid where expression: %v", err)
        }

        tools = filter.ApplyAdvancedFilter(tools, expr)
    }

    // ... 剩余逻辑（输出）...
}
```

**文件**: `internal/filter/advanced.go` (新建, ~20 行)

```go
package filter

import "github.com/yaleh/meta-cc/internal/parser"

// ApplyAdvancedFilter 应用高级过滤表达式
func ApplyAdvancedFilter(tools []parser.ToolCall, expr Expression) []parser.ToolCall {
    var filtered []parser.ToolCall

    for _, tool := range tools {
        // 将 ToolCall 转换为 map（用于表达式求值）
        record := toolCallToMap(tool)

        // 求值表达式
        match, err := expr.Evaluate(record)
        if err != nil {
            // 表达式求值错误，跳过该记录
            continue
        }

        if match {
            filtered = append(filtered, tool)
        }
    }

    return filtered
}

func toolCallToMap(tool parser.ToolCall) map[string]interface{} {
    // 转换为 map（类似 Phase 9 的投影逻辑）
    return map[string]interface{}{
        "uuid":      tool.UUID,
        "timestamp": tool.Timestamp,
        "tool":      tool.Tool,
        "status":    tool.Status,
        "duration":  tool.Duration,
        "error":     tool.Error,
        // ... 其他字段
    }
}
```

### TDD 步骤

**测试文件**: `internal/filter/expression_test.go` (~80 行)

```go
package filter

import "testing"

func TestComparisonExpression(t *testing.T) {
    tests := []struct {
        name       string
        expression Expression
        record     map[string]interface{}
        expected   bool
    }{
        {
            name: "equal string",
            expression: &ComparisonExpression{
                Field:    "tool",
                Operator: "=",
                Value:    "Bash",
            },
            record:   map[string]interface{}{"tool": "Bash"},
            expected: true,
        },
        {
            name: "greater than number",
            expression: &ComparisonExpression{
                Field:    "duration",
                Operator: ">",
                Value:    1000,
            },
            record:   map[string]interface{}{"duration": 1500},
            expected: true,
        },
        // 更多测试用例...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := tt.expression.Evaluate(tt.record)
            if err != nil {
                t.Fatalf("unexpected error: %v", err)
            }
            if result != tt.expected {
                t.Errorf("expected %v, got %v", tt.expected, result)
            }
        })
    }
}

func TestBinaryExpression(t *testing.T) {
    // 测试 AND, OR 逻辑
    left := &ComparisonExpression{Field: "status", Operator: "=", Value: "error"}
    right := &ComparisonExpression{Field: "tool", Operator: "=", Value: "Bash"}

    andExpr := &BinaryExpression{Operator: "AND", Left: left, Right: right}
    orExpr := &BinaryExpression{Operator: "OR", Left: left, Right: right}

    record1 := map[string]interface{}{"status": "error", "tool": "Bash"}
    record2 := map[string]interface{}{"status": "error", "tool": "Edit"}

    // AND: 两个都为 true
    result1, _ := andExpr.Evaluate(record1)
    if !result1 {
        t.Error("AND expression failed for matching record")
    }

    // AND: 一个为 false
    result2, _ := andExpr.Evaluate(record2)
    if result2 {
        t.Error("AND expression should be false")
    }

    // OR: 一个为 true
    result3, _ := orExpr.Evaluate(record2)
    if !result3 {
        t.Error("OR expression should be true")
    }
}

func TestInExpression(t *testing.T) {
    expr := &InExpression{
        Field:  "tool",
        Values: []interface{}{"Bash", "Edit", "Write"},
        Negate: false,
    }

    record1 := map[string]interface{}{"tool": "Bash"}
    record2 := map[string]interface{}{"tool": "Read"}

    result1, _ := expr.Evaluate(record1)
    if !result1 {
        t.Error("IN expression should match Bash")
    }

    result2, _ := expr.Evaluate(record2)
    if result2 {
        t.Error("IN expression should not match Read")
    }
}

func TestBetweenExpression(t *testing.T) {
    expr := &BetweenExpression{
        Field: "duration",
        Lower: 500,
        Upper: 2000,
    }

    tests := []struct {
        value    int
        expected bool
    }{
        {300, false},   // < lower
        {500, true},    // = lower
        {1000, true},   // in range
        {2000, true},   // = upper
        {2500, false},  // > upper
    }

    for _, tt := range tests {
        record := map[string]interface{}{"duration": tt.value}
        result, _ := expr.Evaluate(record)
        if result != tt.expected {
            t.Errorf("BETWEEN for %d: expected %v, got %v", tt.value, tt.expected, result)
        }
    }
}

func TestLikeExpression(t *testing.T) {
    tests := []struct {
        pattern  string
        value    string
        expected bool
    }{
        {"meta%", "meta-cc", true},      // 前缀匹配
        {"meta%", "my-meta", false},
        {"%coach", "meta-coach", true},  // 后缀匹配
        {"%coach", "coach-meta", false},
        {"%meta%", "my-meta-tool", true},// 包含匹配
        {"Bash", "Bash", true},          // 精确匹配
    }

    for _, tt := range tests {
        expr := &LikeExpression{Field: "tool", Pattern: tt.pattern}
        record := map[string]interface{}{"tool": tt.value}
        result, _ := expr.Evaluate(record)
        if result != tt.expected {
            t.Errorf("LIKE '%s' for '%s': expected %v, got %v", tt.pattern, tt.value, tt.expected, result)
        }
    }
}
```

**测试文件**: `internal/filter/parser_test.go` (~50 行)

```go
package filter

import "testing"

func TestParseExpression(t *testing.T) {
    tests := []struct {
        input       string
        expectError bool
    }{
        // 有效表达式
        {"tool='Bash'", false},
        {"status='error' AND tool='Bash'", false},
        {"status='error' OR duration>1000", false},
        {"NOT (status='success')", false},
        {"tool IN ('Bash', 'Edit')", false},
        {"duration BETWEEN 500 AND 2000", false},
        {"tool LIKE 'meta%'", false},
        {"error REGEXP 'permission.*denied'", false},

        // 复杂组合
        {"(status='error' OR duration>1000) AND tool='Bash'", false},

        // 无效表达式
        {"tool=", true},             // 缺少值
        {"AND tool='Bash'", true},   // 缺少左操作数
        {"tool IN ()", true},        // 空列表
        {"BETWEEN 100", true},       // 不完整
    }

    for _, tt := range tests {
        t.Run(tt.input, func(t *testing.T) {
            _, err := ParseExpression(tt.input)
            if tt.expectError && err == nil {
                t.Error("expected error but got none")
            }
            if !tt.expectError && err != nil {
                t.Errorf("unexpected error: %v", err)
            }
        })
    }
}

func TestParsedExpressionEvaluation(t *testing.T) {
    // 测试解析后的表达式求值是否正确
    expr, err := ParseExpression("status='error' AND duration>1000")
    if err != nil {
        t.Fatalf("parse failed: %v", err)
    }

    record1 := map[string]interface{}{
        "status":   "error",
        "duration": 1500,
    }
    result1, _ := expr.Evaluate(record1)
    if !result1 {
        t.Error("expression should match record1")
    }

    record2 := map[string]interface{}{
        "status":   "error",
        "duration": 500,
    }
    result2, _ := expr.Evaluate(record2)
    if result2 {
        t.Error("expression should not match record2")
    }
}
```

### 交付物

**新建文件**:
- `internal/filter/expression.go` (~60 行)
- `internal/filter/parser.go` (~60 行)
- `internal/filter/advanced.go` (~20 行)
- `internal/filter/expression_test.go` (~80 行)
- `internal/filter/parser_test.go` (~50 行)

**修改文件**:
- `cmd/query_tools.go` (~20 行)

**代码量**: ~120 行（新增 Go 源代码，不含测试）

### 验收标准

- ✅ `meta-cc query tools --where "tool='Bash' AND status='error'"` 返回匹配记录
- ✅ `meta-cc query tools --where "duration BETWEEN 500 AND 2000"` 范围查询正确
- ✅ `meta-cc query tools --where "tool IN ('Bash', 'Edit')"` 集合查询正确
- ✅ `meta-cc query tools --where "tool LIKE 'meta%'"` 模式匹配正确
- ✅ 所有单元测试通过
- ✅ 查询性能 < 200ms（1000 turns 会话）

---

## Stage 10.2: 聚合函数

### 目标

实现 `stats aggregate` 命令，支持按字段分组计算统计指标。

### 实现步骤

#### 1. 定义聚合指标

**支持的指标**:
- `count`: 记录数量
- `sum`: 数值字段求和
- `avg`: 平均值
- `min`: 最小值
- `max`: 最大值
- `error_rate`: 错误率（错误数 / 总数）
- `p50`, `p90`, `p95`, `p99`: 百分位数

#### 2. 实现聚合逻辑

**文件**: `internal/analyzer/aggregate.go` (新建, ~60 行)

```go
package analyzer

import (
    "fmt"
    "sort"
    "github.com/yaleh/meta-cc/internal/parser"
)

// AggregateConfig 聚合配置
type AggregateConfig struct {
    GroupBy string   // 分组字段（如 "tool", "status"）
    Metrics []string // 要计算的指标列表
}

// AggregateResult 聚合结果
type AggregateResult struct {
    GroupValue string                 `json:"group_value"`
    Metrics    map[string]interface{} `json:"metrics"`
}

// Aggregate 对 ToolCall 数据进行聚合
func Aggregate(tools []parser.ToolCall, config AggregateConfig) ([]AggregateResult, error) {
    // Step 1: 按 GroupBy 字段分组
    groups := make(map[string][]parser.ToolCall)

    for _, tool := range tools {
        groupValue := getFieldValue(tool, config.GroupBy)
        groups[groupValue] = append(groups[groupValue], tool)
    }

    // Step 2: 对每个分组计算指标
    var results []AggregateResult

    for groupValue, groupTools := range groups {
        metrics := make(map[string]interface{})

        for _, metric := range config.Metrics {
            value, err := calculateMetric(groupTools, metric)
            if err != nil {
                return nil, err
            }
            metrics[metric] = value
        }

        results = append(results, AggregateResult{
            GroupValue: groupValue,
            Metrics:    metrics,
        })
    }

    // Step 3: 按 count 降序排序
    sort.Slice(results, func(i, j int) bool {
        ci, _ := results[i].Metrics["count"].(int)
        cj, _ := results[j].Metrics["count"].(int)
        return ci > cj
    })

    return results, nil
}

// calculateMetric 计算单个指标
func calculateMetric(tools []parser.ToolCall, metric string) (interface{}, error) {
    switch metric {
    case "count":
        return len(tools), nil

    case "error_rate":
        errorCount := 0
        for _, tool := range tools {
            if tool.Status == "error" {
                errorCount++
            }
        }
        return float64(errorCount) / float64(len(tools)), nil

    case "avg_duration":
        sum := 0
        for _, tool := range tools {
            sum += tool.Duration
        }
        return float64(sum) / float64(len(tools)), nil

    case "sum_duration":
        sum := 0
        for _, tool := range tools {
            sum += tool.Duration
        }
        return sum, nil

    case "min_duration":
        if len(tools) == 0 {
            return 0, nil
        }
        min := tools[0].Duration
        for _, tool := range tools {
            if tool.Duration < min {
                min = tool.Duration
            }
        }
        return min, nil

    case "max_duration":
        if len(tools) == 0 {
            return 0, nil
        }
        max := tools[0].Duration
        for _, tool := range tools {
            if tool.Duration > max {
                max = tool.Duration
            }
        }
        return max, nil

    case "p50", "p90", "p95", "p99":
        return calculatePercentile(tools, metric)

    default:
        return nil, fmt.Errorf("unknown metric: %s", metric)
    }
}

// calculatePercentile 计算百分位数
func calculatePercentile(tools []parser.ToolCall, metric string) (int, error) {
    if len(tools) == 0 {
        return 0, nil
    }

    // 提取 duration 并排序
    durations := make([]int, len(tools))
    for i, tool := range tools {
        durations[i] = tool.Duration
    }
    sort.Ints(durations)

    // 计算百分位索引
    var percentile float64
    switch metric {
    case "p50":
        percentile = 0.50
    case "p90":
        percentile = 0.90
    case "p95":
        percentile = 0.95
    case "p99":
        percentile = 0.99
    }

    index := int(float64(len(durations)-1) * percentile)
    return durations[index], nil
}

// getFieldValue 从 ToolCall 中提取字段值
func getFieldValue(tool parser.ToolCall, field string) string {
    switch field {
    case "tool":
        return tool.Tool
    case "status":
        return tool.Status
    // 更多字段...
    default:
        return ""
    }
}
```

#### 3. 实现 stats aggregate 命令

**文件**: `cmd/stats.go` (新建, ~40 行)

```go
package cmd

import (
    "encoding/json"
    "fmt"
    "github.com/spf13/cobra"
    "github.com/yaleh/meta-cc/internal/analyzer"
    "github.com/yaleh/meta-cc/internal/locator"
    "github.com/yaleh/meta-cc/internal/parser"
)

var (
    groupByFlag string
    metricsFlag string
)

var statsAggregateCmd = &cobra.Command{
    Use:   "aggregate",
    Short: "Aggregate statistics by field",
    Long: `Aggregate tool call statistics grouped by a field.

Examples:
  meta-cc stats aggregate --group-by tool --metrics "count,error_rate,avg_duration"
  meta-cc stats aggregate --group-by status --metrics "count,percentage"`,
    RunE: runStatsAggregate,
}

func init() {
    statsCmd.AddCommand(statsAggregateCmd)

    statsAggregateCmd.Flags().StringVar(&groupByFlag, "group-by", "tool", "Field to group by")
    statsAggregateCmd.Flags().StringVar(&metricsFlag, "metrics", "count,error_rate", "Comma-separated metrics to calculate")
}

func runStatsAggregate(cmd *cobra.Command, args []string) error {
    // 定位会话文件
    sessionPath, err := locator.LocateSession(sessionFlag, projectFlag)
    if err != nil {
        return err
    }

    // 解析会话
    tools, err := parser.ExtractToolCalls(sessionPath)
    if err != nil {
        return err
    }

    // 解析指标列表
    metrics := strings.Split(metricsFlag, ",")
    for i := range metrics {
        metrics[i] = strings.TrimSpace(metrics[i])
    }

    // 执行聚合
    config := analyzer.AggregateConfig{
        GroupBy: groupByFlag,
        Metrics: metrics,
    }

    results, err := analyzer.Aggregate(tools, config)
    if err != nil {
        return err
    }

    // 输出结果
    data, _ := json.MarshalIndent(results, "", "  ")
    fmt.Println(string(data))

    return nil
}
```

### TDD 步骤

**测试文件**: `internal/analyzer/aggregate_test.go` (~60 行)

```go
package analyzer

import (
    "testing"
    "github.com/yaleh/meta-cc/internal/testutil"
)

func TestAggregate(t *testing.T) {
    // 生成测试数据：10 个 Bash, 5 个 Edit, 3 个错误
    tools := testutil.GenerateToolCallsMixed([]testutil.ToolSpec{
        {Tool: "Bash", Count: 10, ErrorRate: 0.2},
        {Tool: "Edit", Count: 5, ErrorRate: 0.0},
    })

    config := AggregateConfig{
        GroupBy: "tool",
        Metrics: []string{"count", "error_rate"},
    }

    results, err := Aggregate(tools, config)
    if err != nil {
        t.Fatalf("Aggregate failed: %v", err)
    }

    // 验证分组数量
    if len(results) != 2 {
        t.Errorf("expected 2 groups, got %d", len(results))
    }

    // 验证 Bash 分组
    bashResult := findGroup(results, "Bash")
    if bashResult == nil {
        t.Fatal("Bash group not found")
    }

    bashCount, _ := bashResult.Metrics["count"].(int)
    if bashCount != 10 {
        t.Errorf("Bash count: expected 10, got %d", bashCount)
    }

    bashErrorRate, _ := bashResult.Metrics["error_rate"].(float64)
    if bashErrorRate != 0.2 {
        t.Errorf("Bash error_rate: expected 0.2, got %.2f", bashErrorRate)
    }
}

func TestCalculateMetric(t *testing.T) {
    tools := testutil.GenerateToolCalls(10)

    tests := []struct {
        metric   string
        validate func(value interface{}) bool
    }{
        {
            metric: "count",
            validate: func(v interface{}) bool {
                return v.(int) == 10
            },
        },
        {
            metric: "avg_duration",
            validate: func(v interface{}) bool {
                avg := v.(float64)
                return avg > 0 && avg < 10000
            },
        },
        {
            metric: "p50",
            validate: func(v interface{}) bool {
                return v.(int) > 0
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.metric, func(t *testing.T) {
            value, err := calculateMetric(tools, tt.metric)
            if err != nil {
                t.Fatalf("calculateMetric failed: %v", err)
            }

            if !tt.validate(value) {
                t.Errorf("metric %s validation failed: %v", tt.metric, value)
            }
        })
    }
}

func TestPercentileCalculation(t *testing.T) {
    // 测试百分位数计算准确性
    tools := []parser.ToolCall{
        {Duration: 100},
        {Duration: 200},
        {Duration: 300},
        {Duration: 400},
        {Duration: 500},
        {Duration: 600},
        {Duration: 700},
        {Duration: 800},
        {Duration: 900},
        {Duration: 1000},
    }

    p50, _ := calculatePercentile(tools, "p50")
    if p50 != 500 {
        t.Errorf("p50: expected 500, got %d", p50)
    }

    p90, _ := calculatePercentile(tools, "p90")
    if p90 != 900 {
        t.Errorf("p90: expected 900, got %d", p90)
    }
}

func findGroup(results []AggregateResult, groupValue string) *AggregateResult {
    for i := range results {
        if results[i].GroupValue == groupValue {
            return &results[i]
        }
    }
    return nil
}
```

### 交付物

**新建文件**:
- `internal/analyzer/aggregate.go` (~60 行)
- `internal/analyzer/aggregate_test.go` (~60 行)
- `cmd/stats.go` (~40 行)

**代码量**: ~100 行（新增 Go 源代码，不含测试）

### 验收标准

- ✅ `meta-cc stats aggregate --group-by tool --metrics "count,error_rate"` 输出正确
- ✅ 聚合统计准确（与手动计算对比，误差 < 1%）
- ✅ 百分位数计算正确（p50, p90, p95, p99）
- ✅ 所有单元测试通过

---

## Stage 10.3: 时间序列分析

### 目标

实现 `stats time-series` 命令，分析工具使用、错误率等指标随时间的变化。

### 实现步骤

#### 1. 定义时间分桶逻辑

**支持的时间间隔**:
- `hour`: 按小时分桶
- `day`: 按天分桶
- `week`: 按周分桶

#### 2. 实现时间序列分析

**文件**: `internal/analyzer/timeseries.go` (新建, ~60 行)

```go
package analyzer

import (
    "time"
    "github.com/yaleh/meta-cc/internal/parser"
)

// TimeSeriesConfig 时间序列配置
type TimeSeriesConfig struct {
    Metric   string // 指标: "tool-calls", "error-rate", "duration"
    Interval string // 时间间隔: "hour", "day", "week"
}

// TimeSeriesPoint 时间序列数据点
type TimeSeriesPoint struct {
    Timestamp time.Time `json:"timestamp"`
    Value     float64   `json:"value"`
}

// AnalyzeTimeSeries 生成时间序列数据
func AnalyzeTimeSeries(tools []parser.ToolCall, config TimeSeriesConfig) ([]TimeSeriesPoint, error) {
    if len(tools) == 0 {
        return nil, nil
    }

    // Step 1: 确定时间范围
    minTime := tools[0].Timestamp
    maxTime := tools[len(tools)-1].Timestamp

    // Step 2: 创建时间桶
    buckets := createTimeBuckets(minTime, maxTime, config.Interval)

    // Step 3: 将数据分配到桶中
    bucketData := make(map[time.Time][]parser.ToolCall)
    for _, tool := range tools {
        bucket := findBucket(tool.Timestamp, buckets, config.Interval)
        bucketData[bucket] = append(bucketData[bucket], tool)
    }

    // Step 4: 计算每个桶的指标
    var points []TimeSeriesPoint
    for _, bucket := range buckets {
        data := bucketData[bucket]
        value := calculateTimeSeriesMetric(data, config.Metric)

        points = append(points, TimeSeriesPoint{
            Timestamp: bucket,
            Value:     value,
        })
    }

    return points, nil
}

// createTimeBuckets 创建时间桶列表
func createTimeBuckets(start, end time.Time, interval string) []time.Time {
    var buckets []time.Time

    current := truncateTime(start, interval)
    endTruncated := truncateTime(end, interval)

    for !current.After(endTruncated) {
        buckets = append(buckets, current)
        current = advanceTime(current, interval)
    }

    return buckets
}

// truncateTime 截断时间到指定间隔的开始
func truncateTime(t time.Time, interval string) time.Time {
    switch interval {
    case "hour":
        return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location())
    case "day":
        return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
    case "week":
        // 周一作为一周的开始
        weekday := int(t.Weekday())
        if weekday == 0 {
            weekday = 7 // 周日调整为 7
        }
        daysBack := weekday - 1
        return time.Date(t.Year(), t.Month(), t.Day()-daysBack, 0, 0, 0, 0, t.Location())
    default:
        return t
    }
}

// advanceTime 前进一个时间间隔
func advanceTime(t time.Time, interval string) time.Time {
    switch interval {
    case "hour":
        return t.Add(time.Hour)
    case "day":
        return t.AddDate(0, 0, 1)
    case "week":
        return t.AddDate(0, 0, 7)
    default:
        return t
    }
}

// findBucket 找到时间点所属的桶
func findBucket(t time.Time, buckets []time.Time, interval string) time.Time {
    truncated := truncateTime(t, interval)

    for _, bucket := range buckets {
        if bucket.Equal(truncated) {
            return bucket
        }
    }

    // 如果没找到，返回最近的桶
    return buckets[len(buckets)-1]
}

// calculateTimeSeriesMetric 计算时间序列指标
func calculateTimeSeriesMetric(tools []parser.ToolCall, metric string) float64 {
    if len(tools) == 0 {
        return 0.0
    }

    switch metric {
    case "tool-calls":
        return float64(len(tools))

    case "error-rate":
        errorCount := 0
        for _, tool := range tools {
            if tool.Status == "error" {
                errorCount++
            }
        }
        return float64(errorCount) / float64(len(tools))

    case "avg-duration":
        sum := 0
        for _, tool := range tools {
            sum += tool.Duration
        }
        return float64(sum) / float64(len(tools))

    default:
        return 0.0
    }
}
```

#### 3. 实现 stats time-series 命令

**文件**: `cmd/stats.go` (扩展, ~40 行)

```go
var (
    timeSeriesMetricFlag   string
    timeSeriesIntervalFlag string
)

var statsTimeSeriesCmd = &cobra.Command{
    Use:   "time-series",
    Short: "Analyze metrics over time",
    Long: `Generate time series data for specified metrics.

Examples:
  meta-cc stats time-series --metric tool-calls --interval hour
  meta-cc stats time-series --metric error-rate --interval day`,
    RunE: runStatsTimeSeries,
}

func init() {
    statsCmd.AddCommand(statsTimeSeriesCmd)

    statsTimeSeriesCmd.Flags().StringVar(&timeSeriesMetricFlag, "metric", "tool-calls", "Metric to analyze (tool-calls, error-rate, avg-duration)")
    statsTimeSeriesCmd.Flags().StringVar(&timeSeriesIntervalFlag, "interval", "hour", "Time interval (hour, day, week)")
}

func runStatsTimeSeries(cmd *cobra.Command, args []string) error {
    // 定位和解析会话
    sessionPath, err := locator.LocateSession(sessionFlag, projectFlag)
    if err != nil {
        return err
    }

    tools, err := parser.ExtractToolCalls(sessionPath)
    if err != nil {
        return err
    }

    // 生成时间序列
    config := analyzer.TimeSeriesConfig{
        Metric:   timeSeriesMetricFlag,
        Interval: timeSeriesIntervalFlag,
    }

    points, err := analyzer.AnalyzeTimeSeries(tools, config)
    if err != nil {
        return err
    }

    // 输出结果
    data, _ := json.MarshalIndent(points, "", "  ")
    fmt.Println(string(data))

    return nil
}
```

### TDD 步骤

**测试文件**: `internal/analyzer/timeseries_test.go` (~50 行)

```go
package analyzer

import (
    "testing"
    "time"
)

func TestAnalyzeTimeSeries(t *testing.T) {
    // 生成跨 3 小时的测试数据
    baseTime := time.Date(2025, 10, 3, 10, 0, 0, 0, time.UTC)
    tools := []parser.ToolCall{
        {Timestamp: baseTime, Status: "success"},
        {Timestamp: baseTime.Add(30 * time.Minute), Status: "success"},
        {Timestamp: baseTime.Add(90 * time.Minute), Status: "error"},
        {Timestamp: baseTime.Add(150 * time.Minute), Status: "success"},
    }

    config := TimeSeriesConfig{
        Metric:   "tool-calls",
        Interval: "hour",
    }

    points, err := AnalyzeTimeSeries(tools, config)
    if err != nil {
        t.Fatalf("AnalyzeTimeSeries failed: %v", err)
    }

    // 验证时间桶数量（3 小时 = 3 个桶）
    if len(points) != 3 {
        t.Errorf("expected 3 time buckets, got %d", len(points))
    }

    // 验证第一个小时的数据点
    if points[0].Value != 2.0 {
        t.Errorf("hour 0: expected 2 tool calls, got %.0f", points[0].Value)
    }

    // 验证第二个小时的数据点
    if points[1].Value != 1.0 {
        t.Errorf("hour 1: expected 1 tool call, got %.0f", points[1].Value)
    }
}

func TestTimeBucketing(t *testing.T) {
    tests := []struct {
        name     string
        time     time.Time
        interval string
        expected time.Time
    }{
        {
            name:     "hour truncation",
            time:     time.Date(2025, 10, 3, 14, 35, 0, 0, time.UTC),
            interval: "hour",
            expected: time.Date(2025, 10, 3, 14, 0, 0, 0, time.UTC),
        },
        {
            name:     "day truncation",
            time:     time.Date(2025, 10, 3, 14, 35, 0, 0, time.UTC),
            interval: "day",
            expected: time.Date(2025, 10, 3, 0, 0, 0, 0, time.UTC),
        },
        {
            name:     "week truncation",
            time:     time.Date(2025, 10, 3, 14, 35, 0, 0, time.UTC), // Friday
            interval: "week",
            expected: time.Date(2025, 9, 29, 0, 0, 0, 0, time.UTC), // Monday
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := truncateTime(tt.time, tt.interval)
            if !result.Equal(tt.expected) {
                t.Errorf("expected %v, got %v", tt.expected, result)
            }
        })
    }
}
```

### 交付物

**新建文件**:
- `internal/analyzer/timeseries.go` (~60 行)
- `internal/analyzer/timeseries_test.go` (~50 行)

**修改文件**:
- `cmd/stats.go` (扩展, ~40 行)

**代码量**: ~100 行（新增 Go 源代码，不含测试）

### 验收标准

- ✅ `meta-cc stats time-series --metric tool-calls --interval hour` 输出正确
- ✅ 时间分桶正确（hour, day, week）
- ✅ 指标计算准确（tool-calls, error-rate, avg-duration）
- ✅ 所有单元测试通过

---

## Stage 10.4: 文件级统计

### 目标

实现 `stats files` 命令，分析文件操作频率、错误关联等。

### 实现步骤

#### 1. 实现文件统计分析

**文件**: `internal/analyzer/filestats.go` (新建, ~50 行)

```go
package analyzer

import (
    "sort"
    "github.com/yaleh/meta-cc/internal/parser"
)

// FileStats 文件统计信息
type FileStats struct {
    FilePath    string  `json:"file_path"`
    ReadCount   int     `json:"read_count"`
    EditCount   int     `json:"edit_count"`
    WriteCount  int     `json:"write_count"`
    ErrorCount  int     `json:"error_count"`
    TotalOps    int     `json:"total_ops"`
    ErrorRate   float64 `json:"error_rate"`
}

// AnalyzeFileStats 分析文件级统计
func AnalyzeFileStats(tools []parser.ToolCall) []FileStats {
    fileMap := make(map[string]*FileStats)

    for _, tool := range tools {
        // 提取文件路径（从 tool 的 input 参数中）
        filePath := extractFilePath(tool)
        if filePath == "" {
            continue // 跳过非文件操作
        }

        // 初始化文件统计
        if _, exists := fileMap[filePath]; !exists {
            fileMap[filePath] = &FileStats{
                FilePath: filePath,
            }
        }

        stats := fileMap[filePath]

        // 统计操作类型
        switch tool.Tool {
        case "Read":
            stats.ReadCount++
        case "Edit":
            stats.EditCount++
        case "Write":
            stats.WriteCount++
        }

        // 统计错误
        if tool.Status == "error" {
            stats.ErrorCount++
        }

        stats.TotalOps++
    }

    // 计算错误率
    for _, stats := range fileMap {
        if stats.TotalOps > 0 {
            stats.ErrorRate = float64(stats.ErrorCount) / float64(stats.TotalOps)
        }
    }

    // 转换为切片并排序
    var results []FileStats
    for _, stats := range fileMap {
        results = append(results, *stats)
    }

    // 默认按总操作数降序排序
    sort.Slice(results, func(i, j int) bool {
        return results[i].TotalOps > results[j].TotalOps
    })

    return results
}

// extractFilePath 从 ToolCall 中提取文件路径
func extractFilePath(tool parser.ToolCall) string {
    // 从 tool.Input 中提取 file_path 参数
    if filePath, ok := tool.Input["file_path"].(string); ok {
        return filePath
    }

    // 其他字段名变体
    if filePath, ok := tool.Input["path"].(string); ok {
        return filePath
    }

    return ""
}
```

#### 2. 实现 stats files 命令

**文件**: `cmd/stats.go` (扩展, ~30 行)

```go
var (
    filesSortByFlag string
    filesTopFlag    int
    filesFilterFlag string
)

var statsFilesCmd = &cobra.Command{
    Use:   "files",
    Short: "Analyze file-level statistics",
    Long: `Show statistics for files accessed during the session.

Examples:
  meta-cc stats files --sort-by edit-count --top 20
  meta-cc stats files --sort-by error-count --filter "errors>3"`,
    RunE: runStatsFiles,
}

func init() {
    statsCmd.AddCommand(statsFilesCmd)

    statsFilesCmd.Flags().StringVar(&filesSortByFlag, "sort-by", "total-ops", "Sort by field (total-ops, edit-count, error-count, error-rate)")
    statsFilesCmd.Flags().IntVar(&filesTopFlag, "top", 0, "Show only top N files")
    statsFilesCmd.Flags().StringVar(&filesFilterFlag, "filter", "", "Filter expression (e.g., \"errors>3\")")
}

func runStatsFiles(cmd *cobra.Command, args []string) error {
    // 定位和解析会话
    sessionPath, err := locator.LocateSession(sessionFlag, projectFlag)
    if err != nil {
        return err
    }

    tools, err := parser.ExtractToolCalls(sessionPath)
    if err != nil {
        return err
    }

    // 分析文件统计
    fileStats := analyzer.AnalyzeFileStats(tools)

    // 应用排序
    sortFileStats(fileStats, filesSortByFlag)

    // 应用 top N 限制
    if filesTopFlag > 0 && filesTopFlag < len(fileStats) {
        fileStats = fileStats[:filesTopFlag]
    }

    // 输出结果
    data, _ := json.MarshalIndent(fileStats, "", "  ")
    fmt.Println(string(data))

    return nil
}

func sortFileStats(stats []analyzer.FileStats, sortBy string) {
    sort.Slice(stats, func(i, j int) bool {
        switch sortBy {
        case "edit-count":
            return stats[i].EditCount > stats[j].EditCount
        case "error-count":
            return stats[i].ErrorCount > stats[j].ErrorCount
        case "error-rate":
            return stats[i].ErrorRate > stats[j].ErrorRate
        default: // "total-ops"
            return stats[i].TotalOps > stats[j].TotalOps
        }
    })
}
```

### TDD 步骤

**测试文件**: `internal/analyzer/filestats_test.go` (~40 行)

```go
package analyzer

import "testing"

func TestAnalyzeFileStats(t *testing.T) {
    tools := []parser.ToolCall{
        {Tool: "Read", Input: map[string]interface{}{"file_path": "main.go"}, Status: "success"},
        {Tool: "Edit", Input: map[string]interface{}{"file_path": "main.go"}, Status: "success"},
        {Tool: "Edit", Input: map[string]interface{}{"file_path": "main.go"}, Status: "error"},
        {Tool: "Write", Input: map[string]interface{}{"file_path": "test.go"}, Status: "success"},
    }

    stats := AnalyzeFileStats(tools)

    // 验证文件数量
    if len(stats) != 2 {
        t.Errorf("expected 2 files, got %d", len(stats))
    }

    // 验证 main.go 统计
    mainStats := findFile(stats, "main.go")
    if mainStats == nil {
        t.Fatal("main.go not found")
    }

    if mainStats.ReadCount != 1 {
        t.Errorf("main.go read_count: expected 1, got %d", mainStats.ReadCount)
    }

    if mainStats.EditCount != 2 {
        t.Errorf("main.go edit_count: expected 2, got %d", mainStats.EditCount)
    }

    if mainStats.ErrorCount != 1 {
        t.Errorf("main.go error_count: expected 1, got %d", mainStats.ErrorCount)
    }

    if mainStats.TotalOps != 3 {
        t.Errorf("main.go total_ops: expected 3, got %d", mainStats.TotalOps)
    }

    expectedErrorRate := 1.0 / 3.0
    if mainStats.ErrorRate != expectedErrorRate {
        t.Errorf("main.go error_rate: expected %.2f, got %.2f", expectedErrorRate, mainStats.ErrorRate)
    }
}

func findFile(stats []FileStats, filePath string) *FileStats {
    for i := range stats {
        if stats[i].FilePath == filePath {
            return &stats[i]
        }
    }
    return nil
}
```

### 交付物

**新建文件**:
- `internal/analyzer/filestats.go` (~50 行)
- `internal/analyzer/filestats_test.go` (~40 行)

**修改文件**:
- `cmd/stats.go` (扩展, ~30 行)

**代码量**: ~80 行（新增 Go 源代码，不含测试）

### 验收标准

- ✅ `meta-cc stats files --sort-by edit-count --top 20` 输出正确
- ✅ 文件操作计数准确（read, edit, write, error）
- ✅ 错误率计算正确
- ✅ 排序功能正常（total-ops, edit-count, error-count, error-rate）
- ✅ 所有单元测试通过

---

## 集成测试：高级查询端到端验证

### 测试脚本

**文件**: `tests/integration/advanced_query_test.sh` (~80 行)

```bash
#!/bin/bash
# Phase 10 高级查询集成测试

set -e

echo "=== Phase 10 Advanced Query Test ==="
echo ""

# Step 1: 测试高级过滤器
echo "[1/4] Testing advanced filtering..."

# 布尔组合
RESULT1=$(meta-cc query tools --where "tool='Bash' AND status='error'" --output json 2>/dev/null)
COUNT1=$(echo "$RESULT1" | jq '. | length')
echo "  ✓ Boolean AND filter: $COUNT1 records"

# 集合操作
RESULT2=$(meta-cc query tools --where "tool IN ('Bash', 'Edit')" --output json 2>/dev/null)
COUNT2=$(echo "$RESULT2" | jq '. | length')
echo "  ✓ IN operator: $COUNT2 records"

# 范围查询
RESULT3=$(meta-cc query tools --where "duration BETWEEN 500 AND 2000" --output json 2>/dev/null)
COUNT3=$(echo "$RESULT3" | jq '. | length')
echo "  ✓ BETWEEN operator: $COUNT3 records"

echo ""

# Step 2: 测试聚合统计
echo "[2/4] Testing aggregation..."

AGGREGATE=$(meta-cc stats aggregate --group-by tool --metrics "count,error_rate" --output json 2>/dev/null)
GROUP_COUNT=$(echo "$AGGREGATE" | jq '. | length')

if [ "$GROUP_COUNT" -gt 0 ]; then
    echo "  ✓ Aggregation: $GROUP_COUNT tool groups"

    # 验证指标存在
    FIRST_GROUP=$(echo "$AGGREGATE" | jq '.[0]')
    HAS_COUNT=$(echo "$FIRST_GROUP" | jq 'has("metrics") and .metrics | has("count")')
    HAS_ERROR_RATE=$(echo "$FIRST_GROUP" | jq '.metrics | has("error_rate")')

    if [ "$HAS_COUNT" = "true" ] && [ "$HAS_ERROR_RATE" = "true" ]; then
        echo "  ✓ Metrics calculated: count, error_rate"
    else
        echo "  ✗ Metrics missing"
        exit 1
    fi
else
    echo "  ✗ No aggregation groups found"
    exit 1
fi

echo ""

# Step 3: 测试时间序列
echo "[3/4] Testing time series..."

TIMESERIES=$(meta-cc stats time-series --metric tool-calls --interval hour --output json 2>/dev/null)
POINT_COUNT=$(echo "$TIMESERIES" | jq '. | length')

if [ "$POINT_COUNT" -gt 0 ]; then
    echo "  ✓ Time series: $POINT_COUNT data points"

    # 验证数据点格式
    FIRST_POINT=$(echo "$TIMESERIES" | jq '.[0]')
    HAS_TIMESTAMP=$(echo "$FIRST_POINT" | jq 'has("timestamp")')
    HAS_VALUE=$(echo "$FIRST_POINT" | jq 'has("value")')

    if [ "$HAS_TIMESTAMP" = "true" ] && [ "$HAS_VALUE" = "true" ]; then
        echo "  ✓ Time series data format valid"
    else
        echo "  ✗ Invalid time series format"
        exit 1
    fi
else
    echo "  ✗ No time series data generated"
    exit 1
fi

echo ""

# Step 4: 测试文件统计
echo "[4/4] Testing file statistics..."

FILESTATS=$(meta-cc stats files --sort-by edit-count --top 10 --output json 2>/dev/null)
FILE_COUNT=$(echo "$FILESTATS" | jq '. | length')

if [ "$FILE_COUNT" -gt 0 ]; then
    echo "  ✓ File statistics: $FILE_COUNT files analyzed"

    # 验证统计字段
    FIRST_FILE=$(echo "$FILESTATS" | jq '.[0]')
    HAS_EDIT_COUNT=$(echo "$FIRST_FILE" | jq 'has("edit_count")')
    HAS_ERROR_RATE=$(echo "$FIRST_FILE" | jq 'has("error_rate")')

    if [ "$HAS_EDIT_COUNT" = "true" ] && [ "$HAS_ERROR_RATE" = "true" ]; then
        echo "  ✓ File statistics fields valid"
    else
        echo "  ✗ Missing file statistics fields"
        exit 1
    fi
else
    echo "  ✗ No file statistics generated"
    exit 1
fi

echo ""
echo "=== All Phase 10 Tests Passed ✅ ==="
echo ""
echo "Summary:"
echo "  - Advanced filtering: working (AND, IN, BETWEEN)"
echo "  - Aggregation: working ($GROUP_COUNT groups)"
echo "  - Time series: working ($POINT_COUNT points)"
echo "  - File statistics: working ($FILE_COUNT files)"
```

---

## 文档更新

### README.md 新增章节

**新增内容** (~150 行):

````markdown
## Advanced Query Features (Phase 10)

Phase 10 extends Phase 8 query capabilities with advanced filtering, aggregation, time series analysis, and file-level statistics.

### Advanced Filtering

Use SQL-like expressions for complex queries:

```bash
# Boolean operators
meta-cc query tools --where "tool='Bash' AND status='error'"
meta-cc query tools --where "status='error' OR duration>1000"
meta-cc query tools --where "NOT (status='success')"

# Set operations
meta-cc query tools --where "tool IN ('Bash', 'Edit', 'Write')"
meta-cc query tools --where "status NOT IN ('success')"

# Range queries
meta-cc query tools --where "duration BETWEEN 500 AND 2000"
meta-cc query tools --where "timestamp BETWEEN '2025-10-01' AND '2025-10-03'"

# Pattern matching
meta-cc query tools --where "tool LIKE 'meta%'"
meta-cc query tools --where "error_text REGEXP 'permission.*denied'"

# Complex combinations
meta-cc query tools --where "(status='error' OR duration>1000) AND tool='Bash'"
```

### Aggregation Statistics

Group data and calculate metrics:

```bash
# Group by tool, show counts and error rates
meta-cc stats aggregate --group-by tool --metrics "count,error_rate,avg_duration"

# Group by status
meta-cc stats aggregate --group-by status --metrics "count,percentage"

# Multiple metrics
meta-cc stats aggregate --group-by tool --metrics "count,error_rate,p50,p90,p95"
```

**Output example**:
```json
[
  {
    "group_value": "Bash",
    "metrics": {
      "count": 127,
      "error_rate": 0.157,
      "avg_duration": 1234.5,
      "p50": 800,
      "p90": 2500
    }
  },
  {
    "group_value": "Edit",
    "metrics": {
      "count": 89,
      "error_rate": 0.034,
      "avg_duration": 567.8,
      "p50": 450,
      "p90": 1200
    }
  }
]
```

### Time Series Analysis

Analyze metrics over time:

```bash
# Tool calls per hour
meta-cc stats time-series --metric tool-calls --interval hour --output json

# Error rate per day
meta-cc stats time-series --metric error-rate --interval day

# Average duration per week
meta-cc stats time-series --metric avg-duration --interval week
```

**Use cases**:
- Identify peak usage hours
- Track error trends over time
- Monitor session activity patterns

### File-Level Statistics

Analyze file operations and identify hotspots:

```bash
# Most edited files
meta-cc stats files --sort-by edit-count --top 20

# Files with most errors
meta-cc stats files --sort-by error-count --top 10

# Files with highest error rate
meta-cc stats files --sort-by error-rate --filter "errors>3"
```

**Output example**:
```json
[
  {
    "file_path": "src/main.go",
    "read_count": 12,
    "edit_count": 45,
    "write_count": 8,
    "error_count": 3,
    "total_ops": 65,
    "error_rate": 0.046
  }
]
```

**Use cases**:
- Identify frequently modified files
- Find error-prone files
- Optimize development workflow

### Integration with Claude

Phase 10 features provide rich data for Claude analysis:

**Example @meta-coach workflow**:
```python
# 1. Get aggregated statistics
tool_stats = query("stats aggregate --group-by tool --metrics count,error_rate")

# 2. Identify problematic tools
high_error_tools = [t for t in tool_stats if t["metrics"]["error_rate"] > 0.1]

# 3. Analyze time trends for problematic tools
for tool in high_error_tools:
    trend = query(f"stats time-series --metric error-rate --filter tool='{tool}' --interval day")
    # Provide insights and recommendations...
```

### Performance Benchmarks

| Operation | 1000 turns | 2000 turns | 5000 turns |
|-----------|------------|------------|------------|
| Advanced filtering | < 50ms | < 100ms | < 200ms |
| Aggregation | < 30ms | < 60ms | < 150ms |
| Time series | < 40ms | < 80ms | < 180ms |
| File statistics | < 20ms | < 40ms | < 100ms |
````

---

## Phase 10 验收清单

### 功能验收

- [ ] **Stage 10.1: 高级过滤引擎**
  - [ ] 布尔操作符实现（AND, OR, NOT）
  - [ ] 比较操作符实现（=, !=, >, <, >=, <=）
  - [ ] 集合操作符实现（IN, NOT IN）
  - [ ] 范围操作符实现（BETWEEN ... AND ...）
  - [ ] 模式匹配实现（LIKE, REGEXP）
  - [ ] 表达式解析器工作正常
  - [ ] 单元测试通过

- [ ] **Stage 10.2: 聚合函数**
  - [ ] `stats aggregate` 命令实现
  - [ ] Group-by 功能正常
  - [ ] 基础指标：count, sum, avg, min, max
  - [ ] 高级指标：error_rate, p50, p90, p95, p99
  - [ ] 聚合结果准确（误差 < 1%）
  - [ ] 单元测试通过

- [ ] **Stage 10.3: 时间序列分析**
  - [ ] `stats time-series` 命令实现
  - [ ] 时间分桶正确（hour, day, week）
  - [ ] 指标计算准确（tool-calls, error-rate, avg-duration）
  - [ ] 时间范围覆盖完整会话
  - [ ] 单元测试通过

- [ ] **Stage 10.4: 文件级统计**
  - [ ] `stats files` 命令实现
  - [ ] 文件操作统计准确（read, edit, write, error）
  - [ ] 错误率计算正确
  - [ ] 排序功能正常（4 种排序字段）
  - [ ] 单元测试通过

### 集成验收

- [ ] **端到端测试**
  - [ ] 集成测试脚本通过（`advanced_query_test.sh`）
  - [ ] 所有 Stage 功能无缝集成
  - [ ] 真实项目验证（meta-cc, NarrativeForge, claude-tmux）

- [ ] **性能验收**
  - [ ] 查询性能 < 200ms（复杂查询，1000 turns）
  - [ ] 聚合性能 < 100ms（10 个分组）
  - [ ] 时间序列性能 < 150ms（50 个数据点）
  - [ ] 文件统计性能 < 100ms（100 个文件）

### 代码质量

- [ ] **代码量验收**
  - [ ] Stage 10.1: ~120 行（过滤引擎）
  - [ ] Stage 10.2: ~100 行（聚合函数）
  - [ ] Stage 10.3: ~100 行（时间序列）
  - [ ] Stage 10.4: ~80 行（文件统计）
  - [ ] 总计: ~400 行（目标 350-450 行）

- [ ] **测试覆盖率**
  - [ ] 单元测试覆盖率 ≥ 80%
  - [ ] 所有 TDD 测试通过
  - [ ] 边界条件测试完整

### 文档验收

- [ ] **README.md 更新**
  - [ ] 添加"高级查询功能"章节
  - [ ] 包含所有命令示例
  - [ ] 包含性能基准
  - [ ] 包含 Claude 集成示例

---

## 项目结构（Phase 10 完成后）

```
meta-cc/
├── cmd/
│   ├── query_tools.go              # 更新：--where 标志支持
│   └── stats.go                    # 新增：aggregate, time-series, files 子命令
├── internal/
│   ├── filter/
│   │   ├── expression.go           # 新增：表达式求值
│   │   ├── parser.go               # 新增：表达式解析器
│   │   ├── advanced.go             # 新增：高级过滤应用
│   │   ├── expression_test.go      # 新增：表达式测试
│   │   └── parser_test.go          # 新增：解析器测试
│   └── analyzer/
│       ├── aggregate.go            # 新增：聚合逻辑
│       ├── timeseries.go           # 新增：时间序列分析
│       ├── filestats.go            # 新增：文件统计
│       ├── aggregate_test.go       # 新增：聚合测试
│       ├── timeseries_test.go      # 新增：时间序列测试
│       └── filestats_test.go       # 新增：文件统计测试
├── tests/
│   └── integration/
│       └── advanced_query_test.sh  # 新增：集成测试脚本
├── plans/
│   └── 10/
│       ├── plan.md                 # 本文档
│       └── README.md               # 快速参考
└── README.md                        # 更新：高级查询文档
```

---

## 依赖关系

**Phase 10 依赖**:
- Phase 0-7（MVP 完整工具链 + MCP Server）
- Phase 8（查询命令基础）
- Phase 9（上下文长度应对）

**Phase 10 提供**:
- 高级过滤能力（复杂表达式查询）
- 统计聚合能力（分组计算指标）
- 时间维度分析（趋势和模式）
- 文件维度分析（操作频率和错误关联）

**后续 Phase 可选扩展**:
- Phase 11（Unix 可组合性）：流式输出、退出码标准化
- Phase 12（查询语言增强）：SQL-like 语法优化、查询优化器
- Phase 13（索引功能）：SQLite 索引、跨会话查询

---

## 风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 表达式解析器实现复杂 | 高 | 使用递归下降解析；参考成熟解析器设计；充分测试 |
| 聚合统计误差 | 中 | 使用 float64 精度；与手动计算对比验证 |
| 时间分桶边界问题 | 中 | 明确定义时间截断规则；充分测试边界条件 |
| 文件路径提取不准确 | 中 | 支持多种字段名变体；记录未识别的工具 |
| 查询性能不达标 | 高 | 使用性能测试验证；考虑添加索引（Phase 13） |
| 代码量超预算 | 低 | 每个 Stage 完成后检查；及时重构 |

---

## 实施优先级

**必须完成**（Phase 10 核心功能）:
1. Stage 10.1（高级过滤引擎）- 提供复杂查询能力
2. Stage 10.2（聚合函数）- 提供统计分析能力
3. 集成测试和文档更新

**推荐完成**（增强分析能力）:
4. Stage 10.3（时间序列分析）- 提供时间维度视角
5. Stage 10.4（文件级统计）- 提供文件维度视角

**可选完成**（进一步优化）:
6. 查询性能优化（如果性能不达标）
7. 更多聚合指标（如 stddev, variance）

---

## Phase 10 总结

Phase 10 实现高级查询能力，为 meta-cc 提供丰富的数据分析维度：

### 核心成果

1. **高级过滤引擎**（Stage 10.1）
   - SQL-like 表达式语法
   - 布尔逻辑、集合操作、范围查询、模式匹配
   - 灵活组合，精准数据检索

2. **聚合统计**（Stage 10.2）
   - 按字段分组
   - 多种指标：count, avg, error_rate, percentiles
   - 快速识别模式和异常

3. **时间序列分析**（Stage 10.3）
   - 时间维度分桶（hour/day/week）
   - 趋势识别
   - 活跃度和错误率随时间变化

4. **文件级统计**（Stage 10.4）
   - 文件操作频率
   - 错误关联
   - 高频修改文件识别

### 集成价值

- **@meta-coach**: 使用聚合和时间序列数据进行深度分析
- **MCP Server**: 提供更丰富的查询接口
- **Slash Commands**: 快速获取统计摘要

### 用户价值

- ✅ 精准查询：复杂条件过滤，快速定位问题
- ✅ 统计洞察：自动识别高频错误、慢操作、热点文件
- ✅ 趋势分析：理解工作模式，优化开发流程
- ✅ 性能优异：复杂查询 < 200ms，支持大会话

**Phase 10 完成后，meta-cc 具备生产级的数据分析能力，为 Claude 集成层提供高质量的结构化数据。**

---

## 参考文档

- [Claude Code 官方文档](https://docs.claude.com/en/docs/claude-code)
- [meta-cc 技术方案](../../docs/proposals/meta-cognition-proposal.md)
- [meta-cc 总体实施计划](../../docs/plan.md)
- [Phase 8 实施计划](../8/phase-8-implementation-plan.md)（查询基础）
- [Phase 9 实施计划](../9/plan.md)（上下文管理）

---

**Phase 10 实施准备就绪。开始 TDD 开发流程。**
