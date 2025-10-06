package main

import (
	"encoding/json"
	"io"
	"os"
)

type JSONRPCRequest struct {
	JSONRPC string                 `json:"jsonrpc"`
	ID      interface{}            `json:"id"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params"`
}

type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      interface{}   `json:"id"`
	Result  interface{}   `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
}

type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

var executor *ToolExecutor
var outputWriter io.Writer = os.Stdout

func init() {
	executor = NewToolExecutor()
}

func handleRequest(req JSONRPCRequest) {
	switch req.Method {
	case "initialize":
		handleInitialize(req)
	case "tools/list":
		handleToolsList(req)
	case "tools/call":
		handleToolsCall(req)
	default:
		writeError(req.ID, -32601, "Method not found")
	}
}

func handleInitialize(req JSONRPCRequest) {
	result := map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]interface{}{
			"tools": map[string]bool{},
		},
		"serverInfo": map[string]string{
			"name":    "meta-cc-mcp",
			"version": "1.0.0",
		},
	}
	writeResponse(req.ID, result)
}

func handleToolsList(req JSONRPCRequest) {
	tools := getToolDefinitions()
	result := map[string]interface{}{
		"tools": tools,
	}
	writeResponse(req.ID, result)
}

func handleToolsCall(req JSONRPCRequest) {
	// Extract tool name and arguments
	params := req.Params
	toolName, ok := params["name"].(string)
	if !ok {
		writeError(req.ID, -32602, "Invalid params: missing tool name")
		return
	}

	arguments, ok := params["arguments"].(map[string]interface{})
	if !ok {
		arguments = make(map[string]interface{})
	}

	// Execute tool
	output, err := executor.ExecuteTool(toolName, arguments)
	if err != nil {
		writeError(req.ID, -32603, err.Error())
		return
	}

	// Return result
	result := map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": output,
			},
		},
	}
	writeResponse(req.ID, result)
}

func writeResponse(id interface{}, result interface{}) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	_ = json.NewEncoder(outputWriter).Encode(resp)
}

func writeError(id interface{}, code int, message string) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
		},
	}
	_ = json.NewEncoder(outputWriter).Encode(resp)
}
