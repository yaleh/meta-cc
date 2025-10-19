# Configuration Management Conventions for meta-cc

**Version**: 1.0
**Status**: PROPOSED
**Created**: 2025-10-17 (Iteration 2)
**Applies To**: All meta-cc Go code

---

## 1. Configuration Standard

**Decision**: Use environment variables with centralized Config struct

**Rationale**:
- ✅ Follows 12-Factor App principles
- ✅ Simple (no external dependencies)
- ✅ Standard for CLI tools and servers
- ✅ Works with Docker, Kubernetes, systemd
- ✅ Easy to override in different environments

**Alternatives Considered**:
- **Config files** (YAML/TOML): Adds complexity, requires file management
- **Command-line flags**: Verbose, harder to manage in deployment
- **viper library**: Overkill for simple CLI tool, adds dependency

---

## 2. Environment Variable Naming

### 2.1 Naming Convention

**Convention**: Use `META_CC_` prefix for all meta-cc configuration

**Format**: `META_CC_<COMPONENT>_<PROPERTY>`

**Rules**:
1. All uppercase
2. Underscore separation
3. `META_CC_` prefix for namespacing
4. Component name (optional): `LOG`, `CACHE`, `OUTPUT`
5. Property name: `LEVEL`, `FORMAT`, `ENABLED`

**Examples**:
```bash
# ✓ Good: Clear, namespaced
META_CC_LOG_LEVEL=DEBUG
META_CC_LOG_FORMAT=json
META_CC_CACHE_DIR=/tmp/meta-cc-cache
META_CC_OUTPUT_MODE=inline

# ✗ Bad: No prefix (conflicts possible)
LOG_LEVEL=DEBUG

# ✗ Bad: Inconsistent casing
meta_cc_log_level=DEBUG

# ✗ Bad: Unclear component
META_CC_LEVEL=DEBUG
```

### 2.2 Standard Variables

**Core Configuration**:
```bash
# Logging
META_CC_LOG_LEVEL=INFO              # DEBUG, INFO, WARN, ERROR
META_CC_LOG_FORMAT=text             # text, json
META_CC_LOGGING_ENABLED=true        # true, false

# Output
META_CC_OUTPUT_MODE=auto            # auto, inline, file_ref
META_CC_INLINE_THRESHOLD=8192       # Bytes before switching to file_ref

# Capability Sources
META_CC_CAPABILITY_SOURCES=/path/to/capabilities  # Colon-separated paths

# Session/Project (set by Claude Code)
CC_SESSION_ID=uuid                  # Session UUID
CC_PROJECT_HASH=hash                # Project path hash
```

**External Dependencies** (no META_CC prefix):
```bash
# Claude Code sets these (not our convention)
CLAUDE_CODE_SESSION_ID=session-id
CC_SESSION_ID=uuid
CC_PROJECT_HASH=hash
```

---

## 3. Centralized Config Structure

### 3.1 Config Package Pattern

**Convention**: Create `internal/config` package with centralized Config struct

**Structure**:
```
internal/
  config/
    config.go          # Config struct and loading
    config_test.go     # Config tests
    validation.go      # Validation logic
```

### 3.2 Config Struct Design

**Pattern**:
```go
package config

import (
    "fmt"
    "log/slog"
    "os"
    "strconv"
)

// Config holds all application configuration
type Config struct {
    // Logging configuration
    Log LogConfig

    // Output configuration
    Output OutputConfig

    // Capability configuration
    Capability CapabilityConfig

    // Session configuration (from Claude Code)
    Session SessionConfig
}

// LogConfig holds logging-related configuration
type LogConfig struct {
    Level   slog.Level `env:"META_CC_LOG_LEVEL" default:"INFO"`
    Format  string     `env:"META_CC_LOG_FORMAT" default:"text"` // text, json
    Enabled bool       `env:"META_CC_LOGGING_ENABLED" default:"true"`
}

// OutputConfig holds output-related configuration
type OutputConfig struct {
    Mode            string `env:"META_CC_OUTPUT_MODE" default:"auto"` // auto, inline, file_ref
    InlineThreshold int    `env:"META_CC_INLINE_THRESHOLD" default:"8192"`
}

// CapabilityConfig holds capability source configuration
type CapabilityConfig struct {
    Sources string `env:"META_CC_CAPABILITY_SOURCES" default:""`
}

// SessionConfig holds session information from Claude Code
type SessionConfig struct {
    SessionID   string `env:"CC_SESSION_ID"`
    ProjectHash string `env:"CC_PROJECT_HASH"`
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
    cfg := &Config{
        Log:        loadLogConfig(),
        Output:     loadOutputConfig(),
        Capability: loadCapabilityConfig(),
        Session:    loadSessionConfig(),
    }

    // Validate configuration
    if err := cfg.Validate(); err != nil {
        return nil, fmt.Errorf("invalid configuration: %w", err)
    }

    return cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
    // Validate log level
    validLevels := []string{"DEBUG", "INFO", "WARN", "ERROR"}
    if !contains(validLevels, c.Log.LevelString()) {
        return fmt.Errorf("invalid LOG_LEVEL: %s (must be one of: %v)",
            c.Log.LevelString(), validLevels)
    }

    // Validate log format
    if c.Log.Format != "text" && c.Log.Format != "json" {
        return fmt.Errorf("invalid LOG_FORMAT: %s (must be 'text' or 'json')",
            c.Log.Format)
    }

    // Validate output mode
    validModes := []string{"auto", "inline", "file_ref"}
    if !contains(validModes, c.Output.Mode) {
        return fmt.Errorf("invalid OUTPUT_MODE: %s (must be one of: %v)",
            c.Output.Mode, validModes)
    }

    // Validate inline threshold (must be positive)
    if c.Output.InlineThreshold <= 0 {
        return fmt.Errorf("INLINE_THRESHOLD must be positive, got: %d",
            c.Output.InlineThreshold)
    }

    return nil
}

// loadLogConfig loads logging configuration from environment
func loadLogConfig() LogConfig {
    cfg := LogConfig{
        Level:   slog.LevelInfo, // Default
        Format:  getEnvOrDefault("META_CC_LOG_FORMAT", "text"),
        Enabled: getEnvBool("META_CC_LOGGING_ENABLED", true),
    }

    // Parse log level
    if levelStr := os.Getenv("META_CC_LOG_LEVEL"); levelStr != "" {
        cfg.Level = parseLogLevel(levelStr)
    }

    return cfg
}

// loadOutputConfig loads output configuration from environment
func loadOutputConfig() OutputConfig {
    return OutputConfig{
        Mode:            getEnvOrDefault("META_CC_OUTPUT_MODE", "auto"),
        InlineThreshold: getEnvInt("META_CC_INLINE_THRESHOLD", 8192),
    }
}

// loadCapabilityConfig loads capability configuration from environment
func loadCapabilityConfig() CapabilityConfig {
    return CapabilityConfig{
        Sources: os.Getenv("META_CC_CAPABILITY_SOURCES"),
    }
}

// loadSessionConfig loads session configuration from environment
func loadSessionConfig() SessionConfig {
    return SessionConfig{
        SessionID:   os.Getenv("CC_SESSION_ID"),
        ProjectHash: os.Getenv("CC_PROJECT_HASH"),
    }
}

// Helper functions

func getEnvOrDefault(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if intVal, err := strconv.Atoi(value); err == nil {
            return intVal
        }
    }
    return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
    if value := os.Getenv(key); value != "" {
        return value == "true" || value == "1" || value == "yes"
    }
    return defaultValue
}

func parseLogLevel(levelStr string) slog.Level {
    switch levelStr {
    case "DEBUG":
        return slog.LevelDebug
    case "INFO":
        return slog.LevelInfo
    case "WARN":
        return slog.LevelWarn
    case "ERROR":
        return slog.LevelError
    default:
        return slog.LevelInfo
    }
}

func contains(slice []string, item string) bool {
    for _, s := range slice {
        if s == item {
            return true
        }
    }
    return false
}

// LevelString returns the string representation of log level
func (c LogConfig) LevelString() string {
    switch c.Level {
    case slog.LevelDebug:
        return "DEBUG"
    case slog.LevelInfo:
        return "INFO"
    case slog.LevelWarn:
        return "WARN"
    case slog.LevelError:
        return "ERROR"
    default:
        return "INFO"
    }
}
```

---

## 4. Configuration Loading

### 4.1 Fail-Fast Principle

**Convention**: Load and validate configuration on application startup, fail immediately if invalid

**Rationale**:
- Catch configuration errors early
- Prevent runtime surprises
- Clear error messages at startup
- Follows 12-Factor App principle

**Pattern** (in main.go):
```go
func main() {
    // Load configuration (fail-fast)
    cfg, err := config.Load()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
        os.Exit(1)
    }

    // Initialize components with config
    logger := initLogger(cfg.Log)
    // ... rest of initialization ...

    logger.Info("application started",
        "log_level", cfg.Log.LevelString(),
        "log_format", cfg.Log.Format,
    )

    // ... run application ...
}
```

### 4.2 Default Values

**Convention**: Provide sensible defaults for all non-sensitive configuration

**Required (No Defaults)**:
- Sensitive data (API keys, passwords)
- Environment-specific values (database URLs)
- Session IDs (provided by Claude Code)

**Optional (With Defaults)**:
- Log level (default: INFO)
- Log format (default: text)
- Cache directory (default: OS temp dir)
- Timeouts (default: 30s)

**Example**:
```go
// ✓ Good: Has default
cfg.Log.Level = getEnvOrDefault("META_CC_LOG_LEVEL", "INFO")

// ✓ Good: Required, no default (will fail validation if missing)
cfg.API.Key = os.Getenv("META_CC_API_KEY") // Required
if cfg.API.Key == "" {
    return nil, errors.New("META_CC_API_KEY is required")
}

// ✗ Bad: Sensitive data with default
cfg.API.Key = getEnvOrDefault("META_CC_API_KEY", "default-key") // NEVER!
```

---

## 5. Configuration Validation

### 5.1 Validation Strategy

**Convention**: Validate ALL configuration on load, before application runs

**Validation Types**:

1. **Required Field Validation**:
   ```go
   if cfg.Session.SessionID == "" {
       return fmt.Errorf("CC_SESSION_ID is required")
   }
   ```

2. **Enum Validation**:
   ```go
   validLevels := []string{"DEBUG", "INFO", "WARN", "ERROR"}
   if !contains(validLevels, cfg.Log.Level) {
       return fmt.Errorf("invalid LOG_LEVEL: %s", cfg.Log.Level)
   }
   ```

3. **Range Validation**:
   ```go
   if cfg.Output.InlineThreshold <= 0 || cfg.Output.InlineThreshold > 1000000 {
       return fmt.Errorf("INLINE_THRESHOLD must be between 1 and 1000000")
   }
   ```

4. **Format Validation**:
   ```go
   if cfg.Cache.Dir != "" && !filepath.IsAbs(cfg.Cache.Dir) {
       return fmt.Errorf("CACHE_DIR must be absolute path")
   }
   ```

5. **Dependency Validation**:
   ```go
   if cfg.Feature.Enabled && cfg.Feature.APIKey == "" {
       return fmt.Errorf("FEATURE_API_KEY required when FEATURE_ENABLED=true")
   }
   ```

### 5.2 Helpful Error Messages

**Convention**: Validation errors should explain what's wrong AND how to fix it

**Examples**:
```go
// ✓ Good: Explains problem and solution
return fmt.Errorf("invalid LOG_LEVEL: %s (must be one of: DEBUG, INFO, WARN, ERROR)",
    cfg.Log.Level)

// ✓ Good: Suggests action
return fmt.Errorf("CC_SESSION_ID not set. Are you running inside Claude Code?")

// ✗ Bad: Vague
return errors.New("invalid config")

// ✗ Bad: No guidance
return fmt.Errorf("LOG_LEVEL is invalid: %s", cfg.Log.Level)
```

---

## 6. Configuration Documentation

### 6.1 Self-Documenting Structs

**Convention**: Use struct tags and comments to document configuration

**Pattern**:
```go
type Config struct {
    // Log level for application logging.
    // Valid values: DEBUG, INFO, WARN, ERROR
    // Default: INFO
    LogLevel string `env:"META_CC_LOG_LEVEL" default:"INFO"`

    // Log format for output.
    // Valid values: text, json
    // Default: text
    LogFormat string `env:"META_CC_LOG_FORMAT" default:"text"`

    // Enable or disable logging entirely.
    // Valid values: true, false
    // Default: true
    LoggingEnabled bool `env:"META_CC_LOGGING_ENABLED" default:"true"`
}
```

### 6.2 Configuration Reference

**Convention**: Maintain `docs/configuration.md` with complete reference

**Structure**:
```markdown
# Configuration Reference

## Environment Variables

### META_CC_LOG_LEVEL

**Description**: Controls the minimum log level for output

**Type**: String

**Valid Values**: DEBUG, INFO, WARN, ERROR

**Default**: INFO

**Example**:
bash
export META_CC_LOG_LEVEL=DEBUG


### META_CC_LOG_FORMAT

**Description**: Format for log output

**Type**: String

**Valid Values**: text (human-readable), json (structured)

**Default**: text

**Example**:
bash
export META_CC_LOG_FORMAT=json

```

---

## 7. Configuration Testing

### 7.1 Test Loading and Validation

**Convention**: Test both valid and invalid configurations

**Pattern**:
```go
func TestConfigLoad(t *testing.T) {
    // Test valid config
    os.Setenv("META_CC_LOG_LEVEL", "DEBUG")
    os.Setenv("META_CC_LOG_FORMAT", "json")
    defer os.Unsetenv("META_CC_LOG_LEVEL")
    defer os.Unsetenv("META_CC_LOG_FORMAT")

    cfg, err := config.Load()
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    if cfg.Log.Level != slog.LevelDebug {
        t.Errorf("expected DEBUG level, got %v", cfg.Log.Level)
    }

    if cfg.Log.Format != "json" {
        t.Errorf("expected json format, got %s", cfg.Log.Format)
    }
}

func TestConfigValidation(t *testing.T) {
    // Test invalid log level
    os.Setenv("META_CC_LOG_LEVEL", "INVALID")
    defer os.Unsetenv("META_CC_LOG_LEVEL")

    _, err := config.Load()
    if err == nil {
        t.Fatal("expected validation error for invalid log level")
    }

    if !strings.Contains(err.Error(), "invalid LOG_LEVEL") {
        t.Errorf("error should mention invalid LOG_LEVEL: %v", err)
    }
}

func TestConfigDefaults(t *testing.T) {
    // Clear all config env vars
    os.Clearenv()

    cfg, err := config.Load()
    if err != nil {
        t.Fatalf("should load with defaults: %v", err)
    }

    // Verify defaults
    if cfg.Log.Level != slog.LevelInfo {
        t.Errorf("default log level should be INFO, got %v", cfg.Log.Level)
    }

    if cfg.Log.Format != "text" {
        t.Errorf("default log format should be text, got %s", cfg.Log.Format)
    }

    if !cfg.Log.Enabled {
        t.Error("logging should be enabled by default")
    }
}
```

### 7.2 Test Environment Isolation

**Convention**: Always clean up environment variables in tests

**Pattern**:
```go
func TestWithCleanEnv(t *testing.T) {
    // Save original env
    original := os.Getenv("META_CC_LOG_LEVEL")
    defer func() {
        if original != "" {
            os.Setenv("META_CC_LOG_LEVEL", original)
        } else {
            os.Unsetenv("META_CC_LOG_LEVEL")
        }
    }()

    // Test with modified env
    os.Setenv("META_CC_LOG_LEVEL", "DEBUG")
    // ... run test ...
}

// Or use helper function
func setEnvForTest(t *testing.T, key, value string) {
    t.Helper()
    original := os.Getenv(key)
    t.Cleanup(func() {
        if original != "" {
            os.Setenv(key, original)
        } else {
            os.Unsetenv(key)
        }
    })
    os.Setenv(key, value)
}
```

---

## 8. Configuration Migration

### 8.1 Backward Compatibility

**Convention**: Support old environment variables for one release cycle

**Pattern**:
```go
func loadLogConfig() LogConfig {
    // Try new variable first
    logLevel := os.Getenv("META_CC_LOG_LEVEL")

    // Fall back to old variable (deprecated)
    if logLevel == "" {
        if oldLevel := os.Getenv("LOG_LEVEL"); oldLevel != "" {
            fmt.Fprintf(os.Stderr, "Warning: LOG_LEVEL is deprecated, use META_CC_LOG_LEVEL\n")
            logLevel = oldLevel
        }
    }

    // Default
    if logLevel == "" {
        logLevel = "INFO"
    }

    return LogConfig{
        Level: parseLogLevel(logLevel),
    }
}
```

### 8.2 Migration Guide

**Convention**: Document migration path in CHANGELOG and docs

**Example**:
```markdown
## Migration Guide: v1.0 to v2.0

### Environment Variable Changes

The following environment variables have been renamed:

- `LOG_LEVEL` → `META_CC_LOG_LEVEL`
- `LOG_FORMAT` → `META_CC_LOG_FORMAT`
- `CACHE_DIR` → `META_CC_CACHE_DIR`

**Action Required**: Update your environment configuration before v2.1 (old names will be removed)

**Backward Compatibility**: v2.0 supports both old and new names, but old names will be removed in v2.1
```

---

## 9. Configuration in Different Environments

### 9.1 Development

**Convention**: Use `.env.development` file (ignored by git)

**Example** (`.env.development`):
```bash
# Development configuration
META_CC_LOG_LEVEL=DEBUG
META_CC_LOG_FORMAT=text
META_CC_CAPABILITY_SOURCES=capabilities/commands

# Session variables (simulated for testing)
CC_SESSION_ID=dev-session-123
CC_PROJECT_HASH=-home-user-projects-test
```

**Load with**:
```bash
# Using export
export $(cat .env.development | xargs)

# Or using direnv (recommended)
# Create .envrc:
# source .env.development
direnv allow
```

### 9.2 Production

**Convention**: Set environment variables via deployment system

**Examples**:

**Docker**:
```dockerfile
ENV META_CC_LOG_LEVEL=INFO
ENV META_CC_LOG_FORMAT=json
```

**Kubernetes**:
```yaml
env:
  - name: META_CC_LOG_LEVEL
    value: "INFO"
  - name: META_CC_LOG_FORMAT
    value: "json"
```

**systemd**:
```ini
[Service]
Environment="META_CC_LOG_LEVEL=INFO"
Environment="META_CC_LOG_FORMAT=json"
```

### 9.3 Testing

**Convention**: Override config in tests, don't modify environment

**Pattern**:
```go
func TestWithCustomConfig(t *testing.T) {
    cfg := &config.Config{
        Log: config.LogConfig{
            Level:   slog.LevelDebug,
            Format:  "text",
            Enabled: true,
        },
    }

    // Pass config to components
    logger := initLogger(cfg.Log)
    // ... test with custom config ...
}
```

---

## 10. Best Practices Summary

### Must-Follow Patterns:

1. ✅ **Use META_CC_ prefix** for all configuration
2. ✅ **Centralize in Config struct** (no scattered os.Getenv)
3. ✅ **Validate on startup** (fail-fast)
4. ✅ **Provide sensible defaults** (for non-sensitive values)
5. ✅ **Document in struct tags** (self-documenting)
6. ✅ **Test validation** (valid and invalid cases)
7. ✅ **Clear error messages** (explain what and how to fix)
8. ✅ **Never commit secrets** (use env vars, not code)

### Anti-Patterns to Avoid:

1. ❌ Scattered os.Getenv calls throughout code
2. ❌ No validation (runtime failures instead of startup)
3. ❌ Hardcoded defaults (magic values in code)
4. ❌ Silent failures (missing required config)
5. ❌ Inconsistent naming (LOG_LEVEL vs META_CC_LOG_LEVEL)
6. ❌ Secrets in code or config files
7. ❌ No documentation (undiscoverable configuration)
8. ❌ Missing tests (untested edge cases)

---

## Example: Complete Config Package

See `knowledge/templates/config-management-template.go` for complete, copy-paste ready implementation.

---

**Status**: PROPOSED (Iteration 2)
**Next**: Implement in codebase (Iteration 2-3)
**Generated By**: convention-definer (Bootstrap-013)
