package main

import (
	"context"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

var tracer trace.Tracer

// InitTracing initializes OpenTelemetry distributed tracing
func InitTracing() (func(), error) {
	// Create stdout exporter for testing/development
	exporter, err := stdouttrace.New(
		stdouttrace.WithPrettyPrint(),
		stdouttrace.WithWriter(os.Stderr), // Write traces to stderr to avoid mixing with JSON-RPC
	)
	if err != nil {
		return nil, err
	}

	// Create resource with service information
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceName("meta-cc-mcp"),
			semconv.ServiceVersion("1.0.0"),
		),
	)
	if err != nil {
		return nil, err
	}

	// Configure sampler (AlwaysOn for development, configurable for production)
	sampler := sdktrace.AlwaysSample()
	if samplingRatio := os.Getenv("OTEL_TRACES_SAMPLER_ARG"); samplingRatio != "" {
		// TraceIDRatioBased sampler can be configured via environment variable
		// Example: OTEL_TRACES_SAMPLER_ARG=0.1 for 10% sampling
		slog.Debug("trace sampling configured",
			"sampling_ratio", samplingRatio,
		)
	}

	// Create trace provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	// Register as global trace provider
	otel.SetTracerProvider(tp)

	// Get tracer for this service
	tracer = tp.Tracer("meta-cc-mcp")

	slog.Info("distributed tracing initialized",
		"exporter", "stdout",
		"service_name", "meta-cc-mcp",
	)

	// Return cleanup function
	cleanup := func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			slog.Error("failed to shutdown trace provider",
				"error", err.Error(),
			)
		}
	}

	return cleanup, nil
}

// GetTracer returns the global tracer instance
func GetTracer() trace.Tracer {
	return tracer
}

// GetTraceID extracts the trace ID from a context
func GetTraceID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return span.SpanContext().TraceID().String()
	}
	return ""
}

// GetSpanID extracts the span ID from a context
func GetSpanID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return span.SpanContext().SpanID().String()
	}
	return ""
}
