#!/usr/bin/env python3
"""
validate-links.py - Validate markdown links in documentation

Usage:
    ./validate-links.py [file.md]              # Check one file
    ./validate-links.py [directory]            # Check all .md files

Exit codes:
    0 - All links valid
    1 - One or more broken links found
"""

import os
import re
import sys
from pathlib import Path

# Colors
RED = '\033[0;31m'
GREEN = '\033[0;32m'
YELLOW = '\033[1;33m'
NC = '\033[0m'

# Counters
total_links = 0
valid_links = 0
broken_links = 0
broken_list = []


def heading_to_anchor(heading):
    """Convert heading text to GitHub-style anchor"""
    # Remove markdown formatting
    heading = re.sub(r'[`*_]', '', heading)
    # Lowercase and replace spaces with hyphens
    anchor = heading.lower().replace(' ', '-')
    # Remove non-alphanumeric except hyphens
    anchor = re.sub(r'[^a-z0-9-]', '', anchor)
    return anchor


def check_anchor(file_path, anchor):
    """Check if anchor exists in file"""
    # Remove leading #
    anchor = anchor.lstrip('#')

    with open(file_path, 'r', encoding='utf-8') as f:
        for line in f:
            # Match heading lines
            match = re.match(r'^(#+)\s+(.+)$', line)
            if match:
                heading_text = match.group(2).strip()
                heading_anchor = heading_to_anchor(heading_text)
                if heading_anchor == anchor.lower():
                    return True
    return False


def validate_link(file_path, link_text, link_url):
    """Validate a single link"""
    global total_links, valid_links, broken_links

    total_links += 1

    # Skip external links
    if link_url.startswith(('http://', 'https://')):
        valid_links += 1
        return True

    # Handle anchor-only links
    if link_url.startswith('#'):
        if check_anchor(file_path, link_url):
            valid_links += 1
            return True
        else:
            broken_links += 1
            broken_list.append(f"{file_path}: [{link_text}]({link_url}) - Anchor not found")
            return False

    # Handle file links (with or without anchor)
    link_file = link_url
    link_anchor = None
    if '#' in link_url:
        link_file, link_anchor = link_url.split('#', 1)
        link_anchor = '#' + link_anchor

    # Resolve relative path
    current_dir = os.path.dirname(file_path)
    if link_file.startswith('/'):
        # Absolute path from repo root (not supported in this simple version)
        resolved_path = link_file
    else:
        # Relative path
        resolved_path = os.path.join(current_dir, link_file)

    # Normalize path
    resolved_path = os.path.normpath(resolved_path)

    # Check file exists
    if not os.path.isfile(resolved_path):
        broken_links += 1
        broken_list.append(f"{file_path}: [{link_text}]({link_url}) - File not found: {resolved_path}")
        return False

    # Check anchor if present
    if link_anchor:
        if check_anchor(resolved_path, link_anchor):
            valid_links += 1
            return True
        else:
            broken_links += 1
            broken_list.append(f"{file_path}: [{link_text}]({link_url}) - Anchor not found in {resolved_path}")
            return False

    valid_links += 1
    return True


def validate_file(file_path):
    """Validate all links in a markdown file"""
    print(f"{YELLOW}Checking:{NC} {file_path}")

    with open(file_path, 'r', encoding='utf-8') as f:
        content = f.read()

    # Find all markdown links: [text](url)
    link_pattern = r'\[([^\]]+)\]\(([^)]+)\)'
    for match in re.finditer(link_pattern, content):
        link_text = match.group(1)
        link_url = match.group(2)
        validate_link(file_path, link_text, link_url)


def main():
    """Main function"""
    if len(sys.argv) < 2:
        target = '.'
    else:
        target = sys.argv[1]

    print(f"{YELLOW}Link Validation Tool{NC}")
    print("====================")
    print("")

    target_path = Path(target)

    if not target_path.exists():
        print(f"{RED}Error:{NC} {target} not found")
        sys.exit(2)

    if target_path.is_file():
        if target_path.suffix != '.md':
            print(f"{RED}Error:{NC} Not a markdown file: {target}")
            sys.exit(2)
        validate_file(str(target_path))
    elif target_path.is_dir():
        for md_file in target_path.rglob('*.md'):
            validate_file(str(md_file))
    else:
        print(f"{RED}Error:{NC} {target} is neither a file nor directory")
        sys.exit(2)

    # Summary
    print("")
    print("====================")
    print(f"{YELLOW}Summary{NC}")
    print("====================")
    print(f"Total links: {total_links}")
    print(f"{GREEN}Valid:{NC} {valid_links}")
    print(f"{RED}Broken:{NC} {broken_links}")

    if broken_links > 0:
        print("")
        print("Details:")
        for broken in broken_list:
            print(f"{RED}  ✗{NC} {broken}")
        sys.exit(1)
    else:
        print(f"{GREEN}✓ All links valid!{NC}")
        sys.exit(0)


if __name__ == '__main__':
    main()
