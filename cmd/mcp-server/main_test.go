package main

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

// scannerDefaultMaxTokenSize is the implicit 64 KiB cap bufio.Scanner applied
// before serveRequests replaced it. A frame at or beyond it used to make
// Scan() fail, ending the read loop and killing the server.
const scannerDefaultMaxTokenSize = 64 * 1024

// captureStdout swaps the package's outputWriter for a buffer for the duration
// of fn (the idiom the other tests in this package use) and returns what was
// written to it.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	var buf strings.Builder
	origStdout := outputWriter
	outputWriter = &buf
	defer func() { outputWriter = origStdout }()
	fn()
	return buf.String()
}

// decodeResponses decodes the newline-delimited JSON-RPC responses a read loop
// wrote to stdout.
func decodeResponses(t *testing.T, raw string) []JSONRPCResponse {
	t.Helper()
	var out []JSONRPCResponse
	dec := json.NewDecoder(strings.NewReader(raw))
	for {
		var resp JSONRPCResponse
		err := dec.Decode(&resp)
		if errors.Is(err, io.EOF) {
			return out
		}
		if err != nil {
			t.Fatalf("failed to decode response %d: %v (raw output: %q)", len(out), err, raw)
		}
		out = append(out, resp)
	}
}

// initializeFrame builds a valid initialize request whose total encoded length
// is at least minBytes, padded through a params field the handler ignores.
func initializeFrame(t *testing.T, id int, minBytes int) string {
	t.Helper()
	pad := ""
	if minBytes > 0 {
		// Pad so the total (pad plus the frame's fixed JSON overhead) reaches
		// minBytes; the assertion below catches any shortfall.
		pad = strings.Repeat("a", minBytes)
	}
	frame := `{"jsonrpc":"2.0","id":` + strconv.Itoa(id) + `,"method":"initialize","params":{"pad":"` + pad + `"}}`
	if len(frame) < minBytes {
		t.Fatalf("test fixture bug: frame is %d bytes, wanted at least %d", len(frame), minBytes)
	}
	return frame
}

// TestServeRequests_OversizedFrameKeepsServing is the regression test for the
// defect: a single frame larger than bufio.Scanner's 64 KiB cap used to end the
// read loop (Scan() returned false, scanner.Err() reported bufio.ErrTooLong)
// and terminate the server. The server must instead answer the oversized frame
// and go on to answer the next one.
func TestServeRequests_OversizedFrameKeepsServing(t *testing.T) {
	oversized := initializeFrame(t, 1, scannerDefaultMaxTokenSize+1024)
	followUp := initializeFrame(t, 2, 0)
	input := oversized + "\n" + followUp + "\n"

	var serveErr error
	raw := captureStdout(t, func() {
		serveErr = serveRequests(strings.NewReader(input), maxRequestLineBytes)
	})

	// "Does not drop into the EOF path": the loop returned cleanly at EOF
	// rather than reporting a terminal read error, and it did not emit the
	// -32603 input-error response that path writes.
	if serveErr != nil {
		t.Fatalf("oversized frame terminated the read loop: serveRequests returned %v", serveErr)
	}

	responses := decodeResponses(t, raw)
	if len(responses) != 2 {
		t.Fatalf("expected 2 responses (oversized frame + follow-up), got %d: %q", len(responses), raw)
	}

	// The >64 KiB frame is handled like any other request, not rejected.
	if responses[0].Error != nil {
		t.Errorf("oversized frame was rejected: %+v", responses[0].Error)
	}
	if responses[0].Result == nil {
		t.Errorf("oversized frame produced no result: %q", raw)
	}

	// And the server is still serving afterwards.
	if responses[1].Error != nil {
		t.Errorf("request after the oversized frame failed: %+v", responses[1].Error)
	}
	if id, ok := responses[1].ID.(float64); !ok || int(id) != 2 {
		t.Errorf("expected the follow-up frame (id=2) to be answered, got id=%v (type %T)", responses[1].ID, responses[1].ID)
	}
}

// TestServeRequests_FrameOverBoundIsRejectedPerRequest covers the branch the
// real maxRequestLineBytes (64 MiB) is too large to reach in a unit test: a
// frame beyond the bound is rejected with a JSON-RPC error for that frame
// alone, and the next frame is still served.
func TestServeRequests_FrameOverBoundIsRejectedPerRequest(t *testing.T) {
	const bound = 1024
	overBound := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"pad":"` +
		strings.Repeat("a", bound*2) + `"}}`
	followUp := `{"jsonrpc":"2.0","id":2,"method":"initialize","params":{}}`
	input := overBound + "\n" + followUp + "\n"

	var serveErr error
	raw := captureStdout(t, func() {
		serveErr = serveRequests(strings.NewReader(input), bound)
	})

	if serveErr != nil {
		t.Fatalf("over-bound frame terminated the read loop: serveRequests returned %v", serveErr)
	}

	responses := decodeResponses(t, raw)
	if len(responses) != 2 {
		t.Fatalf("expected 2 responses (over-bound frame + follow-up), got %d: %q", len(responses), raw)
	}
	if responses[0].Error == nil {
		t.Errorf("expected the over-bound frame to be answered with an error, got result %v", responses[0].Result)
	} else if responses[0].Error.Code != -32600 {
		t.Errorf("expected error code -32600 for the over-bound frame, got %d (%s)",
			responses[0].Error.Code, responses[0].Error.Message)
	}
	if responses[1].Error != nil {
		t.Errorf("request after the over-bound frame failed: %+v", responses[1].Error)
	}
	if id, ok := responses[1].ID.(float64); !ok || int(id) != 2 {
		t.Errorf("expected the follow-up frame (id=2) to be answered, got id=%v (type %T)", responses[1].ID, responses[1].ID)
	}
}

// TestServeRequests_OverlongLineTailDoesNotSwallowNextFrame pins the invariant
// behind the decision not to drain an over-long line: whether the read stopped
// mid-line (as here, where the line is longer than the reader's internal buffer)
// or past the newline, the frame that follows must still be served. The tail of
// the over-long line is answered with errors of its own — noisy, but never at
// the cost of a legitimate request.
func TestServeRequests_OverlongLineTailDoesNotSwallowNextFrame(t *testing.T) {
	const bound = 1024
	// Longer than the read loop's internal bufio buffer (4096), so the cap is
	// crossed on a buffer-full read that stopped well before the newline.
	overlong := strings.Repeat("a", 8*1024)

	var serveErr error
	raw := captureStdout(t, func() {
		serveErr = serveRequests(
			strings.NewReader(overlong+"\n"+initializeFrame(t, 7, 0)+"\n"),
			bound,
		)
	})

	if serveErr != nil {
		t.Fatalf("serveRequests returned %v, want nil", serveErr)
	}

	responses := decodeResponses(t, raw)
	if len(responses) < 2 {
		t.Fatalf("expected the over-long line to be answered and the next frame served, got %d responses: %q", len(responses), raw)
	}

	last := responses[len(responses)-1]
	if last.Error != nil {
		t.Fatalf("the frame after the over-long line was rejected (%+v); its bytes were swallowed", last.Error)
	}
	if id, ok := last.ID.(float64); !ok || int(id) != 7 {
		t.Errorf("expected the final response to answer frame id=7, got id=%v (type %T)", last.ID, last.ID)
	}
	for i, resp := range responses[:len(responses)-1] {
		if resp.Error == nil {
			t.Errorf("response %d unexpectedly succeeded: %+v", i, resp.Result)
		}
	}
}

// TestServeRequests_NormalFrameHandlingUnchanged pins the behaviour ordinary
// frames had before the rewrite: a valid initialize frame is answered, and a
// frame that is not JSON is answered with a parse error and does not end the
// loop.
func TestServeRequests_NormalFrameHandlingUnchanged(t *testing.T) {
	input := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}` + "\n" +
		"not json at all\n" +
		`{"jsonrpc":"2.0","id":3,"method":"initialize","params":{}}` + "\n"

	var serveErr error
	raw := captureStdout(t, func() {
		serveErr = serveRequests(strings.NewReader(input), maxRequestLineBytes)
	})

	if serveErr != nil {
		t.Fatalf("serveRequests returned %v, want nil", serveErr)
	}

	responses := decodeResponses(t, raw)
	if len(responses) != 3 {
		t.Fatalf("expected 3 responses (initialize, parse error, initialize), got %d: %q", len(responses), raw)
	}

	if responses[0].Error != nil {
		t.Errorf("initialize frame failed: %+v", responses[0].Error)
	}
	result, ok := responses[0].Result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected an initialize result object, got %T (%v)", responses[0].Result, responses[0].Result)
	}
	if result["protocolVersion"] != "2024-11-05" {
		t.Errorf("expected protocolVersion 2024-11-05, got %v", result["protocolVersion"])
	}

	if responses[1].Error == nil || responses[1].Error.Code != -32700 {
		t.Errorf("expected a -32700 parse error for the non-JSON line, got %+v", responses[1])
	}

	if responses[2].Error != nil {
		t.Errorf("initialize frame after the parse error failed: %+v", responses[2].Error)
	}
	if id, ok := responses[2].ID.(float64); !ok || int(id) != 3 {
		t.Errorf("expected the third frame (id=3) to be answered, got id=%v (type %T)", responses[2].ID, responses[2].ID)
	}
}

// TestServeRequests_FinalLineWithoutNewlineIsServed pins the EOF path: a last
// frame that has no trailing newline (and no blank-line tail to trip over) is
// still handled.
func TestServeRequests_FinalLineWithoutNewlineIsServed(t *testing.T) {
	var serveErr error
	raw := captureStdout(t, func() {
		serveErr = serveRequests(strings.NewReader(`{"jsonrpc":"2.0","id":9,"method":"initialize","params":{}}`), maxRequestLineBytes)
	})

	if serveErr != nil {
		t.Fatalf("serveRequests returned %v, want nil", serveErr)
	}
	responses := decodeResponses(t, raw)
	if len(responses) != 1 {
		t.Fatalf("expected 1 response, got %d: %q", len(responses), raw)
	}
	if id, ok := responses[0].ID.(float64); !ok || int(id) != 9 {
		t.Errorf("expected id=9, got %v (type %T)", responses[0].ID, responses[0].ID)
	}
}
