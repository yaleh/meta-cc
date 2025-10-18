# Cross-Cutting Concerns Management - Reference

This reference documentation provides comprehensive details on the cross-cutting concerns standardization methodology developed in bootstrap-013.

## Core Methodology

**Systematic standardization of**: Error handling, Logging, Configuration

**Three Phases**:
1. Observe (Pattern inventory, baseline metrics, gap analysis)
2. Codify (Convention selection, infrastructure creation, linter development)
3. Automate (Standardization, CI integration, documentation)

## Five Universal Principles

1. **Detect Before Standardize**: Automate identification of non-compliant code
2. **Prioritize by Value**: High-value files first (ROI-based classification)
3. **Infrastructure Enables Scale**: Build sentinels before standardizing call sites
4. **Context Is King**: Enrich errors with operation + resource + type + guidance
5. **Automate Enforcement**: CI blocks non-compliant code

## Knowledge Artifacts

All knowledge artifacts from bootstrap-013 are documented in:
`experiments/bootstrap-013-cross-cutting-concerns/knowledge/`

**Best Practices** (3):
- Go Logging (13 practices)
- Go Error Handling (13 practices)
- Go Configuration (14 practices)

**Templates** (3):
- Logger Setup (log/slog initialization)
- Error Handling Template (sentinel errors, wrapping, context)
- Config Management Template (centralized config, validation)

## File Tier Prioritization

**Tier 1 (ROI > 10x)**: User-facing APIs, public interfaces, error infrastructure
- **Strategy**: Standardize 100%
- **Example**: capabilities.go (16.7x ROI, 25.5% value gain)

**Tier 2 (ROI 5-10x)**: Internal services, CLI commands, data processors
- **Strategy**: Selective standardization 50-80%
- **Example**: Internal utilities (8.3x ROI, 6% value gain)

**Tier 3 (ROI < 5x)**: Test utilities, stubs, deprecated code
- **Strategy**: Defer or skip 0-20%
- **Example**: Stubs (3x ROI, 1% value gain) - deferred

## Effectiveness Validation

**Error Diagnosis Speed**: 60-75% faster with rich context

**ROI by Tier**:
- Tier 1: 16.7x ROI
- Tier 2: 8.3x ROI
- Tier 3: 3x ROI (deferred)

**CI Enforcement**:
- Setup time: 20 minutes
- Regression rate: 0%
- Ongoing maintenance: 0 hours (fully automated)

## Transferability

**Overall**: 80-90% transferable across languages

**Language-Specific Adaptations**:
- Go: 90% (log/slog, fmt.Errorf %w, os.Getenv)
- Python: 80-85% (logging, raise...from, os.environ)
- JavaScript: 75-80% (winston, Error.cause, process.env)
- Rust: 85-90% (tracing, anyhow, thiserror)

**Universal Components** (100%):
- 5 universal principles
- File tier prioritization framework
- ROI calculation method
- Context enrichment structure (operation + resource + type + guidance)

**Language-Specific** (10-20%):
- Specific libraries/tools
- Syntax variations
- Error wrapping mechanisms

## Experiment Results

See full results: `experiments/bootstrap-013-cross-cutting-concerns/` (in progress)

**Key Metrics**:
- Error handling: 70% → 90% consistency (Tier 1)
- Logging: 0.7% → 90% adoption
- Configuration: 40% → 80% centralized
- ROI: 16.7x for Tier 1, 8.3x for Tier 2
- Diagnosis speed: 60-75% improvement
