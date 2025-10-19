# Go Configuration Management Best Practices

**Domain**: Cross-Cutting Concerns
**Language**: Go
**Topic**: Configuration
**Status**: VALIDATED (Iteration 2)

---

## Overview

This document captures best practices for configuration management in Go applications, following 12-Factor App principles and Go community standards.

---

## 1. Store Config in Environment Variables

**Practice**: Use environment variables for all configuration (12-Factor App principle III)

**Rationale**:
- Strict separation of config and code
- Easy to change between deployments
- No accidental config commits
- Works with Docker, Kubernetes, systemd
- Language and OS agnostic

**Example**:
```go
// ✓ Good: Environment variables
logLevel := os.Getenv("META_CC_LOG_LEVEL")
dbURL := os.Getenv("DATABASE_URL")

// ✗ Bad: Hardcoded config
const logLevel = "DEBUG" // Wrong!
const dbURL = "localhost:5432" // Wrong!

// ✗ Bad: Config files (adds complexity)
cfg := loadConfigFile("config.yaml") // Avoid if possible
```

---

## 2. Use Consistent Naming Convention

**Practice**: Prefix all app-specific config with app name, use uppercase with underscores

**Format**: `APP_NAME_COMPONENT_PROPERTY`

**Example**:
```bash
# ✓ Good: Clear, namespaced
META_CC_LOG_LEVEL=DEBUG
META_CC_LOG_FORMAT=json
META_CC_CACHE_DIR=/tmp/meta-cc

# ✗ Bad: No prefix (name conflicts)
LOG_LEVEL=DEBUG

# ✗ Bad: Inconsistent casing
meta_cc_log_level=DEBUG
```

**Exceptions**:
```bash
# External systems (don't use your prefix)
DATABASE_URL=postgres://...
REDIS_URL=redis://...

# Standard environment variables
PATH=/usr/bin
HOME=/home/user
```

---

## 3. Centralize Configuration in a Config Struct

**Practice**: Define single `Config` struct, load once on startup

**Rationale**:
- Single source of truth
- Type-safe access
- Easy to test
- Self-documenting

**Anti-Pattern** (Scattered os.Getenv):
```go
// ✗ Bad: Config scattered throughout code
func initLogger() {
    level := os.Getenv("LOG_LEVEL") // Here
    // ...
}

func initCache() {
    dir := os.Getenv("CACHE_DIR") // And here
    // ...
}

func connectDB() {
    url := os.Getenv("DB_URL") // And here
    // ...
}
```

**Correct Pattern** (Centralized):
```go
// ✓ Good: Centralized config
package config

type Config struct {
    Log   LogConfig
    Cache CacheConfig
    DB    DBConfig
}

type LogConfig struct {
    Level  string
    Format string
}

type CacheConfig struct {
    Dir string
    TTL int
}

type DBConfig struct {
    URL string
}

func Load() (*Config, error) {
    cfg := &Config{
        Log:   loadLogConfig(),
        Cache: loadCacheConfig(),
        DB:    loadDBConfig(),
    }

    if err := cfg.Validate(); err != nil {
        return nil, err
    }

    return cfg, nil
}

// Usage in main.go
cfg, err := config.Load()
if err != nil {
    log.Fatalf("config error: %v", err)
}

initLogger(cfg.Log)
initCache(cfg.Cache)
connectDB(cfg.DB)
```

---

## 4. Validate Configuration on Startup (Fail-Fast)

**Practice**: Load and validate ALL config on startup, fail immediately if invalid

**Rationale**:
- Catch errors early
- Clear error messages
- Prevents runtime surprises
- Follows fail-fast principle

**Example**:
```go
func (c *Config) Validate() error {
    // Validate log level
    validLevels := []string{"DEBUG", "INFO", "WARN", "ERROR"}
    if !contains(validLevels, c.Log.Level) {
        return fmt.Errorf("invalid LOG_LEVEL: %s (must be one of: %v)",
            c.Log.Level, validLevels)
    }

    // Validate required fields
    if c.DB.URL == "" {
        return errors.New("DATABASE_URL is required")
    }

    // Validate ranges
    if c.Cache.TTL < 0 || c.Cache.TTL > 86400 {
        return fmt.Errorf("CACHE_TTL must be between 0 and 86400, got: %d",
            c.Cache.TTL)
    }

    return nil
}

// In main.go
func main() {
    cfg, err := config.Load()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
        os.Exit(1) // Fail fast!
    }

    // ... rest of application ...
}
```

---

## 5. Provide Sensible Defaults for Non-Sensitive Values

**Practice**: All non-sensitive config should have reasonable defaults

**Required (No Defaults)**:
- Sensitive data (API keys, passwords, tokens)
- Environment-specific values (database URLs)
- Values that must be explicitly set

**Optional (With Defaults)**:
- Log level (default: INFO)
- Timeouts (default: 30s)
- Buffer sizes (default: 8KB)
- Port numbers (default: 8080)

**Example**:
```go
// ✓ Good: Has default
cfg.Log.Level = getEnv("META_CC_LOG_LEVEL", "INFO")
cfg.HTTP.Port = getEnvInt("META_CC_HTTP_PORT", 8080)
cfg.HTTP.Timeout = getEnvInt("META_CC_HTTP_TIMEOUT", 30)

// ✓ Good: Required, no default (validation will catch)
cfg.API.Key = os.Getenv("META_CC_API_KEY")
if cfg.API.Key == "" {
    return errors.New("META_CC_API_KEY is required")
}

// ✗ Bad: Sensitive data with default
cfg.API.Key = getEnv("API_KEY", "default-key") // NEVER!
```

---

## 6. Write Helpful Validation Error Messages

**Practice**: Errors should explain what's wrong AND how to fix it

**Example**:
```go
// ✓ Good: Explains problem and solution
return fmt.Errorf("invalid LOG_LEVEL: %s (must be one of: DEBUG, INFO, WARN, ERROR)",
    cfg.Log.Level)

// ✓ Good: Suggests action
return fmt.Errorf("DATABASE_URL not set. Set it with: export DATABASE_URL=postgres://...")

// ✓ Good: Context-aware
if cfg.Feature.Enabled && cfg.Feature.APIKey == "" {
    return errors.New("FEATURE_API_KEY required when FEATURE_ENABLED=true")
}

// ✗ Bad: Vague
return errors.New("invalid config")

// ✗ Bad: No guidance
return fmt.Errorf("LOG_LEVEL is invalid: %s", cfg.Log.Level)
```

---

## 7. Use Struct Tags for Documentation

**Practice**: Document config fields with struct tags and comments

**Example**:
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

    // API key for authentication.
    // Required: Yes
    // Sensitive: Yes
    APIKey string `env:"META_CC_API_KEY" required:"true" sensitive:"true"`
}
```

---

## 8. Maintain Configuration Reference Documentation

**Practice**: Keep `docs/configuration.md` with complete env var reference

**Structure**:
```markdown
# Configuration Reference

## META_CC_LOG_LEVEL

**Description**: Controls the minimum log level for output

**Type**: String (enum)

**Valid Values**: DEBUG, INFO, WARN, ERROR

**Default**: INFO

**Required**: No

**Example**:
\`\`\`bash
export META_CC_LOG_LEVEL=DEBUG
\`\`\`

---

## DATABASE_URL

**Description**: PostgreSQL connection string

**Type**: String (URL)

**Format**: postgres://user:password@host:port/database

**Default**: None

**Required**: Yes

**Sensitive**: Yes (contains password)

**Example**:
\`\`\`bash
export DATABASE_URL=postgres://user:pass@localhost:5432/mydb
\`\`\`
```

---

## 9. Support Environment-Specific Configuration

**Practice**: Use different configs for dev/test/prod, never commit secrets

### Development (.env.development):
```bash
# Development configuration (git-ignored)
META_CC_LOG_LEVEL=DEBUG
META_CC_LOG_FORMAT=text
META_CC_CACHE_DIR=/tmp/meta-cc-dev
DATABASE_URL=postgres://dev:dev@localhost:5432/dev_db
```

### Production (Kubernetes ConfigMap + Secret):
```yaml
# ConfigMap (non-sensitive)
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
data:
  META_CC_LOG_LEVEL: "INFO"
  META_CC_LOG_FORMAT: "json"

---
# Secret (sensitive)
apiVersion: v1
kind: Secret
metadata:
  name: app-secrets
type: Opaque
data:
  DATABASE_URL: <base64-encoded>
  API_KEY: <base64-encoded>
```

---

## 10. Use Helper Functions for Type Conversion

**Practice**: Create helpers for common type conversions with defaults

**Example**:
```go
// String with default
func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

// Integer with default
func getEnvInt(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if intVal, err := strconv.Atoi(value); err == nil {
            return intVal
        }
        // Log warning about invalid integer
        fmt.Fprintf(os.Stderr, "Warning: invalid integer for %s: %s, using default: %d\n",
            key, value, defaultValue)
    }
    return defaultValue
}

// Boolean with default
func getEnvBool(key string, defaultValue bool) bool {
    if value := os.Getenv(key); value != "" {
        lower := strings.ToLower(value)
        return lower == "true" || lower == "1" || lower == "yes" || lower == "on"
    }
    return defaultValue
}

// Duration with default
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
    if value := os.Getenv(key); value != "" {
        if duration, err := time.ParseDuration(value); err == nil {
            return duration
        }
        fmt.Fprintf(os.Stderr, "Warning: invalid duration for %s: %s, using default: %v\n",
            key, value, defaultValue)
    }
    return defaultValue
}
```

---

## 11. Never Commit Secrets to Version Control

**Practice**: Use environment variables for secrets, git-ignore .env files

**.gitignore**:
```gitignore
# Environment files (may contain secrets)
.env
.env.local
.env.development
.env.production

# Sensitive config
config.local.yaml
secrets.yaml
```

**Safe Practices**:
- ✅ Use environment variables for secrets
- ✅ Use secret management tools (Vault, AWS Secrets Manager)
- ✅ Use .env files (git-ignored) for local development
- ✅ Use .env.example (committed) as template

**.env.example** (committed):
```bash
# Example configuration (copy to .env and fill in)
META_CC_LOG_LEVEL=INFO
META_CC_API_KEY=your-api-key-here
DATABASE_URL=postgres://user:password@localhost:5432/db
```

**.env** (git-ignored, actual values):
```bash
META_CC_LOG_LEVEL=DEBUG
META_CC_API_KEY=actual-secret-key-123
DATABASE_URL=postgres://dev:devpass@localhost:5432/dev_db
```

---

## 12. Test Configuration Loading and Validation

**Practice**: Test valid configs, invalid configs, defaults, edge cases

**Example**:
```go
func TestConfigLoad(t *testing.T) {
    // Test valid config
    os.Setenv("META_CC_LOG_LEVEL", "DEBUG")
    defer os.Unsetenv("META_CC_LOG_LEVEL")

    cfg, err := config.Load()
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    if cfg.Log.Level != "DEBUG" {
        t.Errorf("expected DEBUG, got %s", cfg.Log.Level)
    }
}

func TestConfigValidation(t *testing.T) {
    tests := []struct {
        name    string
        env     map[string]string
        wantErr bool
        errMsg  string
    }{
        {
            name:    "valid",
            env:     map[string]string{"LOG_LEVEL": "INFO"},
            wantErr: false,
        },
        {
            name:    "invalid log level",
            env:     map[string]string{"LOG_LEVEL": "INVALID"},
            wantErr: true,
            errMsg:  "invalid LOG_LEVEL",
        },
        {
            name:    "missing required",
            env:     map[string]string{},
            wantErr: true,
            errMsg:  "API_KEY is required",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup
            for k, v := range tt.env {
                os.Setenv(k, v)
                defer os.Unsetenv(k)
            }

            _, err := config.Load()

            if (err != nil) != tt.wantErr {
                t.Errorf("wantErr %v, got %v", tt.wantErr, err)
            }

            if tt.wantErr && tt.errMsg != "" {
                if !strings.Contains(err.Error(), tt.errMsg) {
                    t.Errorf("error should contain %q, got: %v", tt.errMsg, err)
                }
            }
        })
    }
}

func TestConfigDefaults(t *testing.T) {
    os.Clearenv()

    cfg, err := config.Load()
    if err != nil {
        t.Fatalf("should load with defaults: %v", err)
    }

    if cfg.Log.Level != "INFO" {
        t.Errorf("default should be INFO, got %s", cfg.Log.Level)
    }
}
```

---

## 13. Isolate Tests with Environment Cleanup

**Practice**: Always clean up environment variables in tests

**Example**:
```go
// Helper function for test cleanup
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

// Usage
func TestWithEnv(t *testing.T) {
    setEnvForTest(t, "META_CC_LOG_LEVEL", "DEBUG")
    // Test automatically cleans up after
}
```

---

## 14. Support Backward Compatibility During Migration

**Practice**: Support old env vars for one release cycle with deprecation warning

**Example**:
```go
func loadLogLevel() string {
    // Try new variable first
    if level := os.Getenv("META_CC_LOG_LEVEL"); level != "" {
        return level
    }

    // Fall back to old variable (deprecated)
    if oldLevel := os.Getenv("LOG_LEVEL"); oldLevel != "" {
        fmt.Fprintf(os.Stderr,
            "Warning: LOG_LEVEL is deprecated, use META_CC_LOG_LEVEL instead\n")
        return oldLevel
    }

    return "INFO" // Default
}
```

---

## Common Anti-Patterns

### 1. Scattered os.Getenv Calls

❌ Problem: No single source of truth, hard to test

```go
// Bad: Config scattered everywhere
func initLogger() {
    level := os.Getenv("LOG_LEVEL")
    format := os.Getenv("LOG_FORMAT")
    // ...
}

func initDB() {
    url := os.Getenv("DB_URL")
    // ...
}

// Good: Centralized config
cfg, _ := config.Load()
initLogger(cfg.Log)
initDB(cfg.DB)
```

### 2. No Validation (Silent Failures)

❌ Problem: Runtime errors instead of startup failures

```go
// Bad: No validation
cfg.Port = getEnvInt("PORT", 8080)
// Later: port = -1 causes runtime error

// Good: Validate on load
if cfg.Port < 1 || cfg.Port > 65535 {
    return fmt.Errorf("PORT must be 1-65535, got: %d", cfg.Port)
}
```

### 3. Hardcoded Configuration

❌ Problem: Can't change without recompiling

```go
// Bad: Hardcoded
const apiURL = "https://api.example.com"
const timeout = 30 * time.Second

// Good: Environment variables
apiURL := getEnv("API_URL", "https://api.example.com")
timeout := getEnvDuration("TIMEOUT", 30*time.Second)
```

### 4. Secrets in Code or Config Files

❌ Problem: Security risk, accidental commits

```go
// Bad: Secret in code
const apiKey = "secret-key-123" // NEVER!

// Bad: Secret in committed file
// config.yaml (committed):
// api_key: secret-key-123  # Wrong!

// Good: Environment variable
apiKey := os.Getenv("API_KEY")
if apiKey == "" {
    log.Fatal("API_KEY required")
}
```

### 5. Inconsistent Naming

❌ Problem: Hard to remember, name conflicts

```go
// Bad: Inconsistent
LOG_LEVEL=DEBUG
meta_cc_cache_dir=/tmp
MetaCCTimeout=30

// Good: Consistent
META_CC_LOG_LEVEL=DEBUG
META_CC_CACHE_DIR=/tmp
META_CC_TIMEOUT=30
```

### 6. No Defaults for Optional Values

❌ Problem: Application crashes if not set

```go
// Bad: No default, crashes if not set
port := getEnvInt("PORT") // Returns 0 if not set!
http.ListenAndServe(fmt.Sprintf(":%d", port), nil) // :0 = random port!

// Good: Sensible default
port := getEnvInt("PORT", 8080)
http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
```

---

## References

- [The Twelve-Factor App - Config](https://12factor.net/config)
- [Go Best Practices - Configuration](https://peter.bourgon.org/go-best-practices-2016/#configuration)
- [Environment Variables in Go](https://golang.org/pkg/os/#Getenv)
- [Viper (alternative, but adds complexity)](https://github.com/spf13/viper)
- [envconfig (struct tags approach)](https://github.com/kelseyhightower/envconfig)

---

**Status**: VALIDATED
**Source**: Iteration 2 (Bootstrap-013)
**Validation**: Derived from 12-Factor App principles and Go community standards
