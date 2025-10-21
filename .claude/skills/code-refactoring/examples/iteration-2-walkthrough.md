# Iteration 2 Walkthrough

1. **Baseline tests** — Added 5 characterization tests for `calculateSequenceTimeSpan`; coverage lifted from 85% → 100%.
2. **Extract collectOccurrenceTimestamps** — Removed timestamp gathering loop (complexity 10 → 6) while maintaining green tests.
3. **Extract findMinMaxTimestamps** — Split min/max computation; additional unit tests locked behaviour (complexity 6 → 3).
4. **Quality outcome** — Complexity −70%, package coverage 92% → 94%, three commits (≤50 lines) all green.
