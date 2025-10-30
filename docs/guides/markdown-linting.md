# Markdown Linting

This document describes the markdown linting setup for the meta-cc project.

## Configuration

The project uses a `.markdownlint.json` configuration file that defines the linting rules:

```json
{
  "default": true,
  "MD013": {
    "line_length": 120,
    "code_blocks": false,
    "tables": false
  },
  "MD024": {
    "siblings_only": true
  },
  "MD033": false,
  "MD041": false
}
```

## Makefile Integration

A new `lint-markdown` target has been added to the Makefile:

```makefile
lint-markdown:
	@echo "Running markdown linting..."
	@if command -v markdownlint >/dev/null 2>&1; then \
		markdownlint --config .markdownlint.json **/*.md; \
	elif command -v npm >/dev/null 2>&1 && npm list -g markdownlint-cli >/dev/null 2>&1; then \
		npx markdownlint-cli --config .markdownlint.json **/*.md; \
	else \
		echo "markdownlint not found. Install with:"; \
		echo "  npm install -g markdownlint-cli"; \
		echo "Skipping markdown linting..."; \
	fi
```

The `lint` target has been updated to include `lint-markdown`:

```makefile
lint: fmt vet lint-errors lint-error-handling lint-markdown
```

## CI/CD Integration

Markdown linting has been added to the CI workflow in `.github/workflows/ci.yml`:

```yaml
  lint:
    name: Lint
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
        with:
          fetch-depth: 0  # Fetch all history for CHANGELOG validation

      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.24'

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'

      - name: Install markdownlint-cli
        run: npm install -g markdownlint-cli

      - name: Run markdown linting
        run: markdownlint --config .markdownlint.json **/*.md

      - name: Run golangci-lint
        uses: golangci/golangci-lint-action@v6
        with:
          version: v1.64.8
```

## Local Development

To run markdown linting locally:

1. Install markdownlint-cli:
   ```bash
   npm install -g markdownlint-cli
   ```

2. Run the linting:
   ```bash
   make lint-markdown
   ```

   Or run all linting checks:
   ```bash
   make lint
   ```

## Fixing Issues

To fix markdown linting issues, you can:

1. Manually edit the files to comply with the linting rules
2. Use an editor plugin that supports markdownlint
3. Run markdownlint with the `--fix` flag to automatically fix some issues:
   ```bash
   markdownlint --config .markdownlint.json --fix **/*.md
   ```
