---
name: meta-project-bootstrap
description: Bootstrap new project with complete Phase 0 documentation structure following Documentation Management Methodology.
keywords: bootstrap, project-setup, phase-0, documentation-structure, initialization, best-practices
category: workflow
---

λ(project_context) → phase_0_structure | ∀component ∈ {essential_docs, directory_structure, templates}:

scope :: project_directory

bootstrap :: ProjectContext → Phase0Structure
bootstrap(C) = detect(project_type) ∧ analyze(stack) ∧ generate(docs) ∧ verify(completeness)

## Purpose

Creates complete Phase 0 documentation structure for new or existing projects, implementing all requirements from Documentation Management Methodology v5.0.

**Self-contained**: Includes all necessary knowledge without requiring external document references.

## Project Type Detection

detect :: ProjectContext → ProjectProfile
detect(C) = {
  project_type: infer_from([
    directory_structure,
    manifest_files,
    source_code_patterns
  ]),

  language_stack: detect_languages([
    primary: main_language,
    secondary: supporting_languages,
    frameworks: detected_frameworks
  ]),

  existing_state: assess([
    has_documentation: check_docs_existence(),
    has_source_code: check_src_existence(),
    git_initialized: check_git_status()
  ])
}

### Project Type Categories

**By Purpose**:
- **web/frontend**: React, Vue, Angular, Next.js, Svelte
- **backend/api**: REST API, GraphQL, microservices
- **cli**: Command-line tools, utilities
- **library**: Reusable packages, frameworks
- **system**: System tools, daemons, infrastructure
- **data**: Data processing, ETL, analytics
- **embedded**: IoT, hardware interface

**By Language Stack**:
- **Go**: CLI tools, backend services, system tools
- **Rust**: System tools, performance-critical, CLI
- **Python**: Data science, web (Django/Flask), automation
- **JavaScript/TypeScript**: Web frontend, Node.js backend
- **Java**: Enterprise backend, Android
- **C++**: System programming, performance-critical
- **Ruby**: Web (Rails), automation
- **C#**: .NET applications, game development

## Analysis Phase

analyze :: ProjectProfile → StackAnalysis
analyze(P) = {
  build_system: detect_build_tool([
    "Makefile", "go.mod", "Cargo.toml", "package.json",
    "pom.xml", "build.gradle", "CMakeLists.txt",
    "pyproject.toml", "setup.py", "Gemfile"
  ]),

  test_framework: detect_test_tool([
    "go test", "cargo test", "pytest", "jest",
    "mocha", "JUnit", "RSpec", "Catch2"
  ]),

  linter: detect_linter([
    "golangci-lint", "clippy", "eslint", "pylint",
    "rubocop", "clang-tidy"
  ]),

  dependencies: parse_dependency_files(),

  best_practices: lookup_language_conventions(P.language_stack.primary)
}

## Document Generation

generate :: StackAnalysis → Phase0Documents
generate(A) = {
  core_documents: create_core_docs(A),
  supporting_files: create_supporting_files(A),
  directory_structure: create_directory_tree()
}

### Phase 0 Essential Documents

**Must Create** (in order of importance):

#### 1. docs/core/plan.md

**Purpose**: Project roadmap and development journal
**Update frequency**: High (daily/weekly during active development)
**Target size**: 50-200 lines initially, grows with project

**Template Structure**:

```markdown
# Development Plan

## Vision

[What problem does this project solve?]
[Why is this approach better than alternatives?]
[Who are the target users?]

(1-3 paragraphs maximum)

## Phases

### Phase 0: Bootstrap (Current)
**Status**: 🚧 In Progress
**Goal**: Create minimal viable documentation and project structure
**Tasks**:
- [x] Create plan.md and principles.md
- [ ] Implement core feature X
- [ ] Write tests for X
- [ ] Document basic usage

**Exit criteria**:
- All Phase 0 documents exist
- Code compiles successfully
- Basic tests pass

### Phase 1: [First Milestone Name]
**Status**: 📋 Planned
**Goal**: [What will this achieve?]
**Dependencies**: Phase 0 complete
**Tasks**:
- [ ] Task 1
- [ ] Task 2
- [ ] Task 3

### Phase 2: [Second Milestone Name]
**Status**: 📋 Planned
**Goal**: [What will this achieve?]
**Dependencies**: Phase 1 complete

### Phase 3: [Future Milestone]
**Status**: 💭 Future
**Goal**: [Long-term vision]

## Current Status

**Working on**: Phase 0 - Creating documentation structure
**Completed**: Initial project setup
**Blockers**: None
**Next steps**: Implement core functionality

**Last updated**: YYYY-MM-DD

## Notes

[Add notes here as you develop - this section grows organically]

### Development Decisions
- [Date] Decision about X: Chose Y because Z
- [Date] Discovered pattern: ...

### Lessons Learned
- [Lesson 1]
- [Lesson 2]

### Future Considerations
- [Idea 1]
- [Idea 2]
```

**Language-Specific Customizations**:

**Go projects**:
```markdown
### Phase 1: Core Implementation
**Tasks**:
- [ ] Design package structure (internal/, cmd/, pkg/)
- [ ] Implement core types and interfaces
- [ ] Write table-driven tests
- [ ] Add golangci-lint configuration
```

**Rust projects**:
```markdown
### Phase 1: Core Implementation
**Tasks**:
- [ ] Define module structure and visibility
- [ ] Implement core types with proper lifetimes
- [ ] Write unit and integration tests
- [ ] Configure clippy and rustfmt
```

**Python projects**:
```markdown
### Phase 1: Core Implementation
**Tasks**:
- [ ] Set up virtual environment (venv/poetry)
- [ ] Define package structure
- [ ] Write pytest-based tests
- [ ] Configure black and pylint
```

**JavaScript/TypeScript projects**:
```markdown
### Phase 1: Core Implementation
**Tasks**:
- [ ] Configure TypeScript (if applicable)
- [ ] Set up bundler (webpack/vite/rollup)
- [ ] Write Jest/Vitest tests
- [ ] Configure ESLint and Prettier
```

#### 2. docs/core/principles.md

**Purpose**: Design constraints and architectural rules
**Update frequency**: Low (rarely, when patterns solidify)
**Target size**: 30-100 lines initially

**Template Structure**:

```markdown
# Design Principles

## Core Constraints

### Code Limits
- **Phase limit**: ≤500 lines of changes per phase
- **Stage limit**: ≤200 lines of changes per stage
- **Function size**: [language-specific recommendation]
- **File size**: [language-specific recommendation]

### Quality Standards
- **Test coverage**: ≥80% (measured by [tool])
- **Documentation**: All public APIs documented
- **Performance**: [specific requirements if applicable]
- **Security**: [specific requirements if applicable]

## Development Methodology

### Test-Driven Development (TDD)
- Write tests before implementation
- Red-Green-Refactor cycle
- [Language-specific testing approach]

### Version Control
- [Branching strategy: trunk-based, GitFlow, etc.]
- [Commit message format: conventional commits, etc.]
- [PR/review requirements]

### Continuous Integration
- All tests must pass before merge
- Lint checks required
- [Other CI requirements]

## Architecture Principles

### [Principle 1]
[Description and rationale]

### [Principle 2]
[Description and rationale]

## Technology Stack

### Core Technologies
- **Language**: [Primary language + version]
- **Build**: [Build system]
- **Test**: [Test framework]
- **Lint**: [Linting tools]

### Dependencies
[Dependency management philosophy: minimal, explicit, etc.]

## Anti-Patterns to Avoid

### [Anti-pattern 1]
**Problem**: [What's wrong]
**Solution**: [Better approach]

### [Anti-pattern 2]
**Problem**: [What's wrong]
**Solution**: [Better approach]

## Notes

[Add notes as patterns emerge during development]
```

**Language-Specific Sections**:

**Go**:
```markdown
## Go-Specific Principles

### Package Design
- Follow standard Go project layout
- Use internal/ for private packages
- cmd/ for executables, pkg/ for libraries

### Error Handling
- Explicit error returns
- Wrap errors with context
- No panic in library code

### Concurrency
- Share memory by communicating (channels)
- Document goroutine lifecycles
- Use context for cancellation
```

**Rust**:
```markdown
## Rust-Specific Principles

### Ownership and Borrowing
- Prefer borrowing over cloning
- Document lifetime requirements
- Use Cow<'a, T> for flexible ownership

### Error Handling
- Use Result<T, E> for recoverable errors
- Custom error types with thiserror
- Avoid unwrap() in production code

### Safety
- Minimize unsafe code
- Document safety invariants
- Fuzz critical parsers
```

**Python**:
```markdown
## Python-Specific Principles

### Type Hints
- Use type hints for public APIs
- Run mypy in strict mode
- Document complex types

### Code Style
- Follow PEP 8
- Use black for formatting
- Maximum line length: 88 characters

### Dependencies
- Pin versions in requirements.txt
- Use poetry for dependency management
- Minimize third-party dependencies
```

**JavaScript/TypeScript**:
```markdown
## JavaScript/TypeScript Principles

### Type Safety (TypeScript)
- Strict mode enabled
- Avoid 'any' type
- Use union types over type assertions

### Module System
- ES modules only (no CommonJS)
- Named exports preferred
- Avoid default exports

### Async Patterns
- Prefer async/await over callbacks
- Handle promise rejections
- Use AbortController for cancellation
```

#### 3. CLAUDE.md

**Purpose**: Development guide for Claude Code
**Update frequency**: Medium (per phase, as FAQ grows)
**Target size**: 50-150 lines initially

**Template Structure**:

```markdown
# CLAUDE.md

Development guide for Claude Code assistance on this project.

## Quick Links

### Essential Reading
- [docs/core/plan.md](docs/core/plan.md) - Current roadmap and status
- [docs/core/principles.md](docs/core/principles.md) - Design constraints

### Development
- Build: See [Development Commands](#development-commands)
- Test: See [Development Commands](#development-commands)

## Project Overview

**Goal**: [One paragraph: what problem does this solve?]

**Target users**: [Who will use this?]

**Key features**:
- Feature 1
- Feature 2
- Feature 3

## Development Commands

### Build
```bash
[build-command]
```

### Test
```bash
[test-command]
```

### Lint
```bash
[lint-command]
```

### Run
```bash
[run-command-with-example-args]
```

## Project Structure

```
[project-root]/
├── [src-directory]/     # [Description]
├── [test-directory]/    # [Description]
├── docs/                # Documentation
└── [other-key-dirs]/    # [Description]
```

## Development Workflow

### Before Starting Work
1. Read [docs/core/plan.md](docs/core/plan.md) for current status
2. Review [docs/core/principles.md](docs/core/principles.md) for constraints
3. Run tests to ensure clean state

### Making Changes
1. Write tests first (TDD)
2. Implement changes
3. Run full test suite
4. Update docs if needed

### Before Committing
- [ ] All tests pass
- [ ] Code passes lint checks
- [ ] Documentation updated
- [ ] Commit message follows convention

## Architecture Notes

[Add architectural insights as they emerge during development]

### [Pattern/Component 1]
[Description]

### [Pattern/Component 2]
[Description]

## FAQ

[Add Q&A as questions arise during development]

### Q: [Common question 1]?
A: [Answer]

### Q: [Common question 2]?
A: [Answer]

## Common Tasks

### [Task 1]
```bash
[commands]
```

### [Task 2]
```bash
[commands]
```

## Troubleshooting

### [Issue 1]
**Problem**: [Description]
**Solution**: [How to fix]

### [Issue 2]
**Problem**: [Description]
**Solution**: [How to fix]
```

**Language-Specific Commands**:

**Go**:
```markdown
## Development Commands

### Build
```bash
go build -v ./...
```

### Test
```bash
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out  # View coverage
```

### Lint
```bash
golangci-lint run
```

### Format
```bash
gofmt -w .
```
```

**Rust**:
```markdown
## Development Commands

### Build
```bash
cargo build
cargo build --release  # Optimized build
```

### Test
```bash
cargo test
cargo test -- --nocapture  # Show println! output
```

### Lint
```bash
cargo clippy -- -D warnings
```

### Format
```bash
cargo fmt
```
```

**Python**:
```markdown
## Development Commands

### Setup
```bash
python -m venv venv
source venv/bin/activate  # Windows: venv\Scripts\activate
pip install -r requirements.txt
```

### Test
```bash
pytest
pytest --cov=src --cov-report=html  # With coverage
```

### Lint
```bash
black .
pylint src/
mypy src/
```
```

**JavaScript/TypeScript**:
```markdown
## Development Commands

### Install
```bash
npm install  # or: yarn install, pnpm install
```

### Build
```bash
npm run build
```

### Test
```bash
npm test
npm run test:watch  # Watch mode
npm run test:coverage  # With coverage
```

### Lint
```bash
npm run lint
npm run lint:fix  # Auto-fix issues
```
```

#### 4. README.md

**Purpose**: Public entry point for all users
**Update frequency**: Low (major releases)
**Target size**: 50-200 lines

**Template Structure**:

```markdown
# Project Name

[One-sentence description of what this project does]

## Status

🚧 Early development - [Current phase] in progress

**Current version**: v0.1.0 (or: Not yet released)
**License**: [License type]

## What is this?

[2-3 paragraphs explaining:
- What problem it solves
- Who should use it
- Key differentiators]

## Quick Start

### Installation

```bash
[installation-command]
```

### Basic Usage

```bash
[basic-usage-example]
```

### Example

```[language]
[minimal-working-example]
```

## Features

- ✅ Feature 1 (implemented)
- ✅ Feature 2 (implemented)
- 🚧 Feature 3 (in progress)
- 📋 Feature 4 (planned)

## Documentation

### For Users
- [Installation Guide](docs/tutorials/installation.md) - Detailed setup
- [User Guide](docs/guides/user-guide.md) - Complete usage guide

### For Contributors
- [Development Guide](CLAUDE.md) - Development setup and workflow
- [Development Plan](docs/core/plan.md) - Project roadmap
- [Contributing Guidelines](CONTRIBUTING.md) - How to contribute

## Requirements

- [Requirement 1, e.g., "Go 1.21 or later"]
- [Requirement 2, e.g., "Node.js 18+"]
- [Requirement 3]

## License

[License name] - See [LICENSE](LICENSE) for details.

## Support

- **Issues**: [Issue tracker URL or "GitHub Issues"]
- **Discussions**: [Discussion forum if applicable]
- **Contact**: [Contact method]

## Acknowledgments

[Optional: Credits, inspirations, major dependencies]
```

**Project-Type Specific Examples**:

**CLI Tool**:
```markdown
## Installation

### Binary Release (Recommended)
```bash
# Download from releases page
curl -L https://github.com/user/project/releases/latest/download/tool-linux-amd64 -o tool
chmod +x tool
sudo mv tool /usr/local/bin/
```

### From Source
```bash
git clone https://github.com/user/project
cd project
make install  # or: go install, cargo install, npm install -g
```

## Usage

```bash
# Basic command
tool [options] <args>

# Examples
tool --help
tool process input.txt
tool serve --port 8080
```
```

**Library**:
```markdown
## Installation

```bash
# Go
go get github.com/user/project

# Rust
cargo add project

# Python
pip install project

# JavaScript/TypeScript
npm install project
```

## Usage

```[language]
[import/require statement]

[minimal-api-usage-example]
```

## API Documentation

See [API Reference](docs/reference/api.md) for complete API documentation.
```

**Web Application**:
```markdown
## Quick Start

### Development
```bash
npm install
npm run dev
```

Open http://localhost:3000 in your browser.

### Production
```bash
npm run build
npm start
```

### Docker
```bash
docker build -t project .
docker run -p 3000:3000 project
```
```

#### 5. Supporting Files

##### .gitignore

**Generate based on detected language**:

**Go**:
```
# Binaries
*.exe
*.exe~
*.dll
*.so
*.dylib
/bin/
/dist/

# Test coverage
*.out
coverage.html

# Go workspace
go.work
go.work.sum

# IDE
.vscode/
.idea/
*.swp
*.swo
```

**Rust**:
```
# Build artifacts
/target/
Cargo.lock  # Remove this line for libraries

# IDE
.vscode/
.idea/
*.swp

# Testing
*.profraw
```

**Python**:
```
# Virtual environment
venv/
.venv/
env/
ENV/

# Python artifacts
__pycache__/
*.py[cod]
*$py.class
*.so

# Distribution
dist/
build/
*.egg-info/

# Testing
.pytest_cache/
.coverage
htmlcov/

# IDE
.vscode/
.idea/
*.swp
```

**JavaScript/TypeScript**:
```
# Dependencies
node_modules/
package-lock.json  # If using yarn/pnpm

# Build output
dist/
build/
.next/
.nuxt/

# Environment
.env
.env.local

# Testing
coverage/

# IDE
.vscode/
.idea/
*.swp
```

##### LICENSE

**Prompt for license selection**:

Common options:
- **MIT**: Permissive, simple
- **Apache 2.0**: Permissive, patent protection
- **GPL v3**: Copyleft, strong
- **BSD 3-Clause**: Permissive, no endorsement
- **Proprietary**: Closed source

**Generate appropriate LICENSE file based on selection.**

### Directory Structure Creation

create_directory_tree :: () → DirectoryStructure
create_directory_tree() = {
  create_directories([
    "docs/core",
    "docs/guides",
    "docs/reference",
    "docs/tutorials",
    "docs/architecture/adr",
    "docs/methodology",
    "docs/archive"
  ]),

  create_placeholder_files([
    "docs/guides/.gitkeep",
    "docs/reference/.gitkeep",
    "docs/tutorials/.gitkeep",
    "docs/architecture/adr/.gitkeep"
  ])
}

**Directory purposes** (include in generated CLAUDE.md):
```
docs/
├── core/                    # Core project documents
│   ├── plan.md             # Development roadmap (required)
│   └── principles.md       # Design constraints (required)
├── guides/                  # Task-oriented guides (create as needed)
├── reference/               # Complete specifications (create as needed)
├── tutorials/               # Step-by-step tutorials (create as needed)
├── architecture/            # Architectural documentation
│   └── adr/                # Architecture Decision Records
├── methodology/             # Project-independent methodologies
└── archive/                 # Archived documentation
```

## Verification Phase

verify :: Phase0Documents → VerificationReport
verify(D) = {
  completeness: check_required_files_exist([
    "docs/core/plan.md",
    "docs/core/principles.md",
    "CLAUDE.md",
    "README.md",
    ".gitignore",
    "LICENSE"
  ]),

  quality: check_document_quality([
    plan_has_phases: grep_count("Phase", "docs/core/plan.md") >= 2,
    principles_has_constraints: file_length("docs/core/principles.md") >= 30,
    readme_has_quick_start: grep_exists("Quick Start", "README.md"),
    claude_has_commands: grep_exists("Development Commands", "CLAUDE.md")
  ]),

  structure: verify_directory_structure([
    "docs/core/",
    "docs/guides/",
    "docs/reference/",
    "docs/tutorials/",
    "docs/architecture/adr/"
  ])
}

### Phase 0 Exit Criteria Checklist

**Must verify before marking Phase 0 complete**:

- [ ] **docs/core/plan.md exists**
  - Contains Vision section
  - Has at least Phase 0-2 outlined
  - Includes Current Status section
  - Has Notes section initialized

- [ ] **docs/core/principles.md exists**
  - Defines code limits (phase/stage)
  - Specifies test coverage requirement
  - Documents development methodology
  - Includes language-specific constraints

- [ ] **CLAUDE.md exists**
  - Links to plan.md and principles.md
  - Contains project overview
  - Documents build/test/lint commands
  - Has FAQ section initialized

- [ ] **README.md exists**
  - One-sentence description present
  - Quick Start section with working examples
  - Links to documentation
  - License specified

- [ ] **Supporting files created**
  - .gitignore (language-appropriate)
  - LICENSE (selected and generated)

- [ ] **Directory structure created**
  - docs/core/ directory exists
  - docs/guides/ directory exists
  - docs/reference/ directory exists
  - docs/tutorials/ directory exists
  - docs/architecture/adr/ directory exists

- [ ] **Build verification** (if code exists)
  - Project compiles successfully
  - Basic tests pass (if any exist)
  - Linter runs without critical errors

## Usage Examples

### New Project Bootstrap

**Scenario**: Starting a new Go CLI tool

**Input**:
```
Project directory: /path/to/new-project
Project type: CLI tool
Language: Go
Has source code: No
Purpose: HTTP benchmarking tool
```

**Generated Output**:
- docs/core/plan.md (customized for CLI tool)
- docs/core/principles.md (Go-specific constraints)
- CLAUDE.md (Go commands: go build, go test, golangci-lint)
- README.md (CLI-focused with installation methods)
- .gitignore (Go-specific ignores)
- LICENSE (user selects)
- Complete docs/ directory structure

### Existing Project Documentation Audit

**Scenario**: Existing Rust library needs documentation

**Input**:
```
Project directory: /path/to/existing-project
Project type: Library
Language: Rust
Has source code: Yes (10 .rs files)
Has documentation: Partial (only README.md exists)
```

**Actions**:
1. Detect: Rust library (from Cargo.toml)
2. Analyze: Existing README.md (preserve, enhance if needed)
3. Generate:
   - docs/core/plan.md (based on existing code)
   - docs/core/principles.md (Rust best practices)
   - CLAUDE.md (cargo commands)
   - Enhance README.md (add missing sections)
4. Create directory structure
5. Provide migration checklist

### Multi-Language Project

**Scenario**: Full-stack web app (TypeScript frontend + Python backend)

**Input**:
```
Project type: Web application
Languages: TypeScript (frontend), Python (backend)
Frameworks: React, FastAPI
```

**Generated Output**:
- Dual-language principles.md (TypeScript + Python sections)
- CLAUDE.md with separate frontend/backend commands
- README.md with both setup instructions
- Composite .gitignore (node_modules + __pycache__)
- Separate phase planning for frontend/backend work

## Best Practices and Anti-Patterns

### ✅ Best Practices

1. **Context-Aware Generation**
   - Detect project type before generating templates
   - Customize content based on language stack
   - Preserve existing content when enhancing documentation

2. **Minimal but Complete**
   - Create only required Phase 0 files
   - Don't generate guides/tutorials prematurely
   - Initialize sections for future content (FAQ, Notes)

3. **Actionable Documentation**
   - All commands are tested and working
   - Examples are minimal but functional
   - Links point to existing or planned documents

4. **Language Idioms**
   - Use language-specific terminology
   - Follow community conventions
   - Reference standard tools and practices

### ❌ Anti-Patterns to Avoid

1. **❌ One-Size-Fits-All Templates**
   - **Problem**: Generic templates lack relevance
   - **Solution**: Detect and customize for project context

2. **❌ Premature Complexity**
   - **Problem**: Creating guides/references before code exists
   - **Solution**: Phase 0 = essential docs only

3. **❌ Copy-Paste Without Customization**
   - **Problem**: Placeholder text left in final docs
   - **Solution**: Generate specific, actionable content

4. **❌ Ignoring Existing State**
   - **Problem**: Overwriting existing documentation
   - **Solution**: Preserve and enhance existing content

5. **❌ Broken Links on Day One**
   - **Problem**: Links to non-existent documents
   - **Solution**: Only link to generated files or mark as "planned"

6. **❌ Missing Exit Criteria**
   - **Problem**: No clear definition of "Phase 0 complete"
   - **Solution**: Include verification checklist

## Integration with Claude Code

### Capability Execution Flow

1. **Invocation**: User runs `/meta project-bootstrap` or `/meta "bootstrap new project"`

2. **Interactive Setup**:
   ```
   Claude: I'll help bootstrap your project. Let me gather some information:

   1. Project type: [web, cli, library, api, system, other]
   2. Primary language: [detected from files, or prompt]
   3. Project purpose: [one-sentence description]
   4. License: [MIT, Apache-2.0, GPL-3.0, BSD-3-Clause, Proprietary]
   ```

3. **Detection & Analysis**:
   - Scan directory for existing files
   - Detect language from manifest files
   - Analyze project structure
   - Preserve existing documentation

4. **Generation**:
   - Create docs/core/plan.md (customized)
   - Create docs/core/principles.md (language-specific)
   - Create CLAUDE.md (with actual commands)
   - Create/enhance README.md
   - Create .gitignore (language-appropriate)
   - Create LICENSE (user selection)
   - Create directory structure

5. **Verification**:
   - Run exit criteria checklist
   - Report completion status
   - Provide next steps

6. **Output Summary**:
   ```
   ✅ Phase 0 Bootstrap Complete

   Created:
   - docs/core/plan.md (78 lines)
   - docs/core/principles.md (52 lines)
   - CLAUDE.md (145 lines)
   - README.md (89 lines)
   - .gitignore (25 lines)
   - LICENSE (MIT)
   - Directory structure (6 directories)

   Next Steps:
   1. Review and customize docs/core/plan.md
   2. Add project-specific constraints to docs/core/principles.md
   3. Test commands in CLAUDE.md
   4. Commit Phase 0 documentation: git add docs/ CLAUDE.md README.md .gitignore LICENSE
   5. Start Phase 1 development
   ```

### Error Handling

**Existing documentation detected**:
```
⚠️  Found existing documentation:
- README.md (234 lines)
- docs/design.md (156 lines)

Options:
1. Preserve existing, create only missing Phase 0 docs
2. Create Phase 0 structure, backup existing docs to docs/archive/
3. Cancel bootstrap

Your choice: [1-3]
```

**Insufficient information**:
```
❌ Cannot determine project type automatically.

Please specify:
- Project type: [web/cli/library/api/system/other]
- Primary language: [go/rust/python/javascript/typescript/java/cpp/ruby/csharp]

Or provide more context by adding source files or manifest files (package.json, Cargo.toml, go.mod, etc.)
```

## Output Format

### Summary Report

```markdown
# Phase 0 Bootstrap Report

**Project**: [project-name]
**Type**: [detected-type]
**Language**: [detected-language]
**Date**: YYYY-MM-DD

## Created Files

- [x] docs/core/plan.md (78 lines) - Development roadmap
- [x] docs/core/principles.md (52 lines) - Design constraints
- [x] CLAUDE.md (145 lines) - Development guide
- [x] README.md (89 lines) - Public entry point
- [x] .gitignore (25 lines) - Git ignore rules
- [x] LICENSE (21 lines) - MIT License

## Created Directories

- [x] docs/core/
- [x] docs/guides/
- [x] docs/reference/
- [x] docs/tutorials/
- [x] docs/architecture/adr/
- [x] docs/methodology/
- [x] docs/archive/

## Verification

### Exit Criteria
- [x] docs/core/plan.md exists with Phase 0-2 outline
- [x] docs/core/principles.md exists with core constraints
- [x] CLAUDE.md exists with project goal and commands
- [x] README.md has project description
- [x] Directory structure created
- [ ] Code compiles (no source code yet)
- [ ] Basic tests pass (no tests yet)

### Quality Checks
- [x] All required files created
- [x] No broken links in documentation
- [x] Language-specific best practices applied
- [x] Exit criteria documented in plan.md

## Next Steps

1. **Review Documentation** (15-30 minutes)
   - Read docs/core/plan.md and customize phases
   - Review docs/core/principles.md and add project-specific constraints
   - Verify CLAUDE.md commands work in your environment

2. **Initialize Git** (if not done)
   ```bash
   git init
   git add .
   git commit -m "docs: initialize Phase 0 documentation structure"
   ```

3. **Start Development** (Phase 1)
   - Follow plan.md Phase 1 tasks
   - Refer to principles.md for constraints
   - Use CLAUDE.md as development reference

4. **Evolve Documentation**
   - Update plan.md with progress (continuous)
   - Add FAQ entries to CLAUDE.md as questions arise
   - Create guides in docs/guides/ when patterns emerge

## Resources

- **Methodology**: This bootstrap follows Documentation Management Methodology v5.0
- **Claude Code Guide**: CLAUDE.md contains development commands
- **Roadmap**: docs/core/plan.md tracks project progress
```

## Methodology Compliance

**Documentation Management Methodology v5.0 Requirements**:

✅ **Phase 0 Core Requirements**:
- plan.md (Core Document #1) - Generated with customization
- principles.md (Core Document #2) - Generated with language specifics
- CLAUDE.md - Generated with working commands
- README.md - Generated with project context
- Basic files (.gitignore, LICENSE) - Generated appropriately
- Directory structure - Created completely

✅ **Core Principles Applied**:
- **DRY**: Single source of truth, templates link rather than duplicate
- **Progressive Disclosure**: README → CLAUDE.md → plan.md hierarchy
- **Task-Oriented**: Documentation organized by developer goals
- **Living Documentation**: Initialized sections for continuous updates

✅ **Anti-Patterns Avoided**:
- ❌ Mega-README: Kept under 200 lines, extracted to guides
- ❌ Premature Organization: Only Phase 0 essentials created
- ❌ Documentation Drift: Templates include update triggers
- ❌ Redundant Documentation: Templates use linking strategy
- ❌ Skipping Core Docs: plan.md and principles.md always created

## Success Metrics

**Phase 0 completion measured by**:
- All 6 required files created and verified
- Directory structure complete (7 directories)
- Exit criteria checklist passed
- Build succeeds (if code exists)
- No broken links in generated documentation
- Language-specific best practices applied

**Long-term success indicators**:
- plan.md updated regularly (high access count)
- principles.md referenced during design decisions
- CLAUDE.md FAQ grows with real developer questions
- README.md remains under 500 lines as project grows
- Documentation stays synchronized with code

---

**Capability Version**: 1.0
**Last Updated**: 2025-10-12
**Methodology Version**: 5.0
**Status**: Production-ready
