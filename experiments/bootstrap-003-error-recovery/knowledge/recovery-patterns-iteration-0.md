# Recovery Strategy Patterns - Iteration 0

**Version**: 0.1 (Baseline)
**Date**: 2025-10-18
**Patterns**: 5 (covering top error categories)

---

## Pattern 1: Syntax Error Fix-and-Retry

**Applicable to**: Build/Compilation Errors (Category 1)

**Strategy**: Fix syntax error in source code and rebuild

**Steps**:
1. **Locate**: Identify file and line from error message (`file.go:line:col`)
2. **Read**: Read the problematic file section
3. **Fix**: Edit file to correct syntax error (remove unused import, fix type, etc.)
4. **Verify**: Run `go build` or `go test` to verify fix
5. **Retry**: Retry the original failed operation

**Automation Potential**: Semi-automated
- Detection: Fully automated (parse Go error messages)
- Fix suggestion: Automated (linter suggestions)
- Fix application: Manual (requires code understanding)

**Success Rate**: High (>90% for simple syntax errors)

**Time to Recovery**: ~2-5 minutes

**Example**:
```
Error: cmd/root.go:4:2: "fmt" imported and not used

Recovery:
1. Read cmd/root.go
2. Edit cmd/root.go - remove line 4: import "fmt"
3. Bash: go build
4. Verify: Build succeeds
```

**Edge Cases**:
- Error in multiple files: Fix all before rebuilding
- Cascading errors: Fix root cause first, rerun to see remaining errors

---

## Pattern 2: Test Fixture Update

**Applicable to**: Test Failures (Category 2)

**Strategy**: Update test fixtures or test expectations to match current code

**Steps**:
1. **Analyze**: Understand what test expects vs what code produces
2. **Decide**: Determine if code is wrong or test expectation is outdated
3. **Update**: Either fix code or update test fixture/assertion
4. **Verify**: Run test again to confirm pass
5. **Full test**: Run full test suite to check for regressions

**Automation Potential**: Low
- Requires human judgment to determine if code or test is correct
- Cannot automate "is this behavior correct?" decision

**Success Rate**: High (>85% once correct fix identified)

**Time to Recovery**: ~5-15 minutes

**Example**:
```
Error: --- FAIL: TestLoadFixture (0.00s)
    fixtures_test.go:34: Fixture content should contain 'sequence' field

Recovery:
1. Read tests/fixtures/sample-session.jsonl
2. Identify missing 'sequence' field in fixture
3. Edit fixture to add proper 'sequence' field to each entry
4. Bash: go test ./internal/testutil -v
5. Verify: Test passes
```

**Edge Cases**:
- Breaking API change: Multiple tests may need updates
- Fixture format change: May need fixture regeneration script

---

## Pattern 3: Path Correction

**Applicable to**: File Not Found (Category 3)

**Strategy**: Correct file path or create missing file

**Steps**:
1. **Verify**: Confirm file doesn't exist with `ls` or `find`
2. **Locate**: Search for file with correct name in other locations
3. **Decide**: Path typo vs file not created yet
4. **Fix**:
   - If typo: Correct path in command
   - If not created: Create file or reorder workflow
5. **Retry**: Retry original operation with correct path

**Automation Potential**: High
- Path validation can be automated
- Fuzzy path matching can suggest corrections
- "Did you mean?" suggestions

**Success Rate**: Very high (>95% for path typos)

**Time to Recovery**: ~1-3 minutes

**Example**:
```
Error: wc: /home/yale/work/meta-cc/internal/testutil/fixture.go: No such file or directory

Recovery:
1. Bash: ls /home/yale/work/meta-cc/internal/testutil/
2. Find: File is named fixtures.go (not fixture.go)
3. Bash: wc -l /home/yale/work/meta-cc/internal/testutil/fixtures.go
4. Verify: Command succeeds
```

**Edge Cases**:
- File should exist but doesn't: Need to create it first (workflow order issue)
- Permission issue mistaken for file not found: Check permissions

---

## Pattern 4: Read-Then-Write

**Applicable to**: Write Before Read (Category 5)

**Strategy**: Add Read step before Write for existing files, or use Edit instead

**Steps**:
1. **Check existence**: Verify file exists with `ls`
2. **Decide tool**:
   - For modifications: Use Edit tool instead of Write
   - For complete rewrite: Read first, then Write
3. **Read**: Read the existing file content
4. **Write/Edit**: Perform write or edit operation
5. **Verify**: Confirm file has desired content

**Automation Potential**: Fully automated
- Can automatically check file existence before Write
- Can suggest Edit vs Write based on intent
- Can auto-insert Read step in workflow

**Success Rate**: Very high (>98%)

**Time to Recovery**: ~1-2 minutes

**Example**:
```
Error: <tool_use_error>File has not been read yet. Read it first before writing to it.</tool_use_error>

Recovery:
1. Bash: ls internal/testutil/fixtures.go (confirm exists)
2. Read internal/testutil/fixtures.go
3. Edit internal/testutil/fixtures.go (modify specific lines)
   OR
   Write internal/testutil/fixtures.go (complete rewrite after reading)
4. Verify: File updated successfully
```

**Edge Cases**:
- File doesn't exist: Can Write directly without Read (error is misleading)
- Large file: May need pagination for Read step

---

## Pattern 5: Build-Then-Execute

**Applicable to**: Command Not Found (Category 6)

**Strategy**: Build binary before executing command, or add to PATH

**Steps**:
1. **Identify command**: Determine what command is missing
2. **Check buildable**: Is this a project binary (e.g., meta-cc)?
3. **Build**: Run appropriate build command (`make build`)
4. **Execute**: Use local path (`./command`) or install to PATH
5. **Verify**: Command executes successfully

**Automation Potential**: Medium
- Can detect "command not found" automatically
- Can check if Makefile exists and suggest build
- Cannot fully automate (may need dependencies)

**Success Rate**: High (>90% for project binaries)

**Time to Recovery**: ~2-5 minutes (including build time)

**Example**:
```
Error: /bin/bash: line 1: meta-cc: command not found

Recovery:
1. Bash: ls meta-cc (check if binary exists)
2. If not: make build (build the binary)
3. Bash: ./meta-cc --version (use local path)
   OR
   export PATH=$PATH:/home/yale/work/meta-cc
   meta-cc --version
4. Verify: Command succeeds
```

**Edge Cases**:
- Build fails: Need to fix build errors first (cascading recovery)
- System command missing: Install system dependency
- Typo in command name: Suggest correction

---

## Recovery Pattern Summary

| Pattern | Category | Success Rate | MTTR | Automation | Frequency |
|---------|----------|--------------|------|------------|-----------|
| 1. Fix-and-Retry | Build Errors | >90% | 2-5 min | Semi | 15.0% |
| 2. Test Fixture Update | Test Failures | >85% | 5-15 min | Low | 11.2% |
| 3. Path Correction | File Not Found | >95% | 1-3 min | High | 18.7% |
| 4. Read-Then-Write | Workflow Errors | >98% | 1-2 min | Full | 3.0% |
| 5. Build-Then-Execute | Command Errors | >90% | 2-5 min | Medium | 3.7% |

**Coverage**: 5 patterns covering 51.6% of errors (689/1336)

**Overall estimated MTTR**: ~3-7 minutes per error (weighted average)

---

## Baseline MTTR (Mean Time To Recovery)

**Estimated MTTR by category** (manual recovery):
- Build/Compilation: 2-5 minutes
- Test Failures: 5-15 minutes
- File Not Found: 1-3 minutes
- Write Before Read: 1-2 minutes
- Command Not Found: 2-5 minutes

**Average MTTR**: ~4 minutes (simple errors) to ~15 minutes (complex errors)

**Current limitations**:
- No automated recovery
- Manual diagnosis required
- No error context capture
- No recovery verification
- No fallback strategies

---

## Recovery Automation Opportunities

### High Priority (High automation potential, high frequency)

1. **Path Correction** (18.7% of errors)
   - Automated path validation
   - Fuzzy path matching
   - "Did you mean?" suggestions
   - Expected speedup: 5-10x (from 3 min to <30 sec)

2. **Read-Then-Write** (3.0% of errors)
   - Auto-check file existence before Write
   - Auto-suggest Edit vs Write
   - Auto-insert Read step
   - Expected speedup: 10x (from 2 min to <15 sec)

### Medium Priority (Medium automation potential)

3. **Build-Then-Execute** (3.7% of errors)
   - Detect project binaries
   - Auto-suggest `make build`
   - Auto-retry with local path
   - Expected speedup: 3-5x (from 3 min to ~1 min)

4. **Fix-and-Retry** (15.0% of errors)
   - Automated linting suggestions
   - Semi-automated fix application
   - Expected speedup: 2-3x (from 4 min to ~1.5 min)

### Low Priority (Low automation potential, requires human judgment)

5. **Test Fixture Update** (11.2% of errors)
   - Automated test failure parsing
   - Suggested fixes (but manual application)
   - Expected speedup: 1.5-2x (from 10 min to ~5 min)

---

## Gaps and Next Steps

**Missing patterns**:
1. MCP Integration error recovery (17.1% of errors)
2. JSON Parsing error recovery (6.0% of errors)
3. Permission Denied error recovery (0.7% of errors)

**Missing features**:
1. Automated recovery scripts
2. Recovery verification procedures
3. Fallback strategies (if primary recovery fails)
4. Recovery history tracking
5. Success rate measurement

**Iteration 1 goals**:
1. Add 3 more recovery patterns (MCP, JSON, permissions)
2. Implement automated recovery scripts for high-priority patterns
3. Add recovery verification steps
4. Measure actual MTTR with timestamps
5. Implement fallback strategies

---

## Related Prevention Patterns

Each recovery pattern suggests a prevention approach:

| Recovery Pattern | Prevention Approach |
|-----------------|---------------------|
| Fix-and-Retry | Pre-commit linting (catch before commit) |
| Test Fixture Update | CI test validation (catch before merge) |
| Path Correction | Path validation helper (validate before use) |
| Read-Then-Write | Workflow checker (detect pattern before execution) |
| Build-Then-Execute | Build verification (ensure binary exists before commands) |

See `prevention-guidelines-iteration-0.md` for detailed prevention practices.

---

**Version History**:
- v0.1 (2025-10-18): Initial 5 recovery patterns covering 51.6% of errors
