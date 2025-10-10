# /meta-bugs - Project-Level Bug Analysis

## Overview

`/meta-bugs` is a slash command that performs **meta-cognitive analysis** of project-level bugs and fix patterns, complementing `/meta-errors` which focuses on tool-level technical errors.

## Design Philosophy

### `/meta-errors` vs `/meta-bugs`

| Aspect | `/meta-errors` | `/meta-bugs` |
|--------|---------------|-------------|
| **Focus** | Tool execution failures | Workflow and process issues |
| **Data Source** | `tool_result.is_error` | User messages, conversations |
| **Detection** | Error signatures, tool status | Semantic pattern matching |
| **Analysis Level** | Technical (what failed) | Meta-cognitive (why & how to improve) |
| **Output** | Error patterns, technical fixes | Workflow insights, process improvements |

### Key Differences

**`/meta-errors`** (Tool-centric):
- Bash exit codes
- File not found
- Permission denied
- MCP tool errors
- → **Technical debugging**

**`/meta-bugs`** (Workflow-centric):
- Test failures
- Build failures
- Git conflicts
- User corrections (interrupts, rejections)
- Fix effectiveness
- → **Process optimization**

## Implementation Strategy

### Pure MCP-Driven Architecture

`/meta-bugs` uses **only MCP tools** for data collection and analysis, with **no Go code modifications required**:

1. **Data Collection** (MCP Tools):
   - `query_user_messages`: Detect workflow failures and user corrections
   - `query_conversation`: Detect resolution signals
   - `query_tools`: Get error context
   - `query_tool_sequences`: Identify workflow patterns
   - `get_session_stats`: Calculate rates and ratios

2. **Pattern Detection** (Regex Matching):
   - Test failures: `test.*fail|tests?.*failed|make test.*error`
   - Build failures: `build.*fail|make.*error|compilation.*error`
   - User interrupts: `stop|interrupt|cancel|clear`
   - User rejections: `wrong|incorrect|not what|mistake`
   - Resolutions: `test.*pass|build.*success|fixed|resolved`

3. **Analysis** (Claude LLM):
   - Group and classify patterns
   - Calculate statistics (rates, counts, averages)
   - Identify repeated issues
   - Generate actionable recommendations

### Data Flow

```
User message → MCP query → Pattern matching → Classification → Insights → Recommendations
```

Example:
```
User: "tests failed again"
  ↓
query_user_messages(pattern="test.*fail")
  ↓
Detect workflow failure (type: "test")
  ↓
Group similar failures
  ↓
Calculate fix effectiveness
  ↓
Generate recommendation: "Run tests before major refactors"
```

## Output Structure

```markdown
## 📊 Project-Level Bug Analysis

### 🎯 Summary
- Total Workflow Failures: X
- Total User Corrections: Y
- Repeated Issues: Z
- Average Fix Attempts: N
- Resolution Rate: P%

### 🔴 Critical Issues
- Repeated failures (≥3 occurrences)
- Unresolved workflow failures
- High correction rate signals

### 📉 Workflow Failures Breakdown
- Test Failures: count, rate, examples, insights
- Build Failures: count, rate, examples, insights
- Lint Failures: count, rate, examples, insights
- Git Issues: count, rate, examples, insights

### 🔄 User Correction Patterns
- Interruptions: signals of frustration or long operations
- Rejections: expectation mismatches
- Retries: previous approach failed

### 🛠️ Fix Effectiveness Analysis
- Repeated issues
- Unresolved issues
- Average fix attempts
- Resolution rate

### 💡 Recommendations
- Immediate Actions
- Workflow Improvements
- Process Optimizations
- Prevention Strategies
```

## Usage Examples

### Example 1: Detect Test Instability

**User runs**: `/meta-bugs`

**Analysis finds**:
- 12 test failures across project
- Pattern: Type errors after interface changes (8x)
- Average fix attempts: 3.2

**Recommendation**:
> Run `make test` before major refactors to catch type errors early

### Example 2: Identify User Frustration

**User runs**: `/meta-bugs`

**Analysis finds**:
- 7 interruptions (stop/cancel)
- Pattern: User stops Claude during long bash sequences

**Recommendation**:
> Break down tasks into smaller, verifiable steps

### Example 3: Repeated Issues

**User runs**: `/meta-bugs`

**Analysis finds**:
- "is_error field parsing" issue occurred 3 times
- Resolution rate: 85%

**Recommendation**:
> Document common fixes in CLAUDE.md to prevent recurrence

## Comparison with Your Analysis Method

### Alignment with "我的分析" (Your Analysis)

| Your Analysis Method | `/meta-bugs` Implementation |
|---------------------|---------------------------|
| 多维度数据收集 | ✅ MCP tools (messages, conversation, tools) |
| 分层分析方法 | ✅ Detect → Classify → Analyze → Recommend |
| 模式识别 | ✅ Regex + grouping + frequency analysis |
| 修复过程分析 | ✅ Fix cycles, resolution rate, repeated issues |
| 可操作建议 | ✅ Immediate actions + prevention strategies |

### Key Advantages

1. **No Go Code Changes**: Pure MCP-driven, easy to maintain
2. **Leverages Claude's Strengths**: LLM performs semantic analysis
3. **Complementary to `/meta-errors`**: Different analysis layers
4. **Actionable Output**: Every issue → recommendation
5. **Meta-Cognitive Focus**: "How to improve" not just "what failed"

## Technical Notes

### Pattern Detection Accuracy

**High Precision Patterns** (low false positives):
- Test failures: `test.*fail|tests?.*failed`
- Build failures: `build.*fail|make.*error`

**High Recall Patterns** (catch all signals):
- User corrections: `stop|interrupt|wrong|retry`

### Threshold Configuration

- **Repeated issues**: ≥2 occurrences
- **Critical severity**: ≥3 occurrences
- **High correction rate**: >15% of turns

### Scope Support

- `scope: "project"`: Analyze all sessions (default)
- `scope: "session"`: Current session only

## Future Enhancements

Potential improvements (without Go code changes):

1. **Temporal Analysis**: Detect if issues cluster in time
2. **File Correlation**: Identify files with high bug correlation
3. **Git Integration**: Link failures to specific commits
4. **Trend Detection**: Compare current vs historical rates
5. **Custom Patterns**: User-defined failure patterns

All achievable through enhanced MCP queries and regex patterns.

## References

- [CLAUDE.md](../../CLAUDE.md) - Project constraints and principles
- [docs/examples-usage.md](../../docs/examples-usage.md) - Usage guide
- [.claude/commands/meta-errors.md](./meta-errors.md) - Tool error analysis (complementary)
