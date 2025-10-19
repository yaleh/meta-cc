#!/usr/bin/env python3
"""
Generate SKILL.md frontmatter from experiment results.md file.

Usage:
    ./generate-frontmatter.py <results.md> [--output OUTPUT]

Example:
    ./generate-frontmatter.py experiments/bootstrap-006-api-design/results.md --output frontmatter.yaml
"""

import argparse
import re
import sys
from pathlib import Path
from typing import Dict, Any, List, Optional


def extract_convergence_data(content: str) -> Dict[str, Any]:
    """Extract convergence metrics from results.md."""
    data = {}

    # Extract V_instance final value
    match = re.search(r'V_instance\(s[₀-₉\d]+\)\s*=\s*([\d.]+)', content)
    if match:
        data['v_instance'] = float(match.group(1))

    # Extract V_meta final value
    match = re.search(r'V_meta\(s[₀-₉\d]+\)\s*=\s*([\d.]+)', content)
    if match:
        data['v_meta'] = float(match.group(1))

    # Extract iteration count
    matches = re.findall(r'Iteration\s+(\d+):', content)
    if matches:
        data['iterations'] = max(int(m) for m in matches) + 1  # 0-indexed

    # Extract convergence status
    if re.search(r'CONVERGED|CONVERGENCE ACHIEVED', content, re.IGNORECASE):
        data['converged'] = True
    else:
        data['converged'] = False

    return data


def extract_patterns_count(content: str) -> int:
    """Count patterns mentioned in results.md."""
    # Count "### Pattern X:" headers
    pattern_headers = re.findall(r'^###\s+Pattern\s*\d+:', content, re.MULTILINE)
    return len(pattern_headers)


def extract_principles_count(content: str) -> int:
    """Count principles/lessons mentioned in results.md."""
    # Count "### Lesson N:" or "### Principle N:" headers
    lesson_headers = re.findall(r'^###\s+(?:Lesson|Principle)\s+\d+:', content, re.MULTILINE)
    return len(lesson_headers)


def extract_transferability_data(content: str) -> Dict[str, Any]:
    """Extract transferability/reusability metrics."""
    data = {}

    # Look for percentage mentions
    match = re.search(r'(\d+(?:\.\d+)?)\s*%.*?transfer', content, re.IGNORECASE)
    if match:
        data['transferability_pct'] = float(match.group(1))

    # Look for language mentions
    languages = []
    for lang in ['Python', 'TypeScript', 'Rust', 'Java', 'Go', 'JavaScript']:
        if re.search(lang, content, re.IGNORECASE):
            languages.append(lang)
    if languages:
        data['languages'] = languages

    return data


def extract_validation_evidence(content: str) -> List[str]:
    """Extract validation evidence points."""
    evidence = []

    # Look for validation sections
    validation_section = re.search(
        r'## Validation.*?(?=##|\Z)',
        content,
        re.DOTALL | re.IGNORECASE
    )

    if validation_section:
        section_content = validation_section.group(0)
        # Extract bullet points
        bullets = re.findall(r'^\s*[-\*]\s+(.+)$', section_content, re.MULTILINE)
        evidence.extend(bullets[:5])  # Limit to top 5

    return evidence


def generate_skill_name(results_path: Path) -> str:
    """Infer skill name from experiment directory."""
    experiment_name = results_path.parent.name

    # Remove bootstrap-NNN- prefix
    name = re.sub(r'^bootstrap-\d+-', '', experiment_name)

    # Convert kebab-case to Title Case
    name = name.replace('-', ' ').title()

    return name


def extract_when_to_use(content: str, patterns_count: int) -> List[str]:
    """Generate 'when to use' bullet points."""
    use_cases = []

    # If we have an explicit section
    when_section = re.search(
        r'## When to Use.*?(?=##|\Z)',
        content,
        re.DOTALL | re.IGNORECASE
    )

    if when_section:
        section_content = when_section.group(0)
        bullets = re.findall(r'^\s*[-\*]\s+(.+)$', section_content, re.MULTILINE)
        use_cases.extend(bullets[:6])
    else:
        # Infer from patterns
        use_cases.append(f"Need systematic methodology with {patterns_count} validated patterns")
        use_cases.append("Building similar system from scratch")
        use_cases.append("Improving existing implementation quality")

    return use_cases


def generate_description(
    skill_name: str,
    patterns_count: int,
    convergence_data: Dict[str, Any],
    transferability: Dict[str, Any]
) -> str:
    """
    Generate description field (max 400 chars).

    Format: [Methodology summary]. Use when [use cases]. Provides [patterns/tools/metrics]. Validated [evidence].
    """
    parts = []

    # Methodology summary
    parts.append(f"Systematic {skill_name.lower()} methodology")

    # Patterns/tools count
    if patterns_count > 0:
        parts.append(f"with {patterns_count} validated patterns")

    # Use cases (brief)
    parts.append(f"Use when establishing {skill_name.lower()} from scratch or improving existing implementation")

    # Validation metrics
    if convergence_data.get('v_instance'):
        v_i = convergence_data['v_instance']
        parts.append(f"V_instance={v_i:.2f}")

    if transferability.get('transferability_pct'):
        t_pct = transferability['transferability_pct']
        parts.append(f"{t_pct}% transferability")

    description = '. '.join(parts) + '.'

    # Truncate if needed
    if len(description) > 400:
        description = description[:397] + '...'

    return description


def generate_frontmatter(results_path: Path) -> Dict[str, Any]:
    """Generate complete frontmatter YAML from results.md."""
    content = results_path.read_text()

    skill_name = generate_skill_name(results_path)
    convergence_data = extract_convergence_data(content)
    patterns_count = extract_patterns_count(content)
    principles_count = extract_principles_count(content)
    transferability = extract_transferability_data(content)
    validation_evidence = extract_validation_evidence(content)
    when_to_use = extract_when_to_use(content, patterns_count)

    description = generate_description(
        skill_name,
        patterns_count,
        convergence_data,
        transferability
    )

    frontmatter = {
        "name": skill_name,
        "description": description,
        "allowed-tools": ["Read", "Write", "Edit", "Bash", "Grep", "Glob"],
        "_metadata": {
            "extraction_source": results_path.parent.name,
            "patterns_count": patterns_count,
            "principles_count": principles_count,
            "convergence": convergence_data,
            "transferability": transferability,
            "validation_evidence": validation_evidence,
            "when_to_use": when_to_use
        }
    }

    return frontmatter


def main():
    parser = argparse.ArgumentParser(description='Generate SKILL.md frontmatter from results.md')
    parser.add_argument('results_file', type=Path, help='Path to results.md file')
    parser.add_argument('--output', '-o', type=Path, default=None, help='Output YAML file path')
    parser.add_argument('--format', choices=['yaml', 'markdown'], default='yaml', help='Output format')

    args = parser.parse_args()

    if not args.results_file.exists():
        print(f"Error: Results file not found: {args.results_file}", file=sys.stderr)
        sys.exit(1)

    frontmatter = generate_frontmatter(args.results_file)

    if args.format == 'yaml':
        # Generate YAML output
        lines = ["---"]
        lines.append(f"name: {frontmatter['name']}")
        lines.append(f"description: {frontmatter['description']}")
        lines.append(f"allowed-tools: {', '.join(frontmatter['allowed-tools'])}")
        lines.append("---")
        output = '\n'.join(lines)

        if args.output:
            args.output.write_text(output)
            print(f"Generated frontmatter: {args.output}")
        else:
            print(output)

        # Also print metadata to stderr for reference
        print("\nMetadata (for reference):", file=sys.stderr)
        import json
        print(json.dumps(frontmatter['_metadata'], indent=2), file=sys.stderr)

    elif args.format == 'markdown':
        # Generate full markdown header
        import json
        md_lines = ["---"]
        md_lines.append(f"name: {frontmatter['name']}")
        md_lines.append(f"description: {frontmatter['description']}")
        md_lines.append(f"allowed-tools: {', '.join(frontmatter['allowed-tools'])}")
        md_lines.append("---")
        md_lines.append("")
        md_lines.append(f"# {frontmatter['name']}")
        md_lines.append("")
        md_lines.append("**[Tagline here]**")
        md_lines.append("")
        md_lines.append("## When to Use This Skill")
        md_lines.append("")
        for use_case in frontmatter['_metadata']['when_to_use']:
            md_lines.append(f"- {use_case}")
        md_lines.append("")
        md_lines.append("## Validation")
        md_lines.append("")
        md_lines.append(f"- **Patterns**: {frontmatter['_metadata']['patterns_count']}")
        md_lines.append(f"- **Principles**: {frontmatter['_metadata']['principles_count']}")
        if frontmatter['_metadata']['convergence'].get('v_instance'):
            md_lines.append(f"- **V_instance**: {frontmatter['_metadata']['convergence']['v_instance']}")
        if frontmatter['_metadata']['transferability'].get('transferability_pct'):
            md_lines.append(f"- **Transferability**: {frontmatter['_metadata']['transferability']['transferability_pct']}%")
        md_lines.append("")

        output = '\n'.join(md_lines)

        if args.output:
            args.output.write_text(output)
            print(f"Generated markdown: {args.output}")
        else:
            print(output)


if __name__ == '__main__':
    main()
