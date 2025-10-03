# Meta-CC Claude Code Integration Guide

A comprehensive guide to choosing and using the right integration method for your meta-cognition workflow.

## Overview

meta-cc provides three integration methods with Claude Code, each designed for different use cases:

- **MCP Server**: Seamless data access through the Model Context Protocol
- **Slash Commands**: Quick, pre-defined workflows
- **Subagent (@meta-coach)**: Interactive, conversational analysis

This guide helps you understand the **core differences** between these methods and choose the **best approach** for your specific needs.

---

## Quick Comparison

| Feature | MCP Server | Slash Commands | Subagent |
|---------|-----------|----------------|----------|
| **Invocation** | Automatic (Claude decides) | Manual (`/command`) | Manual or auto-delegated (`@agent`) |
| **Context** | Main conversation | Main conversation | **Independent context** |
| **Multi-turn** | ❌ Single call | ❌ Single execution | ✅ Conversational |
| **Parameters** | Structured (JSON schema) | Positional (`$1, $2`) | Natural language |
| **Best for** | Data queries | Repeated workflows | Exploratory analysis |

👉 **[Jump to Decision Framework](#decision-framework)** to find the best method for your task.

---

## Part I: Core Differences Deep Dive

### 1.1 Context Isolation Mechanism

Understanding how each method handles conversation context is crucial for choosing the right approach.

#### Main Conversation Context (MCP & Slash Commands)

```
┌─────────────────────────────────────┐
│   Main Claude Code Conversation    │
│                                     │
│  User: "Analyze my session..."      │
│    ↓                                │
│  Claude: [calls MCP tool]           │  ← MCP executes HERE
│    ↓                                │
│  [Tool result returned]             │
│    ↓                                │
│  Claude: "Your session has..."      │
│                                     │
│  User: /meta-stats                  │  ← Slash executes HERE
│    ↓                                │
│  [Command output displayed]         │
└─────────────────────────────────────┘
         All visible in chat history
```

**Characteristics**:
- ✅ **Full history access**: Can reference previous messages
- ✅ **Context continuity**: Claude remembers earlier analysis
- ⚠️ **Context pollution**: Tool calls and outputs accumulate in history
- ⚠️ **Token consumption**: Each call adds to context length

**When this matters**:
- If you need Claude to correlate current data with earlier conversation → ✅ Good
- If you're doing many repeated queries → ⚠️ Consider impact on context

---

#### Independent Context (Subagent)

```
┌─────────────────────────────────────┐
│   Main Claude Code Conversation    │
│                                     │
│  User: "@meta-coach analyze..."     │
│    ↓                                │
│  [Delegates to subagent]            │
└─────────────────────────────────────┘
                  ↓
        ┌─────────────────────────┐
        │  @meta-coach Context    │  ← SEPARATE window
        │  (Clean slate)          │
        │                         │
        │  [Calls meta-cc tools]  │
        │  [Analyzes results]     │
        │  [Multi-turn dialogue]  │
        │                         │
        │  Returns: Summary only  │
        └─────────────────────────┘
                  ↓
┌─────────────────────────────────────┐
│   Main Conversation (continued)     │
│                                     │
│  Claude: "Here's what @meta-coach   │
│           found: [summary]"         │
└─────────────────────────────────────┘
```

**Characteristics**:
- ✅ **No pollution**: Main conversation stays focused
- ✅ **Specialized reasoning**: Dedicated context for deep analysis
- ❌ **No shared history**: Each invocation starts fresh
- ❌ **Limited continuity**: Cannot reference main conversation details

**When this matters**:
- If you need deep, multi-step analysis → ✅ Good
- If you want to keep main conversation clean → ✅ Good
- If you need to build on previous subagent sessions → ❌ Won't work (each is independent)

---

### 1.2 Invocation Mechanisms

How each method is triggered and controlled.

#### MCP: Autonomous Tool Selection

**How it works**:
```
User: "What's my session error rate?"
  ↓
Claude thinks: "User wants stats... I should use get_session_stats"
  ↓
Claude calls: mcp__meta-insight__get_session_stats(output_format: "json")
  ↓
Result: {"ErrorRate": 0.0, "ErrorCount": 0, ...}
  ↓
Claude responds: "Your error rate is 0%, with 0 errors detected."
```

**Key points**:
- Claude **decides when** to call the tool (based on conversation context)
- Claude **selects parameters** based on user intent
- User **doesn't need to know** the tool exists

**Pros**:
- ✅ Most natural user experience (just ask questions)
- ✅ No need to memorize commands
- ✅ Claude can combine multiple tool calls

**Cons**:
- ⚠️ Less predictable (Claude might not always call the tool)
- ⚠️ User has less control over parameters

---

#### Slash Commands: Explicit Execution

**How it works**:
```
User: /meta-stats
  ↓
Claude Code: Loads .claude/commands/meta-stats.md
  ↓
Executes: meta-cc parse stats --output md
  ↓
Output: [Formatted markdown table]
  ↓
Claude: [Displays output, may add commentary]
```

**Key points**:
- User **explicitly triggers** the command
- Parameters passed as **positional arguments**: `/meta-errors 30`
- Command script **controls execution** (not Claude's judgment)

**Pros**:
- ✅ Fully predictable and controllable
- ✅ Fast execution (no LLM decision overhead)
- ✅ Can include complex Bash scripts

**Cons**:
- ⚠️ User must remember command names
- ⚠️ Limited parameter flexibility (positional only)

---

#### Subagent: Delegated Conversation

**How it works**:
```
User: "@meta-coach I feel stuck, help analyze my workflow"
  ↓
Claude Code: Creates independent context for @meta-coach
  ↓
@meta-coach: "Let me gather data first..."
  ↓
@meta-coach calls: meta-cc parse stats
  ↓
@meta-coach: "I see you have 0 errors. Can you describe what feels stuck?"
  ↓
User (to @meta-coach): "I keep re-running the same tests"
  ↓
@meta-coach calls: meta-cc analyze errors --window 30
  ↓
@meta-coach: "Ah, I found the pattern. Here's what I recommend..."
```

**Key points**:
- User can **mention explicitly** (`@meta-coach`) or Claude **auto-delegates**
- Supports **multi-turn dialogue** within the subagent context
- Subagent has its own **system prompt** and reasoning style

**Pros**:
- ✅ Conversational and exploratory
- ✅ Can adapt based on user responses
- ✅ Keeps main conversation focused

**Cons**:
- ⚠️ Slower (multiple LLM turns)
- ⚠️ Each session starts fresh (no memory of previous sessions)

---

### 1.3 Execution Models

How each method processes information and generates responses.

#### MCP: Data Source

**Mental model**: "A database that Claude queries"

```
┌─────────────────────────────────────────────────────┐
│                  Claude's Reasoning                 │
│                                                     │
│  1. User wants session statistics                  │
│  2. I should call get_session_stats                │
│  3. [Calls tool, receives JSON]                    │
│  4. I'll format this into a readable summary       │
│  5. [Generates natural language response]          │
└─────────────────────────────────────────────────────┘
```

**Characteristics**:
- Raw data retrieval
- Main conversation Claude does the **interpretation**
- Single tool call per invocation
- Can be **combined** with other tools in same turn

**Example output flow**:
```json
// MCP returns (raw):
{"TurnCount": 120, "ToolCallCount": 35, "ErrorRate": 0}

// Claude interprets:
"Your session has 120 turns with 35 tool calls and a 0% error rate.
That's excellent - no errors detected!"
```

---

#### Slash Commands: Pre-Programmed Workflow

**Mental model**: "A script that does exactly what you told it to"

```
┌─────────────────────────────────────────────────────┐
│              Bash Script Execution                  │
│                                                     │
│  1. Check if meta-cc is installed                  │
│  2. Run: meta-cc parse stats --output md           │
│  3. Capture output                                 │
│  4. Display formatted markdown                     │
│  5. (Optional) Claude adds brief commentary        │
└─────────────────────────────────────────────────────┘
```

**Characteristics**:
- Pre-defined logic (written in Bash)
- Can include **multiple meta-cc commands**
- Output is **pre-formatted** (markdown/json/csv)
- Claude's role is minimal (just display + optional context)

**Example output flow**:
```bash
# Script runs:
meta-cc parse stats --output md

# Outputs directly:
| Metric | Value |
|--------|-------|
| Total Turns | 120 |
| Tool Calls | 35 |
| Error Rate | 0.0% |

# Claude adds:
"Here's your session summary. Everything looks healthy!"
```

---

#### Subagent: Independent Analyst

**Mental model**: "A specialized colleague you consult with"

```
┌─────────────────────────────────────────────────────┐
│            @meta-coach Reasoning                    │
│            (Independent Context)                    │
│                                                     │
│  System Prompt:                                     │
│  "You are a meta-cognition coach..."                │
│                                                     │
│  Turn 1: Understand user's concern                 │
│    → Ask clarifying questions                      │
│                                                     │
│  Turn 2: Gather relevant data                      │
│    → Call meta-cc parse stats                      │
│    → Call meta-cc analyze errors                   │
│                                                     │
│  Turn 3: Analyze patterns                          │
│    → Identify root causes                          │
│    → Formulate recommendations                     │
│                                                     │
│  Turn 4: Present tiered suggestions                │
│    → Immediate actions                             │
│    → Optional improvements                         │
│    → Long-term optimizations                       │
│                                                     │
│  Turn 5+: Help implement solutions                 │
│    → Offer to create Hooks/Commands                │
│    → Guide through setup                           │
└─────────────────────────────────────────────────────┘
```

**Characteristics**:
- Has **own personality** and coaching methodology
- Can **reason across multiple tool calls**
- Supports **back-and-forth dialogue**
- Returns only **high-level summary** to main conversation

**Example dialogue**:
```
User: @meta-coach I feel inefficient lately

@meta-coach: Let me analyze your recent sessions.
              [Runs meta-cc parse stats]

              I notice you're using the Read tool 45% of the time,
              and often reading the same files repeatedly.

              What are you typically looking for when you read files?

User: Usually searching for function definitions

@meta-coach: Ah! You might benefit from using Grep instead of Read
              for that. Would you like me to show you how?
```

---

## Part II: Decision Framework

### 2.1 Task Type Decision Tree

Use this flowchart to quickly identify the best integration method:

```
┌─────────────────────────────────────┐
│ What do you need to do?             │
└─────────────────┬───────────────────┘
                  │
        ┌─────────┴─────────┐
        │                   │
    [Simple       [Complex multi-step
     data query]   analysis]
        │                   │
        ├── Is it a one-time    │
        │   question?            │
        │   YES → MCP            │
        │   NO ↓                 │
        │                        │
        ├── Will you repeat      ├── Do you know exactly
        │   this often?          │   what steps to take?
        │   YES → Slash Command  │   YES → Slash Command
        │   NO → MCP             │   NO ↓
                                 │
                                 ├── Do you need guidance
                                 │   or exploration?
                                 │   YES → Subagent
                                 │   NO → MCP (multiple calls)
```

**Quick decision rules**:

1. **Just want data, ask naturally** → MCP
2. **Repeat the same workflow often** → Slash Command
3. **Don't know what's wrong, need help figuring it out** → Subagent
4. **Multi-step analysis with known steps** → Slash Command
5. **Multi-step with unknown/variable steps** → Subagent

---

### 2.2 Use Case Scenarios Matrix

| Scenario | Best Method | Why | Alternative |
|----------|-------------|-----|-------------|
| **Quick stats check** | MCP or Slash | Fast, no ceremony | Either works well |
| **Daily workflow automation** | Slash Command | Predictable, repeatable | - |
| **Debugging repeated errors** | Subagent | Needs exploration and guidance | Slash + manual interpretation |
| **Cross-project comparison** | Slash Command | Parametric (`/meta-compare $path`) | MCP (if schema supports project param) |
| **Learning workflow optimization** | Subagent | Educational, conversational | - |
| **CI/CD integration** | ❌ None | Use meta-cc CLI directly | - |
| **Ad-hoc data exploration** | MCP | Natural questions, flexible | - |
| **Implementing recommendations** | Subagent | Can create files/configs | Manual (Slash just shows info) |
| **Checking if tests improved** | Slash Command | Pre-defined comparison logic | MCP + manual tracking |
| **Understanding context pollution** | Subagent | Needs explanation + analysis | Read this doc 😉 |

---

### 2.3 Anti-Patterns (When NOT to Use)

#### ❌ Don't Use MCP When...

**1. You need exactly the same analysis every time**
```
Bad:  Relying on Claude to remember to call error analysis with window=30
Good: Create /meta-errors-30 slash command
```
Why: Claude's tool selection isn't deterministic.

**2. You need multi-step reasoning**
```
Bad:  "Use MCP to analyze errors, then suggest fixes, then help implement"
Good: Use @meta-coach (multi-turn dialogue)
```
Why: MCP is single-shot data retrieval.

**3. You're building automation/scripts**
```
Bad:  Trying to invoke MCP from external scripts
Good: Call meta-cc CLI directly
```
Why: MCP is Claude Code-specific, not programmatic API.

---

#### ❌ Don't Use Slash Commands When...

**1. The workflow isn't well-defined yet**
```
Bad:  Creating /meta-investigate before knowing what to investigate
Good: Use @meta-coach to explore first, then codify as Slash later
```
Why: Slash commands are rigid scripts.

**2. You need adaptive behavior based on results**
```
Bad:  /meta-fix-errors (can't adapt to different error types)
Good: @meta-coach can analyze errors → suggest specific fixes
```
Why: Bash scripts can't reason.

**3. You only use it once**
```
Bad:  Creating /meta-one-time-check for single use
Good: Just ask Claude → it calls MCP
```
Why: Overhead of creating command file not worth it.

---

#### ❌ Don't Use Subagent When...

**1. You just need quick data**
```
Bad:  @meta-coach what's my error rate?
Good: "What's my error rate?" (Claude uses MCP) or /meta-stats
```
Why: Subagent has overhead; main conversation Claude + MCP is faster.

**2. You need the same exact output format every time**
```
Bad:  @meta-coach give me stats (might format differently each time)
Good: /meta-stats (always same markdown table)
```
Why: Subagent responses vary based on conversation flow.

**3. You want to reference main conversation history**
```
Bad:  @meta-coach "earlier we discussed X, now analyze Y"
Good: Keep analysis in main conversation (MCP tools)
```
Why: Subagent doesn't see main conversation context.

**4. You need to track progress across multiple sessions**
```
Bad:  @meta-coach "remember last time we optimized X?"
Good: Document findings manually or use persistent storage
```
Why: Each subagent invocation is independent (no memory).

---

## Part III: Best Practices

### 3.1 Combining Integration Methods

The three methods are not mutually exclusive - they can work together!

#### Pattern 1: Slash Command → Calls MCP

**Use case**: You want a fixed workflow, but leverage MCP for data access.

**Example** (`.claude/commands/meta-health-check.md`):
```markdown
---
name: meta-health-check
allowed_tools: [Bash]
---

# Health Check Workflow

1. Get session stats via MCP
2. Check error patterns
3. Compare with healthy baseline
4. Report status

[Implementation uses MCP tools programmatically]
```

**Benefit**: Combines Slash's predictability with MCP's seamless integration.

---

#### Pattern 2: Subagent → Calls meta-cc CLI

**Use case**: Complex analysis requiring both reasoning and data.

**How @meta-coach does it**:
```bash
# In subagent context:
# Step 1: Get data
meta-cc parse stats --output json

# Step 2: Reason about the data
# (Subagent's LLM interprets JSON)

# Step 3: Ask follow-up questions to user

# Step 4: Get more specific data
meta-cc analyze errors --window 20 --output json

# Step 5: Generate recommendations
```

**Benefit**: Subagent's reasoning + meta-cc's structured data = powerful analysis.

---

#### Pattern 3: MCP as Foundation, Others as Shortcuts

**Strategy**:
1. **Start with MCP** - Let users ask naturally
2. **Identify common patterns** - What questions repeat?
3. **Create Slash Commands** for those patterns
4. **Add Subagent** when users need guidance

**Example evolution**:
```
Week 1: Users ask "show me stats" → Claude uses MCP
Week 2: Lots of users ask this → Create /meta-stats
Week 3: Users ask "why are my sessions slow?" → Create @meta-coach
```

**Benefit**: Organic growth based on actual usage patterns.

---

### 3.2 Performance Considerations

#### Minimizing Context Pollution

**Problem**: Every MCP call adds to main conversation context.

**Solutions**:

1. **Use Slash Commands for bulk operations**
   ```
   Bad:  Asking Claude to call MCP 20 times to analyze different sessions
   Good: /meta-compare-all (script loops through sessions)
   ```

2. **Use Subagent for exploratory deep dives**
   ```
   Bad:  Long back-and-forth in main conversation
   Good: @meta-coach (keeps main conversation clean)
   ```

3. **Be explicit about output format**
   ```
   Better: "Get stats as JSON" (Claude calls MCP with output_format: "json")
   Good:   /meta-stats (pre-configured for clean markdown)
   ```

---

#### Choosing Output Format

**JSON** - Best for:
- Programmatic processing
- Piping to other tools
- When you need precision

**Markdown** - Best for:
- Human readability
- Slash commands (displays nicely)
- When you want Claude to interpret

**CSV** - Best for:
- Importing to spreadsheets
- Data analysis tools
- Bulk data export

**Recommendation by method**:
- **MCP**: Use JSON (Claude interprets anyway)
- **Slash**: Use Markdown (better UX)
- **Subagent**: Use JSON (subagent reasons over it)

---

#### When to Use --output json vs --output md

**Use JSON when**:
- Feeding data to another tool
- Need exact values for calculations
- Subagent will process it

**Use Markdown when**:
- Final output to user
- Human readability is priority
- Creating reports

**Example**:
```bash
# In Slash Command:
# Step 1: Get data as JSON (for jq processing)
stats=$(meta-cc parse stats --output json)
error_rate=$(echo "$stats" | jq -r '.ErrorRate')

# Step 2: Make decision based on data
if (( $(echo "$error_rate > 5.0" | bc -l) )); then
  # Step 3: Get detailed error report as Markdown (for display)
  meta-cc analyze errors --output md
else
  echo "✅ Session is healthy (${error_rate}% error rate)"
fi
```

---

## Part IV: Real-World Case Studies

> **Note**: These case studies will be populated with actual test results. The structure below shows what will be included.

### 4.1 Case Study Template

Each case study includes:
- **Scenario**: What the user wants to accomplish
- **Test Setup**: Session characteristics
- **Execution Process**: Step-by-step what happened
- **Results**: Output and analysis
- **User Experience Score**: Rated on convenience, accuracy, completeness
- **Lessons Learned**: What worked, what didn't

---

### 4.2 Case 1: Quick Health Check

**Scenario**: Developer wants to quickly check if the current session is progressing normally.

**Question**: "How's my session doing?"

#### Method A: MCP Tools

**Execution**:
```
[To be filled after testing]
```

**Results**:
```
[Actual output and metrics]
```

**UX Score**: ⭐⭐⭐⭐⭐ (5/5)
- Speed: Fast
- Accuracy: High
- Effort: Minimal

---

#### Method B: Slash Commands

**Execution**:
```
User: /meta-stats
[Command output]
```

**Results**:
```
[Actual output and metrics]
```

**UX Score**: ⭐⭐⭐⭐⭐ (5/5)
- Speed: Fastest
- Accuracy: High
- Effort: Minimal (if you remember the command)

---

#### Method C: Subagent

**Execution**:
```
User: @meta-coach How's my session looking?
[Multi-turn dialogue]
```

**Results**:
```
[Actual conversation flow]
```

**UX Score**: ⭐⭐⭐⭐ (4/5)
- Speed: Slower (conversational)
- Accuracy: High + context
- Effort: Minimal but more verbose

**Verdict**: MCP or Slash for this simple task. Subagent is overkill but provides more context.

---

### 4.3 Case 2: Deep Error Analysis

**Scenario**: Developer suspects they're hitting the same error repeatedly but isn't sure.

**Question**: "I keep seeing errors, what's going on?"

#### Method A: MCP Tools

**Execution**:
```
[To be filled - Claude makes multiple MCP calls]
```

**Results**:
```
[Analysis of how well Claude orchestrated the investigation]
```

**UX Score**: ⭐⭐⭐⭐ (4/5)

---

#### Method B: Slash Commands

**Execution**:
```
User: /meta-errors 30
[Pre-programmed workflow runs]
```

**Results**:
```
[Script output]
```

**UX Score**: ⭐⭐⭐⭐ (4/5)

---

#### Method C: Subagent

**Execution**:
```
User: @meta-coach I keep hitting errors, help me understand what's happening
[Multi-turn exploratory analysis]
```

**Results**:
```
[Full conversation with reasoning]
```

**UX Score**: ⭐⭐⭐⭐⭐ (5/5)

**Verdict**: Subagent shines here - provides guided analysis and actionable recommendations.

---

### 4.4 Case 3: Cross-Project Comparison

**Scenario**: Developer wants to compare current session with a previous project to see if workflow improved.

**Question**: "How does this session compare to my work on NarrativeForge?"

#### Method A: MCP Tools

**Execution**:
```
[Requires MCP schema to support --project parameter, or manual comparison]
```

**Results**:
```
[Analysis of feasibility]
```

**UX Score**: ⭐⭐⭐ (3/5) - Possible but not ideal

---

#### Method B: Slash Commands

**Execution**:
```
User: /meta-compare /home/yale/work/NarrativeForge
[Script calls meta-cc with --project parameter]
```

**Results**:
```
[Side-by-side comparison output]
```

**UX Score**: ⭐⭐⭐⭐⭐ (5/5)

**Verdict**: Slash Command is perfect for this - parametric and repeatable.

---

### 4.5 Case 4: Exploratory Workflow Diagnosis

**Scenario**: Developer feels inefficient but doesn't know why.

**Question**: "Something feels off about my workflow, but I can't pinpoint it"

#### Method A: MCP Tools

**Status**: ❌ Not suitable for exploratory, open-ended investigation

---

#### Method B: Slash Commands

**Status**: ❌ Can't create a script without knowing what to look for

---

#### Method C: Subagent

**Execution**:
```
User: @meta-coach I feel like my sessions are inefficient lately, but I'm not sure why

@meta-coach: Let me investigate. First, I'll get an overview of your recent activity.
[Calls: meta-cc parse stats]

@meta-coach: I see you're averaging 120 turns per session with 35 tool calls.
Your error rate is 0%, so that's good.

Let me look at your tool usage patterns...
[Calls: meta-cc parse extract --type tools]

@meta-coach: Interesting - I notice you're using the Read tool 45% of the time,
and you often read the same files multiple times in a row.

What are you typically trying to find when you read files?

User: Usually looking for function definitions or checking how something is implemented

@meta-coach: Ah! That's the inefficiency. Instead of using Read repeatedly, you could:

1. **Immediate**: Use Grep to search for function names directly
   - Faster and more targeted
   - Example: grep -r "def myFunction" instead of reading files

2. **Optional**: Create a /find-function Slash Command
   - Combines Grep + Read for common searches
   - Would you like me to help create this?

3. **Long-term**: Consider using the Glob tool first to find relevant files
   - Then Read only the files that match

Which of these would help you most right now?
```

**UX Score**: ⭐⭐⭐⭐⭐ (5/5)

**Verdict**: **Only Subagent can handle this** - requires guided exploration and adaptive reasoning.

---

## Part V: Advanced Topics

### 5.1 Creating Custom Integrations

#### When to Create a New Slash Command

**Criteria**:
1. You run the same meta-cc command >3 times
2. The workflow has clear, fixed steps
3. You want consistent output format

**Template**:
```markdown
---
name: my-custom-check
description: [Your description]
allowed_tools: [Bash]
argument-hint: [optional]
---

#!/bin/bash

# Your custom workflow using meta-cc
meta-cc parse stats --output json | jq '.ErrorRate'

# Combine multiple commands
if [ condition ]; then
  meta-cc analyze errors --output md
fi
```

---

#### Extending @meta-coach

To customize the subagent's behavior, edit `.claude/agents/meta-coach.md`:

**Add domain-specific knowledge**:
```markdown
## Specialized Analysis for [Your Domain]

When analyzing [specific type of project]:
- Look for [specific patterns]
- Consider [domain constraints]
- Recommend [domain-specific tools]
```

**Add new meta-cc commands**:
```markdown
### Advanced Analysis

# New command you added to meta-cc
meta-cc analyze toolchains --output json
```

---

### 5.2 Troubleshooting Integration Issues

#### MCP Tools Not Being Called

**Symptoms**: You ask for stats but Claude doesn't use the MCP tool.

**Possible causes**:
1. Tool description too vague → Claude doesn't recognize relevance
2. Question too indirect → Rephrase more explicitly
3. MCP server not connected → Check `claude mcp list`

**Solution**:
```
Instead of: "How am I doing?"
Try: "Get my session statistics" or "Use meta-insight to show my stats"
```

---

#### Slash Commands Not Found

**Symptoms**: `/meta-stats` not recognized.

**Checklist**:
1. ✅ File exists at `.claude/commands/meta-stats.md`?
2. ✅ Frontmatter has `name: meta-stats`?
3. ✅ Restarted Claude Code after creating the file?

---

#### Subagent Not Understanding Context

**Symptoms**: @meta-coach seems confused about your request.

**Remember**: Subagent has independent context!

**Solution**:
```
Don't: "@meta-coach why did that error happen?"
       (subagent doesn't know which error)

Do: "@meta-coach I just got an error in my auth module.
     Can you analyze my recent errors and help debug?"
```

---

## Part VI: Quick Reference

### 6.1 Command Cheat Sheet

#### MCP Tools

```bash
# In conversation, mention these naturally:
"Get session statistics"          → get_session_stats
"Analyze error patterns"          → analyze_errors
"Show tool usage"                 → extract_tools

# Claude will translate to MCP calls automatically
```

#### Slash Commands

```bash
/meta-stats              # Session overview
/meta-errors [window]    # Error analysis (default window=20)
/meta-timeline [limit]   # Chronological tool calls (default limit=50)
/meta-compare <path>     # Compare with another project
/meta-help               # Show all commands
```

#### Subagent

```bash
@meta-coach [question]   # Start analysis conversation

# Example questions:
@meta-coach How's my workflow efficiency?
@meta-coach I'm stuck, help me analyze what's wrong
@meta-coach Compare my current session with best practices
```

---

### 6.2 Decision Quick Lookup

| I want to... | Use this |
|--------------|----------|
| Check error rate quickly | MCP (ask naturally) or `/meta-stats` |
| Analyze repeated errors | `/meta-errors 30` |
| Understand why I'm inefficient | `@meta-coach` |
| Compare two projects | `/meta-compare <path>` |
| Get help optimizing workflow | `@meta-coach` |
| See recent tool usage | MCP (ask "show my tools") or `/meta-timeline` |
| Automate daily checks | Create custom Slash Command |
| Explore unknown problem | `@meta-coach` |
| Get exact same report daily | Slash Command |

---

## Part VII: Next Steps

### For New Users

1. **Start with MCP**: Just ask questions naturally, let Claude call tools
2. **Learn Slash Commands**: Use `/meta-help` to see available commands
3. **Try @meta-coach**: When you need guidance or have complex questions

### For Advanced Users

1. **Create Custom Slash Commands**: Automate your common workflows
2. **Extend @meta-coach**: Add domain-specific analysis
3. **Combine Methods**: Use MCP + Slash + Subagent together

### Contributing

Found a better pattern? [Open an issue](https://github.com/yale/meta-cc/issues) or submit a PR with your use case!

---

## Related Documentation

- **[meta-cc README](../README.md)**: Installation and CLI reference
- **[Examples & Usage](examples-usage.md)**: Step-by-step setup guides
- **[Troubleshooting Guide](troubleshooting.md)**: Common issues and solutions
- **[Technical Proposal](proposals/meta-cognition-proposal.md)**: Architecture deep dive

---

## Appendix: Technical Details

### A.1 MCP Protocol Specifics

- **Protocol Version**: 2024-11-05
- **Transport**: stdio
- **Tool Schema**: JSON Schema Draft 7
- **Max Output**: 10,000 tokens (configurable)

### A.2 Slash Command Mechanics

- **Location**: `.claude/commands/*.md`
- **Format**: Markdown with YAML frontmatter
- **Execution**: Bash script embedded in code blocks
- **Context**: Runs in project root directory

### A.3 Subagent Architecture

- **Location**: `.claude/agents/*.md`
- **Context**: Independent conversation window
- **Model**: Can specify different model (default: claude-sonnet-4)
- **Memory**: Stateless (each invocation starts fresh)

---

*Last updated: [To be filled with actual date after testing]*

*Integration testing and case studies will be added after executing the comparison tests.*
