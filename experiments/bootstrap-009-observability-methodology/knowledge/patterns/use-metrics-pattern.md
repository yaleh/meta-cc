# USE Metrics Pattern (Universal)

**Pattern Type**: Observability / Metrics
**Domain**: Resource monitoring (CPU, memory, I/O, etc.)
**Transferability**: Very high (applicable to any system)
**Source**: Brendan Gregg, "Systems Performance" (2013)
**Created**: 2025-10-17 (Bootstrap-009 Iteration 3)

---

## Pattern Overview

**USE** (Utilization, Saturation, Errors) is a metrics methodology for monitoring resource health and capacity planning. It focuses on the system's perspective rather than the user's perspective (like RED).

### The Three Metrics (Per Resource)

1. **Utilization**: How busy is the resource? (0-100%)
2. **Saturation**: Is the resource queuing work? (queue depth or saturation events)
3. **Errors**: Is the resource failing? (error count)

### When to Use USE

- **Resource monitoring**: CPU, memory, disk, network, file descriptors, threads/goroutines
- **Capacity planning**: When you need to predict resource exhaustion
- **System performance**: When diagnosing performance bottlenecks
- **Reliability**: When tracking resource failures

### When NOT to Use USE

- **Request flow monitoring**: Use RED instead for request-driven services
- **Business metrics**: USE is for system resources, not business KPIs

---

## Universal Implementation Pattern

### Resource Types

For each resource in your system, apply USE metrics:

| Resource | Utilization | Saturation | Errors |
|----------|-------------|------------|--------|
| CPU | CPU % busy | Load average, run queue depth | CPU errors (rare) |
| Memory | Memory bytes used / total | Page faults, swap usage | OOM events |
| Disk I/O | Disk % busy | Disk queue depth | I/O errors |
| Network I/O | Network % busy | Network queue depth | Packet loss, timeouts |
| File Descriptors | FDs open / ulimit | FD exhaustion events | "Too many open files" |
| Goroutines/Threads | Active count | Thread pool queue | Thread creation failures |

---

## 1. Utilization Metrics

**Purpose**: How busy is the resource?

**Metric Type**: Gauge (percentage 0-100% or absolute value)

### CPU Utilization

**Metric Name**: `{service}_cpu_utilization_percent`
**Type**: Gauge
**Unit**: Percent (0-100%)

```prometheus
{service}_cpu_utilization_percent
```

**Collection**:
- **Linux**: Read `/proc/{pid}/stat` (utime + stime)
- **Go**: `runtime.ReadMemStats()` (indirect, via CPU profiling)
- **Interval**: 10 seconds (background goroutine)

**Calculation**:
```
cpu_percent = (cpu_time_delta / wall_time_delta) * 100
```

**Example (Go)**:
```go
var cpuUtilization = prometheus.NewGauge(
    prometheus.GaugeOpts{
        Name: "myservice_cpu_utilization_percent",
        Help: "CPU utilization percentage",
    },
)

go func() {
    ticker := time.NewTicker(10 * time.Second)
    for range ticker.C {
        cpuPercent := getCPUUtilization()  // Read /proc/{pid}/stat
        cpuUtilization.Set(cpuPercent)
    }
}()
```

**Prometheus Query**:
```promql
# Current CPU utilization
{service}_cpu_utilization_percent

# Average CPU over 1 hour
avg_over_time({service}_cpu_utilization_percent[1h])
```

**Alerting**:
```yaml
- alert: HighCPUUtilization
  expr: {service}_cpu_utilization_percent > 80
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "CPU utilization > 80% (approaching capacity)"
```

### Memory Utilization

**Metric Name**: `{service}_memory_utilization_bytes`
**Type**: Gauge
**Labels**: type (heap, stack, total)
**Unit**: Bytes

```prometheus
{service}_memory_utilization_bytes{type="heap"}
{service}_memory_utilization_bytes{type="stack"}
{service}_memory_utilization_bytes{type="total"}
```

**Collection**:
- **Go**: `runtime.ReadMemStats()` (HeapAlloc, StackInuse, Sys)
- **Linux**: `/proc/{pid}/status` (VmRSS, VmSize)
- **Interval**: 10 seconds

**Example (Go)**:
```go
var memoryUtilization = prometheus.NewGaugeVec(
    prometheus.GaugeOpts{
        Name: "myservice_memory_utilization_bytes",
        Help: "Memory utilization in bytes",
    },
    []string{"type"},
)

go func() {
    ticker := time.NewTicker(10 * time.Second)
    for range ticker.C {
        var m runtime.MemStats
        runtime.ReadMemStats(&m)
        memoryUtilization.WithLabelValues("heap").Set(float64(m.HeapAlloc))
        memoryUtilization.WithLabelValues("stack").Set(float64(m.StackInuse))
        memoryUtilization.WithLabelValues("total").Set(float64(m.Sys))
    }
}()
```

**Prometheus Query**:
```promql
# Heap memory in MB
{service}_memory_utilization_bytes{type="heap"} / 1024 / 1024

# Memory growth rate (bytes/sec)
rate({service}_memory_utilization_bytes{type="heap"}[1h])
```

**Alerting**:
```yaml
- alert: HighMemoryUtilization
  expr: {service}_memory_utilization_bytes{type="heap"} > 500 * 1024 * 1024
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "Heap memory > 500MB"

- alert: MemoryLeak
  expr: rate({service}_memory_utilization_bytes{type="heap"}[1h]) > 1024 * 1024
  for: 10m
  labels:
    severity: critical
  annotations:
    summary: "Heap growing > 1MB/sec (possible leak)"
```

### Goroutines/Threads

**Metric Name**: `{service}_goroutines_active` (Go) or `{service}_threads_active`
**Type**: Gauge
**Unit**: Count

```prometheus
{service}_goroutines_active
```

**Collection**:
- **Go**: `runtime.NumGoroutine()`
- **Java**: `Thread.activeCount()`
- **Python**: `threading.active_count()`

**Example (Go)**:
```go
var goroutinesActive = prometheus.NewGauge(
    prometheus.GaugeOpts{
        Name: "myservice_goroutines_active",
        Help: "Number of active goroutines",
    },
)

go func() {
    ticker := time.NewTicker(10 * time.Second)
    for range ticker.C {
        goroutinesActive.Set(float64(runtime.NumGoroutine()))
    }
}()
```

**Alerting**:
```yaml
- alert: GoroutineLeak
  expr: {service}_goroutines_active > 1000
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "More than 1000 active goroutines (possible leak)"
```

### File Descriptors

**Metric Name**: `{service}_file_descriptors_open`
**Type**: Gauge
**Unit**: Count

```prometheus
{service}_file_descriptors_open
```

**Collection**:
- **Linux**: Count files in `/proc/{pid}/fd/`
- **macOS**: `lsof -p {pid} | wc -l`

**Example (Go)**:
```go
var fileDescriptorsOpen = prometheus.NewGauge(
    prometheus.GaugeOpts{
        Name: "myservice_file_descriptors_open",
        Help: "Number of open file descriptors",
    },
)

go func() {
    ticker := time.NewTicker(10 * time.Second)
    for range ticker.C {
        fdCount := getOpenFileDescriptors()  // Count /proc/{pid}/fd/
        fileDescriptorsOpen.Set(float64(fdCount))
    }
}()
```

**Alerting**:
```yaml
- alert: HighFileDescriptorUsage
  expr: {service}_file_descriptors_open / 1024 > 0.8
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "File descriptor usage > 80% of ulimit"
```

---

## 2. Saturation Metrics

**Purpose**: Is the resource queuing work it cannot service?

**Metric Type**: Gauge (queue depth) or Counter (saturation events)

### Request Queue Depth

**Metric Name**: `{service}_request_queue_depth`
**Type**: Gauge
**Unit**: Count

```prometheus
{service}_request_queue_depth
```

**Collection**:
- Track pending requests in server handler
- Increment on request arrival, decrement on processing start

**Example (Go)**:
```go
var requestQueueDepth = prometheus.NewGauge(
    prometheus.GaugeOpts{
        Name: "myservice_request_queue_depth",
        Help: "Number of requests waiting to be processed",
    },
)

var requestQueue atomic.Int32

func handleRequest(req Request) {
    requestQueue.Inc()
    requestQueueDepth.Set(float64(requestQueue.Load()))
    defer func() {
        requestQueue.Dec()
        requestQueueDepth.Set(float64(requestQueue.Load()))
    }()
    // ... process request ...
}
```

**Alerting**:
```yaml
- alert: RequestQueueSaturated
  expr: {service}_request_queue_depth > 10
  for: 2m
  labels:
    severity: warning
  annotations:
    summary: "Request queue depth > 10 (service saturated)"
```

### Concurrent Requests

**Metric Name**: `{service}_concurrent_requests`
**Type**: Gauge
**Unit**: Count

```prometheus
{service}_concurrent_requests
```

**Collection**:
- Increment on request start, decrement on completion

**Example (Go)**:
```go
var concurrentRequests = prometheus.NewGauge(
    prometheus.GaugeOpts{
        Name: "myservice_concurrent_requests",
        Help: "Number of requests currently being processed",
    },
)

var concurrentReqs atomic.Int32

func handleRequest(req Request) {
    concurrentReqs.Inc()
    concurrentRequests.Set(float64(concurrentReqs.Load()))
    defer func() {
        concurrentReqs.Dec()
        concurrentRequests.Set(float64(concurrentReqs.Load()))
    }()
    // ... process request ...
}
```

**Alerting**:
```yaml
- alert: HighConcurrency
  expr: {service}_concurrent_requests > 50
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "Concurrent requests > 50 (high load)"
```

### Memory Saturation (GC Pressure)

**Metric Name**: `{service}_memory_pressure_events_total`
**Type**: Counter
**Unit**: Count

```prometheus
{service}_memory_pressure_events_total
```

**Collection**:
- Track GC frequency, flag pressure if GC rate > threshold (e.g., 10 GC/sec)

**Example (Go)**:
```go
var memoryPressureEvents = prometheus.NewCounter(
    prometheus.CounterOpts{
        Name: "myservice_memory_pressure_events_total",
        Help: "Number of GC pressure events",
    },
)

var prevNumGC uint32

go func() {
    ticker := time.NewTicker(10 * time.Second)
    for range ticker.C {
        var m runtime.MemStats
        runtime.ReadMemStats(&m)
        gcDelta := m.NumGC - prevNumGC
        prevNumGC = m.NumGC
        gcRate := float64(gcDelta) / 10.0  // GC/sec
        if gcRate > 10.0 {
            memoryPressureEvents.Inc()
        }
    }
}()
```

**Alerting**:
```yaml
- alert: MemoryPressure
  expr: rate({service}_memory_pressure_events_total[5m]) > 0.1
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "Frequent GC pressure events (memory pressure detected)"
```

---

## 3. Resource Error Metrics

**Purpose**: Is the resource failing?

**Metric Type**: Counter

### Resource Errors

**Metric Name**: `{service}_resource_errors_total`
**Type**: Counter
**Labels**: resource_type (memory, cpu, file_descriptors, goroutines)
**Unit**: Count

```prometheus
{service}_resource_errors_total{resource_type="memory"}
{service}_resource_errors_total{resource_type="file_descriptors"}
```

**Collection**:
- Increment on resource allocation failures (OOM, FD exhaustion, etc.)

**Example (Go)**:
```go
var resourceErrors = prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "myservice_resource_errors_total",
        Help: "Resource-related errors",
    },
    []string{"resource_type"},
)

func handleError(err error) {
    if errors.Is(err, syscall.ENOMEM) {
        resourceErrors.WithLabelValues("memory").Inc()
    } else if errors.Is(err, syscall.EMFILE) {
        resourceErrors.WithLabelValues("file_descriptors").Inc()
    }
}
```

**Alerting**:
```yaml
- alert: ResourceError
  expr: rate({service}_resource_errors_total[5m]) > 0.01
  for: 1m
  labels:
    severity: critical
  annotations:
    summary: "Resource errors detected (OOM, FD exhaustion, etc.)"
```

---

## USE Dashboard Template

### Dashboard: "{Service} - USE Overview"

**Row 1: Utilization Gauges**
- **Panel**: CPU Utilization (%)
  - Query: `{service}_cpu_utilization_percent`
  - Visualization: Gauge
  - Threshold: Green < 60%, Yellow 60-80%, Red > 80%
- **Panel**: Memory Utilization (Heap)
  - Query: `{service}_memory_utilization_bytes{type="heap"} / 1024 / 1024`
  - Visualization: Gauge (MB)
  - Threshold: Green < 200MB, Yellow 200-500MB, Red > 500MB
- **Panel**: Goroutines Active
  - Query: `{service}_goroutines_active`
  - Visualization: Gauge
  - Threshold: Green < 50, Yellow 50-1000, Red > 1000

**Row 2: Utilization Trends**
- **Panel**: CPU Utilization Over Time
  - Query: `{service}_cpu_utilization_percent`
  - Visualization: Line graph
- **Panel**: Memory Utilization Over Time
  - Query: `{service}_memory_utilization_bytes{type="heap"} / 1024 / 1024`
  - Visualization: Line graph

**Row 3: Saturation**
- **Panel**: Request Queue Depth
  - Query: `{service}_request_queue_depth`
  - Visualization: Line graph
  - Threshold: Yellow > 5, Red > 10
- **Panel**: Concurrent Requests
  - Query: `{service}_concurrent_requests`
  - Visualization: Line graph
  - Threshold: Yellow > 20, Red > 50
- **Panel**: GC Pressure Events
  - Query: `rate({service}_memory_pressure_events_total[5m])`
  - Visualization: Line graph

**Row 4: Resource Errors**
- **Panel**: Resource Errors by Type
  - Query: `sum(rate({service}_resource_errors_total[5m])) by (resource_type)`
  - Visualization: Bar chart
- **Panel**: Resource Error Rate
  - Query: `sum(rate({service}_resource_errors_total[5m]))`
  - Visualization: Line graph

---

## Transfer Checklist

When applying USE pattern to a new service:

1. **Identify critical resources**
   - [ ] List all resources (CPU, memory, disk, network, goroutines, FDs, etc.)
   - [ ] Prioritize by criticality (CPU and memory usually most critical)

2. **For each resource, implement USE metrics:**
   - [ ] **Utilization**: How busy? (Gauge, 0-100% or absolute)
   - [ ] **Saturation**: Queuing work? (Gauge for queue depth, Counter for events)
   - [ ] **Errors**: Failing? (Counter for error events)

3. **Implement collection**
   - [ ] Background goroutines for sampling (10s interval typical)
   - [ ] Use runtime APIs (`runtime.ReadMemStats()`, `/proc/{pid}/stat`, etc.)

4. **Create USE dashboard**
   - [ ] Row 1: Utilization gauges (at-a-glance status)
   - [ ] Row 2: Utilization trends (time series)
   - [ ] Row 3: Saturation metrics (queue depths, concurrency)
   - [ ] Row 4: Resource errors (failures)

5. **Configure alerting**
   - [ ] Alert on high utilization (CPU > 80%, memory > threshold)
   - [ ] Alert on saturation (queue depth > 10, concurrency > limit)
   - [ ] Alert on resource errors (OOM, FD exhaustion)

---

## Real-World Examples

### Web Server
```
CPU:
  Utilization: webserver_cpu_utilization_percent
  Saturation:  webserver_cpu_saturation_percent (load average)
  Errors:      (rare)

Memory:
  Utilization: webserver_memory_utilization_bytes{type="heap"}
  Saturation:  webserver_memory_pressure_events_total (paging, swapping)
  Errors:      webserver_resource_errors_total{resource_type="memory"} (OOM)

Network:
  Utilization: webserver_network_utilization_bytes_per_sec
  Saturation:  webserver_network_queue_depth
  Errors:      webserver_network_errors_total (packet loss, timeouts)
```

### Database Server
```
CPU:
  Utilization: db_cpu_utilization_percent
  Saturation:  db_cpu_saturation_percent (run queue depth)
  Errors:      (rare)

Disk:
  Utilization: db_disk_utilization_percent
  Saturation:  db_disk_queue_depth
  Errors:      db_disk_errors_total (I/O errors)

Locks:
  Utilization: db_lock_utilization_percent
  Saturation:  db_lock_wait_queue_depth
  Errors:      db_deadlock_events_total
```

---

## RED + USE Combined

### Diagnostic Workflow

**Scenario**: High latency detected (RED)

1. **RED**: p95 latency > 5s (from `request_duration_seconds`)
2. **USE**: Check resource utilization
   - CPU utilization > 90% → CPU bottleneck
   - Memory pressure events spiking → Memory pressure
   - Request queue depth > 20 → Saturation
3. **Correlation**: High latency coincides with high CPU → CPU is root cause

**Result**: USE metrics explain RED symptoms

---

## Implementation Checklist

- [ ] **Identify resources**: List all critical resources
- [ ] **Define USE metrics**: For each resource, define Utilization, Saturation, Errors
- [ ] **Implement collection**: Background goroutines, runtime APIs, /proc
- [ ] **Verify overhead**: < 2% CPU, < 10MB memory for metrics collection
- [ ] **Create USE dashboard**: Utilization gauges, trends, saturation, errors
- [ ] **Configure alerting**: High utilization, saturation, resource errors
- [ ] **Document thresholds**: CPU > 80%, memory > limit, queue depth > 10

---

## References

- Brendan Gregg, "The USE Method" (2013): http://www.brendangregg.com/usemethod.html
- Brendan Gregg, "Systems Performance: Enterprise and the Cloud" (2013)
- Prometheus Best Practices: https://prometheus.io/docs/practices/naming/
- Go runtime documentation: https://pkg.go.dev/runtime

---

**Pattern Status**: Validated (Bootstrap-009 Iteration 3)
**Transferability**: Universal (any system)
**Complementary**: Use with RED for comprehensive observability
**Next Application**: Bootstrap-009 Iteration 5 (MCP server implementation, after RED in Iteration 4)
