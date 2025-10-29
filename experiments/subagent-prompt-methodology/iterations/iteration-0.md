# BAIME Experiment: Subagent Prompt Construction Methodology
## Iteration 0 - Baseline Definition

---

## Metadata

- **Iteration**: 0
- **Date**: 2025-10-29
- **Status**: In Progress
- **Experiment Type**: Methodology Development

---

## Dual-Layer Objectives

### Meta-Level Objective (V_meta)

**Goal**: Develop a universal, reusable methodology for constructing Claude Code subagent prompts that are compact, effective, and leverage Claude Code features.

**Quality Components** (weighted):

| Component | Weight | Description | Measurement |
|-----------|--------|-------------|-------------|
| **Compactness** | 0.25 | Formal language, minimal verbosity, ≤150 lines | `1 - (lines / 150)` |
| **Generality** | 0.20 | Cross-domain applicability, templateable | `successful_domains / total_domains` |
| **Integration** | 0.25 | Uses Claude Code features (skills/subagents/MCP) | `features_used / total_features` |
| **Maintainability** | 0.15 | Clear structure, easy to modify | Subjective 0-1 |
| **Effectiveness** | 0.15 | Generated subagents work correctly | `success_rate` |

**V_meta Formula**:
```
V_meta = 0.25 * compactness + 0.20 * generality + 0.25 * integration +
         0.15 * maintainability + 0.15 * effectiveness
```

**Convergence Threshold**: V_meta ≥ 0.75

---

### Instance-Level Objective (V_instance)

**Goal**: Construct a concrete `phase-planner-executor` subagent.

**Rationale**: Based on user messages analysis showing frequent requests for:
- "Plan and execute a new phase for a specific new feature"
- Integration of project-planner + stage-executor workflow

**Quality Components** (weighted):

| Component | Weight | Description | Measurement |
|-----------|--------|-------------|-------------|
| **Planning Quality** | 0.30 | Generates valid phase plans | Plan correctness 0-1 |
| **Execution Quality** | 0.30 | Executes stages successfully | Execution success rate |
| **Integration Quality** | 0.20 | Uses Claude Code features | `features_used / 4` (target: 4 features) |
| **Output Quality** | 0.20 | Clear, structured reports | Output quality 0-1 |

**V_instance Formula**:
```
V_instance = 0.30 * planning_quality + 0.30 * execution_quality +
             0.20 * integration_quality + 0.20 * output_quality
```

**Success Criteria**:
- V_instance ≥ 0.80
- Successfully plans and executes a phase from TODO.md
- Prompt ≤150 lines
- Uses ≥3 Claude Code features

---

## Baseline Analysis

### Existing Patterns Identified

#### 1. Structural Pattern

```
---
name: {agent_name}
description: {one_line_description}
---

λ(inputs) → outputs | constraints

{type_signatures}
{function_definitions}
{main_flow}
```

#### 2. Symbolic Language System

**Logic Operators**:
- `∧` (AND), `∨` (OR), `¬` (NOT), `→` (implies), `↔` (bidirectional)

**Quantifiers**:
- `∀` (for all), `∃` (exists)

**Set Operations**:
- `∈` (in), `⊆` (subset), `∪` (union), `∩` (intersection), `⊇` (superset)

**Comparisons**:
- `≤`, `≥`, `=`, `==`, `<`, `>`

**Special Symbols**:
- `|x|` (cardinality/length)
- `Δx` (delta/change)
- `x'` (x prime, next state)

#### 3. Type Signature Pattern

```
function_name :: InputType → OutputType
```

Examples:
```
pre_execution :: Experiment → Context
lifecycle_execution :: (M, Context, A) → (Output, M', A')
domain_analysis :: Experiment → Domain
```

#### 4. Constraint Expression Pattern

```
constraints :: Type → Bool
constraints(x) =
  condition1 ∧ condition2 ∧ ¬anti_pattern ∧ ...
```

#### 5. Reference Mechanisms Found

**File/Capability Reading**:
```
read(file_path)
read(capability)
load(definition)
```

**Examples**:
- `read(iteration_{n-1}.md)`
- `read(meta-agents/*.md)`
- `load(lifecycle_capabilities)`

**Note**: Direct skill/subagent references not yet observed in existing prompts. This is an opportunity for enhancement.

---

## Pattern Analysis Details

### iteration-executor.md Analysis

**Lambda Contract**:
```
λ(experiment, iteration_n) → (M_n, A_n, s_n, V(s_n), convergence) | ∀i ∈ iterations
```

**Key Functions** (9 total):
1. `pre_execution :: Experiment → Context`
2. `meta_agent_context :: M_i → Capabilities`
3. `lifecycle_execution :: (M, Context, A) → (Output, M', A')`
4. `insufficiency_evaluation :: (A, Strategy) → Bool`
5. `system_evolution :: (M, A, Evidence) → (M', A')`
6. `dual_value_calculation :: Output → (V_instance, V_meta, Gaps)`
7. `convergence_evaluation :: (...) → Bool`
8. `state_transition :: (s_{n-1}, Work) → s_n`
9. `documentation :: Iteration → Report`

**Lines**: 108 lines
**Compactness Score**: 1 - (108/150) = 0.28

**Features Used**:
- ✓ Function composition
- ✓ Type signatures
- ✓ Constraint blocks
- ✓ File references (`read()`)
- ✗ Skill references
- ✗ Subagent composition
- ✗ MCP tool usage

**Integration Score**: 4/7 = 0.57

---

### project-planner.md Analysis

**Lambda Contract**:
```
λ(docs, state) → plan | ∀i ∈ iterations
```

**Lines**: 17 lines
**Compactness Score**: 1 - (17/150) = 0.89

**Key Constraints**:
```
∧ |code(i)| ≤ 500 ∧ |test(i)| ≤ 500
∧ ∀s ∈ stages(i): |code(s)| ≤ 200 ∧ |test(s)| ≤ 200
∧ ¬impl ∧ +interfaces
```

**Notable**: Very compact, single-expression style

---

### knowledge-extractor.md Analysis

**Lambda Contract**:
```
λ(experiment_dir, skill_name, options?) → (skill_dir, knowledge_entries, validation_report)
```

**Lines**: 31 lines
**Compactness Score**: 1 - (31/150) = 0.79

**Key Features**:
- Uses optional parameters (`options?`)
- Has validation predicates
- References automation scripts
- Includes version metadata

---

### stage-executor.md Analysis

**Lambda Contract**:
```
λ(plan, constraints) → execution | ∀stage ∈ plan
```

**Lines**: 52 lines
**Compactness Score**: 1 - (52/150) = 0.65

**Key Functions** (8 total):
1. `pre_analysis :: Plan → Validated_Plan`
2. `environment :: System → Ready_State`
3. `execute :: Stage → Result`
4. `pre_commit_hooks :: Code_Changes → Quality_Gate`
5. `quality_assurance :: Result → Validated_Result`
6. `status_matrix :: Task → Status_Report`
7. `risk_assessment :: Issue → Risk_Level`
8. `development_standards :: Code → Validated_Code`

**Features**:
- Includes external reference: `https://pre-commit.com/`
- Has cleanup procedures
- Status matrix with enums

---

### iteration-prompt-designer.md Analysis

**Lambda Contract**:
```
λ(experiment_spec, domain) → ITERATION-PROMPTS.md | structured_for_iteration-executor
```

**Lines**: 136 lines
**Compactness Score**: 1 - (136/150) = 0.09

**Key Functions** (9 total):
1. `domain_analysis :: Experiment → Domain`
2. `architecture_design :: Domain → ArchitectureSpec`
3. `value_function_design :: Domain → (ValueSpec_Instance, ValueSpec_Meta)`
4. `baseline_iteration_spec :: Domain → Iteration0`
5. `subsequent_iteration_spec :: Domain → IterationN`
6. `knowledge_organization_spec :: Domain → KnowledgeSpec`
7. `results_analysis_spec :: Domain → ResultsTemplate`
8. `execution_guidance :: Domain → ExecutionGuide`
9. `template_composition :: (...) → Document`

**Characteristics**:
- Most comprehensive
- Near 150-line limit
- Multiple complex specifications

---

## Baseline Methodology (V0)

### Template Structure

```markdown
---
name: {agent_name}
description: {one_line_task_description}
---

λ({input_params}) → {outputs} | {high_level_constraints}

{type_signature_1} :: {InputType} → {OutputType}
{function_1}({params}) = {definition}

{type_signature_2} :: {InputType} → {OutputType}
{function_2}({params}) = {definition}

...

{main_execution_flow} :: {InputType} → {OutputType}
{main_flow}({params}) =
  {step_1} →
  {step_2} →
  ...
  {result}
```

### Compactness Techniques

1. **Use symbolic logic**: `∧`, `∨`, `¬`, `→` instead of prose
2. **Type annotations**: `function :: Type → Type`
3. **Lambda contracts**: `λ(x) → y | constraints`
4. **Function composition**: Chain operations with `→`
5. **Predicate logic**: Express constraints as predicates
6. **Avoid redundancy**: Factor common patterns

### Missing Elements (To Add)

1. **Skill references**: `skill(name) :: Domain → Result`
2. **Subagent composition**: `agent(type, params) → Output`
3. **MCP tool usage**: `mcp::tool_name(params) → Data`

---

## Initial V_meta Assessment

### Component Scores

| Component | Score | Calculation | Notes |
|-----------|-------|-------------|-------|
| Compactness | 0.54 | avg(0.28, 0.89, 0.79, 0.65, 0.09) | Wide variance |
| Generality | 0.40 | 2/5 prompts easily adaptable | Moderate |
| Integration | 0.40 | 4/10 features used | Missing skills/agents/MCP |
| Maintainability | 0.70 | Clear structure | Subjective |
| Effectiveness | 0.85 | All existing agents work | High |

**V_meta (Iteration 0)**:
```
V_meta = 0.25 * 0.54 + 0.20 * 0.40 + 0.25 * 0.40 + 0.15 * 0.70 + 0.15 * 0.85
       = 0.135 + 0.08 + 0.10 + 0.105 + 0.1275
       = 0.5475
```

**Assessment**: Below convergence threshold (0.75). Key gaps:
- Integration score low (missing Claude Code features)
- Generality could improve (templates not yet formalized)
- Compactness variance too high

---

## Next Iteration Focus

### Priorities for Iteration 1

1. **Enhance Integration** (target: 0.70)
   - Design skill reference syntax
   - Design subagent composition syntax
   - Design MCP tool usage syntax
   - Update template to include these

2. **Improve Generality** (target: 0.60)
   - Create parameterized template
   - Test on different domains
   - Document adaptation process

3. **Stabilize Compactness** (target: 0.65)
   - Define optimal line count range (40-100 lines)
   - Create guidelines for when to decompose

### Concrete Tasks

- [ ] Design Claude Code feature integration syntax
- [ ] Create template with feature placeholders
- [ ] Build phase-planner-executor using new template
- [ ] Test on TODO.md item
- [ ] Measure V_instance and V_meta
- [ ] Iterate based on gaps

---

## Artifacts

### Files Created
- `experiments/subagent-prompt-methodology/iteration-0.md` (this file)

### Data Collected
- 5 existing subagent prompts analyzed
- Pattern library extracted
- Baseline V_meta calculated: 0.5475

---

## Reflection

### What Worked
- Systematic analysis of existing prompts
- Clear pattern extraction
- Quantitative baseline measurement

### What Didn't Work
- Haven't yet explored Claude Code feature integration
- No concrete instance built yet

### Insights
1. Existing prompts don't leverage skills/subagents/MCP
2. High variance in compactness suggests need for guidelines
3. Lambda contract style is well-established and effective

### Next Steps
1. Research Claude Code skill/subagent invocation syntax
2. Design integration patterns
3. Build phase-planner-executor instance
4. Evaluate and iterate

---

**End of Iteration 0**
