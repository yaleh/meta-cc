# Software Development Methodology

This directory contains universal, project-independent methodology guides for software development with Claude Code.

## Available Guides

### [Documentation Management](documentation-management.md)
A comprehensive, language-agnostic guide to managing documentation in software projects using Claude Code.

**Key topics**:
- Phase 0 bootstrap process (plan.md and principles.md)
- Core principles (DRY, Progressive Disclosure, Task-Oriented)
- Commit/Merge/Release checklists
- Synchronization decision matrix
- Plans directory and agent-assisted workflow
- Document size guidelines
- Directory structure reference

**Target audience**: All Claude Code projects

**Status**: Operational manual (v5.0)

### [Role-Based Documentation Architecture](role-based-documentation.md)
Data-driven methodology for organizing and maintaining documentation based on actual usage patterns, with automated health checks and continuous optimization.

**Key topics**:
- 6 document roles (Context Base, Living, Specification, Reference, Episodic, Archive)
- Key metrics (R/E ratio, access density, lifecycle stages)
- Automated health checks (/meta doc-health, doc-evolution, doc-gaps, doc-usage)
- Implementation guide with empirical case study
- Integration with existing workflows

**Target audience**: Projects with meta-cc integration

**Status**: Methodology v1.0 (2025-10-13)

### [Empirical Methodology Development](empirical-methodology-development.md)
Meta-methodology for developing software engineering practices through observation, analysis, and automation. Based on meta-cc project experience.

**Key topics**:
- Empirical Evolutionism: Core meta-principles
- OCA Framework (Observe-Codify-Automate)
- Case study: meta-cc development process analysis
- Methodology extension examples (TDD, Error Handling, Cross-Platform, etc.)
- Scientific Software Engineering philosophy
- Implementation roadmap

**Target audience**: Any software project with observable development process

**Status**: Framework v1.0 (2025-10-13)

### [Bootstrapped Software Engineering](bootstrapped-software-engineering.md)
Meta-methodology framework for self-evolving software development processes, derived from empirical analysis of the meta-cc project.

**Key topics**:
- Self-referential feedback loops and bootstrapping theory
- Multi-dimensional iteration architecture
- Meta-Agent bootstrapping and three-tuple output (O, Aₙ, Mₙ)
- Indirect vs. direct artifacts
- Multi-team concurrency model
- Software Engineering Darwinism and convergence theory

**Target audience**: Any software project with observable development process

**Status**: Theoretical Framework v1.0 (2025-10-13)

### [Value Space Optimization](value-space-optimization.md)
Mathematical framework for training Agents and Meta-Agents from software development history, treating development as optimization in high-dimensional value space.

**Key topics**:
- Value space mathematical model (V: S → ℝ)
- Agent as gradient (∇V), Meta-Agent as Hessian (∇²V)
- Historical data collection and value trajectory reconstruction
- Agent training pipeline (clustering, supervised learning)
- Error value extraction (positive value from failures)
- Meta-Agent training with reinforcement learning (Q-learning)
- Transfer learning and continual learning
- Case study: meta-cc development (277 commits analyzed)

**Target audience**: Projects with comprehensive development history (git + sessions + metrics)

**Status**: Theoretical Framework v1.0 (2025-10-14)

---

## Future Guides

The following methodology guides are planned for future development (using OCA Framework):

- **TDD Methodology**: Test-driven development practices and patterns (partial data collected)
- **Error Handling**: Comprehensive error handling strategies (partial data collected)
- **Cross-Platform Development**: Platform compatibility principles (partial data collected)
- **Version Management**: Semantic versioning and release workflows (partial data collected)
- **Code Review Methodology**: Evidence-based code review practices
- **Performance Optimization**: Benchmarking and optimization strategies
- **And more**: Additional software engineering methodologies

---

## Usage

These methodology documents are designed to be:

1. **Universal**: Apply to any programming language or project type
2. **Project-independent**: No meta-cc-specific content
3. **Operational**: Provide concrete checklists and decision matrices
4. **Reusable**: Can be referenced or adapted by other projects

## Navigation

- **For project-specific guidance**: See [DOCUMENTATION_MAP.md](../DOCUMENTATION_MAP.md)
- **For meta-cc development**: See [CLAUDE.md](../../CLAUDE.md)
- **For meta-cc architecture**: See [principles.md](../principles.md) and [plan.md](../plan.md)

---

**Last Updated**: 2025-10-14
