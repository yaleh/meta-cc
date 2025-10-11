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
  tokens: tokenize(I.toLowerCase(), /[\s\-_]+/),

  scores: [
    {capability: cap, score: score_capability(cap, tokens)}
    for cap in C.capabilities
    if score_capability(cap, tokens) > 0
  ],

  return sort_by_score(scores, descending=true)
}

score_capability :: (Capability, [Token]) → int
score_capability(cap, tokens) = {
  sum([
    3 * count_matches(tokens, cap.name.toLowerCase()),
    2 * count_matches(tokens, cap.description.toLowerCase()),
    1 * count_matches(tokens, cap.keywords.join(" ").toLowerCase()),
    1 * count_matches(tokens, cap.category.toLowerCase())
  ])
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

# Step 5: Execute single capability
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

# Step 6: Execute pipeline
execute_pipeline :: (CompositeIntent) → output
execute_pipeline(composite) = {
  say(""),
  say("## Executing Composite Pipeline"),
  say(""),

  # Order capabilities by pattern
  ordered_caps: order_by_pattern(composite.capabilities, composite.pattern),

  # Execute each capability in order
  results: [],
  for i, cap in enumerate(ordered_caps, start=1) do {
    say("### Step " + i + ": Executing `" + cap.name + "`"),
    say(""),

    try {
      result: execute(cap.name),
      results.append({capability: cap.name, status: "success", output: result}),
      say(""),
      say("✓ Capability completed"),
      say("")
    } catch error {
      say(""),
      say("❌ **Failed**: `" + cap.name + "` - " + error),
      say(""),
      if i == 1:
        say("**Aborting pipeline.**"),
        return,
      else:
        say("**Continuing with remaining capabilities...**"),
        say("")
    }
  },

  say(""),
  say("## Pipeline Complete"),
  say(""),
  say("Executed " + len(results) + " capabilities successfully."),
  say("")
}

order_by_pattern :: (ScoredCapabilities, PipelinePattern) → [Capability]
order_by_pattern(caps, pattern) = {
  if pattern.type == "data_to_viz":
    [find_by_category(caps, ¬"visualization"), find_by_category(caps, "visualization")],
  else if pattern.type == "analysis_to_guidance":
    [find_by_category(caps, ¬"guidance|coaching"), find_by_category(caps, "guidance|coaching")],
  else:
    [cap.capability for cap in caps]
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

notes:
- keyword scoring: name(+3), description(+2), keywords(+1), category(+1), threshold > 0
- composite detection: ≥2 capabilities with score ≥ max(3, best * 0.7)
- pipeline patterns: data_to_viz, analysis_to_guidance, multi_analysis, sequential
- error handling: first failure aborts, subsequent failures show partial results
- alternatives shown when close scores (≥70% or ≥3 points)

patterns:
- data_to_viz: diagnostics → visualization (e.g., meta-errors → meta-viz)
- analysis_to_guidance: assessment → coaching (e.g., meta-quality → meta-coach)
- multi_analysis: multiple diagnostics (e.g., meta-errors + meta-bugs)
- sequential: default fallback

constraints:
- transparent | discoverable | flexible | semantic | non_recursive
- user_visible ∧ explicit_composite ∧ partial_results
