# Phase 12: MCP Project-Level Query Implementation

## Overview

This phase extends the MCP Server to support both project-level (all sessions) and session-level (current session) queries, enabling cross-session metacognitive analysis while maintaining backward compatibility.

## Quick Start

### Implementation Order

1. **Stage 12.1**: Project-level MCP tool definitions (TDD: tests first)
2. **Stage 12.2**: Execution logic with `--project .` flag support (TDD: tests first)
3. **Stage 12.3**: Session-level tools with `_session` suffix (TDD: tests first)
4. **Stage 12.4**: Configuration and documentation updates

### After Each Stage

```bash
# Stage 12.1: Test project-level tool definitions
go test ./internal/mcp -run TestProjectLevelTools
meta-cc mcp list-tools | grep -E "(query_tools|get_stats)" | grep -v "_session"

# Stage 12.2: Test --project flag execution
go test ./internal/mcp -run TestProjectExecution
# Simulate MCP call: query_tools (should use --project .)
./tests/mcp/test-project-flag.sh

# Stage 12.3: Test session-level tools
go test ./internal/mcp -run TestSessionLevelTools
meta-cc mcp list-tools | grep "_session"

# Stage 12.4: Validate configuration and docs
cat ~/.claude/mcp-servers/meta-cc.json | jq '.tools | length'  # Should show all tools
ls docs/mcp-guide.md
```

### After Phase Completion

```bash
# Run all unit tests
go test ./...

# Run integration tests
./tests/integration/mcp_project_scope_test.sh

# Verify project-level queries return multi-session data
meta-cc mcp call query_tools --args '{"limit": 10}'

# Verify session-level queries return current session only
meta-cc mcp call query_tools_session --args '{"limit": 10}'

# Verify backward compatibility
meta-cc mcp call get_session_stats  # Should still work unchanged
```

## Deliverables

### MCP Tools

**Project-Level (Default)**:
- `query_tools` - Query tool calls across all sessions
- `query_user_messages` - Search user messages across all sessions
- `get_stats` - Project-level statistics
- `analyze_errors` - Project-level error analysis
- `query_tool_sequences` - Workflow patterns across sessions
- `query_file_access` - File operation history (project-wide)
- `query_successful_prompts` - Quality prompts across sessions
- `query_context` - Error context analysis

**Session-Level (`_session` suffix)**:
- `query_tools_session` - Current session only
- `query_user_messages_session` - Current session only
- `get_session_stats` - Current session only (EXISTING - maintain compatibility)
- `analyze_errors_session` - Current session only
- `query_tool_sequences_session` - Current session only
- `query_file_access_session` - Current session only
- `query_successful_prompts_session` - Current session only
- `query_context_session` - Current session only

### Configuration

- Updated `.claude/mcp-servers/meta-cc.json` with all tools

### Documentation

- `docs/mcp-guide.md` - Usage guide with examples

### Testing

- Unit tests: `internal/mcp/*_test.go`
- Integration tests: `tests/integration/mcp_project_scope_test.sh`

## Code Budget

- Stage 12.1: Project-level tool definitions (~80 lines)
- Stage 12.2: Execution logic with `--project .` (~100 lines)
- Stage 12.3: Session-level tools (~80 lines)
- Stage 12.4: Configuration and documentation (~40 lines)
- **Total**: ~300 lines (within target)

## Success Criteria

- All unit tests pass (100%)
- Integration tests pass
- Project-level tools query all sessions (verified with multi-session data)
- Session-level tools query current session only
- `get_session_stats` maintains backward compatibility (unchanged behavior)
- Tool naming is consistent: no suffix = project-level, `_session` = session-level
- `--project .` flag correctly passed to CLI commands
- Documentation is clear and includes examples
- No regressions (all existing tests pass)

## Stage Checklist

- [ ] Stage 12.1: Project-level MCP tool definitions
  - [ ] Define `query_tools` (project-level)
  - [ ] Define `query_user_messages` (project-level)
  - [ ] Define `get_stats` (project-level)
  - [ ] Define `analyze_errors` (project-level)
  - [ ] Define remaining project-level tools
  - [ ] Unit tests pass

- [ ] Stage 12.2: Execution logic with `--project .`
  - [ ] Implement `executeTool()` for project-level queries
  - [ ] Add `--project .` flag to CLI invocations
  - [ ] Handle cross-session data aggregation
  - [ ] Unit tests pass

- [ ] Stage 12.3: Session-level tools
  - [ ] Define `query_tools_session`
  - [ ] Define `query_user_messages_session`
  - [ ] Verify `get_session_stats` unchanged
  - [ ] Define remaining `_session` tools
  - [ ] Unit tests pass

- [ ] Stage 12.4: Configuration and documentation
  - [ ] Update `.claude/mcp-servers/meta-cc.json`
  - [ ] Create `docs/mcp-guide.md`
  - [ ] Add usage examples
  - [ ] Update README.md with project-level query section

- [ ] Integration Testing
  - [ ] Project-level queries return multi-session data
  - [ ] Session-level queries return current session data
  - [ ] Backward compatibility verified
  - [ ] Real-world scenarios tested

- [ ] Documentation
  - [ ] README.md updated
  - [ ] Usage examples added
  - [ ] Integration guide updated

## Integration Points

Phase 12 builds on:
- **Phase 0-9**: Core CLI tool and query infrastructure
- **Phase 8**: MCP Server foundation

Phase 12 provides:
- **Project-level analysis**: Cross-session metacognitive insights
- **Session-level analysis**: Focused current session queries
- **Backward compatibility**: Existing `get_session_stats` unchanged
- **Clear API**: Consistent tool naming convention

## Implementation Status

**Phase 12 Status**: 📝 Planning Complete, Ready for Implementation

**Next Action**: Begin Stage 12.1 TDD - write tests for project-level MCP tool definitions
