#!/usr/bin/env python3
"""
Extract frontmatter metadata from SKILL.md
"""

import json
import re
from pathlib import Path

def extract_frontmatter(skill_file):
    """Extract YAML frontmatter from SKILL.md"""
    with open(skill_file, 'r') as f:
        content = f.read()

    # Extract frontmatter (between --- markers)
    frontmatter_match = re.match(r'^---\n(.*?)\n---', content, re.DOTALL)
    if not frontmatter_match:
        return None

    frontmatter = {}
    for line in frontmatter_match.group(1).split('\n'):
        if ':' in line:
            key, value = line.split(':', 1)
            key = key.strip()
            value = value.strip()

            # Parse boolean
            if value.lower() in ('true', 'false'):
                value = value.lower() == 'true'
            # Parse numbers
            elif value.replace('.', '').isdigit():
                value = float(value) if '.' in value else int(value)

            frontmatter[key] = value

    return frontmatter


def extract_metrics(skill_file):
    """Extract skill metrics from SKILL.md"""
    with open(skill_file, 'r') as f:
        content = f.read()

    metrics = {}

    # Extract validation scores
    v_instance_match = re.search(r'V_instance\s*[≥>=]\s*(\d+\.\d+)', content)
    v_meta_match = re.search(r'V_meta\s*[≥>=]\s*(\d+\.\d+)', content)

    if v_instance_match:
        metrics['v_instance_target'] = float(v_instance_match.group(1))
    if v_meta_match:
        metrics['v_meta_target'] = float(v_meta_match.group(1))

    # Extract from frontmatter if available
    frontmatter = extract_frontmatter(skill_file)
    if frontmatter:
        if 'v_instance' in frontmatter:
            metrics['v_instance_achieved'] = frontmatter['v_instance']
        if 'v_meta' in frontmatter:
            metrics['v_meta_achieved'] = frontmatter['v_meta']

    return metrics


def main():
    script_dir = Path(__file__).parent
    skill_dir = script_dir.parent
    inventory_dir = skill_dir / "inventory"
    inventory_dir.mkdir(exist_ok=True)

    skill_file = skill_dir / "SKILL.md"
    if not skill_file.exists():
        print(f"⚠ SKILL.md not found at {skill_file}")
        return

    # Extract frontmatter
    frontmatter = extract_frontmatter(skill_file)
    if frontmatter:
        output_file = inventory_dir / "skill-frontmatter.json"
        with open(output_file, 'w') as f:
            json.dump(frontmatter, f, indent=2)
        print(f"✓ Extracted frontmatter metadata")
        print(f"  Output: {output_file}")

        # Pretty print
        print("\nFrontmatter:")
        for key, value in frontmatter.items():
            print(f"  {key}: {value}")
    else:
        print("⚠ No frontmatter found in SKILL.md")

    # Extract metrics
    metrics = extract_metrics(skill_file)
    if metrics:
        output_file = inventory_dir / "skill-metrics.json"
        with open(output_file, 'w') as f:
            json.dump(metrics, f, indent=2)
        print(f"\n✓ Extracted skill metrics")
        print(f"  Output: {output_file}")

        # Pretty print
        print("\nMetrics:")
        for key, value in metrics.items():
            print(f"  {key}: {value}")
    else:
        print("\n⚠ No metrics found in SKILL.md")


if __name__ == "__main__":
    main()
