# Incremental Commit Protocol

**Purpose**: Ensure clean, revertible git history through disciplined incremental commits

**When to Use**: During ALL refactoring work

**Origin**: Iteration 1 - Problem E3 (No Incremental Commit Discipline)

---

## Core Principle

**Every refactoring step = One commit with passing tests**

**Benefits**:
- **Rollback**: Can revert any single change easily
- **Review**: Small commits easier to review
- **Bisect**: Can use `git bisect` to find which change caused issue
- **Collaboration**: Easy to cherry-pick or rebase individual changes
- **Safety**: Never have large uncommitted work at risk of loss

---

## Commit Frequency Rule

**COMMIT AFTER**:
- Every refactoring step (Extract Method, Rename, Simplify Conditional)
- Every test addition
- Every passing test run after code change
- Approximately every 5-10 minutes of work
- Before taking a break or switching context

**DO NOT COMMIT**:
- While tests are failing (except for WIP commits on feature branches)
- Large batches of changes (>200 lines in single commit)
- Multiple unrelated changes together

---

## Commit Message Convention

### Format

```
<type>(<scope>): <subject>

[optional body]

[optional footer]
```

### Types for Refactoring

| Type | When to Use | Example |
|------|-------------|---------|
| `refactor` | Restructuring code without behavior change | `refactor(sequences): extract collectTimestamps helper` |
| `test` | Adding or modifying tests | `test(sequences): add edge cases for calculateTimeSpan` |
| `docs` | Adding/updating GoDoc comments | `docs(sequences): document calculateTimeSpan parameters` |
| `style` | Formatting, naming (no logic change) | `style(sequences): rename ts to timestamp` |
| `perf` | Performance improvement | `perf(sequences): optimize timestamp collection loop` |

### Scope

**Use package or file name**:
- `sequences` (for internal/query/sequences.go)
- `context` (for internal/query/context.go)
- `file_access` (for internal/query/file_access.go)
- `query` (for changes across multiple files in package)

### Subject Line Rules

**Format**: `<verb> <what> [<pattern>]`

**Verbs**:
- `extract`: Extract Method pattern
- `inline`: Inline Method pattern
- `simplify`: Simplify Conditionals pattern
- `rename`: Rename pattern
- `move`: Move Method/Field pattern
- `add`: Add tests, documentation
- `remove`: Remove dead code, duplication
- `update`: Update existing code/tests

**Examples**:
- ✅ `refactor(sequences): extract collectTimestamps helper`
- ✅ `refactor(sequences): simplify timestamp filtering logic`
- ✅ `refactor(sequences): rename ts to timestamp for clarity`
- ✅ `test(sequences): add edge cases for empty occurrences`
- ✅ `docs(sequences): document calculateSequenceTimeSpan return value`

**Avoid**:
- ❌ `fix bugs` (vague, no scope)
- ❌ `refactor calculateSequenceTimeSpan` (no scope, unclear what changed)
- ❌ `WIP` (not descriptive, avoid on main branch)
- ❌ `refactor: various changes` (not specific)

### Body (Optional but Recommended)

**When to add body**:
- Change is not obvious from subject
- Multiple related changes in one commit
- Need to explain WHY (not WHAT)

**Example**:
```
refactor(sequences): extract collectTimestamps helper

Reduces complexity of calculateSequenceTimeSpan from 10 to 7.
Extracted timestamp collection logic to dedicated helper for clarity.
All tests pass, coverage maintained at 85%.
```

### Footer (For Tracking)

**Pattern**: `Pattern: <pattern-name>`

**Examples**:
```
refactor(sequences): extract collectTimestamps helper

Pattern: Extract Method
```

```
test(sequences): add edge cases for calculateTimeSpan

Pattern: Characterization Tests
```

---

## Commit Workflow (Step-by-Step)

### Before Starting Refactoring

**1. Ensure Clean Baseline**

```bash
git status
```

**Checklist**:
- [ ] No uncommitted changes: `nothing to commit, working tree clean`
- [ ] If dirty: Stash or commit before starting: `git stash` or `git commit`

**2. Create Refactoring Branch** (optional but recommended)

```bash
git checkout -b refactor/calculate-sequence-timespan
```

**Checklist**:
- [ ] Branch created: `refactor/<descriptive-name>`
- [ ] On correct branch: `git branch` shows current branch

---

### During Refactoring (Per Step)

**For Each Refactoring Step**:

#### 1. Make Single Change

- Focused, minimal change (e.g., extract one helper method)
- No unrelated changes in same commit

#### 2. Run Tests

```bash
go test ./internal/query/... -v
```

**Checklist**:
- [ ] All tests pass: PASS / FAIL
- [ ] If FAIL: Fix issue before committing

#### 3. Stage Changes

```bash
git add internal/query/sequences.go internal/query/sequences_test.go
```

**Checklist**:
- [ ] Only relevant files staged: `git status` shows green files
- [ ] No unintended files: Review `git diff --cached`

**Review Staged Changes**:
```bash
git diff --cached
```

**Verify**:
- [ ] Changes are what you intended
- [ ] No debug code, commented code, or temporary changes
- [ ] No unrelated changes sneaked in

#### 4. Commit with Descriptive Message

```bash
git commit -m "refactor(sequences): extract collectTimestamps helper"
```

**Or with body**:
```bash
git commit -m "refactor(sequences): extract collectTimestamps helper

Reduces complexity from 10 to 7.
Extracts timestamp collection logic to dedicated helper.

Pattern: Extract Method"
```

**Checklist**:
- [ ] Commit message follows convention
- [ ] Commit hash: _______________ (from `git log -1 --oneline`)
- [ ] Commit is small (<200 lines): `git show --stat`

#### 5. Verify Commit

```bash
git log -1 --stat
```

**Checklist**:
- [ ] Commit message correct
- [ ] Files changed correct
- [ ] Line count reasonable (<200 insertions + deletions)

**Repeat for each refactoring step**

---

### After Refactoring Complete

**1. Review Commit History**

```bash
git log --oneline
```

**Checklist**:
- [ ] Each commit is small, focused
- [ ] Each commit message is descriptive
- [ ] Commits tell a story of refactoring progression
- [ ] No "fix typo" or "oops" commits (if any, squash them)

**2. Run Final Test Suite**

```bash
go test ./... -v
```

**Checklist**:
- [ ] All tests pass
- [ ] Test coverage: `go test -cover ./internal/query/...`
- [ ] Coverage ≥85%: YES / NO

**3. Verify Each Commit Independently** (optional but good practice)

```bash
git rebase -i HEAD~N  # N = number of commits
# For each commit:
git checkout <commit-hash>
go test ./internal/query/...
```

**Checklist**:
- [ ] Each commit has passing tests: YES / NO
- [ ] Each commit is a valid state: YES / NO
- [ ] If any commit fails tests: Reorder or squash commits

---

## Commit Size Guidelines

### Ideal Commit Size

| Metric | Target | Max |
|--------|--------|-----|
| **Lines changed** | 20-50 | 200 |
| **Files changed** | 1-2 | 5 |
| **Time to review** | 2-5 min | 15 min |
| **Complexity change** | -1 to -3 | -5 |

**Rationale**:
- Small commits easier to review
- Small commits easier to revert
- Small commits easier to understand in history

### When Commit is Too Large

**Signs**:
- >200 lines changed
- >5 files changed
- Commit message says "and" (doing multiple things)
- Hard to write descriptive subject (too complex)

**Fix**:
- Break into multiple smaller commits:
  ```bash
  git reset HEAD~1  # Undo last commit, keep changes
  # Stage and commit parts separately
  git add <file1>
  git commit -m "refactor: <first change>"
  git add <file2>
  git commit -m "refactor: <second change>"
  ```

- Or use interactive staging:
  ```bash
  git add -p <file>  # Stage hunks interactively
  git commit -m "refactor: <specific change>"
  ```

---

## Rollback Scenarios

### Scenario 1: Last Commit Was Mistake

**Undo last commit, keep changes**:
```bash
git reset HEAD~1
```

**Checklist**:
- [ ] Commit removed from history: `git log`
- [ ] Changes still in working directory: `git status`
- [ ] Can re-commit differently: `git add` + `git commit`

**Undo last commit, discard changes**:
```bash
git reset --hard HEAD~1
```

**WARNING**: This DELETES changes permanently
- [ ] Confirm you want to lose changes: YES / NO
- [ ] Backup created if needed: YES / NO / N/A

---

### Scenario 2: Need to Revert Specific Commit

**Revert a commit** (keeps history, creates new commit undoing changes):
```bash
git revert <commit-hash>
```

**Checklist**:
- [ ] Commit hash identified: _______________
- [ ] Revert commit created: `git log -1`
- [ ] Tests pass after revert: PASS / FAIL

**Example**:
```bash
# Revert the "extract helper" commit
git log --oneline  # Find commit hash
git revert abc123  # Revert that commit
git commit -m "revert: extract collectTimestamps helper

Tests failed due to nil pointer. Rolling back to investigate.

Pattern: Rollback"
```

---

### Scenario 3: Multiple Commits Need Rollback

**Revert range of commits**:
```bash
git revert <oldest-commit>..<newest-commit>
```

**Or reset to earlier state**:
```bash
git reset --hard <commit-hash>
```

**Checklist**:
- [ ] Identified rollback point: <commit-hash>
- [ ] Confirmed losing commits OK: YES / NO
- [ ] Branch backed up if needed: `git branch backup-$(date +%Y%m%d)`
- [ ] Tests pass after rollback: PASS / FAIL

---

## Clean History Practices

### Practice 1: Squash Fixup Commits

**Scenario**: Made small "oops" commits (typo fix, forgot file)

**Before Pushing** (local history only):
```bash
git rebase -i HEAD~N  # N = number of commits to review
# Mark fixup commits as "fixup" or "squash"
# Save and close
```

**Example**:
```
pick abc123 refactor: extract collectTimestamps helper
fixup def456 fix: forgot to commit test file
pick ghi789 refactor: extract findMinMax helper
fixup jkl012 fix: typo in variable name
```

**After rebase**:
```
abc123 refactor: extract collectTimestamps helper
ghi789 refactor: extract findMinMax helper
```

**Checklist**:
- [ ] Fixup commits squashed: YES / NO
- [ ] History clean: `git log --oneline`
- [ ] Tests still pass: PASS / FAIL

---

### Practice 2: Reorder Commits Logically

**Scenario**: Commits out of logical order (test commit before code commit)

**Reorder with Interactive Rebase**:
```bash
git rebase -i HEAD~N
# Reorder lines to desired sequence
# Save and close
```

**Example**:
```
# Before:
pick abc123 refactor: extract helper
pick def456 test: add edge case tests
pick ghi789 docs: add GoDoc comments

# After (logical order):
pick def456 test: add edge case tests
pick abc123 refactor: extract helper
pick ghi789 docs: add GoDoc comments
```

**Checklist**:
- [ ] Commits reordered logically: YES / NO
- [ ] Each commit still has passing tests: VERIFY
- [ ] History makes sense: `git log --oneline`

---

## Git Hooks for Enforcement

### Pre-Commit Hook (Prevent Committing Failing Tests)

**Create `.git/hooks/pre-commit`**:
```bash
#!/bin/bash
# Run tests before allowing commit
go test ./... > /dev/null 2>&1
if [ $? -ne 0 ]; then
    echo "❌ Tests failing. Fix tests before committing."
    echo "Run 'go test ./...' to see failures."
    echo ""
    echo "To commit anyway (NOT RECOMMENDED):"
    echo "  git commit --no-verify"
    exit 1
fi

echo "✅ Tests pass. Proceeding with commit."
exit 0
```

**Make executable**:
```bash
chmod +x .git/hooks/pre-commit
```

**Checklist**:
- [ ] Pre-commit hook installed: YES / NO
- [ ] Hook prevents failing test commits: VERIFY
- [ ] Hook can be bypassed if needed: `--no-verify` works

---

### Commit-Msg Hook (Enforce Commit Message Convention)

**Create `.git/hooks/commit-msg`**:
```bash
#!/bin/bash
# Validate commit message format
commit_msg_file=$1
commit_msg=$(cat "$commit_msg_file")

# Pattern: type(scope): subject
pattern="^(refactor|test|docs|style|perf)\([a-z_]+\): .{10,}"

if ! echo "$commit_msg" | grep -qE "$pattern"; then
    echo "❌ Invalid commit message format."
    echo ""
    echo "Required format: type(scope): subject"
    echo "  Types: refactor, test, docs, style, perf"
    echo "  Scope: package or file name (lowercase)"
    echo "  Subject: descriptive (min 10 chars)"
    echo ""
    echo "Example: refactor(sequences): extract collectTimestamps helper"
    echo ""
    echo "Your message:"
    echo "$commit_msg"
    exit 1
fi

echo "✅ Commit message format valid."
exit 0
```

**Make executable**:
```bash
chmod +x .git/hooks/commit-msg
```

**Checklist**:
- [ ] Commit-msg hook installed: YES / NO
- [ ] Hook enforces convention: VERIFY
- [ ] Can be bypassed if needed: `--no-verify` works

---

## Commit Statistics (Track Over Time)

**Refactoring Session**: ___ (e.g., calculateSequenceTimeSpan - 2025-10-19)

| Metric | Value |
|--------|-------|
| **Total commits** | ___ |
| **Commits with passing tests** | ___ |
| **Average commit size** | ___ lines |
| **Largest commit** | ___ lines |
| **Smallest commit** | ___ lines |
| **Rollbacks needed** | ___ |
| **Fixup commits** | ___ |
| **Commits per hour** | ___ |

**Commit Discipline Score**: (Commits with passing tests) / (Total commits) × 100% = ___%

**Target**: 100% commit discipline (every commit has passing tests)

---

## Example Commit Sequence

**Refactoring**: calculateSequenceTimeSpan (Complexity 10 → <8)

```bash
# Baseline
abc123 test: add edge cases for calculateSequenceTimeSpan
def456 refactor(sequences): extract collectOccurrenceTimestamps helper
ghi789 test: add unit tests for collectOccurrenceTimestamps
jkl012 refactor(sequences): extract findMinMaxTimestamps helper
mno345 test: add unit tests for findMinMaxTimestamps
pqr678 refactor(sequences): simplify calculateSequenceTimeSpan using helpers
stu901 docs(sequences): add GoDoc for calculateSequenceTimeSpan
vwx234 test(sequences): verify complexity reduced to 6
```

**Statistics**:
- Total commits: 8
- Average size: ~30 lines
- Largest commit: def456 (extract helper, 45 lines)
- All commits with passing tests: 8/8 (100%)
- Complexity progression: 10 → 7 (def456) → 6 (pqr678)

---

## Notes

- **Discipline**: Commit after EVERY refactoring step
- **Small**: Keep commits <200 lines
- **Passing**: Every commit must have passing tests
- **Descriptive**: Subject line tells what changed
- **Revertible**: Each commit can be reverted independently
- **Story**: Commit history tells story of refactoring progression

---

**Version**: 1.0 (Iteration 1)
**Next Review**: Iteration 2 (refine based on usage data)
**Automation**: See git hooks section for automated enforcement
