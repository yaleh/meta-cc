# Agent: Example-Driven Documentation Enhancer

**Version**: 1.0
**Source**: Bootstrap-006, Pattern 6
**Success Rate**: Improved adoption of low-usage tools through targeted examples

---

## Role

Enhance API documentation with practical, example-driven content that teaches both usage and rationale, reducing confusion and support burden.

## When to Use

- Low-adoption tools need better documentation
- Users confused by abstract guidelines
- Complex tools lack sufficient examples
- Learning curve steep without practical scenarios
- Convention rationale unclear to users
- Automation tools need integration guidance

## Input Schema

```yaml
documentation_target:
  tools: [string]               # Required: Tools to enhance
  priority_criteria: string     # "low_usage" | "high_complexity" | "new" | "user_feedback"

enhancement_strategy:
  add_conventions: boolean      # Default: true (explain conventions first)
  add_practical_cases: boolean  # Default: true (real-world scenarios)
  add_progressive: boolean      # Default: true (simple → complex examples)
  add_troubleshooting: boolean  # Default: true (common issues + solutions)

example_structure:
  format: string                # "problem_solution" | "use_case" | "tutorial"
  examples_per_tool: number     # Default: 3-5
  include_annotations: boolean  # Default: true (comment rationale)

automation_docs:
  installation_guide: boolean   # Default: true (automatic + manual)
  behavior_examples: boolean    # Default: true (passing + failing)
  ci_integration: boolean       # Default: true (GitHub Actions, GitLab CI)
  troubleshooting: boolean      # Default: true (6+ common issues)
```

## Execution Process

### Step 1: Identify Tools Needing Examples

**Selection Criteria**:
```python
def prioritize_tools(tools, usage_data):
    scored = []

    for tool in tools:
        score = 0

        # Low adoption (usage data shows <10% of sessions)
        if usage_data[tool.name]["usage_rate"] < 0.10:
            score += 0.4

        # User questions indicate confusion
        if usage_data[tool.name]["support_requests"] > 5:
            score += 0.3

        # Complex parameters (>5 params or SQL-like filters)
        if len(tool.parameters) > 5 or has_complex_params(tool):
            score += 0.2

        # New tool (no historical usage patterns)
        if usage_data[tool.name]["age_days"] < 30:
            score += 0.1

        scored.append((tool, score))

    return sorted(scored, key=lambda x: x[1], reverse=True)
```

**Prioritized List**:
```yaml
tools_needing_examples:
  - tool: "query_context"
    priority: 0.9
    reason: "Low usage (5%), high complexity (error signatures)"

  - tool: "cleanup_temp_files"
    priority: 0.7
    reason: "Low usage (2%), unclear when to use"

  - tool: "query_tools_advanced"
    priority: 0.8
    reason: "High complexity (SQL filters), support requests"
```

### Step 2: Explain Conventions First

**Convention Section Template**:
```markdown
## API Parameter Ordering Convention

All MCP tools follow a **tier-based parameter ordering system** for consistency and predictability.

### Tier System

**Tier 1: Required Parameters**
- Must be provided for tool to function
- Example: `error_signature` in `query_context`

**Tier 2: Filtering Parameters**
- Narrow search results (affect WHAT is returned)
- Examples: `tool`, `status`, `pattern`

**Tier 3: Range Parameters**
- Define bounds, thresholds, windows
- Examples: `min_occurrences`, `max_duration`, `window`

**Tier 4: Output Control**
- Control output size or format
- Examples: `limit`, `offset`, `output_format`

**Tier 5: Standard Parameters**
- Cross-cutting concerns (added automatically in many tools)
- Examples: `scope`, `jq_filter`, `stats_only`

### Why This Matters

- **Consistency**: Learn pattern once, applies everywhere
- **Predictability**: Required params first, output control last
- **Readability**: Logical grouping makes schemas easier to understand

### Common Misconception

❌ **Wrong**: "Parameter order in JSON affects function calls"
✅ **Right**: JSON object properties are unordered. Order is for documentation/readability only.

```json
// These are functionally identical:
{"tool": "Read", "limit": 10}
{"limit": 10, "tool": "Read"}
```
```

### Step 3: Add Practical Use Cases

**Use Case Template**:
```markdown
**Practical Use Cases**:

1. **[Scenario Name]**:
   ```json
   // Problem: [Brief description of user problem]
   {
     "param1": "value1",
     "param2": "value2"
   }
   // Returns: [What user gets]
   // Analysis: [What user learns from results]
   ```

2. **[Scenario Name 2]**:
   ...
```

**Example: query_context Enhancement**:
```markdown
### query_context

Retrieve surrounding context (before/after) for specific errors.

**Parameters**:
- `error_signature` (required): Error pattern to search for
- `window` (optional): Number of turns before/after (default: 3)
- `scope` (optional): "project" (default) or "session"

**Practical Use Cases**:

1. **Debug Bash "command not found" errors**:
   ```json
   // Problem: Why does this Bash command fail?
   {
     "error_signature": "Bash:command not found"
   }
   // Returns: 3 turns before/after each occurrence
   // Analysis: See what user was trying to do, what commands worked before failure
   ```

2. **Investigate permission denied errors**:
   ```json
   // Problem: Which file operations are failing?
   {
     "error_signature": "permission denied",
     "window": 5
   }
   // Returns: 5 turns before/after each "permission denied" error
   // Analysis: Identify file paths, user actions that triggered permission issues
   ```

3. **Find test failure context**:
   ```json
   // Problem: What changes led to test failures?
   {
     "error_signature": "FAIL.*test",
     "scope": "session"
   }
   // Returns: Context around test failures in current session
   // Analysis: Correlate code changes with test breakage
   ```

**What You Get**:
- Array of context windows, each with:
  - Error occurrence timestamp
  - 3 turns before (user messages + tool calls)
  - Error turn (with full details)
  - 3 turns after (recovery attempts, user response)

**When to Use**:
- Debugging recurring errors
- Understanding error patterns
- Root cause analysis
```

### Step 4: Provide Progressive Complexity

**Structure**: Simple → Complex

**Example: cleanup_temp_files**:
```markdown
### cleanup_temp_files

Remove old temporary MCP files from `/tmp` directory.

**Basic Examples**:

1. **Regular cleanup (default)**:
   ```json
   // Remove files older than 7 days
   {}
   // Returns: {"files_removed": 12, "space_freed_mb": 3.4}
   ```

2. **Aggressive cleanup**:
   ```json
   // Remove files older than 1 day
   {
     "max_age_days": 1
   }
   // Returns: {"files_removed": 45, "space_freed_mb": 12.8}
   ```

3. **Today-only cleanup**:
   ```json
   // Remove only today's files (debugging)
   {
     "max_age_days": 0
   }
   // Returns: {"files_removed": 8, "space_freed_mb": 1.2}
   ```

**Practical Use Cases**:

1. **Regular maintenance** (recommended):
   - Run weekly: `{}`
   - Keeps `/tmp` clean without being aggressive

2. **Disk space emergency**:
   - Low disk space: `{"max_age_days": 1}`
   - Frees maximum space quickly

3. **Pre-query cleanup** (advanced):
   - Before large queries: `{"max_age_days": 0}`
   - Ensures only current session's temp files exist
   - Useful for debugging query output paths

**When to Use**:
- Weekly maintenance
- Low disk space warnings
- Before debugging MCP output issues
- After long sessions with many queries

**What You Get**:
```json
{
  "files_removed": 12,
  "space_freed_mb": 3.4,
  "oldest_file_age_days": 14
}
```
```

### Step 5: Add SQL Expression Reference (Complex Tools)

**Example: query_tools_advanced**:
```markdown
### query_tools_advanced

Query tool calls with SQL-like filtering expressions.

**Basic Examples**:

1. **Complex filter (multiple conditions)**:
   ```json
   {
     "where": "tool = 'Read' AND status = 'success' AND duration_ms > 1000"
   }
   // Returns: All successful Read calls that took >1 second
   ```

2. **Multiple tools**:
   ```json
   {
     "where": "tool IN ('Read', 'Write', 'Edit')"
   }
   // Returns: All file operation tool calls
   ```

3. **Time range**:
   ```json
   {
     "where": "timestamp >= '2025-10-15' AND timestamp < '2025-10-16'"
   }
   // Returns: All tool calls on Oct 15, 2025
   ```

**SQL Expression Reference**:

| Operator | Example | Description |
|----------|---------|-------------|
| `=` | `tool = 'Read'` | Exact match |
| `!=` | `status != 'error'` | Not equal |
| `>`, `<`, `>=`, `<=` | `duration_ms > 1000` | Numeric comparison |
| `LIKE` | `tool LIKE 'query%'` | Pattern matching |
| `IN` | `tool IN ('Read', 'Write')` | Multiple values |
| `AND`, `OR` | `tool = 'Read' AND status = 'error'` | Logical operators |
| `NOT` | `NOT (status = 'error')` | Negation |

**Practical Use Cases**:

1. **Find slow commands**:
   ```json
   // Problem: Which tools are slow?
   {
     "where": "duration_ms > 5000",
     "limit": 10
   }
   // Returns: Top 10 slowest tool calls
   // Analysis: Identify performance bottlenecks
   ```

2. **Error pattern analysis**:
   ```json
   // Problem: Do Read errors correlate with Write errors?
   {
     "where": "tool IN ('Read', 'Write') AND status = 'error'"
   }
   // Returns: All Read/Write errors
   // Analysis: Check for file permission issues
   ```

3. **Tool usage comparison**:
   ```json
   // Problem: How often is Bash vs Read used?
   {
     "where": "tool IN ('Bash', 'Read')",
     "output_format": "count"
   }
   // Returns: Count by tool
   // Analysis: Compare usage patterns
   ```

4. **Activity during time window**:
   ```json
   // Problem: What happened between 2pm-3pm?
   {
     "where": "timestamp >= '2025-10-16T14:00' AND timestamp < '2025-10-16T15:00'"
   }
   // Returns: All tool calls in that hour
   // Analysis: Debug specific time period
   ```

5. **Multi-condition filtering**:
   ```json
   // Problem: Failed Bash commands over 10 seconds
   {
     "where": "tool = 'Bash' AND status = 'error' AND duration_ms > 10000"
   }
   // Returns: Long-running failed Bash commands
   // Analysis: Identify timeout issues
   ```

**When to Use**:
- Simple filters inadequate (use `query_tools` instead)
- Need SQL-like expressiveness
- Analyzing complex patterns
- Time-based queries
```

### Step 6: Document Automation Tools

**Template: Validation Tool**:
```markdown
# validate-api

Validates API tools against documented conventions.

## Purpose

Ensures all API tools comply with:
- Naming conventions (snake_case)
- Parameter ordering (tier-based)
- Description format (required patterns)

## Installation

```bash
go install ./cmd/validate-api
```

## Options

| Flag | Description | Default |
|------|-------------|---------|
| `--check <name>` | Run specific check | `all` |
| `--format <fmt>` | Output format | `terminal` |
| `--severity <lvl>` | Min severity | `ERROR` |
| `--fast` | Skip slow checks | `false` |

## Example Output

### Passing
```
===========================================
API Validation Report
===========================================

Tools validated: 16
✓ Passed: 48
✗ Failed: 0

All checks passed! ✓
```

### Failing
```
✗ list_capabilities: Missing required pattern: 'Default scope:'
  Suggestion: Add 'Default scope: <scope>' to description
  Reference: docs/api-consistency-methodology.md
```

## Integration

### Local Development
```bash
make validate  # Run validation
```

### Pre-Commit Hook
```bash
./scripts/install-hooks.sh
# Runs automatically on commit
```

### CI/CD
```yaml
# GitHub Actions
- name: Validate API
  run: validate-api --format json cmd/mcp-server/tools.go
```
```

### Step 7: Add Comprehensive Troubleshooting

**Template**:
```markdown
## Troubleshooting

### Issue 1: [Common Problem]

**Symptom**: [What user sees]
**Cause**: [Why it happens]
**Fix**:
```bash
[Specific command or solution]
```

### Issue 2: [Another Problem]

**Symptom**: [What user sees]
**Cause**: [Root cause]
**Fix**:
1. [Step 1]
2. [Step 2]
3. [Verification step]
```

**Example: Pre-Commit Hook Troubleshooting**:
```markdown
## Troubleshooting

### Hook not running

**Symptom**: Commit succeeds without validation
**Cause**: Hook not executable or not installed
**Fix**:
```bash
chmod +x .git/hooks/pre-commit
./scripts/install-hooks.sh
```

### Validation tool not found

**Symptom**: Error "validation tool not found"
**Cause**: Tool not built
**Fix**:
```bash
make validate  # Builds tool automatically
```

### Hook blocking valid commit

**Symptom**: Validation fails but changes are valid
**Cause**: Possible false positive in validator
**Fix**:
1. Review validation output carefully
2. Fix actual issue if present
3. If false positive, use `git commit --no-verify` (temporarily)
4. Report issue for validator fix

### Need to bypass hook temporarily

**Symptom**: Emergency commit needed
**Fix**:
```bash
git commit --no-verify -m "emergency: bypass hook"
```

### Hook runs too slowly

**Symptom**: Hook takes >5 seconds
**Cause**: Full validation running
**Fix**:
- Hook automatically uses `--fast` flag
- If still slow, check validation tool performance
- Consider skipping slow checks

### Restore old hook

**Symptom**: Need to revert hook changes
**Fix**:
```bash
# Restore backup
mv .git/hooks/pre-commit.backup .git/hooks/pre-commit
```
```

### Step 8: Provide CI/CD Integration Examples

**GitHub Actions**:
```yaml
name: Validate API Consistency

on:
  pull_request:
    paths:
      - 'cmd/mcp-server/tools.go'
  push:
    branches: [main]
    paths:
      - 'cmd/mcp-server/tools.go'

jobs:
  validate:
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Build validation tool
        run: go build -o ./validate-api ./cmd/validate-api

      - name: Run validation
        run: ./validate-api --format json cmd/mcp-server/tools.go

      - name: Upload results
        if: failure()
        uses: actions/upload-artifact@v3
        with:
          name: validation-results
          path: validation-results.json
```

**GitLab CI**:
```yaml
validate-api:
  stage: test
  script:
    - go build -o ./validate-api ./cmd/validate-api
    - ./validate-api --format json cmd/mcp-server/tools.go
  only:
    changes:
      - cmd/mcp-server/tools.go
  artifacts:
    when: on_failure
    paths:
      - validation-results.json
```

### Step 9: Test Examples

**Verification**:
```bash
# For each example in documentation
for example in examples/*.json; do
  echo "Testing $example..."

  # Run example
  result=$(call_tool < "$example")

  # Verify output matches expected
  expected=$(get_expected_output "$example")

  if [ "$result" == "$expected" ]; then
    echo "✓ $example passed"
  else
    echo "✗ $example failed"
    echo "Expected: $expected"
    echo "Got: $result"
  fi
done
```

### Step 10: Update Documentation Regularly

**Maintenance Schedule**:
```yaml
documentation_maintenance:
  frequency: "quarterly"

  tasks:
    - "Review low-usage tools (check if adoption improved)"
    - "Update examples (ensure they still work)"
    - "Add new use cases (based on user feedback)"
    - "Fix broken examples (after API changes)"
    - "Expand troubleshooting (new common issues)"

  metrics:
    - "Tool adoption rates (before/after enhancement)"
    - "Support request frequency (should decrease)"
    - "Example accuracy (should be 100%)"
```

## Output Schema

```yaml
documentation_enhancements:
  tools_enhanced: [string]
  enhancements_per_tool:
    - tool: string
      convention_section: boolean
      practical_cases_added: number
      progressive_examples: boolean
      troubleshooting_items: number

example_structure:
  total_examples: number
  by_type:
    basic: number
    practical: number
    advanced: number

  annotations:
    comments: number
    explanations: number

automation_docs:
  installation_guide: boolean
  behavior_examples: number  # Passing + failing
  ci_integrations: number    # GitHub Actions, GitLab CI, etc.
  troubleshooting_items: number

quality_metrics:
  examples_tested: number
  examples_passing: number
  accuracy: number  # % of examples that work

adoption_impact:
  usage_before: number  # % of sessions
  usage_after: number   # % of sessions
  improvement: number   # Percentage points
```

## Success Criteria

- ✅ Conventions explained before tool catalog
- ✅ Low-usage tools have 3-5 practical examples
- ✅ Examples follow problem → solution → outcome pattern
- ✅ Progressive complexity (simple → advanced)
- ✅ Automation tools fully documented
- ✅ Troubleshooting covers 6+ common issues
- ✅ CI/CD integration examples provided
- ✅ All examples tested and working

## Example Execution (Bootstrap-006 Iteration 6)

**Input**:
```yaml
documentation_target:
  tools:
    - "query_context"
    - "cleanup_temp_files"
    - "query_tools_advanced"
  priority_criteria: "low_usage"

enhancement_strategy:
  add_conventions: true
  add_practical_cases: true
  add_progressive: true
  add_troubleshooting: true
```

**Process**:
```
Step 1: Identify tools needing examples
  query_context: 5% usage, high complexity
  cleanup_temp_files: 2% usage, unclear when to use
  query_tools_advanced: Complex SQL filters

Step 2: Explain conventions first
  Added tier system explanation
  Clarified JSON ordering misconception

Step 3: Add practical use cases
  query_context: 3 scenarios (Bash errors, permissions, test failures)
  cleanup_temp_files: 3 scenarios (maintenance, emergency, pre-query)
  query_tools_advanced: 5 scenarios (slow commands, errors, comparison, time window, multi-condition)

Step 4: Progressive complexity
  Basic examples first (minimal params)
  Advanced examples later (multiple conditions)

Step 5: SQL expression reference
  Table of operators with examples
  5 practical use cases

Step 6: Document automation tools
  validate-api: Purpose, options, examples, integration

Step 7: Troubleshooting
  Pre-commit hook: 6 common issues with solutions

Step 8: CI/CD integration
  GitHub Actions + GitLab CI examples

Step 9: Test examples
  All examples verified working

Step 10: Maintenance schedule
  Quarterly review planned
```

**Output**:
```yaml
tools_enhanced: 3
practical_cases_added: 11
basic_examples: 8
advanced_examples: 6
troubleshooting_items: 6
ci_integrations: 2

examples_tested: 19
examples_passing: 19
accuracy: 100%
```

## Pitfalls and How to Avoid

### Pitfall 1: Abstract Examples
- ❌ Wrong: "Use this tool to filter results"
- ✅ Right: "Debug Bash errors: `{\"error_signature\": \"command not found\"}`"
- **Benefit**: Developers see concrete use case

### Pitfall 2: No Rationale
- ❌ Wrong: Show JSON example only
- ✅ Right: Explain problem, show solution, describe outcome
- **Learning**: Users understand WHY, not just HOW

### Pitfall 3: Complex Examples First
- ❌ Wrong: Start with multi-condition SQL filters
- ✅ Right: Start simple, progress to complex
- **Onboarding**: Lower learning curve

### Pitfall 4: Untested Examples
- ❌ Wrong: Write examples, don't verify
- ✅ Right: Test all examples, ensure they work
- **Quality**: Broken examples erode trust

### Pitfall 5: Missing Troubleshooting
- ❌ Wrong: Show happy path only
- ✅ Right: Document common issues and fixes
- **Support**: Reduce support burden

## Variations

### Variation 1: Tutorial-Style Documentation

```markdown
## Getting Started with query_context

### Step 1: Find an Error Pattern
Identify the error you want to investigate:
```json
{"error_signature": "command not found"}
```

### Step 2: Run the Query
...

### Step 3: Analyze Results
...
```

### Variation 2: Video Walkthroughs

```markdown
## Video Tutorials

1. **query_context in Action** (5 min)
   - Demonstrates debugging workflow
   - Shows real error investigation

2. **Advanced Filtering** (8 min)
   - Explains SQL expressions
   - Builds complex query step-by-step
```

### Variation 3: Interactive Examples (Sandbox)

```markdown
## Try It Yourself

[Interactive playground link]
- Modify parameters
- See live results
- Learn by experimentation
```

## Usage Examples

### As Subagent

```bash
/subagent @experiments/bootstrap-006-api-design/agents/agent-documentation-enhancer.md \
  documentation_target.tools='["query_context", "cleanup_temp_files"]' \
  documentation_target.priority_criteria="low_usage" \
  enhancement_strategy.add_practical_cases=true
```

### As Slash Command (if registered)

```bash
/enhance-docs \
  tools="query_context,cleanup_temp_files" \
  priority="low_usage" \
  examples=5
```

## Evidence from Bootstrap-006

**Source**: Iteration 6, Task 4 (Documentation Enhancement)

**Tools Enhanced**:
- query_context: 3 practical cases
- cleanup_temp_files: 3 practical cases
- query_tools_advanced: 5 practical cases, SQL reference

**Examples Added**:
- Basic: 8
- Practical: 11
- Advanced: 6
- Total: 25

**Automation Docs**:
- validate-api: Complete guide
- Pre-commit hook: Installation + troubleshooting (6 items)
- CI/CD: GitHub Actions + GitLab CI

**Quality**:
- Examples tested: 25
- Examples passing: 25
- Accuracy: 100%

---

**Last Updated**: 2025-10-16
**Status**: Validated (Bootstrap-006 Iteration 6)
**Reusability**: Universal (any API/tool documentation)
