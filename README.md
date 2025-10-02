# meta-cc

Meta-Cognition tool for Claude Code - analyze session history for workflow optimization.

## Features

- 🔍 Parse Claude Code session history (JSONL format)
- 📊 Statistical analysis of tool usage and errors
- 🎯 Pattern detection for workflow optimization
- 🚀 Zero dependencies - single binary deployment

## Installation

### From Source

```bash
git clone https://github.com/yale/meta-cc.git
cd meta-cc
make build
```

### Cross-Platform Binaries

```bash
# Build for all supported platforms
make cross-compile

# Binaries will be in build/ directory:
# - build/meta-cc-linux-amd64
# - build/meta-cc-linux-arm64
# - build/meta-cc-darwin-amd64
# - build/meta-cc-darwin-arm64
# - build/meta-cc-windows-amd64.exe
```

## Usage

```bash
# Show help
./meta-cc --help

# Show version
./meta-cc --version

# Global options
./meta-cc --session <session-id>    # Specify session ID
./meta-cc --project <path>          # Specify project path
./meta-cc --output json|md|csv      # Output format
```

## Development

### Prerequisites

- Go 1.21 or later
- Make (optional, for build automation)

### Build

```bash
# Using Make
make build

# Or using go directly
go build -o meta-cc
```

### Test

```bash
# Run all tests
make test

# Run with coverage
make test-coverage
# Open coverage.html in browser
```

### Available Make Targets

```bash
make build           # Build for current platform
make test            # Run tests
make test-coverage   # Run tests with coverage report
make clean           # Remove build artifacts
make install         # Install to GOPATH/bin
make cross-compile   # Build for all platforms
make deps            # Download and tidy dependencies
make help            # Show help message
```

## Supported Platforms

- Linux (amd64, arm64)
- macOS (amd64, arm64/Apple Silicon)
- Windows (amd64)

## Project Structure

```
meta-cc/
├── cmd/              # Command definitions (Cobra)
├── internal/         # Internal packages
│   └── testutil/    # Test utilities
├── pkg/              # Public packages
├── tests/            # Test files and fixtures
└── docs/             # Documentation
```

## License

MIT
