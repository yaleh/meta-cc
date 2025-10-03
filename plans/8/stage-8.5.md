# Stage 8.5: Update Slash Commands for Phase 8

## Overview

**Objective**: Update existing Slash Commands to use Phase 8 query capabilities, avoiding context overflow and improving performance.

**Code Estimate**: ~50 lines (configuration changes, no Go code)

**Priority**: High (immediate improvement)

**Time Estimate**: 15-30 minutes

## Problem Statement

Current Slash Commands use old `parse extract` which:
- Loads ALL tool calls into memory at once
- Can cause context overflow in large sessions (>500 turns)
- Doesn't support filtering or limiting
- Less efficient than Phase 8 query commands

## Changes Required

### 1. Update `/meta-stats` Command

**File**: `.claude/commands/meta-stats.md`

**Current Implementation** (line 26):
```bash
meta-cc parse stats --output md
```

**Analysis**: Already optimal! `parse stats` performs aggregation and doesn't extract all data. No change needed.

**Action**: ✅ No change required (already using optimal command)

---

### 2. Update `/meta-timeline` Command

**File**: `.claude/commands/meta-timeline.md`

**Current Implementation** (line 33):
```bash
tools_data=$(meta-cc parse extract --type tools --output json)
```

**Problem**: Extracts ALL tool calls, no limit

**New Implementation**:
```bash
# Use Phase 8 query with limit
LIMIT=${1:-50}
tools_data=$(meta-cc query tools --limit "$LIMIT" --output json)
```

**Changes**:
1. Line 33: Replace `parse extract` with `query tools --limit`
2. Benefit: Only fetches the requested number of tools
3. Backward compatible: Still accepts limit parameter

**Detailed Diff**:
```diff
- # 提取工具调用数据
- tools_data=$(meta-cc parse extract --type tools --output json)
+ # 使用 Phase 8 query 命令（支持分页）
+ tools_data=$(meta-cc query tools --limit "$LIMIT" --output json)
```

---

### 3. Update `/meta-errors` Command (Optional Enhancement)

**File**: `.claude/commands/meta-errors.md`

**Current Implementation**: Uses `parse extract --filter` and `analyze errors`

**Potential Enhancement** (optional):
```bash
# Instead of:
error_data=$(meta-cc parse extract --type tools --filter "status=error" --output json)

# Use Phase 8 query (more efficient):
error_data=$(meta-cc query tools --status error --limit 100 --output json)
```

**Action**: Optional (analyze errors already works well)

---

## Implementation Steps

### Step 1: Backup Current Files
```bash
cp .claude/commands/meta-timeline.md .claude/commands/meta-timeline.md.backup
```

### Step 2: Update meta-timeline.md

**Change 1**: Update the tools data extraction (line 32-33)

Replace:
```bash
# 提取工具调用数据
tools_data=$(meta-cc parse extract --type tools --output json)
```

With:
```bash
# 使用 Phase 8 query 命令（支持分页，避免大会话上下文溢出）
tools_data=$(meta-cc query tools --limit "$LIMIT" --output json)
```

**Change 2**: Update the jq processing (line 36-41)

Replace:
```bash
echo "$tools_data" | jq -r --arg limit "$LIMIT" '
.[-($limit | tonumber):] |
to_entries[] |
"\(.key + 1). Turn \(.key) - **\(.value.ToolName)** \(if .value.Status == "error" then "❌" else "✅" end)"
'
```

With:
```bash
# query 命令已经限制了数量，直接使用结果
echo "$tools_data" | jq -r '
to_entries[] |
"\(.key + 1). **\(.value.ToolName)** \(if .value.Status == "error" or .value.Error != "" then "❌" else "✅" end)"
'
```

**Rationale**:
- `query tools --limit` already returns only the requested items
- No need to slice again with `.[-($limit | tonumber):]`
- Simplifies the jq logic

**Change 3**: Update statistics calculation (line 47-63)

Replace the second jq invocation that slices again:
```bash
echo "$tools_data" | jq -r --arg limit "$LIMIT" '
.[-($limit | tonumber):] |
{
  total: length,
  ...
'
```

With:
```bash
echo "$tools_data" | jq -r '
{
  total: length,
  errors: [.[] | select(.Status == "error" or .Error != "")] | length,
  tools: [.[] | .ToolName] | group_by(.) | map({tool: .[0], count: length}) | sort_by(.count) | reverse
} |
"- **总工具调用**: \(.total) 次",
"- **错误次数**: \(.errors) 次",
"- **错误率**: \(if .total > 0 then (.errors / .total * 100 | floor) else 0 end)%",
"",
"### Top 工具",
(.tools[:5] | .[] | "- \(.tool): \(.count) 次")
'
```

### Step 3: Update Documentation Comments

Add a note at the top of the command explaining the Phase 8 enhancement:

```markdown
---
name: meta-timeline
description: 生成当前会话的时间线视图，显示工具使用和错误的时序分布（Phase 8 增强：支持分页）
allowed_tools: [Bash]
argument-hint: [limit]
---

# meta-timeline：会话时间线视图

**Phase 8 增强**: 现在使用 `query tools` 命令，支持高效分页，避免大会话上下文溢出。

生成当前会话的时间线，可视化展示工具使用和错误分布。
```

---

## Testing Strategy

### Test 1: Small Session (< 50 tools)
```bash
/meta-timeline
# Expected: Shows all tool calls, no errors
```

### Test 2: Custom Limit
```bash
/meta-timeline 20
# Expected: Shows exactly 20 tool calls
```

### Test 3: Large Session (> 500 tools)
```bash
/meta-timeline 100
# Expected: Shows 100 tool calls, no context overflow
```

### Test 4: Error Detection
```bash
# In a session with errors
/meta-timeline
# Expected: Error analysis section shows detected errors
```

---

## Acceptance Criteria

- ✅ `/meta-timeline` uses `query tools --limit` instead of `parse extract`
- ✅ Default limit of 50 works correctly
- ✅ Custom limits work (e.g., `/meta-timeline 20`)
- ✅ No context overflow in large sessions
- ✅ Error detection still works
- ✅ Output format unchanged (backward compatible)
- ✅ Documentation updated with Phase 8 note
- ✅ `/meta-stats` verified to already be optimal (no change)

---

## Dependencies

- ✅ Stage 8.2 completed (`query tools` command available)
- ✅ `meta-cc` binary in PATH
- ✅ jq installed (for JSON processing)

---

## Rollback Plan

If issues occur:
```bash
# Restore from backup
cp .claude/commands/meta-timeline.md.backup .claude/commands/meta-timeline.md
```

---

## Benefits

### Performance
- ✅ Reduced memory usage (only load requested tools)
- ✅ Faster execution (less data processing)
- ✅ No context overflow risk

### User Experience
- ✅ Same familiar interface
- ✅ More reliable in large sessions
- ✅ Clearer documentation

### Maintainability
- ✅ Uses modern Phase 8 capabilities
- ✅ Simpler jq logic (no double slicing)
- ✅ Better separation of concerns

---

## Related Documentation

- Phase 8 Implementation Plan: `/plans/8/phase-8-implementation-plan.md`
- Integration Improvement Proposal: `/tmp/meta-cc-integration-improvement-proposal.md`
- Slash Commands Guide: `docs/examples-usage.md`
