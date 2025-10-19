package main

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
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
	// Create root span for the request
	ctx := context.Background()
	var span trace.Span
	if GetTracer() != nil {
		ctx, span = GetTracer().Start(ctx, "jsonrpc.request",
			trace.WithAttributes(
				attribute.String("rpc.method", req.Method),
				attribute.String("rpc.jsonrpc.version", req.JSONRPC),
			),
		)
		defer span.End()
	}

	// Track request queue depth (arrival)
	RecordRequestQueueInc()

	traceID := GetTraceID(ctx)
	spanID := GetSpanID(ctx)

	slog.Debug("handling JSON-RPC request",
		"method", req.Method,
		"id", req.ID,
		"trace_id", traceID,
		"span_id", spanID,
	)

	// Track concurrent requests (processing starts)
	RecordRequestQueueDec()
	RecordConcurrentRequestInc()
	defer RecordConcurrentRequestDec()

	switch req.Method {
	case "initialize":
		handleInitialize(ctx, req)
	case "tools/list":
		// Record tools/list request (no tool name)
		RecordRequest("list", "tools/list", "success")
		handleToolsList(ctx, req)
	case "tools/call":
		handleToolsCall(ctx, req)
	default:
		slog.Warn("unknown method requested",
			"method", req.Method,
			"id", req.ID,
			"trace_id", traceID,
			"span_id", spanID,
		)
		// Record unknown method as error
		RecordRequest("unknown", req.Method, "invalid")
		RecordError("server", "validation_error", "error")
		if span != nil {
			span.SetStatus(codes.Error, "Method not found")
			span.RecordError(nil)
		}
		writeError(req.ID, -32601, "Method not found")
	}
}

func handleInitialize(ctx context.Context, req JSONRPCRequest) {
	traceID := GetTraceID(ctx)
	slog.Info("initialize request",
		"trace_id", traceID,
	)

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

func handleToolsList(ctx context.Context, req JSONRPCRequest) {
	traceID := GetTraceID(ctx)
	slog.Debug("tools list request",
		"trace_id", traceID,
	)

	tools := getToolDefinitions()
	result := map[string]interface{}{
		"tools": tools,
	}
	writeResponse(req.ID, result)
}

func handleToolsCall(ctx context.Context, req JSONRPCRequest) {
	// Extract tool name and arguments
	params := req.Params
	toolName, ok := params["name"].(string)
	if !ok {
		traceID := GetTraceID(ctx)
		slog.Error("invalid params: missing tool name",
			"error_type", "validation_error",
			"request_id", req.ID,
			"trace_id", traceID,
		)
		// Record validation error
		RecordRequest("unknown", "tools/call", "invalid")
		RecordError("server", "validation_error", "error")
		writeError(req.ID, -32602, "Invalid params: missing tool name")
		return
	}

	arguments, ok := params["arguments"].(map[string]interface{})
	if !ok {
		arguments = make(map[string]interface{})
	}

	// Create tool execution span
	var span trace.Span
	if GetTracer() != nil {
		ctx, span = GetTracer().Start(ctx, "tool.execute",
			trace.WithAttributes(
				attribute.String("tool.name", toolName),
			),
		)
		defer span.End()
	}

	// Create request-scoped logger
	logger, requestID := NewRequestLogger(toolName)

	// Get scope for logging
	scope := "project" // default
	if s, ok := arguments["scope"].(string); ok {
		scope = s
	}

	traceID := GetTraceID(ctx)
	spanID := GetSpanID(ctx)

	start := time.Now()

	logger.Info("tool execution started",
		"scope", scope,
		"trace_id", traceID,
		"span_id", spanID,
	)

	// Execute tool
	output, err := executor.ExecuteTool(cfg, toolName, arguments)
	elapsed := time.Since(start)

	if err != nil {
		errorType := classifyError(err)
		logger.Error("tool execution failed",
			"error", err.Error(),
			"error_type", errorType,
			"duration_ms", elapsed.Milliseconds(),
			"trace_id", traceID,
			"span_id", spanID,
		)

		// Record span error
		if span != nil {
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)
			span.SetAttributes(
				attribute.String("error.type", errorType),
			)
		}

		// Record error metrics
		RecordRequest(toolName, "tools/call", "error")
		RecordError(toolName, errorType, GetErrorSeverity(errorType))
		RecordRequestDuration(toolName, "error", elapsed)

		// Record USE error metrics (resource errors, timeout errors)
		if resourceType := ClassifyResourceError(err); resourceType != "" {
			RecordResourceError(resourceType)
			logger.Debug("resource error detected",
				"resource_type", resourceType,
				"trace_id", traceID,
			)
		}
		if contextType := ClassifyTimeoutError(err); contextType != "" {
			RecordTimeoutError(contextType)
			logger.Debug("timeout error detected",
				"context_type", contextType,
				"trace_id", traceID,
			)
		}

		writeError(req.ID, -32603, err.Error())
		return
	}

	logger.Info("tool execution completed",
		"status", "success",
		"duration_ms", elapsed.Milliseconds(),
		"output_length", len(output),
		"trace_id", traceID,
		"span_id", spanID,
	)

	// Record span success
	if span != nil {
		span.SetStatus(codes.Ok, "success")
		span.SetAttributes(
			attribute.Int("output.length", len(output)),
			attribute.Int64("duration.ms", elapsed.Milliseconds()),
		)
	}

	// Record success metrics
	RecordRequest(toolName, "tools/call", "success")
	RecordRequestDuration(toolName, "success", elapsed)

	// Return result
	result := map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": output,
			},
		},
		"_meta": map[string]interface{}{
			"request_id":  requestID,
			"duration_ms": elapsed.Milliseconds(),
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
