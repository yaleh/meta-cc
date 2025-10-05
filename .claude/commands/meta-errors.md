---
name: meta-errors
description: Analyze error patterns in current session (Phase 14 - simplified query)
allowed_tools: [Bash]
---

# meta-errors: Error Pattern Analysis

Phase 14 Update: Uses new simplified `query errors` command. Aggregation done via jq.

Analyze errors in the session by extracting error list and aggregating patterns.

```bash
# Check if meta-cc is installed
if ! command -v meta-cc &> /dev/null; then
    echo "❌ Error: meta-cc not installed or not in PATH"
    echo ""
    echo "Please install meta-cc:"
    echo "  1. Download or build meta-cc binary"
    echo "  2. Place it in PATH (e.g., /usr/local/bin/meta-cc)"
    echo "  3. Ensure executable permissions: chmod +x /usr/local/bin/meta-cc"
    echo ""
    echo "Details: https://github.com/yale/meta-cc"
    exit 1
fi

# Step 1: Extract all errors using new query errors command
echo "## Error Data Extraction" >&2
echo "" >&2

# Phase 14: Use new query errors command (outputs JSONL)
ERRORS_JSONL=$(meta-cc query errors 2>/dev/null)
EXIT_CODE=$?

if [ $EXIT_CODE -eq 2 ]; then
    echo "✅ No errors detected in current session." >&2
    exit 0
elif [ $EXIT_CODE -eq 1 ]; then
    echo "❌ Error occurred during query." >&2
    exit 1
fi

# Convert JSONL to JSON array for jq processing
ERRORS_JSON=$(echo "$ERRORS_JSONL" | jq -s '.')
ERROR_COUNT=$(echo "$ERRORS_JSON" | jq 'length')
echo "Detected $ERROR_COUNT error tool call(s)." >&2
echo "" >&2

# Step 2: Aggregate error patterns using jq
echo "## Error Pattern Analysis"
echo ""

# Group by signature and count occurrences
PATTERNS=$(echo "$ERRORS_JSON" | jq 'if length > 0 then
    group_by(.signature) | map({
        signature: .[0].signature,
        tool_name: .[0].tool_name,
        count: length,
        first_seen: .[0].timestamp,
        last_seen: .[-1].timestamp,
        sample_error: .[0].error,
        time_span_seconds: ((.[- 1].timestamp | fromdateiso8601) - (.[0].timestamp | fromdateiso8601))
    }) | sort_by(-.count)
else
    []
end')

PATTERN_COUNT=$(echo "$PATTERNS" | jq '. | length')

if [ "$PATTERN_COUNT" -eq 0 ]; then
    echo "✅ No errors detected in current session."
    exit 0
fi

# Step 3: Format output as Markdown
echo "# Error Pattern Analysis"
echo ""
echo "Found $PATTERN_COUNT error pattern(s):"
echo ""

# Show patterns (limit to top 10 if many)
if [ "$PATTERN_COUNT" -gt 10 ]; then
    echo "⚠️  Large error set detected ($PATTERN_COUNT patterns)"
    echo "Showing top 10 patterns to prevent context overflow."
    echo ""

    echo "$PATTERNS" | jq -r '.[:10] | .[] |
        "\n## Pattern: \(.tool_name)\n" +
        "- **Signature**: `\(.signature)`\n" +
        "- **Occurrences**: \(.count) times\n" +
        "- **Error**: \(.sample_error)\n" +
        "\n### Context\n" +
        "- **First Occurrence**: \(.first_seen)\n" +
        "- **Last Occurrence**: \(.last_seen)\n" +
        "- **Time Span**: \(.time_span_seconds) seconds\n" +
        "\n---\n"'

    echo ""
    echo "💡 Tip: Use 'meta-cc query errors | jq' for custom analysis"
else
    echo "$PATTERNS" | jq -r '.[] |
        "\n## Pattern: \(.tool_name)\n" +
        "- **Signature**: `\(.signature)`\n" +
        "- **Occurrences**: \(.count) times\n" +
        "- **Error**: \(.sample_error)\n" +
        "\n### Context\n" +
        "- **First Occurrence**: \(.first_seen)\n" +
        "- **Last Occurrence**: \(.last_seen)\n" +
        "- **Time Span**: \(.time_span_seconds) seconds\n" +
        "\n---\n"'
fi

echo ""

# Step 4: Provide optimization suggestions
echo "---"
echo ""
echo "## Optimization Suggestions"
echo ""
echo "Based on detected error patterns, consider the following optimizations:"
echo ""
echo "1. **Investigate Repeated Errors**"
echo "   - Review error text to identify root causes"
echo "   - Check affected turns for context"
echo ""
echo "2. **Use Claude Code Hooks for Prevention**"
echo "   - Create pre-tool hooks to check error conditions"
echo "   - Example: file existence checks, permission validation"
echo ""
echo "3. **Adjust Workflow**"
echo "   - If errors concentrate in one tool, consider alternatives"
echo "   - Optimize prompts to reduce error frequency"
echo ""
echo "4. **View Full Error List**"
echo "   - Run: \`meta-cc query errors | jq\`"
echo "   - Analyze each error's specific cause and context"
echo ""

# Step 5: Show query examples
echo "## Advanced Query Examples"
echo ""
echo "**Last 50 errors:**"
echo "\`\`\`bash"
echo "meta-cc query errors | jq '.[-50:]'"
echo "\`\`\`"
echo ""
echo "**Errors by specific tool:**"
echo "\`\`\`bash"
echo "meta-cc query errors | jq '[.[] | select(.tool_name == \"Bash\")]'"
echo "\`\`\`"
echo ""
echo "**Count by tool:**"
echo "\`\`\`bash"
echo "meta-cc query errors | jq 'group_by(.tool_name) | map({tool: .[0].tool_name, count: length})'"
echo "\`\`\`"
```

## Output Content

1. **Error Data Extraction**: Count total errors in session
2. **Error Pattern Analysis**: Group errors by signature and show top patterns
3. **Optimization Suggestions**: Provide actionable improvement measures
4. **Advanced Query Examples**: Show how to use jq for custom analysis

## Migration from Phase 13

Phase 14 replaces `analyze errors --window` with `query errors` + jq:

**Old (Phase 13):**
```bash
meta-cc analyze errors --window 50
```

**New (Phase 14):**
```bash
meta-cc query errors | jq '.[-50:]'
```

**Aggregation (Phase 14):**
```bash
meta-cc query errors | jq 'group_by(.signature) | map({sig: .[0].signature, count: length})'
```

## Output Example

```markdown
## Error Pattern Analysis

# Error Pattern Analysis

Found 2 error pattern(s):

## Pattern: Bash

- **Signature**: `Bash:command not found: xyz`
- **Occurrences**: 5 times
- **Error**: command not found: xyz

### Context

- **First Occurrence**: 2025-10-05T10:00:00.000Z
- **Last Occurrence**: 2025-10-05T10:15:00.000Z
- **Time Span**: 900 seconds

---

## Optimization Suggestions

Based on detected error patterns, consider the following optimizations:

1. **Investigate Repeated Errors**
   - Review error text to identify root causes

2. **Use Claude Code Hooks for Prevention**
   - Create pre-tool hooks to check error conditions
```

## Use Cases

- Identify repeated errors to avoid re-debugging
- Discover workflow bottlenecks (operations failing frequently)
- Get optimization suggestions (hooks, alternatives, prompt improvements)
- Custom analysis using jq for advanced filtering

## Related Commands

- `/meta-stats`: View session statistics
- `meta-cc query errors`: Extract error list
- `meta-cc query tools --status error`: Query error tool calls
