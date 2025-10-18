# Retrospective Validation Process

**Version**: 1.0
**Framework**: BAIME
**Purpose**: Validate methodologies against historical data post-creation

---

## Overview

Retrospective validation applies a newly created methodology to historical work to measure effectiveness and identify gaps. This validates that the methodology would have improved past outcomes.

---

## Validation Process

### Phase 1: Data Collection (15 min)

**Gather historical data**:
- Session history (JSONL files)
- Error logs and recovery attempts
- Time measurements
- Quality metrics

**Tools**:
```bash
# Query session data
query_tools --status=error
query_user_messages --pattern="error|fail|bug"
query_context --error-signature="..."
```

### Phase 2: Baseline Measurement (15 min)

**Measure pre-methodology state**:
- Error frequency by category
- Mean Time To Recovery (MTTR)
- Prevention opportunities missed
- Quality metrics

**Example**:
```markdown
## Baseline (Without Methodology)

**Errors**: 1336 total
**MTTR**: 11.25 min average
**Prevention**: 0% (no automation)
**Classification**: Ad-hoc, inconsistent
```

### Phase 3: Apply Methodology (30 min)

**Retrospectively apply patterns**:
1. Classify errors using new taxonomy
2. Identify which patterns would apply
3. Calculate time saved per pattern
4. Measure coverage improvement

**Example**:
```markdown
## With Error Recovery Methodology

**Classification**: 1275/1336 = 95.4% coverage
**Patterns Applied**: 10 recovery patterns
**Time Saved**: 8.25 min per error average
**Prevention**: 317 errors (23.7%) preventable
```

### Phase 4: Calculate Impact (20 min)

**Metrics**:
```
Coverage = classified_errors / total_errors
Time_Saved = (MTTR_before - MTTR_after) × error_count
Prevention_Rate = preventable_errors / total_errors
ROI = time_saved / methodology_creation_time
```

**Example**:
```markdown
## Impact Analysis

**Coverage**: 95.4% (1275/1336)
**Time Saved**: 8.25 min × 1336 = 183.6 hours
**Prevention**: 23.7% (317 errors)
**ROI**: 183.6h saved / 5.75h invested = 31.9x
```

### Phase 5: Gap Analysis (15 min)

**Identify remaining gaps**:
- Uncategorized errors (4.6%)
- Patterns needed for edge cases
- Automation opportunities
- Transferability limits

---

## Confidence Scoring

**Formula**:
```
Confidence = 0.4 × coverage +
             0.3 × validation_sample_size +
             0.2 × pattern_consistency +
             0.1 × expert_review

Where:
- coverage = classified / total (0-1)
- validation_sample_size = min(validated/50, 1.0)
- pattern_consistency = successful_applications / total_applications
- expert_review = binary (0 or 1)
```

**Thresholds**:
- Confidence ≥ 0.80: High confidence, production-ready
- Confidence 0.60-0.79: Medium confidence, needs refinement
- Confidence < 0.60: Low confidence, significant gaps

---

## Validation Criteria

**Methodology is validated if**:
1. Coverage ≥ 80% (methodology handles most cases)
2. Time savings ≥ 30% (significant efficiency gain)
3. Prevention ≥ 10% (automation provides value)
4. ROI ≥ 5x (worthwhile investment)
5. Transferability ≥ 70% (broadly applicable)

---

## Example: Error Recovery Validation

**Historical Data**: 1336 errors from 15 sessions

**Baseline**:
- MTTR: 11.25 min
- No systematic classification
- No prevention tools

**Post-Methodology** (retrospective):
- Coverage: 95.4% (13 categories)
- MTTR: 3 min (73% reduction)
- Prevention: 23.7% (3 automation tools)
- Time saved: 183.6 hours
- ROI: 31.9x

**Confidence Score**:
```
Confidence = 0.4 × 0.954 +
             0.3 × 1.0 +
             0.2 × 0.91 +
             0.1 × 1.0
           = 0.38 + 0.30 + 0.18 + 0.10
           = 0.96 (High confidence)
```

**Validation Result**: ✅ VALIDATED (all criteria met)

---

## Common Pitfalls

**❌ Selection Bias**: Only validating on "easy" cases
- Fix: Use complete dataset, include edge cases

**❌ Overfitting**: Methodology too specific to validation data
- Fix: Test transferability on different project

**❌ Optimistic Timing**: Assuming perfect pattern application
- Fix: Use realistic time estimates (1.2x typical)

**❌ Ignoring Learning Curve**: Assuming immediate proficiency
- Fix: Factor in 2-3 iterations to master patterns

---

## Automation Support

**Validation Script**:
```bash
#!/bin/bash
# scripts/validate-methodology.sh

METHODOLOGY=$1
HISTORY_DIR=$2

# Extract baseline metrics
baseline=$(query_tools --scope=session | jq -r '.[] | .duration' | avg)

# Apply methodology patterns
coverage=$(classify_with_patterns "$METHODOLOGY" "$HISTORY_DIR")

# Calculate impact
time_saved=$(calculate_time_savings "$baseline" "$coverage")
prevention=$(calculate_prevention_rate "$METHODOLOGY")

# Generate report
echo "Coverage: $coverage"
echo "Time Saved: $time_saved"
echo "Prevention: $prevention"
echo "ROI: $(calculate_roi "$time_saved" "$methodology_time")"
```

---

**Source**: Bootstrap-003 Error Recovery Retrospective Validation
**Status**: Production-ready, 96% confidence score
**ROI**: 31.9x validated across 1336 historical errors
