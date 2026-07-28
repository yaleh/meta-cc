package codex

import (
	"context"
	"io"

	"github.com/yaleh/meta-cc/internal/locator"
	"github.com/yaleh/meta-cc/internal/provider/codex/appserver"
)

// ThreadSource is the app-server RPC surface (thread/list, thread/read)
// that Provider's app-server backend depends on — identical in shape to
// this package's own unexported threadSource interface. It is exported
// ONLY so other internal packages' tests (e.g.
// internal/mcp/executor's query_sessions_handler_test.go) can inject a
// fake app-server and exercise the REAL dispatch/ListSessionsPage/
// FetchSessionsBounded code path end to end — including DIR-039's
// partial-result-plus-warning behavior on a mid-pagination failure —
// without spawning a real `codex app-server` subprocess. Production code
// never names this type; it exists purely as a cross-package test seam.
type ThreadSource interface {
	ThreadList(ctx context.Context, params appserver.ThreadListParams) (appserver.ThreadListResult, error)
	ThreadRead(ctx context.Context, params appserver.ThreadReadParams) (appserver.ThreadReadResult, error)
}

// nopCloser is the trivial io.Closer NewProviderForAppServerTest hands back
// alongside src: there is no subprocess/connection to release.
type nopCloser struct{}

func (nopCloser) Close() error { return nil }

// NewProviderForAppServerTest builds a Provider in the given mode whose
// app-server backend is wired directly to src instead of spawning a
// subprocess. Exported test seam — not for production use. See ThreadSource.
func NewProviderForAppServerTest(loc *locator.CodexLocator, mode Mode, src ThreadSource) *Provider {
	backend := &appServerBackend{connect: func(context.Context) (threadSource, io.Closer, error) {
		return src, nopCloser{}, nil
	}}
	return newProvider(loc, mode, backend)
}
