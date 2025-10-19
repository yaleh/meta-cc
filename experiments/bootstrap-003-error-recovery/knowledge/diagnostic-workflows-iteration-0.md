# Root Cause Diagnostic Workflows - Iteration 0

**Version**: 0.1 (Baseline)
**Date**: 2025-10-18
**Workflows**: 5 (covering top error categories)

---

## Workflow 1: Build/Compilation Error Diagnosis

**Applicable to**: Category 1 (Build/Compilation Errors)
**Frequency**: 15.0% of errors

### Step 1: Identify Error Symptoms

**Look for**:
- Bash tool output containing `# github.com/`
- Go compiler error messages
- Build failure indicators

**Error Signature Pattern**:
```
# github.com/yale/meta-cc/[package]
[file].go:[line]:[col]: [error message]
```

### Step 2: Gather Context

**Information to collect**:
1. Which file has the error? (`[file].go:[line]:[col]`)
2. What is the error message? (syntax, type, import issue)
3. What was the last code change?
4. Which Go package is affected?

**Commands**:
```bash
# View the problematic file
Read [file].go

# Check recent changes
git diff [file].go

# Check full build output
go build -v
```

### Step 3: Analyze Root Cause

**Common root causes**:

| Error Pattern | Root Cause | Fix |
|---------------|------------|-----|
| `"fmt" imported and not used` | Unused import | Remove import or use it |
| `undefined: [name]` | Variable/function not defined | Define it or import package |
| `cannot use [X] (type [A]) as type [B]` | Type mismatch | Fix type or add conversion |
| `use of internal package not allowed` | Import violation | Use public API or restructure |
| `go.mod already exists` | Module re-init | Skip `go mod init` |

### Step 4: Verify Diagnosis

**Verification**:
1. Read the error line in source file
2. Confirm error matches one of the known patterns
3. Identify exact fix needed

**Success Criteria**:
- Exact line and column identified
- Error type classified
- Fix approach clear

**Estimated Time**: ~2-5 minutes

**Tools Needed**: Read tool, Bash (go build), git diff

---

## Workflow 2: Test Failure Diagnosis

**Applicable to**: Category 2 (Test Failures)
**Frequency**: 11.2% of errors

### Step 1: Identify Error Symptoms

**Look for**:
- `--- FAIL: TestName`
- `FAIL\tpackage\ttime`
- Test assertion messages

**Error Signature Pattern**:
```
--- FAIL: TestName (time)
    file_test.go:line: [assertion message]
FAIL
FAIL	package	time
```

### Step 2: Gather Context

**Information to collect**:
1. Which test failed? (`TestName`)
2. What was the assertion? (assertion message)
3. What was expected vs actual?
4. Which package is being tested?

**Commands**:
```bash
# Run specific test for details
go test -v ./[package] -run TestName

# View test file
Read [package]/[file]_test.go

# Check if fixtures exist
ls tests/fixtures/
```

### Step 3: Analyze Root Cause

**Common root causes**:

| Error Pattern | Root Cause | Fix |
|---------------|------------|-----|
| `Fixture content should contain 'X'` | Missing fixture data | Update fixture file |
| `Expected [X], got [Y]` | Test expectation outdated | Update test or fix code |
| `Failed to load fixture` | Fixture file missing | Create fixture file |
| `[build failed]` in test | Syntax error in test code | Fix test code syntax |

### Step 4: Verify Diagnosis

**Verification**:
1. Run the specific test in isolation
2. Confirm expected vs actual values
3. Identify if issue is in test or code

**Success Criteria**:
- Test failure reproduced
- Root cause identified (code bug or test bug)
- Fix approach clear

**Estimated Time**: ~3-10 minutes

**Tools Needed**: Bash (go test), Read tool, file operations

---

## Workflow 3: File Not Found Diagnosis

**Applicable to**: Category 3 (File Not Found)
**Frequency**: 18.7% of errors

### Step 1: Identify Error Symptoms

**Look for**:
- `<tool_use_error>File does not exist.</tool_use_error>`
- `no such file or directory`
- Read or Bash tool errors

**Error Signature Pattern**:
```
<tool_use_error>File does not exist.</tool_use_error>
```
or
```
[command]: /path/to/file: No such file or directory
```

### Step 2: Gather Context

**Information to collect**:
1. What file path was attempted? (`/path/to/file`)
2. Which tool reported the error? (Read, Bash, etc.)
3. What operation was attempted? (read, execute, etc.)
4. What is the current working directory?

**Commands**:
```bash
# Check if file exists
ls -la /path/to/file

# Check parent directory
ls -la /path/to/

# Check current directory
pwd
ls -la
```

### Step 3: Analyze Root Cause

**Common root causes**:

| Symptom | Root Cause | Fix |
|---------|------------|-----|
| File path has typo | Typo in path | Correct path spelling |
| File not created yet | Workflow order issue | Create file first |
| File in different directory | Wrong path | Find correct path with `find` |
| Relative vs absolute path | Path confusion | Use absolute path |

### Step 4: Verify Diagnosis

**Verification**:
1. Use `ls` or `find` to locate file
2. Confirm correct path
3. If file doesn't exist, determine if it should be created

**Success Criteria**:
- File location confirmed or non-existence verified
- Correct path identified
- Creation vs path correction decided

**Estimated Time**: ~1-3 minutes

**Tools Needed**: Bash (ls, find, pwd), file operations

---

## Workflow 4: Write Before Read Diagnosis

**Applicable to**: Category 5 (Write Before Read)
**Frequency**: 3.0% of errors

### Step 1: Identify Error Symptoms

**Look for**:
- Write tool error
- `<tool_use_error>File has not been read yet. Read it first before writing to it.</tool_use_error>`

**Error Signature Pattern**:
```
<tool_use_error>File has not been read yet. Read it first before writing to it.</tool_use_error>
```

### Step 2: Gather Context

**Information to collect**:
1. What file was being written? (`file_path` in Write tool input)
2. Does the file already exist?
3. Was this an attempt to overwrite or create new file?

**Commands**:
```bash
# Check if file exists
ls -la [file_path]
```

### Step 3: Analyze Root Cause

**Root cause**: Claude Code safety constraint requires reading existing files before overwriting

**Decision tree**:
- If file exists AND needs modification → Use Edit tool instead
- If file exists AND needs complete rewrite → Read first, then Write
- If file doesn't exist → Use Write directly (no Read needed)

### Step 4: Verify Diagnosis

**Verification**:
1. Confirm file exists with `ls`
2. Determine if Edit or Write is appropriate
3. If Write is needed, add Read step first

**Success Criteria**:
- File existence confirmed
- Correct tool (Edit vs Write) identified
- Workflow corrected

**Estimated Time**: ~1-2 minutes

**Tools Needed**: Bash (ls), Read tool, Edit or Write tool

---

## Workflow 5: Command Not Found Diagnosis

**Applicable to**: Category 6 (Command Not Found)
**Frequency**: 3.7% of errors

### Step 1: Identify Error Symptoms

**Look for**:
- Bash tool error
- `command not found: [command]`

**Error Signature Pattern**:
```
/bin/bash: line 1: [command]: command not found
```

### Step 2: Gather Context

**Information to collect**:
1. What command was attempted? (`[command]`)
2. Is it a custom binary or system command?
3. Should it exist or needs to be built/installed?

**Commands**:
```bash
# Check if command in PATH
which [command]

# Check if binary exists locally
ls -la ./[command]

# For meta-cc specifically
ls -la /home/yale/work/meta-cc/meta-cc
```

### Step 3: Analyze Root Cause

**Common root causes**:

| Command Type | Root Cause | Fix |
|--------------|------------|-----|
| `meta-cc` | Not built yet | Run `make build` |
| `meta-cc` | Not in PATH | Use `./meta-cc` or install to PATH |
| System command | Typo | Fix command name |
| System command | Not installed | Install dependency |

### Step 4: Verify Diagnosis

**Verification**:
1. Use `which` to check PATH
2. Use `ls` to check local directory
3. Determine if build, install, or typo fix needed

**Success Criteria**:
- Command location verified or non-existence confirmed
- Build vs install vs typo identified
- Fix approach clear

**Estimated Time**: ~1-2 minutes

**Tools Needed**: Bash (which, ls, make build)

---

## Workflow Summary Table

| Workflow | Category | Frequency | MTTD | Complexity | Automation Potential |
|----------|----------|-----------|------|------------|---------------------|
| 1. Build/Compilation | Build Errors | 15.0% | 2-5 min | Medium | Medium (linting) |
| 2. Test Failure | Test Failures | 11.2% | 3-10 min | Medium-High | Low (needs analysis) |
| 3. File Not Found | File Errors | 18.7% | 1-3 min | Low | High (path validation) |
| 4. Write Before Read | Workflow Errors | 3.0% | 1-2 min | Low | High (automated check) |
| 5. Command Not Found | Command Errors | 3.7% | 1-2 min | Low | Medium (build check) |

**Coverage**: 5 workflows covering 51.6% of errors (689/1336)

---

## Baseline MTTD (Mean Time To Diagnosis)

**Overall MTTD estimate**: ~3-5 minutes per error

**By category**:
- Build/Compilation: 2-5 minutes
- Test Failures: 3-10 minutes
- File Not Found: 1-3 minutes
- Write Before Read: 1-2 minutes
- Command Not Found: 1-2 minutes

**Current approach**: Manual diagnosis, no automation

---

## Gaps and Next Steps

**Current gaps**:
1. No workflows for MCP Integration errors (17.1%)
2. No workflows for JSON Parsing errors (6.0%)
3. No automated diagnosis tools
4. No error context capture (surrounding tool calls)
5. No recovery procedures (diagnosis only)

**Iteration 1 goals**:
1. Add diagnostic workflows for MCP and JSON parsing errors
2. Create automated error classification script
3. Add recovery procedures to each workflow
4. Implement error context capture (window of N tool calls)
5. Measure actual MTTD with timestamps

---

**Version History**:
- v0.1 (2025-10-18): Initial 5 diagnostic workflows covering 51.6% of errors
