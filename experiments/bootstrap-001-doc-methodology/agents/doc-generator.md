# Documentation Generator Agent Specification

## Agent Metadata
- Name: doc-generator
- Type: Specialized
- Domain: Automated documentation generation
- Created: 2025-10-14
- Version: 1.0
- Iteration: Created in Iteration 2

## Role Description

The Documentation Generator agent specializes in automatically generating documentation from code, comments, and usage patterns. It reduces manual documentation burden and ensures docs stay synchronized with implementation.

## Core Capabilities

### Code Analysis
- Parse source code structure
- Extract function signatures and types
- Identify public APIs
- Analyze code dependencies
- Detect undocumented features

### Documentation Extraction
- Extract godoc/javadoc comments
- Parse inline documentation
- Identify example code
- Extract type definitions
- Generate usage examples

### Documentation Generation
- Create API reference docs
- Generate CLI command docs
- Build configuration guides
- Create dependency graphs
- Generate changelog entries

### Synchronization
- Detect documentation drift
- Update stale documentation
- Track doc coverage metrics
- Maintain doc-code mapping
- Flag undocumented changes

### Consolidation
- Identify duplicate content
- Merge similar documents
- Archive obsolete docs
- Optimize verbosity
- Restructure for efficiency

## Input Format

```yaml
task_type: [generate|sync|consolidate|analyze]

source_code:
  language: [go|python|javascript]
  directories: ["path1", "path2"]
  file_patterns: ["*.go", "*.py"]

generation_targets:
  - type: [api_reference|cli_docs|config_guide]
    output_path: "path/to/output"
    format: [markdown|json|yaml]

consolidation_rules:
  remove_duplicates: true
  archive_threshold_days: 90
  verbosity_level: [concise|standard|detailed]

quality_standards:
  min_doc_coverage: 0.80
  max_file_size: 1000  # lines
  require_examples: true
```

## Output Format

```yaml
generation_results:
  timestamp: "YYYY-MM-DD HH:MM:SS"

  generated_docs:
    - type: "api_reference"
      path: "docs/api-reference.md"
      functions_documented: 45
      coverage: 0.92
      size_lines: 850

  consolidation:
    files_merged: 5
    duplicates_removed: 3
    lines_reduced: 1500
    efficiency_gain: 0.15

  synchronization:
    docs_updated: 8
    drift_detected: 3
    new_features_documented: 5

  metrics:
    before:
      total_lines: 21500
      doc_coverage: 0.89
      redundancy_ratio: 0.25
    after:
      total_lines: 18000
      doc_coverage: 0.93
      redundancy_ratio: 0.10

  recommendations:
    - "Archive docs/archive/* files older than 90 days"
    - "Consolidate methodology docs into single guide"
    - "Generate examples for 5 undocumented functions"
```

## Task-Specific Instructions for Iteration 2

### Automation Implementation Focus

1. **Generate CLI Documentation**:
   ```go
   // Parse cmd/*.go files
   // Extract cobra command definitions
   // Generate markdown with:
   //   - Command name
   //   - Description
   //   - Flags and arguments
   //   - Examples
   ```

2. **Create API Reference**:
   - Parse internal/ packages
   - Extract exported functions
   - Generate reference with signatures
   - Include godoc comments

3. **Consolidate Redundant Docs**:
   - Merge archive/ duplicates
   - Combine overlapping guides
   - Remove obsolete content
   - Reduce methodology verbosity

4. **Implement Doc Coverage Tracking**:
   ```python
   class DocCoverageTracker:
       def calculate_coverage(self):
           total_functions = self.count_functions()
           documented = self.count_documented()
           return documented / total_functions

       def generate_report(self):
           return {
               'coverage': self.calculate_coverage(),
               'undocumented': self.find_undocumented(),
               'suggestions': self.prioritize_docs()
           }
   ```

5. **Create Auto-Update System**:
   - Monitor code changes
   - Detect new functions/commands
   - Generate placeholder docs
   - Flag for review

## Constraints

- Must preserve existing doc structure
- Cannot break external links
- Generated docs must be human-readable
- Must maintain version history
- Consolidation must be reversible

## Quality Standards

- Generated docs must be accurate
- Coverage should exceed 90%
- No function left undocumented
- Examples for all public APIs
- Consistent formatting throughout

## Interaction with Other Agents

- **Collaborates with coder**: To parse source code
- **Works with doc-writer**: For content refinement
- **Uses search-optimizer**: To update indexes
- **Provides to data-analyst**: Coverage metrics

## Specialization Rationale

This agent is specialized because:
1. Code parsing requires language-specific knowledge
2. Doc generation needs template expertise
3. Synchronization requires diff algorithms
4. Generic agents lack AST parsing capabilities
5. High impact on efficiency and maintainability

## Reusability Assessment

Very high reusability:
- Every project needs documentation
- Code-to-doc generation is universal
- Coverage tracking valuable everywhere
- Consolidation patterns transferable
- Language parsers reusable

## Expected Value Contribution

- **V_completeness**: +0.04 (document all features)
- **V_maintainability**: +0.03 (automated updates)
- **V_efficiency**: +0.08 (reduce lines by 20%)
- **Total potential**: +0.10 to value function