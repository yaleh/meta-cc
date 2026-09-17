#!/usr/bin/env bash
# quay-suite.sh — meta-cc's quay suite entrypoint: the value of loop.test_command in
# .quay/config.yml, and the producer of the summary line loop.test_output parses.
#
# WHY THIS EXISTS (task gap-quay-test-output-undeclared-counts-absent). The quay web test card
# rendered every meta-cc round as `pass —/—`. Those counts were honestly ABSENT, not zero: quay
# reads them out of the suite log, and its builtin parser only understands node:test / TAP —
#
#   TEST_COUNT_RE = /^[#ℹ]\s*(pass|fail|cancelled)\s+(\d+)/       (pre-verified-round-record.ts)
#
# — so `go test`'s `ok  <pkg> 0.12s` / `--- FAIL: TestX` never produced a count. The DECLARED
# path (loop.test_output → applyDeclaredTestOutput) matches ONE line with ONE capture group and
# never sums anything, so it needs a line that already carries the totals — which `go test ./...`
# does not print. This entrypoint prints it.
#
# CONTRACT (what quay depends on):
#   * stdout = `go test -v`'s own output, unmodified, with stderr folded in (compile errors are
#     the ONLY explanation for a `[build failed]` red, and quay captures stdout alone), followed
#     by EXACTLY ONE summary line, as the LAST line:
#         quay-suite: pass <N> fail <M> skip <K>
#     .quay/config.yml's loop.test_output regexes bind to it, one capture group each.
#   * the exit status IS go test's status — a red suite stays red. This wrapper can never turn a
#     failing suite green.
#   * `skip` maps to quay's `cancelled` field, so quay's tests = pass+fail+cancelled equals the
#     real number of top-level test functions (the same reading as the vitest reference impl).
#   * counts are TOP-LEVEL test functions only: Go indents subtest result lines, so anchoring the
#     match at column 0 excludes them instead of inflating the totals.
#   * the summary is printed ONLY when the round is accountable — at least one top-level result
#     line was seen AND no package failed to build/setup. Anything else keeps the counts ABSENT
#     (the same honest `—/—` as before) rather than reporting a number that quietly omits a
#     package that never ran. A declared-but-unmatched field does NOT fall back to the builtin
#     parser, so a broken regex here degrades to "absent", never to a fabricated number.
#
# ⛔ DO NOT rename this to scripts/test.sh. quay's testShAt(dir) is literally `<dir>/scripts/test.sh`;
#    its presence flips hasTestSh and switches the round onto quay's own full-suite-runner/TAP
#    path (suiteRunsOutsideRunner goes false) — a different mechanism, not a renamed one.
#
# ⛔ A test that shells out to this script (or to `go test`) and lets the subprocess output reach
#    its OWN stdout injects extra `--- FAIL:` lines and a second summary line into the outer suite
#    log — and quay's declared regex takes the FIRST summary match, not the last. Capture such
#    output (exec buffers, t.Log); never print it.
#
# Usage: bash scripts/quay-suite.sh [go test args...]        (no args ⇒ ./...)
#   The driver runs this with NO arguments and never passes the --buckets/--root/--state-dir/
#   --runner/--log-file/--run-id flags — those belong to a project's own scripts/test.sh, which
#   meta-cc deliberately does not ship (see .quay/config.yml's loop.test_command note).

set -uo pipefail

if [ "$#" -eq 0 ]; then
  set -- ./...
fi

out="$(mktemp "${TMPDIR:-/tmp}/quay-suite.XXXXXX")" || exit 2
trap 'rm -f "${out}"' EXIT

# -v, not -json: the counts are read from the per-test result lines, and replaying a JSON stream
# would need a decoder (jq / node / python) sitting on the fan-in path — a hard dependency this
# entrypoint must not acquire. `go test -v`'s stdout IS the log.
go test -v "$@" >"${out}" 2>&1
rc=$?

cat "${out}"

# One pass, POSIX awk only (no jq/node/python — see the -v note above).
#   * ^-anchored: a subtest result is `    --- PASS: TestX/sub (0.00s)` and must not be counted.
#   * the FULL result-line shape (name + trailing duration) is required, so test output that
#     merely quotes such a line cannot forge a count.
#   * `[build failed]` / `[setup failed]` → counts withheld; see the contract note above.
awk '
  /^--- (PASS|FAIL|SKIP): .* \([0-9.]+s\)$/ {
    n++
    if ($2 == "PASS:") pass++
    else if ($2 == "FAIL:") fail++
    else skip++
  }
  /\[(build|setup) failed\]$/ { unaccountable = 1 }
  END {
    if (n > 0 && !unaccountable)
      printf "quay-suite: pass %d fail %d skip %d\n", pass, fail, skip
  }
' "${out}"

exit "${rc}"
