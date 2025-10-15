# Evolve Capability

## Purpose
Determine when to create specialized testing agents, identify new testing capability needs, and guide Meta-Agent evolution based on testing domain requirements.

## Specialization Triggers

### When to Create Specialized Testing Agent

**Trigger 1: Repeated Testing Pattern**
- Same type of test generation needed multiple times
- Example: Generating table-driven tests for 10+ functions
- Solution: Create `agents/test-generator.md` with table-driven test expertise

**Trigger 2: Domain-Specific Testing Knowledge Required**
- Generic agents lack testing methodology knowledge
- Example: Coverage analysis requiring Go coverage format parsing
- Solution: Create `agents/coverage-analyzer.md` with coverage expertise

**Trigger 3: Complex Testing Task Requiring Specialized Skills**
- Task requires specific testing tools or techniques
- Example: Test optimization requiring profiling and parallel execution
- Solution: Create `agents/test-optimizer.md` with optimization expertise

**Trigger 4: Interface/Dependency Mocking Complexity**
- Need systematic mock generation for multiple interfaces
- Example: Mocking external dependencies across packages
- Solution: Create `agents/mock-designer.md` with mocking expertise

**Trigger 5: Test Quality Assessment**
- Need systematic test quality evaluation
- Example: Reviewing test clarity, maintainability across codebase
- Solution: Create `agents/test-reviewer.md` with quality assessment expertise

### When NOT to Create Specialized Agent

**Anti-Pattern 1: One-Time Task**
- Task only needed once in experiment
- Use generic coder or data-analyst instead

**Anti-Pattern 2: Simple Task**
- Task can be accomplished with basic commands
- No specialized knowledge required

**Anti-Pattern 3: Premature Specialization**
- Haven't tried with generic agents yet
- Unclear if specialization provides value

## Testing Agent Design Template

When creating specialized testing agent:

```markdown
# [Agent Name] Agent

## Identity
You are a [agent name] agent specialized in [specific testing capability].

## Expertise
- [Testing knowledge area 1]
- [Testing knowledge area 2]
- [Testing knowledge area 3]
- [Testing tool proficiency]
- [Testing pattern mastery]

## Responsibilities
- [Primary testing responsibility]
- [Secondary testing responsibility]
- [Testing quality assurance]

## Testing Methodology

### [Key Testing Skill 1]
[Detailed description of testing approach]

### [Key Testing Skill 2]
[Detailed description of testing technique]

### [Key Testing Skill 3]
[Detailed description of testing strategy]

## Tools and Techniques
- **Go Testing**: testing package, testify, subtests, table-driven tests
- **Coverage Analysis**: go test -cover, go tool cover
- **Test Patterns**: [Specific patterns this agent masters]

## Output Format
[What this agent produces: test code, analysis, reports]

## Quality Standards
- [Testing quality criterion 1]
- [Testing quality criterion 2]
- [Testing quality criterion 3]
```

## Common Specialized Testing Agents

### agents/test-generator.md
**When to Create**: Need to generate tests for multiple functions/packages
**Expertise**:
- Table-driven test generation
- Subtest structure creation
- Edge case identification (nil, empty, boundary)
- Error case test generation
- Testify assertion usage

### agents/coverage-analyzer.md
**When to Create**: Need systematic coverage gap analysis
**Expertise**:
- Go coverage report parsing
- Gap prioritization (critical paths first)
- Branch coverage analysis
- Coverage improvement recommendations

### agents/test-optimizer.md
**When to Create**: Need to improve test execution performance
**Expertise**:
- Test profiling and timing analysis
- Parallel execution design (t.Parallel())
- Fixture optimization
- Test dependency reduction

### agents/mock-designer.md
**When to Create**: Need to mock external dependencies systematically
**Expertise**:
- Interface mock generation
- Test double design
- Fixture data creation
- Test isolation patterns

### agents/test-reviewer.md
**When to Create**: Need systematic test quality assessment
**Expertise**:
- Test clarity evaluation
- Maintainability assessment
- Test convention validation
- Refactoring recommendations

## Meta-Agent Capability Evolution

### When to Add Meta-Agent Capability

**Trigger 1: New Testing Analysis Dimension**
- Need to assess testing aspect not in existing capabilities
- Example: Mutation testing, property-based testing coverage
- Solution: Create `meta-agents/mutation-analyze.md`

**Trigger 2: New Testing Coordination Pattern**
- Need new way to coordinate testing agents
- Example: Continuous test generation based on coverage monitoring
- Solution: Extend `meta-agents/execute.md` or create new capability

**Trigger 3: New Testing Reflection Dimension**
- Need to evaluate testing state in new way
- Example: Test-to-code ratio analysis, test documentation coverage
- Solution: Extend `meta-agents/reflect.md`

### Capability Evolution Process

1. **Identify Need**: What testing capability is missing?
2. **Design Specification**: Define capability scope and interface
3. **Create Capability File**: `meta-agents/{capability}.md`
4. **Document Trigger**: Why was this capability needed?
5. **Update Execution Flow**: How does this capability integrate?

## Evolution Documentation

### For Each New Agent

Document in iteration file:
```markdown
## Agent Evolution: {agent-name}

**Trigger**: [What testing need prompted creation]
**Justification**: [Why existing agents insufficient]
**Capabilities**: [What testing expertise this agent provides]
**Expected Usage**: [When and how this agent will be used]
```

### For Each New Capability

Document in iteration file:
```markdown
## Capability Evolution: {capability-name}

**Trigger**: [What testing analysis gap identified]
**Justification**: [Why existing capabilities insufficient]
**Specification**: [What this capability enables]
**Integration**: [How it fits with existing capabilities]
```

## Evolution Validation

### After Creating Specialized Agent

1. **Effectiveness Check**: Does agent perform better than generic agent?
2. **Reusability Check**: Will agent be used in multiple iterations?
3. **Quality Check**: Does agent produce higher quality tests?
4. **Efficiency Check**: Does agent save time or improve outcomes?

### After Adding Capability

1. **Necessity Check**: Was capability truly needed?
2. **Completeness Check**: Does capability cover identified gap?
3. **Integration Check**: Does capability work well with others?
4. **Value Check**: Does capability improve V(s) calculation/progress?

## Key Principles

1. **Let Testing Needs Drive Evolution**: Don't create agents/capabilities speculatively
2. **Validate Before Specialization**: Try with generic agents first
3. **Document Justification**: Always explain why evolution was necessary
4. **Assess Effectiveness**: Validate that specialization improved outcomes
5. **Avoid Over-Engineering**: Simplest solution that meets testing needs
