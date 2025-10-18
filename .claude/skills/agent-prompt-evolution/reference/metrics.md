# Agent Prompt Metrics

**Version**: 1.0
**Purpose**: Quantitative metrics for measuring agent prompt quality
**Framework**: BAIME dual-layer value functions applied to agents

---

## Core Metrics

### 1. Success Rate

**Definition**: Percentage of tasks completed correctly

**Calculation**:
```
Success Rate = correct_completions / total_tasks
```

**Thresholds**:
- ≥90%: Excellent (production-ready)
- 80-89%: Good (minor refinements needed)
- 60-79%: Fair (needs improvement)
- <60%: Poor (major issues)

**Example**:
```
Tasks: 20
Correct: 17
Partial: 2
Failed: 1

Success Rate = 17/20 = 85% (Good)
```

---

### 2. Quality Score

**Definition**: Average quality rating of agent outputs (1-5 scale)

**Rating Criteria**:
- **5**: Perfect - Accurate, complete, well-structured
- **4**: Good - Minor gaps, mostly complete
- **3**: Fair - Acceptable but needs improvement
- **2**: Poor - Significant issues
- **1**: Failed - Incorrect or unusable

**Thresholds**:
- ≥4.5: Excellent
- 4.0-4.4: Good
- 3.5-3.9: Fair
- <3.5: Poor

**Example**:
```
Task 1: 5/5 (perfect)
Task 2: 4/5 (good)
Task 3: 5/5 (perfect)
...
Task 20: 4/5 (good)

Average: 4.35/5 (Good)
```

---

### 3. Efficiency

**Definition**: Time and token usage per task

**Metrics**:
```
Time Efficiency = avg_time_per_task
Token Efficiency = avg_tokens_per_task
```

**Thresholds** (vary by agent type):
- Explore agent: <3 min, <5k tokens
- Code generation: <5 min, <10k tokens
- Analysis: <10 min, <20k tokens

**Example**:
```
Tasks: 20
Total time: 56 min
Total tokens: 92k

Time Efficiency: 2.8 min/task ✅
Token Efficiency: 4.6k tokens/task ✅
```

---

### 4. Reliability

**Definition**: Consistency of agent performance

**Calculation**:
```
Reliability = 1 - (std_dev(success_rate) / mean(success_rate))
```

**Thresholds**:
- ≥0.90: Very reliable (consistent)
- 0.80-0.89: Reliable
- 0.70-0.79: Moderately reliable
- <0.70: Unreliable (erratic)

**Example**:
```
Batch 1: 85% success
Batch 2: 90% success
Batch 3: 87% success
Batch 4: 88% success

Mean: 87.5%
Std Dev: 2.08
Reliability: 1 - (2.08/87.5) = 0.976 (Very reliable)
```

---

## Composite Metrics

### V_instance (Agent Performance)

**Formula**:
```
V_instance = 0.40 × success_rate +
             0.30 × (quality_score / 5) +
             0.20 × efficiency_score +
             0.10 × reliability

Where:
- success_rate ∈ [0, 1]
- quality_score ∈ [1, 5], normalized to [0, 1]
- efficiency_score = 1 - (actual_time / target_time), capped at [0, 1]
- reliability ∈ [0, 1]
```

**Target**: V_instance ≥ 0.80

**Example**:
```
Success Rate: 85% = 0.85
Quality Score: 4.2/5 = 0.84
Efficiency: 2.8 min / 3 min target = 1 - 0.93 = 0.07, but we want faster so: 1.0 (under budget)
Reliability: 0.976

V_instance = 0.40 × 0.85 +
             0.30 × 0.84 +
             0.20 × 1.0 +
             0.10 × 0.976

           = 0.34 + 0.252 + 0.20 + 0.0976
           = 0.890 ✅ (exceeds target)
```

---

### V_meta (Prompt Quality)

**Formula**:
```
V_meta = 0.35 × completeness +
         0.30 × clarity +
         0.20 × adaptability +
         0.15 × maintainability

Where:
- completeness = features_implemented / features_needed
- clarity = 1 - (ambiguous_instructions / total_instructions)
- adaptability = successful_task_types / tested_task_types
- maintainability = 1 - (prompt_complexity / max_complexity)
```

**Target**: V_meta ≥ 0.80

**Example**:
```
Completeness: 8/8 features = 1.0
Clarity: 1 - (2 ambiguous / 20 instructions) = 0.90
Adaptability: 5/6 task types = 0.83
Maintainability: 1 - (150 lines / 300 max) = 0.50

V_meta = 0.35 × 1.0 +
         0.30 × 0.90 +
         0.20 × 0.83 +
         0.15 × 0.50

       = 0.35 + 0.27 + 0.166 + 0.075
       = 0.861 ✅ (exceeds target)
```

---

## Metric Collection

### Automated Collection

**Session Analysis**:
```bash
# Extract agent performance from session
query_tools --tool="Task" --scope=session | \
  jq -r '.[] | select(.status == "success") | .duration' | \
  awk '{sum+=$1; n++} END {print sum/n}'
```

**Example Script**:
```bash
#!/bin/bash
# scripts/measure-agent-metrics.sh

AGENT_NAME=$1
SESSION=$2

# Success rate
total=$(grep "agent=$AGENT_NAME" "$SESSION" | wc -l)
success=$(grep "agent=$AGENT_NAME.*success" "$SESSION" | wc -l)
success_rate=$(echo "scale=2; $success / $total" | bc)

# Average time
avg_time=$(grep "agent=$AGENT_NAME" "$SESSION" | \
  jq -r '.duration' | \
  awk '{sum+=$1; n++} END {print sum/n}')

# Quality (requires manual rating file)
avg_quality=$(cat "${SESSION}.ratings" | \
  grep "$AGENT_NAME" | \
  awk '{sum+=$2; n++} END {print sum/n}')

echo "Agent: $AGENT_NAME"
echo "Success Rate: $success_rate"
echo "Avg Time: ${avg_time}s"
echo "Avg Quality: $avg_quality/5"
```

---

### Manual Collection

**Test Suite Template**:
```markdown
## Agent Test Suite: [Agent Name]

**Iteration**: [N]
**Date**: [YYYY-MM-DD]

### Test Cases

| ID | Task | Result | Quality | Time | Notes |
|----|------|--------|---------|------|-------|
| 1  | [Description] | ✅/❌ | [1-5] | [min] | [Issues] |
| 2  | [Description] | ✅/❌ | [1-5] | [min] | [Issues] |
...

### Summary

- Success Rate: [X]% ([Y]/[Z])
- Avg Quality: [X.X]/5
- Avg Time: [X.X] min
- V_instance: [X.XX]
```

---

## Benchmarking

### Cross-Agent Comparison

**Standard Test Suite**: 20 representative tasks

**Example Results**:
```
| Agent       | Success | Quality | Time  | V_inst |
|-------------|---------|---------|-------|--------|
| Explore v1  | 60%     | 3.1     | 4.2m  | 0.62   |
| Explore v2  | 87.5%   | 4.2     | 2.8m  | 0.89   |
| Explore v3  | 90%     | 4.3     | 2.6m  | 0.91   |
```

**Improvement**: v1 → v3 = +30% success, +1.2 quality, +38% faster

---

### Baseline Comparison

**Industry Baselines** (approximate):
- Generic agent (no tuning): ~50-60% success
- Basic tuned agent: ~70-80% success
- Well-tuned agent: ~85-95% success
- Expert-tuned agent: ~95-98% success

---

## Regression Testing

### Track Metrics Over Time

**Regression Detection**:
```
if current_metric < (previous_metric - threshold):
    alert("REGRESSION DETECTED")
```

**Thresholds**:
- Success Rate: -5% (e.g., 90% → 85%)
- Quality Score: -0.3 (e.g., 4.5 → 4.2)
- Efficiency: +20% time (e.g., 2.8 min → 3.4 min)

**Example**:
```
Iteration 3: 90% success, 4.3 quality, 2.6 min ✅
Iteration 4: 87% success, 4.1 quality, 2.8 min ⚠️ REGRESSION

Analysis: New constraint too restrictive
Action: Revert constraint, re-test
```

---

## Reporting Template

```markdown
## Agent Metrics Report

**Agent**: [Name]
**Version**: [X.Y]
**Test Date**: [YYYY-MM-DD]
**Test Suite**: [Standard 20 | Custom N]

### Performance Metrics

**Success Rate**: [X]% ([Y]/[Z] tasks)
- Target: ≥85%
- Status: ✅/⚠️/❌

**Quality Score**: [X.X]/5
- Target: ≥4.0
- Status: ✅/⚠️/❌

**Efficiency**:
- Time: [X.X] min/task (target: [Y] min)
- Tokens: [X]k tokens/task (target: [Y]k)
- Status: ✅/⚠️/❌

**Reliability**: [X.XX]
- Target: ≥0.85
- Status: ✅/⚠️/❌

### Composite Scores

**V_instance**: [X.XX]
- Target: ≥0.80
- Status: ✅/⚠️/❌

**V_meta**: [X.XX]
- Target: ≥0.80
- Status: ✅/⚠️/❌

### Comparison to Baseline

| Metric        | Baseline | Current | Δ      |
|---------------|----------|---------|--------|
| Success Rate  | [X]%     | [Y]%    | [+/-]% |
| Quality       | [X.X]    | [Y.Y]   | [+/-]  |
| Time          | [X.X]m   | [Y.Y]m  | [+/-]% |
| V_instance    | [X.XX]   | [Y.YY]  | [+/-]  |

### Recommendations

1. [Action item based on metrics]
2. [Action item based on metrics]

### Next Steps

- [ ] [Task for next iteration]
- [ ] [Task for next iteration]
```

---

**Source**: BAIME Agent Prompt Evolution Framework
**Status**: Production-ready, validated across 13 agent types
**Measurement Overhead**: ~5 min per 20-task test suite
