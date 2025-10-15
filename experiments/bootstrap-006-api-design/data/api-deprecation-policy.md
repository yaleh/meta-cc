# API Deprecation Policy for meta-cc MCP Tools

**Document Version**: 1.0
**Created**: 2025-10-15
**Status**: Proposed
**Applies To**: meta-cc MCP API v1.0.0+

---

## 1. Purpose and Scope

### Purpose

This policy defines:
- What constitutes a breaking change
- How and when to deprecate API features
- Communication requirements for deprecations
- Migration support expectations
- Timeline requirements

### Scope

Applies to:
- MCP tool additions, modifications, and removals
- Tool parameter changes
- Response format changes
- Behavior changes

Does NOT apply to:
- Internal implementation changes (no API surface impact)
- Documentation clarifications (non-behavioral)
- Bug fixes that restore documented behavior

---

## 2. Breaking Change Definition

### What is a Breaking Change?

A change that **requires user action** or **alters existing behavior** without explicit opt-in.

### Breaking Change Categories

#### 🔴 **Critical Breaking Changes** (Always Require Deprecation)

1. **Tool Removal**
   - Example: Removing `cleanup_temp_files` tool
   - Impact: Users calling this tool will get "unknown tool" errors

2. **Required Parameter Addition**
   - Example: Making `pattern` required in `query_files` (previously optional)
   - Impact: Existing calls without this parameter fail

3. **Required Parameter Removal**
   - Example: Removing `pattern` from `query_user_messages`
   - Impact: User code breaks

4. **Parameter Default Value Change** (Behavioral)
   - Example: Changing `scope` default from `"project"` to `"session"`
   - Impact: Silent behavior change for users relying on default

5. **Tool Rename**
   - Example: `get_session_stats` → `query_session_stats`
   - Impact: Tool name in user code must change

6. **Response Format Structure Change**
   - Example: Changing from JSONL to JSON array
   - Impact: User parsers break

7. **Behavior Change** (Same Inputs, Different Outputs)
   - Example: `jq_filter` now validates jq syntax (previously accepted any string)
   - Impact: Previously accepted inputs may fail

#### 🟡 **Moderate Breaking Changes** (Require Notice, May Not Need Full Deprecation)

1. **Optional Parameter Default Change** (Non-Behavioral)
   - Example: Changing `output_format` default from `"jsonl"` to `"json"` (if both work)
   - Impact: Output changes but still valid

2. **Error Message Changes** (Format Only)
   - Example: Changing error code from `-32603` to `-32001`
   - Impact: Users parsing error codes must update

3. **Validation Tightening**
   - Example: `jq_filter` rejects invalid jq (previously passed through)
   - Impact: Invalid inputs now fail (was always wrong, now enforced)

#### 🟢 **Non-Breaking Changes** (No Deprecation Needed)

1. **Tool Addition**
   - Example: Adding `query_insights` tool
   - Impact: None on existing tools

2. **Optional Parameter Addition** (With Safe Default)
   - Example: Adding `include_archived: false` to `query_files`
   - Impact: None (default preserves behavior)

3. **Response Field Addition** (Additive Only)
   - Example: Adding `execution_time_ms` to response
   - Impact: None (existing fields unchanged)

4. **Validation Relaxation**
   - Example: `pattern` now accepts regex or glob (previously regex only)
   - Impact: None (existing patterns still work)

5. **Bug Fixes** (Restore Documented Behavior)
   - Example: Fixing `stats_only` calculation error
   - Impact: None (was always supposed to work this way)

6. **Documentation Improvements**
   - Example: Clarifying parameter descriptions
   - Impact: None (behavior unchanged)

---

## 3. Deprecation Process

### Process Overview

```
┌─────────────────────────────────────────────────────────┐
│ 1. Identify Breaking Change                             │
│    ↓                                                     │
│ 2. Design Migration Path                                │
│    ↓                                                     │
│ 3. Announce Deprecation (v N.x.x)                       │
│    ↓                                                     │
│ 4. Migration Period (12+ months)                        │
│    ↓                                                     │
│ 5. Remove Feature (v N+1.0.0)                           │
└─────────────────────────────────────────────────────────┘
```

### Step-by-Step Process

#### Step 1: Identify Breaking Change

**Actions**:
- Classify change using breaking change categories
- Assess user impact (usage statistics)
- Determine if deprecation is required

**Example**:
```yaml
change: "Rename get_session_stats to query_session_stats"
category: "Tool Rename (🔴 Critical Breaking Change)"
usage: "48 calls in project history (frequently used)"
deprecation_required: true
```

#### Step 2: Design Migration Path

**Actions**:
- Design replacement or alternative
- Create migration guide
- Identify automation opportunities

**Example**:
```yaml
migration_path:
  replacement: "query_session_stats"
  compatibility_period: "Both tools work during v1.x.x"
  migration_guide: "docs/migration/get-session-stats.md"
  automation: "Search-and-replace tool name"
```

#### Step 3: Announce Deprecation

**Channels** (ALL required):
1. **Tool Metadata**: Mark tool as deprecated
2. **Changelog**: Add DEPRECATED entry
3. **Documentation**: Update MCP guide
4. **Runtime Warning**: Display warning when tool called

**Timeline**: Deprecation announced in MINOR release (e.g., v1.5.0)

**Example Metadata**:
```json
{
  "tool": "get_session_stats",
  "version": "1.5.0",
  "deprecated": true,
  "deprecated_since": "1.5.0",
  "deprecated_date": "2025-10-15",
  "removal_planned": "2.0.0",
  "removal_date": "2026-10-15 (estimated)",
  "replacement": "query_session_stats",
  "deprecation_reason": "Naming consistency - standardizing on query_* prefix",
  "migration_guide": "https://docs.meta-cc/migration/get-session-stats"
}
```

**Example Runtime Warning**:
```
⚠️  DEPRECATION WARNING
Tool: get_session_stats
Status: Deprecated since v1.5.0
Removal: Scheduled for v2.0.0 (October 2026)
Replacement: Use 'query_session_stats' instead
Migration: https://docs.meta-cc/migration/get-session-stats
This warning will appear until migration is complete.
```

#### Step 4: Migration Period

**Duration**: Minimum 12 months from deprecation announcement

**Requirements**:
- Both old and new features work (if feasible)
- Migration guide available
- Runtime warnings active
- Support for migration questions

**Monitoring**:
- Track usage of deprecated feature
- Monitor migration progress
- Address migration blockers

**Example Timeline**:
```
2025-10-15: v1.5.0 released - get_session_stats deprecated
2025-10 - 2026-09: 12-month migration period
  - Both get_session_stats and query_session_stats work
  - Warnings displayed on every get_session_stats call
  - Migration support provided
2026-10-15: v2.0.0 released - get_session_stats removed
```

#### Step 5: Remove Feature

**Trigger**: MAJOR version bump (e.g., v2.0.0)

**Requirements**:
- Minimum 12 months elapsed since deprecation
- Migration guide published
- Changelog updated with BREAKING CHANGE
- Release notes highlight removals

**Example Removal**:
```yaml
version: "2.0.0"
release_date: "2026-10-15"
breaking_changes:
  - removed: "get_session_stats tool"
    deprecated_since: "1.5.0 (2025-10-15)"
    migration_path: "Use query_session_stats"
    guide: "https://docs.meta-cc/migration/get-session-stats"
```

---

## 4. Notice Periods

### Minimum Notice Periods by Change Type

| Change Type | Minimum Notice | Recommended Notice | Rationale |
|-------------|---------------|-------------------|-----------|
| Tool Removal | 12 months | 12 months | Users need time to update code |
| Tool Rename | 12 months | 12 months | Code changes required |
| Required Parameter Change | 12 months | 12 months | Breaking user code |
| Default Value Change | 12 months | 12 months | Silent behavior change risk |
| Response Format Change | 12 months | 18 months | Parser updates complex |
| Behavior Change | 6 months | 12 months | Depends on severity |
| Validation Tightening | 3 months | 6 months | Users should fix anyway |

### Notice Period Calculation

**Start Date**: Version release date where deprecation announced (e.g., v1.5.0 on 2025-10-15)

**End Date**: Minimum start date + notice period (e.g., 2025-10-15 + 12 months = 2026-10-15)

**Removal Version**: First MAJOR release after end date (e.g., v2.0.0)

---

## 5. Communication Requirements

### Required Communications

For every deprecation, ALL of the following are REQUIRED:

#### 1. Tool Metadata

```json
{
  "deprecated": true,
  "deprecated_since": "version",
  "replacement": "alternative_tool_or_null",
  "migration_guide": "url"
}
```

#### 2. Changelog Entry

```markdown
## [1.5.0] - 2025-10-15

### Deprecated
- **get_session_stats**: Use `query_session_stats` instead. Will be removed in v2.0.0 (October 2026).
  - Reason: Naming consistency (standardizing on query_* prefix)
  - Migration: https://docs.meta-cc/migration/get-session-stats
```

#### 3. Documentation Update

- MCP guide: Mark tool as deprecated
- Migration guide: Create step-by-step migration instructions
- FAQ: Add deprecation to known issues

#### 4. Runtime Warning

- Display warning when deprecated tool called
- Include: deprecation version, removal version, replacement, migration link
- Warning must be visible but not block functionality

### Optional Communications (Recommended)

- Blog post for major deprecations
- GitHub issue for user feedback
- Email notification for known heavy users
- Release notes highlighting deprecations

---

## 6. Migration Support

### Required Migration Support

For every deprecation, provide:

#### 1. Migration Guide

**Format**: Markdown document

**Location**: `docs/migration/{deprecated-feature}.md`

**Contents**:
- What is deprecated
- Why it's deprecated
- Replacement feature
- Step-by-step migration
- Before/after code examples
- Common migration issues

**Example**:
```markdown
# Migration: get_session_stats → query_session_stats

## Summary
`get_session_stats` is deprecated. Use `query_session_stats` instead.

## Before (Deprecated)
get_session_stats()

## After (Recommended)
query_session_stats(scope="session")

## Changes Required
1. Rename tool: get_session_stats → query_session_stats
2. Add scope parameter: scope="session" (explicit)

## Rationale
Standardizing on query_* prefix for consistency.
```

#### 2. Compatibility Shim (If Feasible)

**Definition**: Keep old feature working during migration period.

**Example**: `get_session_stats` internally calls `query_session_stats(scope="session")` + warning

**Benefits**:
- Users can migrate at their own pace
- No immediate breakage
- Warnings remind users to migrate

**When NOT Feasible**:
- Structural changes (response format overhaul)
- Performance reasons (shim too costly)
- Security reasons (old behavior insecure)

#### 3. Automated Migration Tools (If Feasible)

**Examples**:
- Search-and-replace scripts
- Parameter migration scripts
- Response format converters

**Benefit**: Reduces migration effort

---

## 7. Exceptions and Special Cases

### When Deprecation May Be Waived

**Security Vulnerabilities**:
- Critical security issues may require immediate breaking changes
- Still provide migration path, but timeline shortened to 1-3 months
- Example: Parameter allows code injection → must be fixed immediately

**Unused Features**:
- Features with zero or near-zero usage may have shortened deprecation (3-6 months)
- Requires data: usage statistics showing <1% usage
- Example: `cleanup_temp_files` rarely used → shorter deprecation acceptable

**Alpha/Beta Features**:
- Features in alpha/beta (v0.x.x) may be changed without deprecation
- Clear labeling required ("BETA - API may change")

### Accelerated Deprecation Process

For critical issues, accelerated process:
1. Immediate deprecation warning (next PATCH release)
2. 3-6 month migration period (vs. 12 months)
3. Removal in next MINOR or MAJOR (vs. only MAJOR)

**Requirements**:
- Clear rationale (security, data loss risk, etc.)
- Accelerated timeline communicated upfront
- Extra migration support (direct user outreach)

---

## 8. Deprecation Checklist

### Pre-Deprecation

- [ ] Identify breaking change
- [ ] Classify severity (🔴 Critical, 🟡 Moderate, 🟢 Non-Breaking)
- [ ] Analyze usage statistics (how many users affected?)
- [ ] Design replacement or migration path
- [ ] Create migration guide
- [ ] Determine notice period (minimum 12 months for critical)

### Deprecation Announcement (v N.x.x)

- [ ] Add deprecation metadata to tool
- [ ] Update changelog with DEPRECATED section
- [ ] Update MCP guide documentation
- [ ] Implement runtime warning
- [ ] Publish migration guide
- [ ] Create GitHub issue for tracking

### Migration Period (12+ months)

- [ ] Monitor deprecated feature usage
- [ ] Track migration progress
- [ ] Address migration blockers
- [ ] Provide migration support
- [ ] Send reminder communications (6 months, 3 months before removal)

### Removal (v N+1.0.0)

- [ ] Verify minimum notice period elapsed
- [ ] Update changelog with BREAKING CHANGE
- [ ] Remove deprecated feature from code
- [ ] Update documentation (remove feature)
- [ ] Release notes highlight removal
- [ ] Archive migration guide (keep for reference)

---

## 9. Example Deprecation Scenarios

### Scenario 1: Tool Rename (Critical Breaking Change)

**Change**: `get_session_stats` → `query_session_stats`

**Timeline**:
```
v1.5.0 (2025-10-15):
  - Add query_session_stats (new)
  - Deprecate get_session_stats
  - get_session_stats → calls query_session_stats + warning

v1.5.0 - v1.x.x (12 months):
  - Both tools work
  - Migration period

v2.0.0 (2026-10-15):
  - Remove get_session_stats
  - Only query_session_stats remains
```

**Communications**:
- Metadata: deprecated=true, replacement="query_session_stats"
- Changelog: DEPRECATED section
- Runtime warning: Every get_session_stats call
- Migration guide: docs/migration/get-session-stats.md

### Scenario 2: Default Value Change (Critical Breaking Change)

**Change**: `scope` default from `"project"` to `"session"`

**Timeline**:
```
v1.6.0 (2025-11-01):
  - Add warning if scope not explicitly provided
  - "Future versions will default to 'session'. Explicitly set scope='project' to preserve current behavior."

v1.6.0 - v1.x.x (12 months):
  - Default still "project"
  - Warnings encourage explicit scope

v2.0.0 (2026-11-01):
  - Change default to "session"
  - Users who set scope="project" explicitly unaffected
```

**Communications**:
- Warning: When scope parameter omitted
- Changelog: Note future default change
- Migration guide: Explain why change is needed, how to preserve behavior

### Scenario 3: Removing Unused Tool (Moderate Breaking Change)

**Change**: Remove `cleanup_temp_files` (rarely used)

**Timeline**:
```
v1.7.0 (2025-12-01):
  - Deprecate cleanup_temp_files
  - Provide manual cleanup documentation

v1.7.0 - v1.x.x (6 months - shortened due to low usage):
  - Tool still works
  - Warning displayed

v2.0.0 (2026-06-01):
  - Remove cleanup_temp_files
  - Direct users to manual cleanup docs
```

**Communications**:
- Metadata: deprecated=true, replacement=null (manual process)
- Usage data: Show <1% usage justifies shorter timeline
- Migration guide: Manual cleanup steps

---

## 10. Success Metrics

### Deprecation Process Quality

**Metrics**:
1. **Migration Rate**: % of users migrated before removal
   - Target: >90% migration by removal date

2. **Breaking Incident Rate**: # of users broken by removal (didn't migrate)
   - Target: <10% of affected users

3. **Notice Period Compliance**: Actual notice period ÷ minimum required
   - Target: ≥1.0 (always meet minimum)

4. **Migration Support Quality**: User-reported migration difficulty
   - Target: <20% report difficulty migrating

### Impact on V_evolvability

**Before**: has_deprecation_policy = 0.00 (no policy)

**After**: has_deprecation_policy = 1.00 (comprehensive policy)

**Contribution to V_evolvability**: +0.20 (one of five components)

---

## 11. Recommendations

### Best Practices

1. **Prefer Additive Changes**: Design APIs to allow additions without deprecation
2. **Early Warning**: Announce deprecations as early as possible (>12 months if known)
3. **Clear Communication**: Err on side of over-communication
4. **Monitor Usage**: Track deprecated feature usage to assess migration
5. **Provide Tools**: Automate migration where possible

### Anti-Patterns to Avoid

1. ❌ **Silent Breakage**: Never break without warning
2. ❌ **Insufficient Notice**: Don't rush deprecations (<12 months for critical)
3. ❌ **Missing Migration Paths**: Always provide alternatives
4. ❌ **Ignoring Feedback**: Address user migration blockers
5. ❌ **Inconsistent Timelines**: Follow policy consistently

---

**Document Status**: ✅ Ready for Implementation
**Expected Impact**: has_deprecation_policy: 0.00 → 1.00
**Next Steps**: Approve policy → Implement tooling → Apply to first deprecation
