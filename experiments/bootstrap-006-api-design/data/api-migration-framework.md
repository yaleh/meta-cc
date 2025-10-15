# API Migration Framework for meta-cc MCP Tools

**Document Version**: 1.0
**Created**: 2025-10-15
**Status**: Proposed
**Applies To**: meta-cc MCP API v1.0.0+

---

## 1. Framework Purpose

### Goals

This framework provides:
1. **Standard migration process** for users upgrading between API versions
2. **Migration tooling requirements** for automation
3. **Migration documentation templates** for consistency
4. **User support approach** for migration assistance
5. **Success metrics** to track migration effectiveness

### Scope

Applies to:
- Tool renames and removals
- Parameter changes
- Response format changes
- Default value changes
- Any change requiring user action

---

## 2. Migration Checklist

### Standard Migration Steps

Every API change requiring migration MUST follow these steps:

#### Pre-Migration (During Deprecation Announcement)

- [ ] **Identify affected users**
  - Query usage statistics
  - Estimate number of affected calls/users
  - Classify impact (high/medium/low)

- [ ] **Design migration path**
  - Define replacement or alternative
  - Create before/after examples
  - Identify automation opportunities

- [ ] **Create migration guide**
  - Write step-by-step instructions
  - Include code examples
  - Document common issues

- [ ] **Set up compatibility shim** (if feasible)
  - Old feature calls new feature internally
  - Add deprecation warning
  - Maintain during migration period

- [ ] **Announce deprecation**
  - Update tool metadata
  - Add changelog entry
  - Publish migration guide
  - Enable runtime warnings

#### During Migration Period (12+ months)

- [ ] **Monitor migration progress**
  - Track deprecated feature usage
  - Identify migration blockers
  - Measure migration rate

- [ ] **Provide user support**
  - Answer migration questions
  - Update migration guide based on feedback
  - Assist with complex migrations

- [ ] **Send reminders**
  - 6-month reminder
  - 3-month reminder
  - 1-month final warning

- [ ] **Assess readiness for removal**
  - Verify 12-month minimum elapsed
  - Check migration rate (target: >90%)
  - Address outstanding blockers

#### Post-Migration (After Removal)

- [ ] **Remove deprecated feature**
  - Delete from codebase
  - Update documentation
  - Archive migration guide

- [ ] **Measure success**
  - Calculate final migration rate
  - Document lessons learned
  - Update framework based on experience

---

## 3. Migration Documentation Templates

### Template 1: Tool Rename Migration

**Use Case**: Renaming a tool (e.g., `get_session_stats` → `query_session_stats`)

```markdown
# Migration Guide: {old_tool_name} → {new_tool_name}

**Status**: Deprecated since v{X.Y.Z}
**Removal**: Scheduled for v{X+1}.0.0 ({date})
**Migration Deadline**: {date} (12 months)

---

## Summary

`{old_tool_name}` is deprecated. Use `{new_tool_name}` instead.

**Reason for change**: {rationale - e.g., naming consistency}

---

## Quick Migration

### Before (Deprecated)
{old_tool_name}({params})

### After (Recommended)
{new_tool_name}({params})

### What Changed
- Tool name: `{old_tool_name}` → `{new_tool_name}`
- Parameters: {unchanged | changed - list changes}
- Response: {unchanged | changed - list changes}

---

## Step-by-Step Migration

### Step 1: Identify Usage
Find all calls to `{old_tool_name}` in your code:
- Search: `{old_tool_name}` (exact match)
- Estimated effort: {X} minutes per occurrence

### Step 2: Replace Tool Name
Replace `{old_tool_name}` with `{new_tool_name}`:
- Find: `{old_tool_name}(`
- Replace: `{new_tool_name}(`
- Automated: Yes (search-and-replace)

### Step 3: Update Parameters (if changed)
{parameter_changes_details | "No parameter changes required"}

### Step 4: Update Response Parsing (if changed)
{response_changes_details | "No response changes required"}

### Step 5: Test
Verify migration:
- [ ] Code compiles/runs without errors
- [ ] Results match expected behavior
- [ ] No deprecation warnings displayed

---

## Common Issues

### Issue 1: {common_issue_description}
**Symptom**: {what_users_see}
**Cause**: {why_it_happens}
**Solution**: {how_to_fix}

### Issue 2: {common_issue_description}
...

---

## Compatibility Period

**Both tools work**: v{X.Y.Z} - v{X}.x.x (until {date})

During this period:
- `{old_tool_name}` still works (with deprecation warning)
- `{new_tool_name}` is the recommended approach
- Both return identical results

**After v{X+1}.0.0**: Only `{new_tool_name}` will work.

---

## Need Help?

- Migration questions: {support_channel}
- Report migration issues: {issue_tracker}
- Migration tooling: {automation_scripts_if_available}

---

**Last Updated**: {date}
**Status**: {Active | Archived}
```

---

### Template 2: Parameter Change Migration

**Use Case**: Adding required parameter or changing parameter default

```markdown
# Migration Guide: {tool_name} Parameter Change

**Parameter**: `{parameter_name}`
**Change**: {old_behavior} → {new_behavior}
**Deprecated**: v{X.Y.Z}
**Effective**: v{X+1}.0.0 ({date})

---

## Summary

The `{parameter_name}` parameter behavior is changing in v{X+1}.0.0.

**Current** (v{X}.x.x): {current_behavior}
**Future** (v{X+1}.0.0+): {new_behavior}

**Action Required**: {explicit_parameter | change_default_usage}

---

## Migration Scenarios

### Scenario 1: You rely on the default value

**Before** (implicit default):
{tool_name}()  # {parameter_name} defaults to {old_default}

**After** (explicit value):
{tool_name}({parameter_name}={old_default})  # Preserve old behavior

**OR** (adopt new default):
{tool_name}()  # Accept new default {new_default}

### Scenario 2: You explicitly set the parameter

**Before**:
{tool_name}({parameter_name}={value})

**After**:
{tool_name}({parameter_name}={value})  # No change needed

---

## Why This Change?

{rationale_for_change}

---

## Deprecation Warnings

Starting v{X.Y.Z}, you'll see this warning if `{parameter_name}` is not explicitly set:

WARNING: '{parameter_name}' parameter not provided.
Currently defaults to '{old_default}'.
In v{X+1}.0.0, default will change to '{new_default}'.
Explicitly set {parameter_name}='{old_default}' to preserve current behavior.

**How to silence the warning**: Explicitly set `{parameter_name}` parameter.

---

## Timeline

- **v{X.Y.Z}** ({date}): Deprecation warning added
- **v{X}.x.x** ({date_range}): 12-month migration period (both defaults work)
- **v{X+1}.0.0** ({date}): New default takes effect

---

**Last Updated**: {date}
```

---

### Template 3: Tool Removal Migration

**Use Case**: Removing a tool entirely

```markdown
# Migration Guide: {tool_name} Removal

**Tool**: `{tool_name}`
**Deprecated**: v{X.Y.Z}
**Removed**: v{X+1}.0.0 ({date})

---

## Summary

`{tool_name}` is being removed in v{X+1}.0.0.

**Reason**: {rationale - e.g., low usage, superseded by better alternative}

---

## Alternatives

### Option 1: Use {replacement_tool} (Recommended)

**Before**:
{tool_name}({params})

**After**:
{replacement_tool}({new_params})

**Differences**:
- {difference_1}
- {difference_2}

**Migration effort**: {easy | moderate | complex}

### Option 2: {alternative_approach}

{description_of_manual_or_alternative_approach}

---

## Why Removal?

{detailed_rationale}

Usage statistics:
- Total calls: {X} ({Y}% of all tool calls)
- Classification: {rarely_used | low_impact}

---

## Migration Steps

### Step 1: Assess Usage
Check if you use `{tool_name}`:
- Search codebase for `{tool_name}`
- Review call frequency

### Step 2: Choose Alternative
- Recommended: {replacement_tool}
- Alternative: {manual_approach}

### Step 3: Migrate Code
{specific_migration_steps}

### Step 4: Verify
- [ ] No calls to `{tool_name}` remain
- [ ] Alternative approach works as expected
- [ ] No deprecation warnings

---

## Timeline

- **v{X.Y.Z}** ({date}): Tool deprecated (still works with warning)
- **v{X}.x.x** ({date_range}): Migration period
- **v{X+1}.0.0** ({date}): Tool removed (calls will fail)

---

**Last Updated**: {date}
```

---

## 4. Migration Tooling Requirements

### Tool 1: Deprecation Warning System

**Purpose**: Notify users at runtime when deprecated features are used

**Requirements**:
- Detect deprecated feature usage
- Display clear warning message
- Include deprecation version, removal version, replacement
- Link to migration guide
- Allow warning suppression (opt-out)

**Example Implementation**:
```python
def deprecated_tool_wrapper(old_name, new_name, deprecated_since, removal_version, migration_url):
    def wrapper(*args, **kwargs):
        print(f"""
⚠️  DEPRECATION WARNING
Tool: {old_name}
Status: Deprecated since v{deprecated_since}
Removal: Scheduled for v{removal_version}
Replacement: Use '{new_name}' instead
Migration: {migration_url}
        """)
        # Call new tool internally
        return new_tool(*args, **kwargs)
    return wrapper
```

---

### Tool 2: Usage Statistics Tracker

**Purpose**: Monitor deprecated feature usage to track migration progress

**Requirements**:
- Track tool call frequency
- Identify deprecated feature usage
- Calculate migration rate over time
- Generate migration progress reports

**Example Report**:
```yaml
migration_progress:
  tool: "get_session_stats"
  deprecated_since: "2025-10-15"
  removal_date: "2026-10-15"

  usage_trend:
    - month: "2025-10"
      calls: 48  # Baseline
      percentage: 100%

    - month: "2025-11"
      calls: 42
      percentage: 87.5%  # 12.5% migrated

    - month: "2026-04"
      calls: 15
      percentage: 31.3%  # 68.7% migrated

    - month: "2026-09"
      calls: 3
      percentage: 6.3%  # 93.7% migrated ✅

  migration_rate: 93.7%
  target: 90%
  status: "On track"
```

---

### Tool 3: Automated Migration Script (Optional)

**Purpose**: Automate simple migrations (e.g., tool renames)

**Requirements**:
- Detect migration scenarios
- Apply transformations (search-and-replace)
- Generate migration report
- Support dry-run mode

**Example Usage**:
```bash
# Dry run (show changes without applying)
meta-cc migrate --from v1.x.x --to v2.0.0 --dry-run

# Apply migration
meta-cc migrate --from v1.x.x --to v2.0.0 --apply

# Migration report
Analyzing code...
Found 5 occurrences of deprecated features:

1. get_session_stats → query_session_stats (3 occurrences)
   - file1.py:42
   - file2.py:103
   - file3.py:67

2. list_capabilities → query_capabilities (2 occurrences)
   - file4.py:28
   - file5.py:91

Apply these changes? [y/N]
```

**Note**: Automated migration is OPTIONAL. Simple renames are easy to automate; complex changes may require manual migration.

---

### Tool 4: Version Compatibility Checker

**Purpose**: Verify code compatibility with target API version

**Requirements**:
- Analyze code for deprecated feature usage
- Check compatibility with specific version
- Generate compatibility report
- Suggest migrations

**Example Usage**:
```bash
# Check compatibility with v2.0.0
meta-cc check-compatibility --target v2.0.0

# Report
Compatibility Check: Current code → v2.0.0

❌ INCOMPATIBLE (3 issues found)

1. Tool removed: get_session_stats (3 occurrences)
   Migration: Use query_session_stats
   Guide: https://docs.meta-cc/migration/get-session-stats

2. Tool removed: list_capabilities (2 occurrences)
   Migration: Use query_capabilities
   Guide: https://docs.meta-cc/migration/list-capabilities

3. Parameter default changed: query_tools.scope
   Current default: "project"
   New default: "session"
   Fix: Explicitly set scope="project" to preserve behavior
   Guide: https://docs.meta-cc/migration/scope-default

Recommended action: Migrate before upgrading to v2.0.0
```

---

## 5. User Support Plan

### Support Channels

1. **Migration Guides** (Primary)
   - Location: `docs/migration/`
   - Format: Markdown
   - Contents: Step-by-step instructions, examples, common issues

2. **Runtime Warnings** (Proactive)
   - Trigger: When deprecated feature used
   - Message: Deprecation status, replacement, migration link

3. **Changelog** (Reference)
   - Location: `CHANGELOG.md`
   - Format: Markdown with semantic versioning structure
   - Contents: All changes with BREAKING CHANGE flags

4. **GitHub Issues** (Interactive)
   - Purpose: Migration questions, bug reports
   - Label: `migration-support`
   - Response SLA: <48 hours

5. **Documentation** (Comprehensive)
   - MCP guide: Mark deprecated tools
   - API reference: Version compatibility matrix
   - FAQ: Common migration questions

---

### Support Timeline

#### At Deprecation Announcement (v X.Y.Z)

**Actions**:
- Publish migration guide
- Enable runtime warnings
- Add changelog entry
- Create GitHub tracking issue

**User expectation**: Clear migration path, 12+ months to migrate

#### 6 Months Before Removal

**Actions**:
- Send reminder (changelog, runtime warning emphasis)
- Check migration progress (usage statistics)
- Address migration blockers

**User expectation**: Migration is possible, support is available

#### 3 Months Before Removal

**Actions**:
- Send urgent reminder
- Escalate warnings (more prominent)
- Reach out to heavy users directly (if identifiable)

**User expectation**: Deadline approaching, urgency clear

#### 1 Month Before Removal

**Actions**:
- Final warning
- Verify migration readiness
- Consider extension if migration rate <90%

**User expectation**: Last chance to migrate

#### At Removal (v X+1.0.0)

**Actions**:
- Remove deprecated feature
- Archive migration guide (keep for reference)
- Monitor for post-removal issues

**User expectation**: Feature gone, migration guide still available

---

### Escalation Path

For users struggling with migration:

**Level 1: Self-Service**
- Read migration guide
- Check changelog
- Review examples

**Level 2: Community Support**
- Ask in GitHub issues
- Search existing issues
- Check FAQ

**Level 3: Direct Support**
- Email maintainer (for complex migrations)
- Schedule consultation (if available)
- Request migration assistance

---

## 6. Success Metrics

### Metric 1: Migration Rate

**Definition**: Percentage of users who migrated by deadline

**Calculation**:
```
migration_rate = (baseline_usage - current_usage) / baseline_usage × 100%
```

**Target**: >90% migration by removal date

**Example**:
```yaml
baseline_usage: 48 calls/month (at deprecation)
current_usage: 3 calls/month (before removal)
migration_rate: (48 - 3) / 48 × 100% = 93.75% ✅
```

---

### Metric 2: Migration Support Quality

**Definition**: User-reported ease of migration

**Measurement**:
- Survey users who migrated
- Ask: "How difficult was the migration?" (1-5 scale)
- Track common issues reported

**Target**: Average difficulty <2.5 (out of 5)

**Example**:
```yaml
survey_responses: 20
ratings:
  very_easy: 8 (1)
  easy: 10 (2)
  moderate: 2 (3)
  difficult: 0 (4)
  very_difficult: 0 (5)

average: (8×1 + 10×2 + 2×3) / 20 = 1.7 ✅ (Easy)
```

---

### Metric 3: Migration Time

**Definition**: Average time to complete migration

**Measurement**: User-reported or estimated based on complexity

**Target**: <1 hour for simple migrations, <1 day for complex

**Example**:
```yaml
migration: "get_session_stats → query_session_stats"
complexity: simple (tool rename only)
estimated_time: 15 minutes per occurrence
average_occurrences: 3
total_time: 45 minutes ✅
```

---

### Metric 4: Post-Migration Issues

**Definition**: Issues reported after migration

**Target**: <5% of users report post-migration issues

**Example**:
```yaml
users_migrated: 100
issues_reported: 3
issue_rate: 3% ✅

issues:
  - "Unexpected response format" (docs unclear)
  - "Performance regression" (unrelated to migration)
  - "Missing parameter" (migration guide incomplete)

actions:
  - Update migration guide (clarify response format)
  - Investigate performance (separate issue)
  - Add missing parameter to guide
```

---

### Impact on V_evolvability

**Component**: migration_support

**Before**: 0.00 (no migration framework)

**After**: 0.60 (guides, tooling requirements, support plan)

**Contribution**: +0.12 to V_evolvability (0.60 × 0.20 weight)

**Note**: 0.60 (not 1.00) because automation tooling is not yet implemented, only specified.

---

## 7. Continuous Improvement

### Post-Migration Review

After each MAJOR release with removals:

1. **Measure Metrics**
   - Calculate migration rate
   - Survey migration difficulty
   - Track post-migration issues

2. **Analyze Blockers**
   - What prevented users from migrating?
   - Which guides were unclear?
   - What tooling would have helped?

3. **Update Framework**
   - Improve migration guide templates
   - Add new tooling requirements
   - Refine support timeline

4. **Document Lessons**
   - What worked well?
   - What would we do differently?
   - What patterns emerged?

---

### Framework Evolution

This framework itself will evolve:

**v1.0** (Current): Templates, tooling requirements, support plan
**v1.1** (Future): Add automation scripts, improved metrics
**v2.0** (Future): Add AI-assisted migration, cross-version compatibility matrix

---

## 8. Recommendations

### For Maintainers

1. **Start Early**: Begin migration planning during deprecation design
2. **Be Clear**: Migration guides should be step-by-step, not high-level
3. **Monitor Progress**: Track usage monthly, address blockers proactively
4. **Extend if Needed**: If migration rate <90%, consider extending deadline

### For Users

1. **Migrate Early**: Don't wait until deadline
2. **Read Guides**: Migration guides have examples and common issues
3. **Test Thoroughly**: Verify migration works before deploying
4. **Report Issues**: If migration is difficult, report so guides can improve

### For the Project

1. **Automate Tracking**: Use tooling to monitor migration progress automatically
2. **Template Everything**: Consistent migration guides reduce confusion
3. **Support Generously**: Good migration support builds user trust
4. **Learn and Improve**: Each migration teaches lessons for next one

---

**Document Status**: ✅ Ready for Implementation
**Expected Impact**: migration_support: 0.00 → 0.60
**Next Steps**: Implement tooling → Create first migration guide → Monitor metrics
