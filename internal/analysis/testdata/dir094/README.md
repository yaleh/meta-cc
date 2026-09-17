# DIR-094 malformed-session tolerance corpus

Checked-in regression fixtures for the contract that one unusable session file
must never erase the results derived from the rest of the corpus, and must
never be dropped SILENTLY (DIR-018, unified across every enumerating path by
DIR-094). They exist so a future rewrite of the corpus-enumerating paths cannot
lose the tolerance without a test going red — the behavior is pinned to files
on disk, not to an in-test string literal someone can quietly simplify.

`internal/analysis/service_skipped_files_test.go` and
`internal/provider/claude/provider_skips_test.go` install these into a
project-hash transcript directory and assert, per path, that the tool returns
results PLUS a warning naming the file.

| File | Shape | Why it exists |
| --- | --- | --- |
| `valid-session.jsonl` | A real 10-entry Claude session (user → assistant → Grep → Read → Bash error → Bash fix), carrying a `TODO` marker and one error→fix pair | The control: proves the surrounding corpus is still returned intact alongside the warning. It is deliberately *not* empty of signal — the TODO and the error→fix pair make `get_tech_debt` and `analyze_bugs` produce distinguishable output, so "results present" is a real assertion for every tool and not just "the call returned" |
| `empty-session.jsonl` | 0 bytes | The exact shape of the 2026-07-30 dogfooding failure (`8eda8f4e-…jsonl`): a session file Claude Code created but never wrote a message into. Parses with NO error and zero entries — the silent-tolerance case DIR-094 closed |
| `metadata-only-session.jsonl` | Only `file-history-snapshot` / `mode` / `permission-mode` / `system` lines | The zero-message "stub" shape: valid JSON, zero message entries. Same silent path as the empty file, reached without a 0-byte file |
| `malformed-session.jsonl` | `{this is not valid json` | The DIR-018 case (a parse *error*, not a silent empty result) — kept here so one corpus exercises both halves of the contract |
| `whitespace-session.jsonl` | Blank/whitespace lines only | A file that is non-empty on disk but has no content lines; guards against a "size > 0 means usable" assumption |

These are data fixtures, not Go test files: they are read with `os.ReadFile`
from `testdata/`, so `make check-fixtures` (which verifies `testutil.LoadFixture`
references) does not apply to them.

**Observability note:** an MCP tool response only shows this behavior after the
installed plugin binary is refreshed — a stale install keeps serving the old
tolerance. See `docs/guides/troubleshooting.md` ("Malformed session file
tolerated but no warning appears").
