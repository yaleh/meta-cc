# Pattern Application Checklist

Use this checklist when applying refactoring patterns to ensure systematic execution and verification.

## Pre-Execution Phase

### Pattern Selection
- [ ] Identified the problem to solve
- [ ] Selected appropriate pattern from catalog
- [ ] Verified pattern prerequisites are met
- [ ] Read pattern documentation completely
- [ ] Understood pattern variations

### Context Assessment
- [ ] Verified pattern applies to this situation
- [ ] Checked for edge cases specific to this codebase
- [ ] Identified potential conflicts with existing code
- [ ] Reviewed similar past refactorings (if any)

### Risk Assessment
- [ ] Evaluated task complexity (LOW/MODERATE/HIGH)
- [ ] Evaluated safety (LOW/MODERATE/HIGH)
- [ ] Estimated effort (hours, lines, files)
- [ ] Calculated priority score
- [ ] Decided on P1/P2/P3 priority level

### Baseline Capture
- [ ] Ran full test suite (captured results)
- [ ] Ran linter/static analyzer (captured results)
- [ ] Measured code coverage (captured baseline)
- [ ] Documented current line counts
- [ ] Documented current duplication metrics (if applicable)
- [ ] Took git commit snapshot (note SHA)

### Planning
- [ ] Created task document (using template)
- [ ] Defined success criteria (quantitative + qualitative)
- [ ] Listed verification steps
- [ ] Created rollback plan
- [ ] Allocated time budget
- [ ] Notified team (if collaborative project)

## Execution Phase

### Pattern-Specific Steps

Follow the pattern's step-by-step procedure from the methodology document.

#### Pattern 1: Verify Before Remove
- [ ] Step 1: Run static analyzer on specific scope
- [ ] Step 2: Check test coverage for target code
- [ ] Step 3: Search for references across project
- [ ] Step 4: Verify with runtime analysis (if applicable)
- [ ] Step 5: Document verification results
- [ ] Step 6: Make decision (remove or keep)

#### Pattern 2: InputSchema Builder Extraction
- [ ] Step 1: Identify duplication patterns
- [ ] Step 2: Extract smallest reusable unit first
- [ ] Step 3: Create helper function
- [ ] Step 4: Test helper in isolation
- [ ] Step 5: Refactor one usage
- [ ] Step 6: Run tests, verify behavioral equivalence
- [ ] Step 7: Refactor remaining usages incrementally
- [ ] Step 8: Remove old code

#### Pattern 3: Risk-Based Task Prioritization
- [ ] Step 1: List all candidate tasks
- [ ] Step 2: Assess value for each (0.0-1.0)
- [ ] Step 3: Assess safety for each (0.0-1.0)
- [ ] Step 4: Estimate effort for each (0.0-1.0)
- [ ] Step 5: Calculate priority = (value × safety) / effort
- [ ] Step 6: Sort by priority descending
- [ ] Step 7: Select P1 tasks (priority ≥ 1.0)
- [ ] Step 8: Skip P3 tasks (priority < 0.5) if time-constrained

#### Pattern 4: Incremental Test Addition
- [ ] Step 1: Identify low-coverage package
- [ ] Step 2: Create test file (naming convention)
- [ ] Step 3: Write test for first function (success case)
- [ ] Step 4: Run test, verify passes
- [ ] Step 5: Add failure case test
- [ ] Step 6: Add edge case tests
- [ ] Step 7: Repeat for remaining functions
- [ ] Step 8: Measure coverage improvement

### Continuous Verification
- [ ] Run tests after each logical change
- [ ] Check for compilation errors immediately
- [ ] Verify linter passes after each file edit
- [ ] Review diffs before committing
- [ ] Check performance (if applicable)

### Progress Tracking
- [ ] Update task status regularly
- [ ] Document issues as they arise
- [ ] Note time spent on each step
- [ ] Capture lessons learned in real-time

## Post-Execution Phase

### Verification
- [ ] All tests pass (compare to baseline)
- [ ] Linter/static analyzer passes (no new issues)
- [ ] Code coverage maintained or improved
- [ ] No compilation errors or warnings
- [ ] Performance not degraded (if measured)
- [ ] Documentation updated (if needed)

### Success Criteria Check
- [ ] Quantitative criteria met (lines reduced, coverage improved, etc.)
- [ ] Qualitative criteria met (readability, maintainability)
- [ ] No unintended side effects
- [ ] Rollback not needed

### Measurement
- [ ] Calculate actual ΔV (code quality, maintainability, safety, effort)
- [ ] Measure actual lines changed
- [ ] Measure actual time spent
- [ ] Compare actual vs. estimated
- [ ] Document variance reasons

### Documentation
- [ ] Updated task template with actual results
- [ ] Documented issues encountered
- [ ] Documented lessons learned
- [ ] Updated methodology (if pattern needed adaptation)
- [ ] Created before/after examples (for future reference)

### Code Review (if applicable)
- [ ] Created pull request with clear description
- [ ] Linked to task document
- [ ] Responded to reviewer feedback
- [ ] Merged after approval

### Knowledge Sharing
- [ ] Shared results with team
- [ ] Updated team documentation
- [ ] Added to pattern library (if new variation discovered)
- [ ] Scheduled retrospective (if needed)

## Rollback Procedure

If something goes wrong:

- [ ] Stop execution immediately
- [ ] Document what went wrong
- [ ] Check if quick fix is possible (< 30 minutes)
- [ ] If no quick fix: execute rollback plan
- [ ] git revert or git reset to baseline SHA
- [ ] Verify tests pass after rollback
- [ ] Analyze root cause
- [ ] Re-plan with different approach or pattern
- [ ] Update risk assessment based on learnings

## Pattern Composition

If applying multiple patterns:

- [ ] Identify pattern dependencies
- [ ] Sequence patterns in correct order
- [ ] Apply Pattern 3 (Risk Prioritization) first
- [ ] Apply Pattern 1 (Verify Before Remove) before any deletions
- [ ] Apply Pattern 4 (Incremental Test Addition) continuously
- [ ] Apply Pattern 2 (Builder Extraction) after verification

## Common Pitfalls to Avoid

- [ ] ❌ Don't skip baseline capture
- [ ] ❌ Don't refactor without tests
- [ ] ❌ Don't bulk-change multiple files simultaneously
- [ ] ❌ Don't assume duplication without verification
- [ ] ❌ Don't skip verification steps
- [ ] ❌ Don't ignore linter warnings
- [ ] ❌ Don't commit without review
- [ ] ❌ Don't force-apply patterns that don't fit

## Checklist Usage Notes

1. **Before starting**: Print or copy this checklist
2. **During execution**: Check off items as completed
3. **If blocked**: Document blocker and pause execution
4. **After completion**: Archive checklist with task documentation
5. **For review**: Use completed checklist as audit trail

## Customization

Adapt this checklist for your project by:
- Adding project-specific verification steps
- Adjusting pattern steps to match your tools
- Including team-specific approval requirements
- Adding integration with issue tracking systems
- Including CI/CD pipeline checks

---

**Version**: 1.0
**Last Updated**: 2025-10-16
**Source**: bootstrap-004-refactoring-guide methodology
