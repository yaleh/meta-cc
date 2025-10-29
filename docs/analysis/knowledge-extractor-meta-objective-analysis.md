# Knowledge Extractor Meta Objective Analysis
## 如何使 knowledge-extractor 更好地遵循 BAIME Meta Objective

**分析日期**: 2025-10-29
**分析对象**: knowledge-extractor v2.1
**参考案例**: subagent-prompt-construction methodology (V_meta = 0.709)

---

## 问题陈述

knowledge-extractor 在提取 subagent-prompt-construction 实验时，生成的 skill 违反了实验的核心 meta objective（紧凑性）：

**症状**:
- `examples/phase-planner-executor.md`: 393 行（期望 ≤150 行）
- 包含大量分析内容，而非紧凑的可复制示例

**根本原因**:
knowledge-extractor 不理解或不遵循实验的 **V_meta** 定义。

---

## 当前 knowledge-extractor 的 Meta Objective 理解

### 分析 knowledge-extractor.md

#### ✅ 理解的部分

1. **V_instance 验证**:
   ```
   ∧ validation_report.V_instance ≥ 0.85
   ```
   - 知道要验证实例质量
   - 有明确的阈值（0.85）

2. **硬编码的质量约束**:
   ```
   ∧ |lines(SKILL.md)| ≤ 40
   ∧ line_limit(reference/patterns.md) ≤ 400
   ```
   - 对特定文件有行数限制
   - 但这些是"一刀切"的规则

3. **结构验证**:
   ```
   ∧ structure(skill_dir) validated by validate-skill.sh
   ```
   - 有结构检查机制

#### ❌ 缺失的部分

1. **V_meta 完全缺失**:
   - 没有读取 `results.md` 中的 V_meta 定义
   - 没有识别 meta objective 的组件
   - 没有验证提取的 skill 是否反映 meta objective

2. **组件理解缺失**:
   ```
   实验的 V_meta 组件:
   - Compactness: 0.25 weight, target ≤150 lines
   - Generality: 0.20 weight
   - Integration: 0.25 weight
   - Maintainability: 0.15 weight
   - Effectiveness: 0.15 weight
   ```

   knowledge-extractor 不知道：
   - 这个实验的 meta objective 是什么
   - 每个组件的权重和目标
   - 如何在提取中遵循这些组件

3. **动态约束缺失**:
   - 硬编码 `≤40 lines` for SKILL.md
   - 但 `examples/` 没有行数约束
   - 应该从 `V_meta.compactness` 中提取约束

4. **验证报告不完整**:
   ```
   当前: validation_report.V_instance ≥ 0.85
   缺失: validation_report.V_meta_compliance
   ```

---

## 对比：实验 Meta Objective vs. 提取的 Skill

### Case Study: subagent-prompt-construction

#### 实验的 Meta Objective (results.md)

```yaml
V_meta_components:
  compactness:
    weight: 0.25
    measure: "1 - (lines / 150)"
    target: "≤150 lines per subagent"

  generality:
    weight: 0.20
    measure: "successful_domains / total_domains"
    target: "3+ diverse use cases"

  integration:
    weight: 0.25
    measure: "features_used / total_features"
    target: "Use ≥3 Claude Code features"

  maintainability:
    weight: 0.15
    measure: "Subjective 0-1"
    target: "Clear structure, easy to modify"

  effectiveness:
    weight: 0.15
    measure: "success_rate of generated subagents"
    target: "Generated agents work correctly"
```

#### knowledge-extractor 的行为

**应该遵循的约束（从 V_meta 推导）**:
1. **Compactness** (0.25 weight, highest):
   - ✅ SKILL.md ≤ 40 lines（符合）
   - ❌ examples/ ≤ 150 lines（违反：393 lines）
   - ❓ reference/*.md 应该紧凑但详细（模糊）

2. **Integration** (0.25 weight, highest):
   - ✅ 展示 agent composition
   - ✅ 展示 MCP tool usage
   - ✅ 模板中包含 skills_required
   - 总体：遵循较好

3. **Maintainability** (0.15 weight):
   - ✅ 清晰的目录结构
   - ✅ 交叉引用
   - 总体：遵循良好

4. **Generality** (0.20 weight):
   - ❓ 提供模板（支持通用性）
   - ❓ 只有 1 个示例（限制通用性展示）
   - 总体：中性

5. **Effectiveness** (0.15 weight):
   - ✅ 自动化脚本（支持有效性）
   - ✅ 验证脚本
   - 总体：遵循良好

**总体评估**: 50% 遵循，50% 违反/模糊

---

## 核心问题分析

### Problem 1: 没有读取 Meta Objective

**当前流程**:
```
knowledge-extractor 读取:
├── results.md (部分内容)
├── iterations/*.md
└── experiment artifacts

但不解析:
└── V_meta definition
    ├── Components
    ├── Weights
    └── Targets
```

**应该的流程**:
```
knowledge-extractor 读取:
├── results.md
│   ├── V_meta definition ← 解析这个
│   │   ├── Component 1: name, weight, measure, target
│   │   ├── Component 2: ...
│   │   └── ...
│   └── V_instance definition
└── 根据 V_meta 调整提取策略
```

### Problem 2: 硬编码约束 vs. 动态约束

**当前**:
```
# knowledge-extractor.md
∧ |lines(SKILL.md)| ≤ 40          # 硬编码
∧ line_limit(reference/patterns.md) ≤ 400  # 硬编码
# examples/ 没有约束
```

**问题**:
- 不同实验可能有不同的 meta objective
- 例如：documentation-management 可能不强调紧凑性
- 例如：testing-strategy 可能强调覆盖率 > 紧凑性

**应该**:
```
# knowledge-extractor.md
∧ constraints = extract_from_meta_objective(results.md)
∧ ∀component ∈ V_meta.components:
    if component.name == "compactness" then
      apply_compactness_constraint(component.target)
    elif component.name == "integration" then
      ensure_integration_examples(component.target)
    ...
```

### Problem 3: Examples 的定位混淆

**当前理解**:
```
∧ detail(patterns, templates, examples, metrics) →
    reference/*.md ∪ templates/ ∪ examples/
```

这条规则意味着"把详细内容放到这三个地方"，但：
- 没有区分 examples/ 和 reference/ 的职责
- 导致 examples/ 被当作"详细教学内容"的存放地

**应该**:
```
examples/ :: Compact, Copy-Paste Ready
examples/ =
  ∀artifact ∈ validated_artifacts:
    if compactness_required(V_meta) then
      ensure(|artifact| ≤ V_meta.compactness.target)
    copy_or_link(artifact)

reference/ :: Detailed Analysis and Guides
reference/ =
  patterns.md ∪ guides/*.md ∪ case-studies/*.md
  where case-studies/ = detailed_analysis(iterations/)
```

### Problem 4: 验证不完整

**当前**:
```
validation_report = {
  V_instance: 0.895,
  structure_valid: true,
  automation_works: true
}
```

**缺失**:
```
V_meta_compliance: {
  compactness: {
    target: "≤150 lines",
    actual: {
      "SKILL.md": 61,        # ✅
      "examples/*.md": 393,  # ❌
      "reference/*.md": 1400 # ✅ (详细文档允许)
    },
    compliant: false
  },
  integration: { ... },
  ...
}
```

---

## 改进方案

### 方案 1: Meta-Aware Extraction Protocol

#### 1.1 读取和解析 Meta Objective

```markdown
parse_meta_objective :: ResultsFile → MetaObjective
parse_meta_objective(results.md) =
  V_meta_section = extract_section(results.md, "V_meta Component Breakdown") →
  components = ∀row ∈ V_meta_section:
    {
      name: row.component,
      weight: row.weight,
      score: row.score,
      target: infer_target(row.notes)  # 从 notes 推断目标
    } →
  MetaObjective(components, formula)
```

**示例输出**:
```json
{
  "compactness": {
    "weight": 0.25,
    "score": 0.65,
    "target": "≤150 lines",
    "priority": "high"  // weight ≥ 0.20
  },
  "integration": {
    "weight": 0.25,
    "score": 0.857,
    "target": "≥3 Claude Code features",
    "priority": "high"
  },
  ...
}
```

#### 1.2 动态约束生成

```markdown
generate_constraints :: MetaObjective → Constraints
generate_constraints(meta_obj) =
  constraints = {} →

  # Compactness 约束
  if "compactness" ∈ meta_obj.components ∧ meta_obj.compactness.weight ≥ 0.20 then
    target_lines = parse_number(meta_obj.compactness.target) →
    constraints.examples_max_lines = target_lines
    constraints.SKILL_max_lines = min(40, target_lines / 3)

  # Integration 约束
  if "integration" ∈ meta_obj.components ∧ meta_obj.integration.weight ≥ 0.20 then
    constraints.require_integration_examples = true
    constraints.min_features = parse_number(meta_obj.integration.target)

  # ... 其他组件

  return constraints
```

#### 1.3 约束应用

```markdown
apply_constraints :: (Artifacts, Constraints) → ValidatedArtifacts
apply_constraints(artifacts, constraints) =

  # Examples 紧凑性
  ∀example ∈ artifacts.examples:
    if |example| > constraints.examples_max_lines then
      if is_compact_version_possible(example) then
        create_compact_version(example) →
        move_detailed_analysis(example → reference/case-studies/)
      else
        link_to_source(example)

  # Integration 展示
  if constraints.require_integration_examples then
    ensure(∃example: demonstrates_integration(example, constraints.min_features))

  return validated_artifacts
```

#### 1.4 Meta Compliance 验证

```markdown
validate_meta_compliance :: (Skill, MetaObjective) → ComplianceReport
validate_meta_compliance(skill, meta_obj) =
  report = {} →

  # 检查每个高权重组件
  ∀component ∈ meta_obj.components where component.weight ≥ 0.15:
    report[component.name] = check_component_compliance(skill, component)

  report.overall_compliant = ∀c ∈ report: c.compliant
  return report
```

**示例报告**:
```json
{
  "compactness": {
    "target": "≤150 lines",
    "actual": {
      "SKILL.md": 61,
      "examples/phase-planner-executor.md": 109,  // ✅ fixed
      "reference/patterns.md": 418
    },
    "compliant": true,
    "notes": "Examples within target, reference docs allowed to be detailed"
  },
  "integration": {
    "target": "≥3 features",
    "actual": {
      "agents": 2,
      "mcp_tools": 2,
      "skills": 0
    },
    "feature_count": 4,
    "compliant": true
  },
  "overall_compliant": true
}
```

---

### 方案 2: 分层提取策略

#### 2.1 三层输出结构

```
skill/
├── SKILL.md              # 极简概览（≤40 lines）
├── README.md             # 快速开始（≤200 lines）
│
├── examples/             # 紧凑示例（遵循 compactness target）
│   ├── example-1.md      # ≤ V_meta.compactness.target
│   └── README.md         # 指向 case-studies/
│
├── templates/            # 可复用模板
│
├── reference/            # 参考文档（可以详细）
│   ├── patterns.md       # ≤400 lines（硬限制）
│   ├── integration-patterns.md
│   ├── symbolic-language.md
│   └── case-studies/     # 详细分析（不受紧凑性限制）
│       └── example-1-analysis.md
│
├── scripts/              # 自动化脚本
└── inventory/            # 元数据
```

**规则**:
```markdown
extraction_strategy :: (Experiment, MetaObjective) → SkillStructure
extraction_strategy(exp, meta_obj) =

  # Layer 1: 极简层（总是紧凑）
  SKILL.md = extract_lambda_contract(exp) | |SKILL.md| ≤ 40

  # Layer 2: 示例层（根据 meta_obj.compactness 决定）
  if meta_obj.compactness.weight ≥ 0.20 then
    examples/ = compact_examples(exp, meta_obj.compactness.target)
  else
    examples/ = full_examples(exp)

  # Layer 3: 参考层（可以详细，但有结构）
  reference/ = {
    patterns.md | |patterns.md| ≤ 400,
    case-studies/ = detailed_analysis(exp.iterations) | no_line_limit
  }
```

#### 2.2 Examples 处理逻辑

```markdown
process_example :: (Artifact, CompactnessTarget) → Example
process_example(artifact, target) =

  if |artifact| ≤ target then
    # 已经紧凑，直接复制
    copy(artifact → examples/)

  elif is_source_available(artifact) then
    # 链接到源文件
    link(artifact → examples/)

  else
    # 需要拆分
    compact_part = extract_core_definition(artifact) →
    analysis_part = extract_analysis(artifact) →

    copy(compact_part → examples/) |
      |compact_part| ≤ target ∧
      add_reference(compact_part → case-studies/)

    copy(analysis_part → reference/case-studies/)
```

---

### 方案 3: Config.json 驱动的提取

#### 3.1 实验配置文件

```json
// experiments/subagent-prompt-methodology/config.json
{
  "experiment": {
    "name": "subagent-prompt-construction",
    "domain": "Subagent prompt construction",
    "status": "near_convergence",
    "v_meta": 0.709,
    "v_instance": 0.895
  },

  "meta_objective": {
    "components": [
      {
        "name": "compactness",
        "weight": 0.25,
        "priority": "high",
        "targets": {
          "subagent_prompts": "≤150 lines",
          "examples": "≤150 lines",
          "SKILL.md": "≤40 lines"
        },
        "enforcement": "strict"
      },
      {
        "name": "integration",
        "weight": 0.25,
        "priority": "high",
        "targets": {
          "min_features": 3,
          "feature_types": ["agents", "mcp_tools", "skills"]
        },
        "enforcement": "validate"
      },
      {
        "name": "maintainability",
        "weight": 0.15,
        "priority": "medium",
        "targets": {
          "structure": "clear",
          "cross_references": "extensive"
        },
        "enforcement": "best_effort"
      }
    ]
  },

  "extraction_rules": {
    "examples_strategy": "compact_only",  // compact_only | detailed | hybrid
    "case_studies": true,
    "automation_priority": "high"
  }
}
```

#### 3.2 Config-Driven Extraction

```markdown
extract_with_config :: (Experiment, Config) → Skill
extract_with_config(exp, config) =

  # 读取 meta objective
  meta_obj = config.meta_objective →

  # 生成约束
  constraints = generate_constraints(meta_obj) →

  # 应用提取规则
  if config.extraction_rules.examples_strategy == "compact_only" then
    examples = create_compact_examples(exp, constraints.examples_max_lines)
  elif config.extraction_rules.examples_strategy == "hybrid" then
    examples = create_hybrid_examples(exp, constraints)

  # 创建 case studies（如果配置启用）
  if config.extraction_rules.case_studies then
    case_studies = create_case_studies(exp.iterations)

  # 验证合规性
  compliance = validate_compliance(skill, meta_obj) →
  assert(compliance.overall_compliant)

  return skill
```

---

## 具体改进建议

### Immediate Fix (v2.1 → v2.2)

**修改 knowledge-extractor.md**:

```markdown
# 添加 meta objective 解析
parse_meta_objective :: ResultsFile → MetaObjective
parse_meta_objective(results) =
  read(results.md) →
  extract_section("V_meta Component Breakdown") →
  parse_components(weight, score, target)

# 添加动态约束
apply_meta_constraints :: (MetaObjective, Artifacts) → ValidatedArtifacts
apply_meta_constraints(meta_obj, artifacts) =

  # Compactness 约束（如果 weight ≥ 0.20）
  if meta_obj.compactness.weight ≥ 0.20 then
    target = parse_number(meta_obj.compactness.target) →
    ∀example ∈ artifacts.examples:
      ensure(|example| ≤ target ∨ is_link(example))

  # Integration 约束
  if meta_obj.integration.weight ≥ 0.20 then
    min_features = parse_number(meta_obj.integration.target) →
    ensure(∃example: feature_count(example) ≥ min_features)

  return validated_artifacts

# 更新主流程
λ(experiment_dir, skill_name, options?) → (skill_dir, knowledge_entries, validation_report) |
  ∧ meta_obj = parse_meta_objective(experiment_dir/results.md)    # 新增
  ∧ constraints = generate_constraints(meta_obj)                   # 新增
  ∧ artifacts = extract_artifacts(experiment_dir)
  ∧ validated_artifacts = apply_meta_constraints(meta_obj, artifacts)  # 新增
  ∧ compliance_report = validate_meta_compliance(skill, meta_obj)      # 新增
  ...
  ∧ validation_report = {
      V_instance: ...,
      V_meta_compliance: compliance_report  # 新增
    }
```

### Medium-Term Enhancement (v3.0)

1. **Config.json 支持**:
   - 允许实验定义 config.json
   - Config 中明确 meta objective 和提取规则
   - knowledge-extractor 优先使用 config

2. **Case Studies 目录**:
   - `reference/case-studies/` 存放详细分析
   - Examples 保持紧凑
   - 自动交叉引用

3. **Meta Compliance Dashboard**:
   - 生成 `COMPLIANCE.md`
   - 列出所有 meta objective 组件的合规性
   - 提供改进建议

---

## 验证方案

### Test Case 1: subagent-prompt-construction

**Input**:
- V_meta.compactness.weight = 0.25 (high priority)
- V_meta.compactness.target = "≤150 lines"

**Expected Output**:
- ✅ SKILL.md ≤ 40 lines
- ✅ examples/phase-planner-executor.md ≤ 150 lines
- ✅ reference/case-studies/phase-planner-executor-analysis.md (no limit)
- ✅ Compliance report shows compactness: compliant

### Test Case 2: documentation-management (hypothetical)

**Input**:
- V_meta.compactness.weight = 0.10 (low priority)
- V_meta.completeness.weight = 0.30 (high priority)

**Expected Output**:
- ✅ SKILL.md ≤ 40 lines（总是紧凑）
- ✅ examples/ 可以详细（compactness 不是高优先级）
- ✅ reference/ 强调完整性（大量文档）

---

## 总结

### 核心问题

knowledge-extractor v2.1 **不理解或不遵循** BAIME 实验的 meta objective，导致：
1. 提取的 skill 违反实验的核心约束（如紧凑性）
2. 验证报告不完整（只有 V_instance，缺 V_meta compliance）
3. 硬编码约束无法适应不同实验的 meta objective

### 关键改进

1. **Parse Meta Objective**: 从 results.md 读取 V_meta 定义
2. **Dynamic Constraints**: 根据 meta objective 生成提取约束
3. **Meta Compliance**: 验证提取的 skill 是否符合 meta objective
4. **Layered Structure**: Examples 紧凑 + Case Studies 详细

### 优先级

**P0 (立即修复)**:
- Examples 紧凑性约束
- 创建 case-studies/ 目录
- 手动重组 subagent-prompt-construction

**P1 (v2.2)**:
- Parse meta objective from results.md
- Dynamic compactness constraints
- Meta compliance validation

**P2 (v3.0)**:
- Config.json support
- Full meta objective component support
- Compliance dashboard

---

**建议**：先手动修复 subagent-prompt-construction，然后升级 knowledge-extractor 到 v2.2。
