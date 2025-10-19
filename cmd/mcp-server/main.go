package main

import (
	"bufio"
	"encoding/json"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yaleh/meta-cc/internal/config"
)

// Global configuration (loaded at startup)
var cfg *config.Config

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
		"server_name", "meta-cc-mcp",
		"version", "1.0.0",
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
	StartResourceMonitoring(10 * time.Second)

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

	scanner := bufio.NewScanner(os.Stdin)

	slog.Info("MCP server ready", "status", "listening")

	for scanner.Scan() {
		line := scanner.Text()

		// Parse JSON-RPC request
		var req JSONRPCRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			slog.Error("failed to parse JSON-RPC request",
				"error", err.Error(),
				"error_type", "parse_error",
				"input_length", len(line),
			)
			writeError(nil, -32700, "Parse error")
			continue
		}

		// Handle request
		handleRequest(req)
	}

	if err := scanner.Err(); err != nil {
		slog.Error("scanner error",
			"error", err.Error(),
			"error_type", "io_error",
		)
		writeError(nil, -32603, "Input error: "+err.Error())
	}
}
