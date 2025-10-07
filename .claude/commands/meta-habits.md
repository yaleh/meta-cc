---
name: meta-habits
description: Analyze user's work habits and patterns using MCP meta-insight. Examines prompt types, work rhythm, tool preferences, git activity, and task completion rates to provide personalized productivity insights.
---

λ(scope) → user_behavior_insights | ∀habit ∈ {prompt_patterns, tool_usage, workflow_sequences}:

scope :: project | session

analyze :: UserMessages → Behavior_Insights
analyze(U) = collect(prompts) ∧ extract(features) ∧ detect(sequences) ∧ characterize(style)

collect :: Scope → UserPrompts
collect(S) = {
  all_prompts: mcp_meta_insight.query_user_messages(
    pattern=".*",
    limit=200,
    scope=scope
  ),

  successful_patterns: mcp_meta_insight.query_successful_prompts(
    min_quality_score=0.8,
    limit=30,
    scope=scope
  )
}

extract :: UserPrompts → PromptFeatures
extract(U) = {
  // 统计包含引用的消息数量（不是引用符号的总出现次数）
  file_references: count(U.all_prompts | filter("@[a-zA-Z0-9/_.-]+")),
  // Implementation: jq -r '.content' | grep -c '@'

  subagent_usage: count(U.all_prompts | filter("@agent-")),
  // Implementation: jq -r '.content' | grep -c '@agent-'

  slash_commands: count(U.all_prompts | filter("^/")),
  // Implementation: jq -r '.content' | grep -c '^/'

  multi_file_refs: count(U.all_prompts | filter("@.*@")),
  // Implementation: jq -r '.content' | grep -c '@.*@'

  doc_refs: count(U.all_prompts | filter("@docs|@plans|@CLAUDE")),
  // Implementation: jq -r '.content' | grep -cE '@docs|@plans|@CLAUDE'

  specific_locations: count(U.all_prompts | filter("Lines? \\d+|:\\d+")),
  // Implementation: jq -r '.content' | grep -cE 'Lines? [0-9]+|:[0-9]+'
}

classify :: UserPrompts → PromptTypes
classify(U) = {
  planning: filter(U.all_prompts,
    pattern="计划|Phase.*plan|设计|规划|分析.*如何|建议|思考"
  ),

  execution: filter(U.all_prompts,
    pattern="执行|实现|implement|修改|fix|create|add|apply|do"
  ),

  verification: filter(U.all_prompts,
    pattern="测试|验证|test|check|运行|确认|validate"
  ),

  documentation: filter(U.all_prompts,
    pattern="文档|doc|README|注释|comment|update.*md"
  ),

  meta_reflection: filter(U.all_prompts,
    pattern="/meta-|@agent-meta-coach|分析.*习惯|优化.*prompt"
  )
}

detect_sequences :: UserPrompts → PromptSequences
detect_sequences(U) = {
  // 用户工作流序列
  typical_workflows: identify([
    planning → execution → verification,
    question → clarification → execution,
    error → debug → fix → verify
  ]),

  // 用户反馈信号
  positive_feedback: filter(U.all_prompts,
    pattern="好的|继续|下一个|apply|执行.*above|perfect|excellent"
  ),

  negative_feedback: filter(U.all_prompts,
    pattern="修改|重新|不对|fix|change|incorrect|wrong"
  ),

  clarification: filter(U.all_prompts,
    pattern="为什么|如何|能否|可以.*吗|what|why|how|can you"
  ),

  interruptions: filter(U.all_prompts,
    pattern="^/clear$|\\[Request interrupted"
  ),

  // Prompt 改进迭代
  refinement_cycles: detect([
    vague_prompt → clarification → refined_prompt
  ])
}

measure_tool_adoption :: PromptFeatures → ToolUsageMetrics
measure_tool_adoption(F) = {
  file_ref_rate: F.file_references / total_prompts,

  subagent_adoption: F.subagent_usage / total_prompts,

  slash_cmd_adoption: F.slash_commands / total_prompts,

  advanced_refs: F.specific_locations / F.file_references,  // 精确引用比例

  doc_awareness: F.doc_refs / F.file_references,  // 文档引用比例
}

analyze_workflow :: PromptSequences → WorkflowInsights
analyze_workflow(S) = {
  completion_rate: count(S.typical_workflows.complete) / count(S.typical_workflows.started),

  feedback_ratio: {
    positive: count(S.positive_feedback) / total_prompts,
    negative: count(S.negative_feedback) / total_prompts,
    clarification: count(S.clarification) / total_prompts
  },

  interruption_rate: count(S.interruptions) / total_prompts,

  iteration_efficiency: {
    one_shot_success: prompts_with_positive_feedback / total_execution_prompts,
    avg_refinement_cycles: avg(S.refinement_cycles.length)
  }
}

characterize :: (PromptTypes, ToolUsageMetrics, WorkflowInsights) → UserStyle
characterize(P, T, W) = {
  communication_style: classify({
    "结构化型": T.file_ref_rate > 0.5 ∧ T.advanced_refs > 0.1,
    "探索型": W.feedback_ratio.clarification > 0.2,
    "直接型": W.iteration_efficiency.one_shot_success > 0.6,
    "协作型": T.subagent_adoption > 0.15
  }),

  planning_style: classify({
    "文档驱动": T.doc_awareness > 0.3,
    "增量开发": P.planning < P.execution ∧ W.completion_rate > 0.7,
    "瀑布式": P.planning > 0.3 ∧ planning_precedes_execution
  }),

  tool_proficiency: classify({
    "专家级用户": T.file_ref_rate > 0.6 ∧ T.subagent_adoption > 0.2,
    "高级用户": T.file_ref_rate > 0.4 ∧ T.subagent_adoption > 0.1,
    "熟练用户": T.file_ref_rate > 0.2,
    "新手": T.file_ref_rate < 0.2 ∧ T.subagent_adoption < 0.05
  })
}

output :: Analysis → Report
output(A) = {
  summary: {
    total_prompts: count(A.all_prompts),
    communication_style: A.user_style.communication_style,
    planning_style: A.user_style.planning_style,
    tool_proficiency: A.user_style.tool_proficiency
  },

  prompt_patterns: {
    type_distribution: percentage(A.prompt_types),
    typical_sequences: top_5(A.sequences.typical_workflows),
    refinement_efficiency: A.workflow.iteration_efficiency
  },

  tool_usage: {
    file_reference_rate: A.tool_metrics.file_ref_rate,
    subagent_adoption: A.tool_metrics.subagent_adoption,
    slash_command_favorites: frequency(A.slash_commands),
    advanced_techniques: {
      specific_line_refs: A.tool_metrics.advanced_refs,
      multi_file_context: A.features.multi_file_refs / total_prompts
    }
  },

  workflow_health: {
    completion_rate: A.workflow.completion_rate,
    positive_feedback_rate: A.workflow.feedback_ratio.positive,
    interruption_rate: A.workflow.interruption_rate,
    one_shot_success_rate: A.workflow.iteration_efficiency.one_shot_success
  },

  recommendations: {
    strengths: identify_from(A.user_style, A.workflow),
    improvements: suggest_from(A.tool_metrics, A.successful_patterns),
    next_level_techniques: recommend_advanced(A.tool_proficiency)
  }
} where ¬execute(recommendations)

constraints:
- user_focused: analyze user behavior, not Claude behavior
- evidence_based: ∀insight → ∃data_point
- actionable: recommendations → concrete ∧ implementable
- privacy_aware: aggregate statistics only, no sensitive data exposure
- non_judgmental: descriptive ∧ ¬prescriptive
- comprehensive: cover 5 dimensions (prompts, rhythm, tools, completion, success)
- correct_counting: use grep -c for message count, not grep -o | wc -l for occurrence count
