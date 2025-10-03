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
			"description": "Extract tool usage data from the current session with pagination (Phase 8 enhanced)",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"limit": map[string]interface{}{
						"type":        "integer",
						"default":     100,
						"description": "Maximum number of tools to extract (default 100, prevents overflow)",
					},
					"output_format": map[string]interface{}{
						"type":    "string",
						"enum":    []string{"json", "md"},
						"default": "json",
					},
				},
			},
		},
		{
			"name":        "query_tools",
			"description": "Query tool calls with flexible filtering and pagination (Phase 8)",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"tool": map[string]interface{}{
						"type":        "string",
						"description": "Filter by tool name (e.g., 'Bash', 'Read', 'Edit')",
					},
					"status": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"error", "success"},
						"description": "Filter by execution status",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"default":     20,
						"description": "Maximum number of results (default 20)",
					},
					"output_format": map[string]interface{}{
						"type":    "string",
						"enum":    []string{"json", "md"},
						"default": "json",
					},
				},
			},
		},
		{
			"name":        "query_user_messages",
			"description": "Search user messages with regex pattern matching (Phase 8)",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"pattern": map[string]interface{}{
						"type":        "string",
						"description": "Regex pattern to match in message content (required)",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"default":     10,
						"description": "Maximum number of results (default 10)",
					},
					"output_format": map[string]interface{}{
						"type":    "string",
						"enum":    []string{"json", "md"},
						"default": "json",
					},
				},
				"required": []string{"pattern"},
			},
		},
		{
			"name":        "query_context",
			"description": "Query context around specific errors (Stage 8.10)",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"error_signature": map[string]interface{}{
						"type":        "string",
						"description": "Error pattern ID to query (required)",
					},
					"window": map[string]interface{}{
						"type":        "integer",
						"default":     3,
						"description": "Context window size in turns before/after (default 3)",
					},
					"output_format": map[string]interface{}{
						"type":    "string",
						"enum":    []string{"json", "md"},
						"default": "json",
					},
				},
				"required": []string{"error_signature"},
			},
		},
		{
			"name":        "query_tool_sequences",
			"description": "Query repeated tool call sequences (workflow patterns from Stage 8.11)",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"min_occurrences": map[string]interface{}{
						"type":        "integer",
						"default":     3,
						"description": "Minimum occurrences to report (default 3)",
					},
					"pattern": map[string]interface{}{
						"type":        "string",
						"description": "Specific sequence pattern to match (e.g., 'Read -> Edit -> Bash')",
					},
					"output_format": map[string]interface{}{
						"type":    "string",
						"enum":    []string{"json", "md"},
						"default": "json",
					},
				},
			},
		},
		{
			"name":        "query_file_access",
			"description": "Query file access history (read/edit/write operations)",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"file": map[string]interface{}{
						"type":        "string",
						"description": "File path to query (required)",
					},
					"output_format": map[string]interface{}{
						"type":    "string",
						"enum":    []string{"json", "md"},
						"default": "json",
					},
				},
				"required": []string{"file"},
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
		cmdArgs = []string{"query", "tools", "--output", outputFormat}

		// Add default limit to prevent overflow
		if limit, ok := args["limit"].(float64); ok {
			cmdArgs = append(cmdArgs, "--limit", fmt.Sprintf("%.0f", limit))
		} else {
			cmdArgs = append(cmdArgs, "--limit", "100") // Default 100
		}
	case "query_tools":
		cmdArgs = []string{"query", "tools", "--output", outputFormat}

		if tool, ok := args["tool"].(string); ok && tool != "" {
			cmdArgs = append(cmdArgs, "--tool", tool)
		}
		if status, ok := args["status"].(string); ok && status != "" {
			cmdArgs = append(cmdArgs, "--status", status)
		}
		if limit, ok := args["limit"].(float64); ok {
			cmdArgs = append(cmdArgs, "--limit", fmt.Sprintf("%.0f", limit))
		} else {
			cmdArgs = append(cmdArgs, "--limit", "20")
		}
	case "query_user_messages":
		pattern, ok := args["pattern"].(string)
		if !ok || pattern == "" {
			return "", fmt.Errorf("pattern parameter is required")
		}

		cmdArgs = []string{"query", "user-messages", "--match", pattern, "--output", outputFormat}

		if limit, ok := args["limit"].(float64); ok {
			cmdArgs = append(cmdArgs, "--limit", fmt.Sprintf("%.0f", limit))
		} else {
			cmdArgs = append(cmdArgs, "--limit", "10")
		}
	case "query_context":
		errorSignature, ok := args["error_signature"].(string)
		if !ok || errorSignature == "" {
			return "", fmt.Errorf("error_signature parameter is required")
		}

		cmdArgs = []string{"query", "context", "--error-signature", errorSignature, "--output", outputFormat}

		if window, ok := args["window"].(float64); ok {
			cmdArgs = append(cmdArgs, "--window", fmt.Sprintf("%.0f", window))
		} else {
			cmdArgs = append(cmdArgs, "--window", "3")
		}
	case "query_tool_sequences":
		cmdArgs = []string{"query", "tool-sequences", "--output", outputFormat}

		if minOccurrences, ok := args["min_occurrences"].(float64); ok {
			cmdArgs = append(cmdArgs, "--min-occurrences", fmt.Sprintf("%.0f", minOccurrences))
		} else {
			cmdArgs = append(cmdArgs, "--min-occurrences", "3")
		}
		if pattern, ok := args["pattern"].(string); ok && pattern != "" {
			cmdArgs = append(cmdArgs, "--pattern", pattern)
		}
	case "query_file_access":
		file, ok := args["file"].(string)
		if !ok || file == "" {
			return "", fmt.Errorf("file parameter is required")
		}

		cmdArgs = []string{"query", "file-access", "--file", file, "--output", outputFormat}
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
