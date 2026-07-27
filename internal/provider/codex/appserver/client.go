package appserver

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
)

// maxLineBytes bounds a single JSON-RPC frame. `thread/read` responses can
// be large (a whole thread's turns/items), so this is generous, but still
// finite: a malformed or hostile peer can never make the scanner grow
// without bound.
const maxLineBytes = 64 * 1024 * 1024

// Client speaks the `codex app-server` newline-delimited JSON-RPC protocol
// over an arbitrary io.Writer (server stdin) / io.Reader (server stdout)
// pair. It is deliberately transport-agnostic: production code wires it to
// a real subprocess (see Process/StartProcess), while tests wire it
// directly to an in-process io.Pipe fake, giving full protocol coverage
// without spawning any process.
type Client struct {
	w io.Writer

	// writeCh serializes writes to w through a single dedicated goroutine
	// (writeLoop) rather than a mutex held across the (blocking) Write
	// call. This is what lets Call bound a stuck write by ctx/done: a
	// mutex would still make a second Call block inside Write itself
	// (uninterruptibly) if the transport's peer never reads; a channel
	// send can instead be raced against ctx.Done()/c.done in a select.
	writeCh chan writeJob

	nextID int64

	pendingMu sync.Mutex
	pending   map[string]chan Envelope

	done    chan struct{}
	readErr error

	closed atomic.Bool
}

// writeJob is one queued frame for writeLoop to write to the transport.
type writeJob struct {
	line []byte
}

// NewClient starts a background read loop over r and a background write
// loop over w, returning a Client ready to issue requests. The read loop
// terminates (and Client becomes unusable, with all in-flight and future
// Calls failing) when r returns EOF or an error; the write loop terminates
// when the client shuts down (via readLoop EOF/error, a failed write, or
// Close).
func NewClient(w io.Writer, r io.Reader) *Client {
	c := &Client{
		w:       w,
		writeCh: make(chan writeJob),
		pending: make(map[string]chan Envelope),
		done:    make(chan struct{}),
	}
	go c.readLoop(r)
	go c.writeLoop()
	return c
}

// writeLoop is the single goroutine that ever calls c.w.Write, so no mutex
// is needed to prevent interleaved frames. If a write fails (e.g. the
// peer's stdin pipe/process is gone), the whole connection is considered
// broken: shutdown unblocks every other in-flight and future Call via
// c.done, the same as a read-side EOF/error.
func (c *Client) writeLoop() {
	for {
		select {
		case job := <-c.writeCh:
			if _, err := c.w.Write(job.line); err != nil {
				c.shutdown(fmt.Errorf("write: %w", err))
				return
			}
		case <-c.done:
			return
		}
	}
}

func (c *Client) readLoop(r io.Reader) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), maxLineBytes)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var env Envelope
		if err := json.Unmarshal(line, &env); err != nil {
			// A malformed frame is a protocol violation from the peer, not
			// a reason to tear down calls already correctly correlated;
			// skip it and keep reading.
			continue
		}
		if env.IsNotification() {
			// Notifications (thread/started, configWarning, etc.) are out
			// of scope for the read-only surface this client implements;
			// drain and discard rather than blocking the read loop.
			continue
		}
		c.dispatch(env)
	}
	err := scanner.Err()
	if err == nil {
		err = io.EOF
	}
	c.shutdown(err)
}

func (c *Client) dispatch(env Envelope) {
	c.pendingMu.Lock()
	ch, ok := c.pending[string(env.ID)]
	if ok {
		delete(c.pending, string(env.ID))
	}
	c.pendingMu.Unlock()
	if ok {
		ch <- env
	}
}

func (c *Client) shutdown(err error) {
	if !c.closed.CompareAndSwap(false, true) {
		return
	}
	// readErr is set under pendingMu (not just guarded by the closed
	// atomic) because closed.CompareAndSwap and this plain write are two
	// separate memory operations: a reader that observes closed==true via
	// its own atomic Load is only guaranteed to see writes that happened
	// before the CAS in program order, not writes (like this one) issued
	// after it. Routing both the write here and every read (see connErr)
	// through the same mutex closes that window.
	c.pendingMu.Lock()
	c.readErr = err
	pending := c.pending
	c.pending = make(map[string]chan Envelope)
	c.pendingMu.Unlock()
	for _, ch := range pending {
		close(ch)
	}
	close(c.done)
}

// connErr returns the error that caused the connection to close (or will,
// once shutdown has run), synchronized against shutdown's write via
// pendingMu. Safe to call regardless of whether shutdown has run yet.
func (c *Client) connErr() error {
	c.pendingMu.Lock()
	defer c.pendingMu.Unlock()
	return c.readErr
}

// Call issues a JSON-RPC request and blocks until a matching response
// arrives, ctx is done, or the connection closes. result may be nil when
// the caller doesn't need the response payload.
func (c *Client) Call(ctx context.Context, method string, params, result interface{}) error {
	if c.closed.Load() {
		return fmt.Errorf("codex app-server: connection closed: %w", c.connErr())
	}

	id := atomic.AddInt64(&c.nextID, 1)
	idBytes, err := json.Marshal(id)
	if err != nil {
		return err
	}
	paramsBytes, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("codex app-server: encode params for %s: %w", method, err)
	}

	req := struct {
		ID     json.RawMessage `json:"id"`
		Method string          `json:"method"`
		Params json.RawMessage `json:"params"`
	}{idBytes, method, paramsBytes}
	line, err := json.Marshal(req)
	if err != nil {
		return err
	}

	ch := make(chan Envelope, 1)
	key := string(idBytes)
	c.pendingMu.Lock()
	c.pending[key] = ch
	c.pendingMu.Unlock()
	defer func() {
		c.pendingMu.Lock()
		delete(c.pending, key)
		c.pendingMu.Unlock()
	}()

	// Hand the frame to writeLoop rather than writing directly: this send
	// (and the whole Call) stays bounded by ctx/c.done even if the
	// transport's peer never reads (e.g. a closed/unread pipe, or a stuck
	// process), which a direct blocking Write cannot guarantee.
	select {
	case c.writeCh <- writeJob{line: append(line, '\n')}:
	case <-ctx.Done():
		return ctx.Err()
	case <-c.done:
		return fmt.Errorf("codex app-server: connection closed while sending %s request: %w", method, c.connErr())
	}

	select {
	case env, ok := <-ch:
		if !ok {
			return fmt.Errorf("codex app-server: connection closed while waiting for %s: %w", method, c.connErr())
		}
		if env.Error != nil {
			return fmt.Errorf("codex app-server: %s failed: %w", method, env.Error)
		}
		if result != nil && len(env.Result) > 0 {
			if err := json.Unmarshal(env.Result, result); err != nil {
				return fmt.Errorf("codex app-server: decode %s result: %w", method, err)
			}
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-c.done:
		return fmt.Errorf("codex app-server: connection closed while waiting for %s: %w", method, c.connErr())
	}
}

// Initialize performs the mandatory handshake, which must be the first call
// on a fresh connection.
func (c *Client) Initialize(ctx context.Context, info ClientInfo) (InitializeResult, error) {
	var result InitializeResult
	err := c.Call(ctx, "initialize", InitializeParams{ClientInfo: info}, &result)
	return result, err
}

// ThreadList calls the stable "thread/list" method.
func (c *Client) ThreadList(ctx context.Context, params ThreadListParams) (ThreadListResult, error) {
	var result ThreadListResult
	err := c.Call(ctx, "thread/list", params, &result)
	return result, err
}

// ThreadRead calls the stable "thread/read" method.
func (c *Client) ThreadRead(ctx context.Context, params ThreadReadParams) (ThreadReadResult, error) {
	var result ThreadReadResult
	err := c.Call(ctx, "thread/read", params, &result)
	return result, err
}

// Close unblocks any in-flight Calls with an error and marks the client
// unusable. It does not touch the underlying transport (io.Reader/Writer)
// lifecycle — callers that own a process (see Process) close stdio/kill the
// process separately.
func (c *Client) Close() error {
	c.shutdown(fmt.Errorf("client closed"))
	return nil
}
