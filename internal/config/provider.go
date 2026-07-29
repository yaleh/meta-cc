package config

import (
	"os"
	"strings"
	"sync"
)

// HostEnv is the narrowly scoped environment contract (DIR-073) the plugin
// manifests use to tell the MCP binary which agent host launched it:
// "claude" for Claude Code, "codex" for the Codex CLI. Manual/standalone
// installations may leave it unset (documented claude fallback) or set it
// explicitly to change the omitted-provider default without touching query
// arguments.
const HostEnv = "META_CC_HOST"

// Validated META_CC_HOST values — one per supported agent host.
const (
	HostClaude = "claude"
	HostCodex  = "codex"
)

// StandaloneHostFallback is the deterministic default for processes with no
// META_CC_HOST: the pre-DIR-073 Claude-only behavior, preserved so manual
// MCP installations keep working unchanged.
const StandaloneHostFallback = HostClaude

// ProviderConfig holds the resolved host/default-provider pair.
type ProviderConfig struct {
	// Host is the normalized META_CC_HOST value ("" when unset).
	Host string

	// Default is the provider name every provider-aware tool resolves an
	// omitted/empty `provider` argument to: Host when set, otherwise
	// StandaloneHostFallback. Always a valid provider name (validated by
	// Config.Validate).
	Default string
}

// loadProviderConfig reads and normalizes META_CC_HOST. Validation happens
// in Config.Validate so an invalid host fails the whole Load() visibly.
func loadProviderConfig() ProviderConfig {
	host := strings.ToLower(strings.TrimSpace(os.Getenv(HostEnv)))
	def := host
	if def == "" {
		def = StandaloneHostFallback
	}
	return ProviderConfig{Host: host, Default: def}
}

var (
	processMu      sync.RWMutex
	processDefault = StandaloneHostFallback
)

// OmittedProviderDefault is the single source of truth every provider-aware
// tool (executor handlers, pipeline, stage-1 discovery, raw-file selection,
// analysis dispatch, query_sessions) uses to resolve an omitted/empty
// `provider` argument. It returns the value Load() validated and published
// at startup, or StandaloneHostFallback before/without Load().
func OmittedProviderDefault() string {
	processMu.RLock()
	defer processMu.RUnlock()
	return processDefault
}

// applyProcessDefault publishes the validated default after a successful
// Load(). Kept unexported: production processes get exactly one source of
// truth (the validated config); tests use SwapProcessDefault.
func applyProcessDefault(def string) {
	processMu.Lock()
	processDefault = def
	processMu.Unlock()
}

// SwapProcessDefault replaces the process-wide omitted-provider default and
// returns a restore function. It is the override seam for tests and for
// embedding processes that manage configuration outside Load().
func SwapProcessDefault(def string) (restore func()) {
	processMu.Lock()
	prev := processDefault
	processDefault = def
	processMu.Unlock()
	return func() { applyProcessDefault(prev) }
}
