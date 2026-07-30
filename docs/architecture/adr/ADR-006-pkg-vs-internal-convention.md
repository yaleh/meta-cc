---
id: ADR-006
title: "pkg/ vs internal/ Directory Convention"
status: accepted
date: 2026-03-10
---
# ADR-006: pkg/ vs internal/ Directory Convention

## Status

Accepted

## Context

The meta-cc Go module uses both `pkg/` and `internal/` top-level directories. In standard Go convention:

- `internal/` packages are **private to the module** — they cannot be imported by external Go modules. Packages within the same module may import each other freely.
- `pkg/` packages are **publicly importable** — any external Go module can import them using the full module path (e.g., `github.com/yaleh/meta-cc/pkg/pipeline`).

The current layout includes:

| Directory | Packages |
|-----------|----------|
| `internal/` | `parser`, `analyzer`, `query`, `locator` |
| `pkg/` | `pipeline`, `output` |

Both `pkg/pipeline` and `pkg/output` import from `internal/parser`, which is correct direction. However, at one point `internal/query` imported `pkg/pipeline` directly to load session data. This created a **layering inversion**: a private internal package depended on a nominally public package, making the dependency graph confusing and circular in intent.

The root cause: `pkg/pipeline` and `pkg/output` are **not genuinely public APIs**. They are not intended to be consumed by external Go modules — they are only used by this module's own binaries (`cmd/mcp-server`). They ended up in `pkg/` for historical reasons rather than by deliberate design.

## Decision

1. **`internal/` is the default** for all new packages. A package belongs in `internal/` unless it is explicitly designed to be a stable, versioned public API consumable by external Go modules.

2. **`pkg/` is reserved for genuine public API surface**. Moving a package to `pkg/` is a deliberate commitment to external consumers. It implies stability guarantees that `internal/` does not.

3. **`pkg/pipeline` and `pkg/output` remain in `pkg/` for now** due to their existing location. They are not treated as public API in practice — no external module should rely on them — but renaming them is deferred as it is a low-risk, low-priority refactor.

4. **`internal/` packages must never import `pkg/` packages**. This direction of dependency is forbidden because it means a private implementation detail depends on a public (or nominally public) package, which inverts the intended layering. Any such dependency must be broken via an interface.

5. **The SessionLoader interface decouples `internal/query` from `pkg/pipeline`**. Rather than `internal/query` importing `pkg/pipeline` directly, `internal/query` defines a `SessionLoader` interface. The `pkg/pipeline.SessionPipeline` type satisfies that interface, and the dependency is injected at the call site in `cmd/`. This is the canonical fix for the layering inversion.

### Dependency Rules Summary

```
cmd/           → internal/, pkg/   (binary wires everything together)
pkg/pipeline   → internal/parser, internal/locator  (ok: pkg→internal)
pkg/output     → internal/parser   (ok: pkg→internal)
internal/query → internal/parser   (ok: internal→internal)
internal/query → pkg/pipeline      (FORBIDDEN — was fixed via SessionLoader interface)
```

Allowed dependency directions:

```
cmd/ ──────────────────┐
  │                    │
  ▼                    ▼
pkg/  ──────────► internal/
```

`internal/` packages must not depend on `pkg/` packages.

## Consequences

### Positive

- **Clear layering**: `internal/` packages are isolated from public API concerns; they can evolve freely without worrying about external consumers.
- **No circular imports**: Forbidding `internal/ → pkg/` prevents the class of import cycles that motivated this ADR.
- **Testability**: `internal/query` accepts a `SessionLoader` interface, making it straightforward to inject test doubles without depending on the full pipeline.
- **Explicit public API boundary**: New contributors know that `pkg/` is a deliberate, considered choice — not the default.

### Negative

- **Historical inconsistency**: `pkg/pipeline` and `pkg/output` are misnamed by this convention but are left in place to avoid unnecessary churn.
- **Interface overhead**: The SessionLoader interface adds a small amount of indirection that would be unnecessary if `pkg/pipeline` were in `internal/`.

### Future Work

- `pkg/pipeline` and `pkg/output` can be moved to `internal/` in a future cleanup phase. This is a mechanical rename with no behavioral change.
- If a genuine public API is ever needed (e.g., to let external tools embed meta-cc's parser), a new `pkg/` package should be designed explicitly for that purpose with stability guarantees documented.

## Related Decisions

- [ADR-001](ADR-001-two-layer-architecture.md) - Two-Layer Architecture Design (establishes the internal/external split at the system level)

## Notes

The Go specification defines `internal` package visibility at:
https://pkg.go.dev/cmd/go#hdr-Internal_Directories

The convention of placing public packages in `pkg/` is a common Go project layout pattern but is not enforced by the toolchain — only `internal/` has toolchain-enforced visibility restrictions.
