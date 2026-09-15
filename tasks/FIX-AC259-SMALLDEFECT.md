---
id: FIX-AC259-SMALLDEFECT
title: AC-258 user-scope 驱动取证任务
status: ready
labels: []
parent: null
children: []
extra: {}
---
## Proposal

**Defect.** The DIR-078 documentation-contract gate in
`internal/release/doc_contract_test.go` enumerates its file set in
`currentMarkdownFiles` (line 103), and the repo-root part of that enumeration is a
hard-coded two-element list at line 126:

```go
for _, top := range []string{"README.md", "CLAUDE.md"} {
```

So every other top-level markdown page — `CONTRIBUTING.md`, `AGENTS.md`,
`SECURITY.md`, `TODO.md`, `CODE_OF_CONDUCT.md`, `CHANGELOG.md` — is **exempt**
from all three checks the gate exists to enforce. `CONTRIBUTING.md` is the one
repository file that still carries the defect class the gate names:
`TestCurrentDocsGoVersionAtLeastBaseline` (line 393) exists to fail when "a
documented Go prerequisite is below the go.mod baseline", and `CONTRIBUTING.md`
documents one. The gate is therefore green while the defect is live, in the one
file it is not looking at — the same shape as `FIX-MCP-SCANNER` (a gate whose own
scan exempts the file that carries the violation it names), reached by a
different mechanism: there an explicit `grep -v "main\.go"`, here a file-set
enumeration that stops at two names.

The target is reached from the repo's own commit path: `make check-docs`
(Makefile line 280) runs `go test -short ./internal/release/...`.

**Measured before-readings (all run 2026-09-15 on `ssh orangevps`,
`/home/yale/work/meta-cc`, branch `main`, tree clean apart from pre-existing dirt).**

1. The baseline and the stale claim disagree:

   ```
   $ grep -n "^go " go.mod
   3:go 1.24.0
   $ grep -n "Go 1.2" CONTRIBUTING.md
   18:- Go 1.21 or later
   ```

2. The gate reports success anyway:

   ```
   $ make check-docs
   === Documentation Contract Check (DIR-078) ===
   ok  \tgithub.com/yaleh/meta-cc/internal/release\t0.050s

   $ go test ./internal/release/... -run TestCurrentDocsGoVersionAtLeastBaseline -v
   === RUN   TestCurrentDocsGoVersionAtLeastBaseline
   --- PASS: TestCurrentDocsGoVersionAtLeastBaseline (0.01s)
   PASS
   ok  \tgithub.com/yaleh/meta-cc/internal/release\t0.040s
   ```

3. The gate would fire on that exact line if the file were in its set. The
   predicate `belowBaselineGoVersions` + `goVersionRe` were copied **verbatim**
   (a test-only, unexported helper cannot be called from outside the package) and
   run against every top-level `*.md`:

   ```
   CONTRIBUTING.md: 1 below-baseline hit(s) [{line:18 version:1.21}]
   AGENTS.md: 0    CHANGELOG.md: 0    CLAUDE.md: 0    CODE_OF_CONDUCT.md: 0
   README.md: 0    SECURITY.md: 0     TODO.md: 0      lambda-expression-rewrite-report.md: 0
   ```

4. **Positive control** proving the only difference is the file set — run in a
   disposable clone (`git clone --local /home/yale/work/meta-cc /tmp/ac259ctrl`),
   never in the live checkout. Appending the identical claim to a page the gate
   *does* cover:

   ```
   $ printf '\n- Go 1.21 or later\n' >> docs/reference/features.md
   $ go test ./internal/release/... -run TestCurrentDocsGoVersionAtLeastBaseline
   --- FAIL: TestCurrentDocsGoVersionAtLeastBaseline (0.01s)
       doc_contract_test.go:402: stale Go prerequisite: docs/reference/features.md:140
       documents Go 1.21, below the go.mod baseline Go 1.24.
   ```

5. **The fix path is verified, not assumed.** In the same disposable clone, (a)
   widening `currentMarkdownFiles` to read every top-level `*.md` makes the gate
   fail on the real defect:

   ```
   --- FAIL: TestCurrentDocsGoVersionAtLeastBaseline (0.01s)
       doc_contract_test.go:407: stale Go prerequisite: CONTRIBUTING.md:18 documents Go 1.21,
       below the go.mod baseline Go 1.24.
   ```

   and (b) after correcting the `CONTRIBUTING.md` line, `go test ./internal/release/...`
   is `ok` again. Widening the file set costs exactly one new finding and zero
   removed-tool findings (the fence-scanner was copied verbatim too and reports 0
   hits for every top-level page, including `CHANGELOG.md`, whose
   `query_summaries(...)` mention at line 415 is prose outside a code fence).

6. No other gate covers `CONTRIBUTING.md`:
   `grep -rn "CONTRIBUTING" internal/release/*.go scripts/ Makefile .github/workflows/*.yml`
   returns nothing.

**User-visible impact.** `CONTRIBUTING.md:18` tells a new contributor to install
Go 1.21, while `go.mod` declares `go 1.24.0`; their first `go build` / `make commit`
fails with a toolchain-too-old error that the repo's own documentation gate was
built to prevent. The line has not been touched since `CONTRIBUTING.md` was added
(`5bb925c`, 2025-10-08), i.e. it went stale silently when the baseline moved.

**Not the same defect as any open task.** `tasks/FIX-MCP-SCANNER.md` (done) is the
`bufio.NewScanner` frame cap; `tasks/DIR-078.md` (done) *created* this gate and
describes exactly this defect class, but its scope manifest never included the
top-level pages; `grep -rln "CONTRIBUTING" tasks/` is empty.

**Scope.** Two non-task files change: the gate (`internal/release/doc_contract_test.go`)
and the defective artifact it should have been checking (`CONTRIBUTING.md`).

## Plan

1. **Add the failing test first.** Append to `internal/release/doc_contract_test.go`
   a regression that locks the file set, in the same synthetic-`t.TempDir()` style
   as `TestBrokenRelativeLinks`:

   ```go
   // TestCurrentMarkdownFilesCoversTopLevelDocs is the regression for the DIR-078
   // scope gap: the gate enumerated only README.md and CLAUDE.md at the repo root,
   // so every other top-level current page (CONTRIBUTING.md, AGENTS.md, ...) was
   // exempt from the checks the gate exists to enforce -- including the
   // Go-prerequisite-vs-go.mod baseline check that CONTRIBUTING.md:18 violates.
   func TestCurrentMarkdownFilesCoversTopLevelDocs(t *testing.T) {
   \troot := t.TempDir()
   \tif err := os.MkdirAll(filepath.Join(root, "docs"), 0o755); err != nil {
   \t\tt.Fatal(err)
   \t}
   \tif err := os.WriteFile(filepath.Join(root, "docs", "guide.md"), []byte("# guide\n"), 0o644); err != nil {
   \t\tt.Fatal(err)
   \t}
   \tfor _, top := range []string{"README.md", "CLAUDE.md", "CONTRIBUTING.md"} {
   \t\tif err := os.WriteFile(filepath.Join(root, top), []byte("# "+top+"\n"), 0o644); err != nil {
   \t\t\tt.Fatal(err)
   \t\t}
   \t}

   \tgot := currentMarkdownFiles(t, root)
   \tindex := make(map[string]bool, len(got))
   \tfor _, g := range got {
   \t\tindex[g] = true
   \t}
   \tif !index["CONTRIBUTING.md"] {
   \t\tt.Errorf("CONTRIBUTING.md missing from the doc-contract file set %v: "+
   \t\t\t"the gate exempts the top-level pages that carry its defect class", got)
   \t}
   }
   ```

   Run it now, before touching production code, and capture the output — it must
   fail with the exempted set printed:
   `CONTRIBUTING.md missing from the doc-contract file set [docs/guide.md README.md CLAUDE.md]`.

2. **Widen the file set (the fix).** Replace the `[]string{"README.md", "CLAUDE.md"}`
   loop in `currentMarkdownFiles` with an enumeration of every top-level `*.md`,
   keeping the existing `classifyDocScope(...) == scopeCurrent` filter so the
   allowlist / historical-prefix machinery still governs scope, and keeping the
   existing "no current markdown files resolved" fail-closed guard:

   ```go
   tops, dirErr := os.ReadDir(root)
   if dirErr != nil {
   \tt.Fatalf("reading repo root: %v", dirErr)
   }
   for _, e := range tops {
   \tif e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
   \t\tcontinue
   \t}
   \tif s, _ := classifyDocScope(e.Name()); s == scopeCurrent {
   \t\trel = append(rel, e.Name())
   \t}
   }
   ```

3. **Let the widened gate expose the real defect, then fix the artifact.**
   `go test ./internal/release/...` must now fail with
   `stale Go prerequisite: CONTRIBUTING.md:18 documents Go 1.21, below the go.mod baseline Go 1.24.`
   Correct `CONTRIBUTING.md:18` to mirror `README.md:244`:

   ```
   - Go 1.24 or later (matches the `go` directive in `go.mod`)
   ```

4. **Keep the gate honest (negative control).** Append a below-baseline claim to an
   already-covered page and confirm the gate still fires; then revert:

   ```
   printf '\n- Go 1.21 or later\n' >> docs/reference/features.md
   go test ./internal/release/... -run TestCurrentDocsGoVersionAtLeastBaseline   # must FAIL
   git checkout -- docs/reference/features.md
   ```

5. **Verify.** `make check-docs`, then `go test ./internal/release/...`, then
   `make commit` (= `normalize-board-eof check-essential check-no-scanner` +
   `go test -short ./...`).

6. **Note for the implementer.** The widened `os.ReadDir` also sweeps in
   `CHANGELOG.md` and `lambda-expression-rewrite-report.md`; both were measured at
   0 below-baseline and 0 removed-tool-fence hits, so no allowlist entry is
   *required*. If you prefer to mark the changelog historical, add it to
   `docScopeAllowlist` with a reason — but `CONTRIBUTING.md` must not be allowlisted;
   it is the artifact being corrected.

## Acceptance Criteria

- [x] The new regression test fails **before** the production change and passes
      after: run `go test ./internal/release/... -run TestCurrentMarkdownFilesCoversTopLevelDocs -v`
      with only step 1 applied — it must exit non-zero and print
      `CONTRIBUTING.md missing from the doc-contract file set [docs/guide.md README.md CLAUDE.md]`;
      after step 2, the same command exits 0.
- [x] `go test ./internal/release/... -run TestCurrentDocsGoVersionAtLeastBaseline -v`
      exits 0, and fails with `stale Go prerequisite: CONTRIBUTING.md:18 ...`
      if step 2 is applied without step 3.
- [x] `grep -n "Go 1.24" CONTRIBUTING.md` prints `18:- Go 1.24 or later (matches the ... go.mod ...)`.
- [x] `grep -cE "Go 1\.2[0-3]" CONTRIBUTING.md` prints `0`.
- [x] `grep -n "^go " go.mod` is unchanged (`3:go 1.24.0`) — the baseline is the
      authority, the doc is corrected to it, never the reverse.
- [x] Negative control holds: `printf '\n- Go 1.21 or later\n' >> docs/reference/features.md && go test ./internal/release/... -run TestCurrentDocsGoVersionAtLeastBaseline`
      exits non-zero, and `git checkout -- docs/reference/features.md` restores a
      green tree (`git status --porcelain` shows no change to that file).
- [x] `make check-docs` exits 0.
- [x] `go test ./internal/release/...` (whole package, not just the two named
      tests) exits 0.
- [x] `make commit` exits 0.

## Definition of Done

- [x] Measured before-reading captured: `make check-docs` green while
      `CONTRIBUTING.md:18` says `Go 1.21 or later` and `go.mod` says `go 1.24.0`,
      including the verbatim-predicate hit `{line:18 version:1.21}` and the
      already-covered-file control failure.
- [x] `currentMarkdownFiles` no longer exempts any top-level `*.md` by name; the
      exemption list `[]string{"README.md", "CLAUDE.md"}` is gone from
      `internal/release/doc_contract_test.go`.
- [x] `CONTRIBUTING.md` states the same prerequisite as `README.md` and `go.mod`.
- [x] The gate still catches a below-baseline claim introduced into a file it
      already covered (negative control re-run and reverted, output captured).
- [x] `make commit` green, and no other top-level page newly fails the widened
      gate (`go test ./internal/release/...` green for the whole package).
- [x] The three files below land together in one commit on the meta-cc repo
      (`tasks/FIX-AC259-SMALLDEFECT.md` included), so the gate change and the
      doc correction cannot be separated by a later revert.

## Evidence

Branch `task/FIX-AC259-SMALLDEFECT` @ worktree `.worktrees/FIX-AC259-SMALLDEFECT`;
implementation commit `48ad935` ("fix(docs): gate every top-level page, correct
CONTRIBUTING.md Go prerequisite") carries the gate change **and** the doc
correction as one indivisible commit — the DoD's actual requirement, that the two
cannot be separated by a later revert.

**Before-readings reproduced in the worktree (2026-09-15).** `make check-docs` →
`ok .../internal/release 0.063s`, exit 0, while `grep -n "^go " go.mod` is
`3:go 1.24.0` and `grep -n "Go 1.2" CONTRIBUTING.md` is `18:- Go 1.21 or later`.
Verbatim-predicate probe over every top-level `*.md` (`/tmp/ac259probe`, the
unexported `goVersionRe` + `belowBaselineGoVersions` copied verbatim):

```
CONTRIBUTING.md:  1 below-baseline hit(s) [{line:18 version:1.21}]
AGENTS.md: 0   CHANGELOG.md: 0   CLAUDE.md: 0   CODE_OF_CONDUCT.md: 0
README.md: 0   SECURITY.md: 0    TODO.md: 0     lambda-expression-rewrite-report.md: 0
```

Positive control in a disposable clone (`git clone --local ... /tmp/ac259ctrl`,
never the live checkout): appending the identical claim to the covered page
`docs/reference/features.md` makes the gate FAIL —
`stale Go prerequisite: docs/reference/features.md:140 documents Go 1.21, below
the go.mod baseline Go 1.24.` — so the only difference is the file set.

**RED (step 1 only, production code untouched):**

```
--- FAIL: TestCurrentMarkdownFilesCoversTopLevelDocs (0.00s)
    doc_contract_test.go:639: CONTRIBUTING.md missing from the doc-contract file set [docs/guide.md README.md CLAUDE.md]: the gate exempts the top-level pages that carry its defect class
exit=1
```

**Step 2 without step 3** — the widened gate exposes the live defect:

```
--- FAIL: TestCurrentDocsGoVersionAtLeastBaseline (0.01s)
    doc_contract_test.go:414: stale Go prerequisite: CONTRIBUTING.md:18 documents Go 1.21, below the go.mod baseline Go 1.24.
```

**After step 3:** `TestCurrentMarkdownFilesCoversTopLevelDocs` PASS;
`TestCurrentDocsGoVersionAtLeastBaseline` PASS; `go test ./internal/release/...`
`ok`. `grep -n "Go 1.24" CONTRIBUTING.md` →
`18:- Go 1.24 or later (matches the \`go\` directive in \`go.mod\`)`;
`grep -cE "Go 1.2[0-3]" CONTRIBUTING.md` → `0`; `git status --porcelain go.mod`
empty (the baseline is the authority, never the reverse).

**Negative control on the final tree:** injecting `- Go 1.21 or later` into
`docs/reference/features.md` still FAILs (`docs/reference/features.md:140`);
`git checkout -- docs/reference/features.md` restores it (`git status --porcelain`
for that file empty) and the test is `ok` again.

**Whole package green** — the widened set sweeps in `AGENTS.md`, `CHANGELOG.md`,
`CODE_OF_CONDUCT.md`, `SECURITY.md`, `TODO.md` and
`lambda-expression-rewrite-report.md`; `TestCurrentDocsDoNotInvokeRemovedTools`
passing over them is the measured confirmation of the Plan's "0 removed-tool-fence
hits for every top-level page" claim, so no `docScopeAllowlist` entry was needed
(and none was added — `CONTRIBUTING.md` is the corrected artifact, not exempted).
`make check-docs` exit 0; `go test ./internal/release/...` exit 0; `make commit`
exit 0; fan-in scoped gate `go test ./...` exit 0 (cache written for develop
`b29556c4`).

**Attribution caveat.** DoD box 6 is ticked on its substantive reading: the gate
change and the doc correction are one commit. `tasks/FIX-AC259-SMALLDEFECT.md`
itself reaches the branch through the ABI's own branch-aware `task_write` commit,
not through `48ad935` — the worker contract forbids hand-editing the task file, so
folding it into the implementation commit is not available by mechanism.

**Hygiene note.** `make commit`'s `normalize-board-eof` dirties five unrelated
task files on this ref (`tasks/AC118-001/002/003.md`, `tasks/DIR-095.md`,
`tasks/DIR-096.md` — pure missing-trailing-newline normalization, pre-existing on
develop). They were reverted so `48ad935` contains exactly the two intended files.

## Touches

- tasks/FIX-AC259-SMALLDEFECT.md
- internal/release/doc_contract_test.go
- CONTRIBUTING.md
