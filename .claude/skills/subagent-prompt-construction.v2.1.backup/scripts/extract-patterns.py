#!/usr/bin/env python3
"""
Extract pattern summary from reference/patterns.md
"""

import json
import os
import re
from pathlib import Path

def extract_patterns(patterns_file):
    """Extract pattern information from patterns.md"""
    patterns = []

    with open(patterns_file, 'r') as f:
        content = f.read()

    # Find all pattern sections (## Pattern N: ...)
    pattern_sections = re.findall(
        r'## (Pattern \d+: .+?)\n(.*?)(?=\n## |\Z)',
        content,
        re.DOTALL
    )

    for title, body in pattern_sections:
        # Extract use case
        use_case_match = re.search(r'\*\*Use Case\*\*: (.+)', body)
        use_case = use_case_match.group(1) if use_case_match else "N/A"

        # Extract when to use
        when_match = re.search(r'\*\*When to Use\*\*:\n(.*?)(?=\n\n|\*\*)', body, re.DOTALL)
        when_to_use = []
        if when_match:
            when_to_use = [
                line.strip('- ').strip()
                for line in when_match.group(1).split('\n')
                if line.strip().startswith('-')
            ]

        # Check validation status
        validated = "validated" in body.lower() and "✅" in body

        patterns.append({
            "name": title,
            "use_case": use_case,
            "when_to_use": when_to_use,
            "validated": validated
        })

    return patterns


def extract_integration_patterns(integration_file):
    """Extract integration pattern information"""
    patterns = []

    with open(integration_file, 'r') as f:
        content = f.read()

    # Find all pattern sections
    pattern_sections = re.findall(
        r'## (Pattern \d+: .+?)\n(.*?)(?=\n## |\Z)',
        content,
        re.DOTALL
    )

    for title, body in pattern_sections:
        # Extract syntax
        syntax_match = re.search(r'### Syntax\n+```\n(.+?)\n```', body, re.DOTALL)
        syntax = syntax_match.group(1).strip() if syntax_match else "N/A"

        patterns.append({
            "name": title,
            "syntax": syntax
        })

    return patterns


def main():
    script_dir = Path(__file__).parent
    skill_dir = script_dir.parent
    inventory_dir = skill_dir / "inventory"
    inventory_dir.mkdir(exist_ok=True)

    # Extract construction patterns
    patterns_file = skill_dir / "reference" / "patterns.md"
    if patterns_file.exists():
        patterns = extract_patterns(patterns_file)
        output = {
            "skill": "subagent-prompt-construction",
            "pattern_count": len(patterns),
            "patterns": patterns
        }

        output_file = inventory_dir / "patterns-summary.json"
        with open(output_file, 'w') as f:
            json.dump(output, f, indent=2)

        print(f"✓ Extracted {len(patterns)} construction patterns")
        print(f"  Output: {output_file}")
    else:
        print(f"⚠ patterns.md not found at {patterns_file}")

    # Extract integration patterns
    integration_file = skill_dir / "reference" / "integration-patterns.md"
    if integration_file.exists():
        int_patterns = extract_integration_patterns(integration_file)
        output = {
            "skill": "subagent-prompt-construction",
            "integration_pattern_count": len(int_patterns),
            "patterns": int_patterns
        }

        output_file = inventory_dir / "integration-patterns-summary.json"
        with open(output_file, 'w') as f:
            json.dump(output, f, indent=2)

        print(f"✓ Extracted {len(int_patterns)} integration patterns")
        print(f"  Output: {output_file}")
    else:
        print(f"⚠ integration-patterns.md not found at {integration_file}")


if __name__ == "__main__":
    main()
