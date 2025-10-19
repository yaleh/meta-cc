# Recovery Strategy Patterns

**Version**: 1.0
**Source**: Bootstrap-003 Error Recovery Methodology
**Last Updated**: 2025-10-18

This document provides proven recovery patterns for each error category.

---

## Pattern 1: Syntax Error Fix-and-Retry

**Applicable to**: Build/Compilation Errors (Category 1)

**Strategy**: Fix syntax error in source code and rebuild

**Steps**:
1. **Locate**: Identify file and line from error (`file.go:line:col`)
2. **Read**: Read the problematic file section
3. **Fix**: Edit file to correct syntax error
4. **Verify**: Run `go build` or `go test`
5. **Retry**: Retry original operation

**Automation**: Semi-automated (detection automatic, fix manual)

**Success Rate**: >90%

**Time to Recovery**: 2-5 minutes

**Example**:
```
Error: cmd/root.go:4:2: "fmt" imported and not used

Recovery:
1. Read cmd/root.go
2. Edit cmd/root.go - remove line 4: import "fmt"
3. Bash: go build
4. Verify: Build succeeds
```

---

## Pattern 2: Test Fixture Update

**Applicable to**: Test Failures (Category 2)

**Strategy**: Update test fixtures or expectations to match current code

**Steps**:
1. **Analyze**: Understand test expectation vs code output
2. **Decide**: Determine if code or test is incorrect
3. **Update**: Fix code or update test fixture/assertion
4. **Verify**: Run test again
5. **Full test**: Run complete test suite

**Automation**: Low (requires human judgment)

**Success Rate**: >85%

**Time to Recovery**: 5-15 minutes

**Example**:
```
Error: --- FAIL: TestLoadFixture (0.00s)
    fixtures_test.go:34: Missing 'sequence' field

Recovery:
1. Read tests/fixtures/sample-session.jsonl
2. Identify missing 'sequence' field
3. Edit fixture to add 'sequence' field
4. Bash: go test ./internal/testutil -v
5. Verify: Test passes
```

---

## Pattern 3: Path Correction ⚠️ AUTOMATABLE

**Applicable to**: File Not Found (Category 3)

**Strategy**: Correct file path or create missing file

**Steps**:
1. **Verify**: Confirm file doesn't exist (`ls` or `find`)
2. **Locate**: Search for file with correct name
3. **Decide**: Path typo vs file not created
4. **Fix**:
   - If typo: Correct path
   - If not created: Create file or reorder workflow
5. **Retry**: Retry with correct path

**Automation**: High (path validation, fuzzy matching, "did you mean?")

**Success Rate**: >95%

**Time to Recovery**: 1-3 minutes

**Example**:
```
Error: No such file: /path/internal/testutil/fixture.go

Recovery:
1. Bash: ls /path/internal/testutil/
2. Find: File is fixtures.go (not fixture.go)
3. Bash: wc -l /path/internal/testutil/fixtures.go
4. Verify: Success
```

---

## Pattern 4: Read-Then-Write ⚠️ AUTOMATABLE

**Applicable to**: Write Before Read (Category 5)

**Strategy**: Add Read step before Write, or use Edit

**Steps**:
1. **Check existence**: Verify file exists
2. **Decide tool**:
   - For modifications: Use Edit
   - For complete rewrite: Read then Write
3. **Read**: Read existing file content
4. **Write/Edit**: Perform operation
5. **Verify**: Confirm desired content

**Automation**: Fully automated (can auto-insert Read step)

**Success Rate**: >98%

**Time to Recovery**: 1-2 minutes

**Example**:
```
Error: File has not been read yet.

Recovery:
1. Bash: ls internal/testutil/fixtures.go
2. Read internal/testutil/fixtures.go
3. Edit internal/testutil/fixtures.go
4. Verify: Updated successfully
```

---

## Pattern 5: Build-Then-Execute

**Applicable to**: Command Not Found (Category 6)

**Strategy**: Build binary before executing, or add to PATH

**Steps**:
1. **Identify**: Determine missing command
2. **Check buildable**: Is this a project binary?
3. **Build**: Run build command (`make build`)
4. **Execute**: Use local path or install to PATH
5. **Verify**: Command executes

**Automation**: Medium (can detect and suggest build)

**Success Rate**: >90%

**Time to Recovery**: 2-5 minutes

**Example**:
```
Error: meta-cc: command not found

Recovery:
1. Bash: ls meta-cc (check if exists)
2. If not: make build
3. Bash: ./meta-cc --version
4. Verify: Command runs
```

---

## Pattern 6: Pagination for Large Files ⚠️ AUTOMATABLE

**Applicable to**: File Size Exceeded (Category 4)

**Strategy**: Use offset/limit or alternative tools

**Steps**:
1. **Detect**: File size check before read
2. **Choose approach**:
   - **Option A**: Read with offset/limit
   - **Option B**: Use grep/head/tail
   - **Option C**: Process in chunks
3. **Execute**: Apply chosen approach
4. **Verify**: Obtained needed information

**Automation**: Fully automated (can auto-detect and paginate)

**Success Rate**: 100%

**Time to Recovery**: 1-2 minutes

**Example**:
```
Error: File exceeds 25000 tokens

Recovery:
1. Bash: wc -l large-file.jsonl  # Check size
2. Read large-file.jsonl offset=0 limit=1000  # Read first 1000 lines
3. OR: Bash: head -n 1000 large-file.jsonl
4. Verify: Got needed content
```

---

## Pattern 7: JSON Schema Fix

**Applicable to**: JSON Parsing Errors (Category 7)

**Strategy**: Fix JSON structure or update schema

**Steps**:
1. **Validate**: Use `jq` to check JSON validity
2. **Locate**: Find exact parsing error location
3. **Analyze**: Determine if JSON or code schema is wrong
4. **Fix**:
   - If JSON: Fix structure (commas, braces, types)
   - If schema: Update Go struct tags/types
5. **Test**: Verify parsing succeeds

**Automation**: Medium (syntax validation yes, schema fix no)

**Success Rate**: >85%

**Time to Recovery**: 3-8 minutes

**Example**:
```
Error: json: cannot unmarshal string into field .count of type int

Recovery:
1. Read testdata/fixture.json
2. Find: "count": "42" (string instead of int)
3. Edit: Change to "count": 42
4. Bash: go test ./internal/parser
5. Verify: Test passes
```

---

## Pattern 8: String Exact Match

**Applicable to**: String Not Found (Edit Errors) (Category 13)

**Strategy**: Re-read file and copy exact string

**Steps**:
1. **Re-read**: Read file to get current content
2. **Locate**: Find target section (grep or visual)
3. **Copy exact**: Copy current string exactly (no retyping)
4. **Retry Edit**: Use exact old_string
5. **Verify**: Edit succeeds

**Automation**: High (auto-refresh content before edit)

**Success Rate**: >95%

**Time to Recovery**: 1-3 minutes

**Example**:
```
Error: String to replace not found in file

Recovery:
1. Read internal/parser/parse.go  # Fresh read
2. Grep: Search for target function
3. Copy exact string from current file
4. Edit with exact old_string
5. Verify: Edit succeeds
```

---

## Pattern 9: MCP Server Health Check

**Applicable to**: MCP Server Errors (Category 9)

**Strategy**: Check server health, restart if needed

**Steps**:
1. **Check status**: Verify MCP server is running
2. **Test connection**: Simple query to test connectivity
3. **Restart**: If down, restart MCP server
4. **Optimize query**: If timeout, add pagination/filters
5. **Retry**: Retry original query

**Automation**: Medium (health checks yes, query optimization no)

**Success Rate**: >80%

**Time to Recovery**: 2-10 minutes

**Example**:
```
Error: MCP server connection failed

Recovery:
1. Bash: ps aux | grep mcp-server
2. If not running: Restart MCP server
3. Test: Simple query (e.g., get_session_stats)
4. If working: Retry original query
5. Verify: Query succeeds
```

---

## Pattern 10: Permission Fix

**Applicable to**: Permission Denied (Category 10)

**Strategy**: Change permissions or use appropriate user

**Steps**:
1. **Check current**: `ls -la` to see permissions
2. **Identify owner**: `ls -l` shows file owner
3. **Fix permission**:
   - Option A: `chmod` to add permissions
   - Option B: `chown` to change owner
   - Option C: Use sudo (if appropriate)
4. **Retry**: Retry original operation
5. **Verify**: Operation succeeds

**Automation**: Low (security implications)

**Success Rate**: >90%

**Time to Recovery**: 1-3 minutes

**Example**:
```
Error: Permission denied: /path/to/file

Recovery:
1. Bash: ls -la /path/to/file
2. See: -r--r--r-- (read-only)
3. Bash: chmod u+w /path/to/file
4. Retry: Write operation
5. Verify: Success
```

---

## Recovery Pattern Selection

### Decision Tree

```
Error occurs
├─ Build/compilation? → Pattern 1 (Fix-and-Retry)
├─ Test failure? → Pattern 2 (Test Fixture Update)
├─ File not found? → Pattern 3 (Path Correction) ⚠️ AUTOMATE
├─ File too large? → Pattern 6 (Pagination) ⚠️ AUTOMATE
├─ Write before read? → Pattern 4 (Read-Then-Write) ⚠️ AUTOMATE
├─ Command not found? → Pattern 5 (Build-Then-Execute)
├─ JSON parsing? → Pattern 7 (JSON Schema Fix)
├─ String not found (Edit)? → Pattern 8 (String Exact Match)
├─ MCP server? → Pattern 9 (MCP Health Check)
├─ Permission denied? → Pattern 10 (Permission Fix)
└─ Other? → Consult taxonomy for category
```

---

## Automation Priority

**High Priority** (Full automation possible):
1. Pattern 3: Path Correction (validate-path.sh)
2. Pattern 4: Read-Then-Write (check-read-before-write.sh)
3. Pattern 6: Pagination (check-file-size.sh)

**Medium Priority** (Partial automation):
4. Pattern 5: Build-Then-Execute
5. Pattern 7: JSON Schema Fix
6. Pattern 9: MCP Server Health

**Low Priority** (Manual required):
7. Pattern 1: Syntax Error Fix
8. Pattern 2: Test Fixture Update
9. Pattern 10: Permission Fix

---

## Best Practices

### General Recovery Workflow

1. **Classify**: Match error to category (use taxonomy.md)
2. **Select pattern**: Choose appropriate recovery pattern
3. **Execute steps**: Follow pattern steps systematically
4. **Verify**: Confirm recovery successful
5. **Document**: Note if pattern needs refinement

### Efficiency Tips

- Keep taxonomy.md open for quick classification
- Use automation tools when available
- Don't skip verification steps
- Track recurring errors for prevention

### Common Mistakes

❌ **Don't**: Retry without understanding error
❌ **Don't**: Skip verification step
❌ **Don't**: Ignore automation opportunities
✅ **Do**: Classify error first
✅ **Do**: Follow pattern steps systematically
✅ **Do**: Verify recovery completely

---

**Source**: Bootstrap-003 Error Recovery Methodology
**Framework**: BAIME (Bootstrapped AI Methodology Engineering)
**Status**: Production-ready, validated with 1336 errors
