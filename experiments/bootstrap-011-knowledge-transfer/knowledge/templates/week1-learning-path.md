# Week-1 Learning Path: meta-cc Contributor

**Version**: 1.0
**Created**: 2025-10-17
**Target Role**: New contributor completing first week
**Time Estimate**: 20-40 hours (over first 5 work days)
**Prerequisites**: Completed Day-1 path (working environment + first PR submitted)

---

## Overview

Welcome to Week-1! You've completed Day-1 (environment setup + first contribution). Now let's build deep understanding and deliver a meaningful feature.

**Learning Objectives**:
- ✅ Master core architecture (parser → analyzer → query flow)
- ✅ Understand all major modules (parser, analyzer, query, MCP)
- ✅ Learn common development workflows (test-driven development, debugging)
- ✅ Deliver a meaningful contribution (good first issue or small feature)
- ✅ Begin building code ownership in one area

**Success Criteria**:
- You can explain architecture in 2-3 sentences
- You understand parser → analyzer → query → output flow
- You've delivered a meaningful PR (good first issue or small feature)
- You can navigate codebase confidently using grep/LSP
- You've begun specializing in one module area

**Time Investment**: 20-40 hours over first week (4-8 hours per day)

---

## Section 1: Architecture Deep Dive (4-8 hours, Day 2-3)

### Learning Objectives
- Understand high-level architecture (two-layer: CLI + core logic)
- Master data flow: Session JSONL → Parser → Analyzer → Query → Output
- Understand module responsibilities and boundaries
- Know where to find code for different concerns

### Steps

#### 1.1 Read Architecture Documentation (1-2 hours)

**Key Docs to Read**:
```bash
# Architecture overview
cat docs/architecture/proposals/meta-cognition-proposal.md

# ADR (Architecture Decision Records)
ls docs/architecture/adr/
cat docs/architecture/adr/0001-* # Read first few ADRs

# Core principles
cat docs/core/principles.md
```

**Questions to Answer**:
1. What is the two-layer architecture? (CLI layer + core logic layer)
2. What problem does meta-cc solve? (Meta-cognition for Claude Code session history)
3. Why separate CLI from core logic? (Reusability, testability)
4. What are the main constraints? (See docs/core/principles.md)

#### 1.2 Trace Data Flow (2-3 hours)

Follow a query from input to output:

```bash
# 1. Start with CLI entry point
cat cmd/query_tools.go

# Notice pattern:
# - Parse flags
# - Call internal/locator to find session files
# - Call internal/parser to parse JSONL
# - Call internal/query to filter/analyze
# - Format and output results

# 2. Follow into parser
cat internal/parser/parser.go
cat internal/parser/types.go

# Key types:
# - Session: Represents entire session
# - ToolCall: Represents tool invocation
# - Message: Represents user/assistant message

# 3. Follow into analyzer
cat internal/analyzer/analyzer.go

# Analyzer extracts patterns from parsed data

# 4. Follow into query
cat internal/query/engine.go

# Query engine filters and aggregates data
```

**Checkpoint**: Can you trace `meta-cc query-tools --status error` from CLI → Parser → Query → Output?

#### 1.3 Draw Architecture Diagram (1-2 hours)

Create a mental model (or sketch on paper):

```
User runs CLI command
    ↓
cmd/query_tools.go (CLI layer)
    ↓
internal/locator (find session files)
    ↓
internal/parser (parse JSONL → Session struct)
    ↓
internal/analyzer (extract patterns, statistics)
    ↓
internal/query (filter, aggregate)
    ↓
pkg/output (format as JSON/YAML/table)
    ↓
Output to user
```

**Key Boundaries**:
- **cmd/**: CLI interface only (no business logic)
- **internal/**: All core logic (reusable, testable)
- **pkg/**: Shared utilities (output formatters, etc.)

#### 1.4 Explore MCP Server (1-2 hours)

MCP integration is a major feature:

```bash
# MCP server entry point
cat cmd/mcp-server/main.go

# MCP tools (implements MCP protocol)
cat cmd/mcp-server/tools.go
cat cmd/mcp-server/capabilities.go

# Notice:
# - Each tool wraps a core query function
# - MCP layer adds protocol handling (request/response)
# - Core logic reused from internal/ packages
```

**Understanding MCP**:
- MCP = Model Context Protocol (Claude integration)
- Provides tools like `query_tools`, `query_user_messages`, etc.
- Allows Claude to query session history directly
- See: docs/guides/mcp.md for full reference

### Validation Checkpoint ✓

Before proceeding, verify:
- [ ] Can explain two-layer architecture (CLI + core logic)
- [ ] Can trace data flow from CLI → Parser → Query → Output
- [ ] Understand parser/analyzer/query responsibilities
- [ ] Know where MCP server code lives (cmd/mcp-server/)
- [ ] Can draw architecture diagram from memory

**Time Check**: You should be 4-8 hours into Week-1 (Day 2-3). Take a break!

---

## Section 2: Core Module Mastery (6-12 hours, Day 3-4)

### Learning Objectives
- Deep dive into each major module (parser, analyzer, query)
- Understand module APIs and data structures
- Read and understand module tests
- Run module tests and understand coverage
- Identify potential contribution areas in each module

### Steps

#### 2.1 Parser Module Deep Dive (2-4 hours)

**Read Core Files**:
```bash
# Type definitions
cat internal/parser/types.go

# Key types to understand:
# - Session: Top-level container
# - ToolCall: Tool invocation record
# - Message: User/assistant message
# - Error: Error record

# Parser implementation
cat internal/parser/parser.go

# Notice:
# - Reads JSONL line by line
# - Unmarshals JSON to Go structs
# - Handles errors gracefully

# Parser tests
cat internal/parser/parser_test.go

# Understand test patterns:
# - Table-driven tests
# - Test fixtures in testdata/
# - Coverage of error cases
```

**Run Parser Tests**:
```bash
# Run all parser tests
go test -v ./internal/parser/

# Run with coverage
go test -cover ./internal/parser/

# Run specific test
go test -v ./internal/parser/ -run TestParseSession
```

**Checkpoint**: Can you explain what `parser.ParseSession()` does and how it handles JSONL?

#### 2.2 Analyzer Module Deep Dive (2-4 hours)

**Read Core Files**:
```bash
# Analyzer implementation
cat internal/analyzer/analyzer.go
cat internal/analyzer/patterns.go

# Analyzer extracts:
# - Tool usage statistics
# - Error patterns
# - Message patterns
# - Workflow sequences

# Analyzer tests
cat internal/analyzer/analyzer_test.go
```

**Understanding Analysis**:
- **Pattern Detection**: Identifies common tool sequences
- **Statistics**: Aggregates counts, frequencies
- **Error Analysis**: Categorizes and counts errors

**Checkpoint**: Can you list 3 types of patterns the analyzer extracts?

#### 2.3 Query Module Deep Dive (2-4 hours)

**Read Core Files**:
```bash
# Query engine
cat internal/query/engine.go
cat internal/query/filter.go

# Query supports:
# - Filtering (by tool, status, pattern)
# - Aggregation (counts, statistics)
# - Sorting and limiting
# - Output formatting

# Query tests
cat internal/query/engine_test.go
```

**Understanding Queries**:
```bash
# Example queries:
./meta-cc query-tools --tool Bash --status error
# → Filter tools where name=Bash AND status=error

./meta-cc query-user-messages --pattern "fix.*bug"
# → Filter messages matching regex "fix.*bug"

./meta-cc query-tool-sequences --min-occurrences 3
# → Find tool sequences that occurred 3+ times
```

**Checkpoint**: Can you explain how `query-tools --status error` filters tool calls?

### Validation Checkpoint ✓

Before proceeding, verify:
- [ ] Understand parser types (Session, ToolCall, Message)
- [ ] Know what analyzer extracts (patterns, statistics, errors)
- [ ] Understand query filtering and aggregation
- [ ] Can run tests for each module
- [ ] Identified one module you find most interesting (future specialization area)

**Time Check**: You should be 10-20 hours into Week-1 (Day 3-4). Halfway there!

---

## Section 3: Development Workflows (4-8 hours, Day 4-5)

### Learning Objectives
- Master test-driven development (TDD) workflow
- Learn debugging techniques (print debugging, delve)
- Understand CI/CD pipeline (linting, testing, building)
- Practice git workflow (branching, committing, PR)
- Learn code review process

### Steps

#### 3.1 Test-Driven Development (TDD) (2-3 hours)

**TDD Workflow**:
```bash
# 1. Write test first (red)
vim internal/parser/parser_test.go

# Add failing test:
func TestParseNewFeature(t *testing.T) {
    // Test for feature that doesn't exist yet
    result := parser.NewFeature(input)
    if result != expected {
        t.Errorf("expected %v, got %v", expected, result)
    }
}

# 2. Run test (should fail)
go test ./internal/parser/ -run TestParseNewFeature
# → FAIL (as expected)

# 3. Implement feature (green)
vim internal/parser/parser.go
# Add NewFeature() implementation

# 4. Run test again (should pass)
go test ./internal/parser/ -run TestParseNewFeature
# → PASS

# 5. Refactor if needed
# Clean up code while keeping tests green
```

**Practice TDD**:
- Choose a small feature from "good first issue" list
- Write test first
- Implement feature
- Verify test passes

#### 3.2 Debugging Techniques (1-2 hours)

**Print Debugging**:
```go
// Add debug prints
fmt.Printf("DEBUG: value = %+v\n", myStruct)
log.Printf("DEBUG: entering function with args: %v", args)

// Run test with debug output
go test -v ./internal/parser/ -run TestSpecificCase
```

**Delve Debugger** (optional, advanced):
```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug a test
dlv test ./internal/parser/ -- -test.run TestSpecificCase

# Delve commands:
# break <location>  - Set breakpoint
# continue          - Continue execution
# next              - Step over
# step              - Step into
# print <var>       - Print variable
# list              - Show source code
```

**Checkpoint**: Can you add debug prints to trace execution flow through a function?

#### 3.3 CI/CD Pipeline Understanding (1-2 hours)

**Local CI Simulation**:
```bash
# Run full CI pipeline locally
make all

# This runs:
# 1. make lint    - golangci-lint (code quality)
# 2. make test    - go test (all tests)
# 3. make build   - go build (compile binaries)

# Fix lint errors
make lint
# → Fix reported issues

# Ensure tests pass
make test
# → All tests must pass

# Verify build succeeds
make build
# → Binaries in bin/
```

**Understanding CI**:
- CI runs automatically on every PR
- Must pass: lint + test + build
- Check `.github/workflows/` for CI configuration

#### 3.4 Git Workflow Practice (1-2 hours)

**Feature Branch Workflow**:
```bash
# 1. Create feature branch
git checkout -b feat/my-feature

# 2. Make changes
vim internal/parser/parser.go
vim internal/parser/parser_test.go

# 3. Run tests locally
make all

# 4. Commit with proper message
git add internal/parser/
git commit -m "feat(parser): add NewFeature support

Adds NewFeature() to parse new JSONL field.

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>"

# 5. Push to fork
git push origin feat/my-feature

# 6. Create PR on GitHub
# Include:
# - Clear description
# - Test plan
# - Link to related issue
```

### Validation Checkpoint ✓

Before proceeding, verify:
- [ ] Understand TDD workflow (test → implement → verify)
- [ ] Can add debug prints to trace execution
- [ ] Can run `make all` and fix issues
- [ ] Understand git feature branch workflow
- [ ] Know commit message format (conventional commits)

**Time Check**: You should be 14-28 hours into Week-1 (Day 4-5). Almost done!

---

## Section 4: Meaningful Contribution (6-12 hours, Day 5)

### Learning Objectives
- Find and claim a "good first issue"
- Implement feature with tests
- Submit PR with proper description
- Respond to code review feedback
- Merge your first meaningful contribution

### Steps

#### 4.1 Find Good First Issue (1 hour)

**Where to Look**:
```bash
# GitHub issues with label
# → Browse: https://github.com/yaleh/meta-cc/labels/good%20first%20issue

# Or search existing issues for ideas:
# - Add test coverage for uncovered code
# - Improve error messages
# - Add documentation
# - Small feature additions
```

**Good First Issue Criteria**:
- Small scope (can be completed in 4-8 hours)
- Clear requirements
- Has tests (or can be easily tested)
- Improves codebase (not just cosmetic)

**Claim the Issue**:
- Comment: "I'd like to work on this"
- Wait for maintainer confirmation
- Ask questions if anything unclear

#### 4.2 Implement Feature (3-6 hours)

**TDD Approach**:
```bash
# 1. Write tests first
vim internal/{module}/{feature}_test.go

# 2. Run tests (should fail)
go test ./internal/{module}/

# 3. Implement feature
vim internal/{module}/{feature}.go

# 4. Run tests (should pass)
go test ./internal/{module}/

# 5. Add integration test (if applicable)
vim cmd/{command}_test.go

# 6. Run full test suite
make test

# 7. Verify lint passes
make lint

# 8. Build and manually test
make build
./bin/meta-cc {your-command} {test-args}
```

**Code Quality Checklist**:
- [ ] Tests written and passing
- [ ] Lint passes (no warnings)
- [ ] Code documented (comments on exported functions)
- [ ] Error handling (graceful failures)
- [ ] Edge cases handled

#### 4.3 Create Pull Request (1-2 hours)

**PR Description Template**:
```markdown
## Summary
Brief description of what this PR does.

## Motivation
Why is this change needed? (Link to issue if applicable)

## Changes
- Added {feature} to {module}
- Updated {tests} to cover {cases}
- Documented {functions} with clear comments

## Test Plan
- [ ] Unit tests added and passing (`go test ./internal/{module}/`)
- [ ] Integration test added (if applicable)
- [ ] Manual testing performed: `./meta-cc {command} {args}`
- [ ] Lint passes (`make lint`)
- [ ] Full CI passes (`make all`)

## Type of Change
- [ ] Bug fix (non-breaking change fixing an issue)
- [ ] New feature (non-breaking change adding functionality)
- [ ] Breaking change (fix or feature causing existing functionality to not work as expected)
- [ ] Documentation update

## Screenshots (if applicable)
[Add screenshots of before/after if visual change]

## Related Issues
Closes #123 (if this PR closes an issue)
```

**Submit PR**:
```bash
# Push your branch
git push origin feat/my-feature

# Create PR via GitHub UI or gh CLI
gh pr create --title "feat(parser): add NewFeature support" \
             --body "$(cat pr-description.md)"
```

#### 4.4 Respond to Code Review (1-3 hours)

**Code Review Etiquette**:
- Be receptive to feedback (learning opportunity!)
- Ask questions if feedback unclear
- Make requested changes promptly
- Test changes after each revision
- Thank reviewers for their time

**Typical Review Cycle**:
```bash
# 1. Reviewer requests changes
# → Read feedback carefully

# 2. Make changes locally
vim internal/{module}/{feature}.go

# 3. Run tests again
make all

# 4. Commit changes
git add internal/{module}/
git commit -m "refactor: address code review feedback

- Improved error handling as suggested
- Added missing test case
- Fixed edge case bug"

# 5. Push update
git push origin feat/my-feature

# 6. PR updates automatically
# → Wait for re-review

# 7. Iterate until approved
```

**Checkpoint**: Can you respond professionally to code review feedback?

### Validation Checkpoint ✓

Before celebrating, verify:
- [ ] Found and claimed a good first issue
- [ ] Implemented feature with tests
- [ ] PR submitted with clear description
- [ ] Responded to code review feedback
- [ ] PR approved and merged (or pending final review)

**Time Check**: You should be 20-40 hours into Week-1 (Day 5). Week complete!

---

## Week-1 Complete! 🎉

You now have:
- ✅ Deep understanding of meta-cc architecture
- ✅ Mastery of core modules (parser, analyzer, query)
- ✅ Proficiency in development workflows (TDD, debugging, git)
- ✅ Meaningful contribution delivered (good first issue or feature)
- ✅ Foundation for Month-1 learning (code ownership, mentoring)

### What You've Accomplished

**Technical Skills**:
- Architecture understanding (two-layer: CLI + core logic)
- Module expertise (parser → analyzer → query → output flow)
- Development workflows (TDD, debugging, CI/CD, git)
- Code review participation

**Knowledge Gained**:
- Data flow from JSONL session files to query results
- Module boundaries and responsibilities
- Test-driven development practice
- Git feature branch workflow
- PR creation and code review process

**Contributions**:
- Week-1 contribution: Meaningful feature or good first issue
- Building code ownership in one module area
- Beginning to mentor others (answer questions, review PRs)

### Next Steps

**Immediate**:
- Continue responding to code review on Week-1 PR
- Celebrate meaningful contribution! 🎉
- Begin exploring Month-1 path preview

**Month-1 Path Preview**:
- Advanced architecture topics (concurrency, performance, scalability)
- Complex feature delivery (multi-module, cross-cutting)
- Code ownership (become module expert)
- Mentoring capability (help other contributors)
- Community participation (discussions, issue triage)

**Specialization Areas** (choose one to focus Month-1):
- **Parser Expert**: Optimize parsing, handle edge cases, improve error handling
- **Analyzer Expert**: Build new pattern detectors, improve statistics
- **Query Expert**: Add query features, optimize performance
- **MCP Expert**: Enhance MCP integration, add new tools
- **CLI Expert**: Improve user experience, add commands, better output

### Resources for Continued Learning

**Deep Dives**:
- `docs/architecture/proposals/meta-cognition-proposal.md`: Full architecture
- `docs/architecture/adr/`: Architecture decision records
- `docs/guides/plugin-development.md`: Plugin system deep dive
- `docs/guides/mcp.md`: MCP integration complete reference

**Advanced Topics**:
- Concurrency patterns in Go
- Performance optimization techniques
- Advanced testing (benchmarks, fuzz testing)
- Documentation best practices

### Getting Help

**Stuck on something?**
- Review architecture docs (docs/architecture/)
- Search GitHub Issues for similar problems
- Ask in PR comments or discussions
- Review recent PRs for examples
- Reach out to module maintainers

**Questions to ask yourself**:
- Can I explain architecture to a new contributor?
- Do I understand parser → analyzer → query flow?
- Have I delivered a meaningful contribution?
- Am I ready to own a module area?

If yes to all four, you've successfully completed Week-1! 🚀

---

## Appendix: Common Week-1 Issues

### Issue: Tests failing locally but not in CI

**Solution**:
```bash
# Ensure same Go version as CI
go version  # Should match .github/workflows/

# Clean test cache
go clean -testcache

# Re-run tests
go test ./...

# Check for race conditions
go test -race ./...
```

### Issue: Lint errors not caught locally

**Solution**:
```bash
# Install same golangci-lint version as CI
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run full lint
golangci-lint run

# Auto-fix some issues
golangci-lint run --fix
```

### Issue: Understanding complex code flow

**Solution**:
```bash
# Add debug prints at key points
fmt.Printf("DEBUG: entering %s with %+v\n", funcName, args)

# Use delve to step through
dlv test ./internal/{module}/ -- -test.run TestSpecificCase

# Read tests to understand expected behavior
cat internal/{module}/*_test.go
```

### Issue: PR review taking too long

**Solution**:
- Be patient - maintainers are volunteers
- Ensure PR description is clear
- Respond promptly to initial feedback
- Keep PR scope small (easier to review)
- Ping politely after 3-5 days if no response

---

**Week-1 Learning Path Version**: 1.0
**Created**: 2025-10-17
**Maintained by**: meta-cc contributors
**Feedback**: Please open an issue if you find errors or have suggestions!
