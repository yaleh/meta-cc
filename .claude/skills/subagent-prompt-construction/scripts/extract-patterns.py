#!/usr/bin/env python3
"""extract-patterns.py - Extract and summarize patterns from reference directory"""

import json
import re
from pathlib import Path
from typing import Dict, List


def extract_patterns(reference_dir: Path) -> Dict:
    """Extract patterns from reference/patterns.md"""
    patterns_file = reference_dir / "patterns.md"

    if not patterns_file.exists():
        return {"error": "patterns.md not found"}

    content = patterns_file.read_text()

    patterns = []

    # Extract pattern sections
    pattern_regex = r"## Pattern \d+: (.+?)\n\n\*\*Use case\*\*: (.+?)\n\n\*\*Structure\*\*:\n```\n(.+?)\n```"

    for match in re.finditer(pattern_regex, content, re.DOTALL):
        name = match.group(1).strip()
        use_case = match.group(2).strip()
        structure = match.group(3).strip()

        patterns.append({
            "name": name,
            "use_case": use_case,
            "structure": structure
        })

    return {
        "patterns_count": len(patterns),
        "patterns": patterns
    }


def extract_integration_patterns(reference_dir: Path) -> Dict:
    """Extract integration patterns from reference/integration-patterns.md"""
    integration_file = reference_dir / "integration-patterns.md"

    if not integration_file.exists():
        return {"error": "integration-patterns.md not found"}

    content = integration_file.read_text()

    integrations = []

    # Extract integration sections
    integration_regex = r"## \d+\. (.+?)\n\n\*\*Pattern\*\*:\n```\n(.+?)\n```"

    for match in re.finditer(integration_regex, content, re.DOTALL):
        name = match.group(1).strip()
        pattern = match.group(2).strip()

        integrations.append({
            "name": name,
            "pattern": pattern
        })

    return {
        "integration_patterns_count": len(integrations),
        "integration_patterns": integrations
    }


def extract_symbols(reference_dir: Path) -> Dict:
    """Extract symbolic language operators from reference/symbolic-language.md"""
    symbols_file = reference_dir / "symbolic-language.md"

    if not symbols_file.exists():
        return {"error": "symbolic-language.md not found"}

    content = symbols_file.read_text()

    # Count sections
    logic_ops = len(re.findall(r"### .+? \(.+?\)\n\*\*Symbol\*\*: `(.+?)`", content[:2000]))
    quantifiers = len(re.findall(r"### .+?\n\*\*Symbol\*\*: `(.+?)`", content[2000:4000]))
    set_ops = len(re.findall(r"### .+?\n\*\*Symbol\*\*: `(.+?)`", content[4000:6000]))

    return {
        "logic_operators": logic_ops,
        "quantifiers": quantifiers,
        "set_operations": set_ops,
        "total_symbols": logic_ops + quantifiers + set_ops
    }


def main():
    """Main entry point"""
    skill_dir = Path(__file__).parent.parent
    reference_dir = skill_dir / "reference"

    if not reference_dir.exists():
        print(json.dumps({"error": "reference directory not found"}, indent=2))
        return

    # Extract all patterns
    patterns = extract_patterns(reference_dir)
    integrations = extract_integration_patterns(reference_dir)
    symbols = extract_symbols(reference_dir)

    # Combine results
    result = {
        "skill": "subagent-prompt-construction",
        "patterns": patterns,
        "integration_patterns": integrations,
        "symbolic_language": symbols,
        "summary": {
            "total_patterns": patterns.get("patterns_count", 0),
            "total_integration_patterns": integrations.get("integration_patterns_count", 0),
            "total_symbols": symbols.get("total_symbols", 0)
        }
    }

    # Save to inventory
    inventory_dir = skill_dir / "inventory"
    inventory_dir.mkdir(exist_ok=True)

    output_file = inventory_dir / "patterns-summary.json"
    output_file.write_text(json.dumps(result, indent=2))

    print(f"✅ Patterns extracted to {output_file}")
    print(f"   - {result['summary']['total_patterns']} core patterns")
    print(f"   - {result['summary']['total_integration_patterns']} integration patterns")
    print(f"   - {result['summary']['total_symbols']} symbolic operators")


if __name__ == "__main__":
    main()
