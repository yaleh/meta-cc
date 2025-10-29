# Symbolic Language Reference

## Overview

Formal symbolic syntax for compact, expressive subagent prompts using lambda-calculus and predicate logic.

**Benefits**:
- 30-50% more compact than prose
- Precise semantics
- Easy to parse and understand
- Reduces ambiguity

---

## Logic Operators

| Symbol | Name | Usage | Example |
|--------|------|-------|---------|
| `∧` | AND | Logical conjunction | `valid(x) ∧ complete(y)` |
| `∨` | OR | Logical disjunction | `error(x) ∨ timeout(y)` |
| `¬` | NOT | Logical negation | `¬empty(list)` |
| `→` | Implies/Then | Sequencing or implication | `validate(x) → process(x)` |
| `↔` | Iff | Bidirectional implication | `converged ↔ V_meta ≥ 0.75` |

### Examples

**Conjunction**:
```
constraints(code) =
  |code| ≤ 500 ∧
  coverage(code) ≥ 0.80 ∧
  all_tests_pass(code)
```

**Sequencing**:
```
process :: Input → Output
process(input) =
  validate(input) →
  transform(input) →
  output(result)
```

**Negation**:
```
should_continue :: State → Bool
should_continue(s) =
  ¬converged(s) ∧ ¬exceeded_iterations(s)
```

---

## Quantifiers

| Symbol | Name | Usage | Example |
|--------|------|-------|---------|
| `∀x` | For all | Universal quantification | `∀stage ∈ stages: valid(stage)` |
| `∃x` | Exists | Existential quantification | `∃error ∈ errors: critical(error)` |
| `∃!x` | Exists unique | Unique existence | `∃!solution: optimal(solution)` |

### Examples

**Universal**:
```
all_stages_valid :: Plan → Bool
all_stages_valid(plan) =
  ∀stage ∈ plan.stages:
    |code(stage)| ≤ 200 ∧
    |test(stage)| ≤ 200 ∧
    coverage(stage) ≥ 0.80
```

**Existential**:
```
has_errors :: Results → Bool
has_errors(results) =
  ∃result ∈ results: result.status == "error"
```

**Conditional with Quantifier**:
```
execute_until_error :: [Stage] → [Result]
execute_until_error(stages) =
  results = [] →
  ∀stage ∈ stages:
    result = execute(stage) →
    if result.status == "error" then
      return results + [result]
    else
      results = results + [result] →
  results
```

---

## Set Operations

| Symbol | Name | Usage | Example |
|--------|------|-------|---------|
| `∈` | Element of | Membership test | `x ∈ valid_inputs` |
| `∉` | Not element of | Non-membership | `x ∉ blacklist` |
| `⊆` | Subset of | Subset relation | `completed ⊆ total_tasks` |
| `⊇` | Superset of | Superset relation | `features ⊇ required_features` |
| `∪` | Union | Set union | `all = errors ∪ warnings` |
| `∩` | Intersection | Set intersection | `common = set1 ∩ set2` |

### Examples

**Membership**:
```
validate_complexity :: String → Bool
validate_complexity(c) =
  c ∈ {simple, moderate, complex}
```

**Subset**:
```
check_dependencies :: Agent → Bool
check_dependencies(agent) =
  agent.dependencies ⊆ available_agents
```

**Union**:
```
all_issues :: Analysis → [Issue]
all_issues(analysis) =
  analysis.errors ∪ analysis.warnings ∪ analysis.suggestions
```

---

## Comparisons

| Symbol | Name | Usage | Example |
|--------|------|-------|---------|
| `=` | Equals (assignment) | Variable binding | `x = compute(y)` |
| `==` | Equals (comparison) | Equality test | `status == "complete"` |
| `≠` | Not equals | Inequality test | `x ≠ 0` |
| `<` | Less than | Numeric comparison | `count < 10` |
| `>` | Greater than | Numeric comparison | `score > 0.80` |
| `≤` | Less or equal | Numeric comparison | `|code| ≤ 500` |
| `≥` | Greater or equal | Numeric comparison | `coverage ≥ 0.80` |

### Examples

**Assignment**:
```
process :: Input → Output
process(input) =
  validated = validate(input) →
  transformed = transform(validated) →
  output(transformed)
```

**Comparison in Constraints**:
```
constraints :: Code → Bool
constraints(code) =
  |code| ≤ 500 ∧
  coverage(code) ≥ 0.80 ∧
  complexity(code) < 15
```

---

## Special Symbols

| Symbol | Name | Usage | Example |
|--------|------|-------|---------|
| `|x|` | Cardinality/Length | Count elements or length | `|stages| == 5` |
| `Δx` | Delta | Change in value | `Δcoverage = 0.15` |
| `x'` | Prime | Next state | `s' = transition(s)` |
| `x_n` | Subscript | Indexed variable | `iteration_3` |

### Examples

**Cardinality**:
```
count_completed :: [Result] → Int
count_completed(results) =
  |r ∈ results : r.status == "complete"|
```

**Delta**:
```
improvement :: (Before, After) → Float
improvement(before, after) =
  Δquality = after.quality - before.quality →
  Δquality / before.quality
```

**State Transition**:
```
iterate :: State → State
iterate(s) =
  work = execute(s) →
  s' = update_state(s, work) →
  s'
```

---

## Type Signatures

### Basic Syntax

```
function_name :: InputType → OutputType
```

### Examples

**Simple Function**:
```
validate :: Input → Bool
```

**Multiple Inputs**:
```
execute_stage :: (Plan, StageNumber) → StageResult
```

**Higher-Order Function**:
```
map :: (a → b, [a]) → [b]
```

**Constrained Function**:
```
parse :: String → Maybe Requirements  | valid_format
```

---

## Lambda Contracts

### Structure

```
λ(inputs) → outputs | constraints
```

### Examples

**Simple Contract**:
```
λ(code) → refactored_code | |code| ≤ 500
```

**Complex Contract**:
```
λ(feature_spec, todo_ref?) → (plan, execution_report, status) |
  TDD ∧ code_limits ∧ test_coverage ≥ 0.80
```

**With Quantifiers**:
```
λ(plan) → results |
  ∀stage ∈ plan.stages: |code(stage)| ≤ 200
```

---

## Function Definition Patterns

### Simple Definition

```
function_name :: InputType → OutputType
function_name(param) = expression
```

### Multi-Step Definition

```
function_name :: InputType → OutputType
function_name(param) =
  step1 = compute_step1(param) →
  step2 = compute_step2(step1) →
  result = finalize(step2) →
  result
```

### Conditional Definition

```
function_name :: InputType → OutputType
function_name(param) =
  if condition(param) then
    branch1(param)
  else
    branch2(param)
```

### Pattern Matching

```
function_name :: Maybe a → Result
function_name(m) =
  case m of
    Just x → process(x)
    Nothing → error("No value")
```

---

## Constraint Blocks

### Structure

```
constraints :: ContextType → Bool
constraints(ctx) =
  constraint1 ∧
  constraint2 ∧
  constraint3 ∧
  ...
```

### Example

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

## Best Practices

### 1. Use Type Signatures Everywhere

**Good**:
```
parse :: String → Requirements
validate :: Requirements → Bool
execute :: Plan → Result
```

**Bad**:
```
parse(str) = ...  # No type signature
```

### 2. Prefer Symbolic Logic

**Good**:
```
valid(x) ∧ complete(y) ∧ ¬empty(z) → process
```

**Bad**:
```
"Check that x is valid and y is complete and z is not empty, then process"
```

### 3. Use Quantifiers for Loops

**Good**:
```
∀stage ∈ stages: execute(stage) → validate(result)
```

**Bad**:
```
"For each stage in stages, execute it and validate the result"
```

### 4. Explicit Constraints

**Good**:
```
constraints :: Code → Bool
constraints(code) =
  |code| ≤ 500 ∧ coverage(code) ≥ 0.80
```

**Bad**:
```
"Make sure code isn't too long and has good coverage"
```

### 5. Clear Sequencing

**Good**:
```
validate(input) → transform → output → result
```

**Bad**:
```
"First validate, then transform, then output"
```

---

## Common Patterns

### Error Handling

```
process :: Input → Result | Error
process(input) =
  validated = validate(input) →
  if ¬valid(validated) then
    error("Invalid input") →
  transformed = transform(validated) →
  if error(transformed) then
    handle_error(transformed) →
  output(transformed)
```

### Loop with Early Exit

```
execute_until :: Condition → [Stage] → [Result]
execute_until(cond, stages) =
  results = [] →
  ∀stage ∈ stages:
    result = execute(stage) →
    if cond(result) then
      return results + [result]
    else
      results = results + [result] →
  results
```

### Accumulation

```
aggregate :: [Result] → Summary
aggregate(results) =
  total = |results| →
  completed = |r ∈ results : r.status == "complete"| →
  failed = |r ∈ results : r.status == "error"| →
  {total, completed, failed, percentage: completed/total}
```

---

## References

- Lambda Calculus: https://en.wikipedia.org/wiki/Lambda_calculus
- Predicate Logic: https://en.wikipedia.org/wiki/First-order_logic
- Type Theory: https://en.wikipedia.org/wiki/Type_theory
