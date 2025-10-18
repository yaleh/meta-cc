# Technical Debt Management Methodology - Reference

This reference documentation provides comprehensive details on the SQALE-based technical debt quantification methodology developed in bootstrap-012.

## Core Methodology Components

**Six Components** (complete methodology):
1. Measurement Framework (SQALE Index calculation)
2. Categorization Framework (Code smell taxonomy)
3. Prioritization Framework (Value-effort matrix)
4. Paydown Framework (Phased roadmap)
5. Tracking Framework (Trend analysis)
6. Prevention Framework (Proactive practices)

## SQALE Methodology

**SQALE (Software Quality Assessment based on Lifecycle Expectations)**:
- Industry-standard debt quantification
- Development cost: LOC / 30 (30 LOC/hour productivity)
- Remediation cost: Graduated complexity thresholds
- TD Ratio: (Debt / Development Cost) × 100%
- Rating: A (≤5%) to E (>50%)

## Knowledge Artifacts

All knowledge artifacts from bootstrap-012 are documented in:
`experiments/bootstrap-012-technical-debt/knowledge/`

**Patterns** (3):
- SQALE-Based Debt Quantification (90% reusable)
- Code Smell Taxonomy Mapping (80% reusable)
- Value-Effort Prioritization Matrix (95% reusable)

**Principles** (3):
- Pay High-Value Low-Effort Debt First
- SQALE Provides Objective Baseline
- Complexity Drives Maintainability Debt

**Templates** (4):
- SQALE Index Report Template
- Code Smell Categorization Template
- Remediation Cost Breakdown Template
- Transfer Guide Template

**Best Practices** (3):
- Use SQALE standard productivity (30 LOC/hour)
- Apply graduated complexity thresholds
- Categorize debt by SQALE characteristics

## Effectiveness Validation

**Speedup**: 4.5x vs manual approach
- Manual: 9 hours (ad-hoc review, subjective)
- Methodology: 2 hours (tool-based, SQALE)

**Accuracy**: Subjective → Objective (SQALE standard)
**Reproducibility**: Low → High (industry standard)

## Transferability

**Overall**: 85% transferable across languages

**Language-Specific Adaptations**:
- Go: 90% (native)
- Python: 85% (threshold 10→12, tools: radon, pylint, pytest-cov)
- JavaScript: 85% (threshold 10→8, tools: eslint, jscpd, nyc)
- Java: 90% (tools: PMD, JaCoCo, CheckStyle)
- Rust: 80% (threshold 10→15, tools: cargo-geiger, clippy, skip OO smells)

**Universal Components** (13/16, 81%):
- SQALE formulas (100%)
- Prioritization matrix (100%)
- Paydown roadmap structure (100%)
- Tracking approach (95%)
- Prevention practices (85%)

**Language-Specific** (3/16, 19%):
- Complexity threshold calibration (±20%)
- Tool selection (language-specific)
- OO smells applicability (OO languages only)

## Experiment Results

See full results: `experiments/bootstrap-012-technical-debt/results.md`

**Key Metrics**:
- V_instance = 0.805 (CONVERGED)
- V_meta = 0.855 (CONVERGED)
- 4 iterations, ~7 hours total
- 4.5x speedup, 85% transferability
- meta-cc debt: 66 hours, 15.52% TD ratio, rating C
- Paydown roadmap: 31.5 hours → rating B (8.23%)
