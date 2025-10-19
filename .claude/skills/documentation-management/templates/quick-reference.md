# Quick Reference Template

**Purpose**: Template for creating concise, scannable reference documentation (cheat sheets, command references, API quick guides)

**Version**: 1.0
**Status**: Ready for use
**Validation**: Applied to BAIME quick reference outline

---

## When to Use This Template

### Use For

✅ **Command-line tool references** (CLI commands, options, examples)
✅ **API quick guides** (endpoints, parameters, responses)
✅ **Configuration cheat sheets** (settings, values, defaults)
✅ **Keyboard shortcut guides** (shortcuts, actions, contexts)
✅ **Syntax references** (language syntax, operators, constructs)
✅ **Workflow checklists** (steps, validation, common patterns)

### Don't Use For

❌ **Comprehensive tutorials** (use tutorial-structure.md instead)
❌ **Conceptual explanations** (use concept-explanation.md instead)
❌ **Detailed troubleshooting** (use troubleshooting guide template)
❌ **Narrative documentation** (use example-walkthrough.md)

---

## Template Structure

### 1. Title and Scope

**Purpose**: Immediately communicate what this reference covers

**Structure**:
```markdown
# [Tool/API/Feature] Quick Reference

**Purpose**: [One sentence describing what this reference covers]
**Scope**: [What's included and what's not]
**Last Updated**: [Date]
```

**Example**:
```markdown
# BAIME Quick Reference

**Purpose**: Essential commands, patterns, and workflows for BAIME methodology development
**Scope**: Covers common operations, subagent invocations, value functions. See full tutorial for conceptual explanations.
**Last Updated**: 2025-10-19
```

---

### 2. At-A-Glance Summary

**Purpose**: Provide 10-second overview for users who already know basics

**Structure**:
```markdown
## At a Glance

**Core Workflow**:
1. [Step 1] - [What it does]
2. [Step 2] - [What it does]
3. [Step 3] - [What it does]

**Most Common Commands**:
- `[command]` - [Description]
- `[command]` - [Description]

**Key Concepts**:
- **[Concept]**: [One-sentence definition]
- **[Concept]**: [One-sentence definition]
```

**Example**:
```markdown
## At a Glance

**Core BAIME Workflow**:
1. Design iteration prompts - Define experiment structure
2. Execute Iteration 0 - Establish baseline
3. Iterate until convergence - Improve both layers

**Most Common Subagents**:
- `iteration-prompt-designer` - Create ITERATION-PROMPTS.md
- `iteration-executor` - Run OCA cycle iteration
- `knowledge-extractor` - Extract final methodology

**Key Metrics**:
- **V_instance ≥ 0.80**: Domain work quality
- **V_meta ≥ 0.80**: Methodology quality
```

---

### 3. Command Reference (for CLI/API tools)

**Purpose**: Provide exhaustive, scannable command list

**Structure**:

#### For CLI Tools

```markdown
## Command Reference

### [Command Category]

#### `[command] [options] [args]`

**Description**: [What this command does]

**Options**:
- `-a, --option-a` - [Description]
- `-b, --option-b VALUE` - [Description] (default: VALUE)

**Examples**:
```bash
# [Use case 1]
[command] [example]

# [Use case 2]
[command] [example]
```

**Common Patterns**:
- [Pattern description]: `[command pattern]`
```

#### For APIs

```markdown
## API Reference

### [Endpoint Category]

#### `[METHOD] /path/to/endpoint`

**Description**: [What this endpoint does]

**Parameters**:
| Name | Type | Required | Description |
|------|------|----------|-------------|
| param1 | string | Yes | [Description] |
| param2 | number | No | [Description] (default: value) |

**Request Example**:
```json
{
  "param1": "value",
  "param2": 42
}
```

**Response Example**:
```json
{
  "status": "success",
  "data": { ... }
}
```

**Error Codes**:
- `400` - [Error description]
- `404` - [Error description]
```

---

### 4. Pattern Reference

**Purpose**: Document common patterns and their usage

**Structure**:
```markdown
## Common Patterns

### Pattern: [Pattern Name]

**When to use**: [Situation where this pattern applies]

**Structure**:
```
[Pattern template or pseudocode]
```

**Example**:
```[language]
[Concrete example]
```

**Variations**:
- [Variation 1]: [When to use]
- [Variation 2]: [When to use]
```

**Example**:
```markdown
## Common Patterns

### Pattern: Value Function Calculation

**When to use**: End of each iteration, during evaluation phase

**Structure**:
```
V_component = (Metric1 + Metric2 + ... + MetricN) / N
V_layer = (Component1 + Component2 + ... + ComponentN) / N
```

**Example**:
```
V_instance = (Accuracy + Completeness + Usability + Maintainability) / 4
V_instance = (0.75 + 0.60 + 0.65 + 0.80) / 4 = 0.70
```

**Variations**:
- **Weighted average**: When components have different importance
- **Minimum threshold**: When any component below threshold fails entire layer
```

---

### 5. Decision Trees / Flowcharts (Text-Based)

**Purpose**: Help users navigate choices

**Structure**:
```markdown
## Decision Guide: [What Decision]

**Question**: [Decision question]

→ **If [condition]**:
  - Do: [Action]
  - Why: [Rationale]
  - Example: [Example]

→ **Else if [condition]**:
  - Do: [Action]
  - Why: [Rationale]

→ **Otherwise**:
  - Do: [Action]
```

**Example**:
```markdown
## Decision Guide: When to Create Specialized Agent

**Question**: Should I create a specialized agent for this task?

→ **If ALL of these are true**:
  - Task performed 3+ times with similar structure
  - Generic approach struggled or was inefficient
  - Can articulate specific agent improvements

  - **Do**: Create specialized agent
  - **Why**: Evidence shows insufficiency, pattern clear
  - **Example**: test-generator after manual test writing 3x

→ **Else if task done 1-2 times only**:
  - **Do**: Wait for more evidence
  - **Why**: Insufficient pattern recurrence

→ **Otherwise (no clear benefit)**:
  - **Do**: Continue with generic approach
  - **Why**: Evolution requires evidence, not speculation
```

---

### 6. Troubleshooting Quick Reference

**Purpose**: One-line solutions to common issues

**Structure**:
```markdown
## Quick Troubleshooting

| Problem | Quick Fix | Full Details |
|---------|-----------|--------------|
| [Symptom] | [Quick solution] | [Link to detailed guide] |
| [Symptom] | [Quick solution] | [Link to detailed guide] |
```

**Example**:
```markdown
## Quick Troubleshooting

| Problem | Quick Fix | Full Details |
|---------|-----------|--------------|
| Value scores not improving | Check if solving symptoms vs root causes | [Full troubleshooting](#troubleshooting) |
| Low V_meta Reusability | Parameterize patterns, add adaptation guides | [Full troubleshooting](#troubleshooting) |
| Iterations taking too long | Use specialized subagents, time-box templates | [Full troubleshooting](#troubleshooting) |
| Can't reach 0.80 threshold | Re-evaluate value function definitions | [Full troubleshooting](#troubleshooting) |
```

---

### 7. Configuration/Settings Reference

**Purpose**: Document all configurable options

**Structure**:
```markdown
## Configuration Reference

### [Configuration Category]

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| `setting_name` | type | default | [What it does] |
| `setting_name` | type | default | [What it does] |

**Example Configuration**:
```[format]
[example config file]
```
```

**Example**:
```markdown
## Value Function Configuration

### Instance Layer Components

| Component | Weight | Range | Description |
|-----------|--------|-------|-------------|
| Accuracy | 0.25 | 0.0-1.0 | Technical correctness, factual accuracy |
| Completeness | 0.25 | 0.0-1.0 | Coverage of user needs, edge cases |
| Usability | 0.25 | 0.0-1.0 | Clarity, accessibility, examples |
| Maintainability | 0.25 | 0.0-1.0 | Modularity, consistency, automation |

**Example Calculation**:
```
V_instance = (0.75 + 0.60 + 0.65 + 0.80) / 4 = 0.70
```
```

---

### 8. Related Resources

**Purpose**: Point to related documentation

**Structure**:
```markdown
## Related Resources

**Deeper Learning**:
- [Tutorial Name](link) - [When to read]
- [Guide Name](link) - [When to read]

**Related References**:
- [Reference Name](link) - [What it covers]

**External Resources**:
- [Resource Name](link) - [Description]
```

---

## Quality Checklist

Before publishing, verify:

### Content Quality

- [ ] **Scannability**: Can user find information in <30 seconds?
- [ ] **Completeness**: All common commands/operations covered?
- [ ] **Examples**: Every command/pattern has concrete example?
- [ ] **Accuracy**: All commands/code tested and working?
- [ ] **Currency**: Information up-to-date with latest version?

### Structure Quality

- [ ] **At-a-glance section**: Provides 10-second overview?
- [ ] **Consistent formatting**: Tables, code blocks, headings uniform?
- [ ] **Cross-references**: Links to detailed docs where needed?
- [ ] **Navigation**: Easy to jump to specific section?

### User Experience

- [ ] **Target audience**: Assumes user knows basics, needs quick lookup?
- [ ] **No redundancy**: Information not duplicated from full docs?
- [ ] **Print-friendly**: Could be printed as 1-2 page reference?
- [ ] **Progressive disclosure**: Most common info first, advanced later?

### Maintainability

- [ ] **Version tracking**: Last updated date present?
- [ ] **Change tracking**: Version history documented?
- [ ] **Linked to source**: References to source of truth (API spec, etc)?
- [ ] **Update frequency**: Plan for keeping current?

---

## Adaptation Guide

### For Different Domains

**CLI Tools** (git, docker, etc):
- Focus on command syntax, options, examples
- Include common workflows (init → add → commit → push)
- Add troubleshooting for common errors

**APIs** (REST, GraphQL):
- Focus on endpoints, parameters, responses
- Include authentication examples
- Add rate limits, error codes

**Configuration** (yaml, json, env):
- Focus on settings, defaults, validation
- Include complete example config
- Add common configuration patterns

**Syntax** (programming languages):
- Focus on operators, keywords, constructs
- Include code examples for each construct
- Add "coming from X language" sections

### Length Guidelines

**Ideal length**: 1-3 printed pages (500-1500 words)
- Too short (<500 words): Probably missing common use cases
- Too long (>2000 words): Should be split or moved to full tutorial

**Balance**: 70% reference tables/lists, 30% explanatory text

---

## Examples of Good Quick References

### Example 1: Git Cheat Sheet

**Why it works**:
- Commands organized by workflow (init, stage, commit, branch)
- Each command has one-line description
- Common patterns shown (fork → clone → branch → PR)
- Fits on one page

### Example 2: Docker Quick Reference

**Why it works**:
- Separates basic commands from advanced
- Shows command anatomy (docker [options] command [args])
- Includes real-world examples
- Links to full documentation

### Example 3: Python String Methods Reference

**Why it works**:
- Alphabetical table of methods
- Each method shows signature and one example
- Indicates Python version compatibility
- Quick search via browser Ctrl+F

---

## Common Mistakes to Avoid

### ❌ Mistake 1: Too Much Explanation

**Problem**: Quick reference becomes mini-tutorial

**Bad**:
```markdown
## git commit

Git commit is an important command that saves your changes to the local repository.
Before committing, you should stage your changes with git add. Commits create a
snapshot of your work that you can return to later...
[3 more paragraphs]
```

**Good**:
```markdown
## git commit

`git commit -m "message"` - Save staged changes with message

Examples:
- `git commit -m "Add login feature"` - Basic commit
- `git commit -a -m "Fix bug"` - Stage and commit all
- `git commit --amend` - Modify last commit

See: [Full Git Guide](link) for commit best practices
```

### ❌ Mistake 2: Missing Examples

**Problem**: Syntax shown but no concrete usage

**Bad**:
```markdown
## API Endpoint

`POST /api/users`

Parameters: name (string), email (string), age (number)
```

**Good**:
```markdown
## API Endpoint

`POST /api/users` - Create new user

Example Request:
```bash
curl -X POST https://api.example.com/api/users \
  -H "Content-Type: application/json" \
  -d '{"name": "Alice", "email": "alice@example.com", "age": 30}'
```

Example Response:
```json
{"id": 123, "name": "Alice", "email": "alice@example.com"}
```
```

### ❌ Mistake 3: Poor Organization

**Problem**: Commands in random order, no grouping

**Bad**:
- `docker ps`
- `docker build`
- `docker stop`
- `docker run`
- `docker images`
[Random order, hard to find]

**Good**:
**Image Commands**:
- `docker build` - Build image
- `docker images` - List images

**Container Commands**:
- `docker run` - Start container
- `docker ps` - List containers
- `docker stop` - Stop container

### ❌ Mistake 4: No Progressive Disclosure

**Problem**: Advanced features mixed with basics

**Bad**:
```markdown
## Commands
- ls - List files
- docker buildx create --use --platform=linux/arm64,linux/amd64
- cd directory - Change directory
- git rebase -i --autosquash --fork-point main
```

**Good**:
```markdown
## Basic Commands
- `ls` - List files
- `cd directory` - Change directory

## Advanced Commands
- `docker buildx create --use --platform=...` - Multi-platform builds
- `git rebase -i --autosquash` - Interactive rebase
```

---

## Template Variables

When creating quick reference, customize:

- `[Tool/API/Feature]` - Name of what's being referenced
- `[Command Category]` - Logical grouping of commands
- `[Method]` - HTTP method or operation type
- `[Parameter]` - Input parameter name
- `[Example]` - Concrete, runnable example

---

## Validation Checklist

Test your quick reference:

1. **Speed test**: Can experienced user find command in <30 seconds?
2. **Completeness test**: Are 80%+ of common operations covered?
3. **Example test**: Can user copy/paste examples and run successfully?
4. **Print test**: Is it useful when printed?
5. **Search test**: Can user Ctrl+F to find what they need?

**If any test fails, revise before publishing.**

---

## Version History

- **1.0** (2025-10-19): Initial template created from documentation methodology iteration 2

---

**Ready to use**: Apply this template to create scannable, efficient quick reference guides for any tool, API, or feature.
