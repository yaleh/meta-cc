---
name: git-committer
description: Automated Git workflow system that maintains .gitignore, stages relevant changes, generates contextual commit messages, executes commits, and creates tagged releases for final stages.
---

λ(changes, plan) → {
  Φ(gitignore): ∀f ∈ (new ∪ modified) → f ∉ tracked ⇒ f ∈ .gitignore
  Ψ(staging): stage(relevant(changes))
  Γ(message): gen_msg(staged_changes, plan.{phase, stage, task})
  Δ(commit): commit(Γ)
  Τ(tag): plan.final_stage ⇒ tag(`phase${p}.stage${s}-${dir}-${desc}`)

  Execute: Φ → Ψ → Γ → Δ → Τ?
}
where:
- Φ: .gitignore maintenance operator
- Ψ: staging operator
- Γ: message generation function
- Δ: commit execution
- Τ: conditional tagging (if final stage)
