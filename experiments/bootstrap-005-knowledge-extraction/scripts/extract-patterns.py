#!/usr/bin/env python3
"""
Extract patterns from BAIME experiment iteration reports and results.md files.

Usage:
    ./extract-patterns.py <experiment_dir> [--output OUTPUT]

Example:
    ./extract-patterns.py experiments/bootstrap-006-api-design --output data/patterns.json
"""

import argparse
import json
import re
import sys
from pathlib import Path
from typing import List, Dict, Any


def extract_patterns_from_results(results_path: Path) -> List[Dict[str, Any]]:
    """
    Extract patterns from results.md file.

    Looks for:
    - "### Pattern: NAME" or "### Pattern N: NAME" sections
    - **Context**: ...
    - **Solution**: ...
    - **Evidence**: ...
    - **Reusability**: ...
    """
    patterns = []

    if not results_path.exists():
        return patterns

    content = results_path.read_text()
    lines = content.split('\n')

    i = 0
    while i < len(lines):
        line = lines[i]

        # Match pattern headers: "### Pattern: NAME" or "### Pattern 1: NAME"
        match = re.match(r'^###\s+Pattern\s*(?:\d+)?:\s+(.+)$', line)
        if match:
            pattern_name = match.group(1).strip()
            pattern = {
                "name": pattern_name,
                "source_file": "results.md",
                "source_line": i + 1,
                "context": None,
                "problem": None,
                "solution": None,
                "example": None,
                "evidence": None,
                "reusability": None
            }

            # Extract pattern components
            i += 1
            while i < len(lines) and not lines[i].startswith('###'):
                line = lines[i]

                if line.startswith('**Context**:'):
                    pattern['context'] = line.replace('**Context**:', '').strip()
                elif line.startswith('**Problem**:'):
                    pattern['problem'] = line.replace('**Problem**:', '').strip()
                elif line.startswith('**Solution**:'):
                    # Multi-line solution
                    solution_lines = [line.replace('**Solution**:', '').strip()]
                    i += 1
                    while i < len(lines) and not lines[i].startswith('**') and not lines[i].startswith('###'):
                        if lines[i].strip():
                            solution_lines.append(lines[i].strip())
                        i += 1
                    pattern['solution'] = ' '.join(solution_lines)
                    continue
                elif line.startswith('**Evidence**:'):
                    pattern['evidence'] = line.replace('**Evidence**:', '').strip()
                elif line.startswith('**Reusability**:') or line.startswith('**Transferability**:'):
                    pattern['reusability'] = line.replace('**Reusability**:', '').replace('**Transferability**:', '').strip()
                elif line.startswith('**Example**:'):
                    # Mark that example exists
                    pattern['example'] = "See source"

                i += 1

            patterns.append(pattern)
        else:
            i += 1

    return patterns


def extract_patterns_from_iterations(iteration_dir: Path) -> List[Dict[str, Any]]:
    """
    Extract pattern mentions from iteration reports.

    Looks for:
    - "Pattern X: NAME" mentions
    - "Applied Pattern X" references
    """
    patterns = []

    if not iteration_dir.exists():
        return patterns

    iteration_files = sorted(iteration_dir.glob('iteration-*.md'))

    for iteration_file in iteration_files:
        content = iteration_file.read_text()
        lines = content.split('\n')

        for i, line in enumerate(lines):
            # Match "Pattern N: NAME" or "- Pattern N: NAME"
            match = re.match(r'^[-\*]?\s*Pattern\s+(\d+):\s+(.+)$', line)
            if match:
                pattern_num = match.group(1)
                pattern_name = match.group(2).strip()

                patterns.append({
                    "name": pattern_name,
                    "source_file": iteration_file.name,
                    "source_line": i + 1,
                    "pattern_number": int(pattern_num),
                    "context": "iteration_mention"
                })

    return patterns


def deduplicate_patterns(patterns: List[Dict[str, Any]]) -> List[Dict[str, Any]]:
    """Deduplicate patterns by name, preferring results.md sources."""
    seen = {}

    # Sort to prioritize results.md sources
    sorted_patterns = sorted(patterns, key=lambda p: (p.get('source_file', '') != 'results.md', p.get('name', '')))

    for pattern in sorted_patterns:
        name = pattern.get('name', '').lower()
        if name not in seen:
            seen[name] = pattern

    return list(seen.values())


def main():
    parser = argparse.ArgumentParser(description='Extract patterns from BAIME experiment')
    parser.add_argument('experiment_dir', type=Path, help='Path to experiment directory')
    parser.add_argument('--output', '-o', type=Path, default=None, help='Output JSON file path')
    parser.add_argument('--format', choices=['json', 'markdown'], default='json', help='Output format')

    args = parser.parse_args()

    if not args.experiment_dir.exists():
        print(f"Error: Experiment directory not found: {args.experiment_dir}", file=sys.stderr)
        sys.exit(1)

    # Extract from results.md
    results_path = args.experiment_dir / 'results.md'
    patterns_from_results = extract_patterns_from_results(results_path)

    # Extract from iterations/
    iteration_dir = args.experiment_dir
    patterns_from_iterations = extract_patterns_from_iterations(iteration_dir)

    # Combine and deduplicate
    all_patterns = patterns_from_results + patterns_from_iterations
    unique_patterns = deduplicate_patterns(all_patterns)

    # Output
    output_data = {
        "experiment": args.experiment_dir.name,
        "patterns_count": len(unique_patterns),
        "patterns": unique_patterns
    }

    if args.format == 'json':
        json_output = json.dumps(output_data, indent=2)

        if args.output:
            args.output.write_text(json_output)
            print(f"Extracted {len(unique_patterns)} patterns to {args.output}")
        else:
            print(json_output)

    elif args.format == 'markdown':
        md_lines = [
            f"# Patterns from {args.experiment_dir.name}",
            "",
            f"**Total**: {len(unique_patterns)} patterns",
            ""
        ]

        for i, pattern in enumerate(unique_patterns, 1):
            md_lines.append(f"## Pattern {i}: {pattern['name']}")
            md_lines.append("")

            if pattern.get('context'):
                md_lines.append(f"**Context**: {pattern['context']}")
            if pattern.get('problem'):
                md_lines.append(f"**Problem**: {pattern['problem']}")
            if pattern.get('solution'):
                md_lines.append(f"**Solution**: {pattern['solution']}")
            if pattern.get('evidence'):
                md_lines.append(f"**Evidence**: {pattern['evidence']}")
            if pattern.get('reusability'):
                md_lines.append(f"**Reusability**: {pattern['reusability']}")

            md_lines.append(f"**Source**: {pattern['source_file']} (line {pattern.get('source_line', 'unknown')})")
            md_lines.append("")

        md_output = '\n'.join(md_lines)

        if args.output:
            args.output.write_text(md_output)
            print(f"Extracted {len(unique_patterns)} patterns to {args.output}")
        else:
            print(md_output)


if __name__ == '__main__':
    main()
