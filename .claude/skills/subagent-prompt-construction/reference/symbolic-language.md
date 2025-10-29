# Symbolic Language Reference

Formal syntax for compact, expressive subagent prompts using lambda calculus and predicate logic.

---

## Lambda Calculus

### Function Definition
```
λ(parameters) → output | constraints
```

**Example**:
```
λ(feature_spec, todo_ref?) → (plan, execution_report, status) | TDD ∧ code_limits
```

### Type Signatures
```
function_name :: InputType → OutputType
```

**Examples**:
```
parse_feature :: FeatureSpec → Requirements
execute_stage :: (Plan, StageNumber) → StageResult
quality_check :: StageResult → QualityReport
```

### Function Application
```
function_name(arguments) = definition
```

**Example**:
```
parse_feature(spec) =
  extract(objectives, scope, constraints) ∧
  identify(deliverables) ∧
  assess(complexity)
```

---

## Logic Operators

### Conjunction (AND)
**Symbol**: `∧`

**Usage**:
```
condition1 ∧ condition2 ∧ condition3
```

**Example**:
```
validate(input) ∧ process(input) ∧ output(result)
```

**Semantics**: All conditions must be true

### Disjunction (OR)
**Symbol**: `∨`

**Usage**:
```
condition1 ∨ condition2 ∨ condition3
```

**Example**:
```
is_complete(task) ∨ is_blocked(task) ∨ is_cancelled(task)
```

**Semantics**: At least one condition must be true

### Negation (NOT)
**Symbol**: `¬`

**Usage**:
```
¬condition
```

**Example**:
```
¬empty(results) ∧ ¬error(status)
```

**Semantics**: Condition must be false

### Implication (THEN)
**Symbol**: `→`

**Usage**:
```
step1 → step2 → step3 → result
```

**Example**:
```
parse(input) → validate → process → output
```

**Semantics**: Sequential execution or logical implication

### Bidirectional Implication
**Symbol**: `↔`

**Usage**:
```
condition1 ↔ condition2
```

**Example**:
```
valid_input(x) ↔ passes_schema(x) ∧ passes_constraints(x)
```

**Semantics**: Conditions are equivalent (both true or both false)

---

## Quantifiers

### Universal Quantifier (For All)
**Symbol**: `∀`

**Usage**:
```
∀element ∈ collection: predicate(element)
```

**Examples**:
```
∀stage ∈ plan.stages: execute(stage)
∀result ∈ results: result.status == "complete"
∀stage ∈ stages: |code(stage)| ≤ 200
```

**Semantics**: Predicate must be true for all elements

### Existential Quantifier (Exists)
**Symbol**: `∃`

**Usage**:
```
∃element ∈ collection: predicate(element)
```

**Examples**:
```
∃error ∈ results: error.severity == "critical"
∃stage ∈ stages: stage.status == "failed"
```

**Semantics**: Predicate must be true for at least one element

### Unique Existence
**Symbol**: `∃!`

**Usage**:
```
∃!element ∈ collection: predicate(element)
```

**Example**:
```
∃!config ∈ configs: config.name == "production"
```

**Semantics**: Exactly one element satisfies predicate

---

## Set Operations

### Element Of
**Symbol**: `∈`

**Usage**:
```
element ∈ collection
```

**Example**:
```
"complete" ∈ valid_statuses
stage ∈ plan.stages
```

### Subset
**Symbol**: `⊆`

**Usage**:
```
set1 ⊆ set2
```

**Example**:
```
completed_stages ⊆ all_stages
required_tools ⊆ available_tools
```

### Superset
**Symbol**: `⊇`

**Usage**:
```
set1 ⊇ set2
```

**Example**:
```
all_features ⊇ implemented_features
```

### Union
**Symbol**: `∪`

**Usage**:
```
set1 ∪ set2
```

**Example**:
```
errors ∪ warnings → all_issues
agents_required ∪ mcp_tools_required → dependencies
```

### Intersection
**Symbol**: `∩`

**Usage**:
```
set1 ∩ set2
```

**Example**:
```
completed_stages ∩ tested_stages → verified_stages
```

---

## Comparison Operators

### Equality
**Symbols**: `=`, `==`

**Usage**:
```
variable = value
expression == expected
```

**Examples**:
```
status = "complete"
result.count == expected_count
```

### Inequality
**Symbols**: `≠`, `!=`

**Usage**:
```
value ≠ unwanted
```

**Example**:
```
status ≠ "error"
```

### Less Than
**Symbol**: `<`

**Usage**:
```
value < threshold
```

**Example**:
```
error_count < 5
```

### Less Than or Equal
**Symbol**: `≤`

**Usage**:
```
value ≤ maximum
```

**Examples**:
```
|code| ≤ 200
coverage ≥ 0.80 ∧ lines ≤ 150
```

### Greater Than
**Symbol**: `>`

**Usage**:
```
value > minimum
```

**Example**:
```
coverage > 0.75
```

### Greater Than or Equal
**Symbol**: `≥`

**Usage**:
```
value ≥ threshold
```

**Examples**:
```
coverage(stage) ≥ 0.80
integration_score ≥ 0.50
```

---

## Special Symbols

### Cardinality/Length
**Symbol**: `|x|`

**Usage**:
```
|collection|
|string|
```

**Examples**:
```
|code(stage)| ≤ 200
|results| > 0
|plan.stages| == expected_count
```

### Delta (Change)
**Symbol**: `Δx`

**Usage**:
```
Δvariable
```

**Examples**:
```
ΔV_meta = V_meta(s_1) - V_meta(s_0)
Δcoverage = coverage_after - coverage_before
```

### Prime (Next State)
**Symbol**: `x'`

**Usage**:
```
variable'
```

**Examples**:
```
state' = update(state)
results' = results + [new_result]
```

### Subscript (Iteration)
**Symbol**: `x_n`

**Usage**:
```
variable_n
```

**Examples**:
```
V_meta_1 = evaluate(methodology_1)
iteration_n
stage_i
```

---

## Composite Patterns

### Conditional Logic
```
if condition then
  action1
else if condition2 then
  action2
else
  action3
```

**Compact form**:
```
condition ? action1 : action2
```

### Pattern Matching
```
match value:
  case pattern1 → action1
  case pattern2 → action2
  case _ → default_action
```

### List Comprehension
```
[expression | element ∈ collection, predicate(element)]
```

**Example**:
```
completed = [s | s ∈ stages, s.status == "complete"]
error_count = |[r | r ∈ results, r.status == "error"]|
```

---

## Compactness Examples

### Verbose vs. Compact

**Verbose** (50 lines):
```
First, we need to validate the input to ensure it meets all requirements.
If the input is valid, we should proceed to extract the objectives.
After extracting objectives, we need to identify the scope.
Then we should assess the complexity of the task.
Finally, we return the parsed requirements.
```

**Compact** (5 lines):
```
parse :: Input → Requirements
parse(input) =
  validate(input) ∧
  extract(objectives, scope) ∧
  assess(complexity) → requirements
```

### Constraints

**Verbose**:
```
For each stage in the execution plan:
  - The code should not exceed 200 lines
  - The tests should not exceed 200 lines
  - The test coverage should be at least 80%
For the entire phase:
  - The total code should not exceed 500 lines
  - TDD compliance must be maintained
  - All tests must pass
```

**Compact**:
```
constraints :: PhaseExecution → Bool
constraints(exec) =
  ∀stage ∈ exec.plan.stages:
    |code(stage)| ≤ 200 ∧
    |test(stage)| ≤ 200 ∧
    coverage(stage) ≥ 0.80 ∧
  |code(exec.phase)| ≤ 500 ∧
  tdd_compliance(exec) ∧
  all_tests_pass(exec)
```

---

## Style Guidelines

### Function Names
- Use snake_case: `parse_feature`, `execute_stage`
- Descriptive verbs: `extract`, `validate`, `generate`
- Domain-specific terminology: `quality_check`, `error_analysis`

### Type Names
- Use PascalCase: `FeatureSpec`, `StageResult`, `PhaseReport`
- Singular nouns: `Plan`, `Result`, `Report`
- Composite types: `(Plan, StageNumber)`, `[StageResult]`

### Variable Names
- Use snake_case: `feature_spec`, `stage_num`, `recent_errors`
- Abbreviations acceptable: `req`, `exec`, `ctx`
- Descriptive in context: `plan`, `result`, `report`

### Spacing
- Spaces around operators: `x ∧ y`, `a ≤ b`
- No spaces in function calls: `function(arg1, arg2)`
- Indent nested blocks consistently

### Line Length
- Target: ≤80 characters
- Break long expressions at logical operators
- Align continuations with opening delimiter

---

## Common Idioms

### Sequential Steps
```
step1 → step2 → step3 → result
```

### Conditional Execution
```
if condition then action else alternative
```

### Iteration with Predicate
```
∀element ∈ collection: predicate(element)
```

### Filtering
```
filtered = [x | x ∈ collection, predicate(x)]
```

### Aggregation
```
total = sum([metric(x) | x ∈ collection])
```

### Validation
```
valid(x) = constraint1(x) ∧ constraint2(x) ∧ constraint3(x)
```

---

## Related Resources

- **Patterns**: `patterns.md` (how to use symbolic language in patterns)
- **Integration Patterns**: `integration-patterns.md` (agent/MCP/skill syntax)
- **Template**: `../templates/subagent-template.md` (symbolic language in practice)
- **Example**: `../examples/phase-planner-executor.md` (real-world usage)
