package appserver

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// MinSupportedVersion is the lowest Codex CLI version DIR-029 has verified
// exposes the stable thread/list + thread/read(includeTurns) app-server
// surface this backend targets. Confirmed empirically against Codex CLI
// 0.145.0 (`codex app-server generate-json-schema` output plus a live
// initialize/thread-list handshake against a scratch CODEX_HOME — see
// docs/reference/codex-app-server.md). Versions below this are reported as
// unsupported rather than assumed compatible.
var MinSupportedVersion = Version{Major: 0, Minor: 145, Patch: 0}

// Version is a parsed Codex CLI version.
type Version struct {
	Major, Minor, Patch int
	Raw                 string
}

// Less reports whether v is strictly older than other.
func (v Version) Less(other Version) bool {
	if v.Major != other.Major {
		return v.Major < other.Major
	}
	if v.Minor != other.Minor {
		return v.Minor < other.Minor
	}
	return v.Patch < other.Patch
}

func (v Version) String() string {
	if v.Raw != "" {
		return v.Raw
	}
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// SupportsPagination reports whether v exposes the thread/list cursor-based
// pagination surface (cursor/nextCursor/backwardsCursor) DIR-029's backend
// already relies on. This is part of the same stable, empirically-confirmed
// thread/list surface as MinSupportedVersion — not a separately negotiated
// capability — so it is currently equivalent to
// !v.Less(MinSupportedVersion). It exists as its own named, documented,
// unit-testable capability query (DIR-032) so callers reason about "can I
// use cursor-based continuation" explicitly (see codex.Provider.ListSessionsPage)
// rather than re-deriving it from MinSupportedVersion inline every time.
func (v Version) SupportsPagination() bool {
	return !v.Less(MinSupportedVersion)
}

// ExperimentalTurnPaginationMinVersion is a placeholder floor for a
// not-yet-empirically-confirmed turn/item-level pagination method some
// future Codex app-server release may expose (distinct from the stable
// thread/list pagination SupportsPagination reports on). No known Codex CLI
// version has been confirmed to support turn/item-level pagination, so it
// is set intentionally out of reach of any real version string — this is
// the single place a future confirmed version/method pair would get wired
// in, rather than scattering an "always unsupported" assumption across
// callers (see docs/reference/codex-history-model.md's "Experimental
// pagination" section).
var ExperimentalTurnPaginationMinVersion = Version{Major: 9999, Minor: 0, Patch: 0}

// SupportsExperimentalTurnPagination reports whether v satisfies
// ExperimentalTurnPaginationMinVersion. It always returns false for every
// real Codex CLI version today — meta-cc never issues a turn/item
// pagination request, and callers must keep using thread/read(includeTurns)
// (the confirmed, stable, non-paginated full-turn fetch) regardless of this
// result, per the Contract's "experimental methods are capability-
// negotiated and optional" — a false result here must fail safely to
// existing behavior, never be treated as a hard error.
func (v Version) SupportsExperimentalTurnPagination() bool {
	return !v.Less(ExperimentalTurnPaginationMinVersion)
}

// ParseVersion extracts a dotted major.minor.patch version from CLI output
// such as "codex-cli 0.145.0". It takes the last whitespace-separated field
// and parses leading digits from each dot-separated component, so trailing
// non-numeric suffixes (pre-release tags, etc.) don't cause a hard failure.
func ParseVersion(raw string) (Version, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return Version{}, fmt.Errorf("appserver: empty version string")
	}
	fields := strings.Fields(trimmed)
	numeric := fields[len(fields)-1]
	parts := strings.SplitN(numeric, ".", 3)

	nums := [3]int{}
	for i := 0; i < 3; i++ {
		if i >= len(parts) {
			break
		}
		digits := parts[i]
		end := len(digits)
		for end > 0 && (digits[end-1] < '0' || digits[end-1] > '9') {
			end--
		}
		digits = digits[:end]
		if digits == "" {
			return Version{}, fmt.Errorf("appserver: cannot parse codex version from %q", raw)
		}
		n, err := strconv.Atoi(digits)
		if err != nil {
			return Version{}, fmt.Errorf("appserver: cannot parse codex version from %q: %w", raw, err)
		}
		nums[i] = n
	}
	return Version{Major: nums[0], Minor: nums[1], Patch: nums[2], Raw: trimmed}, nil
}

// DetectResult reports whether an installed codex binary was found and, if
// so, whether it satisfies MinSupportedVersion.
type DetectResult struct {
	Found     bool
	Version   Version
	Supported bool
	Err       error // set when Found but --version output could not be parsed, or the command could not run
}

// DetectCLIVersion runs `command --version` with a short bounded timeout.
// It never returns a Go error itself: absence, a timeout, or an unparseable
// version are all reported via DetectResult so callers (auto-mode fallback,
// the version-gated E2E test) can decide to skip/fall back instead of
// failing outright.
func DetectCLIVersion(ctx context.Context, command string) DetectResult {
	if command == "" {
		command = "codex"
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, command, "--version") //nolint:gosec // command is caller-controlled configuration, not untrusted input
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return DetectResult{Found: false, Err: err}
	}

	v, err := ParseVersion(out.String())
	if err != nil {
		return DetectResult{Found: true, Err: err}
	}
	return DetectResult{Found: true, Version: v, Supported: !v.Less(MinSupportedVersion)}
}
