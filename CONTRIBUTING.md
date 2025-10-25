# Contributing to meta-cc

We welcome contributions from the community! Thank you for your interest in improving meta-cc. This document provides guidelines to help you through the contribution process.

Please also read and adhere to our [Code of Conduct](CODE_OF_CONDUCT.md).

## How to Contribute

You can contribute in several ways:
- Reporting bugs
- Suggesting enhancements
- Improving documentation
- Submitting pull requests with new features or bug fixes

## Development Setup

### Prerequisites
- Go 1.21 or later
- `make`
- `golangci-lint` (for local linting)

### Setup Instructions
```bash
# 1. Fork and clone the repository
git clone https://github.com/yaleh/meta-cc.git
cd meta-cc

# 2. Install dependencies
go mod download

# 3. Run the full local verification suite
make all
```
The `make all` command will run linters, tests, and build the binaries, ensuring your environment is set up correctly.

## Code Style

- We follow standard Go formatting. Use `gofmt` or your IDE's auto-formatter.
- Run `make lint` before committing to catch common style issues and errors.
- For comments, focus on *why* a piece of code exists, especially for complex logic, rather than *what* it does.

## Testing

- All new features and bug fixes must include corresponding tests.
- The project maintains a test coverage target of ≥80%.
- Run the test suite with `make test` before submitting a pull request.
- To generate a detailed coverage report, run `make test-coverage`. This creates an `coverage.html` file you can open in your browser.

## Commit Message Format

We follow the [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) specification. This creates a clean and readable git history and helps automate the release process.

A commit message should be structured as follows:
```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

- **Types**: `feat`, `fix`, `docs`, `test`, `refactor`, `chore`, `build`, `ci`.
- **Scope**: The part of the codebase your changes affect (e.g., `parser`, `mcp`, `docs`, `ci`).

**Example:**
```
feat(query): add --limit flag to query tools command

This allows users to limit the number of results returned,
which is crucial for preventing context overflow when used
by an LLM.
```

## Pull Request Process

1.  **Fork the repository** and create your branch from the `main` branch.
2.  Make your changes, adhering to the code style and testing guidelines.
3.  Ensure the full test and lint suite passes with `make all`.
4.  Commit your changes using the Conventional Commits format.
5.  Push your branch to your fork and open a pull request against the `main` branch.
6.  In the pull request description, clearly explain the problem and your solution. Link to any relevant issues.
7.  A maintainer will review your PR. Be prepared to discuss your changes and make adjustments.

## Questions?

If you have any questions, feel free to open an issue or start a thread in the repository's "Discussions" tab.
