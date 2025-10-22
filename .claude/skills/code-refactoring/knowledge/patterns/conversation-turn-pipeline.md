# Conversation Turn Pipeline

**Problem**: Conversation queries bundled user/assistant extraction, duration math, and output assembly into one 80+ line function, inflating cyclomatic complexity (25) and risking regressions when adding filters.

**Solution**: Extract helpers for user indexing, assistant metrics, turn collection, and timestamp finalization. Each step focuses on a single responsibility, enabling targeted unit tests and reuse across similar commands.

**Evidence**: `cmd/query_conversation.go` (CLI iteration-3) reduced `buildConversationTurns` to a coordinator with helper functions ≤6 complexity.

**When to Use**: Any CLI/API that pairs multi-role messages into aggregate records (e.g., chat analytics, ticket conversations) where duplicating loops would obscure business rules.
