package main

import (
	"log/slog"
	"os"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/v3/process"
)

var (
	// RED Metrics: Rate
	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mcp_server_requests_total",
			Help: "Total number of MCP requests received",
		},
		[]string{"tool_name", "method", "status"},
	)

	toolCallsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mcp_server_tool_calls_total",
			Help: "Total number of tool calls executed",
		},
		[]string{"tool_name", "scope", "status"},
	)

	// RED Metrics: Errors
	errorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mcp_server_errors_total",
			Help: "Total number of errors encountered",
		},
		[]string{"tool_name", "error_type", "severity"},
	)

	// RED Metrics: Duration
	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mcp_server_request_duration_seconds",
			Help:    "Request processing duration (end-to-end)",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0, 5.0, 10.0, 30.0},
		},
		[]string{"tool_name", "status"},
	)

	toolExecutionDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mcp_server_tool_execution_duration_seconds",
			Help:    "Tool execution duration (excludes network/parsing)",
			Buckets: []float64{0.001, 0.01, 0.05, 0.1, 0.5, 1.0, 5.0, 10.0, 30.0},
		},
		[]string{"tool_name", "scope"},
	)

	// USE Metrics: Utilization
	goroutinesActive = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "mcp_server_goroutines_active",
			Help: "Number of active goroutines",
		},
	)

	memoryUtilization = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mcp_server_memory_utilization_bytes",
			Help: "Memory utilization of MCP server process",
		},
		[]string{"type"},
	)

	cpuUtilization = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "mcp_server_cpu_utilization_percent",
			Help: "CPU utilization percentage",
		},
	)

	fileDescriptorsOpen = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "mcp_server_file_descriptors_open",
			Help: "Number of open file descriptors",
		},
	)

	// USE Metrics: Saturation
	requestQueueDepth = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "mcp_server_request_queue_depth",
			Help: "Number of requests waiting to be processed",
		},
	)

	concurrentRequests = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "mcp_server_concurrent_requests",
			Help: "Number of requests currently being processed",
		},
	)

	memoryPressureEvents = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "mcp_server_memory_pressure_events_total",
			Help: "Number of GC pressure events (GC rate > 10/sec)",
		},
	)

	// USE Metrics: Errors
	resourceErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mcp_server_resource_errors_total",
			Help: "Resource-related errors (OOM, FD exhaustion, etc.)",
		},
		[]string{"resource_type"},
	)

	timeoutErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mcp_server_timeout_errors_total",
			Help: "Context timeout errors",
		},
		[]string{"context_type"},
	)
)

// Atomic counters for saturation metrics (thread-safe)
var (
	requestQueueCounter   atomic.Int32
	concurrentReqsCounter atomic.Int32
	prevNumGC             uint32
	prevCPUTime           time.Duration
	prevCPUTimestamp      time.Time
)

func init() {
	// Register RED metrics
	prometheus.MustRegister(requestsTotal)
	prometheus.MustRegister(toolCallsTotal)
	prometheus.MustRegister(errorsTotal)
	prometheus.MustRegister(requestDuration)
	prometheus.MustRegister(toolExecutionDuration)

	// Register USE metrics
	prometheus.MustRegister(goroutinesActive)
	prometheus.MustRegister(memoryUtilization)
	prometheus.MustRegister(cpuUtilization)
	prometheus.MustRegister(fileDescriptorsOpen)
	prometheus.MustRegister(requestQueueDepth)
	prometheus.MustRegister(concurrentRequests)
	prometheus.MustRegister(memoryPressureEvents)
	prometheus.MustRegister(resourceErrors)
	prometheus.MustRegister(timeoutErrors)

	// Initialize CPU tracking
	prevCPUTimestamp = time.Now()

	slog.Debug("Prometheus metrics registered",
		"metrics_count", 15,
		"red_metrics", 5,
		"use_metrics", 10,
		"cardinality_estimate", 1043,
	)
}

// RecordRequest records request metrics (Rate)
func RecordRequest(toolName, method, status string) {
	requestsTotal.WithLabelValues(toolName, method, status).Inc()
}

// RecordToolCall records tool call metrics (Rate)
func RecordToolCall(toolName, scope, status string) {
	toolCallsTotal.WithLabelValues(toolName, scope, status).Inc()
}

// RecordError records error metrics (Errors)
func RecordError(toolName, errorType, severity string) {
	errorsTotal.WithLabelValues(toolName, errorType, severity).Inc()
}

// RecordRequestDuration records request duration (Duration)
func RecordRequestDuration(toolName, status string, duration time.Duration) {
	requestDuration.WithLabelValues(toolName, status).Observe(duration.Seconds())
}

// RecordToolExecutionDuration records tool execution duration (Duration)
func RecordToolExecutionDuration(toolName, scope string, duration time.Duration) {
	toolExecutionDuration.WithLabelValues(toolName, scope).Observe(duration.Seconds())
}

// UpdateResourceMetrics updates USE resource metrics (called periodically)
func UpdateResourceMetrics() {
	// Update goroutines
	goroutinesActive.Set(float64(runtime.NumGoroutine()))

	// Update memory metrics
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	memoryUtilization.WithLabelValues("heap").Set(float64(m.Alloc))
	memoryUtilization.WithLabelValues("stack").Set(float64(m.StackInuse))
	memoryUtilization.WithLabelValues("total").Set(float64(m.Sys))

	// Update CPU utilization
	cpuPercent := GetCPUUtilization()
	if cpuPercent >= 0 {
		cpuUtilization.Set(cpuPercent)
	}

	// Update file descriptors
	fdCount := GetFileDescriptorCount()
	if fdCount >= 0 {
		fileDescriptorsOpen.Set(float64(fdCount))
	}

	// Check for GC pressure (GC rate > 10/sec)
	gcDelta := m.NumGC - prevNumGC
	prevNumGC = m.NumGC
	gcRate := float64(gcDelta) / 10.0 // Assuming 10s interval
	if gcRate > 10.0 {
		memoryPressureEvents.Inc()
		slog.Warn("memory pressure detected",
			"gc_rate_per_sec", gcRate,
			"threshold", 10.0,
		)
	}

	// Update saturation gauges from atomic counters
	requestQueueDepth.Set(float64(requestQueueCounter.Load()))
	concurrentRequests.Set(float64(concurrentReqsCounter.Load()))
}

// GetCPUUtilization returns CPU utilization percentage (0-100%)
// Returns -1 on error
func GetCPUUtilization() float64 {
	pid := int32(os.Getpid())
	proc, err := process.NewProcess(pid)
	if err != nil {
		slog.Debug("failed to get process for CPU tracking",
			"error", err,
		)
		return -1
	}

	cpuPercent, err := proc.CPUPercent()
	if err != nil {
		slog.Debug("failed to get CPU percent",
			"error", err,
		)
		return -1
	}

	return cpuPercent
}

// GetFileDescriptorCount returns number of open file descriptors
// Returns -1 on error
func GetFileDescriptorCount() int {
	pid := int32(os.Getpid())
	proc, err := process.NewProcess(pid)
	if err != nil {
		slog.Debug("failed to get process for FD tracking",
			"error", err,
		)
		return -1
	}

	fds, err := proc.NumFDs()
	if err != nil {
		// NumFDs is not available on all platforms
		slog.Debug("file descriptor tracking not available",
			"error", err,
		)
		return -1
	}

	return int(fds)
}

// StartResourceMonitoring starts background goroutine to collect resource metrics
func StartResourceMonitoring(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			UpdateResourceMetrics()
		}
	}()

	slog.Info("resource monitoring started",
		"interval_seconds", interval.Seconds(),
	)
}

// GetErrorSeverity classifies error severity based on error type
func GetErrorSeverity(errorType string) string {
	switch errorType {
	case "parse_error", "validation_error":
		return "error"
	case "execution_error":
		return "error"
	case "io_error", "network_error":
		return "warning"
	default:
		return "error"
	}
}

// RecordRequestQueueInc increments request queue depth (called when request arrives)
func RecordRequestQueueInc() {
	requestQueueCounter.Add(1)
	requestQueueDepth.Set(float64(requestQueueCounter.Load()))
}

// RecordRequestQueueDec decrements request queue depth (called when processing starts)
func RecordRequestQueueDec() {
	requestQueueCounter.Add(-1)
	requestQueueDepth.Set(float64(requestQueueCounter.Load()))
}

// RecordConcurrentRequestInc increments concurrent request count (called when processing starts)
func RecordConcurrentRequestInc() {
	concurrentReqsCounter.Add(1)
	concurrentRequests.Set(float64(concurrentReqsCounter.Load()))
}

// RecordConcurrentRequestDec decrements concurrent request count (called when processing completes)
func RecordConcurrentRequestDec() {
	concurrentReqsCounter.Add(-1)
	concurrentRequests.Set(float64(concurrentReqsCounter.Load()))
}

// RecordResourceError records a resource-related error (USE Error metric)
func RecordResourceError(resourceType string) {
	resourceErrors.WithLabelValues(resourceType).Inc()
}

// RecordTimeoutError records a context timeout error (USE Error metric)
func RecordTimeoutError(contextType string) {
	timeoutErrors.WithLabelValues(contextType).Inc()
}

// ClassifyResourceError classifies error into resource type
// Returns resource type or empty string if not a resource error
func ClassifyResourceError(err error) string {
	if err == nil {
		return ""
	}

	errMsg := err.Error()
	// Check for common resource errors
	if errorContains(errMsg, "out of memory") || errorContains(errMsg, "cannot allocate") {
		return "memory"
	}
	if errorContains(errMsg, "too many open files") || errorContains(errMsg, "file descriptor") {
		return "file_descriptors"
	}
	if errorContains(errMsg, "goroutine") {
		return "goroutines"
	}
	return ""
}

// ClassifyTimeoutError classifies error into timeout context type
// Returns context type or empty string if not a timeout error
func ClassifyTimeoutError(err error) string {
	if err == nil {
		return ""
	}

	errMsg := err.Error()
	if errorContains(errMsg, "deadline exceeded") || errorContains(errMsg, "timeout") {
		// Try to infer context type
		if errorContains(errMsg, "tool") {
			return "tool_execution"
		}
		if errorContains(errMsg, "request") {
			return "request"
		}
		return "other"
	}
	return ""
}

// errorContains is a simple case-sensitive substring check
func errorContains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[0:len(substr)] == substr || errorContains(s[1:], substr)))
}
