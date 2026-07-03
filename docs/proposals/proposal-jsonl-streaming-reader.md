# Proposal: JSONL Streaming Reader with Early Image Filtering

**Status**: Implemented (2026-06-23, see internal/parser/streaming_reader.go — bufio.Reader.ReadBytes with early image truncation)
**Date**: 2026-03-12
**Scope**: I/O layer — parser and all query executor paths

---

## 背景

### 问题根源

Claude Code 将 MCP 工具返回的截图（base64 编码图片）完整序列化到 JSONL 会话文件。每张截图产生一行超大 JSON，例如：

- 文件 `670a30a2-f413-4fdc-b2e4-ae05779aff05.jsonl` 第 262 行：**6.8 MB**
- 行结构：`type=user`，包含 `tool_result`，其中嵌套单个 `image` block，`source.data` 字段为 3.4 MB base64 字符串

图片 block 在 JSONL 中的实际 JSON 结构为（嵌套于 `tool_result.content[]` 数组内）：

```json
{
  "type": "user",
  "message": {
    "role": "user",
    "content": [{
      "type": "tool_result",
      "tool_use_id": "...",
      "content": [{
        "type": "image",
        "source": {
          "type": "base64",
          "media_type": "image/png",
          "data": "<3.4MB base64 string>"
        }
      }]
    }]
  }
}
```

**重要**：`image` block 嵌套于 `tool_result.content[]` 数组中，不是 `tool_result` 的直接子字段。这使得"跳过整个 `tool_result`"（策略 A）和"精确截断 `data` 字段"（策略 B）的适用边界与检测逻辑都需要针对此嵌套结构设计。

### 当前实现的 Scanner 分布（完整清单）

经代码全量扫描，实际存在 **10 处** `bufio.Scanner` 读取 JSONL 的位置，原文档列出 7 处，遗漏了以下 3 处：

| 文件 | 位置/函数 | 上限 | 是否在原文档中 |
|------|----------|------|--------------|
| `internal/types/constants.go` | `MaxScannerLineBytes` 定义 | 4 MB | 是（定义） |
| `internal/parser/reader.go` | `ParseEntries` | 4 MB（via alias） | 是 |
| `cmd/mcp-server/query_executor.go` | `processFileWithTimeRange` | 4 MB | 是 |
| `cmd/mcp-server/query_executor.go` | `processFile` | 4 MB | 是 |
| `internal/query/jq/stage2_executor.go` | `readJSONLFile` | 10 MB | 是 |
| `internal/query/stage2_executor.go` | `readJSONLFile` | 10 MB | 是 |
| `internal/mcp/filters/filters.go` | `loadTurnsForSession` | 10 MB | 是 |
| `cmd/mcp-server/handlers_query.go` | `loadTurnsForSession` | 10 MB | 是 |
| **`internal/query/files/file_inspector.go`** | `InspectFiles` | **10 MB** | **否（遗漏）** |
| **`cmd/mcp-server/handlers_stage1.go`** | `countLines` | **无显式上限（64KB 默认）** | **否（遗漏）** |
| **`cmd/mcp-server/main.go`** | stdin JSON-RPC 读取 | **64KB 默认** | **否（不适用）** |

**注**：`cmd/mcp-server/main.go` 的 Scanner 读取的是 JSON-RPC 协议消息（stdin），不是 JSONL 会话文件，**不在本 Proposal 范围内**，应单独处理。

`cmd/mcp-server/handlers_stage1.go` 的 `countLines` 函数仅计数行数，不解析 JSON 内容。当遇到超大行时会返回 `ErrTooLong`，导致文件记录数统计失败，属于静默错误（调用方不检查返回错误）。

`internal/query/files/file_inspector.go` 的 `InspectFiles` 不仅计数，还对每行调用 `json.Unmarshal`，遇到 10 MB+ 超大行会失败，影响文件检查功能。

### 行为不一致清单（已确认）

通过读取实际代码，各路径对 `ErrTooLong` 的实际处理方式如下：

| 文件 | 函数 | 遇到 ErrTooLong 时的行为 |
|------|------|----------------------|
| `internal/parser/reader.go` | `ParseEntries` | `return nil, error`（**整个文件失败**） |
| `cmd/mcp-server/query_executor.go` | `processFile` | `return results, error`（**整个文件失败，调用方记录 warning 后继续**） |
| `cmd/mcp-server/query_executor.go` | `processFileWithTimeRange` | 同上 |
| `internal/query/jq/stage2_executor.go` | `readJSONLFile` | `return nil, error`（**整个查询失败，不是跳过文件**） |
| `internal/query/stage2_executor.go` | `readJSONLFile` | 同上 |
| `internal/mcp/filters/filters.go` | `loadTurnsForSession` | scanner.Err() 未检查，**静默丢失错误** |
| `cmd/mcp-server/handlers_query.go` | `loadTurnsForSession` | 同上 |
| `internal/query/files/file_inspector.go` | `InspectFiles` | scanner.Err() 检查不明确，需确认 |
| `cmd/mcp-server/handlers_stage1.go` | `countLines` | `return 0, error`（调用方忽略）|

**严重性差异**：`internal/query/jq/stage2_executor.go` 的失败会导致整个 stage2 查询中断并返回错误给用户，不只是跳过文件。这与文档原文"部分路径静默跳过"的描述不符，实际更严重。

### 后果

- 包含截图的会话文件在 4 MB 上限路径下完全不可读——所有条目丢失。
- 10 MB 上限仅是权宜之计；若 Claude Code 返回多张截图，单行仍可突破此限。
- 当前处理模式：**先把整行读入内存 → json.Unmarshal 全量反序列化 → 判断是图片 → 丢弃**。在 I/O 最后阶段才过滤，内存已经被占用。
- meta-cc 所有查询接口对图片二进制内容的实际需求为零；没有任何 MCP 工具需要读取 `source.data` 字段的内容。

---

## 目标

1. **消除硬性行长限制**：任意大小的行都能被读取，不因行长报错而跳过文件。
2. **不 materialize 二进制内容**：在 I/O 层识别并跳过（或截断）图片 base64 数据，不将其加载到内存。
3. **零功能回归**：已有查询结果不受影响；图片行的结构元数据（`type`、`sessionId`、`timestamp`、block 类型标识）仍可被查询。
4. **统一常量管理**：消除各处散落的字面量 `10*1024*1024`，明确 `MaxScannerLineBytes` 的最终命运（见下文）。

---

## 方案设计

### 核心思路：在 I/O 层早期过滤，而非反序列化后过滤

用 `bufio.Reader.ReadBytes('\n')` 替代 `bufio.Scanner`。`ReadBytes` 无硬性行长限制，遇到超长行按需分配内存直至换行符，完整返回行字节。关键在于：获得完整行字节后，**在调用 `json.Unmarshal` 之前**，先做廉价的字节级判断，决定如何处理该行。

**`ReadBytes` 的内存行为（重要）**：`ReadBytes` 内部通过 `ReadSlice` 循环，将跨越内部缓冲区边界的片段 append 到累积切片，最终分配一个包含完整行的新 `[]byte`。这意味着：
- 对超大行，`ReadBytes` 仍会一次性在堆上持有整行字节。
- **必须配合早期过滤策略**，在 `Unmarshal` 之前截断大字段，才能将超大字节切片的生命周期限制在截断之前，让 GC 尽快回收。
- 如果省略早期过滤，`ReadBytes` 与 `Scanner` 的区别仅是"不报错"而非"不占内存"。

### 两种过滤策略

#### 策略 A：Peek 快速跳过（仅适用于明确不需要 tool_result 内容的路径）

在已知查询只关心文本内容（`text` block）时，对包含 `image` block 的 `tool_result` 行可以完全跳过，无需解析。

**触发条件**（字节检测顺序，从低开销到高开销）：
1. `bytes.Contains(line, []byte(`"type":"image"`))`——命中则整行跳过

**代价**：
- 误跳过*同时含有文本内容和图片的* `tool_result` 行——这类行在实践中存在（MCP 工具可以返回混合内容）。
- **不会**误跳过只含文本 `tool_result` 的行，因为这些行不含 `"type":"image"`。

**适用范围（明确边界）**：

| 查询路径 | 是否适用策略 A | 原因 |
|---------|--------------|------|
| token 用量统计 | 是 | 仅需 `usage` 字段，不需要 `tool_result` content |
| 工具调用次数/类型统计 | 是 | 仅需 `tool_use` block，不需要 `tool_result` |
| 用户文本消息查询 | 是 | 仅需 `type=user` 且 `content[].type=text` |
| 时间线分析 | 是 | 仅需 `timestamp` |
| 工具输出内容全文检索 | **否** | 需要 `tool_result` content 字段 |
| stage2 jq 查询（任意表达式） | **否** | 无法静态确定是否需要 `tool_result` |
| `loadTurnsForSession`（context expand） | **否** | 需要完整 turn 结构 |

**结论**：策略 A **不应作为默认行为**，只能在调用方可以静态保证不需要 `tool_result` 的少数路径中使用。在通用底层 reader 中不应默认开启。

#### 策略 B：流式截断 image source.data（通用方案，推荐默认）

不跳过整行，而是在字节流层面识别并截断 `source.data` 字段的值，保留行的完整 JSON 结构。

**检测条件**（必须同时满足，减少误匹配）：
```
bytes.Contains(line, []byte(`"type":"image"`)) &&
bytes.Contains(line, []byte(`"type":"base64"`))
```

**截断逻辑**：定位 `"data":"` 子串后，找到紧随其后的 JSON 字符串的结束引号（非转义），将字符串值替换为占位符。

**base64 的转义安全性**：RFC 4648 base64 字符集（`A-Za-z0-9+/=`）不包含 `\` 或 `"`，因此 base64 字符串内部不存在转义引号。可以安全地搜索第一个未转义的 `"` 作为字符串结束符，无需完整的 JSON 字符串解析器。

**实现要点**：
- 用 `bytes.Index(line, []byte(`"data":"`))` 定位起始位置（注意：`"data"` 在 JSON 中可能有空格，实际 Claude Code 输出通常无空格）
- 找到值起始引号后，线性扫描找到结束引号，用 `[]byte(`"<binary-omitted>"`)` 替换
- 替换后行长度大幅缩小（从 6.8 MB 降至约 200 字节）
- 替换后调用 `json.Valid()` 验证；不合法则降级为跳过整行并记录警告

**关于 `encoding/json.Decoder.Token()` 的评估**：

该方案在文档中被列为备选，需要明确拒绝原因：

1. Go 标准库的 `json.Decoder.Token()` 即使遇到大字符串值，也会将整个 token 值分配到内存（`bufio.Reader` 内部缓冲机制），**不能实现真正的内存节省**。
2. 重组删除字段后的 JSON 需要手动序列化，实现复杂且容易引入 key 顺序变化，影响下游 jq 查询。
3. `encoding/json/v2`（`jsontext`）的 `SkipValue()` 可以真正跳过而不分配，但目前仍是实验性 API（`go get golang.org/x/exp/jsonv2`），不应引入实验性依赖。

**结论**：策略 B 的字节替换方案在 base64 场景下是正确的、安全的，且性能优于 Token 流。

#### 多图片 block 的处理

一个 `tool_result` 可能包含多个 `image` block（例如截图序列）。上述策略 B 的截断逻辑**必须循环处理**，不能在替换第一个 `"data":"..."` 后停止。实现时应在 while 循环中重复定位并替换，直到行中不再包含 `"type":"base64"` 为止。

原文档仅在验收标准中提到"多图片 block"，但截断逻辑描述未明确要求循环，这是设计遗漏。

#### 非 image 的大 tool_result（文本内容巨大的工具返回）

部分工具（如 Bash 执行长输出、Read 读取大文件）可能返回数百 KB 的纯文本 `tool_result`。这类行：
- 不触发策略 A 或 B 的检测条件（不含 `"type":"image"`）
- 会被 `ReadBytes` 完整读入内存
- 目前 10 MB 上限下勉强能处理；若改为 `ReadBytes` 则无上限，但也不会截断

**当前 Proposal 不覆盖此场景**，属于"不在范围内"，但需要在文档中明确说明：`ReadBytes` 方案对此类行不提供内存保护，这是有意识的取舍（纯文本内容的查询确实需要这些数据）。

### 改动范围（修订版）

**新增**：`internal/parser/streaming_reader.go`

- 提供 `ReadLineFiltered(reader *bufio.Reader, strategy FilterStrategy) ([]byte, bool, error)` 接口
- `FilterStrategy` 类型为 int 常量，定义 `StrategyDefault`（策略 B）、`StrategySkipImage`（策略 A）
- `bool` 返回值表示"行是否被完全跳过"
- 内部实现：`reader.ReadBytes('\n')`，再按策略处理

**修改**：以下文件中所有 JSONL 文件读取的 `bufio.Scanner` + `scanner.Buffer(buf, N)` 模式替换为新接口：

| 文件 | 函数 | 推荐策略 | 备注 |
|------|------|---------|------|
| `internal/parser/reader.go` | `ParseEntries` | B | 需要保留 tool_result 元数据 |
| `cmd/mcp-server/query_executor.go` | `processFile` | B | stage1 查询，需通用兼容 |
| `cmd/mcp-server/query_executor.go` | `processFileWithTimeRange` | B | 同上 |
| `internal/query/jq/stage2_executor.go` | `readJSONLFile` | B | stage2 jq，表达式任意 |
| `internal/query/stage2_executor.go` | `readJSONLFile` | B | 同上 |
| `internal/mcp/filters/filters.go` | `loadTurnsForSession` | B | context expand，需完整 turn |
| `cmd/mcp-server/handlers_query.go` | `loadTurnsForSession` | B | 同上 |
| **`internal/query/files/file_inspector.go`** | `InspectFiles` | **B** | **原文档遗漏** |
| **`cmd/mcp-server/handlers_stage1.go`** | **`countLines`** | **专用字节计数法（`bufio.Reader.ReadBytes('\n')` 循环或 `bytes.Count`）** | **原文档遗漏；仅计行数，可用字节计数法完全绕开 Scanner** |

**`countLines` 的特殊处理**：该函数仅统计行数，不解析 JSON，使用 `bufio.Reader` + `ReadBytes('\n')` 循环计换行符的方式替代，完全不依赖 Scanner，性能更好，无行长限制。`bytes.Count(fileBytes, []byte("\n"))` 需要将整个文件读入内存，不适合大文件；`bufio.Reader.ReadBytes` 逐块读取，内存占用固定，为推荐方案。

**不修改**：
- `cmd/mcp-server/main.go` 的 stdin Scanner：读取 JSON-RPC 协议消息，不是 JSONL 会话文件，行长不受会话内容影响。

**修改**：`internal/types/constants.go`

见下文 `MaxScannerLineBytes` 处置方案。

---

## `MaxScannerLineBytes` 常量的处置

原文档对此的描述是"移除或标记废弃"，过于模糊，需要明确决策。

### 三种选项

**选项 1：彻底移除**
- 移除 `internal/types/constants.go` 中的定义和 `internal/parser/aliases.go` 中的别名
- 优点：彻底消除误用可能
- 风险：若有未被发现的 Scanner 使用点，编译会报错（此处是优点而非风险）
- **推荐**：改造完成后应选此选项

**选项 2：保留为软限/注释常量**
- 更名为 `DeprecatedScannerLineBytes`，并添加注释说明废弃原因
- Go 无官方废弃机制，依赖注释提示，效果有限
- **不推荐**：残留常量容易被误用

**选项 3：转型为软限警告阈值**
- 保留常量，用于 `ReadBytes` 之后检测行长度，若超过阈值记录 debug 日志
- `if len(line) > MaxScannerLineBytes { slog.Debug("large line detected", ...) }`
- 优点：保留可观测性，便于发现异常大行
- 缺点：常量语义从"硬性限制"变为"监控阈值"，需重命名（如 `LargeLineWarnBytes`）

**决策**：
- **改造阶段（实施时）**：选项 3，常量重命名为 `LargeLineWarnBytes`，用于监控日志，阈值保持 4 MB
- **稳定后（验收测试通过后）**：选项 1，移除常量，监控逻辑改用硬编码或配置

`internal/parser/aliases.go` 中的 `MaxScannerLineBytes` 别名在改造阶段同步更新为 `LargeLineWarnBytes` 别名，稳定后同步移除。

---

## 权衡分析

### bufio.Reader.ReadBytes vs bufio.Scanner

| 维度 | bufio.Scanner | bufio.Reader.ReadBytes |
|------|--------------|----------------------|
| 行长限制 | 硬性上限，超出报错 | 无（受可用内存限制） |
| API 简洁性 | 高（`for scanner.Scan()`） | 稍低（需手动处理 `isPrefix`，或用 `ReadBytes`） |
| 超大行行为 | ErrTooLong，整行丢失 | 正常返回，调用方可处理 |
| 内存分配（普通行） | 预分配固定 buffer，复用 | 按需分配，但 `bufio.Reader` 内部 buffer 会被复用 |
| 内存分配（超大行） | 失败报错 | 分配完整行大小的新切片 |
| 与早期过滤配合 | 不适用（先读完整行再报错） | **必须配合**，否则超大行仍占内存 |

**注意**：`bufio.Reader.ReadLine()` 不适合此场景——它对超过内部 buffer 的行只返回前一部分（`isPrefix=true`），需要调用方手动拼接。`ReadBytes('\n')` 自动处理拼接，是正确选择。

### 简单字节替换 vs encoding/json.Decoder Token 流

| 维度 | 字节替换（推荐） | Decoder Token 流 |
|------|---------|----------------|
| 实现复杂度 | 低 | 高 |
| 内存节省 | 替换后原大切片可被 GC | Token 流仍将 string token 分配到内存 |
| 正确性风险 | 低（base64 无转义字符） | 低（标准库保证） |
| 性能 | 高（单次线性扫描） | 低（完整词法分析） |
| 实验性依赖 | 无 | `jsontext` 仍是实验性 API |

对于 `source.data` 的定位，base64 字符串（RFC 4648）不含 `\` 或 `"`，字节替换方案在此场景是**严格正确的**，通过单元测试覆盖边界情况即可。

---

## 风险

### R1：字节替换引入静默错误

**风险**：错误定位 `source.data` 边界，导致生成格式错误的 JSON，后续 `json.Unmarshal` 失败。
**缓解**：
- 替换逻辑需要完整的单元测试，覆盖以下场景（见测试策略章节）。
- 替换后调用 `json.Valid()` 验证；不合法则记录警告并跳过整行（退化为已有行为）。
- 若 `"data":` key 在 JSON 中带空格（如 `"data" : "`），当前实现会漏检。需使用正则 `"data"\s*:\s*"` 或同时检测多种格式。

### R2：性能回归（超大行内存分配）

**风险**：`ReadBytes` 对超大行一次性分配大块内存，GC 压力增加。
**缓解**：
- 早期过滤（策略 B）在截断后，原始大切片在 `Unmarshal` 调用前即可被 GC（调用方只持有截断后的小切片）。
- 对于普通行（< 64KB），`bufio.Reader` 默认 4096 字节内部缓冲已足够，仅分配最终返回切片，无额外分配。
- 添加行长度监控日志（debug 级别），阈值使用 `LargeLineWarnBytes`。

### R3：行为变更影响现有测试

**风险**：现有单元测试依赖 `ErrTooLong` 行为或依赖图片 `source.data` 被反序列化。
**缓解**：
- 全量运行 `make commit` 捕获回归。
- 策略 B 将 `source.data` 值替换为 `"<binary-omitted>"`——若有测试断言该字段的原始内容，需更新为接受占位符值。
- 当前 `internal/types/session.go` 的 `ToolResult.UnmarshalJSON` 将 `content` 处理为 `[]text_block`，不处理 `image` block（会丢失）。策略 B 不影响此行为（占位符字符串符合 string 类型约束）。

### R4：分散的硬编码字面量未完全清除

**风险**：遗漏 `internal/query/files/file_inspector.go` 和 `handlers_stage1.go` 两处（原文档未列出）。
**缓解**：
- 添加 CI 检查规则（`make lint` 自定义规则或 `grep` 检查），禁止 JSONL 读取路径中直接使用 `bufio.NewScanner`（或要求附带注释说明理由）。
- 实施时用全局搜索 `bufio.Scanner` + `bufio.NewScanner` 确认无遗漏，以本文档的完整清单（10 处）为基线。

### R5：`readJSONLFile` 的错误语义变更（新增风险）

**风险**：`internal/query/jq/stage2_executor.go` 和 `internal/query/stage2_executor.go` 的 `readJSONLFile` 当前在 `ErrTooLong` 时 `return nil, error`，导致**整个 stage2 查询失败**。改为 `ReadBytes` 后，超大行会被正常处理（截断后解析），不再导致查询失败。

这是正向改变，但需要在 stage2 测试中验证：原来因行长超限而失败的查询，现在应返回结果而非错误。**相关测试用例需要更新期望值。**

---

## 测试策略

### 验证"修复前后的行为差异"的具体方法

测试策略的核心是**可证伪性**：每个测试必须能够在改造前失败、改造后通过。

#### T1：行长超限不再导致文件失败

```go
// 构造含 5 MB base64 image block 的 JSONL 行
func TestStreamingReader_LargeImageLine_NotSkipped(t *testing.T) {
    // 构造一个包含 5MB base64 数据的合法 JSONL 行
    bigData := strings.Repeat("A", 5*1024*1024) // base64 字符
    line := buildImageToolResultLine(bigData)

    // 在改造前：bufio.Scanner 会返回 ErrTooLong，ParseEntries 返回 error
    // 在改造后：该行应被成功读取（data 被替换为占位符），其他字段正常
    entries, err := parseFromLine(line + "\n")
    assert.NoError(t, err)
    assert.Equal(t, "<binary-omitted>", extractImageData(entries[0]))
}
```

#### T2：相邻正常行在超大行后仍被解析

```go
// 在改造前：超大行导致 ParseEntries 提前返回，后续行丢失
// 在改造后：超大行被截断处理，后续行正常解析
func TestStreamingReader_NormalLineAfterLargeLine_Preserved(t *testing.T) {
    content := buildImageLine(5*1024*1024) + "\n" +
        `{"type":"user","uuid":"normal-uuid","timestamp":"...","message":{...}}` + "\n"
    entries, err := ParseEntries(content)
    assert.NoError(t, err)
    assert.Len(t, entries, 2) // 两行都应解析成功
    assert.Equal(t, "normal-uuid", entries[1].UUID)
}
```

#### T3：多图片 block 均被截断

```go
func TestStreamingReader_MultipleImageBlocks_AllTruncated(t *testing.T) {
    line := buildMultiImageToolResultLine([]string{
        strings.Repeat("A", 1*1024*1024), // image 1
        strings.Repeat("B", 2*1024*1024), // image 2
    })
    result := stripImageData([]byte(line))
    assert.True(t, json.Valid(result))
    assert.Equal(t, 2, countOccurrences(result, "<binary-omitted>"))
    assert.NotContains(t, string(result), strings.Repeat("A", 100))
    assert.NotContains(t, string(result), strings.Repeat("B", 100))
}
```

#### T4：纯文本 tool_result 不受影响

```go
func TestStreamingReader_TextToolResult_Unchanged(t *testing.T) {
    line := `{"type":"user","message":{"content":[{"type":"tool_result",` +
        `"content":"bash output text here","tool_use_id":"xyz"}]}}`
    result := stripImageData([]byte(line))
    assert.Equal(t, []byte(line), result) // 无变化
}
```

#### T5：替换后 JSON 合法性验证

```go
func TestStreamingReader_AfterStrip_ValidJSON(t *testing.T) {
    // 构造合法的含 image 的 JSONL 行
    line := buildRealWorldImageLine()
    result := stripImageData([]byte(line))
    assert.True(t, json.Valid(result), "替换后必须是合法 JSON")
}
```

#### T6：`countLines` 不再因超大行失败（针对 handlers_stage1.go 的特殊情况）

```go
func TestCountLines_LargeImageLine_NoError(t *testing.T) {
    // 在改造前：bufio.Scanner 默认 64KB 上限，ErrTooLong，返回 0, error
    // 在改造后：行计数正确
    count, err := countLines(fileWithLargeImageLine)
    assert.NoError(t, err)
    assert.Equal(t, expectedLineCount, count)
}
```

### 测试覆盖要求

`streaming_reader.go` 的单元测试覆盖率 ≥ 80%，必须包含：

| 测试场景 | 验证目标 |
|---------|---------|
| 普通文本行 | 不触发过滤，原样返回 |
| 纯文本 tool_result | 不触发过滤，原样返回 |
| 含单个 image block 的 tool_result | `source.data` 替换为占位符 |
| 含多个 image block 的 tool_result | 所有 `source.data` 均被替换 |
| 非 tool_result 的 image（如 assistant message inline image） | 正确处理（base64 仍被截断） |
| 替换后 JSON 合法性 | `json.Valid()` 通过 |
| `json.Valid()` 失败的降级路径 | 行被跳过，记录警告 |
| 策略 A 对含 image 行的处理 | 整行跳过，`bool` 返回 `true` |
| 策略 A 对不含 image 的 tool_result 行 | 不跳过 |
| 行尾无换行符（文件末行） | 正确处理 |
| 空行 | 跳过，不报错 |

---

## 验收标准

1. 文件 `670a30a2-f413-4fdc-b2e4-ae05779aff05.jsonl` 可被完整解析，第 262 行不再触发错误，其他行的数据正确返回。
2. 所有现有测试通过（`make commit` 绿色）。
3. 内存使用：解析包含单张 6.8 MB 截图的文件时，堆峰值增量 < 1 MB（相比不含截图的同等文件），通过 `go test -memprofile` 验证。
4. `internal/types/constants.go` 中 `MaxScannerLineBytes` 重命名为 `LargeLineWarnBytes`（或在稳定后移除），各文件无 `10*1024*1024` 字面量残留。
5. 新增 `streaming_reader` 的单元测试覆盖率 ≥ 80%，包含本文档测试策略章节中列出的全部场景。
6. `internal/query/files/file_inspector.go` 和 `cmd/mcp-server/handlers_stage1.go` 的 Scanner 改造完成（原文档未列入验收，补充）。

---

## 不在范围内

- 支持跨行 JSON（JSONL 格式本身保证单条记录单行）。
- 将图片数据存储到外部存储或缓存。
- 修改 Claude Code 的 JSONL 写入行为（上游，不可控）。
- 对 `ParseEntriesFromContent`（测试辅助函数）应用同样的流式改造——该函数操作内存字符串，不涉及 I/O，可保持现状。
- 对大型纯文本 `tool_result`（如 Bash 长输出）的内存限制——此类行不含图片，确实需要完整内容，内存使用属于正常范围。
- `cmd/mcp-server/main.go` 的 stdin JSON-RPC Scanner——读取协议消息，行长不受会话内容影响，不在本范围。

---

## 重大问题记录（架构师审查标注）

### ISSUE-1：原文档 Scanner 数量统计不完整

**等级**：中（影响改造范围完整性）

原文档列出 7 处 Scanner，实际代码中有 10 处涉及 JSONL 文件读取。遗漏：
- `internal/query/files/file_inspector.go:89`（10 MB 上限，`InspectFiles` 函数）
- `cmd/mcp-server/handlers_stage1.go:350`（无显式上限，`countLines` 函数）

`countLines` 使用默认 64KB 上限，比 4MB 更严格，对于包含截图的会话文件必然失败。

**处置**：已在本文档的完整清单和改动范围中补充。

### ISSUE-2：策略 A 的适用边界不清晰

**等级**：中（影响方案正确性）

原文档表述"策略 A 适用于 content 类型为 string 的查询"，但 Claude Code 的 `tool_result` 可以同时包含文本和图片（混合 content 数组）。用 `bytes.Contains(line, []byte(`"tool_result"`))`（原文建议）跳过整行会误丢失包含文本的 `tool_result`。

实际上策略 A 应该以 `"type":"image"` 而非 `"tool_result"` 作为检测键，且应明确标注"不适用于 stage2 查询"。

**处置**：已在本文档策略 A 章节中更正检测条件和适用范围表格。

### ISSUE-3：多图片 block 的循环替换未在截断逻辑中明确

**等级**：低（影响实现正确性）

原文档在"实现要点"中只描述了单次替换，未说明多图片 block 场景需要循环。但验收标准第 5 条已要求覆盖此测试。实现时需确保使用循环。

**处置**：已在策略 B 章节补充"循环处理"要求。

### ISSUE-4：stage2 失败语义与原描述不符

**等级**：中（影响测试期望值设置）

原文档称"部分路径静默跳过"，但实际上 `internal/query/jq/stage2_executor.go` 的 `readJSONLFile` 在 `ErrTooLong` 时 `return nil, error`，导致整个 stage2 查询以错误返回，而非跳过文件。这使改造后的行为变化更大——原来返回错误的查询，现在将返回数据。需要在 stage2 集成测试中更新期望值。

**处置**：已在 R5 风险章节记录，并在测试策略中要求覆盖此场景。

### ISSUE-5：`json.Decoder.Token()` 的内存节省评估有误

**等级**：低（影响方案对比可信度）

原文档将 `encoding/json.Decoder Token 流` 列为备选，但 Go 标准库的 `Token()` 实现仍会将完整 string token 分配到内存，无法实现"不 materialize 大字段"的目标。真正能跳过大字段的是实验性的 `jsontext.SkipValue()`，但引入实验性依赖不可接受。

**处置**：已在策略 B 章节补充明确拒绝 Token 流的原因。
