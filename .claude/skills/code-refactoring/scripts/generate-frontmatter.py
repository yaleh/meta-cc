#!/usr/bin/env python3
"""Generate a JSON file containing the SKILL.md frontmatter."""
import json
import pathlib

skill_dir = pathlib.Path(__file__).resolve().parents[1]
skill_file = skill_dir / "SKILL.md"
output_file = skill_dir / "inventory" / "skill-frontmatter.json"
output_file.parent.mkdir(parents=True, exist_ok=True)

frontmatter = {}
in_frontmatter = False
with skill_file.open("r", encoding="utf-8") as fh:
    for line in fh:
        line = line.rstrip("\n")
        if line.strip() == "---":
            if not in_frontmatter:
                in_frontmatter = True
                continue
            else:
                break
        if in_frontmatter and ":" in line:
            key, value = line.split(":", 1)
            frontmatter[key.strip()] = value.strip()

output_file.write_text(json.dumps(frontmatter, indent=2), encoding="utf-8")
print(json.dumps(frontmatter, indent=2))
