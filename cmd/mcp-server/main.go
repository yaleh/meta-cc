package main

import (
	"bufio"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Setup cleanup on exit
	defer func() {
		if err := CleanupSessionCache(); err != nil {
			// Log error to stderr but don't fail
			os.Stderr.WriteString("Warning: Failed to cleanup session cache: " + err.Error() + "\n")
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

	for scanner.Scan() {
		line := scanner.Text()

		// Parse JSON-RPC request
		var req JSONRPCRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			writeError(nil, -32700, "Parse error")
			continue
		}

		// Handle request
		handleRequest(req)
	}

	if err := scanner.Err(); err != nil {
		writeError(nil, -32603, "Input error: "+err.Error())
	}
}
