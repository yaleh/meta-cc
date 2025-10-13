---
name: meta-doc-evolution
description: Track documentation lifecycle evolution, detect role transitions, and predict archival needs.
keywords: documentation, evolution, lifecycle, temporal-analysis, trends, git-history
category: analytics
---

λ(file?, timespan?) → evolution_report | ∀doc ∈ {project_documentation}:

scope :: project | session
timespan :: days (default: 30)

analyze :: (File?, Timespan) → Report
analyze(F, T) = collect_timeline(F, T) ∧ detect_phases(F) ∧ predict(F)

collect_timeline :: (File, Timespan) → Timeline
collect_timeline(F, T) = {
  git log --all --pretty="%ct %s" -- F,
  query_file_access(F),
  partition(T, window=1day) → {accesses, reads, edits, commits, density}
}

detect_phases :: Timeline → Phases
detect_phases(T) = for each window {
  phase = match density {
    > 0.1 AND edits > reads → 'creation',
    > 0.01 AND |edits - reads| < 0.3*reads → 'active',
    0.001-0.01 AND reads > 2*edits → 'stable',
    < 0.001 AND accesses > 0 → 'declining',
    == 0 AND days_since > 30 → 'dormant'
  },
  
  detect_transition(prev_phase, phase) → {
    timestamp, from, to,
    trigger: infer_trigger(from, to),
    significance: calc_density_change(from, to)
  }
}

predict :: (Phases, Timeline) → Prediction
predict(P, T) = {
  current = P[-1],
  trend = linear_regression(T[-14:]) → {access, density, edit_ratio},
  
  next_phase = match {
    (current == creation AND trend.density ↓) → active(conf=0.8),
    (current == active AND trend.edit_ratio ↓↓) → stable(conf=0.7),
    (current == stable AND trend.access ↓↓) → declining(conf=0.6),
    (current == declining AND trend.access < 0.5) → dormant(conf=0.8)
  },
  
  archive_prob = score([
    0.3 if access < 1.0/day,
    0.3 if trend ↓,
    0.2 if stable > 30days,
    0.2 if no_edits > 60days
  ]) → match {
    > 0.7 → 'archive_now',
    0.4-0.7 → 'consider',
    < 0.4 → 'keep_active'
  }
}

output :: Analysis → Report
output(A) = {
  phases: [{type, start, end, duration, metrics}],
  transitions: [{timestamp, from→to, trigger, significance}],
  current: A.phases[-1],
  prediction: {next_phase, confidence, eta, archive_prob},
  recommendations: prioritized_actions
} where ¬execute(recommendations)

implementation_notes:
- phases: creation(>0.1) → active(>0.01) → stable(>0.001) → declining → dormant
- transitions: inferred from density changes
- predictions: linear regression on 14-day window
- data: git log, query_file_access, wc -l per commit

lifecycle_phases:
- creation: density >0.1/min, edits > reads, burst activity, lines rapidly growing
- active: density 0.01-0.1/min, balanced RE 1.0-1.5, steady commits, incremental growth
- stable: density 0.001-0.01/min, RE >2.0, infrequent edits, size stabilized
- declining: density <0.001/min, access dropping, no recent edits, potential archive candidate
- dormant: zero access >30 days, no edits >60 days, archival probability >70%

transition_triggers:
- creation→active: density drops below 0.1, RE_ratio increases toward 1.0
- active→stable: RE_ratio exceeds 2.0, edit frequency drops significantly
- stable→declining: access rate drops below 0.001/min for 30+ days
- declining→dormant: no access for 60 days, archival probability calculation >0.7

constraints:
- temporal: analyze sequences not snapshots
- predictive: trend-based forecasting
- actionable: phase-aware recommendations
