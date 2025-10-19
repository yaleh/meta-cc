# Backward Compatibility Guidelines for meta-cc MCP API

**Document Version**: 1.0
**Created**: 2025-10-15
**Status**: Proposed
**Applies To**: meta-cc MCP API v1.0.0+

---

## 1. Compatibility Guarantees

### What We Promise

Within a **MAJOR version** (e.g., all v1.x.x releases):

✅ **Guaranteed Compatible**:
- Existing tools will continue to work
- Required parameters will not be removed
- Parameter defaults will not change behavior
- Response format structure will remain stable
- Error codes will remain stable
- Documented behavior will not change (unless it was a bug)

✅ **Safe Additions**:
- New tools may be added
- New optional parameters may be added (with safe defaults)
- New response fields may be added (existing fields unchanged)
- New values for existing enums may be added

❌ **Not Guaranteed**:
- Performance characteristics (may improve or degrade)
- Internal implementation (may change)
- Undocumented behavior (may change)
- Error messages text (codes stable, text may improve)

### What May Break Across Major Versions

Between **MAJOR versions** (e.g., v1.x.x → v2.0.0):

⚠️ **May Change** (After 12-month deprecation):
- Tools may be removed or renamed
- Required parameters may be added or removed
- Parameter defaults may change
- Response formats may evolve
- Behavior may change

✅ **Always Provided**:
- 12-month minimum notice period
- Migration guide
- Deprecation warnings in v1.x.x
- Compatibility shim during migration (if feasible)

---

## 2. Safe Evolution Patterns

### Pattern 1: Adding New Tools

**✅ SAFE** - Can be done in MINOR release

**Example**:
```yaml
v1.2.0:
  new_tool: "query_insights"
  impact: None (doesn't affect existing tools)
  version_bump: MINOR
```

**Guidelines**:
- Choose descriptive, consistent name (follow query_* pattern)
- Document thoroughly
- Add examples to MCP guide

**Rationale**: Users not calling new tool are unaffected.

---

### Pattern 2: Adding Optional Parameters

**✅ SAFE** - Can be done in MINOR release

**Requirements**:
- Parameter MUST be optional
- Parameter MUST have safe default that preserves existing behavior
- Default MUST be documented

**Example**:
```yaml
tool: "query_files"
change: "Add include_archived: false"
existing_behavior: "Only shows non-archived files"
new_parameter:
  name: "include_archived"
  type: boolean
  default: false
  description: "If true, include archived files"
impact: None (default=false preserves behavior)
version_bump: MINOR (1.3.0)
```

**Anti-Pattern** ❌:
```yaml
# BAD: Adding required parameter
tool: "query_files"
change: "Add required file_type parameter"
impact: All existing calls fail (missing required param)
version_bump: MAJOR (breaking change)
```

**Guidelines**:
- Always provide sensible default
- Default should be "most common use case"
- Document both default and rationale

---

### Pattern 3: Adding Response Fields

**✅ SAFE** - Can be done in MINOR release

**Requirements**:
- New fields MUST be additions only
- Existing fields MUST NOT be removed or renamed
- Existing fields MUST NOT change type or meaning
- Response structure MUST remain valid

**Example**:
```yaml
tool: "query_tools"
change: "Add execution_time_ms to response"

before:
  {
    "tool": "query_tools",
    "status": "success",
    "results": [...]
  }

after:
  {
    "tool": "query_tools",
    "status": "success",
    "results": [...],
    "execution_time_ms": 150  # NEW FIELD
  }

impact: None (parsers ignore unknown fields)
version_bump: MINOR (1.4.0)
```

**Anti-Pattern** ❌:
```yaml
# BAD: Removing or renaming field
before:
  { "results": [...] }

after:
  { "data": [...] }  # Renamed results → data

impact: BREAKING (parsers expect "results")
version_bump: MAJOR
```

**Guidelines**:
- Additive only - never remove or rename
- Keep field types stable (don't change string → number)
- Keep field meaning stable (don't reuse field names for different data)

---

### Pattern 4: Relaxing Validation

**✅ SAFE** - Can be done in MINOR or PATCH release

**Definition**: Accepting more input formats than before.

**Example**:
```yaml
tool: "query_user_messages"
parameter: "pattern"
change: "Accept glob patterns in addition to regex"

before: pattern accepts only regex
after: pattern accepts regex OR glob

impact: None (existing regex patterns still work)
version_bump: MINOR (1.5.0)
```

**Guidelines**:
- Previously valid inputs MUST still be valid
- New inputs can be added
- Validation must be backward compatible

**Anti-Pattern** ❌:
```yaml
# BAD: Tightening validation (see Pattern 5)
before: pattern accepts any string
after: pattern must be valid regex (rejects invalid)

impact: BREAKING (previously accepted invalid inputs now fail)
version_bump: MAJOR or MINOR with careful deprecation
```

---

### Pattern 5: Tightening Validation

**⚠️ POTENTIALLY BREAKING** - Requires careful handling

**Definition**: Rejecting inputs that were previously accepted.

**When It's Safe** (PATCH release):
- Fixing security vulnerabilities
- Rejecting inputs that always caused errors anyway
- Restoring documented behavior (was always supposed to reject)

**Example (Safe)**:
```yaml
tool: "query_tools"
parameter: "jq_filter"
issue: "Invalid jq filters cause runtime errors anyway"
change: "Validate jq filter syntax upfront"

before: Accepts any string, fails at runtime
after: Rejects invalid jq at call time

impact: Low (was going to fail anyway, now fails faster)
version_bump: PATCH (1.5.1) or MINOR (1.6.0)
rationale: "Fail-fast improvement, errors happened anyway"
```

**When It's Breaking** (MAJOR release):
- Rejecting inputs that previously "worked" (even if incorrectly)
- Changing validation rules arbitrarily

**Example (Breaking)**:
```yaml
tool: "query_files"
parameter: "threshold"
change: "Must be 1-100 (previously accepted 0)"

before: Accepts threshold=0 (returns all files)
after: Rejects threshold=0

impact: BREAKING (users with threshold=0 fail)
version_bump: MAJOR (2.0.0)
migration: 12-month deprecation warning if threshold=0
```

**Guidelines**:
- Classify as SAFE only if inputs always failed anyway
- Otherwise, treat as BREAKING and deprecate
- Provide clear error messages for rejected inputs

---

### Pattern 6: Changing Parameter Defaults

**❌ BREAKING** - Requires MAJOR version + deprecation

**Definition**: Changing the default value of an optional parameter.

**Why It's Breaking**: Users relying on default value get different behavior.

**Example**:
```yaml
tool: "query_tools"
parameter: "scope"
change: "Default from 'project' to 'session'"

before: scope defaults to "project"
after: scope defaults to "session"

impact: BREAKING (behavior changes for users not setting scope)
version_bump: MAJOR (2.0.0)
migration_path:
  v1.6.0: Warn if scope not explicitly provided
  v1.6.0-v1.x.x: 12-month warning period
  v2.0.0: Change default to "session"
```

**Deprecation Path**:
1. **v1.6.0**: Add warning if parameter not explicitly set
   ```
   WARNING: 'scope' parameter not provided. Currently defaults to 'project'.
   In v2.0.0, default will change to 'session'.
   Explicitly set scope='project' to preserve current behavior.
   ```
2. **v1.6.0 - v1.x.x**: 12-month warning period
3. **v2.0.0**: Change default

**Guidelines**:
- Avoid changing defaults unless necessary
- Always deprecate first (12-month warning)
- Make warning actionable (tell users how to preserve behavior)

---

### Pattern 7: Renaming Tools

**❌ BREAKING** - Requires MAJOR version + deprecation

**Example**:
```yaml
old_name: "get_session_stats"
new_name: "query_session_stats"
reason: "Naming consistency (query_* pattern)"

migration_path:
  v1.5.0:
    - Add query_session_stats (new)
    - Deprecate get_session_stats
    - get_session_stats → calls query_session_stats + warning

  v1.5.0 - v1.x.x:
    - Both tools work (12-month overlap)

  v2.0.0:
    - Remove get_session_stats
    - Only query_session_stats remains
```

**Guidelines**:
- Add new tool first
- Deprecate old tool (mark in metadata)
- Old tool can call new tool internally (compatibility shim)
- Provide migration guide (search-and-replace instructions)
- Remove old tool in next MAJOR version

---

### Pattern 8: Removing Tools

**❌ BREAKING** - Requires MAJOR version + deprecation

**Example**:
```yaml
tool: "cleanup_temp_files"
reason: "Rarely used, manual cleanup sufficient"
usage: "6 calls total (0.3% of all tool calls)"

migration_path:
  v1.7.0:
    - Mark cleanup_temp_files deprecated
    - Provide manual cleanup documentation
    - Warning: "This tool will be removed in v2.0.0"

  v1.7.0 - v1.x.x:
    - 6-12 month deprecation period (shorter if very low usage)

  v2.0.0:
    - Remove cleanup_temp_files
    - Update docs with manual cleanup instructions
```

**Guidelines**:
- Only remove if usage is very low OR replacement is clearly better
- Provide alternative (replacement tool or manual process)
- 12-month deprecation for frequently used, 6-month for rarely used
- Document alternative clearly

---

### Pattern 9: Changing Response Format Structure

**❌ BREAKING** - Requires MAJOR version + careful migration

**Example**:
```yaml
change: "Hybrid mode threshold from 8KB to 16KB"

before: Results >8KB → file_ref mode
after: Results >16KB → file_ref mode

impact: Moderate (more inline results, users expecting file_ref may break)
version_bump: MINOR (behavior change but not breaking)
rationale: "Improving usability, file_ref mode still works"
```

**Example (Breaking)**:
```yaml
change: "JSONL → JSON array format"

before:
  {"tool":"query_tools","result":{"..."}}\n
  {"tool":"query_tools","result":{"..."}}

after:
  [
    {"tool":"query_tools","result":{"..."}},
    {"tool":"query_tools","result":{"..."}}
  ]

impact: BREAKING (parsers expect JSONL, not JSON)
version_bump: MAJOR (2.0.0)
migration: Add output_format parameter, deprecate default
```

**Guidelines**:
- Structural changes are always BREAKING
- Provide output_format parameter to choose (best option)
- Or use compatibility shim that converts new → old format
- 12-18 month deprecation (parsing changes are complex)

---

## 3. Testing Strategy

### Backward Compatibility Testing

**Goal**: Ensure changes don't break existing usage.

#### Test 1: Existing Parameter Combinations

**Method**: Regression test suite with real-world parameter combinations.

**Example**:
```yaml
test: "query_tools backward compatibility"
cases:
  - params: {}  # No parameters
    expected: "Default behavior preserved"

  - params: { scope: "project" }
    expected: "Explicit scope works"

  - params: { stats_only: true }
    expected: "Stats-only mode works"

  - params: { scope: "session", stats_only: true }
    expected: "Combined params work"
```

**Verdict**:
- All tests MUST pass for MINOR/PATCH release
- If any fail → MAJOR release required

#### Test 2: Response Format Stability

**Method**: JSON schema validation.

**Example**:
```yaml
test: "query_tools response format"
schema: |
  {
    "type": "object",
    "required": ["tool", "status"],
    "properties": {
      "tool": { "type": "string" },
      "status": { "type": "string" },
      "results": { "type": "array" }
    },
    "additionalProperties": true  # New fields allowed
  }
```

**Verdict**:
- Existing required fields MUST remain required
- Existing field types MUST remain stable
- New fields can be added (additionalProperties: true)

#### Test 3: Default Value Stability

**Method**: Test with parameters omitted.

**Example**:
```yaml
test: "Default values preserved"
cases:
  - tool: "query_tools"
    params: {}  # No scope provided
    expected_scope: "project"  # Current default

  - tool: "query_files"
    params: {}  # No threshold provided
    expected_threshold: 5  # Current default
```

**Verdict**:
- Defaults MUST NOT change in MINOR/PATCH
- If default changes → MAJOR + deprecation

---

### Compatibility Testing Checklist

Before releasing ANY version:

**For PATCH releases**:
- [ ] All existing tests pass
- [ ] Bug fix restores documented behavior
- [ ] No parameter changes
- [ ] No response format changes
- [ ] Error messages may improve (codes unchanged)

**For MINOR releases**:
- [ ] All existing tests pass
- [ ] New features are additive only
- [ ] New parameters have safe defaults
- [ ] Response format additions only (no removals/renames)
- [ ] Defaults unchanged
- [ ] Validation relaxed OR unchanged (not tightened)

**For MAJOR releases**:
- [ ] 12-month deprecation period completed
- [ ] Migration guide published
- [ ] Changelog has BREAKING CHANGE section
- [ ] Deprecated features removed
- [ ] All breaking changes documented

---

## 4. Compatibility Anti-Patterns

### ❌ Anti-Pattern 1: "Stealth Breaking Change"

**Description**: Changing behavior without announcing it as breaking.

**Example**:
```yaml
# BAD
version: 1.6.0 (MINOR)
change: "Changed scope default from 'project' to 'session'"
announcement: "Improved default behavior"
impact: BREAKING (users relying on default get different results)

# GOOD
version: 1.6.0 (MINOR)
change: "Add deprecation warning for implicit scope"
announcement: "Future default will change (v2.0.0)"

version: 2.0.0 (MAJOR)
change: "Changed scope default from 'project' to 'session'"
announcement: "BREAKING CHANGE after 12-month deprecation"
```

---

### ❌ Anti-Pattern 2: "Required Parameter Surprise"

**Description**: Adding required parameter without deprecation.

**Example**:
```yaml
# BAD
version: 1.5.0 (MINOR)
change: "Added required parameter 'file_type' to query_files"
impact: BREAKING (all existing calls fail)

# GOOD
version: 1.5.0 (MINOR)
change: "Added optional parameter 'file_type' with default 'all'"
impact: None (default preserves behavior)

# OR
version: 1.5.0 (MINOR)
change: "Add deprecation warning if file_type not provided"
version: 2.0.0 (MAJOR)
change: "Make file_type required (after 12-month deprecation)"
```

---

### ❌ Anti-Pattern 3: "Field Rename Chaos"

**Description**: Renaming response fields without compatibility.

**Example**:
```yaml
# BAD
version: 1.7.0 (MINOR)
change: "Renamed 'results' field to 'data'"
impact: BREAKING (parsers expect 'results')

# GOOD
version: 1.7.0 (MINOR)
change: "Add 'data' field (duplicate of 'results'), deprecate 'results'"
response:
  {
    "results": [...],  # Deprecated but still present
    "data": [...]      # New field (same content)
  }

version: 2.0.0 (MAJOR)
change: "Remove deprecated 'results' field"
response:
  {
    "data": [...]  # Only new field remains
  }
```

---

### ❌ Anti-Pattern 4: "Undocumented Behavior Dependency"

**Description**: Relying on undocumented behavior, then breaking it.

**Example**:
```yaml
# UNDOCUMENTED BEHAVIOR
tool: "query_tools"
behavior: "Returns tools in alphabetical order"
documentation: "Returns list of tools" (order not specified)

# USER DEPENDS ON IT
user_code: "Assumes first result is 'cleanup_temp_files' (alphabetical)"

# DEVELOPER CHANGES IT
version: 1.6.0 (MINOR)
change: "Return tools in usage frequency order"
impact: BREAKS user code relying on alphabetical order

# SOLUTION
version: 1.6.0 (MINOR)
change: "Add 'sort_by' parameter (default: alphabetical for compatibility)"
documentation: "Now documents sort order explicitly"
```

**Lesson**: Document ALL observable behavior, or treat changes as potentially breaking.

---

### ❌ Anti-Pattern 5: "Version Number Confusion"

**Description**: Using wrong version bump type.

**Example**:
```yaml
# BAD
version: 1.5.0 → 1.5.1 (PATCH)
change: "Renamed get_session_stats to query_session_stats"
reason: "Just a small naming fix"
impact: BREAKING (tool name changed)

# GOOD
version: 1.5.0 → 2.0.0 (MAJOR)
change: "Renamed get_session_stats to query_session_stats"
reason: "Naming consistency (BREAKING CHANGE)"
deprecation: "12-month period in v1.x.x"
```

**Rule**: When in doubt, bump MAJOR. Better safe than breaking users.

---

## 5. Edge Cases and Special Scenarios

### Edge Case 1: Bug Fixes That Change Behavior

**Scenario**: Documented behavior differs from actual behavior. Which is "correct"?

**Answer**: **Documentation is the contract.** Actual behavior is the bug.

**Example**:
```yaml
documented: "stats_only returns only statistics (no detail)"
actual: "stats_only returns statistics AND details"

fix: "Make actual behavior match documentation (remove details)"
classification: PATCH (fixing bug to match docs)
impact: May break users relying on buggy behavior

mitigation:
  - If usage is high: Treat as BREAKING, deprecate
  - If usage is low: Fix as PATCH, note in changelog
```

**Guideline**:
- Check usage of buggy behavior
- If significant usage: Deprecate first (even for bugs)
- If minimal usage: Fix as PATCH, document in changelog

---

### Edge Case 2: Performance "Improvements" That Break Assumptions

**Scenario**: Performance change alters observable behavior.

**Example**:
```yaml
change: "query_tools now returns cached results (faster)"
before: "Always queries live data"
after: "Returns cached data (up to 60s old)"

impact: Potentially BREAKING (users expecting live data)

solution:
  - Add cache_ttl parameter (default: 0 for live data)
  - Document caching behavior explicitly
  - Let users opt in to caching (cache_ttl > 0)
```

**Guideline**: Performance changes are BREAKING if they alter observable behavior (timing, data freshness, order).

---

### Edge Case 3: Security Fixes Requiring Breaking Changes

**Scenario**: Security vulnerability requires immediate fix that breaks compatibility.

**Approach**:
1. Fix IMMEDIATELY (don't wait for deprecation)
2. Use PATCH or MINOR (not MAJOR) if urgent
3. Communicate clearly in release notes
4. Provide migration path (even if accelerated)

**Example**:
```yaml
vulnerability: "jq_filter allows code injection"
fix: "Strict jq validation (rejects invalid syntax)"

version: 1.5.1 (PATCH - security fix)
change: "jq_filter now validates syntax strictly"
impact: May break users with invalid jq filters
rationale: "Security fix - code injection risk"
migration: "Fix invalid jq filters (see guide)"
timeline: "Immediate (security)"
```

**Guideline**: Security > compatibility. Fix first, help users migrate quickly.

---

### Edge Case 4: Deprecating Multiple Tools Simultaneously

**Scenario**: Multiple tools need deprecation at once (e.g., naming consistency sweep).

**Approach**:
1. **Batch deprecation**: Deprecate all in single MINOR release
2. **Unified migration period**: All have same 12-month timeline
3. **Comprehensive migration guide**: Cover all renames at once
4. **Single MAJOR release**: Remove all in same version

**Example**:
```yaml
v1.8.0 (MINOR):
  deprecated:
    - get_session_stats → query_session_stats
    - list_capabilities → query_capabilities
    - cleanup_temp_files → (removed, manual cleanup)

  migration_period: 12 months (all)

v2.0.0 (MAJOR):
  removed:
    - get_session_stats
    - list_capabilities
    - cleanup_temp_files

  migration_guide: "docs/migration/v1-to-v2-bulk-renames.md"
```

**Guideline**: Batch related deprecations to minimize major version churn.

---

## 6. Compatibility Success Metrics

### Metric 1: Breaking Change Rate

**Definition**: # of MAJOR releases / # of total releases

**Target**: <20% (most releases should be MINOR/PATCH)

**Example**:
```yaml
releases:
  - v1.0.0 (MAJOR - initial)
  - v1.1.0 (MINOR)
  - v1.2.0 (MINOR)
  - v1.2.1 (PATCH)
  - v1.3.0 (MINOR)
  - v2.0.0 (MAJOR - breaking changes)

total_releases: 6
major_releases: 2
breaking_rate: 33%  (2/6)
```

### Metric 2: Deprecation Compliance Rate

**Definition**: # of deprecations with full 12-month period / # of total deprecations

**Target**: 100% (all deprecations follow policy)

### Metric 3: Migration Success Rate

**Definition**: % of users who migrated successfully during deprecation period

**Target**: >90%

**Measurement**: Usage of deprecated feature drops to <10% of original by end of migration period

### Impact on V_evolvability

**Component**: backward_compatible_design

**Before**: 0.50 (some patterns allow extension, but risks exist)

**After**: 0.80 (clear patterns, testing, safe evolution)

**Contribution**: +0.06 to V_evolvability (0.30 × 0.20 weight)

---

## 7. Recommendations

### For API Developers

1. **Design for Extension**: Prefer optional parameters over required
2. **Test Backward Compatibility**: Run regression tests before release
3. **Document Defaults**: Make default behavior explicit
4. **Avoid Renames**: Naming consistency is good, but renames are costly
5. **Validate Assumptions**: Check usage before assuming features are unused

### For API Users

1. **Explicit Parameters**: Don't rely on defaults for critical behavior
2. **Version Awareness**: Track which version you're using
3. **Heed Warnings**: Deprecation warnings are migration deadlines
4. **Test Updates**: Test new versions before deploying

### For the Project

1. **Automate Testing**: CI/CD should include compatibility tests
2. **Monitor Usage**: Track tool/parameter usage to assess impact
3. **Clear Communication**: Document every change (changelog, migration guides)

---

**Document Status**: ✅ Ready for Implementation
**Expected Impact**: backward_compatible_design: 0.50 → 0.80
**Next Steps**: Implement compatibility tests → Apply to first changes → Monitor success rate
