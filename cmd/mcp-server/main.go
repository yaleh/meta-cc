package main

import (
	"bufio"
	"encoding/json"
	"os"
)

func main() {
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
