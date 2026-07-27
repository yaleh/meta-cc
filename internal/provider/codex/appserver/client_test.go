package appserver

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"sync"
	"testing"
	"time"
)

// fakeServer is an in-process, in-memory stand-in for `codex app-server`
// that speaks the exact same newline-delimited JSON-RPC framing over a pair
// of io.Pipes. This is the "fake app-server test harness" the DIR-029 plan
// calls for: it gives full, unconditional coverage of the protocol/
// handshake/pagination/timeout/fallback logic without spawning any real
// process or depending on a specific installed Codex CLI version.
type fakeServer struct {
	w    io.Writer
	wMu  sync.Mutex
	stop chan struct{}
}

// newFakePair wires a Client to a fakeServer over two io.Pipes and starts
// the server's read loop. handler is invoked for every request the client
// sends and returns either a result (marshaled into the response) or an
// RPCError.
func newFakePair(t *testing.T, handler func(method string, params json.RawMessage) (interface{}, *RPCError)) (*Client, *fakeServer) {
	t.Helper()
	clientReadR, serverWriteW := io.Pipe()
	serverReadR, clientWriteW := io.Pipe()

	client := NewClient(clientWriteW, clientReadR)
	server := &fakeServer{w: serverWriteW, stop: make(chan struct{})}

	go server.serve(t, serverReadR, handler)

	t.Cleanup(func() {
		close(server.stop)
		_ = client.Close()
		_ = clientWriteW.Close()
		_ = serverWriteW.Close()
	})
	return client, server
}

func (s *fakeServer) serve(t *testing.T, r io.Reader, handler func(method string, params json.RawMessage) (interface{}, *RPCError)) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), maxLineBytes)
	for scanner.Scan() {
		var req struct {
			ID     json.RawMessage `json:"id"`
			Method string          `json:"method"`
			Params json.RawMessage `json:"params"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
			continue
		}
		result, rpcErr := handler(req.Method, req.Params)
		s.respond(t, req.ID, result, rpcErr)
	}
}

func (s *fakeServer) respond(t *testing.T, id json.RawMessage, result interface{}, rpcErr *RPCError) {
	t.Helper()
	env := struct {
		ID     json.RawMessage `json:"id"`
		Result interface{}     `json:"result,omitempty"`
		Error  *RPCError       `json:"error,omitempty"`
	}{ID: id, Result: result, Error: rpcErr}
	line, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("fakeServer: marshal response: %v", err)
	}
	s.wMu.Lock()
	defer s.wMu.Unlock()
	_, _ = s.w.Write(append(line, '\n'))
}

// notify sends an unsolicited server->client notification (no id), used to
// verify the client drains and ignores these without disrupting request
// correlation.
func (s *fakeServer) notify(t *testing.T, method string, params interface{}) {
	t.Helper()
	env := struct {
		Method string      `json:"method"`
		Params interface{} `json:"params"`
	}{Method: method, Params: params}
	line, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("fakeServer: marshal notification: %v", err)
	}
	s.wMu.Lock()
	defer s.wMu.Unlock()
	_, _ = s.w.Write(append(line, '\n'))
}

func TestClientInitializeHandshake(t *testing.T) {
	client, _ := newFakePair(t, func(method string, params json.RawMessage) (interface{}, *RPCError) {
		if method != "initialize" {
			return nil, &RPCError{Code: -1, Message: "unexpected method " + method}
		}
		return InitializeResult{
			CodexHome: "/tmp/codex-home", PlatformFamily: "unix", PlatformOS: "linux", UserAgent: "fake/1.0",
		}, nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	result, err := client.Initialize(ctx, ClientInfo{Name: "meta-cc-test", Version: "0.0.1"})
	if err != nil {
		t.Fatalf("Initialize: %v", err)
	}
	if result.CodexHome != "/tmp/codex-home" || result.PlatformOS != "linux" {
		t.Fatalf("unexpected InitializeResult: %#v", result)
	}
}

func TestClientThreadListAndThreadRead(t *testing.T) {
	client, _ := newFakePair(t, func(method string, params json.RawMessage) (interface{}, *RPCError) {
		switch method {
		case "initialize":
			return InitializeResult{}, nil
		case "thread/list":
			var p ThreadListParams
			_ = json.Unmarshal(params, &p)
			if len(p.SourceKinds) == 0 {
				return nil, &RPCError{Code: -2, Message: "expected explicit sourceKinds"}
			}
			return ThreadListResult{Data: []Thread{{ID: "t1", CWD: "/repo", CreatedAt: 100}}}, nil
		case "thread/read":
			var p ThreadReadParams
			_ = json.Unmarshal(params, &p)
			if p.ThreadID != "t1" || !p.IncludeTurns {
				return nil, &RPCError{Code: -3, Message: "unexpected thread/read params"}
			}
			return ThreadReadResult{Thread: Thread{ID: "t1", Turns: []Turn{{ID: "turn1", Status: "completed"}}}}, nil
		default:
			return nil, &RPCError{Code: -1, Message: "unexpected method " + method}
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	listResult, err := client.ThreadList(ctx, ThreadListParams{SourceKinds: []string{"cli", "exec"}, ModelProviders: []string{}})
	if err != nil {
		t.Fatalf("ThreadList: %v", err)
	}
	if len(listResult.Data) != 1 || listResult.Data[0].ID != "t1" {
		t.Fatalf("unexpected ThreadListResult: %#v", listResult)
	}

	readResult, err := client.ThreadRead(ctx, ThreadReadParams{ThreadID: "t1", IncludeTurns: true})
	if err != nil {
		t.Fatalf("ThreadRead: %v", err)
	}
	if len(readResult.Thread.Turns) != 1 || readResult.Thread.Turns[0].ID != "turn1" {
		t.Fatalf("unexpected ThreadReadResult: %#v", readResult)
	}
}

func TestClientPropagatesRPCError(t *testing.T) {
	client, _ := newFakePair(t, func(method string, params json.RawMessage) (interface{}, *RPCError) {
		return nil, &RPCError{Code: 42, Message: "boom"}
	})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := client.ThreadList(ctx, ThreadListParams{})
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestClientIgnoresNotificationsWhileWaitingForResponse(t *testing.T) {
	client, server := newFakePair(t, func(method string, params json.RawMessage) (interface{}, *RPCError) {
		return InitializeResult{UserAgent: "fake"}, nil
	})
	// Fire a notification before the client even issues a request, proving
	// the read loop drains it rather than blocking or corrupting
	// correlation for the real request that follows.
	server.notify(t, "configWarning", map[string]string{"summary": "test"})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	result, err := client.Initialize(ctx, ClientInfo{Name: "x", Version: "1"})
	if err != nil {
		t.Fatalf("Initialize after notification: %v", err)
	}
	if result.UserAgent != "fake" {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestClientCallRespectsContextTimeout(t *testing.T) {
	// Server never responds; the call must fail via ctx, not hang forever.
	client, _ := newFakePair(t, func(method string, params json.RawMessage) (interface{}, *RPCError) {
		<-make(chan struct{}) // block forever (test cleanup tears the pipe down)
		return nil, nil
	})
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	_, err := client.ThreadList(ctx, ThreadListParams{})
	if err == nil {
		t.Fatalf("expected timeout error")
	}
}

func TestClientCallFailsAfterConnectionCloses(t *testing.T) {
	clientReadR, serverWriteW := io.Pipe()
	serverReadR, clientWriteW := io.Pipe()
	client := NewClient(clientWriteW, clientReadR)
	_ = serverReadR

	// Close the server's write end (simulating the process exiting): the
	// client's read loop hits EOF and any outstanding/future call must fail
	// promptly with a clear error rather than hang.
	_ = serverWriteW.Close()

	// Give the read loop a moment to observe EOF.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		_, err := client.ThreadList(ctx, ThreadListParams{})
		cancel()
		if err != nil {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("expected Call to fail after connection close")
}

func TestParseVersionAndDetect(t *testing.T) {
	v, err := ParseVersion("codex-cli 0.145.0")
	if err != nil {
		t.Fatalf("ParseVersion: %v", err)
	}
	if v.Major != 0 || v.Minor != 145 || v.Patch != 0 {
		t.Fatalf("unexpected version: %#v", v)
	}
	if v.Less(MinSupportedVersion) {
		t.Fatalf("0.145.0 should satisfy MinSupportedVersion")
	}

	old := Version{Major: 0, Minor: 100, Patch: 0}
	if !old.Less(MinSupportedVersion) {
		t.Fatalf("0.100.0 should be older than MinSupportedVersion")
	}

	if _, err := ParseVersion(""); err == nil {
		t.Fatalf("expected error for empty version string")
	}
}

func TestDetectCLIVersionAbsentBinary(t *testing.T) {
	result := DetectCLIVersion(context.Background(), "meta-cc-definitely-not-a-real-binary")
	if result.Found {
		t.Fatalf("expected Found=false for a nonexistent binary")
	}
}
