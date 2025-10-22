#!/usr/bin/env python3
"""Extract bullet list of patterns with iteration references."""
import json
import pathlib

skill_dir = pathlib.Path(__file__).resolve().parents[1]
patterns_file = skill_dir / "reference" / "patterns.md"
summary_file = skill_dir / "knowledge" / "patterns-summary.json"

patterns = []
current = None
with patterns_file.open("r", encoding="utf-8") as fh:
    for line in fh:
        line = line.strip()
        if line.startswith("- **") and "**" in line[3:]:
            name = line[4:line.find("**", 4)]
            rest = line[line.find("**", 4) + 2:].strip(" -")
            patterns.append({"name": name, "description": rest})

summary = {
    "pattern_count": len(patterns),
    "patterns": patterns,
}
summary_file.write_text(json.dumps(summary, indent=2), encoding="utf-8")
print(json.dumps(summary, indent=2))
