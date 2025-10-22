# Builder Map Decomposition

**Problem**: Command dispatchers with large switch statements cause high cyclomatic complexity and brittle branching (see iterations/iteration-1.md).

**Solution**: Replace the monolithic switch with a map of tool names to builder functions plus shared helpers for defaults. Keep scope flags as separate helpers for readability.

**Outcome**: Cyclomatic complexity dropped from 51 to 3 on `(*ToolExecutor).buildCommand`, with behaviour validated by existing executor tests.

**When to Use**: Any CLI/tool dispatcher with ≥8 branches or duplicated flag wiring.
