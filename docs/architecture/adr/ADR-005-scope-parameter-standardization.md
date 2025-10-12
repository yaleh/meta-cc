# ADR-005: Scope Parameter Standardization

## Status

Accepted

## Context

The MCP server provides 14 query tools that can analyze session data. Each tool needs to support two query scopes:

1. **Project Scope** - Analyze all sessions in the current project
2. **Session Scope** - Analyze only the current session

### Problem Statement

**Inconsistent Defaults**:
- Some tools default to `project` scope (e.g., `query_tools`)
- Some tools default to `session` scope (e.g., `get_session_stats`)
- Users are confused about which scope is used
- Documentation is inconsistent

**User Expectations**:
- Natural language queries like "Show me all errors" - Should this be project or session?
- Command names suggest scope (e.g., `get_session_stats` → session scope)
- But most analysis benefits from project-wide view

**Performance Considerations**:
- Project scope: Slower, more data, but more insights
- Session scope: Faster, less data, but limited view
- Default should optimize for common case

### Requirements

1. **Consistent defaults** across all tools
2. **Intuitive behavior** matching user expectations
3. **Explicit override** when needed
4. **Performance optimization** for common case
5. **Clear documentation** of scope behavior

## Decision

We standardize on **project scope as the default** for all query tools, with explicit override option.

### Default Scope: Project

**All query tools default to `scope: "project"`**:

- `get_session_stats` → **project** (was session)
- `query_tools` → project (unchanged)
- `query_user_messages` → project (unchanged)
- `query_assistant_messages` → project (unchanged)
- `query_conversation` → project (unchanged)
- `query_files` → project (unchanged)
- `query_context` → project (unchanged)
- `query_tool_sequences` → project (unchanged)
- `query_file_access` → project (unchanged)
- `query_project_state` → project (unchanged)
- `query_successful_prompts` → project (unchanged)
- `query_tools_advanced` → project (unchanged)
- `query_time_series` → project (unchanged)

**Rationale**:
1. **More insights** - Project-wide analysis reveals patterns across sessions
2. **Consistent behavior** - All tools behave the same way
3. **Matches user expectations** - "Show me errors" usually means "all errors in this project"
4. **Better recommendations** - Historical data improves suggestions
5. **Explicit is better** - Users must opt into session scope

### Explicit Override

**Scope Parameter**:
```typescript
// Default: project scope
query_tools({ status: "error" })

// Explicit: session scope only
query_tools({ status: "error", scope: "session" })

// Explicit: project scope (redundant but clear)
query_tools({ status: "error", scope: "project" })
```

**Valid Values**:
- `"project"` - Analyze all sessions in current project (default)
- `"session"` - Analyze only current session

### Tool Name Clarification

**`get_session_stats`**:
- Despite the name, defaults to **project scope**
- Returns statistics for all sessions in project
- Name kept for backward compatibility
- Documentation clarifies actual behavior

**Alternative Considered**: Rename to `get_project_stats`
- **Rejected**: Breaking change, existing documentation references it
- **Mitigation**: Clear documentation, deprecation notice

## Consequences

### Positive Impacts

1. **Consistent Defaults**
   - All tools default to project scope
   - Predictable behavior
   - Less cognitive load for users

2. **Better Insights**
   - Project-wide analysis reveals patterns
   - Historical data improves recommendations
   - More valuable default behavior

3. **Explicit Intent**
   - Users must explicitly request session scope
   - Reduces accidental narrow analysis
   - Clear documentation of scope behavior

4. **Matches Natural Language**
   - "Show me errors" → project scope (intuitive)
   - "Show me errors in this session" → session scope (explicit)

5. **Future-Proof**
   - Easy to add more scopes (e.g., `workspace`, `global`)
   - Consistent parameter naming
   - Extensible design

### Negative Impacts

1. **Performance Impact**
   - Project scope is slower than session scope
   - More data to process and transfer
   - Mitigation: Hybrid output mode, query optimization, caching

2. **Tool Name Confusion**
   - `get_session_stats` defaults to project scope
   - Name suggests session scope
   - Mitigation: Documentation, deprecation notice, consider rename in v2.0

3. **Breaking Change**
   - `get_session_stats` behavior changes
   - Existing code relying on session default will break
   - Mitigation: Version bump, migration guide, backward compatibility flag

### Risks

1. **User Confusion**
   - Risk: Users expect `get_session_stats` to return session scope
   - Mitigation: Clear documentation, console warnings, examples

2. **Performance Degradation**
   - Risk: Default project scope is too slow for large projects
   - Mitigation: Hybrid output mode, query optimization, scope override

3. **Breaking Existing Workflows**
   - Risk: Slash commands and subagents rely on old defaults
   - Mitigation: Update all integration code, integration tests

## Implementation

### Completed

- [x] Update MCP tool parameter parsing
- [x] Default all tools to `scope: "project"`
- [x] Add scope override parameter
- [x] Update tool descriptions
- [x] Update integration guide documentation
- [x] Update MCP output modes documentation

### In Progress

- [ ] Update slash commands to handle new default
- [ ] Update subagents to handle new default
- [ ] Add deprecation notice for `get_session_stats` name
- [ ] Add console warnings for scope clarification

### Code Changes

**Before**:
```go
// Different defaults across tools
func GetSessionStats() {
    scope := "session"  // Session scope default
}

func QueryTools() {
    scope := "project"  // Project scope default
}
```

**After**:
```go
// Consistent defaults across all tools
func GetSessionStats(scope string) {
    if scope == "" {
        scope = "project"  // Project scope default
    }
}

func QueryTools(scope string) {
    if scope == "" {
        scope = "project"  // Project scope default
    }
}
```

### Migration Guide

**For Users**:
```typescript
// Old behavior (get_session_stats defaulted to session)
get_session_stats()  // Returned current session only

// New behavior (defaults to project)
get_session_stats()  // Returns all sessions in project
get_session_stats({ scope: "session" })  // Explicit session scope
```

**For Integration Code**:
```javascript
// Old slash command
const stats = await mcp.get_session_stats();  // Session scope

// New slash command
const stats = await mcp.get_session_stats({ scope: "session" });  // Explicit
```

## Related Decisions

- [ADR-001](ADR-001-two-layer-architecture.md) - Two-Layer Architecture Design
- [ADR-003](ADR-003-mcp-server-integration.md) - MCP Server Integration Strategy
- [ADR-004](ADR-004-hybrid-output-mode.md) - Hybrid Output Mode Design

## Notes

### Design Rationale

The key insight is that **meta-cognition requires historical context**:

- Workflow patterns emerge across sessions
- Error trends reveal systemic issues
- Quality metrics improve with more data
- Personalized recommendations need history

**Session scope** is useful for:
- Debugging current session issues
- Quick health checks
- Performance analysis of current work

**Project scope** is useful for:
- Workflow optimization
- Pattern detection
- Long-term trend analysis
- Comprehensive recommendations

Since **project scope provides more value** in 80% of use cases, it should be the default.

### Scope Parameter Philosophy

We follow the **"Explicit is better than implicit"** principle from Python's Zen:

- Default to the **most useful** option (project scope)
- Require explicit override for **narrow** scope (session)
- Clear documentation of defaults
- Consistent naming across all tools

This differs from the **"Principle of least surprise"** which might suggest `get_session_stats` should default to session scope. We prioritize **usefulness over naming consistency**.

**Alternative Considered**: Different defaults per tool
- `get_session_stats` → session scope (matches name)
- `query_tools` → project scope (matches behavior)
- **Rejected**: Inconsistent, confusing, error-prone

### Future Enhancements

**Potential Scope Values**:
- `"session"` - Current session only
- `"project"` - All sessions in current project (current default)
- `"workspace"` - All projects in workspace (future)
- `"global"` - All projects across workspaces (future)
- `"last_n_sessions"` - Last N sessions (future)

**Backward Compatibility**:
- Add `scope_version` parameter to opt into new defaults
- Maintain old behavior with `scope_version: "1.0"`
- New behavior with `scope_version: "2.0"`

### References

- [Integration Guide](../integration-guide.md)
- [MCP Guide](../mcp-guide.md)
- [Tool Descriptions](../../cmd/server.go)
