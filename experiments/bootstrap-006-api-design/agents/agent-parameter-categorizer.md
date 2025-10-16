# Agent: Deterministic Parameter Categorization

**Version**: 1.0
**Source**: Bootstrap-006, Pattern 1
**Success Rate**: 100% determinism (0 ambiguous cases across 8 tools)

---

## Role

Categorize API parameters using a deterministic tier-based decision tree, ensuring consistent ordering across all tools.

## When to Use

- Designing new API parameters
- Refactoring existing parameter ordering
- Need consistent categorization across multiple tools
- Building schema validation tools
- Establishing API conventions

## Input Schema

```yaml
target_api:
  file: string                  # Required: File with API definitions
  type: string                  # "json_schema" | "openapi" | "graphql" | "protobuf"

parameters:
  - name: string                # Required: Parameter name
    type: string                # Required: Data type (string, number, boolean, etc.)
    description: string         # Required: Parameter description
    required: boolean           # Required: Is parameter required?

categorization:
  tier_definitions:             # Default: 5-tier system
    tier_1: "required"          # Must be provided
    tier_2: "filtering"         # Narrows results
    tier_3: "range"             # Defines bounds/thresholds
    tier_4: "output_control"    # Controls output size/format
    tier_5: "standard"          # Cross-cutting concerns

  ordering_strategy:
    by_tier: boolean            # Default: true (order by tier number)
    add_comments: boolean       # Default: true (add tier comments)
```

## Execution Process

### Step 1: Read Parameter Definitions

```bash
# Extract parameters from API definition
grep -A 3 "\"properties\":" tools.go | less

# Count parameters per tool
jq '.tools[].parameters | length' api-schema.json
```

**Output**:
```yaml
extracted_parameters:
  tool: "query_tools"
  parameters:
    - name: "tool"
      type: "string"
      description: "Filter by tool name"
      required: false
    - name: "status"
      type: "string"
      description: "Filter by status (error/success)"
      required: false
    # ... more parameters
```

### Step 2: Apply Tier-Based Decision Tree

**Decision Logic**:
```
For each parameter:

Question 1: "Is this required for tool to function?"
  YES → Tier 1 (Required)
  NO → Continue

Question 2: "Does this filter WHAT is returned?"
  YES → Tier 2 (Filtering)
  NO → Continue

Question 3: "Does this define a range or threshold?"
  YES → Tier 3 (Range)
  NO → Continue

Question 4: "Does this control output size/format?"
  YES → Tier 4 (Output Control)
  NO → Continue

Question 5: "Is this a standard cross-cutting parameter?"
  YES → Tier 5 (Standard)
  NO → Unclassified (needs manual review)
```

**Example Application**:
```yaml
parameter: "tool"
  question_1: "Required?" → NO
  question_2: "Filters results?" → YES (filters to specific tool)
  result: Tier 2 (Filtering)

parameter: "limit"
  question_1: "Required?" → NO
  question_2: "Filters results?" → NO (doesn't change WHAT, just HOW MUCH)
  question_3: "Range/threshold?" → NO
  question_4: "Output control?" → YES (controls result count)
  result: Tier 4 (Output Control)

parameter: "min_occurrences"
  question_1: "Required?" → NO
  question_2: "Filters results?" → NO
  question_3: "Range/threshold?" → YES (minimum threshold)
  result: Tier 3 (Range)

parameter: "scope"
  question_1: "Required?" → NO
  question_2: "Filters results?" → NO (cross-cutting)
  question_3: "Range/threshold?" → NO
  question_4: "Output control?" → NO
  question_5: "Standard parameter?" → YES
  result: Tier 5 (Standard)
```

### Step 3: Categorize All Parameters

**Process**:
```python
def categorize_parameters(parameters):
    categorized = {
        "tier_1": [],
        "tier_2": [],
        "tier_3": [],
        "tier_4": [],
        "tier_5": [],
        "unclassified": []
    }

    for param in parameters:
        # Apply decision tree
        tier = apply_decision_tree(param)
        categorized[tier].append(param)

    return categorized
```

**Output**:
```yaml
categorization_results:
  tool: "query_tools"
  tier_1: []  # No required parameters
  tier_2:
    - "tool"
    - "status"
  tier_3: []  # No range parameters
  tier_4:
    - "limit"
  tier_5:
    - "scope"
    - "jq_filter"
    - "stats_only"
```

### Step 4: Order Parameters by Tier

**Ordering Rule**: Tier 1 → Tier 2 → Tier 3 → Tier 4 → Tier 5

**Before** (arbitrary order):
```go
properties := map[string]Property{
    "limit": {...},         // Tier 4
    "tool": {...},          // Tier 2
    "scope": {...},         // Tier 5
    "status": {...},        // Tier 2
    "jq_filter": {...},     // Tier 5
}
```

**After** (tier-based order):
```go
properties := map[string]Property{
    // Tier 2: Filtering
    "tool": {...},
    "status": {...},
    // Tier 4: Output Control
    "limit": {...},
    // Tier 5: Standard Parameters
    "scope": {...},
    "jq_filter": {...},
}
```

### Step 5: Add Tier Comments

```go
properties := map[string]Property{
    // Tier 2: Filtering
    "tool": {
        Type: "string",
        Description: "Filter by tool name",
    },
    "status": {
        Type: "string",
        Description: "Filter by status (error/success)",
    },

    // Tier 4: Output Control
    "limit": {
        Type: "number",
        Description: "Maximum number of results",
    },

    // Tier 5: Standard Parameters (added automatically)
    "scope": {...},
    "jq_filter": {...},
    "stats_only": {...},
}
```

### Step 6: Verify Determinism

**Verification Checklist**:
- [ ] Every parameter assigned to exactly one tier
- [ ] No ambiguous cases (parameters fitting multiple tiers)
- [ ] Decision criteria consistently applied
- [ ] Ordering matches tier sequence (1→2→3→4→5)

**Example Verification**:
```yaml
verification_results:
  total_parameters: 15
  categorized: 15
  ambiguous: 0
  determinism_rate: 100%

  tier_distribution:
    tier_1: 0
    tier_2: 5
    tier_3: 2
    tier_4: 3
    tier_5: 5
```

### Step 7: Apply to All Tools

**Incremental Process**:
```bash
# Process Tool 1
categorize_tool "query_tools"
verify_ordering
commit

# Process Tool 2
categorize_tool "query_user_messages"
verify_ordering
commit

# ... repeat for all tools
```

**Batch Tracking**:
```yaml
tools_processed:
  - tool: "query_tools"
    parameters_categorized: 8
    ambiguous_cases: 0
    status: "COMPLETED"

  - tool: "query_user_messages"
    parameters_categorized: 10
    ambiguous_cases: 0
    status: "COMPLETED"

  # ... 6 more tools
```

### Step 8: Generate Categorization Report

```markdown
## Parameter Categorization Report

**Date**: 2025-10-16
**Tools Analyzed**: 8
**Total Parameters**: 60

### Tier Distribution

| Tier | Count | Percentage |
|------|-------|------------|
| Tier 1 (Required) | 2 | 3.3% |
| Tier 2 (Filtering) | 18 | 30.0% |
| Tier 3 (Range) | 8 | 13.3% |
| Tier 4 (Output Control) | 12 | 20.0% |
| Tier 5 (Standard) | 20 | 33.3% |

### Determinism

- **Ambiguous Cases**: 0
- **Determinism Rate**: 100%
- **Manual Review Required**: 0

### Tools Requiring Reordering

- query_tools (60% compliant → 100%)
- query_user_messages (75% compliant → 100%)
- query_conversation (40% compliant → 100%)
- query_tool_sequences (67% compliant → 100%)
- query_successful_prompts (0% compliant → 100%)

### Tools Already Compliant

- query_context (100% compliant)
- query_assistant_messages (100% compliant)
- query_time_series (100% compliant)
```

### Step 9: Document Decision Criteria

**Create Reference Document**:
```markdown
# Parameter Tier System

## Tier Definitions

### Tier 1: Required Parameters
**Criteria**: Must be provided for tool to function
**Examples**: error_signature (query_context), pattern (query_user_messages)
**Marker**: required=true in schema

### Tier 2: Filtering Parameters
**Criteria**: Narrows search results (affects WHAT is returned)
**Examples**: tool, status, pattern_target
**Markers**: String/enum types, filters data

### Tier 3: Range Parameters
**Criteria**: Defines bounds, thresholds, windows
**Examples**: min_occurrences, max_duration, window, start_turn, end_turn
**Markers**: Numeric types, prefixes like min_*, max_*, threshold

### Tier 4: Output Control Parameters
**Criteria**: Controls output size or format
**Examples**: limit, offset, output_format
**Markers**: Affects HOW MUCH, not WHAT

### Tier 5: Standard Parameters
**Criteria**: Cross-cutting concerns added automatically
**Examples**: scope, jq_filter, stats_only, inline_threshold_bytes
**Markers**: Common across all/most tools
```

### Step 10: Update API Convention Documentation

```markdown
## API Parameter Ordering Convention

All tools MUST order parameters according to the tier system:

1. **Tier 1** (Required)
2. **Tier 2** (Filtering)
3. **Tier 3** (Range)
4. **Tier 4** (Output Control)
5. **Tier 5** (Standard)

**Rationale**:
- Consistency: Users learn pattern once, applies everywhere
- Predictability: Logical grouping (required → filters → bounds → output)
- Readability: Tier comments clarify parameter purpose

**Enforcement**:
- Validation tool: `./validate-api --check-ordering`
- Pre-commit hook: Blocks non-compliant commits
- Documentation: Generate from schema (order preserved)
```

## Output Schema

```yaml
categorization_report:
  tool: string
  parameters_categorized: number
  ambiguous_cases: number

  tier_distribution:
    tier_1: number
    tier_2: number
    tier_3: number
    tier_4: number
    tier_5: number

  compliance:
    before: number        # % compliance before reordering
    after: number         # % compliance after reordering

  changes_made:
    - parameter: string
      old_position: number
      new_position: number
      tier: string

determinism_metrics:
  total_parameters: number
  categorized: number
  ambiguous: number
  determinism_rate: number  # % (should be 100%)

tools_status:
  - tool: string
    status: "COMPLETED" | "ALREADY_COMPLIANT" | "NEEDS_REVIEW"
    compliance_before: number
    compliance_after: number

reference_docs:
  tier_definitions: string  # Path to tier reference doc
  api_convention: string    # Path to API convention doc
```

## Success Criteria

- ✅ 100% determinism (no ambiguous cases)
- ✅ All parameters categorized into exactly one tier
- ✅ Tier comments added for readability
- ✅ Ordering matches tier sequence (1→2→3→4→5)
- ✅ Reference documentation created
- ✅ API convention updated

## Example Execution (Bootstrap-006 Iteration 4)

**Input**:
```yaml
target_api:
  file: "cmd/mcp-server/tools.go"
  type: "json_schema"

tools_to_analyze: 8
parameters_per_tool: 5-12
```

**Process**:
```
Step 1: Read parameters from 8 tools
  Extracted: 60 parameters

Step 2-3: Apply decision tree, categorize
  Tier 1: 2 parameters
  Tier 2: 18 parameters
  Tier 3: 8 parameters
  Tier 4: 12 parameters
  Tier 5: 20 parameters
  Ambiguous: 0 (100% determinism)

Step 4-5: Reorder and add comments
  5 tools reordered (60 lines changed)
  3 tools already compliant (verified)

Step 6: Verify determinism
  Determinism rate: 100%

Step 7: Process all tools
  8/8 tools completed

Step 8: Generate report
  Compliance: 67.5% → 100%

Step 9-10: Document criteria and conventions
  Created tier-definitions.md
  Updated api-parameter-convention.md
```

**Output**:
```yaml
determinism_rate: 100%
ambiguous_cases: 0
tools_reordered: 5
tools_verified: 3
compliance_improvement: 32.5 percentage points

tier_distribution:
  tier_1: 2 (3.3%)
  tier_2: 18 (30.0%)
  tier_3: 8 (13.3%)
  tier_4: 12 (20.0%)
  tier_5: 20 (33.3%)
```

## Pitfalls and How to Avoid

### Pitfall 1: Ambiguous Parameter Purpose
- ❌ Wrong: "This could be filtering OR output control"
- ✅ Right: Apply decision tree questions in order (Tier 2 before Tier 4)
- **Example**: `max_results` → Not filtering (doesn't change WHAT), is output control (controls HOW MUCH) → Tier 4

### Pitfall 2: Confusing Filtering with Range
- ❌ Wrong: "min_occurrences filters results, so Tier 2"
- ✅ Right: Tier 3 (defines minimum threshold, not a filter)
- **Distinction**: Filtering selects items; range defines bounds on numeric values

### Pitfall 3: Not Adding Tier Comments
- ❌ Wrong: Reorder without comments
- ✅ Right: Add `// Tier 2: Filtering` comments
- **Benefit**: Clarifies intent for future maintainers

### Pitfall 4: Forcing Parameters into Tiers
- ❌ Wrong: Force unusual parameter into tier system
- ✅ Right: Mark as "unclassified," review manually
- **Example**: Custom parameters with unique semantics

### Pitfall 5: Inconsistent Decision Criteria
- ❌ Wrong: Different developers use different interpretations
- ✅ Right: Document decision tree, apply consistently
- **Solution**: Create reference doc with examples

## Variations

### Variation 1: GraphQL Schema Arguments

```graphql
type Query {
  queryTools(
    # Tier 2: Filtering
    tool: String
    status: String
    # Tier 4: Output Control
    limit: Int
    # Tier 5: Standard
    scope: String
  ): [ToolResult]
}
```

### Variation 2: REST API Query Parameters

```yaml
GET /api/tools?tool=<name>&status=<status>&limit=<n>&scope=<scope>

Ordering in documentation:
  1. tool (Tier 2)
  2. status (Tier 2)
  3. limit (Tier 4)
  4. scope (Tier 5)
```

### Variation 3: CLI Flags

```bash
# Tier-based flag ordering
query-tools \
  --tool <name> \          # Tier 2: Filtering
  --status <status> \      # Tier 2: Filtering
  --limit <n> \            # Tier 4: Output Control
  --scope <scope>          # Tier 5: Standard
```

### Variation 4: Configuration Files

```yaml
query_tools:
  # Tier 2: Filtering
  tool: "example"
  status: "error"
  # Tier 4: Output Control
  limit: 100
  # Tier 5: Standard
  scope: "project"
```

## Language-Specific Adaptations

### Go (JSON Schema)
- Define properties in map (order preserved in modern Go)
- Add tier comments before each group
- Use builder pattern for common parameters

### Python (Type Hints)
- Use dataclasses with ordering
- Add docstrings with tier info
- Leverage field() for metadata

### TypeScript (Interface)
- Define interfaces with ordered properties
- Add JSDoc comments for tiers
- Use extends for standard parameters

### Protobuf
- Field numbers don't affect ordering
- Order in documentation follows tier system
- Add comments for tier groups

## Usage Examples

### As Subagent

```bash
/subagent @experiments/bootstrap-006-api-design/agents/agent-parameter-categorizer.md \
  target_api.file="cmd/mcp-server/tools.go" \
  target_api.type="json_schema" \
  categorization.add_comments=true
```

### As Slash Command (if registered)

```bash
/categorize-params \
  file="cmd/mcp-server/tools.go" \
  add-comments
```

## Evidence from Bootstrap-006

**Source**: Iteration 4, Task 1 (Parameter Reordering)

**Tools Analyzed**: 8
- query_tools
- query_user_messages
- query_conversation
- query_context
- query_assistant_messages
- query_time_series
- query_tool_sequences
- query_successful_prompts

**Results**:
- Parameters categorized: 60
- Ambiguous cases: 0
- Determinism rate: 100%
- Tools reordered: 5
- Tools already compliant: 3
- Compliance improvement: 67.5% → 100%

**Time Spent**: ~2 hours for 8 tools

**Quality**:
- Decision tree applied consistently
- No manual review required
- 100% test pass rate after reordering

---

**Last Updated**: 2025-10-16
**Status**: Validated (Bootstrap-006 Iteration 4)
**Reusability**: Universal (any parametric interface with 5-tier structure)
