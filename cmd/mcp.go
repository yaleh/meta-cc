package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start MCP (Model Context Protocol) server",
	Long: `Start an MCP server that exposes meta-cc functionality via the Model Context Protocol.
This allows Claude Code and other MCP clients to query session statistics, analyze errors,
and extract tool usage data.`,
	RunE: runMCPServer,
}

func init() {
	rootCmd.AddCommand(mcpCmd)
}

type jsonRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type jsonRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

func runMCPServer(cmd *cobra.Command, args []string) error {
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		line := scanner.Bytes()

		var req jsonRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			sendError(req.ID, -32700, "Parse error", err.Error())
			continue
		}

		handleRequest(req)
	}

	return scanner.Err()
}

func handleRequest(req jsonRPCRequest) {
	switch req.Method {
	case "initialize":
		handleInitialize(req)
	case "tools/list":
		handleToolsList(req)
	case "tools/call":
		handleToolsCall(req)
	default:
		sendError(req.ID, -32601, "Method not found", fmt.Sprintf("Unknown method: %s", req.Method))
	}
}

func handleInitialize(req jsonRPCRequest) {
	result := map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{},
		},
		"serverInfo": map[string]interface{}{
			"name":    "meta-cc",
			"version": Version,
		},
	}
	sendResponse(req.ID, result)
}

func handleToolsList(req jsonRPCRequest) {
	tools := []map[string]interface{}{
		{
			"name":        "get_session_stats",
			"description": "Get statistics for the current Claude Code session",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"output_format": map[string]interface{}{
						"type":    "string",
						"enum":    []string{"json", "md"},
						"default": "json",
					},
				},
			},
		},
		{
			"name":        "analyze_errors",
			"description": "Analyze error patterns in the current session",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"output_format": map[string]interface{}{
						"type":    "string",
						"enum":    []string{"json", "md"},
						"default": "json",
					},
				},
			},
		},
		{
			"name":        "extract_tools",
			"description": "Extract tool usage data from the current session",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"output_format": map[string]interface{}{
						"type":    "string",
						"enum":    []string{"json", "md"},
						"default": "json",
					},
				},
			},
		},
	}

	result := map[string]interface{}{
		"tools": tools,
	}
	sendResponse(req.ID, result)
}

func handleToolsCall(req jsonRPCRequest) {
	var params struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}

	if err := json.Unmarshal(req.Params, &params); err != nil {
		sendError(req.ID, -32602, "Invalid params", err.Error())
		return
	}

	// Execute the appropriate meta-cc command based on tool name
	output, err := executeTool(params.Name, params.Arguments)
	if err != nil {
		sendError(req.ID, -32603, "Tool execution failed", err.Error())
		return
	}

	result := map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": output,
			},
		},
	}
	sendResponse(req.ID, result)
}

func executeTool(name string, args map[string]interface{}) (string, error) {
	outputFormat := "json"
	if format, ok := args["output_format"].(string); ok {
		outputFormat = format
	}

	var cmdArgs []string

	switch name {
	case "get_session_stats":
		cmdArgs = []string{"parse", "stats", "--output", outputFormat}
	case "analyze_errors":
		cmdArgs = []string{"analyze", "errors", "--output", outputFormat}
	case "extract_tools":
		cmdArgs = []string{"parse", "extract", "--type", "tools", "--output", outputFormat}
	default:
		return "", fmt.Errorf("unknown tool: %s", name)
	}

	// Execute meta-cc command internally
	return executeMetaCCCommand(cmdArgs)
}

func executeMetaCCCommand(args []string) (string, error) {
	// Save original stdout
	oldStdout := os.Stdout
	oldArgs := os.Args

	// Create a pipe to capture output
	r, w, err := os.Pipe()
	if err != nil {
		return "", err
	}

	// Replace stdout
	os.Stdout = w
	os.Args = append([]string{"meta-cc"}, args...)

	// Channel to capture the output
	outCh := make(chan string)
	go func() {
		var buf []byte
		buf = make([]byte, 1024*1024) // 1MB buffer
		n, _ := r.Read(buf)
		outCh <- string(buf[:n])
	}()

	// Execute the command
	err = rootCmd.Execute()

	// Close writer and restore
	w.Close()
	os.Stdout = oldStdout
	os.Args = oldArgs

	// Get output
	output := <-outCh

	if err != nil {
		return "", err
	}

	return output, nil
}

func sendResponse(id interface{}, result interface{}) {
	resp := jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	data, _ := json.Marshal(resp)
	fmt.Println(string(data))
}

func sendError(id interface{}, code int, message string, data interface{}) {
	resp := jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: map[string]interface{}{
			"code":    code,
			"message": message,
			"data":    data,
		},
	}
	jsonData, _ := json.Marshal(resp)
	fmt.Println(string(jsonData))
}
