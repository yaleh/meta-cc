// Package version is the single machine-readable source of truth for the
// current meta-cc release version.
//
// release.json is a small checked-in metadata file — the ONLY place the
// current release version is hand-edited. Every other shipped version
// surface (the Claude Code plugin manifest, the Claude Code marketplace
// manifest, the Codex plugin manifest, and the MCP server's
// initialize/serverInfo response) must agree with the value embedded here.
// Agreement is enforced offline by internal/release's consistency tests,
// which run as part of `go test ./...` (and therefore `make commit` /
// `make test` / CI) with no network access required.
//
// To bump the release version, use
// scripts/release/bump-plugin-version.sh --version X.Y.Z, which updates
// this file together with every other manifest. See
// docs/guides/release-process.md for the full procedure.
package version

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
)

//go:embed release.json
var releaseJSON []byte

// ServerName is the MCP server's advertised name in the initialize response.
const ServerName = "meta-cc-mcp"

type release struct {
	Version string `json:"version"`
}

// Version is the current meta-cc release version (e.g. "3.4.0"), parsed at
// package-init time from the embedded release.json.
var Version = mustLoadVersion()

// Commit and BuildTime are per-build fingerprints, distinct from the
// hand-bumped release Version above. Unlike Version (loaded from the
// go:embed'd release.json, which cannot be overridden by the linker), these
// are plain package-level vars so that `go build -ldflags "-X ...Commit=..."`
// can inject the actual git commit and build timestamp at link time. The
// Makefile's LDFLAGS (see Makefile's COMMIT/BUILD_TIME shell-derived
// variables) targets these exact symbols:
//
//	-X github.com/yaleh/meta-cc/internal/version.Commit=$(COMMIT)
//	-X github.com/yaleh/meta-cc/internal/version.BuildTime=$(BUILD_TIME)
//
// If a build does not go through the Makefile's LDFLAGS (e.g. `go build`
// invoked directly, or `go test`), these retain their zero-value defaults
// below.
var (
	Commit    = "unknown"
	BuildTime = "unknown"
)

func mustLoadVersion() string {
	var r release
	if err := json.Unmarshal(releaseJSON, &r); err != nil {
		panic(fmt.Sprintf("internal/version: failed to parse release.json: %v", err))
	}
	v := strings.TrimSpace(r.Version)
	if v == "" {
		panic("internal/version: release.json is missing a non-empty \"version\" field")
	}
	return v
}
