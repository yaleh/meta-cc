# Query Cookbook

## Overview

This cookbook provides practical, ready-to-use query examples for common analysis scenarios. Each example includes:

- **Use case**: When to use this query
- **Query**: Complete code example
- **Output**: Expected result format
- **Analysis**: How to interpret results

---

## Table of Contents

1. [Error Analysis](#error-analysis)
2. [Tool Usage Patterns](#tool-usage-patterns)
3. [File Operations](#file-operations)
4. [Message Search](#message-search)
5. [Performance Analysis](#performance-analysis)
6. [Workflow Optimization](#workflow-optimization)
7. [Time-Based Analysis](#time-based-analysis)
8. [Session Debugging](#session-debugging)
9. [Quality Assessment](#quality-assessment)
10. [Advanced Composition](#advanced-composition)

---

## Error Analysis

### Example 1.1: Find All Errors by Tool

**Use case**: Identify which tools are failing most often

**Query**:
```javascript
query({
  resource: "tools",
  filter: {
    tool_status: "error"
  },
  aggregate: {
    function: "count",
    field: "tool_name"
  },
  output: {
    sort_by: "count",
    sort_order: "desc"
  }
})
```

**Output**:
```json
[
  {"tool_name": "Bash", "count": 23},
  {"tool_name": "Read", "count": 8},
  {"tool_name": "Edit", "count": 5}
]
```

**Analysis**:
- Bash has highest error rate (23 failures)
- Focus debugging efforts on Bash commands
- Check for common error patterns

---

### Example 1.2: Recent Errors Only

**Use case**: Debug errors from current session or recent work

**Query**:
```javascript
query({
  resource: "tools",
  filter: {
    tool_status: "error"
  },
  output: {
    sort_by: "timestamp",
    sort_order: "desc",
    limit: 10
  },
  scope: "session"
})
```

**Output**:
```json
[
  {
    "tool_name": "Bash",
    "timestamp": "2025-10-23T15:30:00Z",
    "error": "command not found: npm",
    "input": {"command": "npm test"}
  }
]
```

**Analysis**:
- Most recent error: npm command not found
- Likely environment issue (npm not in PATH)
- Check Node.js installation

---

### Example 1.3: Error Rate by Git Branch

**Use case**: Find which branches have most issues

**Query**:
```javascript
query({
  resource: "tools",
  transform: {
    group_by: "git_branch"
  },
  jq_filter: `
    group_by(.git_branch) |
    map({
      branch: .[0].git_branch,
      total: length,
      errors: map(select(.status == "error")) | length,
      error_rate: ((map(select(.status == "error")) | length) / length * 100 | round)
    }) |
    sort_by(.error_rate) |
    reverse
  `
})
```

**Output**:
```json
[
  {"branch": "feature/refactor", "total": 50, "errors": 10, "error_rate": 20},
  {"branch": "main", "total": 100, "errors": 5, "error_rate": 5},
  {"branch": "develop", "total": 30, "errors": 1, "error_rate": 3}
]
```

**Analysis**:
- feature/refactor has 20% error rate (highest)
- Suggests complexity or instability in refactor work
- Consider more frequent testing or smaller changes

---

## Tool Usage Patterns

### Example 2.1: Most Frequently Used Tools

**Use case**: Understand which tools you rely on most

**Query**:
```javascript
query({
  resource: "tools",
  aggregate: {
    function: "count",
    field: "tool_name"
  },
  output: {
    sort_by: "count",
    sort_order: "desc",
    limit: 10
  }
})
```

**Output**:
```json
[
  {"tool_name": "Bash", "count": 234},
  {"tool_name": "Read", "count": 156},
  {"tool_name": "Edit", "count": 89},
  {"tool_name": "Write", "count": 45}
]
```

**Analysis**:
- Bash is most used (234 calls)
- Heavy reliance on shell commands
- Consider creating Bash command shortcuts

---

### Example 2.2: Tool Usage Over Time

**Use case**: See how tool usage changes during development

**Query**:
```javascript
query({
  resource: "tools",
  transform: {
    extract: ["tool_name", "timestamp"]
  },
  jq_filter: `
    group_by(.timestamp[0:10]) |
    map({
      date: .[0].timestamp[0:10],
      tools: group_by(.tool_name) | map({tool: .[0].tool_name, count: length})
    })
  `
})
```

**Output**:
```json
[
  {
    "date": "2025-10-22",
    "tools": [
      {"tool": "Read", "count": 45},
      {"tool": "Bash", "count": 30}
    ]
  },
  {
    "date": "2025-10-23",
    "tools": [
      {"tool": "Edit", "count": 60},
      {"tool": "Bash", "count": 50}
    ]
  }
]
```

**Analysis**:
- Oct 22: Heavy reading (exploration phase)
- Oct 23: More editing (implementation phase)
- Natural progression from reading to writing

---

### Example 2.3: Unused Tools

**Use case**: Find tools you never or rarely use

**Query**:
```javascript
query({
  resource: "tools",
  aggregate: {
    function: "count",
    field: "tool_name"
  },
  jq_filter: `
    map(select(.count < 5)) |
    sort_by(.count)
  `
})
```

**Output**:
```json
[
  {"tool_name": "NotebookEdit", "count": 1},
  {"tool_name": "WebFetch", "count": 2},
  {"tool_name": "SlashCommand", "count": 3}
]
```

**Analysis**:
- NotebookEdit rarely used (consider if needed)
- WebFetch only 2 times (check if useful)
- Low usage may indicate unfamiliarity or limited use cases

---

## File Operations

### Example 3.1: Most Edited Files

**Use case**: Identify development hotspots

**Query**:
```javascript
query({
  resource: "tools",
  filter: {
    tool_name: "Edit"
  },
  jq_filter: `
    map(.input.file_path) |
    group_by(.) |
    map({file: .[0], count: length}) |
    sort_by(.count) |
    reverse |
    .[0:10]
  `
})
```

**Output**:
```json
[
  {"file": "/home/user/project/cmd/mcp-server/executor.go", "count": 23},
  {"file": "/home/user/project/internal/query/tools.go", "count": 18},
  {"file": "/home/user/project/README.md", "count": 12}
]
```

**Analysis**:
- executor.go is most edited (23 times)
- Suggests complex logic or frequent changes
- Consider refactoring if editing too often

---

### Example 3.2: File Read Errors

**Use case**: Find files that cause read errors (missing, permission issues)

**Query**:
```javascript
query({
  resource: "tools",
  filter: {
    tool_name: "Read",
    tool_status: "error"
  },
  jq_filter: `
    map({
      file: .input.file_path,
      error: .error,
      timestamp: .timestamp
    }) |
    group_by(.file) |
    map({
      file: .[0].file,
      error_count: length,
      errors: map(.error) | unique
    })
  `
})
```

**Output**:
```json
[
  {
    "file": "/tmp/missing-file.txt",
    "error_count": 5,
    "errors": ["file not found"]
  },
  {
    "file": "/root/.ssh/config",
    "error_count": 3,
    "errors": ["permission denied"]
  }
]
```

**Analysis**:
- /tmp/missing-file.txt: Repeatedly missing (check logic)
- /root/.ssh/config: Permission issues (run with correct user)

---

### Example 3.3: Files with Highest Error Rate

**Use case**: Find unreliable or problematic files

**Query**:
```javascript
query({
  resource: "tools",
  filter: {
    tool_name: "Read|Edit|Write"
  },
  jq_filter: `
    map({file: .input.file_path, status: .status}) |
    group_by(.file) |
    map({
      file: .[0].file,
      total: length,
      errors: map(select(.status == "error")) | length
    }) |
    map(. + {error_rate: (.errors / .total * 100 | round)}) |
    map(select(.error_rate > 0)) |
    sort_by(.error_rate) |
    reverse
  `
})
```

**Output**:
```json
[
  {"file": "/tmp/flaky.txt", "total": 10, "errors": 8, "error_rate": 80},
  {"file": "config.yaml", "total": 20, "errors": 5, "error_rate": 25}
]
```

**Analysis**:
- /tmp/flaky.txt: 80% error rate (very problematic)
- Likely temporary file with inconsistent existence
- Consider different approach or error handling

---

## Message Search

### Example 4.1: Find User Intent

**Use case**: What did I ask about earlier?

**Query**:
```javascript
query({
  resource: "messages",
  filter: {
    role: "user",
    content_match: "Phase 24|unified query"
  },
  output: {
    limit: 5
  },
  jq_filter: `
    map({
      turn: .turn_sequence,
      preview: .content[0:100],
      timestamp: .timestamp
    })
  `
})
```

**Output**:
```json
[
  {
    "turn": 45,
    "preview": "Let's start Phase 24 implementation. I want to create a unified query interface...",
    "timestamp": "2025-10-23T10:00:00Z"
  }
]
```

**Analysis**:
- Turn 45: User requested Phase 24 implementation
- Clear intent: Create unified query interface
- Use for context in follow-up work

---

### Example 4.2: Find Specific Discussions

**Use case**: Search for technical discussions on specific topics

**Query**:
```javascript
query({
  resource: "messages",
  filter: {
    content_match: "snake_case|schema|standardization"
  },
  jq_filter: `
    map({
      role: .role,
      turn: .turn_sequence,
      preview: .content[0:150]
    }) |
    .[0:10]
  `
})
```

**Output**:
```json
[
  {
    "role": "assistant",
    "turn": 50,
    "preview": "I'll standardize the schema to use snake_case throughout. This includes ToolCall struct with fields like tool_name, session_id..."
  }
]
```

**Analysis**:
- Turn 50: Schema standardization discussion
- Assistant explained snake_case conversion
- Reference for understanding design decisions

---

### Example 4.3: Count Messages by Type

**Use case**: Understand conversation balance (user vs assistant)

**Query**:
```javascript
query({
  resource: "messages",
  aggregate: {
    function: "count",
    field: "role"
  }
})
```

**Output**:
```json
[
  {"role": "user", "count": 45},
  {"role": "assistant", "count": 45}
]
```

**Analysis**:
- Balanced conversation (45 user, 45 assistant)
- Good back-and-forth interaction
- No one-sided conversation

---

## Performance Analysis

### Example 5.1: Slowest Tool Calls

**Use case**: Identify performance bottlenecks

**Query**:
```javascript
query({
  resource: "tools",
  jq_filter: `
    map(select(.duration)) |
    map({
      tool: .tool_name,
      duration: .duration,
      input: .input
    }) |
    sort_by(.duration) |
    reverse |
    .[0:10]
  `
})
```

**Output**:
```json
[
  {
    "tool": "Bash",
    "duration": 45000,
    "input": {"command": "npm test"}
  },
  {
    "tool": "Read",
    "duration": 3000,
    "input": {"file_path": "/large/file.json"}
  }
]
```

**Analysis**:
- npm test takes 45s (slowest)
- Large file read takes 3s
- Consider optimizing test suite or file processing

---

### Example 5.2: Average Duration by Tool

**Use case**: Compare tool performance

**Query**:
```javascript
query({
  resource: "tools",
  jq_filter: `
    map(select(.duration)) |
    group_by(.tool_name) |
    map({
      tool: .[0].tool_name,
      avg_duration: (map(.duration) | add / length | round),
      count: length
    }) |
    sort_by(.avg_duration) |
    reverse
  `
})
```

**Output**:
```json
[
  {"tool": "Bash", "avg_duration": 5000, "count": 234},
  {"tool": "Read", "avg_duration": 200, "count": 156},
  {"tool": "Edit", "avg_duration": 100, "count": 89}
]
```

**Analysis**:
- Bash averages 5s per call (slowest)
- Read and Edit much faster (200ms, 100ms)
- Bash commands are performance bottleneck

---

### Example 5.3: Tool Calls Above Threshold

**Use case**: Find unusually slow operations

**Query**:
```javascript
query({
  resource: "tools",
  jq_filter: `
    map(select(.duration > 10000)) |
    map({
      tool: .tool_name,
      duration: .duration,
      timestamp: .timestamp,
      input: .input
    })
  `
})
```

**Output**:
```json
[
  {
    "tool": "Bash",
    "duration": 45000,
    "timestamp": "2025-10-23T14:00:00Z",
    "input": {"command": "npm test"}
  }
]
```

**Analysis**:
- Only 1 call exceeded 10s threshold
- npm test at 14:00 took 45s
- Investigate why tests were slow at that time

---

## Workflow Optimization

### Example 6.1: Common Tool Sequences

**Use case**: Find repeated workflows to automate

**Query**:
```javascript
query({
  resource: "tools",
  jq_filter: `
    map({tool: .tool_name, turn: .turn_sequence}) |
    group_by(.turn) |
    map(map(.tool)) |
    map(. as $tools |
      if length > 1 then
        range(0; length-1) | $tools[.] + " -> " + $tools[.+1]
      else
        empty
      end
    ) |
    group_by(.) |
    map({sequence: .[0], count: length}) |
    map(select(.count >= 3)) |
    sort_by(.count) |
    reverse
  `
})
```

**Output**:
```json
[
  {"sequence": "Read -> Edit -> Bash", "count": 12},
  {"sequence": "Bash -> Read -> Edit", "count": 8},
  {"sequence": "Edit -> Bash -> Read", "count": 5}
]
```

**Analysis**:
- "Read → Edit → Bash" repeated 12 times
- Common pattern: Read file, edit, run command
- Consider creating slash command for this workflow

---

### Example 6.2: Session Productivity Metrics

**Use case**: Measure development velocity

**Query**:
```javascript
query({
  resource: "entries",
  scope: "session",
  jq_filter: `
    {
      total_turns: length,
      total_tools: map(select(.type == "tool_result")) | length,
      user_messages: map(select(.type == "user_message")) | length,
      assistant_messages: map(select(.type == "assistant_message")) | length,
      errors: map(select(.message.content[]? | select(.type == "tool_result" and .is_error))) | length
    } |
    . + {
      tools_per_turn: (.total_tools / .total_turns * 10 | round / 10),
      error_rate: (.errors / .total_tools * 100 | round)
    }
  `
})
```

**Output**:
```json
{
  "total_turns": 50,
  "total_tools": 123,
  "user_messages": 25,
  "assistant_messages": 25,
  "errors": 5,
  "tools_per_turn": 2.5,
  "error_rate": 4
}
```

**Analysis**:
- 50 turns with 123 tool calls (2.5 tools/turn)
- 4% error rate (acceptable)
- Balanced user/assistant interaction (25 each)
- Good productivity metrics

---

### Example 6.3: Error Recovery Patterns

**Use case**: How do you respond to errors?

**Query**:
```javascript
query({
  resource: "tools",
  jq_filter: `
    map(select(.status == "error")) |
    map({
      tool: .tool_name,
      turn: .turn_sequence,
      error: .error
    }) |
    group_by(.turn) |
    map({
      turn: .[0].turn,
      errors: map(.tool) | unique,
      error_count: length
    }) |
    map(select(.error_count > 1))
  `
})
```

**Output**:
```json
[
  {
    "turn": 45,
    "errors": ["Bash", "Bash"],
    "error_count": 2
  }
]
```

**Analysis**:
- Turn 45: 2 consecutive Bash errors
- Suggests retry attempts or debugging
- Pattern: Error → Retry → Success/Failure

---

## Time-Based Analysis

### Example 7.1: Activity by Hour

**Use case**: When are you most productive?

**Query**:
```javascript
query({
  resource: "entries",
  jq_filter: `
    map({hour: .timestamp[11:13], type: .type}) |
    group_by(.hour) |
    map({
      hour: .[0].hour,
      activity: length
    }) |
    sort_by(.hour)
  `
})
```

**Output**:
```json
[
  {"hour": "09", "activity": 15},
  {"hour": "10", "activity": 45},
  {"hour": "11", "activity": 60},
  {"hour": "14", "activity": 30}
]
```

**Analysis**:
- Peak productivity: 11am (60 activities)
- Morning ramp-up: 9am (15) → 11am (60)
- Post-lunch slowdown: 2pm (30)
- Schedule complex tasks for late morning

---

### Example 7.2: Daily Error Trends

**Use case**: Are errors increasing or decreasing?

**Query**:
```javascript
query({
  resource: "tools",
  jq_filter: `
    map({
      date: .timestamp[0:10],
      status: .status
    }) |
    group_by(.date) |
    map({
      date: .[0].date,
      total: length,
      errors: map(select(.status == "error")) | length
    }) |
    map(. + {error_rate: (.errors / .total * 100 | round)})
  `
})
```

**Output**:
```json
[
  {"date": "2025-10-21", "total": 100, "errors": 10, "error_rate": 10},
  {"date": "2025-10-22", "total": 120, "errors": 8, "error_rate": 7},
  {"date": "2025-10-23", "total": 150, "errors": 6, "error_rate": 4}
]
```

**Analysis**:
- Error rate improving: 10% → 4%
- More tool calls but fewer errors
- Suggests learning or code stabilization

---

### Example 7.3: Work Duration by Day

**Use case**: How long do you work each day?

**Query**:
```javascript
query({
  resource: "entries",
  jq_filter: `
    group_by(.timestamp[0:10]) |
    map({
      date: .[0].timestamp[0:10],
      first: .[0].timestamp,
      last: .[-1].timestamp,
      turns: length
    }) |
    map(. + {
      duration_hours: (
        (.last | fromdateiso8601) - (.first | fromdateiso8601)
      ) / 3600 | round
    })
  `
})
```

**Output**:
```json
[
  {"date": "2025-10-21", "first": "2025-10-21T09:00:00Z", "last": "2025-10-21T17:30:00Z", "turns": 100, "duration_hours": 9},
  {"date": "2025-10-22", "first": "2025-10-22T08:30:00Z", "last": "2025-10-22T16:00:00Z", "turns": 120, "duration_hours": 8}
]
```

**Analysis**:
- Oct 21: 9-hour day with 100 turns
- Oct 22: 8-hour day with 120 turns
- Higher turns/hour on Oct 22 (more productive)

---

## Session Debugging

### Example 8.1: Last 10 Actions

**Use case**: What did I just do?

**Query**:
```javascript
query({
  resource: "entries",
  scope: "session",
  output: {
    limit: 10,
    sort_by: "timestamp",
    sort_order: "desc"
  },
  jq_filter: `
    map({
      type: .type,
      timestamp: .timestamp,
      content: (
        if .type == "user_message" then
          .message.content[0].text[0:50]
        elif .type == "tool_result" then
          .message.content[0].tool_use_id
        else
          ""
        end
      )
    })
  `
})
```

**Output**:
```json
[
  {
    "type": "tool_result",
    "timestamp": "2025-10-23T15:35:00Z",
    "content": "tool-123"
  },
  {
    "type": "user_message",
    "timestamp": "2025-10-23T15:30:00Z",
    "content": "Please run the tests"
  }
]
```

**Analysis**:
- Last action: Tool result at 15:35
- Before that: User message at 15:30
- Currently waiting for next user input

---

### Example 8.2: Context Before Error

**Use case**: What led to this error?

**Query**:
```javascript
// First, find error turn
const errorTurn = query({
  resource: "tools",
  filter: {tool_status: "error"},
  output: {limit: 1, sort_by: "timestamp", sort_order: "desc"}
})[0].turn_sequence

// Then, get context window
query({
  resource: "entries",
  jq_filter: `
    map(select(.turn_sequence >= ${errorTurn - 3} and .turn_sequence <= ${errorTurn + 1}))
  `
})
```

**Output**:
```json
[
  {"turn_sequence": 42, "type": "user_message", "content": "Edit the config"},
  {"turn_sequence": 43, "type": "assistant_message", "content": "I'll edit config.yaml"},
  {"turn_sequence": 44, "type": "tool_use", "tool_name": "Edit"},
  {"turn_sequence": 45, "type": "tool_result", "status": "error", "error": "permission denied"}
]
```

**Analysis**:
- Turn 42: User requested edit
- Turn 43: Assistant planned edit
- Turn 44: Edit attempted
- Turn 45: Permission denied error
- Root cause: Insufficient file permissions

---

### Example 8.3: Current Session Summary

**Use case**: Quick overview of current session

**Query**:
```javascript
query({
  resource: "entries",
  scope: "session",
  jq_filter: `
    {
      session_id: .[0].session_id,
      git_branch: .[0].git_branch,
      start_time: .[0].timestamp,
      end_time: .[-1].timestamp,
      total_turns: length,
      messages: map(select(.type | contains("message"))) | length,
      tools: map(select(.type == "tool_result")) | length,
      errors: map(select(.message.content[]? | select(.type == "tool_result" and .is_error))) | length
    }
  `
})
```

**Output**:
```json
{
  "session_id": "session-abc123",
  "git_branch": "feature/unified-query",
  "start_time": "2025-10-23T10:00:00Z",
  "end_time": "2025-10-23T15:30:00Z",
  "total_turns": 50,
  "messages": 40,
  "tools": 123,
  "errors": 5
}
```

**Analysis**:
- 5.5-hour session on feature/unified-query branch
- 50 turns with 123 tool calls
- 5 errors (4% error rate)
- Active development session

---

## Quality Assessment

### Example 9.1: Code Quality Indicators

**Use case**: Assess code quality from patterns

**Query**:
```javascript
query({
  resource: "tools",
  jq_filter: `
    {
      test_runs: map(select(.input.command? | contains("test"))) | length,
      lint_runs: map(select(.input.command? | contains("lint"))) | length,
      build_runs: map(select(.input.command? | contains("build"))) | length,
      total_edits: map(select(.tool_name == "Edit")) | length,
      ratio_test_to_edit: (
        map(select(.input.command? | contains("test"))) | length
      ) / (
        map(select(.tool_name == "Edit")) | length
      ) * 100 | round
    }
  `
})
```

**Output**:
```json
{
  "test_runs": 45,
  "lint_runs": 30,
  "build_runs": 25,
  "total_edits": 89,
  "ratio_test_to_edit": 51
}
```

**Analysis**:
- Test runs: 45 (good testing discipline)
- Lint runs: 30 (code quality checks)
- 51% test-to-edit ratio (tests run after ~half of edits)
- Suggests TDD or frequent testing

---

### Example 9.2: Documentation Changes

**Use case**: Track documentation updates

**Query**:
```javascript
query({
  resource: "tools",
  filter: {
    tool_name: "Edit|Write"
  },
  jq_filter: `
    map(select(.input.file_path | test("\\.md$"))) |
    map({
      file: .input.file_path,
      timestamp: .timestamp
    }) |
    group_by(.file) |
    map({
      file: .[0].file,
      edits: length
    }) |
    sort_by(.edits) |
    reverse
  `
})
```

**Output**:
```json
[
  {"file": "README.md", "edits": 12},
  {"file": "docs/guides/unified-query-api.md", "edits": 8},
  {"file": "CHANGELOG.md", "edits": 5}
]
```

**Analysis**:
- README updated 12 times (frequently maintained)
- New guide created and edited 8 times
- CHANGELOG updated 5 times (good practice)
- Documentation kept in sync with code

---

### Example 9.3: Commit Pattern Analysis

**Use case**: Understand commit discipline

**Query**:
```javascript
query({
  resource: "tools",
  filter: {
    tool_name: "Bash"
  },
  jq_filter: `
    map(select(.input.command | contains("git commit"))) |
    map({
      timestamp: .timestamp,
      message: .input.command | match("git commit -m \"([^\"]+)\"").captures[0].string
    }) |
    {
      total_commits: length,
      commits: .[0:10]
    }
  `
})
```

**Output**:
```json
{
  "total_commits": 23,
  "commits": [
    {
      "timestamp": "2025-10-23T15:00:00Z",
      "message": "feat: add unified query API"
    },
    {
      "timestamp": "2025-10-23T14:00:00Z",
      "message": "test: add query parameter tests"
    }
  ]
}
```

**Analysis**:
- 23 commits in session (frequent commits)
- Commit messages follow convention (feat:, test:)
- Good atomic commit practice

---

## Advanced Composition

### Example 10.1: Multi-Stage Analysis

**Use case**: Complex analysis requiring multiple queries

**Query (Step 1 - Find hotspots)**:
```javascript
const hotspots = query({
  resource: "tools",
  filter: {tool_name: "Edit"},
  jq_filter: `
    map(.input.file_path) |
    group_by(.) |
    map({file: .[0], edits: length}) |
    map(select(.edits > 10)) |
    map(.file)
  `
})
// Result: ["/home/user/project/cmd/mcp.go", "/home/user/project/internal/query.go"]
```

**Query (Step 2 - Analyze hotspot errors)**:
```javascript
const hotspotErrors = hotspots.flatMap(file =>
  query({
    resource: "tools",
    filter: {
      tool_name: "Read|Edit|Write",
      tool_status: "error"
    },
    jq_filter: `
      map(select(.input.file_path == "${file}")) |
      map({
        file: .input.file_path,
        error: .error,
        timestamp: .timestamp
      })
    `
  })
)
```

**Output**:
```json
[
  {
    "file": "/home/user/project/cmd/mcp.go",
    "error": "syntax error line 45",
    "timestamp": "2025-10-23T14:00:00Z"
  }
]
```

**Analysis**:
- Hotspot files identified (>10 edits)
- Errors found in hotspot files
- mcp.go had syntax error (high churn area)

---

### Example 10.2: Comparative Analysis

**Use case**: Compare two time periods

**Query (Current week)**:
```javascript
const currentWeek = query({
  resource: "tools",
  filter: {
    time_range: {
      start: "2025-10-16T00:00:00Z",
      end: "2025-10-23T23:59:59Z"
    }
  },
  aggregate: {
    function: "count",
    field: "tool_name"
  }
})
```

**Query (Previous week)**:
```javascript
const previousWeek = query({
  resource: "tools",
  filter: {
    time_range: {
      start: "2025-10-09T00:00:00Z",
      end: "2025-10-15T23:59:59Z"
    }
  },
  aggregate: {
    function: "count",
    field: "tool_name"
  }
})
```

**Comparison**:
```javascript
const comparison = currentWeek.map(curr => {
  const prev = previousWeek.find(p => p.tool_name === curr.tool_name) || {count: 0}
  return {
    tool: curr.tool_name,
    current: curr.count,
    previous: prev.count,
    change: curr.count - prev.count,
    change_pct: Math.round((curr.count - prev.count) / prev.count * 100)
  }
})
```

**Output**:
```json
[
  {"tool": "Bash", "current": 250, "previous": 200, "change": 50, "change_pct": 25},
  {"tool": "Edit", "current": 100, "previous": 150, "change": -50, "change_pct": -33}
]
```

**Analysis**:
- Bash usage increased 25% (more command-line work)
- Edit usage decreased 33% (less file editing)
- Shift toward infrastructure/testing work

---

### Example 10.3: Workflow Efficiency Score

**Use case**: Calculate productivity metrics

**Query**:
```javascript
query({
  resource: "entries",
  scope: "session",
  jq_filter: `
    {
      tools: map(select(.type == "tool_result")),
      messages: map(select(.type | contains("message")))
    } |
    {
      total_tools: .tools | length,
      total_messages: .messages | length,
      errors: .tools | map(select(.message.content[]? | select(.is_error))) | length,
      test_runs: .tools | map(select(.message.content[].input.command? | contains("test"))) | length,
      edits: .tools | map(select(.message.content[].tool_name == "Edit")) | length
    } |
    . + {
      error_rate: (.errors / .total_tools * 100 | round),
      tools_per_message: (.total_tools / .total_messages * 10 | round / 10),
      test_to_edit_ratio: (
        if .edits > 0 then
          (.test_runs / .edits * 100 | round)
        else
          0
        end
      )
    } |
    . + {
      efficiency_score: (
        (100 - .error_rate) * 0.4 +
        (.tools_per_message * 10) * 0.3 +
        (.test_to_edit_ratio) * 0.3
      ) | round
    }
  `
})
```

**Output**:
```json
{
  "total_tools": 123,
  "total_messages": 50,
  "errors": 5,
  "test_runs": 45,
  "edits": 89,
  "error_rate": 4,
  "tools_per_message": 2.5,
  "test_to_edit_ratio": 51,
  "efficiency_score": 79
}
```

**Analysis**:
- Efficiency score: 79/100 (good)
- Components:
  - Low error rate (4%) = 38.4 points
  - Good tool usage (2.5/msg) = 7.5 points
  - Testing discipline (51%) = 15.3 points
- Areas to improve: Increase tools per message

---

## Tips and Best Practices

### 1. Start Simple, Add Complexity

Begin with basic filters, then add aggregation:
```javascript
// Start
query({resource: "tools", filter: {tool_name: "Bash"}})

// Add aggregation
query({
  resource: "tools",
  filter: {tool_name: "Bash"},
  aggregate: {function: "count", field: "status"}
})

// Add jq transform
query({
  resource: "tools",
  filter: {tool_name: "Bash"},
  aggregate: {function: "count", field: "status"},
  jq_filter: "map(select(.status == \"error\"))"
})
```

### 2. Use jq for Complex Logic

Structured query for filtering, jq for transformations:
```javascript
query({
  resource: "tools",
  filter: {tool_status: "error"},  // Efficient filtering
  jq_filter: "group_by(.tool_name) | ..."  // Complex transform
})
```

### 3. Leverage Output Control

Sort and limit for better UX:
```javascript
output: {
  sort_by: "timestamp",
  sort_order: "desc",
  limit: 10
}
```

### 4. Combine Queries for Insights

Use multiple queries for complex analysis (see Example 10.1-10.3)

### 5. Save Useful Queries

Create a personal query library for frequent analysis

---

## Related Documentation

- [Unified Query API Guide](../guides/unified-query-api.md) - Complete API reference
- [Migration Guide](../guides/migration-to-unified-query.md) - Migrate from old tools
- [MCP Guide](../guides/mcp.md) - MCP server documentation

---

## Contributing

Have a useful query pattern? Contribute to this cookbook:

1. Fork the repository
2. Add your example to this file
3. Submit a pull request

**Template for new examples**:
```markdown
### Example X.Y: [Title]

**Use case**: [When to use this query]

**Query**:
```javascript
query({...})
```

**Output**:
```json
[...]
```

**Analysis**:
- [Interpretation point 1]
- [Interpretation point 2]
```
