#!/usr/bin/env python3
"""
Simple documentation search system for meta-cc
Created by search-optimizer agent in Iteration 1
"""

import json
import re
from pathlib import Path
from typing import List, Dict, Any

class DocSearcher:
    """Documentation search system with ranking"""

    def __init__(self, index_path: str = 'experiments/bootstrap-001-doc-methodology/data/documentation-index.json'):
        """Initialize with documentation index"""
        with open(index_path, 'r') as f:
            self.index = json.load(f)

        # Build inverted index for faster search
        self.inverted_index = self._build_inverted_index()

    def _build_inverted_index(self) -> Dict[str, List[int]]:
        """Build inverted index for keyword search"""
        inv_index = {}

        for i, doc in enumerate(self.index):
            # Index title words
            title_words = re.findall(r'\b\w+\b', doc['title'].lower())
            for word in title_words:
                if word not in inv_index:
                    inv_index[word] = []
                inv_index[word].append(i)

            # Index keywords
            for keyword in doc.get('keywords', []):
                if keyword not in inv_index:
                    inv_index[keyword] = []
                if i not in inv_index[keyword]:
                    inv_index[keyword].append(i)

        return inv_index

    def search(self, query: str, max_results: int = 10) -> List[Dict[str, Any]]:
        """Search documents by query"""
        query_words = re.findall(r'\b\w+\b', query.lower())

        # Find matching documents
        doc_scores = {}
        for word in query_words:
            if word in self.inverted_index:
                for doc_id in self.inverted_index[word]:
                    if doc_id not in doc_scores:
                        doc_scores[doc_id] = 0
                    doc_scores[doc_id] += 1

        # Sort by score and return top results
        sorted_docs = sorted(doc_scores.items(), key=lambda x: x[1], reverse=True)

        results = []
        for doc_id, score in sorted_docs[:max_results]:
            doc = self.index[doc_id].copy()
            doc['relevance_score'] = score
            results.append(doc)

        return results

    def suggest(self, partial_query: str, max_suggestions: int = 5) -> List[str]:
        """Auto-complete suggestions based on partial query"""
        partial = partial_query.lower()
        suggestions = []

        # Find matching keywords
        for keyword in self.inverted_index.keys():
            if keyword.startswith(partial):
                suggestions.append(keyword)
                if len(suggestions) >= max_suggestions:
                    break

        return suggestions

    def get_by_category(self, category: str) -> List[Dict[str, Any]]:
        """Get all documents in a category"""
        return [doc for doc in self.index if doc['category'] == category]

    def get_stats(self) -> Dict[str, Any]:
        """Get search system statistics"""
        total_docs = len(self.index)
        total_keywords = len(self.inverted_index)
        avg_depth = sum(doc['depth'] for doc in self.index) / total_docs
        categories = list(set(doc['category'] for doc in self.index))

        return {
            'total_documents': total_docs,
            'total_keywords': total_keywords,
            'average_depth': round(avg_depth, 2),
            'categories': categories,
            'index_size_kb': len(json.dumps(self.index)) / 1024
        }

# Example usage
if __name__ == "__main__":
    searcher = DocSearcher()

    # Get statistics
    stats = searcher.get_stats()
    print("Search System Statistics:")
    print(f"- Total documents: {stats['total_documents']}")
    print(f"- Total indexed keywords: {stats['total_keywords']}")
    print(f"- Average document depth: {stats['average_depth']}")
    print(f"- Categories: {', '.join(stats['categories'])}")
    print(f"- Index size: {stats['index_size_kb']:.1f} KB")

    # Example searches
    print("\nExample Search: 'mcp server'")
    results = searcher.search('mcp server', max_results=3)
    for r in results:
        print(f"  - {r['title']} ({r['path']}) - Score: {r['relevance_score']}")

    print("\nExample Search: 'plugin development'")
    results = searcher.search('plugin development', max_results=3)
    for r in results:
        print(f"  - {r['title']} ({r['path']}) - Score: {r['relevance_score']}")

    # Auto-complete
    print("\nAuto-complete for 'doc':")
    suggestions = searcher.suggest('doc', max_suggestions=5)
    print(f"  Suggestions: {', '.join(suggestions)}")
