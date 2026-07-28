package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/yaleh/meta-cc/internal/mcp/metrics"
	"github.com/yaleh/meta-cc/internal/version"
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
	// Top-level request guard (DIR-060): any request-scoped panic — e.g. a
	// slice-bounds overflow from a hostile pagination parameter — must degrade
	// to a JSON-RPC error response instead of crashing the whole server
	// process (a denial of service triggerable by any client). Registered
	// first so it runs last (LIFO), after the span/metrics defers below have
	// unwound. handleRequest is invoked synchronously per request from main's
	// read loop, so recovering here keeps the loop serving subsequent requests.
	defer func() {
		if r := recover(); r != nil {
			writePanicError(req.ID, req.Method, r)
		}
	}()

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
	metrics.RecordRequestQueueInc()

	traceID := GetTraceID(ctx)
	spanID := GetSpanID(ctx)

	slog.Debug("handling JSON-RPC request",
		"method", req.Method,
		"id", req.ID,
		"trace_id", traceID,
		"span_id", spanID,
	)

	// Track concurrent requests (processing starts)
	metrics.RecordRequestQueueDec()
	metrics.RecordConcurrentRequestInc()
	defer metrics.RecordConcurrentRequestDec()

	switch req.Method {
	case "initialize":
		handleInitialize(ctx, req)
	case "tools/list":
		// Record tools/list request (no tool name)
		metrics.RecordRequest("list", "tools/list", "success")
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
		metrics.RecordRequest("unknown", req.Method, "invalid")
		metrics.RecordError("server", "validation_error", "error")
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
			"name":    version.ServerName,
			"version": version.Version,
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
		metrics.RecordRequest("unknown", "tools/call", "invalid")
		metrics.RecordError("server", "validation_error", "error")
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
		metrics.RecordRequest(toolName, "tools/call", "error")
		metrics.RecordError(toolName, errorType, metrics.GetErrorSeverity(errorType))
		metrics.RecordRequestDuration(toolName, "error", elapsed)

		// Record USE error metrics (resource errors, timeout errors)
		if resourceType := metrics.ClassifyResourceError(err); resourceType != "" {
			metrics.RecordResourceError(resourceType)
			logger.Debug("resource error detected",
				"resource_type", resourceType,
				"trace_id", traceID,
			)
		}
		if contextType := metrics.ClassifyTimeoutError(err); contextType != "" {
			metrics.RecordTimeoutError(contextType)
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
	metrics.RecordRequest(toolName, "tools/call", "success")
	metrics.RecordRequestDuration(toolName, "success", elapsed)

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

// writePanicError converts a recovered panic into a JSON-RPC internal-error
// response and records the failure, so a request-scoped panic degrades
// gracefully rather than killing the server (DIR-060).
func writePanicError(id interface{}, method string, recovered interface{}) {
	slog.Error("recovered from panic while handling request",
		"method", method,
		"id", id,
		"panic", fmt.Sprintf("%v", recovered),
		"error_type", "internal_panic",
	)
	metrics.RecordRequest(method, "request", "error")
	metrics.RecordError("server", "internal_panic", "error")
	writeError(id, -32603, fmt.Sprintf("internal error: recovered from panic: %v", recovered))
}
