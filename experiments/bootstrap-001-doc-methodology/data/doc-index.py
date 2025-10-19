#!/usr/bin/env python3
import os
import json
import re
from pathlib import Path

def extract_title(content):
    """Extract first H1 heading as title"""
    match = re.search(r'^#\s+(.+)$', content, re.MULTILINE)
    return match.group(1) if match else "Untitled"

def extract_summary(content):
    """Extract first paragraph after title"""
    lines = content.split('\n')
    for i, line in enumerate(lines):
        if line.startswith('#') and i+1 < len(lines):
            for j in range(i+1, min(i+10, len(lines))):
                if lines[j].strip() and not lines[j].startswith('#'):
                    return lines[j].strip()[:150]
    return ""

def calculate_depth(path):
    """Calculate depth from docs root"""
    parts = Path(path).parts
    if 'docs' in parts:
        return len(parts) - parts.index('docs') - 1
    return 0

def categorize(path):
    """Categorize based on directory"""
    if 'guides' in path: return 'guide'
    elif 'reference' in path: return 'reference'
    elif 'tutorials' in path: return 'tutorial'
    elif 'core' in path: return 'core'
    elif 'methodology' in path: return 'methodology'
    elif 'architecture' in path: return 'architecture'
    elif 'archive' in path: return 'archive'
    else: return 'general'

def extract_keywords(content):
    """Extract important keywords"""
    # Simple keyword extraction - could be enhanced with TF-IDF
    words = re.findall(r'\b[A-Za-z]{4,}\b', content.lower())
    word_freq = {}
    for word in words:
        word_freq[word] = word_freq.get(word, 0) + 1
    # Return top 10 most frequent words
    sorted_words = sorted(word_freq.items(), key=lambda x: x[1], reverse=True)
    return [word for word, freq in sorted_words[:10]]

# Build index
index = []
docs_dir = Path('docs')

for md_file in docs_dir.rglob('*.md'):
    rel_path = md_file.relative_to(docs_dir)
    
    with open(md_file, 'r', encoding='utf-8') as f:
        content = f.read()
    
    entry = {
        'title': extract_title(content),
        'path': str(rel_path),
        'category': categorize(str(rel_path)),
        'depth': calculate_depth(str(md_file)),
        'summary': extract_summary(content),
        'keywords': extract_keywords(content),
        'size': len(content)
    }
    
    index.append(entry)

# Sort by importance (core first, then guides, then others)
category_priority = {'core': 0, 'guide': 1, 'reference': 2, 'tutorial': 3, 'methodology': 4, 'architecture': 5, 'archive': 6, 'general': 7}
index.sort(key=lambda x: (category_priority.get(x['category'], 8), x['depth'], x['title']))

# Save index
with open('experiments/bootstrap-001-doc-methodology/data/documentation-index.json', 'w') as f:
    json.dump(index, f, indent=2)

print(f"Index created with {len(index)} documents")
print(f"Average depth: {sum(d['depth'] for d in index) / len(index):.1f}")
print(f"Categories: {set(d['category'] for d in index)}")
