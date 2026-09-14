# Testing Guide

How to test changes in meta-cc: the **two-stage acceptance workflow** for
iterating fast without weakening the gate that promotes a task.

## Two-stage acceptance workflow

`make commit` is the acceptance gate for most tasks in this repo, and it should
stay that way -- it runs the workspace checks plus the whole short-mode suite
and is what actually certifies a change. The problem is its *cost*: it takes
roughly a minute, so using it as the inner loop of an edit-compile-test cycle
means every typo and every compile error costs a full minute of wall clock.

Split the work into two stages:

| Stage | Command | When | Cost |
|-------|---------|------|------|
| 1. Scoped | `make test-scoped PKGS=./internal/<pkg>/...` | During iteration, after every edit | seconds |
| 2. Full | `make commit` | Before promoting a task to `ready`/`done` | ~60s |

Stage 1 is a **fail-fast** gate, not a substitute for stage 2: it narrows what
runs, it does not lower the bar. Stage 2 is unchanged and remains the only
thing that certifies the change.

### Stage 1: `make test-scoped`

```bash
make test-scoped PKGS=./internal/parser/...
```

The target runs two things:

1. `go test -short <PKGS>` -- the tests of the packages you actually touched.
2. `go build ./...` -- the **whole** module, deliberately not `<PKGS>`.

The second half is the point. The most common way a scoped change breaks the
build is not inside the package you edited -- it is in a *caller* somewhere
else that no longer compiles against the changed signature. Building only
`<PKGS>` would miss exactly that, so the build step is always module-wide.

`PKGS` defaults to `./...`, so a bare `make test-scoped` is a full short-mode
run. Always pass `PKGS` in practice; the default exists only so the target is
still meaningful without arguments.

Point `PKGS` at more than one package by listing them space-separated (quote
the value so the shell does not split it into separate make arguments):

```bash
make test-scoped PKGS="./internal/parser/... ./internal/analyzer/..."
```

For a quick sanity check that formatting and vet still pass, the existing
`make dev` (format + build, <10s) composes well with this target.

### Stage 2: `make commit`

```bash
make commit
```

Run it before moving a task to `ready` or `done`, and before any `make push`.
This is the gate recorded in the task's `extra.acceptance` and evaluated by the
quay acceptance gate; skipping it means the task is promoted on a stage-1
result that was never meant to certify anything.

## Why: the scoped gate is not a new idea

Early tasks in this repo did this by hand. DIR-005, DIR-006, and DIR-007 each
carried a hand-written scoped acceptance string of the form:

```yaml
acceptance: "export PATH=/usr/local/go/bin:$PATH && cd /tmp/meta-cc-worktrees/DIR-005
  && go test ./internal/mcp/executor/... ./internal/mcp/tools/... && go build ./..."
```

That ran the touched packages' tests plus a module-wide build in well under ten
seconds, and it stayed targeted. The fleet later standardized on full
`make commit` for every task, which bought uniformity but lost the fast inner
loop.

The cost of that loss shows up in the retry counts. In the gate-events log
analyzed for this change, acceptance retries concentrated heavily in a few
tasks -- DIR-056 alone ran the acceptance gate 9 times, DIR-053 5 times, and
DIR-058 4 times. Those were overwhelmingly fix-compile-rerun cycles, where the
failure was a compile error or a single failing package that a scoped check
would have surfaced in seconds instead of after another ~60s full run. See
`tasks/DIR-090.md` for the underlying analysis.

`make test-scoped` is that early pattern, mechanized once in the `Makefile`
instead of being re-derived by hand in each task's acceptance string.

## Scoped gates in per-task acceptance strings

A task's `extra.acceptance` may use the scoped target when the change is
genuinely confined to one package, which makes the acceptance run fast:

```yaml
extra:
  acceptance: "make test-scoped PKGS=\"./internal/parser/...\""
```

Two constraints apply:

- **`make commit` stays the promotion gate.** A scoped acceptance string is a
  convenience for a task whose blast radius really is one package; it is not a
  general relaxation. If the task can affect anything outside `PKGS`, its
  acceptance must be `make commit`.
- **`go build ./...` is still module-wide.** The scoped target always verifies
  the full module compiles, so a scoped acceptance string cannot certify a
  change that breaks a caller elsewhere.

## Related targets

| Target | Scope | Notes |
|--------|-------|-------|
| `make dev` | format + build | <10s, no tests |
| `make test-scoped PKGS=...` | scoped tests + full build | stage 1; seconds |
| `make test` | all packages, `-short` | no build check |
| `make commit` | workspace checks + `make test` | stage 2 promotion gate |
| `make push` | all checks + full suite + coverage + lint + build | pre-push |
| `make test-all` | full suite + coverage profile | single pass, no `-short` |

## See Also

- [Build Quality Gates](build-quality-gates.md) - What each check group covers
- [Corpus Tolerance Gate](../reference/corpus-tolerance-gate.md) - Malformed-session-file regression gate and how to register a new tool
- [Plugin Development](plugin-development.md) - Plugin-specific test workflow
- [Design Principles](../core/principles.md) - Testing protocol and failure protocol
- [Repository Structure](../reference/repository-structure.md) - Where tests live
