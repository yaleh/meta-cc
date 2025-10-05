package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

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
						"enum":    []string{"jsonl", "tsv"},
						"default": "jsonl",
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
						"enum":    []string{"jsonl", "tsv"},
						"default": "jsonl",
					},
				},
			},
		},
		// Phase 12 Stage 12.3: Session-level tools with _session suffix
		{
			"name":        "analyze_errors_session",
			"description": "Analyze error patterns in the current session only",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"output_format": map[string]interface{}{
						"type":    "string",
						"enum":    []string{"jsonl", "tsv"},
						"default": "jsonl",
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
						"enum":    []string{"jsonl", "tsv"},
						"default": "jsonl",
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
						"enum":    []string{"jsonl", "tsv"},
						"default": "jsonl",
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
						"enum":    []string{"jsonl", "tsv"},
						"default": "jsonl",
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
						"enum":    []string{"jsonl", "tsv"},
						"default": "jsonl",
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
						"enum":    []string{"jsonl", "tsv"},
						"default": "jsonl",
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
						"enum":    []string{"jsonl", "tsv"},
						"default": "jsonl",
					},
				},
				"required": []string{"file"},
			},
		},
		{
			"name":        "query_project_state",
			"description": "Query current project state from session (Stage 8.12)",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"output_format": map[string]interface{}{
						"type":    "string",
						"enum":    []string{"jsonl", "tsv"},
						"default": "jsonl",
					},
				},
			},
		},
		{
			"name":        "query_successful_prompts",
			"description": "Query successful prompt patterns (Stage 8.12)",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"min_quality_score": map[string]interface{}{
						"type":        "number",
						"default":     0.8,
						"description": "Minimum quality score (0.0-1.0, default 0.8)",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"default":     10,
						"description": "Maximum number of results (default 10)",
					},
					"output_format": map[string]interface{}{
						"type":    "string",
						"enum":    []string{"jsonl", "tsv"},
						"default": "jsonl",
					},
				},
			},
		},
		// Phase 10: Advanced Query Tools
		{
			"name":        "query_tools_advanced",
			"description": "Query tool calls with SQL-like filter expressions (Phase 10)",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"where": map[string]interface{}{
						"type":        "string",
						"description": "SQL-like filter expression (e.g., \"tool='Bash' AND status='error'\")",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"default":     20,
						"description": "Maximum number of results (default 20)",
					},
					"output_format": map[string]interface{}{
						"type":    "string",
						"enum":    []string{"jsonl", "tsv"},
						"default": "jsonl",
					},
				},
				"required": []string{"where"},
			},
		},
		{
			"name":        "aggregate_stats",
			"description": "Aggregate statistics grouped by field (Phase 10)",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"group_by": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"tool", "status", "uuid"},
						"default":     "tool",
						"description": "Field to group by (tool, status, or uuid)",
					},
					"metrics": map[string]interface{}{
						"type":        "string",
						"default":     "count,error_rate",
						"description": "Comma-separated metrics (count, error_rate)",
					},
					"where": map[string]interface{}{
						"type":        "string",
						"description": "Optional filter expression",
					},
					"output_format": map[string]interface{}{
						"type":    "string",
						"enum":    []string{"jsonl", "tsv"},
						"default": "jsonl",
					},
				},
			},
		},
		{
			"name":        "query_time_series",
			"description": "Analyze metrics over time (Phase 10)",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"metric": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"tool-calls", "error-rate"},
						"default":     "tool-calls",
						"description": "Metric to analyze (tool-calls or error-rate)",
					},
					"interval": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"hour", "day", "week"},
						"default":     "hour",
						"description": "Time interval for bucketing",
					},
					"where": map[string]interface{}{
						"type":        "string",
						"description": "Optional filter expression",
					},
					"output_format": map[string]interface{}{
						"type":    "string",
						"enum":    []string{"jsonl", "tsv"},
						"default": "jsonl",
					},
				},
			},
		},
		{
			"name":        "query_files",
			"description": "File-level operation statistics (Phase 10)",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"sort_by": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"total_ops", "edit_count", "read_count", "write_count", "error_count", "error_rate"},
						"default":     "total_ops",
						"description": "Sort field",
					},
					"top": map[string]interface{}{
						"type":        "integer",
						"default":     20,
						"description": "Limit results to top N files",
					},
					"where": map[string]interface{}{
						"type":        "string",
						"description": "Optional filter expression",
					},
					"output_format": map[string]interface{}{
						"type":    "string",
						"enum":    []string{"jsonl", "tsv"},
						"default": "jsonl",
					},
				},
			},
		},
		// Phase 12 Stage 12.3: Additional session-level tools with _session suffix
		{
			"name":        "query_tools_session",
			"description": "Query tool calls in the current session only. For project-level queries, use query_tools.",
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
						"enum":    []string{"jsonl", "tsv"},
						"default": "jsonl",
					},
				},
			},
		},
		{
			"name":        "query_user_messages_session",
			"description": "Search user messages in the current session only using regex pattern matching",
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
						"enum":    []string{"jsonl", "tsv"},
						"default": "jsonl",
					},
				},
				"required": []string{"pattern"},
			},
		},
		{
			"name":        "query_tool_sequences_session",
			"description": "Query repeated tool call sequences (workflow patterns) in the current session only",
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
						"enum":    []string{"jsonl", "tsv"},
						"default": "jsonl",
					},
				},
			},
		},
		{
			"name":        "query_file_access_session",
			"description": "Query file access history (read/edit/write operations) in the current session only",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"file": map[string]interface{}{
						"type":        "string",
						"description": "File path to query (required)",
					},
					"output_format": map[string]interface{}{
						"type":    "string",
						"enum":    []string{"jsonl", "tsv"},
						"default": "jsonl",
					},
				},
				"required": []string{"file"},
			},
		},
		{
			"name":        "query_successful_prompts_session",
			"description": "Query successful prompt patterns in the current session only",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"min_quality_score": map[string]interface{}{
						"type":        "number",
						"default":     0.8,
						"description": "Minimum quality score (0.0-1.0, default 0.8)",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"default":     10,
						"description": "Maximum number of results (default 10)",
					},
					"output_format": map[string]interface{}{
						"type":    "string",
						"enum":    []string{"jsonl", "tsv"},
						"default": "jsonl",
					},
				},
			},
		},
		{
			"name":        "query_context_session",
			"description": "Query context around specific errors in the current session only",
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
						"enum":    []string{"jsonl", "tsv"},
						"default": "jsonl",
					},
				},
				"required": []string{"error_signature"},
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

// Phase 14: Helper functions to reduce duplication
func getOutputFormat(args map[string]interface{}) string {
	if format, ok := args["output_format"].(string); ok {
		return format
	}
	return "jsonl" // Phase 13: default to jsonl
}

func addLimitFlag(cmdArgs []string, args map[string]interface{}, defaultLimit string) []string {
	if limit, ok := args["limit"].(float64); ok {
		return append(cmdArgs, "--limit", fmt.Sprintf("%.0f", limit))
	}
	return append(cmdArgs, "--limit", defaultLimit)
}

func addToolFilter(cmdArgs []string, args map[string]interface{}) []string {
	if tool, ok := args["tool"].(string); ok && tool != "" {
		return append(cmdArgs, "--tool", tool)
	}
	return cmdArgs
}

func addStatusFilter(cmdArgs []string, args map[string]interface{}) []string {
	if status, ok := args["status"].(string); ok && status != "" {
		return append(cmdArgs, "--status", status)
	}
	return cmdArgs
}

func buildToolCommand(name string, args map[string]interface{}) ([]string, error) {
	outputFormat := getOutputFormat(args)

	// Phase 12: Determine if this is a project-level or session-level tool
	isProjectLevel := !strings.HasSuffix(name, "_session") && name != "get_session_stats"

	var cmdArgs []string

	switch {
	// Backward compatibility: get_session_stats retains original behavior (session-only)
	case name == "get_session_stats":
		cmdArgs = []string{"parse", "stats"}

	case strings.HasPrefix(name, "analyze_errors"):
		cmdArgs = []string{"analyze", "errors"}
		if isProjectLevel {
			cmdArgs = append([]string{"--project", "."}, cmdArgs...)
		}

	case name == "extract_tools":
		cmdArgs = []string{"query", "tools"}
		if isProjectLevel {
			cmdArgs = append([]string{"--project", "."}, cmdArgs...)
		}
		cmdArgs = addLimitFlag(cmdArgs, args, "100") // Default 100 to prevent overflow

	case strings.HasPrefix(name, "query_tools"):
		cmdArgs = []string{"query", "tools"}
		if isProjectLevel {
			cmdArgs = append([]string{"--project", "."}, cmdArgs...)
		}
		cmdArgs = addToolFilter(cmdArgs, args)
		cmdArgs = addStatusFilter(cmdArgs, args)
		cmdArgs = addLimitFlag(cmdArgs, args, "20")

	case strings.HasPrefix(name, "query_user_messages"):
		pattern, ok := args["pattern"].(string)
		if !ok || pattern == "" {
			return nil, fmt.Errorf("pattern parameter is required")
		}
		cmdArgs = []string{"query", "user-messages", "--match", pattern}
		if isProjectLevel {
			cmdArgs = append([]string{"--project", "."}, cmdArgs...)
		}
		cmdArgs = addLimitFlag(cmdArgs, args, "10")

	case strings.HasPrefix(name, "query_context"):
		errorSignature, ok := args["error_signature"].(string)
		if !ok || errorSignature == "" {
			return nil, fmt.Errorf("error_signature parameter is required")
		}
		cmdArgs = []string{"query", "context", "--error-signature", errorSignature}
		if isProjectLevel {
			cmdArgs = append([]string{"--project", "."}, cmdArgs...)
		}
		if window, ok := args["window"].(float64); ok {
			cmdArgs = append(cmdArgs, "--window", fmt.Sprintf("%.0f", window))
		} else {
			cmdArgs = append(cmdArgs, "--window", "3")
		}

	case strings.HasPrefix(name, "query_tool_sequences"):
		cmdArgs = []string{"query", "tool-sequences"}
		if isProjectLevel {
			cmdArgs = append([]string{"--project", "."}, cmdArgs...)
		}
		if minOccurrences, ok := args["min_occurrences"].(float64); ok {
			cmdArgs = append(cmdArgs, "--min-occurrences", fmt.Sprintf("%.0f", minOccurrences))
		} else {
			cmdArgs = append(cmdArgs, "--min-occurrences", "3")
		}
		if pattern, ok := args["pattern"].(string); ok && pattern != "" {
			cmdArgs = append(cmdArgs, "--pattern", pattern)
		}

	case strings.HasPrefix(name, "query_file_access"):
		file, ok := args["file"].(string)
		if !ok || file == "" {
			return nil, fmt.Errorf("file parameter is required")
		}
		cmdArgs = []string{"query", "file-access", "--file", file}
		if isProjectLevel {
			cmdArgs = append([]string{"--project", "."}, cmdArgs...)
		}

	case name == "query_project_state":
		cmdArgs = []string{"query", "project-state"}
		if isProjectLevel {
			cmdArgs = append([]string{"--project", "."}, cmdArgs...)
		}

	case strings.HasPrefix(name, "query_successful_prompts"):
		cmdArgs = []string{"query", "successful-prompts"}
		if isProjectLevel {
			cmdArgs = append([]string{"--project", "."}, cmdArgs...)
		}
		if minQualityScore, ok := args["min_quality_score"].(float64); ok {
			cmdArgs = append(cmdArgs, "--min-quality-score", fmt.Sprintf("%.2f", minQualityScore))
		}
		cmdArgs = addLimitFlag(cmdArgs, args, "10")

	// Phase 10: Advanced Query Tools
	case name == "query_tools_advanced":
		where, ok := args["where"].(string)
		if !ok || where == "" {
			return nil, fmt.Errorf("where parameter is required")
		}
		cmdArgs = []string{"query", "tools", "--filter", where}
		if isProjectLevel {
			cmdArgs = append([]string{"--project", "."}, cmdArgs...)
		}
		cmdArgs = addLimitFlag(cmdArgs, args, "20")

	case name == "aggregate_stats":
		cmdArgs = []string{"stats", "aggregate"}
		if isProjectLevel {
			cmdArgs = append([]string{"--project", "."}, cmdArgs...)
		}
		if groupBy, ok := args["group_by"].(string); ok && groupBy != "" {
			cmdArgs = append(cmdArgs, "--group-by", groupBy)
		} else {
			cmdArgs = append(cmdArgs, "--group-by", "tool")
		}
		if metrics, ok := args["metrics"].(string); ok && metrics != "" {
			cmdArgs = append(cmdArgs, "--metrics", metrics)
		} else {
			cmdArgs = append(cmdArgs, "--metrics", "count,error_rate")
		}
		if where, ok := args["where"].(string); ok && where != "" {
			cmdArgs = append(cmdArgs, "--filter", where)
		}

	case name == "query_time_series":
		cmdArgs = []string{"stats", "time-series"}
		if isProjectLevel {
			cmdArgs = append([]string{"--project", "."}, cmdArgs...)
		}
		if metric, ok := args["metric"].(string); ok && metric != "" {
			cmdArgs = append(cmdArgs, "--metric", metric)
		} else {
			cmdArgs = append(cmdArgs, "--metric", "tool-calls")
		}
		if interval, ok := args["interval"].(string); ok && interval != "" {
			cmdArgs = append(cmdArgs, "--interval", interval)
		} else {
			cmdArgs = append(cmdArgs, "--interval", "hour")
		}
		if where, ok := args["where"].(string); ok && where != "" {
			cmdArgs = append(cmdArgs, "--filter", where)
		}

	case name == "query_files":
		cmdArgs = []string{"stats", "files"}
		if isProjectLevel {
			cmdArgs = append([]string{"--project", "."}, cmdArgs...)
		}
		if sortBy, ok := args["sort_by"].(string); ok && sortBy != "" {
			cmdArgs = append(cmdArgs, "--sort-by", sortBy)
		} else {
			cmdArgs = append(cmdArgs, "--sort-by", "total_ops")
		}
		if top, ok := args["top"].(float64); ok {
			cmdArgs = append(cmdArgs, "--top", fmt.Sprintf("%.0f", top))
		} else {
			cmdArgs = append(cmdArgs, "--top", "20")
		}
		if where, ok := args["where"].(string); ok && where != "" {
			cmdArgs = append(cmdArgs, "--filter", where)
		}

	default:
		return nil, fmt.Errorf("unknown tool: %s", name)
	}

	// Add output format at the end
	cmdArgs = append(cmdArgs, "--output", outputFormat)
	return cmdArgs, nil
}

func executeTool(name string, args map[string]interface{}) (string, error) {
	cmdArgs, err := buildToolCommand(name, args)
	if err != nil {
		return "", err
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
