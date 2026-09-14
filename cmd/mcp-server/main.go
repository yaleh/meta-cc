package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yaleh/meta-cc/internal/config"
	"github.com/yaleh/meta-cc/internal/mcp/metrics"
	"github.com/yaleh/meta-cc/internal/parser"
	"github.com/yaleh/meta-cc/internal/version"
)

// Global configuration (loaded at startup)
var cfg *config.Config

// maxRequestLineBytes bounds a single JSON-RPC request frame read from stdin.
//
// The raw bufio.Scanner this replaced capped a token at an implicit 64 KiB, and
// crossing that cap made Scan() return false — which ended the read loop and
// terminated the whole server, so one oversized frame (a large tools/call
// argument blob, or a long tool result echoed back in a request) was a shutdown
// trigger instead of a rejected request. The bound is explicit here: generous
// enough for real frames, still finite so a malformed or hostile peer can never
// make the reader grow without bound. Crossing it is handled per-request in
// serveRequests, never as a shutdown condition.
const maxRequestLineBytes = 64 * 1024 * 1024

func main() {
	// Load configuration with fail-fast validation
	var err error
	cfg, err = config.Load()
	if err != nil {
		slog.Error("configuration error", "error", err)
		os.Exit(1)
	}

	// Initialize structured logging with configuration
	InitLogger(cfg)

	slog.Info("MCP server starting",
		"server_name", version.ServerName,
		"version", version.Version,
		"commit", version.Commit,
		"build_time", version.BuildTime,
	)

	// Initialize distributed tracing
	tracingCleanup, err := InitTracing()
	if err != nil {
		slog.Error("failed to initialize tracing",
			"error", err.Error(),
			"error_type", classifyError(err),
		)
		// Continue without tracing (non-fatal)
	} else {
		defer tracingCleanup()
	}

	// Start resource monitoring (USE metrics)
	metrics.StartResourceMonitoring(10 * time.Second)

	// Setup cleanup on exit
	defer func() {
		slog.Info("MCP server shutting down")
		if err := CleanupSessionCache(); err != nil {
			slog.Error("failed to cleanup session cache",
				"error", err.Error(),
				"error_type", classifyError(err),
			)
		}
	}()

	// Handle interrupt signals gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		// Cleanup will be called by defer
		os.Exit(0)
	}()

	slog.Info("MCP server ready", "status", "listening")

	if err := serveRequests(os.Stdin, maxRequestLineBytes); err != nil {
		slog.Error("scanner error",
			"error", err.Error(),
			"error_type", "io_error",
		)
		writeError(nil, -32603, "Input error: "+err.Error())
	}
}

// serveRequests runs the JSON-RPC read loop over r, dispatching every complete
// frame to handleRequest until r reaches EOF. It returns a non-nil error only
// for a genuine terminal read failure on r: an oversized frame is answered with
// a JSON-RPC error and the loop keeps serving, because terminating the server
// on one bad request is precisely the defect this loop exists to avoid.
//
// maxLineBytes caps a single frame. It is a parameter rather than a direct read
// of maxRequestLineBytes so tests can exercise the oversized-frame path without
// materialising a maxRequestLineBytes-long line.
func serveRequests(r io.Reader, maxLineBytes int) error {
	reader := bufio.NewReader(r)
	for {
		line, err := parser.ReadLineBounded(reader, maxLineBytes)

		if errors.Is(err, parser.ErrLineTooLong) {
			// A per-request rejection, not a terminal condition — this is the
			// defect the raw bufio.Scanner carried, where crossing its implicit
			// 64 KiB cap ended the read loop and killed the server.
			//
			// The offending line is deliberately NOT drained: ReadLineBounded
			// checks its cap only after appending a whole ReadSlice chunk, and
			// ReadSlice returns through the '\n' when it finds one, so the over-
			// long line may already be fully consumed. Discarding "to the next
			// newline" would therefore swallow the next legitimate frame. When
			// the read did stop mid-line, the tail is simply read back as the
			// next line and rejected by dispatchRequest's parse-error path —
			// noisy, but it can never drop a real request.
			slog.Error("JSON-RPC request exceeds maximum line length",
				"error", err.Error(),
				"error_type", "request_too_long",
				"max_line_bytes", maxLineBytes,
			)
			writeError(nil, -32600, fmt.Sprintf("Request too large: exceeds %d bytes", maxLineBytes))
			continue
		}
		if err != nil && !errors.Is(err, io.EOF) {
			// A genuine read error on stdin is terminal, the same as it was
			// through bufio.Scanner.Err() before this loop was rewritten.
			return err
		}

		// Blank lines are skipped, including the empty tail an input ending in
		// '\n' produces: the previous scanner loop emitted no token for that
		// tail either, and a mid-stream blank line was only ever a spurious
		// -32700 parse error. io.EOF still carries the stream's final,
		// newline-less line (if any), so process it before stopping.
		if trimmed := bytes.TrimRight(line, "\r\n"); len(trimmed) > 0 {
			dispatchRequest(trimmed)
		}
		if errors.Is(err, io.EOF) {
			return nil
		}
	}
}

// dispatchRequest parses one JSON-RPC frame and handles it. A frame that is not
// valid JSON is answered with a parse error and dropped, exactly as the
// previous scanner-based loop did.
func dispatchRequest(line []byte) {
	var req JSONRPCRequest
	if err := json.Unmarshal(line, &req); err != nil {
		slog.Error("failed to parse JSON-RPC request",
			"error", err.Error(),
			"error_type", "parse_error",
			"input_length", len(line),
		)
		writeError(nil, -32700, "Parse error")
		return
	}

	handleRequest(req)
}
