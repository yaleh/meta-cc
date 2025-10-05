package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yale/meta-cc/internal/mcp"
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
	// Phase 12 Revision: Use consolidated tools with scope parameter
	tools := getConsolidatedToolsList()

	result := map[string]interface{}{
		"tools": tools,
	}
	sendResponse(req.ID, result)
}

// Phase 12 Note: Legacy _session tools have been consolidated into scope parameter
// Old approach: 22 tools (11 base + 11 _session variants)
// New approach: 14 tools with scope parameter (reduces API complexity by ~36%)

func handleToolsCall(req jsonRPCRequest) {
	var params struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}

	if err := json.Unmarshal(req.Params, &params); err != nil {
		sendError(req.ID, ErrInvalidParams, "Invalid params", err.Error())
		return
	}

	// Execute the appropriate meta-cc command based on tool name
	output, err := executeTool(params.Name, params.Arguments)
	if err != nil {
		code := categorizeError(err)
		errType := categorizeErrorType(err)
		sendErrorWithType(req.ID, code, errType, categorizeMessage(err), err.Error())
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

// Phase 13 Enhancement: Error handling with structured codes
const (
	ErrInvalidParams   = -32602 // Invalid method parameters
	ErrNotFound        = -32001 // Session/resource not found
	ErrNoResults       = -32002 // Query returned no results
	ErrExecutionFailed = -32603 // Tool execution failed
)

// categorizeError maps error types to JSON-RPC error codes
func categorizeError(err error) int {
	if err == nil {
		return 0
	}

	errMsg := err.Error()
	if strings.Contains(errMsg, "session location failed") || strings.Contains(errMsg, "not found") {
		return ErrNotFound
	}
	if strings.Contains(errMsg, "no results") || strings.Contains(errMsg, "empty") {
		return ErrNoResults
	}
	if strings.Contains(errMsg, "required") || strings.Contains(errMsg, "invalid") {
		return ErrInvalidParams
	}
	return ErrExecutionFailed
}

// categorizeMessage provides user-friendly error messages
func categorizeMessage(err error) string {
	if err == nil {
		return ""
	}

	errMsg := err.Error()
	switch {
	case strings.Contains(errMsg, "session location failed"):
		return "Session not found"
	case strings.Contains(errMsg, "no results"):
		return "Query returned no results"
	case strings.Contains(errMsg, "required"):
		return "Missing required parameter"
	case strings.Contains(errMsg, "invalid"):
		return "Invalid parameter value"
	default:
		return "Tool execution failed"
	}
}

// categorizeErrorType provides string error type codes for better error categorization
// This complements the numeric JSON-RPC error codes with semantic type identifiers
func categorizeErrorType(err error) string {
	if err == nil {
		return ""
	}

	errMsg := err.Error()
	switch {
	case strings.Contains(errMsg, "session location failed") || strings.Contains(errMsg, "not found"):
		return "SESSION_NOT_FOUND"
	case strings.Contains(errMsg, "no results") || strings.Contains(errMsg, "empty"):
		return "NO_RESULTS"
	case strings.Contains(errMsg, "required") || strings.Contains(errMsg, "invalid"):
		return "INVALID_PARAMS"
	default:
		return "EXECUTION_FAILED"
	}
}

// Phase 14.2: Use refactored command builder from internal/mcp
func executeTool(name string, args map[string]interface{}) (string, error) {
	// Phase 12 Revision: Map _session suffix to scope parameter
	toolName := name
	scope := "session" // Default to session for backward compatibility

	// Handle legacy _session suffix tools
	if strings.HasSuffix(name, "_session") {
		toolName = strings.TrimSuffix(name, "_session")
		scope = "session"
		// Inject scope into args if not present
		if _, hasScope := args["scope"]; !hasScope {
			args["scope"] = scope
		}
	} else if name != "get_session_stats" {
		// Non-_session tools default to project scope (Phase 12 design)
		// But we maintain session default for backward compatibility
		// Users can explicitly set scope=project in args
		if _, hasScope := args["scope"]; !hasScope {
			args["scope"] = scope
		}
	}

	// Import and use the refactored builder
	// Note: This requires importing internal/mcp, which creates a circular dependency
	// Solution: Move BuildToolCommand to a shared package or keep it here
	cmdArgs, err := buildToolCommandInternal(toolName, args)
	if err != nil {
		return "", err
	}

	// Execute meta-cc command internally
	return executeMetaCCCommand(cmdArgs)
}

// buildToolCommandInternal delegates to the refactored builder
func buildToolCommandInternal(name string, args map[string]interface{}) ([]string, error) {
	return mcp.BuildToolCommand(name, args)
}

// Phase 14: Add streaming support for large sessions
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

	// Channel to capture the output with streaming support
	outCh := make(chan string)
	errCh := make(chan error)

	go func() {
		var output strings.Builder
		buf := make([]byte, 8192) // 8KB buffer for streaming

		for {
			n, readErr := r.Read(buf)
			if n > 0 {
				output.Write(buf[:n])
			}
			if readErr == io.EOF {
				break
			}
			if readErr != nil {
				errCh <- readErr
				return
			}
		}
		outCh <- output.String()
	}()

	// Execute the command
	execErr := rootCmd.Execute()

	// Close writer and restore
	w.Close()
	os.Stdout = oldStdout
	os.Args = oldArgs

	// Get output (wait for streaming to complete)
	select {
	case output := <-outCh:
		if execErr != nil {
			return "", execErr
		}
		return output, nil
	case streamErr := <-errCh:
		return "", streamErr
	}
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

// sendErrorWithType sends an error response with both numeric code and string type
// Phase 14 enhancement: Adds semantic error type for better error categorization
func sendErrorWithType(id interface{}, code int, errType string, message string, data interface{}) {
	resp := jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: map[string]interface{}{
			"code":    code,
			"type":    errType, // Semantic error type (e.g., "SESSION_NOT_FOUND")
			"message": message,
			"data":    data,
		},
	}
	jsonData, _ := json.Marshal(resp)
	fmt.Println(string(jsonData))
}
