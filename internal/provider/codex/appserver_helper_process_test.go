package codex

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

// helperProcessEnvVar, when set to "1" in this test binary's own process
// environment, makes TestMain skip the normal test suite and instead run
// this same compiled binary as a minimal fake `codex app-server`: a
// newline-delimited JSON-RPC stdio server that answers "initialize" and
// "thread/list"/"thread/read" and then keeps reading — never exiting on its
// own — until its stdin is closed. That's exactly the shape of a real
// long-lived app-server child.
//
// The regression test in appserver_provider_regression_test.go points
// META_CC_CODEX_APP_SERVER_BIN at this same test binary (os.Executable())
// and sets this env var, so `codex app-server` inside connectProcess's
// exec.CommandContext call spawns a real, separate OS process — not an
// in-process fake — reproducing the actual subprocess-lifecycle bug
// (premature context cancellation killing the child right after connect()
// returns) rather than merely exercising a hand-rolled threadSource.
const helperProcessEnvVar = "META_CC_TEST_APPSERVER_HELPER"

func TestMain(m *testing.M) {
	if os.Getenv(helperProcessEnvVar) == "1" {
		runFakeAppServerHelperProcess()
		os.Exit(0)
	}
	os.Exit(m.Run())
}

// runFakeAppServerHelperProcess implements just enough of the app-server
// wire protocol (see appserver.Envelope / Client) to answer initialize,
// thread/list, and thread/read requests.
func runFakeAppServerHelperProcess() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 0, 64*1024), 64*1024*1024)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var req struct {
			ID     json.RawMessage `json:"id"`
			Method string          `json:"method"`
		}
		if err := json.Unmarshal(line, &req); err != nil {
			continue
		}

		var resp map[string]interface{}
		switch req.Method {
		case "initialize":
			resp = map[string]interface{}{
				"id": req.ID,
				"result": map[string]string{
					"codexHome": "/fake", "platformFamily": "unix",
					"platformOs": "linux", "userAgent": "fake-helper",
				},
			}
		case "thread/list":
			resp = map[string]interface{}{
				"id":     req.ID,
				"result": map[string]interface{}{"data": []interface{}{}, "nextCursor": nil},
			}
		case "thread/read":
			resp = map[string]interface{}{
				"id":     req.ID,
				"result": map[string]interface{}{"thread": map[string]interface{}{"id": "fake"}},
			}
		default:
			resp = map[string]interface{}{
				"id":    req.ID,
				"error": map[string]interface{}{"code": -32601, "message": "unknown method " + req.Method},
			}
		}

		out, err := json.Marshal(resp)
		if err != nil {
			continue
		}
		fmt.Fprintln(os.Stdout, string(out))
	}
}
