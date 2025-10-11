# Capability Development Guide

This guide explains how to create and extend meta-cc capabilities using the multi-source discovery system.

## Capability Structure

A capability is a markdown file with frontmatter metadata:

```markdown
---
name: my-capability
description: Short description of what this capability does.
keywords: keyword1, keyword2, keyword3
category: diagnostics
---

# Capability Implementation

Your capability implementation here...
Can include:
- MCP tool calls
- File operations
- Data analysis
- Visualization
```

### Frontmatter Fields

- **name**: Unique capability identifier (kebab-case, required)
- **description**: One-sentence description (required)
- **keywords**: Comma-separated keywords for semantic matching (required)
- **category**: Category for grouping (required)
  - Values: `diagnostics`, `assessment`, `visualization`, `analysis`, `guidance`

## Local Development Workflow

**Important**: meta-cc now defaults to loading capabilities from GitHub (`yaleh/meta-cc@main/commands`). For local development, you must explicitly configure a local source.

### 1. Create Capability Directory

```bash
mkdir -p ~/dev/my-capabilities
```

### 2. Configure Local Source

```bash
# REQUIRED for local development
export META_CC_CAPABILITY_SOURCES="~/dev/my-capabilities"

# Or combine with other sources (e.g., project's capability directory)
export META_CC_CAPABILITY_SOURCES="~/dev/my-capabilities:capabilities/commands"
```

**Why this is needed**: Without explicit configuration, meta-cc will load capabilities from GitHub, and your local changes won't be visible.

### 3. Create Capability File

```bash
cat > ~/dev/my-capabilities/my-feature.md <<EOF
---
name: my-feature
description: My custom feature analysis.
keywords: feature, custom, analysis
category: analysis
---

# My Feature

Implementation here...
EOF
```

### 4. Test Capability

```bash
# List capabilities (verify yours appears)
echo '{"method":"tools/call","params":{"name":"list_capabilities","arguments":{}}}' | meta-cc-mcp

# Get capability content
echo '{"method":"tools/call","params":{"name":"get_capability","arguments":{"name":"my-feature"}}}' | meta-cc-mcp

# Use via /meta command
/meta "my feature"
```

### 5. Iterate

- Edit capability file
- Changes reflect immediately (no cache for local sources)
- Test with `/meta` command

## Testing Against Branches

Test capabilities from different branches before merging:

```bash
# Test from develop branch
export META_CC_CAPABILITY_SOURCES="yaleh/meta-cc@develop/commands"
/meta "show errors"

# Test from feature branch
export META_CC_CAPABILITY_SOURCES="yaleh/meta-cc@feature/new-capability/commands"
/meta "new capability"

# Test from pull request commit
export META_CC_CAPABILITY_SOURCES="yaleh/meta-cc@abc123def/commands"
/meta "experimental"
```

**Cache Behavior**:
- Branches: 1-hour cache (changes propagate within 1 hour)
- To force refresh: Restart MCP server or wait for cache expiration
- Local sources: No cache (immediate reflection)

**Development Workflow**:
1. Create feature branch with new capability
2. Test via `META_CC_CAPABILITY_SOURCES="your-repo@feature-branch/commands"`
3. Iterate on the branch until satisfied
4. Submit pull request
5. After merge, capability available via main branch

## Publishing Capabilities

### Method 1: GitHub Repository

1. Create GitHub repo: `username/meta-cc-capabilities`
2. Add capabilities: `capabilities/my-feature.md`
3. Users install via:
   ```bash
   # Latest (main branch, 1-hour cache)
   export META_CC_CAPABILITY_SOURCES="username/meta-cc-capabilities/capabilities"

   # Stable version (tag, 7-day cache, recommended)
   export META_CC_CAPABILITY_SOURCES="username/meta-cc-capabilities@v1.0.0/capabilities"
   ```

**Recommendation**: Use semantic versioning tags (v1.0.0, v1.1.0, etc.) for stable releases. This enables:
- **7-day cache**: Faster loading, reduced CDN requests
- **Version pinning**: Users can opt into specific versions for stability
- **Immutability**: Tags don't change, ensuring consistency

**Release Workflow**:
```bash
# Create a release tag
git tag -a v1.0.0 -m "Release v1.0.0: Initial stable release"
git push origin v1.0.0

# Users can now pin to this version
export META_CC_CAPABILITY_SOURCES="username/meta-cc-capabilities@v1.0.0/capabilities"
```

### Method 2: Fork and PR

1. Fork `yaleh/meta-cc`
2. Add capability: `commands/meta-my-feature.md`
3. Submit PR
4. After merge, available to all users via default source

## Best Practices

1. **Clear frontmatter**: Accurate description and keywords
2. **Keywords**: Include synonyms and common variations
3. **Category**: Choose appropriate category for grouping
4. **Documentation**: Include usage examples in capability
5. **Testing**: Test with various natural language intents
6. **MCP tools**: Use existing MCP tools for data access
7. **Composition**: Design capabilities that can combine with others

## Example Capability

```markdown
---
name: meta-dependencies
description: Analyze project dependencies and detect security issues.
keywords: dependencies, npm, security, vulnerabilities, packages
category: assessment
---

# Dependency Analysis

This capability analyzes project dependencies for:
- Outdated packages
- Security vulnerabilities
- License issues
- Circular dependencies

## Implementation

1. **Detect package manager**:
   - Check for package.json (npm)
   - Check for go.mod (Go)
   - Check for requirements.txt (Python)

2. **Analyze dependencies**:
   ```
   Call mcp_meta_cc.query_tools(tool="Bash", pattern="npm|go|pip")
   ```

3. **Security scan**:
   - Run npm audit (if npm)
   - Check for known CVEs
   - Report vulnerabilities

4. **Recommendations**:
   - List outdated packages
   - Suggest security updates
   - Recommend version pinning
```

## Multi-Source Priority

When same-name capabilities exist in multiple sources, left-most source wins:

```bash
# Priority: my-dev > official
export META_CC_CAPABILITY_SOURCES="~/dev/test:.claude/commands"
```

**Use Cases**:
- Test capability changes before PR
- Override official capability with custom version
- Fork and customize capabilities

## Troubleshooting

### Capability Not Found

- Verify frontmatter is valid YAML
- Check filename matches frontmatter `name` field
- Verify source path in META_CC_CAPABILITY_SOURCES

### Semantic Matching Fails

- Add more keywords to frontmatter
- Use exact capability name: `/meta "meta-my-capability"`
- Check keyword spelling

### Cache Not Updating

- Local sources bypass cache automatically
- GitHub sources: wait 1 hour or restart meta-cc-mcp
- Check if META_CC_CAPABILITY_SOURCES includes local path

## Testing

Test your capability:

```bash
# Unit test: Parse frontmatter
sed -n '/^---$/,/^---$/p' my-capability.md | sed '1d;$d' | python3 -c "import sys, yaml; yaml.safe_load(sys.stdin)"

# Integration test: List capabilities
export META_CC_CAPABILITY_SOURCES="~/dev/my-caps"
echo '{"method":"tools/call","params":{"name":"list_capabilities","arguments":{}}}' | meta-cc-mcp | jq '.capabilities[] | select(.name=="my-capability")'

# Semantic matching test
/meta "my capability keywords"
# Should match your capability
```

## Advanced: Composite Capabilities

Design capabilities that can work together:

```markdown
---
name: meta-test-report
description: Generate comprehensive test report with visualization.
keywords: test, report, coverage, dashboard
category: assessment
---

# Test Report

1. **Analyze test results**:
   - Parse test output
   - Calculate coverage
   - Identify failures

2. **Generate visualization**:
   - Call meta-viz with test data
   - Create dashboard with charts

3. **Export report**:
   - Generate markdown summary
   - Include charts and metrics
```

## Integration with MCP Tools

Use existing MCP tools in your capabilities:

**Available Tools**:
- `get_session_stats()` - Session statistics
- `query_tools(tool, status, limit)` - Tool call analysis
- `query_user_messages(pattern, limit)` - Search user messages
- `query_assistant_messages(pattern, limit)` - Search assistant messages
- `query_conversation(pattern, limit)` - Search conversation
- `query_files(threshold)` - File operation stats
- `query_context(error_signature, window)` - Error context
- `query_tool_sequences(pattern, min_occurrences)` - Workflow patterns
- `query_file_access(file)` - File history
- `query_project_state()` - Project evolution
- `query_successful_prompts(min_quality_score, limit)` - High-quality prompts
- `query_tools_advanced(where, limit)` - SQL-like filtering
- `query_time_series(metric, interval, where)` - Time-based metrics
- `cleanup_temp_files(max_age_days)` - File cleanup

**Example**:
```markdown
# Error Analysis Capability

Call mcp_meta_cc.query_tools(status="error", limit=10) to get recent errors.
Call mcp_meta_cc.query_context(error_signature="<pattern>", window=3) for context.
```

## Community Guidelines

When contributing capabilities to the community:

1. **Clear naming**: Use descriptive, consistent names
2. **Documentation**: Include usage examples and expected output
3. **Testing**: Test with multiple projects and scenarios
4. **Dependencies**: Document any required tools or configuration
5. **License**: Include appropriate license information
6. **Maintenance**: Respond to issues and update as needed

## Resources

- [MCP Tools Reference](mcp-tools-reference.md) - Complete MCP tool documentation
- [Integration Guide](integration-guide.md) - Choosing MCP/Slash/Subagent
- [Official Slash Commands](.claude/commands/) - Examples of existing capabilities
- [Phase 22 Plan](../plans/22/plan.md) - Technical details of multi-source system
