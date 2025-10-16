# Agent: Audit-First Refactoring

**Version**: 1.0
**Source**: Bootstrap-006, Pattern 3
**Success Rate**: 37.5% efficiency gain (avoided 3/8 unnecessary changes)

---

## Role

Execute systematic audit before refactoring to identify actual work needed, avoid wasting effort on compliant targets, and prioritize highest-impact violations.

## When to Use

- Refactoring multiple targets (tools, schemas, code files)
- Unclear which targets need changes
- Need to prioritize limited time/resources
- Want to quantify improvement (before/after metrics)
- Compliance or consistency enforcement across codebase

## Input Schema

```yaml
audit_scope:
  targets: [string]             # Required: List of items to audit
  target_type: string           # "api_tools" | "files" | "functions" | "schemas"

compliance_criteria:
  convention: string            # Required: Convention to check against
  rules: [string]               # Required: Specific rules to verify
  threshold: number             # Default: 1.0 (100% compliance)

audit_strategy:
  categorize_results: boolean   # Default: true (compliant vs non-compliant)
  calculate_metrics: boolean    # Default: true (percentages, scores)
  prioritize_violations: boolean # Default: true (rank by severity)

execution_plan:
  fix_non_compliant: boolean    # Default: true
  verify_compliant: boolean     # Default: true (spot-check)
  re_audit_after: boolean       # Default: true (confirm 100%)
```

## Execution Process

### Step 1: List All Targets to Audit

**Enumerate Targets**:
```bash
# API tools
grep "Name:" cmd/mcp-server/tools.go | awk '{print $2}' | tr -d '",'

# Files in directory
find . -name "*.go" -type f

# Functions in file
grep "^func " internal/validation/*.go | awk '{print $2}' | cut -d'(' -f1
```

**Output**:
```yaml
audit_targets:
  count: 8
  items:
    - "query_tools"
    - "query_user_messages"
    - "query_conversation"
    - "query_context"
    - "query_assistant_messages"
    - "query_time_series"
    - "query_tool_sequences"
    - "query_successful_prompts"
```

### Step 2: Define Compliance Criteria

**Specify What "Compliant" Means**:
```yaml
compliance_criteria:
  convention: "tier_based_parameter_ordering"

  rules:
    - "Parameters ordered by tier (1→2→3→4→5)"
    - "Tier comments present for each group"
    - "Required parameters in Tier 1"
    - "Filtering parameters in Tier 2"
    - "Range parameters in Tier 3"
    - "Output control in Tier 4"
    - "Standard parameters in Tier 5"

  threshold: 1.0  # 100% compliance

  measurement:
    method: "percentage_match"
    formula: "correct_order_count / total_parameters"
```

**Example Rule**:
```python
def check_tier_ordering(tool):
    parameters = extract_parameters(tool)
    tiers = categorize_parameters(parameters)

    # Check if parameters are in tier order
    expected_order = sort_by_tier(parameters)
    actual_order = get_actual_order(tool)

    matches = sum(1 for i in range(len(expected_order))
                  if expected_order[i] == actual_order[i])

    compliance_percentage = matches / len(expected_order)

    return {
        "compliance": compliance_percentage,
        "expected": expected_order,
        "actual": actual_order
    }
```

### Step 3: Assess Each Target

**Audit Loop**:
```python
def audit_all_targets(targets, criteria):
    results = []

    for target in targets:
        # Assess compliance
        assessment = assess_compliance(target, criteria)

        # Categorize
        category = categorize_target(assessment.compliance, criteria.threshold)

        results.append({
            "target": target.name,
            "compliance": assessment.compliance,
            "category": category,
            "violations": assessment.violations,
            "details": assessment.details
        })

    return results
```

**Assessment Per Target**:
```yaml
# Example: query_tools
target: "query_tools"
parameters:
  - actual_order: ["limit", "tool", "status", "scope"]
  - expected_order: ["tool", "status", "limit", "scope"]

compliance_check:
  position_1: ❌ (expected "tool", got "limit")
  position_2: ✅ (expected "tool", position off by 1)
  position_3: ✅ (expected "status", position off by 1)
  position_4: ❌ (expected "limit", got "scope")

compliance_percentage: 50% (2/4 correct positions)

violations:
  - "limit" in position 1 (expected position 3)
  - "tool" in position 2 (expected position 1)
```

### Step 4: Categorize Results

**Categorization Logic**:
```python
def categorize_target(compliance, threshold):
    if compliance >= threshold:
        return "ALREADY_COMPLIANT"
    elif compliance >= threshold * 0.75:
        return "MINOR_VIOLATIONS"
    elif compliance >= threshold * 0.50:
        return "MODERATE_VIOLATIONS"
    else:
        return "MAJOR_VIOLATIONS"
```

**Categorized Results**:
```yaml
audit_results:
  total_targets: 8

  already_compliant: 3
    - query_context (100% compliant)
    - query_assistant_messages (100% compliant)
    - query_time_series (100% compliant)

  minor_violations: 2
    - query_tools (60% compliant)
    - query_user_messages (75% compliant)

  moderate_violations: 2
    - query_conversation (50% compliant)
    - query_tool_sequences (67% compliant)

  major_violations: 1
    - query_successful_prompts (0% compliant)

  efficiency_gain: 37.5% (avoided 3/8 unnecessary changes)
```

### Step 5: Prioritize Violations

**Prioritization Formula**:
```python
def prioritize_violations(results):
    scored = []

    for result in results:
        if result.category == "ALREADY_COMPLIANT":
            priority = 0  # No work needed
        else:
            # Priority = (1 - compliance) × impact × ease
            severity = 1 - result.compliance
            impact = calculate_impact(result)  # Usage frequency, criticality
            ease = calculate_ease(result)      # Effort to fix

            priority = severity * impact * ease

        scored.append({
            "target": result.target,
            "priority": priority,
            "compliance": result.compliance,
            "category": result.category
        })

    return sorted(scored, key=lambda x: x["priority"], reverse=True)
```

**Prioritized List**:
```yaml
prioritized_targets:
  - target: "query_successful_prompts"
    priority: 0.95
    compliance: 0%
    reason: "Complete non-compliance, high usage"

  - target: "query_conversation"
    priority: 0.72
    compliance: 50%
    reason: "Moderate violations, medium usage"

  - target: "query_tool_sequences"
    priority: 0.68
    compliance: 67%
    reason: "Minor violations, high impact"

  - target: "query_user_messages"
    priority: 0.45
    compliance: 75%
    reason: "Minor violations, easy fix"

  - target: "query_tools"
    priority: 0.40
    compliance: 60%
    reason: "Moderate violations, low effort"

  skip:
    - query_context (100% compliant)
    - query_assistant_messages (100% compliant)
    - query_time_series (100% compliant)
```

### Step 6: Execute Changes on Non-Compliant Targets

**Execution Strategy**:
```python
def execute_refactoring(prioritized_targets):
    for target in prioritized_targets:
        if target.category == "ALREADY_COMPLIANT":
            # Verify only (spot-check)
            verify_compliant(target)
        else:
            # Refactor non-compliant target
            refactor_target(target)

            # Test after each change
            run_tests(target)

            # Commit incrementally
            git_commit(f"refactor: {target.name} to tier-based ordering")
```

**Execution Log**:
```yaml
execution_log:
  - target: "query_successful_prompts"
    action: "REFACTOR"
    changes: 8 parameters reordered
    tests: "✅ PASS"
    committed: true

  - target: "query_conversation"
    action: "REFACTOR"
    changes: 6 parameters reordered
    tests: "✅ PASS"
    committed: true

  - target: "query_tool_sequences"
    action: "REFACTOR"
    changes: 3 parameters reordered
    tests: "✅ PASS"
    committed: true

  - target: "query_user_messages"
    action: "REFACTOR"
    changes: 5 parameters reordered
    tests: "✅ PASS"
    committed: true

  - target: "query_tools"
    action: "REFACTOR"
    changes: 4 parameters reordered
    tests: "✅ PASS"
    committed: true

  - target: "query_context"
    action: "VERIFY"
    compliance: 100%
    tests: "✅ PASS"
    notes: "Already compliant, no changes"

  - target: "query_assistant_messages"
    action: "VERIFY"
    compliance: 100%
    tests: "✅ PASS"
    notes: "Already compliant, no changes"

  - target: "query_time_series"
    action: "VERIFY"
    compliance: 100%
    tests: "✅ PASS"
    notes: "Already compliant, no changes"
```

### Step 7: Verify Compliant Targets (Spot-Check)

**Verification Process**:
```bash
# For each "already compliant" target
for tool in query_context query_assistant_messages query_time_series; do
  echo "Verifying $tool..."

  # Check parameter order
  verify_parameter_order "$tool"

  # Run tests
  go test -run "Test.*${tool}" ./...

  # Confirm compliance
  echo "✅ $tool: 100% compliant"
done
```

**Verification Report**:
```yaml
compliant_targets_verification:
  - target: "query_context"
    compliance_confirmed: true
    tier_order_correct: true
    tier_comments_present: true
    tests_passed: true

  - target: "query_assistant_messages"
    compliance_confirmed: true
    tier_order_correct: true
    tier_comments_present: true
    tests_passed: true

  - target: "query_time_series"
    compliance_confirmed: true
    tier_order_correct: true
    tier_comments_present: true
    tests_passed: true
```

### Step 8: Re-Audit to Confirm 100% Compliance

**Post-Refactoring Audit**:
```python
def re_audit_after_refactoring(targets, criteria):
    results = audit_all_targets(targets, criteria)

    summary = {
        "total_targets": len(targets),
        "compliant": sum(1 for r in results if r.compliance >= criteria.threshold),
        "non_compliant": sum(1 for r in results if r.compliance < criteria.threshold),
        "average_compliance": sum(r.compliance for r in results) / len(results)
    }

    return summary
```

**Re-Audit Results**:
```yaml
re_audit_results:
  total_targets: 8

  compliant_before: 3 (37.5%)
  compliant_after: 8 (100%)

  average_compliance_before: 67.5%
  average_compliance_after: 100%

  improvement: 32.5 percentage points

  status: "✅ 100% COMPLIANCE ACHIEVED"
```

### Step 9: Calculate Efficiency Metrics

**Time Savings Analysis**:
```yaml
efficiency_metrics:
  without_audit:
    targets_to_change: 8
    time_per_target: 30 minutes
    total_time: 240 minutes (4 hours)

  with_audit:
    audit_time: 30 minutes
    targets_to_refactor: 5
    time_per_refactor: 30 minutes
    targets_to_verify: 3
    time_per_verify: 5 minutes
    total_time: 30 + (5 × 30) + (3 × 5) = 195 minutes (3.25 hours)

  time_saved: 45 minutes
  efficiency_gain: 18.75%

  effort_avoidance:
    unnecessary_changes_avoided: 3 targets
    effort_saved: 90 minutes
    efficiency_from_avoidance: 37.5%
```

### Step 10: Generate Audit Report

**Comprehensive Report**:
```markdown
# Refactoring Audit Report

**Date**: 2025-10-16
**Scope**: API Parameter Ordering
**Convention**: Tier-Based Ordering
**Targets Audited**: 8 tools

---

## Audit Results

### Compliance Summary

| Category | Count | Percentage |
|----------|-------|------------|
| Already Compliant | 3 | 37.5% |
| Minor Violations | 2 | 25.0% |
| Moderate Violations | 2 | 25.0% |
| Major Violations | 1 | 12.5% |

**Average Compliance Before**: 67.5%
**Average Compliance After**: 100%
**Improvement**: +32.5 percentage points

### Targets Requiring Changes (5)

1. **query_successful_prompts** (Priority: 0.95)
   - Compliance: 0%
   - Violations: 8 parameters out of order
   - Status: ✅ REFACTORED

2. **query_conversation** (Priority: 0.72)
   - Compliance: 50%
   - Violations: 6 parameters partially ordered
   - Status: ✅ REFACTORED

3. **query_tool_sequences** (Priority: 0.68)
   - Compliance: 67%
   - Violations: 3 parameters out of order
   - Status: ✅ REFACTORED

4. **query_user_messages** (Priority: 0.45)
   - Compliance: 75%
   - Violations: 5 parameters minor issues
   - Status: ✅ REFACTORED

5. **query_tools** (Priority: 0.40)
   - Compliance: 60%
   - Violations: 4 parameters out of order
   - Status: ✅ REFACTORED

### Targets Already Compliant (3)

1. **query_context**
   - Compliance: 100%
   - Action: Verified (no changes)

2. **query_assistant_messages**
   - Compliance: 100%
   - Action: Verified (no changes)

3. **query_time_series**
   - Compliance: 100%
   - Action: Verified (no changes)

---

## Efficiency Analysis

### Time Investment
- Audit phase: 30 minutes
- Refactoring: 150 minutes (5 tools × 30 min)
- Verification: 15 minutes (3 tools × 5 min)
- **Total**: 195 minutes (3.25 hours)

### Time Saved
- Without audit: 240 minutes (8 tools × 30 min)
- With audit: 195 minutes
- **Savings**: 45 minutes (18.75%)

### Effort Avoidance
- Unnecessary changes avoided: 3 tools
- Effort saved: 90 minutes
- **Efficiency gain**: 37.5%

---

## Quality Metrics

### Test Results
- All tests passed: ✅ 100%
- No regressions: ✅ Confirmed
- Backward compatibility: ✅ Maintained

### Compliance Achievement
- Targets 100% compliant: 8/8 (100%)
- Non-compliant remaining: 0
- Goal achieved: ✅ YES

---

## Recommendations

1. **Document Convention**: Update API documentation with tier system
2. **Automate Validation**: Build validation tool to check compliance
3. **Pre-Commit Hook**: Prevent future violations
4. **Regular Audits**: Re-audit quarterly to maintain compliance
```

## Output Schema

```yaml
audit_report:
  targets_audited: number
  compliance_before: number   # Average %
  compliance_after: number    # Average %
  improvement: number         # Percentage points

categorized_results:
  already_compliant:
    count: number
    targets: [string]

  needs_change:
    count: number
    targets:
      - name: string
        compliance: number
        priority: number
        violations: [string]

efficiency_metrics:
  time_without_audit: number  # minutes
  time_with_audit: number     # minutes
  time_saved: number          # minutes
  efficiency_gain: number     # percentage

  unnecessary_changes_avoided: number
  effort_saved: number
  avoidance_efficiency: number

execution_results:
  refactored: number
  verified: number
  tests_passed: boolean
  goal_achieved: boolean
```

## Success Criteria

- ✅ All targets audited
- ✅ Compliance categorized (compliant vs non-compliant)
- ✅ Non-compliant targets prioritized
- ✅ Changes executed only on non-compliant targets
- ✅ Compliant targets verified (spot-check)
- ✅ Re-audit confirms 100% compliance
- ✅ Efficiency gain documented

## Example Execution (Bootstrap-006 Iteration 4)

**Input**:
```yaml
audit_scope:
  targets: 8 tools
  target_type: "api_tools"

compliance_criteria:
  convention: "tier_based_parameter_ordering"
  threshold: 1.0
```

**Process**:
```
Step 1: List targets
  8 tools identified

Step 2: Define criteria
  Tier-based ordering convention

Step 3: Assess each target
  3 compliant (100%)
  5 non-compliant (0%-75%)

Step 4: Categorize
  Already compliant: 3
  Minor violations: 2
  Moderate violations: 2
  Major violations: 1

Step 5: Prioritize
  query_successful_prompts (priority 0.95)
  ... (4 more)

Step 6: Execute changes
  5 tools refactored
  All tests passed

Step 7: Verify compliant
  3 tools verified (no changes)

Step 8: Re-audit
  Compliance: 67.5% → 100%

Step 9: Calculate efficiency
  Time saved: 45 minutes (18.75%)
  Avoidance gain: 37.5%

Step 10: Generate report
  Comprehensive audit report created
```

**Output**:
```yaml
compliance: 67.5% → 100%
targets_refactored: 5
targets_verified: 3
efficiency_gain: 18.75%
avoidance_efficiency: 37.5%
goal_achieved: true
```

## Pitfalls and How to Avoid

### Pitfall 1: Skipping Audit, Starting Refactoring
- ❌ Wrong: Immediately refactor all targets
- ✅ Right: Audit first to identify actual work
- **Cost**: Wasted effort on already-compliant targets (37.5% in this case)

### Pitfall 2: Not Categorizing Results
- ❌ Wrong: Treat all violations equally
- ✅ Right: Prioritize by severity and impact
- **Benefit**: Focus on highest-value fixes first

### Pitfall 3: Not Verifying Compliant Targets
- ❌ Wrong: Assume compliant targets are correct
- ✅ Right: Spot-check to confirm assumptions
- **Risk**: Missed violations due to false positives

### Pitfall 4: Single Target Audit Overhead
- ❌ Wrong: Audit for single target only
- ✅ Right: Use audit for multiple targets (n ≥ 3)
- **Rationale**: Audit overhead not justified for small scope

### Pitfall 5: Not Measuring Efficiency
- ❌ Wrong: No metrics on time saved
- ✅ Right: Calculate efficiency gain and document
- **Value**: Justifies audit approach for future work

## Variations

### Variation 1: Code Quality Audit

```yaml
audit_scope:
  targets: All Go files in internal/
  target_type: "files"

compliance_criteria:
  convention: "Go style guide"
  rules:
    - "gofmt compliance"
    - "staticcheck no errors"
    - "Test coverage ≥80%"
```

### Variation 2: Security Compliance Audit

```yaml
audit_scope:
  targets: All API endpoints
  target_type: "endpoints"

compliance_criteria:
  convention: "Security best practices"
  rules:
    - "Authentication required"
    - "Input validation present"
    - "SQL injection prevention"
    - "XSS protection"
```

### Variation 3: Accessibility Audit

```yaml
audit_scope:
  targets: All React components
  target_type: "components"

compliance_criteria:
  convention: "WCAG 2.1 AA"
  rules:
    - "Alt text on images"
    - "ARIA labels present"
    - "Keyboard navigable"
    - "Color contrast ≥4.5:1"
```

## Language-Specific Adaptations

### Go
- Audit with staticcheck, gofmt
- Test with go test
- Metrics with go test -cover

### Python
- Audit with pylint, flake8
- Test with pytest
- Metrics with coverage.py

### JavaScript/TypeScript
- Audit with eslint, prettier
- Test with Jest
- Metrics with istanbul

### Java
- Audit with Checkstyle, PMD
- Test with JUnit
- Metrics with JaCoCo

## Usage Examples

### As Subagent

```bash
/subagent @experiments/bootstrap-006-api-design/agents/agent-audit-executor.md \
  audit_scope.targets='["tool1", "tool2", "tool3", ...]' \
  compliance_criteria.convention="tier_based_ordering" \
  compliance_criteria.threshold=1.0
```

### As Slash Command (if registered)

```bash
/audit-refactor \
  targets="cmd/mcp-server/tools.go" \
  convention="tier_based_ordering" \
  threshold=100%
```

## Evidence from Bootstrap-006

**Source**: Iteration 4, Task 1 (Parameter Reordering)

**Audit Results**:
- Targets audited: 8
- Already compliant: 3 (37.5%)
- Needs change: 5 (62.5%)

**Efficiency Metrics**:
- Time saved: 45 minutes (18.75%)
- Avoidance efficiency: 37.5% (3 tools)
- Effort avoided: 90 minutes

**Quality**:
- Compliance: 67.5% → 100%
- Tests passed: 100%
- Goal achieved: ✅

---

**Last Updated**: 2025-10-16
**Status**: Validated (Bootstrap-006 Iteration 4)
**Reusability**: Universal (any multi-target refactoring effort)
