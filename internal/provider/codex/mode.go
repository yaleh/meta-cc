package codex

import (
	"fmt"
	"os"
)

// Mode selects how the Codex provider resolves conversation history.
type Mode string

const (
	// ModeAuto negotiates with an installed `codex app-server`, falling
	// back to the SQLite/rollout files backend on absence, timeout,
	// incompatible version, or repeated process failure (circuit
	// breaker). This is the default.
	ModeAuto Mode = "auto"
	// ModeAppServer requires the app-server backend and fails clearly
	// (rather than silently using files) when it is unavailable.
	ModeAppServer Mode = "app_server"
	// ModeFiles uses only the existing SQLite/rollout files backend,
	// matching pre-DIR-029 behavior exactly. Useful offline or for
	// forensic access to raw session state.
	ModeFiles Mode = "files"
)

// backendModeEnvVar selects the Codex history backend mode. Unset or empty
// defaults to ModeAuto.
const backendModeEnvVar = "META_CC_CODEX_BACKEND"

// ResolveMode reads backendModeEnvVar and returns the requested Mode, or an
// error naming the invalid value. Exported as a standalone function (not
// only reachable via NewProvider) so callers/tests can validate
// configuration independently of constructing a Provider.
func ResolveMode() (Mode, error) {
	return parseMode(os.Getenv(backendModeEnvVar))
}

func parseMode(raw string) (Mode, error) {
	switch Mode(raw) {
	case "", ModeAuto:
		return ModeAuto, nil
	case ModeAppServer, ModeFiles:
		return Mode(raw), nil
	default:
		return "", fmt.Errorf("invalid %s %q: must be \"auto\", \"app_server\", or \"files\"", backendModeEnvVar, raw)
	}
}
