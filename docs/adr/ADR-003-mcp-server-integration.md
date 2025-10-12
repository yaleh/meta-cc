# ADR-003: MCP Server Integration Strategy

## Status

Accepted

## Context

The meta-cc project needs to provide Claude with programmatic access to session data. We need to decide how to integrate with Claude Code to enable:

1. **Autonomous querying** - Claude can query session data during natural conversations
2. **Multiple integration patterns** - Support different use cases and workflows
3. **Consistent user experience** - Unified approach across integration methods

### Integration Options Available

Claude Code supports three integration patterns:

1. **MCP Server** - Programmatic tool access, Claude calls autonomously
2. **Slash Commands** - User-triggered fixed reports (e.g., `/meta-stats`)
3. **Subagents** - Multi-turn conversational analysis (e.g., `@meta-coach`)

### Key Questions

- Which integration pattern should be primary?
- How should they complement each other?
- How to avoid duplication of functionality?
- How to ensure consistent behavior across patterns?

## Decision

We adopt a **three-tier integration strategy** with MCP as the foundation:

### Tier 1: MCP Server (Foundation Layer)

**Primary integration method** - Claude autonomously calls tools during conversation.

**Responsibilities**:
- Provide 14 query tools for session data access
- Support both inline and file_ref output modes
- Handle scope management (project vs. session)
- Execute queries with filtering and formatting

**Use Cases**:
- Natural language queries: "Show me all errors in this session"
- Conversational analysis: "What patterns do you see in my workflow?"
- Autonomous investigation: Claude queries data to answer questions
- **Coverage**: 80% of use cases

**Tools Provided**:
- `get_session_stats` - Session statistics
- `query_tools` - Filter tool calls
- `query_user_messages` - Search user messages
- `query_assistant_messages` - Search assistant messages
- `query_conversation` - Search conversations
- `query_files` - File operation stats
- `query_context` - Error context analysis
- `query_tool_sequences` - Workflow patterns
- `query_file_access` - File operation history
- `query_project_state` - Project evolution
- `query_successful_prompts` - High-quality prompts
- `query_tools_advanced` - SQL-like filtering
- `query_time_series` - Metrics over time
- `cleanup_temp_files` - Temp file management

### Tier 2: Slash Commands (Quick Reports)

**User-triggered fixed reports** - Predefined analysis workflows.

**Responsibilities**:
- Provide quick, standardized reports
- Call MCP tools under the hood
- Format results for readability

**Use Cases**:
- Repeated analysis workflows
- Quick health checks
- Standardized reports for documentation
- **Coverage**: 15% of use cases

**Commands Provided**:
- `/meta-stats` - Session statistics summary
- `/meta-errors` - Error analysis and patterns
- `/meta-habits` - Work pattern analysis
- `/meta-quality-scan` - Code quality assessment
- `/meta-timeline` - Project evolution timeline
- `/meta-tech-debt` - Technical debt tracking
- `/meta-bugs` - Bug pattern analysis

**Implementation**: Slash commands call MCP tools and format results.

### Tier 3: Subagents (Conversational Analysis)

**Multi-turn analysis** - Deep conversational exploration.

**Responsibilities**:
- Multi-turn conversation for complex analysis
- Interactive questioning and clarification
- Personalized recommendations

**Use Cases**:
- Exploratory workflow analysis
- Personalized coaching
- Complex problem investigation
- **Coverage**: 5% of use cases

**Subagents Provided**:
- `@meta-coach` - Workflow optimization coaching
- `@meta-architect` - Architecture analysis
- `@meta-focus-analyzer` - Attention pattern analysis

**Implementation**: Subagents call MCP tools and maintain conversation state.

### Integration Hierarchy

```
User Query
    ↓
Natural language? → MCP Server (80%)
    ↓
Quick report? → Slash Command → MCP Server (15%)
    ↓
Complex exploration? → Subagent → MCP Server (5%)
```

**Key Principle**: All integration methods use MCP as the data access layer.

## Consequences

### Positive Impacts

1. **Single Source of Truth**
   - MCP server is the only data access layer
   - Slash commands and subagents are thin wrappers
   - Consistency guaranteed across integration methods

2. **Autonomous Querying**
   - Claude can query data during natural conversations
   - No explicit commands needed
   - More natural user experience

3. **Flexibility**
   - Users can choose integration method based on use case
   - MCP for ad-hoc queries
   - Slash commands for repeated workflows
   - Subagents for exploration

4. **Maintainability**
   - Single codebase for data access (MCP server)
   - Slash commands and subagents are lightweight
   - Changes to data layer only affect MCP server

5. **Extensibility**
   - New slash commands are easy to add (call MCP tools)
   - New subagents are easy to add (call MCP tools)
   - New MCP tools benefit all integration methods

### Negative Impacts

1. **MCP Server Dependency**
   - Slash commands and subagents cannot work without MCP server
   - MCP server downtime affects all integration methods
   - Mitigation: MCP server is local (no network dependency)

2. **Potential Feature Duplication**
   - Risk of reimplementing MCP logic in slash commands
   - Mitigation: Code review, clear guidelines

3. **Learning Curve**
   - Users need to understand three integration methods
   - Mitigation: Clear documentation, usage examples

### Risks

1. **MCP Server Performance**
   - Risk: Slow MCP queries block all integration methods
   - Mitigation: Hybrid output mode, query optimization, caching

2. **Breaking Changes**
   - Risk: MCP tool changes break slash commands and subagents
   - Mitigation: Versioning, integration tests, changelog

3. **Over-Reliance on MCP**
   - Risk: All integration methods too tightly coupled to MCP
   - Mitigation: Clear interfaces, abstraction layers

## Implementation

### Completed

- [x] MCP server with 14 query tools (`cmd/server.go`)
- [x] Hybrid output mode (inline vs. file_ref)
- [x] Scope management (project vs. session)
- [x] Slash commands calling MCP tools
- [x] Integration guide documentation

### In Progress

- [ ] Subagent implementation
- [ ] Error handling and retry logic
- [ ] Performance optimization

### MCP Server Configuration

**Location**: `lib/server-config.json`

```json
{
  "mcpServers": {
    "meta-cc": {
      "command": "meta-cc",
      "args": ["server"],
      "env": {
        "META_CC_INLINE_THRESHOLD": "8192"
      }
    }
  }
}
```

**Environment Variables**:
- `META_CC_INLINE_THRESHOLD` - Threshold for inline vs. file_ref mode (default: 8192 bytes)

## Related Decisions

- [ADR-001](ADR-001-two-layer-architecture.md) - Two-Layer Architecture Design
- [ADR-004](ADR-004-hybrid-output-mode.md) - Hybrid Output Mode Design

## Notes

### Design Principle

> "MCP is the data layer. Slash commands and subagents are presentation layers."

This principle ensures:
- No data access logic in slash commands or subagents
- All data queries go through MCP server
- Consistency across integration methods

### Example: Error Analysis

**MCP Tool (`query_tools`)**:
```typescript
query_tools({
  status: "error",
  limit: 10
})
// Returns: [{tool: "Bash", error: "...", timestamp: "..."}]
```

**Slash Command (`/meta-errors`)**:
```markdown
# Error Analysis

Found 10 errors in this session:

1. **Bash** (3 occurrences)
   - "command not found: npm"
   - "permission denied"

2. **Read** (2 occurrences)
   - "file not found: config.json"
```

**Subagent (`@meta-coach`)**:
```
User: Help me understand why I'm getting so many errors

Agent: Let me analyze your error patterns...
[Calls query_tools internally]

I see you have 10 errors, mostly from Bash commands.
The root cause appears to be missing dependencies.

Would you like me to help you set up your project?
```

All three methods use the same MCP tool but present results differently.

### Integration Method Selection Guide

| Use Case | Integration Method | Rationale |
|----------|-------------------|-----------|
| "Show me all errors" | MCP (auto) | Natural language query |
| "What's my test coverage?" | MCP (auto) | Ad-hoc question |
| "Give me a quick health check" | `/meta-stats` | Repeated workflow |
| "Analyze my workflow patterns" | `@meta-coach` | Multi-turn exploration |
| "Why am I getting this error?" | MCP (auto) | Autonomous investigation |

### References

- [Integration Guide](../integration-guide.md)
- [MCP Guide](../mcp-guide.md)
- [Slash Commands](../../.claude/commands/)
- [Subagents](../../.claude/agents/)
