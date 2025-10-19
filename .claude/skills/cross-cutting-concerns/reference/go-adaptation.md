# Go-Specific Adaptations
- Error wrapping: fmt.Errorf("context: %w", err)
- Logging: slog (structured logging)
- Config: viper or env vars
- Middleware: net/http middleware pattern
