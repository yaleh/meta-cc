# Stage 8.6: Update @meta-coach Documentation

## Overview

**Objective**: Update @meta-coach Subagent documentation to include Phase 8 query capabilities and best practices.

**Code Estimate**: ~80 lines (documentation additions)

**Priority**: High (enables Subagent to use new features)

**Time Estimate**: 20-30 minutes

## Problem Statement

Current @meta-coach documentation:
- Only mentions old `parse extract` commands
- Doesn't include Phase 8 `query` commands
- Lacks guidance on iterative querying
- Missing best practices for avoiding context overflow

## Changes Required

### 1. Add Phase 8 Query Capabilities Section

**File**: `.claude/agents/meta-coach.md`

**Location**: After the "Analysis Tools" section (after line 62)

**New Section**:

```markdown
## Phase 8 Enhanced Query Capabilities

Phase 8 introduces powerful `query` commands for flexible data retrieval. Use these for efficient, targeted analysis.

### Query Tool Calls

**Basic Usage**:
```bash
# Query all tool calls (use with caution in large sessions)
meta-cc query tools --output json

# Query specific tool
meta-cc query tools --tool Bash --limit 20 --output json

# Query errors only
meta-cc query tools --status error --limit 10 --output json

# Complex filtering
meta-cc query tools --where "tool=Edit,status=error" --output json

# Sort by timestamp (newest first)
meta-cc query tools --sort-by timestamp --reverse --limit 30 --output json
```

**Key Benefits**:
- ✅ Pagination support (`--limit`, `--offset`)
- ✅ Efficient filtering (tool, status, complex conditions)
- ✅ Sorting capabilities
- ✅ Avoids context overflow in large sessions

### Query User Messages

**Basic Usage**:
```bash
# Search user messages with regex
meta-cc query user-messages --pattern "fix.*bug" --limit 10 --output json

# Find error-related messages
meta-cc query user-messages --pattern "error|fail|issue" --limit 20 --output json

# Sort by timestamp (newest first)
meta-cc query user-messages --sort-by timestamp --reverse --limit 5 --output json
```

**Use Cases**:
- Find when user mentioned specific topics
- Identify recurring concerns
- Track feature requests or bug reports
- Correlate user messages with error patterns

### Iterative Analysis Pattern (Recommended)

For large sessions (>500 turns), use an iterative approach to avoid context overflow:

**Step 1: Get Overview (Limited)**
```bash
# Get statistics first
stats=$(meta-cc parse stats --output json)

# Get recent tool usage (limited)
recent_tools=$(meta-cc query tools --limit 100 --sort-by timestamp --reverse --output json)
```

**Step 2: Identify Patterns**
```bash
# Analyze the limited dataset to find top tool
top_tool=$(echo "$recent_tools" | jq -r '
  [.[] | .ToolName] |
  group_by(.) |
  map({tool: .[0], count: length}) |
  sort_by(.count) |
  reverse |
  .[0].tool
')
```

**Step 3: Deep Dive (Targeted Query)**
```bash
# Now query ONLY that specific tool
meta-cc query tools --tool "$top_tool" --limit 50 --output json

# If errors found, query ONLY errors for that tool
meta-cc query tools --tool "$top_tool" --status error --output json
```

**Step 4: Iterate**
```bash
# Repeat for other interesting tools or patterns
# Each query is small and focused
```

**Why This Works**:
- 🎯 Each query fetches only relevant data
- 📉 Avoids loading entire session history
- 🔄 Allows progressive refinement
- 💡 Discovers insights step-by-step

### Best Practices for Query Commands

1. **Always Use `--limit` for Initial Exploration**
   ```bash
   # ✅ Good: Limited initial query
   meta-cc query tools --limit 50 --output json

   # ❌ Avoid: Unbounded query in large sessions
   meta-cc query tools --output json  # Can overflow context
   ```

2. **Prefer `query` Over `parse extract`**
   ```bash
   # ✅ Good: Filtered query with limit
   meta-cc query tools --tool Bash --limit 20 --output json

   # ❌ Old way: Extract all, then filter manually
   meta-cc parse extract --type tools --output json | jq '.[] | select(.ToolName == "Bash") | .[0:20]'
   ```

3. **Use Specific Filters to Reduce Data**
   ```bash
   # ✅ Good: Query only what you need
   meta-cc query tools --status error --tool Edit --limit 10

   # ❌ Avoid: Query all, filter later
   meta-cc query tools --limit 1000 | jq 'filter by tool and status'
   ```

4. **Leverage Sorting for Recent Analysis**
   ```bash
   # ✅ Good: Get most recent errors
   meta-cc query tools --status error --sort-by timestamp --reverse --limit 10
   ```

5. **Start Broad, Then Narrow**
   ```bash
   # Step 1: Overview (limited)
   meta-cc query tools --limit 100

   # Step 2: Identify issues (narrow filter)
   meta-cc query tools --status error --limit 20

   # Step 3: Deep dive (specific tool)
   meta-cc query tools --tool Bash --status error
   ```
```

---

### 2. Update Example Interactions

**Location**: Update "Example 2: Tool Usage Optimization" (around line 130)

**Add Phase 8 Variant**:

```markdown
### Example 2: Tool Usage Optimization (Phase 8 Enhanced)

```
Developer: "My sessions feel slower lately..."

Coach: Let me check your tool usage patterns using Phase 8 query capabilities.

[Runs: meta-cc parse stats --output json]

I can see you have 1,247 tool calls in this session. Let me analyze the recent patterns.

[Runs: meta-cc query tools --limit 200 --sort-by timestamp --reverse --output json]

Interesting findings from the last 200 tool calls:
- You're using the Read tool 42% of the time (84 out of 200)
- You're using Bash 28% of the time (56 out of 200)
- Most Read calls are for files in the same directory

Let me check if there's a specific file pattern:

[Runs: meta-cc query tools --tool Read --limit 50 --output json]

I found a pattern:
1. You Read the same 3 files repeatedly: `auth.js`, `utils.js`, `config.js`
2. You often Read → Edit → Read the same file (verification pattern)
3. The Read-Edit-Read cycle happens 12 times in the last 50 Read operations

Some optimization suggestions:
1. **Reduce verification Reads**: Trust your edits more, or use `git diff` to verify
2. **Batch related edits**: Edit multiple related files before verifying
3. **Use Grep for searching**: Instead of Read → search manually, use Grep with context

Would you like me to create a custom Slash Command that:
- Reads all 3 config files at once?
- Shows a summary instead of full content?
- Combines Read + Grep for common searches?
```
```

---

### 3. Add New Coaching Scenario

**Location**: After Example 3 (around line 169)

**Add**:

```markdown
### Example 4: Large Session Analysis (Phase 8 Pattern)

```
Developer: "This session has over 2000 turns, and I'm getting context overflow errors..."

Coach: Large sessions need an iterative analysis approach. Let me demonstrate.

[Step 1: Get high-level stats]
meta-cc parse stats --output json

I see:
- 2,347 turns total
- 1,892 tool calls
- 47 errors (2.5% error rate)

This is too large to analyze all at once. Let me use Phase 8 queries to drill down.

[Step 2: Get recent tools only]
meta-cc query tools --limit 100 --sort-by timestamp --reverse --output json

In the last 100 tool calls:
- Bash: 38 times
- Edit: 27 times
- Read: 23 times
- 5 errors detected (all Bash)

[Step 3: Focus on the errors]
meta-cc query tools --tool Bash --status error --limit 20 --output json

All 5 Bash errors are the same: `npm test` failing on `auth.test.js`.

[Step 4: Find when this started]
meta-cc query user-messages --pattern "auth|test" --limit 10 --output json

I found that you first mentioned "auth test failing" about 90 minutes ago.
You've been trying different approaches since then.

**Recommendation**:
Instead of re-running the same test, let's:
1. Focus on understanding the actual error in `auth.test.js`
2. Use Read to examine the test file
3. Check recent changes to auth module
4. Stop the test-retry loop

This iterative approach:
- ✅ Analyzed a 2000+ turn session without overflow
- ✅ Found the core issue in 4 targeted queries
- ✅ Each query fetched < 100 items
- ✅ Total context: ~400 items vs 1892 (79% reduction)
```
```

---

## Implementation Steps

### Step 1: Backup Current File
```bash
cp .claude/agents/meta-coach.md .claude/agents/meta-coach.md.backup
```

### Step 2: Insert Phase 8 Section

**Location**: After line 62 (after "Cross-Project Analysis" section)

**Action**: Add the entire "Phase 8 Enhanced Query Capabilities" section

### Step 3: Update Example 2

**Location**: Line 130-147

**Action**: Add the "Phase 8 Enhanced" variant after the existing example

### Step 4: Add Example 4

**Location**: After line 169 (after Example 3)

**Action**: Insert the new "Large Session Analysis" example

### Step 5: Update Best Practices Section

**Location**: Line 170-177 (Best Practices)

**Action**: Add Phase 8 best practice:

```markdown
## Best Practices

1. **Be Data-Driven**: Always base insights on actual session data, not assumptions
2. **Encourage Discovery**: Guide developers to their own insights rather than prescribing solutions
3. **Respect Context**: Understand that each developer's workflow is unique
4. **Iterate and Adapt**: Treat optimization as an ongoing process
5. **Celebrate Progress**: Acknowledge improvements and learning
6. **Use Phase 8 Iterative Pattern**: For large sessions (>500 turns), use targeted queries with limits to avoid context overflow
```

---

## Testing Strategy

### Test 1: Verify Phase 8 Commands in Subagent
```
@meta-coach 分析我的工作流
```

**Expected**: @meta-coach uses `query` commands instead of `parse extract`

### Test 2: Large Session Handling
```
@meta-coach 这个会话很大，帮我找出主要问题
```

**Expected**: Uses iterative pattern with `--limit` flags

### Test 3: Specific Tool Analysis
```
@meta-coach 我的 Bash 命令总是失败，帮我分析
```

**Expected**: Uses `query tools --tool Bash --status error`

---

## Acceptance Criteria

- ✅ Phase 8 query section added to documentation
- ✅ Iterative analysis pattern documented with examples
- ✅ Best practices updated with Phase 8 guidance
- ✅ Example interactions show Phase 8 usage
- ✅ @meta-coach can use `query tools` and `query user-messages`
- ✅ Large session handling explained
- ✅ Backward compatible (old commands still mentioned)

---

## Dependencies

- ✅ Stage 8.2 completed (`query tools` available)
- ✅ Stage 8.3 completed (`query user-messages` available)
- ✅ `meta-cc` binary in PATH

---

## Benefits

### For @meta-coach Subagent
- ✅ Can handle large sessions without context overflow
- ✅ More efficient data retrieval
- ✅ Better coaching based on targeted queries
- ✅ Demonstrates best practices to users

### For Users
- ✅ Learn Phase 8 capabilities through coaching
- ✅ See iterative analysis pattern in action
- ✅ Get more relevant insights (filtered data)
- ✅ Better performance in large sessions

### For Project
- ✅ Showcases Phase 8 value
- ✅ Provides real-world usage examples
- ✅ Documents best practices
- ✅ Improves Subagent effectiveness

---

## Related Documentation

- Phase 8 Implementation Plan: `/plans/8/phase-8-implementation-plan.md`
- Stage 8.2 (Query Tools): `/plans/8/phase-8-implementation-plan.md#stage-82`
- Stage 8.3 (Query Messages): `/plans/8/phase-8-implementation-plan.md#stage-83`
- Integration Improvement Proposal: `/tmp/meta-cc-integration-improvement-proposal.md`
