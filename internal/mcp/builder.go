package mcp

import (
	"fmt"
)

// Semantic default limits (Phase 9 + Phase 14 standardization)
const (
	DefaultLimitSmall  = 10  // user messages, prompts (high signal)
	DefaultLimitMedium = 20  // tool calls (moderate signal)
	DefaultLimitLarge  = 100 // extract operations (bulk data)
)

// CommandBuilder provides fluent interface for building meta-cc commands
// Reduces duplication and improves maintainability (Phase 14 refactoring)
type CommandBuilder struct {
	baseCmd      []string               // Base command parts (e.g., ["query", "tools"])
	scope        string                 // "project" or "session" (default: "session")
	filters      map[string]string      // Key-value filters (e.g., "tool" -> "Bash")
	params       map[string]string      // Required parameters (e.g., "match" -> "pattern")
	limit        int                    // Result limit (0 = no limit)
	outputFormat string                 // Output format (default: "jsonl")
	extraFlags   map[string]interface{} // Additional flags (window, min-occurrences, etc.)
}

// NewCommandBuilder creates a new command builder with base command parts
func NewCommandBuilder(parts ...string) *CommandBuilder {
	return &CommandBuilder{
		baseCmd:      parts,
		scope:        "session", // Default to session scope
		filters:      make(map[string]string),
		params:       make(map[string]string),
		outputFormat: "jsonl", // Phase 13: default to jsonl
		extraFlags:   make(map[string]interface{}),
	}
}

// WithScope sets the scope (project or session)
// Project scope adds "--project ." prefix to command
func (b *CommandBuilder) WithScope(scope string) *CommandBuilder {
	b.scope = scope
	return b
}

// WithFilter adds a filter flag (e.g., --tool, --status)
func (b *CommandBuilder) WithFilter(key, value string) *CommandBuilder {
	if value != "" {
		b.filters[key] = value
	}
	return b
}

// WithRequiredParam adds a required parameter (e.g., --match for user-messages)
func (b *CommandBuilder) WithRequiredParam(key, value string) *CommandBuilder {
	b.params[key] = value
	return b
}

// WithLimit sets the result limit
func (b *CommandBuilder) WithLimit(limit int) *CommandBuilder {
	b.limit = limit
	return b
}

// WithOutputFormat sets the output format
func (b *CommandBuilder) WithOutputFormat(format string) *CommandBuilder {
	if format != "" {
		b.outputFormat = format
	}
	return b
}

// WithExtraFlag adds additional flags (e.g., --window, --min-occurrences)
func (b *CommandBuilder) WithExtraFlag(key string, value interface{}) *CommandBuilder {
	b.extraFlags[key] = value
	return b
}

// Build constructs the final command array
func (b *CommandBuilder) Build() []string {
	cmd := []string{}

	// Add project scope prefix if needed
	if b.scope == "project" {
		cmd = append(cmd, "--project", ".")
	}

	// Add base command
	cmd = append(cmd, b.baseCmd...)

	// Add required parameters first
	for key, value := range b.params {
		cmd = append(cmd, "--"+key, value)
	}

	// Add filters
	for key, value := range b.filters {
		cmd = append(cmd, "--"+key, value)
	}

	// Add extra flags
	for key, value := range b.extraFlags {
		cmd = append(cmd, "--"+key, fmt.Sprintf("%v", value))
	}

	// Add limit if non-zero
	if b.limit > 0 {
		cmd = append(cmd, "--limit", fmt.Sprintf("%d", b.limit))
	}

	// Add output format at the end
	cmd = append(cmd, "--output", b.outputFormat)

	return cmd
}

// BuildToolCommand builds a command from MCP tool name and arguments
// This is the main entry point for MCP server integration
func BuildToolCommand(toolName string, args map[string]interface{}) ([]string, error) {
	// Extract common parameters
	outputFormat := getStringArg(args, "output_format", "jsonl")
	scope := getStringArg(args, "scope", "session") // Phase 12: default to session for backward compat

	// Route to appropriate builder based on tool name
	switch toolName {
	case "get_session_stats":
		// Backward compatibility: always session-only
		return NewCommandBuilder("parse", "stats").
			WithScope("session").
			WithOutputFormat(outputFormat).
			Build(), nil

	case "analyze_errors":
		return NewCommandBuilder("analyze", "errors").
			WithScope(scope).
			WithOutputFormat(outputFormat).
			Build(), nil

	case "extract_tools":
		limit := getIntArg(args, "limit", DefaultLimitLarge)
		return NewCommandBuilder("query", "tools").
			WithScope(scope).
			WithLimit(limit).
			WithOutputFormat(outputFormat).
			Build(), nil

	case "query_tools":
		tool := getStringArg(args, "tool", "")
		status := getStringArg(args, "status", "")
		limit := getIntArg(args, "limit", DefaultLimitMedium)

		return NewCommandBuilder("query", "tools").
			WithScope(scope).
			WithFilter("tool", tool).
			WithFilter("status", status).
			WithLimit(limit).
			WithOutputFormat(outputFormat).
			Build(), nil

	case "query_user_messages":
		pattern := getStringArg(args, "pattern", "")
		if pattern == "" {
			return nil, fmt.Errorf("pattern parameter is required")
		}
		limit := getIntArg(args, "limit", DefaultLimitSmall)

		return NewCommandBuilder("query", "user-messages").
			WithScope(scope).
			WithRequiredParam("match", pattern).
			WithLimit(limit).
			WithOutputFormat(outputFormat).
			Build(), nil

	case "query_context":
		errorSignature := getStringArg(args, "error_signature", "")
		if errorSignature == "" {
			return nil, fmt.Errorf("error_signature parameter is required")
		}
		window := getIntArg(args, "window", 3)

		return NewCommandBuilder("query", "context").
			WithScope(scope).
			WithRequiredParam("error-signature", errorSignature).
			WithExtraFlag("window", window).
			WithOutputFormat(outputFormat).
			Build(), nil

	case "query_tool_sequences":
		minOccurrences := getIntArg(args, "min_occurrences", 3)
		pattern := getStringArg(args, "pattern", "")

		builder := NewCommandBuilder("query", "tool-sequences").
			WithScope(scope).
			WithExtraFlag("min-occurrences", minOccurrences).
			WithOutputFormat(outputFormat)

		if pattern != "" {
			builder.WithExtraFlag("pattern", pattern)
		}

		return builder.Build(), nil

	case "query_file_access":
		file := getStringArg(args, "file", "")
		if file == "" {
			return nil, fmt.Errorf("file parameter is required")
		}

		return NewCommandBuilder("query", "file-access").
			WithScope(scope).
			WithRequiredParam("file", file).
			WithOutputFormat(outputFormat).
			Build(), nil

	case "query_project_state":
		return NewCommandBuilder("query", "project-state").
			WithScope(scope).
			WithOutputFormat(outputFormat).
			Build(), nil

	case "query_successful_prompts":
		minQualityScore := getFloatArg(args, "min_quality_score", 0.8)
		limit := getIntArg(args, "limit", DefaultLimitSmall)

		return NewCommandBuilder("query", "successful-prompts").
			WithScope(scope).
			WithExtraFlag("min-quality-score", fmt.Sprintf("%.2f", minQualityScore)).
			WithLimit(limit).
			WithOutputFormat(outputFormat).
			Build(), nil

	// Phase 10: Advanced query tools
	case "query_tools_advanced":
		where := getStringArg(args, "where", "")
		if where == "" {
			return nil, fmt.Errorf("where parameter is required")
		}
		limit := getIntArg(args, "limit", DefaultLimitMedium)

		return NewCommandBuilder("query", "tools").
			WithScope(scope).
			WithFilter("filter", where).
			WithLimit(limit).
			WithOutputFormat(outputFormat).
			Build(), nil

	case "aggregate_stats":
		groupBy := getStringArg(args, "group_by", "tool")
		metrics := getStringArg(args, "metrics", "count,error_rate")
		where := getStringArg(args, "where", "")

		builder := NewCommandBuilder("stats", "aggregate").
			WithScope(scope).
			WithExtraFlag("group-by", groupBy).
			WithExtraFlag("metrics", metrics).
			WithOutputFormat(outputFormat)

		if where != "" {
			builder.WithFilter("filter", where)
		}

		return builder.Build(), nil

	case "query_time_series":
		metric := getStringArg(args, "metric", "tool-calls")
		interval := getStringArg(args, "interval", "hour")
		where := getStringArg(args, "where", "")

		builder := NewCommandBuilder("stats", "time-series").
			WithScope(scope).
			WithExtraFlag("metric", metric).
			WithExtraFlag("interval", interval).
			WithOutputFormat(outputFormat)

		if where != "" {
			builder.WithFilter("filter", where)
		}

		return builder.Build(), nil

	case "query_files":
		sortBy := getStringArg(args, "sort_by", "total_ops")
		top := getIntArg(args, "top", DefaultLimitMedium)
		where := getStringArg(args, "where", "")

		builder := NewCommandBuilder("stats", "files").
			WithScope(scope).
			WithExtraFlag("sort-by", sortBy).
			WithExtraFlag("top", top).
			WithOutputFormat(outputFormat)

		if where != "" {
			builder.WithFilter("filter", where)
		}

		return builder.Build(), nil

	default:
		return nil, fmt.Errorf("unknown tool: %s", toolName)
	}
}

// Helper functions to extract typed arguments from map[string]interface{}

func getStringArg(args map[string]interface{}, key, defaultValue string) string {
	if val, ok := args[key].(string); ok {
		return val
	}
	return defaultValue
}

func getIntArg(args map[string]interface{}, key string, defaultValue int) int {
	if val, ok := args[key].(float64); ok {
		return int(val)
	}
	return defaultValue
}

func getFloatArg(args map[string]interface{}, key string, defaultValue float64) float64 {
	if val, ok := args[key].(float64); ok {
		return val
	}
	return defaultValue
}
