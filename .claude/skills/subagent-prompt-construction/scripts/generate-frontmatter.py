#!/usr/bin/env python3
"""generate-frontmatter.py - Generate skill frontmatter inventory"""

import json
import re
from pathlib import Path
from typing import Dict


def extract_frontmatter(skill_md: Path) -> Dict:
    """Extract YAML frontmatter from SKILL.md"""
    if not skill_md.exists():
        return {"error": "SKILL.md not found"}

    content = skill_md.read_text()

    # Extract frontmatter between --- delimiters
    match = re.search(r"^---\n(.+?)\n---", content, re.DOTALL | re.MULTILINE)
    if not match:
        return {"error": "No frontmatter found"}

    frontmatter_text = match.group(1)

    # Parse YAML-style frontmatter
    frontmatter = {}
    for line in frontmatter_text.split("\n"):
        if ":" in line:
            key, value = line.split(":", 1)
            key = key.strip()
            value = value.strip()

            # Try to parse as number or boolean
            if value.replace(".", "").isdigit():
                value = float(value) if "." in value else int(value)
            elif value.lower() in ["true", "false"]:
                value = value.lower() == "true"
            elif value.endswith("%"):
                value = int(value[:-1])

            frontmatter[key] = value

    return frontmatter


def extract_lambda_contract(skill_md: Path) -> str:
    """Extract lambda contract from SKILL.md"""
    if not skill_md.exists():
        return ""

    content = skill_md.read_text()

    # Find lambda contract (starts with λ)
    match = re.search(r"^λ\(.+?\).*$", content, re.MULTILINE)
    if match:
        return match.group(0)

    return ""


def main():
    """Main entry point"""
    skill_dir = Path(__file__).parent.parent
    skill_md = skill_dir / "SKILL.md"

    if not skill_md.exists():
        print(json.dumps({"error": "SKILL.md not found"}, indent=2))
        return

    # Extract frontmatter and lambda contract
    frontmatter = extract_frontmatter(skill_md)
    lambda_contract = extract_lambda_contract(skill_md)

    # Calculate metrics
    skill_lines = len(skill_md.read_text().split("\n"))

    # Count examples
    examples_dir = skill_dir / "examples"
    examples_count = len(list(examples_dir.glob("*.md"))) if examples_dir.exists() else 0

    # Count reference files
    reference_dir = skill_dir / "reference"
    reference_count = len(list(reference_dir.glob("*.md"))) if reference_dir.exists() else 0

    # Count case studies
    case_studies_dir = reference_dir / "case-studies" if reference_dir.exists() else None
    case_studies_count = len(list(case_studies_dir.glob("*.md"))) if case_studies_dir and case_studies_dir.exists() else 0

    # Combine results
    result = {
        "skill": "subagent-prompt-construction",
        "frontmatter": frontmatter,
        "lambda_contract": lambda_contract,
        "metrics": {
            "skill_md_lines": skill_lines,
            "examples_count": examples_count,
            "reference_files_count": reference_count,
            "case_studies_count": case_studies_count
        },
        "compliance": {
            "skill_md_under_40_lines": skill_lines <= 40,
            "has_lambda_contract": len(lambda_contract) > 0,
            "has_examples": examples_count > 0,
            "has_reference": reference_count > 0
        }
    }

    # Save to inventory
    inventory_dir = skill_dir / "inventory"
    inventory_dir.mkdir(exist_ok=True)

    output_file = inventory_dir / "skill-frontmatter.json"
    output_file.write_text(json.dumps(result, indent=2))

    print(f"✅ Frontmatter extracted to {output_file}")
    print(f"   - SKILL.md: {skill_lines} lines ({'✅' if skill_lines <= 40 else '⚠️  over'})")
    print(f"   - Examples: {examples_count}")
    print(f"   - Reference files: {reference_count}")
    print(f"   - Case studies: {case_studies_count}")


if __name__ == "__main__":
    main()
