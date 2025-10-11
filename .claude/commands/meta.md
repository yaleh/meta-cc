---
name: meta
description: Unified meta-cognition command with semantic capability matching. Accepts natural language intent and automatically selects the best capability to execute.
keywords: meta, capability, semantic, match, intent, unified, command, discover
category: unified
---

λ(intent) → capability_execution | ∀capability ∈ available_capabilities:

intent :: string  # Natural language description of user intent (from $1)

# Step 1: Discover available capabilities
discover :: void → CapabilityIndex
discover() = {
  result: mcp_meta_cc.list_capabilities(),

  if result.error then
    error("Failed to load capabilities: " + result.error),

  capabilities: result.capabilities,
  source_count: result.source_count,

  log("Loaded " + len(capabilities) + " capabilities from " + source_count + " source(s)")
}

# Step 2: Semantic matching using keyword scoring
match :: (intent, CapabilityIndex) → ScoredCapabilities
match(I, C) = {
  # Tokenize intent (lowercase, split on spaces/punctuation)
  intent_tokens: tokenize(I.toLowerCase(), /[\s\-_]+/),

  # Score each capability
  scores: [],
  for cap in C.capabilities do {
    score: 0,

    # Check name match (high weight)
    for token in intent_tokens do {
      if token in cap.name.toLowerCase():
        score += 3
    },

    # Check description match (medium weight)
    for token in intent_tokens do {
      if token in cap.description.toLowerCase():
        score += 2
    },

    # Check keywords match (medium weight)
    for keyword in cap.keywords do {
      for token in intent_tokens do {
        if token in keyword.toLowerCase() || keyword.toLowerCase() in token:
          score += 1
      }
    },

    # Check category match (low weight)
    for token in intent_tokens do {
      if token in cap.category.toLowerCase():
        score += 1
    },

    if score > 0:
      scores.append({
        capability: cap,
        score: score
      })
  },

  # Sort by score (descending)
  scores.sort(key=lambda x: x.score, reverse=true),

  return scores
}

# Step 3: Composite detection
detect_composite :: (ScoredCapabilities) → CompositeIntent | null
detect_composite(scored) = {
  if len(scored) < 2:
    return null,

  best_score: scored[0].score,
  threshold: max(3, best_score * 0.7),

  # Find all candidates that meet threshold
  candidates: [cap for cap in scored if cap.score >= threshold],

  if len(candidates) < 2:
    return null,

  return {
    capabilities: candidates,
    pattern: detect_pipeline_pattern(candidates)
  }
}

# Step 4: Detect pipeline pattern
detect_pipeline_pattern :: (ScoredCapabilities) → PipelinePattern
detect_pipeline_pattern(capabilities) = {
  categories: [cap.capability.category for cap in capabilities],
  names: [cap.capability.name for cap in capabilities],

  # Pattern 1: Data → Visualization
  # Any analysis/diagnostics/assessment + visualization
  has_data_source: any(c in ["diagnostics", "analysis", "assessment", "tracking"] for c in categories),
  has_viz: "visualization" in categories,

  if has_data_source and has_viz:
    return {
      type: "data_to_viz",
      description: "Generate data, then visualize it",
      steps: [
        "Execute data generation capability",
        "Extract structured output",
        "Pass to visualization capability"
      ]
    },

  # Pattern 2: Analysis → Guidance
  # Diagnostics/assessment + guidance/coaching
  has_diagnostics: any(c in ["diagnostics", "assessment", "analysis"] for c in categories),
  has_guidance: any(c in ["guidance", "coaching"] for c in categories),

  if has_diagnostics and has_guidance:
    return {
      type: "analysis_to_guidance",
      description: "Analyze state, then provide recommendations",
      steps: [
        "Execute diagnostic capability",
        "Extract key findings",
        "Pass to guidance capability for recommendations"
      ]
    },

  # Pattern 3: Multi-Analysis
  # Multiple diagnostics/analysis capabilities
  analysis_count: sum(1 for c in categories if c in ["diagnostics", "analysis", "assessment"]),

  if analysis_count >= 2:
    return {
      type: "multi_analysis",
      description: "Execute multiple analyses in sequence",
      steps: [
        "Execute first analysis",
        "Execute second analysis",
        "Combine insights"
      ]
    },

  # Default: Sequential
  return {
    type: "sequential",
    description: "Execute capabilities in order",
    steps: [
      "Execute each capability sequentially",
      "Display all results"
    ]
  }
}

# Step 5: User confirmation for composite
confirm_composite :: (CompositeIntent) → bool
confirm_composite(composite) = {
  say(""),
  say("🔍 **Detected Composite Intent**"),
  say(""),
  say("Multiple high-scoring capabilities found:"),
  say(""),

  for i, scored_cap in enumerate(composite.capabilities, start=1) do {
    cap: scored_cap.capability,
    score: scored_cap.score,
    say(i + ". **" + cap.name + "** (score: " + score + ")"),
    say("   - " + cap.description),
    say("   - Category: `" + cap.category + "`"),
    say("")
  },

  say("**Proposed Pipeline**: `" + composite.pattern.type + "`"),
  say(""),
  say("**Description**: " + composite.pattern.description),
  say(""),
  say("**Steps**:"),
  for step in composite.pattern.steps do {
    say("  - " + step)
  },
  say(""),

  say("**Question**: Execute this composite pipeline?"),
  say(""),
  say("_(If you want only a single capability, please specify which one)_"),
  say(""),

  # Note: In actual implementation, Claude will ask user and wait for response
  # For now, we proceed with best match only (user can interrupt)
  return false  # Default: don't auto-execute composite (user must confirm)
}

# Step 6: Execute single capability
execute :: (capability_name) → output
execute(name) = {
  # Get full capability content
  result: mcp_meta_cc.get_capability(name=name),

  if result.error then
    error("Failed to get capability: " + result.error),

  content: result.content,
  source: result.source,

  # Parse frontmatter to get metadata
  frontmatter: parse_frontmatter(content),

  # Display capability info
  say(""),
  say("## Executing Capability: **" + frontmatter.name + "**"),
  say(""),
  say("**Description**: " + frontmatter.description),
  say("**Category**: " + frontmatter.category),
  say("**Source**: " + source),
  say(""),
  say("---"),
  say(""),

  # Execute the capability by interpreting its content
  # The capability is a Claude Code slash command (markdown with lambda calculus)
  # Claude will interpret the implementation and call appropriate MCP tools
  interpret_and_execute(content)
}

# Step 7: Execute pipeline
execute_pipeline :: (CompositeIntent) → output
execute_pipeline(composite) = {
  say(""),
  say("## Executing Composite Pipeline"),
  say(""),

  results: [],
  pattern: composite.pattern,

  if pattern.type == "data_to_viz" then {
    # Find data and viz capabilities
    data_cap: null,
    viz_cap: null,

    for scored_cap in composite.capabilities do {
      if scored_cap.capability.category == "visualization":
        viz_cap = scored_cap.capability,
      else:
        data_cap = scored_cap.capability
    },

    if data_cap == null or viz_cap == null then {
      error("Invalid data_to_viz pipeline: missing data or viz capability"),
      return
    },

    # Execute data capability
    say("### Step 1: Executing Data Capability"),
    say(""),
    say("Running: **" + data_cap.name + "**"),
    say(""),

    try {
      data_result: execute(data_cap.name),
      results.append({
        capability: data_cap.name,
        status: "success",
        output: data_result
      }),
      say(""),
      say("✓ Data capability completed"),
      say("")
    } catch error {
      say(""),
      say("❌ **Failed to execute** `" + data_cap.name + "`: " + error),
      say(""),
      say("**Aborting pipeline.**"),
      return
    },

    # Execute viz capability
    say("### Step 2: Executing Visualization Capability"),
    say(""),
    say("Running: **" + viz_cap.name + "**"),
    say(""),
    say("_(Visualization will use data from previous step)_"),
    say(""),

    try {
      viz_result: execute(viz_cap.name),
      results.append({
        capability: viz_cap.name,
        status: "success",
        output: viz_result
      }),
      say(""),
      say("✓ Visualization capability completed"),
      say("")
    } catch error {
      say(""),
      say("⚠️ **Warning**: `" + viz_cap.name + "` execution failed: " + error),
      say(""),
      say("**Showing partial results** from `" + data_cap.name + "`"),
      say("")
    }

  } else if pattern.type == "analysis_to_guidance" then {
    # Find analysis and guidance capabilities
    analysis_cap: composite.capabilities[0].capability,
    guidance_cap: composite.capabilities[1].capability,

    # Override if guidance is first
    if guidance_cap.category in ["guidance", "coaching"]:
      pass  # Keep order
    else if analysis_cap.category in ["guidance", "coaching"]:
      analysis_cap, guidance_cap = guidance_cap, analysis_cap,

    # Execute analysis capability
    say("### Step 1: Executing Analysis Capability"),
    say(""),
    say("Running: **" + analysis_cap.name + "**"),
    say(""),

    try {
      analysis_result: execute(analysis_cap.name),
      results.append({
        capability: analysis_cap.name,
        status: "success",
        output: analysis_result
      }),
      say(""),
      say("✓ Analysis capability completed"),
      say("")
    } catch error {
      say(""),
      say("❌ **Failed to execute** `" + analysis_cap.name + "`: " + error),
      say(""),
      say("**Aborting pipeline.**"),
      return
    },

    # Execute guidance capability
    say("### Step 2: Executing Guidance Capability"),
    say(""),
    say("Running: **" + guidance_cap.name + "**"),
    say(""),
    say("_(Guidance will be based on analysis results)_"),
    say(""),

    try {
      guidance_result: execute(guidance_cap.name),
      results.append({
        capability: guidance_cap.name,
        status: "success",
        output: guidance_result
      }),
      say(""),
      say("✓ Guidance capability completed"),
      say("")
    } catch error {
      say(""),
      say("⚠️ **Warning**: `" + guidance_cap.name + "` execution failed: " + error),
      say(""),
      say("**Showing partial results** from `" + analysis_cap.name + "`"),
      say("")
    }

  } else {
    # Sequential execution
    for i, scored_cap in enumerate(composite.capabilities, start=1) do {
      cap: scored_cap.capability,

      say("### Step " + i + ": Executing `" + cap.name + "`"),
      say(""),

      try {
        result: execute(cap.name),
        results.append({
          capability: cap.name,
          status: "success",
          output: result
        }),
        say(""),
        say("✓ Capability completed"),
        say("")
      } catch error {
        say(""),
        say("❌ **Failed to execute** `" + cap.name + "`: " + error),
        say(""),
        if i == 1:
          say("**Aborting pipeline.**"),
          return,
        else:
          say("**Continuing with remaining capabilities...**"),
          say("")
      }
    }
  },

  say(""),
  say("## Pipeline Complete"),
  say(""),
  say("Executed " + len(results) + " capabilities successfully."),
  say("")
}

# Main workflow
main :: intent → void
main(I) = {
  say("# Unified Meta Command"),
  say(""),
  say("**User Intent**: " + I),
  say(""),

  # Step 1: Discover capabilities
  say("## Step 1: Discovering Capabilities"),
  say(""),
  index: discover(),
  say("✓ Loaded " + len(index.capabilities) + " capabilities"),
  say(""),

  # Step 2: Semantic matching
  say("## Step 2: Semantic Matching"),
  say(""),
  scored: match(I, index),

  if len(scored) == 0 then {
    say("❌ **No matching capabilities found** for: `" + I + "`"),
    say(""),
    say("### Available Capabilities:"),
    say(""),

    # Group by category
    categories: group_by_category(index.capabilities),

    for category in categories.keys().sort() do {
      say("#### " + category.capitalize()),
      say(""),
      for cap in categories[category].sort_by(name) do {
        say("- **" + cap.name + "**: " + cap.description),
        say("  - Keywords: `" + join(cap.keywords, "`, `") + "`")
      },
      say("")
    },

    say("### Usage"),
    say(""),
    say("Try one of these examples:"),
    say("- `/meta \"show errors\"`"),
    say("- `/meta \"quality check\"`"),
    say("- `/meta \"visualize timeline\"`"),
    say("- `/meta \"help me improve workflow\"`"),
    say(""),

    return
  },

  # Check for composite intent
  composite: detect_composite(scored),

  if composite != null then {
    # Composite execution detected
    say("🔍 **Detected Composite Intent**"),
    say(""),
    say("Multiple high-scoring capabilities found:"),
    say(""),

    for i, scored_cap in enumerate(composite.capabilities, start=1) do {
      cap: scored_cap.capability,
      score: scored_cap.score,
      say(i + ". **" + cap.name + "** (score: " + score + ")"),
      say("   - " + cap.description),
      say("   - Category: `" + cap.category + "`"),
      say("")
    },

    say("**Proposed Pipeline**: `" + composite.pattern.type + "`"),
    say(""),
    say("**Description**: " + composite.pattern.description),
    say(""),
    say("**Steps**:"),
    for step in composite.pattern.steps do {
      say("  - " + step)
    },
    say(""),
    say("---"),
    say(""),

    # For now, proceed with best match only
    # User can interrupt and request full pipeline execution
    say("**Note**: Proceeding with best match. Reply 'execute pipeline' to run all capabilities."),
    say(""),

    best: scored[0],
    say("## Step 3: Executing Best Match"),
    say(""),
    say("🎯 **Selected**: `" + best.capability.name + "` (score: " + best.score + ")"),
    say(""),
    execute(best.capability.name)
  } else {
    # Single capability execution
    best: scored[0],

    say("🎯 **Best Match**: `" + best.capability.name + "` (score: " + best.score + ")"),
    say(""),

    # Show alternatives if multiple high scores (but not composite)
    if len(scored) > 1 and scored[1].score >= max(3, best.score * 0.7) then {
      say("**Other Possible Matches**:"),
      for i in range(1, min(4, len(scored))) do {
        alt: scored[i],
        say("- `" + alt.capability.name + "` (score: " + alt.score + ")")
      },
      say(""),
      say("*Proceeding with best match: `" + best.capability.name + "`*"),
      say("")
    },

    # Step 3: Execute capability
    say("## Step 3: Executing Capability"),
    say(""),
    execute(best.capability.name)
  }
}

# Entry point
main($1)

implementation_notes:
- semantic matching uses keyword-based scoring (no ML required)
- scoring algorithm: name (+3), description (+2), keywords (+1), category (+1)
- threshold for match: score > 0 (sorted by score)
- Claude interprets capability content and executes the instructions
- capabilities can call MCP tools, read files, analyze data, etc.
- /meta is a meta-layer that orchestrates capability discovery and execution
- no-match case lists all available capabilities grouped by category
- alternatives shown when scores are close (within 70% or ≥3 points)
- composite detection: ≥2 capabilities with score ≥ max(3, best_score * 0.7)
- pipeline patterns: data_to_viz, analysis_to_guidance, multi_analysis, sequential
- composite execution: user can request "execute pipeline" to run all capabilities
- default behavior: show composite detection, proceed with best match (user can override)
- error handling: first capability failure aborts pipeline, second+ shows partial results

composite_patterns:
- data_to_viz: diagnostics/analysis/assessment → visualization
  - Example: meta-errors → meta-viz
  - Flow: generate data, extract metrics, visualize
- analysis_to_guidance: diagnostics/assessment → guidance/coaching
  - Example: meta-quality-scan → meta-coach
  - Flow: analyze state, extract findings, provide recommendations
- multi_analysis: multiple diagnostics/analysis capabilities
  - Example: meta-errors + meta-bugs
  - Flow: run analyses in sequence, combine insights
- sequential: default fallback for any composite
  - Flow: execute each capability in order

constraints:
- transparent: all steps visible in main conversation
- discoverable: lists capabilities when no match found
- flexible: accepts any natural language intent
- semantic: keyword-based matching is simple but effective
- non_recursive: ¬execute(/meta) within capability
- user_visible: all execution happens in main conversation thread
- explicit_composite: composite detection shown, user must confirm full pipeline
- partial_results: if pipeline fails mid-execution, show what succeeded
