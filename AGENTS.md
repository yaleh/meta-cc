# Repository Guidelines

## Project Structure & Module Organization
- `cmd/` holds CLI and MCP entry points; `main.go` wires shared packages.
- Core logic lives in `internal/` (analysis, session parsing) with reusable helpers in `pkg/`.
- Plugin-facing content sits under `capabilities/` and `.claude/` (commands, agents, skills); docs and diagrams live in `docs/`.
- Integration and regression suites live beside source code, with longer MCA workflows in `tests/integration/`; automation scripts are in `scripts/`.

## Build, Test, and Development Commands
- `make dev` formats code then builds both binaries for quick local iteration.
- `make build` produces `meta-cc` and `meta-cc-mcp`; `make install` drops the CLI into your `GOBIN`.
- `make test` runs short Go tests; `make test-all` includes slow E2E coverage; `make test-coverage` emits `coverage.html`.
- Quality bundles: `make pre-commit` (full lint + tests), `make ci` (all checks), `make bundle-capabilities` (tarball in `build/`).

## Coding Style & Naming Conventions
- Target Go toolchain `go1.24.9`; rely on `gofmt`, `goimports`, and `golangci-lint` via `make fmt` and `make lint`.
- Follow Go idiomatic naming (`MetaClient`, `parseSession`); keep filenames snake_case.
- Error handling leans on `internal/mcerrors`; run `make lint-errors` to verify sentinel wrapping.
- Prefer package-scoped docs for exported APIs; keep agent/spec markdowns concise and task-scoped.

## Testing Guidelines
- Use Go’s `testing` package with `TestXxx` naming; table tests live next to implementations in `internal/`.
- Integration flows execute from `tests/integration/` scripts; mark long-running tests with `//go:build e2e` and rely on `make test-all`.
- Maintain ≥80% coverage (`make test-coverage-check`); attach new fixtures under `tests/fixtures/` and register cleanup in tests.

## Commit & Pull Request Guidelines
- Commit messages follow Conventional Commits (`type(scope): summary`) as seen in history (`docs(methodology): …`).
- Before pushing, run `make pre-commit` and confirm lint, vet, and sentinel error checks pass.
- Pull requests should describe intent, link issues, and note any capability bundle impacts; add screenshots for UI/tooling output when relevant.
- For release-ready changes, mention if `bundle-capabilities` or `sync-plugin-files` needs to run.

## Security & Configuration Tips
- Scan dependencies with `make security`; ensure new capabilities reference vetted tools only.
- Keep API tokens out of repo—load them via environment reads in `internal/config` and document variables in `docs/reference/`.
