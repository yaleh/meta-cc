# Error Recovery Methodology

**Version**: 1.0
**Date**: 2025-10-16
**Status**: Validated through bootstrap-003 experiment
**Source**: Extracted from iterations 0-4 of meta-cc error recovery work

---

## Table of Contents

1. [Introduction](#introduction)
2. [When to Use This Methodology](#when-to-use-this-methodology)
3. [Core Principles](#core-principles)
4. [Pattern Catalog](#pattern-catalog)
   - [Pattern 1: Hierarchical Error Taxonomy Design](#pattern-1-hierarchical-error-taxonomy-design)
   - [Pattern 2: Root Cause Analysis Framework](#pattern-2-root-cause-analysis-framework)
   - [Pattern 3: Recovery Strategy Categorization](#pattern-3-recovery-strategy-categorization)
   - [Pattern 4: Prevention Mechanism Design](#pattern-4-prevention-mechanism-design)
   - [Pattern 5: Agent Specialization for Error Domains](#pattern-5-agent-specialization-for-error-domains)
   - [Pattern 6: Error Handling Pipeline Architecture](#pattern-6-error-handling-pipeline-architecture)
   - [Pattern 7: Automation Classification for Error Handling](#pattern-7-automation-classification-for-error-handling)
   - [Pattern 8: Value-Driven Convergence Assessment](#pattern-8-value-driven-convergence-assessment)
5. [Pattern Application Framework](#pattern-application-framework)
6. [Decision Trees](#decision-trees)
7. [Success Metrics](#success-metrics)
8. [Appendices](#appendices)

---

## Introduction

### Overview

This methodology provides a **systematic approach to error recovery system development**, emphasizing **completeness**, **measurability**, and **automation**. It was developed through the bootstrap-003 error recovery experiment on the meta-cc project and validated across 1,145 real-world errors.

The methodology consists of eight core patterns that address the complete error handling lifecycle:

1. **Taxonomy** (organizing errors systematically)
2. **Diagnosis** (understanding root causes)
3. **Recovery** (fixing errors effectively)
4. **Prevention** (avoiding errors proactively)
5. **Specialization** (when to create specialized capabilities)
6. **Architecture** (complete error handling pipeline)
7. **Automation** (when to automate recovery)
8. **Convergence** (when to declare system complete)

### Methodology Philosophy

**Core Beliefs**:
- Error handling is a **system**, not isolated capabilities (detection + diagnosis + recovery + prevention)
- **Taxonomy first**: Organization enables systematic diagnosis and recovery
- **Root cause focus**: Understanding "why" enables effective "how to fix"
- **Automation potential varies**: Not all errors can be recovered automatically (classify 51% preventable, 67% automatable)
- **Evidence-based**: Use real error data (1,145 errors) not hypothetical scenarios
- **Measurable progress**: Track value function V(s) to quantify improvement

**Non-Goals**:
- Zero errors (unrealistic - focus on effective handling, not elimination)
- 100% automation (manual recovery needed for logic/design errors)
- Perfect classification (80% coverage is often sufficient, 100% may not justify cost)
- Premature prevention (build detection/diagnosis/recovery first, prevention last)

### Success Stories

This methodology enabled:
- **Iteration 1**: Organized 1,145 errors into 7 categories, 25 subcategories (100% coverage)
- **Iteration 2**: Created 16 diagnostic procedures covering 79.9% of errors (54 root causes identified)
- **Iteration 3**: Developed 54 recovery strategies (67% automatic/semi-automatic)
- **Iteration 4**: Designed 8 prevention mechanisms (351 errors preventable, 30.7% of total)
- **Overall**: Improved error handling capability from V=0.34 to V=0.72 (+111.8%) in 15-16 hours

---

## When to Use This Methodology

### Ideal Scenarios

Use this methodology when:
- ✅ System experiences diverse error types (>100 unique errors)
- ✅ Error handling is currently reactive and ad-hoc
- ✅ No systematic error taxonomy exists
- ✅ Error data is available (session logs, error reports)
- ✅ Goal is production-ready error handling system
- ✅ Resources available for systematic development (15-20 hours minimum)

### Poor Fit Scenarios

Don't use this methodology when:
- ❌ System has <50 errors total (too small to justify taxonomy)
- ❌ Errors are homogeneous (e.g., only network timeouts - specialized handling better)
- ❌ No error data available (collect data first)
- ❌ Quick fix needed (use ad-hoc debugging instead)
- ❌ System is being deprecated (don't invest in error handling)

### Prerequisites

**Required**:
- Error history data (session logs, error reports with messages and context)
- Ability to query error data (SQL, log analysis, MCP tools)
- Development environment for testing
- Version control (git)

**Recommended**:
- Static analysis tools
- Test infrastructure (for validation)
- Meta-Agent framework (for systematic development)
- Existing error handling baseline (for comparison)

---

## Core Principles

### Principle 1: Taxonomy Enables Everything

**Organization is prerequisite for systematic handling**.

**Rationale**: Without taxonomy, cannot prioritize, diagnose systematically, or design targeted recovery strategies.

**Application**: Always start with taxonomy (Pattern 1). Detection enables diagnosis, diagnosis enables recovery, recovery informs prevention.

### Principle 2: Root Cause Over Symptom

**Fix causes, not symptoms**.

**Rationale**: Symptom-based fixes are brittle and incomplete. Root cause analysis creates durable solutions.

**Application**: Use Pattern 2 (Root Cause Analysis) to trace symptom → proximate → root. Fix root causes to prevent recurrence.

### Principle 3: Not All Errors Are Equal

**Prioritize by frequency × severity**.

**Rationale**: 80% of value comes from handling 20% of error types. High-frequency or high-severity errors deserve most attention.

**Application**: Calculate priority = frequency × severity. Address high-priority errors first. Accept 70-80% coverage as sufficient.

### Principle 4: Automation Potential Varies

**Classify by automation feasibility: automatic, semi-automatic, manual**.

**Rationale**: Some errors can be auto-recovered (path typos), others need human judgment (logic errors). Misclassifying creates unsafe automation.

**Application**: Use Pattern 7 to classify automation potential. Be conservative - prefer semi-automatic over automatic when uncertain.

### Principle 5: Prevention Comes Last

**Build detection → diagnosis → recovery → prevention in order**.

**Rationale**: Prevention requires understanding from diagnosis/recovery. Building prevention first is premature optimization.

**Application**: Follow pipeline order (Pattern 6). Prevention mechanisms emerge naturally from recovery insights.

### Principle 6: Completeness Over Perfection

**Complete error handling pipeline beats perfect individual components**.

**Rationale**: System needs all four dimensions (detect, diagnose, recover, prevent). Optimizing one dimension in isolation creates weak systems.

**Application**: Develop balanced system - address all four dimensions to 70% rather than one dimension to 95%.

### Principle 7: Convergence Is Detectable

**Use value function V(s) to measure progress and detect convergence**.

**Rationale**: Subjective "good enough" is unreliable. Quantitative metrics enable objective convergence assessment.

**Application**: Track V(s) = 0.4·V_detection + 0.3·V_diagnosis + 0.2·V_recovery + 0.1·V_prevention. Converge when V ≥ 0.70-0.75.

---

## Pattern Catalog

---

## Pattern 1: Hierarchical Error Taxonomy Design

### Context

**When to use**: When you have diverse error types (>100 unique) and need systematic organization.

**When not to use**: For homogeneous errors (<50 types), or when errors are already well-categorized.

**Frequency**: Once at beginning of error handling system development, then maintained incrementally.

### Problem

**Unorganized errors lead to**:
- **No prioritization basis**: Cannot tell which errors are critical vs. minor
- **Duplicate handling efforts**: Same error fixed multiple times without knowledge capture
- **Inconsistent severity assessment**: "Critical" means different things to different people
- **Poor diagnosis**: Cannot identify patterns or common root causes
- **Ad-hoc recovery**: Each error treated as unique, no systematic approach

**Example chaos** (Bootstrap-003 baseline):
- 1,145 total errors
- 654 unique error types
- No categorization system
- Generic "Error" messages providing no information (50 cases, 4.37%)
- Cannot identify highest-priority errors objectively

### Solution

**Design hierarchical taxonomy with categories, subcategories, and severity levels using MECE principle (Mutually Exclusive, Collectively Exhaustive)**.

**Core Idea**: Organize errors into 5-10 major categories, each with 2-5 subcategories. Classify all errors by category + severity. Use taxonomy as foundation for diagnosis, recovery, and prevention.

### Detailed Steps

1. **Collect error data**: Query error history (session logs, error reports)
   ```bash
   # Example: MCP query for meta-cc project
   mcp__meta-cc__query_tools(status="error", scope="project")
   # Result: 1,145 error records
   ```

2. **Analyze error distribution**: Identify major error types
   - Group by tool/component (e.g., Bash errors, file errors, MCP errors)
   - Count frequencies (e.g., Bash: 586 errors, 51%)
   - Identify patterns (e.g., "file not found" appears 101 times)

3. **Define major categories (5-10)**: High-level groupings
   - **Guideline**: Categories should be mutually exclusive (no overlap)
   - **Example categories**:
     1. File Operations Errors
     2. Command Execution Errors
     3. MCP Integration Errors
     4. User Interruption Errors
     5. Resource Limit Errors
     6. Tool Coordination Errors
     7. Other/Uncategorized

4. **Define subcategories (2-5 per category)**: Specific error types
   - **Example: File Operations → Subcategories**:
     - file_not_found (101 errors)
     - read_before_write_violation (57 errors)
     - string_not_found (28 errors)
     - token_limit_exceeded (7 errors)

5. **Define severity levels (3-5)**: Impact classification
   - **Critical**: Blocks all work, system-breaking, data loss risk (< 1% of errors)
   - **High**: Blocks current task, requires immediate fix (40-45% of errors)
   - **Medium**: Degrades experience, workaround available (40-45% of errors)
   - **Low**: Minor inconvenience, rare, minimal impact (10-15% of errors)

6. **Create classification rules (10-20)**: Pattern matching for automation
   - **Rule format**: IF error_message matches PATTERN THEN category = X, severity = Y
   - **Example rule**:
     ```yaml
     rule: file_not_found_classification
     pattern: "File does not exist|No such file|cannot find file"
     category: file_operations
     subcategory: file_not_found
     severity: high  # Blocks operation
     ```

7. **Classify all errors**: Apply rules to error dataset
   - Manual classification for first 50-100 errors (establish patterns)
   - Automated classification for remaining errors (using rules)
   - Handle edge cases manually
   - Target: 100% coverage (every error assigned category + severity)

8. **Validate taxonomy**:
   - **MECE check**: No errors fit multiple categories (mutually exclusive)
   - **Completeness check**: All errors fit into categories (collectively exhaustive)
   - **Balance check**: No category has >60% of errors (split large categories)
   - **Severity distribution check**: Critical <5%, High 30-50%, Medium 30-50%, Low 10-20%

9. **Document taxonomy**:
   - Create taxonomy.yaml or taxonomy.md
   - Include: category definitions, subcategory definitions, severity criteria, classification rules
   - Provide examples for each category

10. **Calculate metrics**:
    ```yaml
    total_errors: 1145
    categories: 7
    subcategories: 25
    classification_rules: 17
    coverage: 100% (1145/1145 errors classified)

    by_severity:
      critical: 5 (0.4%)
      high: 491 (42.9%)
      medium: 506 (44.2%)
      low: 143 (12.5%)
    ```

### Example: Bootstrap-003 Iteration 1

#### Initial State (Iteration 0)
- **Errors**: 1,145 errors, 654 unique types
- **Organization**: None (unorganized error logs)
- **Categorization**: 5 rough categories (from preliminary analysis)
- **Severity**: No systematic assessment

#### Taxonomy Design Process

**Step 1**: Collect error data
```bash
$ mcp__meta-cc__query_tools(status="error", scope="project")
# Retrieved: 1,145 error records with messages, timestamps, tools
```

**Step 2**: Analyze distribution
```
Top error-prone tools:
- Bash: 586 errors (51.2%)
- Read: 184 errors (16.1%)
- Edit: 101 errors (8.8%)
- MCP tools: 137 errors (12.0%)

Top error patterns:
- File does not exist: 101 occurrences
- Read-before-write violation: 57 occurrences
- jq syntax error: 47 occurrences
```

**Step 3**: Define categories (7 major)
1. **File Operations Errors** (16.8%, 192 errors)
2. **Command Execution Errors** (51.2%, 586 errors)
3. **MCP Integration Errors** (12.0%, 137 errors)
4. **User Interruption Errors** (3.1%, 35 errors)
5. **Resource Limit Errors** (1.6%, 18 errors)
6. **Tool Coordination Errors** (2.3%, 26 errors)
7. **Other/Uncategorized Errors** (13.2%, 151 errors)

**Step 4**: Define subcategories (25 total)

**File Operations** (4 subcategories):
- file_not_found (101 errors, high severity)
- read_before_write_violation (57 errors, high severity)
- string_not_found (28 errors, medium severity)
- token_limit_exceeded (7 errors, medium severity)

**Command Execution** (6 subcategories):
- command_not_found (110 errors, medium severity)
- syntax_error (45 errors, medium severity)
- build_failure (180 errors, high severity)
- test_failure (150 errors, medium severity)
- permission_denied (12 errors, high severity)
- generic_execution_error (89 errors, low severity)

**MCP Integration** (6 subcategories):
- jq_syntax_error (59 errors, high severity)
- parse_error (35 errors, medium severity)
- mcp_tool_execution_failed (21 errors, high severity)
- session_not_found (8 errors, medium severity)
- mcp_connection_error (5 errors, critical severity)
- capability_loading_error (11 errors, high severity)

**Step 5**: Define severity levels (4 levels)

| Severity | Impact | Response Time | Bootstrap-003 Count | % |
|----------|--------|---------------|---------------------|---|
| Critical | Blocks all work, system-breaking | < 1 hour | 5 | 0.4% |
| High | Blocks current task | < 4 hours | 491 | 42.9% |
| Medium | Degrades experience | < 1 day | 506 | 44.2% |
| Low | Minor inconvenience | < 1 week | 143 | 12.5% |

**Step 6**: Create classification rules (17 rules)

```yaml
rules:
  - rule_id: R01_file_not_found
    pattern: "File does not exist|No such file|FileNotFoundError"
    category: file_operations
    subcategory: file_not_found
    severity: high
    examples: ["Read error: File does not exist at /path/to/file"]

  - rule_id: R02_read_before_write
    pattern: "File not read before|Must read file before"
    category: file_operations
    subcategory: read_before_write_violation
    severity: high
    examples: ["Write error: File not read before write"]

  - rule_id: R03_jq_syntax
    pattern: "jq: error|jq: parse error|invalid jq filter"
    category: mcp_integration
    subcategory: jq_syntax_error
    severity: high
    examples: ["jq: parse error at line 1, column 5"]

  # ... 14 more rules
```

**Step 7**: Classify all errors (100% coverage achieved)

```yaml
classification_results:
  total_errors: 1145
  classified: 1145
  coverage: 100.0%

  by_category:
    command_execution: 586 (51.2%)
    file_operations: 192 (16.8%)
    other_errors: 151 (13.2%)
    mcp_integration: 137 (12.0%)
    user_interruption: 35 (3.1%)
    tool_coordination: 26 (2.3%)
    resource_limits: 18 (1.6%)
```

#### Results

- **V_detection improvement**: 0.50 → 0.80 (+0.30, +60%)
- **Taxonomy structure**: 7 categories, 25 subcategories, 100% coverage
- **Classification rules**: 17 automated rules
- **Severity assessment**: 4 levels, clear criteria
- **Time spent**: ~3 hours
- **Reusability**: High (taxonomy structure transferable to other projects)

### Before/After Comparison

#### Before (No Taxonomy)
```
Error 1: "Error"
  - No category
  - No severity
  - Cannot prioritize
  - Ad-hoc handling

Error 2: "File does not exist at /path/to/file.txt"
  - No category
  - No severity
  - Treated as unique
  - Manual investigation each time

Error 3: "jq: parse error at line 1, column 5"
  - No category
  - No severity
  - No pattern recognition
```

#### After (With Taxonomy)
```
Error 1: "Error"
  - Category: other_errors
  - Subcategory: miscellaneous
  - Severity: varies
  - Priority: LOW (insufficient information)

Error 2: "File does not exist at /path/to/file.txt"
  - Category: file_operations
  - Subcategory: file_not_found
  - Severity: high (blocks operation)
  - Priority: HIGH (101 occurrences, 8.8%)
  - Root causes: typo (40%), deleted (25%), wrong directory (20%), never existed (15%)

Error 3: "jq: parse error at line 1, column 5"
  - Category: mcp_integration
  - Subcategory: jq_syntax_error
  - Severity: high (breaks MCP queries)
  - Priority: HIGH (59 occurrences, 5.2%)
  - Root causes: invalid syntax (70%), Python syntax in jq (20%), missing dot (10%)
```

### Verification Checklist

- [ ] Collected comprehensive error data (>100 errors)
- [ ] Defined 5-10 major categories (mutually exclusive)
- [ ] Defined 2-5 subcategories per category
- [ ] Defined 3-5 severity levels with clear criteria
- [ ] Created 10-20 classification rules with patterns
- [ ] Classified all errors (target: 100% coverage)
- [ ] Validated MECE (no overlaps, no gaps)
- [ ] Validated severity distribution (Critical <5%, High 30-50%, Medium 30-50%, Low 10-20%)
- [ ] Documented taxonomy (categories, subcategories, severity, rules, examples)
- [ ] Calculated coverage metrics

### Pitfalls and How to Avoid

**Pitfall 1**: Too many categories (>15)
- ❌ Wrong: 20 fine-grained categories (hard to remember, poor organization)
- ✅ Right: 5-10 major categories (easy to navigate)

**Pitfall 2**: Categories overlap (violates MECE)
- ❌ Wrong: "File Errors" and "Read Errors" (overlap for "read file not found")
- ✅ Right: "File Operations" (includes all file-related errors regardless of operation)

**Pitfall 3**: Severity based on emotion, not impact
- ❌ Wrong: "This error annoys me" → Critical
- ✅ Right: "This error blocks all work for all users" → Critical

**Pitfall 4**: No classification rules (all manual)
- ❌ Wrong: Manually classify each of 1,145 errors every time
- ✅ Right: Create 17 rules, automatically classify 80%+ of errors

**Pitfall 5**: Premature subcategories
- ❌ Wrong: Define 50 subcategories upfront based on intuition
- ✅ Right: Start with 15-20 subcategories, add more as needed based on actual errors

### Variations

**Variation 1: Lightweight Taxonomy (Small Systems)**
- Use when: <200 errors total
- Structure: 3-5 categories, 1-3 subcategories, 2-3 severity levels
- Example: {file_errors, command_errors, other} × {high, medium, low}

**Variation 2: Multi-Dimensional Taxonomy**
- Use when: Errors need multiple classification dimensions
- Structure: Category + Source + Severity
- Example: "file_not_found (file_operations, user_input, high)"

**Variation 3: Domain-Specific Taxonomy**
- Use when: Errors are domain-specific (e.g., database, network, security)
- Structure: Domain categories (database_errors, network_errors, security_errors)
- Example: database_errors → {connection_failed, query_timeout, constraint_violation}

**Variation 4: Temporal Taxonomy**
- Use when: Errors have temporal patterns (startup vs. runtime vs. shutdown)
- Structure: Phase + Category
- Example: {startup_errors, runtime_errors, shutdown_errors} × {file, command, network}

### Reusability

**Language Agnostic**: ✅ Concept applies universally
**Domain Agnostic**: ✅ Error classification needed in any software project

**Transferability Assessment**:

**Go Projects**:
- Error types: panic, nil pointer, type errors, goroutine leaks
- Categories: runtime_errors, concurrency_errors, type_errors, resource_errors
- **Transferability**: HIGH (same hierarchical approach, different error types)

**Python Projects**:
- Error types: Exception subclasses, traceback, import errors
- Categories: exception_errors, import_errors, type_errors, runtime_errors
- **Transferability**: HIGH (Exception hierarchy already hierarchical)

**JavaScript/TypeScript Projects**:
- Error types: TypeError, ReferenceError, SyntaxError, Promise rejections
- Categories: type_errors, reference_errors, syntax_errors, async_errors
- **Transferability**: HIGH (Error types map naturally to categories)

**Java Projects**:
- Error types: Exception subclasses, NullPointerException, checked vs. unchecked
- Categories: checked_exceptions, unchecked_exceptions, runtime_errors
- **Transferability**: HIGH (Exception hierarchy built-in)

**Web Applications**:
- Error types: HTTP errors (4xx, 5xx), client errors, server errors
- Categories: client_errors, server_errors, network_errors, auth_errors
- **Transferability**: HIGH (HTTP status codes provide natural categories)

**Database Systems**:
- Error types: Connection errors, query errors, constraint violations, deadlocks
- Categories: connection_errors, query_errors, integrity_errors, concurrency_errors
- **Transferability**: HIGH (database error codes map to categories)

**System Administration**:
- Error types: Disk full, process crashes, permission denied, network unreachable
- Categories: resource_errors, process_errors, permission_errors, network_errors
- **Transferability**: HIGH (system error categories universal)

### Key Takeaways

- ✅ **Start with taxonomy** - organization enables everything else
- ✅ **Use MECE principle** - categories mutually exclusive, collectively exhaustive
- ✅ **Classify by frequency × severity** - prioritize high-impact errors
- ✅ **Create classification rules** - automate 80%+ of categorization
- ✅ **100% coverage target** - every error gets category + severity
- ✅ **Validate severity distribution** - avoid "everything is critical" trap
- ✅ **Document thoroughly** - taxonomy is foundation, must be clear

**One-sentence summary**: Design hierarchical error taxonomy with 5-10 categories, 2-5 subcategories each, 3-5 severity levels, and 10-20 classification rules to achieve 100% error coverage.

---

## Pattern 2: Root Cause Analysis Framework

### Context

**When to use**: After taxonomy is established (Pattern 1) and you need systematic diagnosis procedures.

**When not to use**: Before taxonomy exists, or when errors are simple (single obvious cause).

**Frequency**: Once to establish framework, then applied to each error category.

### Problem

**Ad-hoc diagnosis leads to**:
- **Symptom-based fixes**: Treat symptoms, not root causes (error recurs)
- **Inconsistent investigation**: Different people investigate differently
- **Missing root causes**: Stop at proximate cause, miss deeper root
- **No systematic approach**: Each error diagnosed from scratch
- **Cannot measure accuracy**: No way to validate diagnosis correctness

**Example problem** (Bootstrap-003 pre-diagnosis):
- 654 unique error types identified
- No systematic diagnostic procedures
- Root cause identification was manual, ad-hoc, inconsistent
- Generic "Error" messages provided no diagnostic information (50 cases)
- Cannot tell if "file not found" is due to typo vs. deletion vs. wrong directory

### Solution

**Establish systematic root cause analysis framework with three methodologies (5 Whys, Fault Tree Analysis, Causal Chain Analysis) and create diagnostic procedures for each high-priority error category**.

**Core Idea**: Use proven methodologies to trace from symptom to root cause. Create reusable diagnostic procedures mapping error types → likely root causes → verification methods. Document as decision trees for systematic application.

### Detailed Steps

1. **Select root cause analysis methodologies (3-5)**:

   **Methodology 1: 5 Whys**
   - **Purpose**: Iteratively ask "why" to trace causation
   - **Depth**: Typically 3-5 levels deep
   - **Example**:
     - Symptom: File not found
     - Why 1: Path incorrect
     - Why 2: Typo in path
     - Why 3: Copied from documentation with typo
     - Root: Documentation has incorrect path example

   **Methodology 2: Fault Tree Analysis**
   - **Purpose**: Work backward from error, identify immediate causes with AND/OR logic
   - **Structure**: Tree diagram showing causal relationships
   - **Example**:
     ```
     File Not Found (top event)
       ├─ OR: Path Issues
       │   ├─ AND: Typo in path
       │   ├─ AND: Wrong working directory
       │   └─ AND: Relative path used incorrectly
       └─ OR: File Issues
           ├─ AND: File deleted
           ├─ AND: File moved
           └─ AND: File never created
     ```

   **Methodology 3: Causal Chain Analysis**
   - **Purpose**: Linear chain from root → proximate → symptom
   - **Structure**: Root cause → contributing factors → proximate cause → symptom
   - **Example**:
     ```
     Root: Documentation has incorrect example path
       → Contributing: Developer copied example without verification
       → Proximate: Path has typo (extra slash)
       → Symptom: File not found error
     ```

2. **Prioritize error categories for diagnosis**: Select high-frequency or high-severity
   - **Criteria**: Priority = frequency × severity
   - **Example**: Bootstrap-003 addressed 3 categories (79.9% of all errors):
     - command_execution (586 errors, 51.2%)
     - file_operations (192 errors, 16.8%)
     - mcp_integration (137 errors, 12.0%)

3. **Create diagnostic procedure template**: Standard structure for all procedures
   ```yaml
   diagnostic_procedure:
     subcategory: file_not_found
     category: file_operations

     initial_assessment:
       - Check file path syntax
       - Verify working directory
       - Review recent tool sequence
       - Extract error context

     investigation_steps:
       - step_1: "Verify file path correctness"
       - step_2: "Check if file existed previously"
       - step_3: "Search for similar paths (typo detection)"
       - step_4: "Review file lifecycle (creation/deletion)"

     root_causes:
       - cause_1:
           name: "Typo in file path"
           probability: 40%
           indicators: ["Path differs by 1-3 characters from existing file"]
           verification: ["Fuzzy match (Levenshtein distance < 3)", "Suggest corrected path"]
           causal_chain: "Incorrect path → File system lookup fails → File not found error"

       - cause_2:
           name: "File deleted or moved"
           probability: 25%
           indicators: ["File accessed earlier in session", "Recent deletion operation in tool sequence"]
           verification: ["Check tool sequence for rm/mv commands", "Search file in previous states"]
           causal_chain: "File deleted → File no longer exists → File system lookup fails → File not found"

       - cause_3:
           name: "Wrong working directory"
           probability: 20%
           indicators: ["Relative path used", "cd command before error"]
           verification: ["Check current working directory", "Convert to absolute path and test"]
           causal_chain: "Relative path + wrong cwd → Resolved path incorrect → File not found"

       - cause_4:
           name: "File never existed"
           probability: 15%
           indicators: ["No previous file reference in conversation", "Parent directory doesn't exist"]
           verification: ["Search entire conversation history", "Check parent directory existence"]
           causal_chain: "File creation step skipped → File never created → File not found"

     decision_tree:
       - condition: "Path has typo (fuzzy match found)"
         root_cause: "Typo in file path"
         confidence: high
         next_action: "Suggest corrected path"

       - condition: "File accessed earlier AND tool sequence shows deletion"
         root_cause: "File deleted in workflow"
         confidence: high
         next_action: "Identify deletion point, suggest recovery"

       - condition: "Relative path used AND cd command before error"
         root_cause: "Working directory mismatch"
         confidence: high
         next_action: "Convert to absolute path"

       - condition: "No previous file reference found"
         root_cause: "File never created"
         confidence: medium
         next_action: "Verify file creation step"
   ```

4. **For each high-priority subcategory, create diagnostic procedure**:
   - Apply template
   - Identify 3-5 likely root causes based on error analysis
   - Assign probability estimates (sum to ~100%, realistic based on data)
   - Define indicators (how to recognize each root cause)
   - Define verification methods (how to confirm root cause)
   - Create causal chains (root → proximate → symptom)
   - Build decision tree (if-then diagnostic logic)

5. **Define diagnostic tools (optional)**: Automation helpers
   - **path_validator**: Validate file paths, suggest corrections
   - **file_lifecycle_tracker**: Track file creation/deletion through conversation
   - **command_sequence_analyzer**: Analyze tool execution sequence
   - **fuzzy_string_matcher**: Find similar strings for typo detection

6. **Validate diagnostic procedures**: Test on sample errors
   - Select 10-20 sample errors per subcategory
   - Apply diagnostic procedure manually
   - Verify root cause identification accuracy
   - Refine decision trees based on results

7. **Calculate diagnostic coverage**:
   ```yaml
   coverage_metrics:
     total_errors: 1145
     diagnostic_procedures_created: 16
     errors_covered_by_procedures: 915
     coverage_percentage: 79.9%

     by_category:
       file_operations: 96.9% (186/192 errors)
       command_execution: 84.8% (497/586 errors)
       mcp_integration: 93.4% (128/137 errors)

     root_causes_identified: 54
     average_root_causes_per_procedure: 3.4
     decision_trees_created: 16
   ```

8. **Document framework**: Create diagnostic_framework.md
   - Include: 3 methodologies, diagnostic procedure template, validation methods
   - Provide: Examples for each methodology
   - Define: Quality metrics (accuracy, coverage, completeness)

### Example: Bootstrap-003 Iteration 2

#### Diagnostic Procedures Created (16 total)

**File Operations** (4 procedures):

1. **file_not_found** (101 errors):
   - Root causes: Typo (40%), deleted (25%), wrong directory (20%), never existed (15%)
   - Decision tree: 4 branches
   - Tools: path_validator, file_lifecycle_tracker

2. **read_before_write_violation** (57 errors):
   - Root causes: Protocol forgotten (70%), complex workflow (20%), async operation (10%)
   - Decision tree: 3 branches
   - Tools: protocol_validator, workflow_analyzer

3. **string_not_found** (28 errors):
   - Root causes: Incorrect old_string (50%), whitespace mismatch (30%), file changed (20%)
   - Decision tree: 3 branches
   - Tools: string_matcher, diff_analyzer

4. **token_limit_exceeded** (7 errors):
   - Root causes: Large file (60%), large output (30%), pagination disabled (10%)
   - Decision tree: 3 branches
   - Tools: token_counter, pagination_manager

**Command Execution** (5 procedures):

5. **command_not_found** (110 errors):
   - Root causes: Not installed (60%), typo (30%), wrong path (10%)
   - Decision tree: 3 branches
   - Tools: command_validator, package_finder

6. **syntax_error** (45 errors):
   - Root causes: Quote mismatch (40%), bracket mismatch (30%), invalid operator (30%)
   - Decision tree: 3 branches
   - Tools: bash_syntax_checker, delimiter_balancer

7. **build_failure** (180 errors):
   - Root causes: Dependency missing (40%), compilation error (35%), config invalid (25%)
   - Decision tree: 3 branches
   - Tools: dependency_checker, build_log_analyzer

**MCP Integration** (5 procedures):

12. **jq_syntax_error** (59 errors):
    - Root causes: Invalid syntax (70%), Python syntax in jq (20%), missing dot accessor (10%)
    - Decision tree: 3 branches
    - Tools: jq_validator, jq_syntax_fixer

13. **mcp_tool_execution_failed** (21 errors):
    - Root causes: Invalid parameters (50%), session issue (30%), server error (20%)
    - Decision tree: 3 branches
    - Tools: parameter_validator, session_checker

**... 4 more procedures**

#### Results

- **Diagnostic procedures**: 16 complete procedures
- **Root causes identified**: 54 total (3.4 average per procedure)
- **Error coverage**: 79.9% (915/1145 errors)
- **Decision trees**: 16 if-then diagnostic logic trees
- **Diagnostic tools specified**: 7 tools (not implemented, specification complete)
- **V_diagnosis improvement**: 0.35 → 0.70 (+0.35, +100%)
- **Time spent**: ~4 hours

### Before/After Comparison

#### Before (No Diagnostic Framework)

```
Error: "File does not exist at /path/to/file.txt"

Investigation process:
- Developer manually checks path
- Developer guesses: "Maybe typo?"
- Developer tries a few variations
- Developer gives up or tries different approach
- No systematic method
- Cannot learn from previous similar errors

Outcome: Time wasted, inconsistent diagnosis, no knowledge capture
```

#### After (With Diagnostic Framework)

```
Error: "File does not exist at /path/to/file.txt"

Investigation process:
1. Apply diagnostic procedure for file_not_found
2. Initial assessment: Check path syntax, verify cwd, review tool sequence
3. Decision tree evaluation:
   - Fuzzy match? → Found "path/to/flie.txt" (Levenshtein distance = 1)
   - ROOT CAUSE: Typo in file path (confidence: high)
4. Next action: Suggest corrected path "path/to/file.txt"
5. Verification: File exists at corrected path? Yes
6. Recovery: Use corrected path

Outcome: Systematic diagnosis in <30 seconds, high confidence, knowledge captured
```

### Verification Checklist

- [ ] Selected 3-5 root cause analysis methodologies
- [ ] Prioritized error categories by frequency × severity
- [ ] Created diagnostic procedure template
- [ ] Created 10-20 diagnostic procedures (target: 70-80% error coverage)
- [ ] Identified 3-5 root causes per procedure with probability estimates
- [ ] Defined indicators for each root cause
- [ ] Defined verification methods for each root cause
- [ ] Created causal chains (root → proximate → symptom)
- [ ] Built decision trees (if-then diagnostic logic)
- [ ] Specified diagnostic tools (optional)
- [ ] Validated procedures on sample errors
- [ ] Calculated coverage metrics
- [ ] Documented framework

### Pitfalls and How to Avoid

**Pitfall 1**: Stopping at proximate cause, not finding root
- ❌ Wrong: "File not found" → "Path incorrect" (proximate, not root)
- ✅ Right: "File not found" → "Path incorrect" → "Typo" → "Copied from bad documentation" (root)

**Pitfall 2**: Subjective probability estimates
- ❌ Wrong: "Typo feels like 90%" (no data)
- ✅ Right: "Typo observed in 40 of 100 file_not_found cases → 40% probability" (data-driven)

**Pitfall 3**: No verification methods
- ❌ Wrong: "Root cause is typo" (asserted but not verified)
- ✅ Right: "Root cause is typo (verified by fuzzy matching, Levenshtein distance < 3)"

**Pitfall 4**: Decision trees too complex
- ❌ Wrong: 20-branch decision tree with nested conditions
- ✅ Right: 3-5 branch decision tree with clear if-then logic

**Pitfall 5**: Generic root causes
- ❌ Wrong: "User error" (not actionable)
- ✅ Right: "Typo in file path" (specific, actionable)

### Variations

**Variation 1: Lightweight Diagnosis (Simple Errors)**
- Use when: Errors have 1-2 obvious causes
- Structure: Single methodology (5 Whys), simplified decision tree
- Example: "Command not found" → "Not installed" (80%) or "Typo" (20%)

**Variation 2: Data-Driven Probabilities**
- Use when: Large error dataset available
- Structure: Calculate probabilities from actual error data
- Example: Analyze 100 file_not_found errors → 42 typos → 42% probability

**Variation 3: Automated Diagnosis**
- Use when: Diagnostic tools can be implemented
- Structure: Automated decision tree execution with tool invocations
- Example: Run path_validator → fuzzy_match → suggest_correction (fully automated)

**Variation 4: Collaborative Diagnosis**
- Use when: Errors require domain expertise
- Structure: Diagnostic procedure with expert consultation steps
- Example: Build failure → Run build_log_analyzer → Consult build expert if ambiguous

### Reusability

**Language Agnostic**: ✅ Root cause analysis methodologies are universal
**Domain Agnostic**: ⚠️ Diagnostic procedures are domain-specific, but framework is universal

**Transferability Assessment**:

**Software Errors (Any Language)**:
- Framework: 5 Whys, Fault Tree, Causal Chain apply universally
- Procedures: Need adaptation (different error types in Go vs. Python)
- Decision trees: Structure transfers, content needs adaptation
- **Transferability**: HIGH (framework), MEDIUM (procedures)

**System Administration**:
- Framework: All 3 methodologies applicable
- Example: "Service failed to start" → Why? → Port occupied → Why? → Previous instance didn't terminate
- **Transferability**: HIGH

**Database Operations**:
- Framework: All 3 methodologies applicable
- Example: "Deadlock detected" → Fault tree of lock acquisition sequences
- **Transferability**: HIGH

**Network Operations**:
- Framework: All 3 methodologies applicable
- Example: "Connection timeout" → Causal chain of network path issues
- **Transferability**: HIGH

**Quality Assurance**:
- Framework: All 3 methodologies applicable (5 Whys classic in QA)
- Example: "Test failed" → Why? → Assertion incorrect → Why? → Requirements changed
- **Transferability**: HIGH

**Business Processes**:
- Framework: 5 Whys originated in manufacturing (Toyota Production System)
- Example: "Customer complaint" → Why? → Delivery late → Why? → Warehouse understaffed
- **Transferability**: HIGH

### Key Takeaways

- ✅ **Use multiple methodologies** - 5 Whys, Fault Tree, Causal Chain complement each other
- ✅ **Focus on root, not symptom** - trace from symptom to root cause systematically
- ✅ **Data-driven probabilities** - estimate root cause likelihood from actual error data
- ✅ **Define verification methods** - confirm root cause with objective tests
- ✅ **Create decision trees** - enable systematic diagnostic application
- ✅ **Target 70-80% coverage** - diminishing returns beyond 80%
- ✅ **Document thoroughly** - diagnostic procedures are reusable knowledge assets

**One-sentence summary**: Establish systematic root cause analysis framework with 3 methodologies (5 Whys, Fault Tree, Causal Chain), create diagnostic procedures for high-priority errors with decision trees, and identify 3-5 root causes per procedure with probability estimates and verification methods.

---

## Pattern 3: Recovery Strategy Categorization

### Context

**When to use**: After diagnostic procedures exist (Pattern 2) and you need systematic recovery strategies.

**When not to use**: Before diagnosis is established, or when errors cannot be recovered (system-level failures).

**Frequency**: Once to establish framework, then applied to each root cause.

### Problem

**Ad-hoc recovery leads to**:
- **Inconsistent recovery quality**: Some errors recovered well, others poorly
- **No automation strategy**: Cannot determine which recoveries can be automated
- **Missing validation**: Cannot verify recovery succeeded
- **No rollback plans**: Recovery failures leave system in bad state
- **Duplicate recovery efforts**: Same recovery implemented multiple times

**Example problem** (Bootstrap-003 pre-recovery):
- 54 root causes identified (from diagnostic procedures)
- No documented recovery procedures for any root cause
- No automation strategy (which recoveries can be automatic vs. manual)
- Manual fixes only, ad-hoc and context-dependent
- No knowledge capture or learning from successful recoveries

### Solution

**Create systematic recovery procedures for each root cause, classify automation potential (automatic/semi-automatic/manual), and establish validation framework with success criteria and rollback procedures**.

**Core Idea**: For each root cause identified in diagnostic procedures, design recovery strategy with 7 components: metadata, prerequisites, recovery steps, validation checks, success criteria, rollback procedure, common pitfalls. Classify automation potential realistically.

### Detailed Steps

1. **Create recovery procedure template**: Standard structure with 7 components
   ```yaml
   recovery_procedure:
     metadata:
       subcategory_id: file_not_found
       root_cause: "Typo in file path"
       automation_classification: automatic
       priority: high

     prerequisites:
       - "Fuzzy match found (corrected path identified)"
       - "Corrected path exists in file system"

     recovery_steps:
       - step_1: "Identify corrected path using fuzzy matching (Levenshtein distance < 3)"
       - step_2: "Verify corrected path exists in file system"
       - step_3: "Replace incorrect path with corrected path in original command"
       - step_4: "Retry operation with corrected path"

     validation_checks:
       - "File exists at corrected path (file system check)"
       - "Operation succeeds with corrected path (no file_not_found error)"
       - "Corrected path unambiguous (only one match)"

     success_criteria:
       - "Original operation completes successfully"
       - "No file_not_found error occurs"
       - "Result matches expected output"

     rollback_procedure:
       - condition: "Multiple fuzzy matches found (ambiguous)"
         action: "Report ambiguity to user, request manual selection"
       - condition: "Corrected path doesn't exist"
         action: "Fall back to manual path entry"
       - condition: "Operation fails with corrected path"
         action: "Report unexpected error, escalate to manual investigation"

     common_pitfalls:
       - warning: "Don't auto-correct if Levenshtein distance > 3 (too different, likely not a typo)"
       - warning: "Don't auto-correct if multiple matches (ambiguous, require user confirmation)"
       - edge_case: "File exists but permissions deny access (different error, not typo)"
       - risk_mitigation: "Always verify corrected path exists before retrying"
   ```

2. **Define automation classification criteria**: Categorize recovery strategies

   **Automatic (20-30% of strategies)**:
   - Deterministic solution (always same fix)
   - No user input required
   - Safe to execute automatically
   - Low risk of side effects
   - Fast execution (<1 second)
   - Examples: path typo correction, protocol enforcement, pagination

   **Semi-automatic (40-50% of strategies)**:
   - Solution requires user confirmation
   - Multiple valid options exist
   - Moderate risk requires verification
   - May require system changes (install package, restart server)
   - Examples: dependency installation, server restart, permission fixing

   **Manual (20-40% of strategies)**:
   - Requires human judgment
   - Logic or design errors
   - High complexity
   - Context-dependent solution
   - Examples: code regression fixes, test logic errors, design refactoring

3. **For each root cause (from diagnostic procedures), create recovery procedure**:
   - Map root cause → recovery strategy
   - Apply recovery procedure template (7 components)
   - Classify automation potential (automatic/semi-automatic/manual)
   - Define prerequisites (what must be true before recovery)
   - List recovery steps (numbered, ordered actions)
   - Define validation checks (how to verify recovery succeeded)
   - Define success criteria (when is recovery complete)
   - Create rollback procedure (what to do if recovery fails)
   - Document common pitfalls (warnings, edge cases, risk mitigations)

4. **Create recovery automation tools (optional)**: Implementation specifications
   - **path_corrector**: Automatic path typo correction (fuzzy matching + file system validation)
   - **protocol_enforcer**: Automatic Read insertion before Write/Edit
   - **dependency_installer**: Guided dependency installation (semi-automatic, user confirmation)
   - **jq_syntax_fixer**: jq filter syntax validation and correction (semi-automatic)

5. **Establish recovery validation framework**: Systematic validation methodology

   **Validation Principles**:
   - Every recovery must have objective validation checks
   - Success criteria must be measurable
   - Rollback procedures must be defined for failure cases
   - Validation should test actual success, not just absence of error

   **Validation Check Types**:
   - **Existence checks**: Resource exists (file, command, dependency)
   - **Operation success checks**: Operation completes without errors
   - **State consistency checks**: System state is correct and consistent
   - **Behavioral checks**: Correct behavior verified (original operation succeeds)

6. **Calculate recovery coverage and automation metrics**:
   ```yaml
   recovery_coverage:
     diagnostic_procedures: 16
     recovery_procedures: 16
     coverage_percentage: 100.0%  # 100% of diagnostic procedures have recovery

     root_causes: 54
     root_causes_with_recovery: 54
     root_cause_coverage: 100.0%

   automation_classification:
     total_strategies: 54
     automatic: 11 (20.4%)
     semi_automatic: 25 (46.3%)
     manual: 18 (33.3%)

     by_category:
       file_operations:
         automatic: 5 (50.0%)      # Highest automation potential
         semi_automatic: 3 (30.0%)
         manual: 2 (20.0%)

       command_execution:
         automatic: 2 (8.7%)       # Lower automation (complex errors)
         semi_automatic: 10 (43.5%)
         manual: 11 (47.8%)

       mcp_integration:
         automatic: 4 (19.0%)      # Medium automation
         semi_automatic: 12 (57.1%)
         manual: 5 (23.8%)

   success_rate_estimates:
     automatic_strategies: 85-95% (average 90%)
     semi_automatic_strategies: 60-85% (average 73%)
     manual_strategies: 50-80% (average 67%)
     overall_weighted_average: 76%
   ```

7. **Validate recovery procedures**: Test on sample errors
   - Select 5-10 sample errors per subcategory
   - Apply recovery procedure
   - Measure success rate (does recovery succeed?)
   - Refine steps based on failures
   - Adjust automation classification if needed (e.g., automatic → semi-automatic if success rate <85%)

8. **Document recovery framework**: Create recovery_framework.md
   - Include: Recovery procedure template, automation classification criteria, validation framework
   - Provide: Examples for each automation classification
   - Define: Success metrics (success rate, automation percentage, coverage)

### Example: Bootstrap-003 Iteration 3

#### Recovery Strategies Created (54 total, mapping all root causes)

**File Operations** (10 strategies):

**file_not_found** (4 strategies for 4 root causes):

1. **correct_path_typo** (Automatic):
   - Root cause: Typo in file path (40% of file_not_found errors)
   - Prerequisites: Fuzzy match found, suggested path exists
   - Steps: Identify corrected path → Verify exists → Replace path → Retry operation
   - Validation: File exists, operation succeeds
   - Success criteria: Operation completes without file_not_found error
   - Rollback: Report ambiguity if multiple matches
   - Success rate estimate: 90%

2. **recreate_deleted_file** (Semi-automatic):
   - Root cause: File deleted or moved (25%)
   - Prerequisites: File content known from history, deletion identified
   - Steps: Retrieve last content → Verify deletion → Prompt user → Recreate → Verify → Retry
   - Validation: File exists, content matches, operation succeeds
   - Success criteria: File recreated with correct content
   - Rollback: Ask user for content source if recreation fails
   - Requires: User confirmation (file content correctness)
   - Success rate estimate: 75%

3. **convert_to_absolute_path** (Automatic):
   - Root cause: Wrong working directory (20%)
   - Prerequisites: File exists in different directory
   - Steps: Get cwd → Search for file → Convert to absolute path → Retry
   - Validation: Absolute path valid, operation succeeds
   - Success criteria: Operation works with absolute path
   - Rollback: Present options if multiple files found
   - Success rate estimate: 95%

4. **add_file_creation_step** (Manual):
   - Root cause: File never existed (15%)
   - Prerequisites: File never existed, path correct
   - Steps: Verify path → Ask user for content source → Create file → Verify → Retry
   - Validation: File exists, has content, operation succeeds
   - Success criteria: File created, operation completes
   - Rollback: Check parent directory exists, verify permissions
   - Requires: Human judgment (what file content should be)
   - Success rate estimate: 60%

**read_before_write_violation** (2 strategies):

5. **insert_read_before_write** (Automatic):
   - Root cause: Protocol forgotten (70% of violations)
   - Prerequisites: File exists, Write/Edit attempted without Read
   - Steps: Detect Write/Edit without prior Read → Insert Read automatically → Execute Read → Proceed with Write/Edit
   - Validation: Read succeeds, Write/Edit succeeds, protocol satisfied
   - Success criteria: Protocol enforced transparently, user unaware
   - Rollback: N/A (Read is always safe, no rollback needed)
   - Success rate estimate: 98%

6. **workflow_simplification** (Manual):
   - Root cause: Complex workflow causing protocol confusion (20%)
   - Prerequisites: Multi-step workflow identified
   - Steps: Analyze workflow → Identify protocol violation points → Redesign workflow → Test
   - Validation: Workflow simpler, protocol violations eliminated
   - Success criteria: No more read_before_write errors
   - Requires: Human judgment (workflow redesign)
   - Success rate estimate: 50%

**Command Execution** (23 strategies):

**command_not_found** (3 strategies):

10. **install_missing_command** (Semi-automatic):
    - Root cause: Command not installed (60% of command_not_found)
    - Prerequisites: Package database available, package identified
    - Steps: Identify package providing command → Prompt user for install confirmation → Install package → Verify command available → Retry
    - Validation: Command exists (`command -v <cmd>`), operation succeeds
    - Success criteria: Command installed, operation completes
    - Rollback: If installation fails, report error and suggest manual installation
    - Requires: User confirmation (system modification)
    - Success rate estimate: 85%

**syntax_error** (3 strategies):

13. **auto_fix_delimiter_mismatch** (Automatic):
    - Root cause: Quote/bracket mismatch (70% of syntax errors)
    - Prerequisites: Delimiter imbalance detected (odd count)
    - Steps: Detect unbalanced delimiters → Balance automatically (add closing quote/bracket) → Verify syntax → Retry
    - Validation: Syntax valid (`bash -n`), operation succeeds
    - Success criteria: Syntax error resolved, operation completes
    - Rollback: If auto-fix creates invalid syntax, report to user
    - Success rate estimate: 80%

**MCP Integration** (21 strategies):

**jq_syntax_error** (3 strategies):

20. **fix_python_syntax_in_jq** (Automatic):
    - Root cause: Python syntax used in jq (20% of jq errors)
    - Prerequisites: Python-specific syntax detected (e.g., `f"..."`, `if ... else`)
    - Steps: Detect Python syntax → Convert to jq syntax → Verify jq validity → Retry
    - Validation: jq filter valid (`echo '{}' | jq '<filter>'`), query succeeds
    - Success criteria: jq query executes correctly
    - Rollback: If conversion fails, report invalid syntax to user
    - Success rate estimate: 75%

**... 51 more recovery strategies**

#### Results

- **Recovery procedures**: 16 complete procedures (100% of diagnostic procedures)
- **Recovery strategies**: 54 (1:1 mapping with root causes)
- **Automation classification**:
  - Automatic: 11 strategies (20.4%)
  - Semi-automatic: 25 strategies (46.3%)
  - Manual: 18 strategies (33.3%)
- **Validation framework**: Complete (checks, criteria, rollback for all procedures)
- **Recovery automation tools specified**: 18 tools (not implemented, specification complete)
- **V_recovery improvement**: 0.25 → 0.70 (+0.45, +180%)
- **Time spent**: ~4-5 hours

### Before/After Comparison

#### Before (No Recovery Procedures)

```
Error: "File does not exist at /path/to/flie.txt"
Root cause identified: Typo in file path

Recovery process:
- Developer manually corrects path
- Developer retries operation
- No validation (did it work?)
- No documentation (how was it fixed?)
- Next time: Same manual process

Issues:
- Manual, slow (1-2 minutes per error)
- Not automated (requires developer intervention)
- No validation (success assumed)
- No knowledge capture (same error fixed same way repeatedly)
```

#### After (With Recovery Procedures)

```
Error: "File does not exist at /path/to/flie.txt"
Root cause identified: Typo in file path
Recovery procedure: correct_path_typo (Automatic)

Recovery process:
1. Fuzzy match: "flie.txt" → "file.txt" (Levenshtein distance = 1)
2. Verification: File exists at "/path/to/file.txt" ✓
3. Path replacement: "/path/to/flie.txt" → "/path/to/file.txt"
4. Retry: Operation succeeds ✓
5. Validation: File accessed, operation complete ✓

Result:
- Automated (no developer intervention)
- Fast (<1 second)
- Validated (success confirmed objectively)
- Documented (recovery procedure reusable)

Success rate: 90% (automatic recovery succeeds 9/10 times)
```

### Verification Checklist

- [ ] Created recovery procedure template (7 components)
- [ ] Defined automation classification criteria (automatic/semi-automatic/manual)
- [ ] Created recovery procedures for all root causes (target: 100% of root causes)
- [ ] Classified automation potential for each strategy
- [ ] Defined prerequisites for each recovery
- [ ] Listed recovery steps (numbered, ordered)
- [ ] Defined validation checks (objective, measurable)
- [ ] Defined success criteria (when recovery is complete)
- [ ] Created rollback procedures (what to do if recovery fails)
- [ ] Documented common pitfalls (warnings, edge cases, risk mitigations)
- [ ] Specified recovery automation tools (optional)
- [ ] Established recovery validation framework
- [ ] Validated procedures on sample errors
- [ ] Calculated coverage and automation metrics

### Pitfalls and How to Avoid

**Pitfall 1**: Over-optimistic automation classification
- ❌ Wrong: "This recovery looks simple, mark as automatic"
- ✅ Right: "Success rate <85%, mark as semi-automatic (requires user confirmation)"

**Pitfall 2**: No rollback procedures
- ❌ Wrong: Assume recovery always succeeds
- ✅ Right: Define rollback for every failure mode

**Pitfall 3**: Missing validation checks
- ❌ Wrong: "Recovery succeeded" (assumed but not verified)
- ✅ Right: "Recovery succeeded: File exists ✓, operation completes ✓, no errors ✓"

**Pitfall 4**: Recovery steps too vague
- ❌ Wrong: "Fix the error"
- ✅ Right: "1. Identify typo using fuzzy matching, 2. Verify corrected path exists, 3. Replace path, 4. Retry"

**Pitfall 5**: Ignoring prerequisites
- ❌ Wrong: Attempt recovery without checking prerequisites
- ✅ Right: Verify prerequisites first (e.g., "fuzzy match found" before path correction)

### Variations

**Variation 1: Lightweight Recovery (Simple Errors)**
- Use when: Errors have obvious single recovery
- Structure: Simplified template (3 components: steps, validation, success criteria)
- Example: "Command not found" → "Install package" → "Verify command exists"

**Variation 2: Multi-Stage Recovery**
- Use when: Recovery requires multiple attempts (fallback strategies)
- Structure: Primary recovery + fallback 1 + fallback 2 + manual escalation
- Example: "File not found" → Try typo correction → Try search → Ask user

**Variation 3: User-Guided Recovery**
- Use when: Errors require user judgment at multiple points
- Structure: Interactive recovery with user prompts at decision points
- Example: "Build failure" → Analyze log → Prompt user: "Missing dependency? [Y/N]" → If Y: Install

**Variation 4**: Transactional Recovery**
- Use when: Recovery may have side effects
- Structure: Begin transaction → Attempt recovery → Validate → Commit if success, rollback if failure
- Example: Database schema change → Backup → Apply change → Test → Commit if tests pass, restore backup if fail

### Reusability

**Language Agnostic**: ⚠️ Recovery strategies are domain-specific, but framework is universal
**Domain Agnostic**: ⚠️ Automation classification and validation framework are universal

**Transferability Assessment**:

**Software Errors (Any Language)**:
- Framework: 7-component recovery procedure template transfers
- Automation classification: Automatic/semi-automatic/manual applies universally
- Strategies: Need adaptation (different recovery actions in Go vs. Python)
- **Transferability**: HIGH (framework), MEDIUM (strategies)

**System Administration**:
- Framework: All components apply
- Example: "Service failed to start" → Automatic: Check port, Semi-auto: Restart service, Manual: Reconfigure
- **Transferability**: HIGH

**Database Operations**:
- Framework: All components apply
- Example: "Deadlock detected" → Automatic: Retry transaction, Semi-auto: Kill blocking transaction, Manual: Redesign schema
- **Transferability**: HIGH

**Network Operations**:
- Framework: All components apply
- Example: "Connection timeout" → Automatic: Retry with exponential backoff, Semi-auto: Restart network service, Manual: Reconfigure firewall
- **Transferability**: HIGH

**DevOps/Infrastructure**:
- Framework: All components apply (especially rollback procedures critical)
- Example: "Deployment failed" → Automatic: Rollback to previous version, Semi-auto: Scale down new version, Manual: Debug and fix
- **Transferability**: HIGH

**Web Applications**:
- Framework: All components apply
- Example: "API request failed (429 rate limit)" → Automatic: Exponential backoff retry, Semi-auto: Request rate limit increase, Manual: Redesign request pattern
- **Transferability**: HIGH

### Key Takeaways

- ✅ **1:1 mapping: root causes → recovery strategies** - every root cause needs recovery
- ✅ **7-component template** - metadata, prerequisites, steps, validation, success criteria, rollback, pitfalls
- ✅ **Conservative automation classification** - prefer semi-automatic over automatic when uncertain
- ✅ **Validation is mandatory** - every recovery must have objective validation checks
- ✅ **Rollback for failure modes** - define what to do when recovery fails
- ✅ **Target 100% coverage** - all diagnostic procedures should have recovery procedures
- ✅ **Document thoroughly** - recovery procedures are operational playbooks

**One-sentence summary**: Create systematic recovery procedures with 7 components (metadata, prerequisites, steps, validation, success criteria, rollback, pitfalls) for each root cause, classify automation potential (automatic/semi-automatic/manual), and establish validation framework to verify recovery success.

---

*[Continue with Patterns 4-8...]*

