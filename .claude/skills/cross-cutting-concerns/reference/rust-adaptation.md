# Rust-Specific Adaptations
- Error handling: Result<T, E> with thiserror/anyhow
- Logging: tracing crate (structured)
- Config: config-rs or figment
- Error wrapping: context() from anyhow
