# Confidence Scoring Methodology

**Version**: 1.0
**Purpose**: Quantify validation confidence for methodologies
**Range**: 0.0-1.0 (threshold: 0.80 for production)

---

## Confidence Formula

```
Confidence = 0.4 × coverage +
             0.3 × validation_sample_size +
             0.2 × pattern_consistency +
             0.1 × expert_review

Where all components ∈ [0, 1]
```

---

## Component 1: Coverage (40% weight)

**Definition**: Percentage of cases methodology handles

**Calculation**:
```
coverage = handled_cases / total_cases
```

**Example** (Error Recovery):
```
coverage = 1275 classified / 1336 total
         = 0.954
```

**Thresholds**:
- 0.95-1.0: Excellent (comprehensive)
- 0.80-0.94: Good (most cases covered)
- 0.60-0.79: Fair (significant gaps)
- <0.60: Poor (incomplete)

---

## Component 2: Validation Sample Size (30% weight)

**Definition**: How much data was used for validation

**Calculation**:
```
validation_sample_size = min(validated_count / 50, 1.0)
```

**Rationale**: 50+ validated cases provides statistical confidence

**Example** (Error Recovery):
```
validation_sample_size = min(1336 / 50, 1.0)
                       = min(26.72, 1.0)
                       = 1.0
```

**Thresholds**:
- 50+ cases: 1.0 (high confidence)
- 20-49 cases: 0.4-0.98 (medium confidence)
- 10-19 cases: 0.2-0.38 (low confidence)
- <10 cases: <0.2 (insufficient data)

---

## Component 3: Pattern Consistency (20% weight)

**Definition**: Success rate when patterns are applied

**Calculation**:
```
pattern_consistency = successful_applications / total_applications
```

**Measurement**:
1. Apply each pattern to 5-10 representative cases
2. Count successes (problem solved correctly)
3. Calculate success rate per pattern
4. Average across all patterns

**Example** (Error Recovery):
```
Pattern 1 (Fix-and-Retry): 9/10 = 0.90
Pattern 2 (Test Fixture): 10/10 = 1.0
Pattern 3 (Path Correction): 8/10 = 0.80
...
Pattern 10 (Permission Fix): 10/10 = 1.0

Average: 91/100 = 0.91
```

**Thresholds**:
- 0.90-1.0: Excellent (reliable patterns)
- 0.75-0.89: Good (mostly reliable)
- 0.60-0.74: Fair (needs refinement)
- <0.60: Poor (unreliable)

---

## Component 4: Expert Review (10% weight)

**Definition**: Binary validation by domain expert

**Values**:
- 1.0: Reviewed and approved by expert
- 0.5: Partially reviewed or peer-reviewed
- 0.0: Not reviewed

**Review Criteria**:
1. Patterns are correct and complete
2. No critical gaps identified
3. Transferability claims validated
4. Automation tools tested
5. Documentation is accurate

**Example** (Error Recovery):
```
expert_review = 1.0 (fully reviewed and validated)
```

---

## Complete Example: Error Recovery

**Component Values**:
```
coverage = 1275/1336 = 0.954
validation_sample_size = min(1336/50, 1.0) = 1.0
pattern_consistency = 91/100 = 0.91
expert_review = 1.0 (reviewed)
```

**Confidence Calculation**:
```
Confidence = 0.4 × 0.954 +
             0.3 × 1.0 +
             0.2 × 0.91 +
             0.1 × 1.0

           = 0.382 + 0.300 + 0.182 + 0.100
           = 0.964
```

**Interpretation**: **96.4% confidence** (High - Production Ready)

---

## Confidence Bands

### High Confidence (0.80-1.0)

**Characteristics**:
- ≥80% coverage
- ≥20 validated cases
- ≥75% pattern consistency
- Reviewed by expert

**Actions**: Deploy to production, recommend broadly

**Example Methodologies**:
- Error Recovery (0.96)
- Testing Strategy (0.87)
- CI/CD Pipeline (0.85)

---

### Medium Confidence (0.60-0.79)

**Characteristics**:
- 60-79% coverage
- 10-19 validated cases
- 60-74% pattern consistency
- May lack expert review

**Actions**: Use with caution, monitor results, refine gaps

**Example**:
- New methodology with limited validation
- Partial coverage of domain

---

### Low Confidence (<0.60)

**Characteristics**:
- <60% coverage
- <10 validated cases
- <60% pattern consistency
- Not reviewed

**Actions**: Do not use in production, requires significant refinement

**Example**:
- Untested methodology
- Insufficient validation data

---

## Adjustments for Domain Complexity

**Adjust thresholds for complex domains**:

**Simple Domain** (e.g., file operations):
- Target: 0.85+ (higher expectations)
- Coverage: ≥90%
- Patterns: 3-5 sufficient

**Medium Domain** (e.g., testing):
- Target: 0.80+ (standard)
- Coverage: ≥80%
- Patterns: 6-8 typical

**Complex Domain** (e.g., distributed systems):
- Target: 0.75+ (realistic)
- Coverage: ≥70%
- Patterns: 10-15 needed

---

## Confidence Over Time

**Track confidence across iterations**:

```
Iteration 0: N/A (baseline only)
Iteration 1: 0.42 (low - initial patterns)
Iteration 2: 0.63 (medium - expanded)
Iteration 3: 0.79 (approaching target)
Iteration 4: 0.88 (high - converged)
Iteration 5: 0.87 (stable)
```

**Convergence**: Confidence stable ±0.05 for 2 iterations

---

## Confidence vs. V_meta

**Different but related**:

**V_meta**: Methodology quality (completeness, transferability, automation)
**Confidence**: Validation strength (how sure we are V_meta is accurate)

**Relationship**:
- High V_meta, Low Confidence: Good methodology, insufficient validation
- High V_meta, High Confidence: Production-ready
- Low V_meta, High Confidence: Well-validated but incomplete methodology
- Low V_meta, Low Confidence: Needs significant work

---

## Reporting Template

```markdown
## Validation Confidence Report

**Methodology**: [Name]
**Version**: [X.Y]
**Validation Date**: [YYYY-MM-DD]

### Confidence Score: [X.XX]

**Components**:
- Coverage: [X.XX] ([handled]/[total] cases)
- Sample Size: [X.XX] ([count] validated cases)
- Pattern Consistency: [X.XX] ([successes]/[applications])
- Expert Review: [X.XX] ([status])

**Confidence Band**: [High/Medium/Low]

**Recommendation**: [Deploy/Refine/Rework]

**Gaps Identified**:
1. [Gap description]
2. [Gap description]

**Next Steps**:
1. [Action item]
2. [Action item]
```

---

## Automation

**Confidence Calculator**:
```bash
#!/bin/bash
# scripts/calculate-confidence.sh

METHODOLOGY=$1
HISTORY=$2

# Calculate coverage
coverage=$(calculate_coverage "$METHODOLOGY" "$HISTORY")

# Calculate sample size
sample_size=$(count_validated_cases "$HISTORY")
sample_score=$(echo "scale=2; if ($sample_size >= 50) 1.0 else $sample_size/50" | bc)

# Calculate pattern consistency
consistency=$(measure_pattern_consistency "$METHODOLOGY")

# Expert review (manual input)
expert_review=${3:-0.0}

# Calculate confidence
confidence=$(echo "scale=3; 0.4*$coverage + 0.3*$sample_score + 0.2*$consistency + 0.1*$expert_review" | bc)

echo "Confidence: $confidence"
echo "  Coverage: $coverage"
echo "  Sample Size: $sample_score"
echo "  Consistency: $consistency"
echo "  Expert Review: $expert_review"
```

---

**Source**: BAIME Retrospective Validation Framework
**Status**: Production-ready, validated across 13 methodologies
**Average Confidence**: 0.86 (median 0.87)
