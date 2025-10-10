# ADR-001: Two-Layer Architecture Design

## Status

Accepted

## Context

The meta-cc project needs to analyze Claude Code session history to provide metacognitive insights and workflow optimization. The challenge is to design a system that can:

1. **Process large volumes of session data efficiently** - Session histories can contain thousands of tool calls, messages, and events
2. **Distinguish between data processing and semantic analysis** - Raw data extraction vs. intelligent interpretation
3. **Maintain performance and cost-effectiveness** - Avoid unnecessary LLM calls for simple data operations
4. **Support multiple integration patterns** - CLI tool, MCP server, slash commands, and subagents
5. **Enable autonomous querying** - Allow Claude to query session data during natural conversations

The key question was: **Should we build a monolithic system with LLM integration throughout, or separate concerns into distinct layers?**

## Decision

We adopt a **two-layer architecture** with clear separation of responsibilities:

### Layer 1: meta-cc CLI Tool (Pure Data Processing)

- **No LLM calls** - Pure rule-based analysis and data extraction
- **Responsibilities**:
  - Parse Claude Code session history (JSONL files from `~/.claude/projects/`)
  - Detect patterns using statistical and rule-based analysis
  - Output structured JSON/JSONL for consumption by Layer 2
- **Technologies**: Go, JSONL parsing, regex, statistical analysis

### Layer 2: Claude Code Integration (LLM-Powered)

- **LLM-driven semantic understanding** - Interpret data and generate recommendations
- **Responsibilities**:
  - MCP Server: Programmatic access for autonomous queries during conversations
  - Slash Commands: Quick analysis reports (`/meta-stats`, `/meta-errors`)
  - Subagents: Multi-turn conversational analysis (`@meta-coach`)
- **Technologies**: Claude Code integration, MCP protocol, natural language generation

### Data Flow

```
Session Data → CLI Tool → Structured JSON/JSONL → Claude Integration → Insights
   (JSONL)     (Layer 1)      (Data)             (Layer 2)          (User)
```

## Consequences

### Positive Impacts

1. **Clear Separation of Concerns**
   - Data extraction logic is testable without LLM
   - CLI tool can be used standalone or programmatically
   - Claude integration focuses purely on semantic analysis

2. **Cost Efficiency**
   - No LLM calls for simple data queries
   - Reduced token consumption (only structured data sent to LLM)
   - CLI tool can process large datasets without API costs

3. **Performance**
   - Fast data extraction without network latency
   - CLI tool can be cached, versioned, and distributed
   - Parallel processing possible at data layer

4. **Testability**
   - Layer 1: Unit tests for data extraction logic
   - Layer 2: Integration tests for LLM interactions
   - Clear interfaces between layers

5. **Flexibility**
   - CLI tool can be used in CI/CD pipelines
   - MCP server enables autonomous Claude queries
   - Multiple integration patterns supported

### Negative Impacts

1. **Increased System Complexity**
   - Two codebases to maintain (CLI tool + integration layer)
   - Need to ensure compatibility between layers
   - Documentation must cover both layers

2. **Interface Maintenance**
   - Data format changes require updates to both layers
   - Versioning strategy needed for breaking changes

3. **Learning Curve**
   - Users need to understand when to use CLI vs. MCP vs. slash commands
   - Documentation must explain the layering clearly

### Risks

1. **Data Format Changes**
   - Risk: Claude Code session format changes could break both layers
   - Mitigation: Version-aware parsing, schema validation, integration tests

2. **Performance Bottlenecks**
   - Risk: Large data transfers between layers
   - Mitigation: Hybrid output mode (inline vs. file_ref), streaming support

3. **Feature Duplication**
   - Risk: Temptation to add LLM features to CLI tool
   - Mitigation: Strict adherence to layer responsibilities, code review

## Implementation

### Completed

- [x] CLI tool with JSONL parsing (`internal/parser`)
- [x] Rule-based pattern analysis (`internal/analyzer`)
- [x] Query engine for filtering data (`internal/query`)
- [x] MCP server integration (`cmd/server.go`)
- [x] 14 MCP query tools (stats, tools, messages, files, etc.)
- [x] Slash commands (`/meta-stats`, `/meta-errors`, etc.)
- [x] Hybrid output mode (inline vs. file_ref)

### In Progress

- [ ] Subagents for multi-turn analysis
- [ ] Advanced workflow pattern detection
- [ ] Time-series analysis improvements

## Related Decisions

- [ADR-003](ADR-003-mcp-server-integration.md) - MCP Server Integration Strategy
- [ADR-004](ADR-004-hybrid-output-mode.md) - Hybrid Output Mode Design

## Notes

### Design Principle

> "The CLI tool extracts data; Claude extracts meaning."

This principle guides all architectural decisions:

- If it's **data extraction** → CLI tool (Layer 1)
- If it's **semantic understanding** → Claude integration (Layer 2)

### Example: Error Analysis

**Layer 1 (CLI Tool)**:
- Extract all tool calls with `status: "error"`
- Group by error type, count occurrences
- Output structured JSON: `{"error_type": "FileNotFound", "count": 15, "examples": [...]}`

**Layer 2 (Claude Integration)**:
- Interpret error patterns: "File not found errors suggest missing project initialization"
- Generate recommendations: "Run `npm install` before starting development"
- Provide contextual advice based on project type

### References

- [Implementation Plan](../plan.md)
- [Design Principles](../principles.md)
- [Technical Proposal](../proposals/meta-cognition-proposal.md)
