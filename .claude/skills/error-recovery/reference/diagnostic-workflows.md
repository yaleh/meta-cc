# Diagnostic Workflows

**Version**: 2.0
**Source**: Bootstrap-003 Error Recovery Methodology
**Last Updated**: 2025-10-18
**Coverage**: 78.7% of errors (8 workflows)

Step-by-step diagnostic procedures for common error categories.

---

## Workflow 1: Build/Compilation Errors (15.0%)

**MTTD**: 2-5 minutes

### Symptoms
- `go build` fails
- Error messages: `*.go:[line]:[col]: [error]`

### Diagnostic Steps

**Step 1: Identify Error Location**
```bash
go build 2>&1 | tee build-error.log
grep "\.go:" build-error.log
```

**Step 2: Classify Error Type**
- Syntax error (braces, semicolons)
- Type error (mismatches)
- Import error (unused/missing)
- Definition error (undefined references)

**Step 3: Inspect Context**
```bash
sed -n '[line-5],[line+5]p' [file]
```

### Tools
- `go build`, `grep`, `sed`
- IDE/editor

### Success Criteria
- Root cause identified
- Fix approach clear

### Automation
Medium (linters, IDE integration)

---

## Workflow 2: Test Failures (11.2%)

**MTTD**: 3-10 minutes

### Symptoms
- `go test` fails
- `FAIL` messages in output

### Diagnostic Steps

**Step 1: Identify Failing Test**
```bash
go test ./... -v 2>&1 | tee test-output.log
grep "FAIL:" test-output.log
```

**Step 2: Isolate Test**
```bash
go test ./internal/parser -run TestParseSession
```

**Step 3: Analyze Failure**
- Assertion failure (expected vs actual)
- Panic (runtime error)
- Timeout
- Setup failure

**Step 4: Inspect Code/Data**
```bash
cat [test_file].go | grep -A 20 "func Test[Name]"
cat tests/fixtures/[fixture]
```

### Tools
- `go test`, `grep`
- Test fixtures

### Success Criteria
- Understand why assertion failed
- Know expected vs actual behavior

### Automation
Low (requires understanding intent)

---

## Workflow 3: File Not Found (18.7%)

**MTTD**: 1-3 minutes

### Symptoms
- `File does not exist`
- `No such file or directory`

### Diagnostic Steps

**Step 1: Verify Non-Existence**
```bash
ls [path]
find . -name "[filename]"
```

**Step 2: Search for Similar Files**
```bash
find . -iname "*[partial_name]*"
ls [directory]/
```

**Step 3: Classify Issue**
- Path typo (wrong name/location)
- File not created yet
- Wrong working directory
- Case sensitivity issue

**Step 4: Fuzzy Match**
```bash
# Use automation tool
./scripts/validate-path.sh [attempted_path]
```

### Tools
- `ls`, `find`
- `validate-path.sh` (automation)

### Success Criteria
- Know exact cause (typo vs missing)
- Found correct path or know file needs creation

### Automation
**High** (path validation, fuzzy matching)

---

## Workflow 4: File Size Exceeded (6.3%)

**MTTD**: 1-2 minutes

### Symptoms
- `File content exceeds maximum allowed tokens`
- Read operation fails with size error

### Diagnostic Steps

**Step 1: Check File Size**
```bash
wc -l [file]
du -h [file]
```

**Step 2: Determine Strategy**
- Use offset/limit parameters
- Use grep/head/tail
- Process in chunks

**Step 3: Execute Alternative**
```bash
# Option A: Pagination
Read [file] offset=0 limit=1000

# Option B: Selective reading
grep "pattern" [file]
head -n 1000 [file]
```

### Tools
- `wc`, `du`
- Read tool with pagination
- `grep`, `head`, `tail`
- `check-file-size.sh` (automation)

### Success Criteria
- Got needed information without full read

### Automation
**Full** (size check, auto-pagination)

---

## Workflow 5: Write Before Read (5.2%)

**MTTD**: 1-2 minutes

### Symptoms
- `File has not been read yet`
- Write/Edit tool error

### Diagnostic Steps

**Step 1: Verify File Exists**
```bash
ls [file]
```

**Step 2: Determine Operation Type**
- Modification → Use Edit tool
- Complete rewrite → Read then Write
- New file → Write directly (no Read needed)

**Step 3: Add Read Step**
```bash
Read [file]
Edit [file] old_string="..." new_string="..."
```

### Tools
- Read, Edit, Write tools
- `check-read-before-write.sh` (automation)

### Success Criteria
- File read before modification
- Correct tool chosen (Edit vs Write)

### Automation
**Full** (auto-insert Read step)

---

## Workflow 6: Command Not Found (3.7%)

**MTTD**: 2-5 minutes

### Symptoms
- `command not found`
- Bash execution fails

### Diagnostic Steps

**Step 1: Identify Command Type**
```bash
which [command]
type [command]
```

**Step 2: Check if Project Binary**
```bash
ls ./[command]
ls bin/[command]
```

**Step 3: Build if Needed**
```bash
# Check build system
ls Makefile
cat Makefile | grep [command]

# Build
make build
```

**Step 4: Execute with Path**
```bash
./[command] [args]
# OR
PATH=$PATH:./bin [command] [args]
```

### Tools
- `which`, `type`
- `make`
- Project build system

### Success Criteria
- Command found or built
- Can execute successfully

### Automation
Medium (can detect and suggest build)

---

## Workflow 7: JSON Parsing Errors (6.0%)

**MTTD**: 3-8 minutes

### Diagnostic Steps

**Step 1: Validate JSON Syntax**
```bash
jq . [file.json]
cat [file.json] | python -m json.tool
```

**Step 2: Locate Parsing Error**
```bash
# Error message shows line/field
# View context around error
sed -n '[line-5],[line+5]p' [file.json]
```

**Step 3: Classify Issue**
- Syntax error (commas, braces)
- Type mismatch (string vs int)
- Missing field
- Schema change

**Step 4: Fix or Update**
- Fix JSON structure
- Update Go struct definition
- Update test fixtures

### Tools
- `jq`, `python -m json.tool`
- Go compiler (for schema errors)

### Success Criteria
- JSON is valid
- Schema matches code expectations

### Automation
Medium (syntax validation yes, schema fix no)

---

## Workflow 8: String Not Found (Edit) (3.2%)

**MTTD**: 1-3 minutes

### Symptoms
- `String to replace not found in file`
- Edit operation fails

### Diagnostic Steps

**Step 1: Re-Read File**
```bash
Read [file]
```

**Step 2: Locate Target Section**
```bash
grep -n "target_pattern" [file]
```

**Step 3: Copy Exact String**
- View file content
- Copy exact string (including whitespace)
- Don't retype (preserves formatting)

**Step 4: Retry Edit**
```bash
Edit [file] old_string="[exact_copied_string]" new_string="[new]"
```

### Tools
- Read tool
- `grep -n`

### Success Criteria
- Found exact current string
- Edit succeeds

### Automation
High (auto-refresh before edit)

---

## Diagnostic Workflow Selection

### Decision Tree

```
Error occurs
├─ Build fails? → Workflow 1
├─ Test fails? → Workflow 2
├─ File not found? → Workflow 3 ⚠️ AUTOMATE
├─ File too large? → Workflow 4 ⚠️ AUTOMATE
├─ Write before read? → Workflow 5 ⚠️ AUTOMATE
├─ Command not found? → Workflow 6
├─ JSON parsing? → Workflow 7
├─ Edit string not found? → Workflow 8
└─ Other? → See taxonomy.md
```

---

## Best Practices

### General Diagnostic Approach

1. **Reproduce**: Ensure error is reproducible
2. **Classify**: Match to error category
3. **Follow workflow**: Use appropriate diagnostic workflow
4. **Document**: Note findings for future reference
5. **Verify**: Confirm diagnosis before fix

### Time Management

- Set time limit per diagnostic step (5-10 min)
- If stuck, escalate or try different approach
- Use automation tools when available

### Common Mistakes

❌ Skip verification steps
❌ Assume root cause without evidence
❌ Try fixes without diagnosis
✅ Follow workflow systematically
✅ Use tools/automation
✅ Document findings

---

**Source**: Bootstrap-003 Error Recovery Methodology
**Framework**: BAIME (Bootstrapped AI Methodology Engineering)
**Status**: Production-ready, validated with 1336 errors
