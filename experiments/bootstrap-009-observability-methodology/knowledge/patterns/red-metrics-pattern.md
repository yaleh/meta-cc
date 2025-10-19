# RED Metrics Pattern (Universal)

**Pattern Type**: Observability / Metrics
**Domain**: Request-driven services
**Transferability**: Very high (applicable to any request/response service)
**Source**: Tom Wilkie (Weaveworks), Prometheus best practices
**Created**: 2025-10-17 (Bootstrap-009 Iteration 3)

---

## Pattern Overview

**RED** (Rate, Errors, Duration) is a metrics methodology for monitoring the health of request-driven services from the user's perspective. It provides three golden metrics that capture service quality and performance.

### The Three Metrics

1. **Rate**: How many requests per second
2. **Errors**: How many requests are failing
3. **Duration**: How long requests take to complete

### When to Use RED

- **Request-driven services**: HTTP APIs, RPC services, message processors, database queries
- **User-facing services**: When user experience depends on request success/latency
- **SLO monitoring**: When you need to track service-level objectives

### When NOT to Use RED

- **Batch processing**: No clear request/response pattern
- **Event streams**: Continuous data flows without discrete requests
- **Pure data pipelines**: Focus on throughput, not individual request latency

---

## Universal Implementation Pattern

### 1. Rate Metrics

**Metric Name**: `{service}_requests_total`
**Type**: Counter
**Labels**: endpoint, method, status

```prometheus
{service}_requests_total{endpoint="...", method="...", status="..."}
```

**Purpose**: Track how many requests are being served

**Prometheus Queries**:
```promql
# Requests per second (5-minute window)
rate({service}_requests_total[5m])

# Requests per second by endpoint
sum(rate({service}_requests_total[5m])) by (endpoint)

# Requests per second by status
sum(rate({service}_requests_total[5m])) by (status)
```

**Instrumentation Points**:
- Increment on every request received
- Label with endpoint, method, status

**Example (Go)**:
```go
var requestsTotal = prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "myservice_requests_total",
        Help: "Total number of requests",
    },
    []string{"endpoint", "method", "status"},
)

func handleRequest(req Request) {
    defer func() {
        requestsTotal.WithLabelValues(
            req.Endpoint,
            req.Method,
            status,
        ).Inc()
    }()
    // ... process request ...
}
```

### 2. Error Metrics

**Metric Name**: `{service}_errors_total`
**Type**: Counter
**Labels**: endpoint, error_type

```prometheus
{service}_errors_total{endpoint="...", error_type="..."}
```

**Purpose**: Track how many requests are failing

**Computed Metric**: Error Rate (%)
```promql
# Error rate as percentage
sum(rate({service}_errors_total[5m])) / sum(rate({service}_requests_total[5m])) * 100
```

**Instrumentation Points**:
- Increment on every request failure
- Label with endpoint, error type (validation, execution, timeout, etc.)

**Example (Go)**:
```go
var errorsTotal = prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "myservice_errors_total",
        Help: "Total number of errors",
    },
    []string{"endpoint", "error_type"},
)

func handleRequest(req Request) error {
    err := processRequest(req)
    if err != nil {
        errorsTotal.WithLabelValues(
            req.Endpoint,
            classifyError(err),
        ).Inc()
        return err
    }
    return nil
}
```

### 3. Duration Metrics

**Metric Name**: `{service}_request_duration_seconds`
**Type**: Histogram
**Labels**: endpoint, status
**Buckets**: [0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0, 5.0, 10.0, 30.0]

```prometheus
{service}_request_duration_seconds{endpoint="...", status="..."}
```

**Purpose**: Track how long requests take (percentiles: p50, p95, p99)

**Prometheus Queries**:
```promql
# p95 latency (5-minute window)
histogram_quantile(0.95, rate({service}_request_duration_seconds_bucket[5m]))

# p50, p95, p99 latency
histogram_quantile(0.50, rate({service}_request_duration_seconds_bucket[5m]))
histogram_quantile(0.95, rate({service}_request_duration_seconds_bucket[5m]))
histogram_quantile(0.99, rate({service}_request_duration_seconds_bucket[5m]))

# p95 latency by endpoint
histogram_quantile(0.95, sum(rate({service}_request_duration_seconds_bucket[5m])) by (endpoint, le))
```

**Instrumentation Points**:
- Measure request start to completion
- Observe duration in seconds

**Example (Go)**:
```go
var requestDuration = prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
        Name:    "myservice_request_duration_seconds",
        Help:    "Request duration in seconds",
        Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0, 5.0, 10.0, 30.0},
    },
    []string{"endpoint", "status"},
)

func handleRequest(req Request) {
    start := time.Now()
    defer func() {
        duration := time.Since(start)
        requestDuration.WithLabelValues(
            req.Endpoint,
            status,
        ).Observe(duration.Seconds())
    }()
    // ... process request ...
}
```

---

## Cardinality Control

**Critical**: Avoid high-cardinality labels to prevent metric explosion

### Good Labels (Low Cardinality)
- `endpoint` (10-100 unique values)
- `method` (GET, POST, PUT, DELETE - 4 values)
- `status` (success, error, timeout - 3-10 values)
- `error_type` (validation, execution, network - 5-10 values)

### Bad Labels (High Cardinality)
- `user_id` (millions of users → millions of time series)
- `request_id` (every request unique → unbounded cardinality)
- `ip_address` (thousands of IPs)
- `session_id` (thousands of sessions)

### Cardinality Example
```
Good: 50 endpoints × 4 methods × 3 statuses = 600 time series ✓
Bad:  1 endpoint × 1,000,000 request_ids = 1,000,000 time series ✗
```

---

## RED Dashboard Template

### Dashboard: "{Service} - RED Overview"

**Row 1: Rate**
- **Panel**: Request Rate (req/s)
  - Query: `sum(rate({service}_requests_total[5m]))`
  - Visualization: Single stat with sparkline
- **Panel**: Request Rate by Endpoint
  - Query: `sum(rate({service}_requests_total[5m])) by (endpoint)`
  - Visualization: Stacked area graph

**Row 2: Errors**
- **Panel**: Error Rate (%)
  - Query: `sum(rate({service}_errors_total[5m])) / sum(rate({service}_requests_total[5m])) * 100`
  - Visualization: Single stat with threshold (Green < 1%, Yellow 1-5%, Red > 5%)
- **Panel**: Error Rate Over Time
  - Query: Same as above
  - Visualization: Line graph

**Row 3: Duration**
- **Panel**: Latency p50/p95/p99
  - Queries:
    - p50: `histogram_quantile(0.50, rate({service}_request_duration_seconds_bucket[5m]))`
    - p95: `histogram_quantile(0.95, rate({service}_request_duration_seconds_bucket[5m]))`
    - p99: `histogram_quantile(0.99, rate({service}_request_duration_seconds_bucket[5m]))`
  - Visualization: Multi-line graph
- **Panel**: p95 Latency by Endpoint
  - Query: `histogram_quantile(0.95, sum(rate({service}_request_duration_seconds_bucket[5m])) by (endpoint, le))`
  - Visualization: Multi-line graph

---

## Alerting Rules Template

### Rate Alerts

```yaml
- alert: RequestRateDropped
  expr: sum(rate({service}_requests_total[5m])) < 0.1 * avg_over_time(sum(rate({service}_requests_total[5m]))[1h:5m])
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "Request rate dropped to < 10% of 1-hour average"
```

### Error Alerts

```yaml
- alert: HighErrorRate
  expr: sum(rate({service}_errors_total[5m])) / sum(rate({service}_requests_total[5m])) > 0.05
  for: 5m
  labels:
    severity: critical
  annotations:
    summary: "Error rate > 5% (SLO violation)"

- alert: EndpointErrorRateHigh
  expr: sum(rate({service}_errors_total[5m])) by (endpoint) / sum(rate({service}_requests_total[5m])) by (endpoint) > 0.10
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "Endpoint error rate > 10%"
```

### Duration Alerts

```yaml
- alert: HighP95Latency
  expr: histogram_quantile(0.95, rate({service}_request_duration_seconds_bucket[5m])) > 5.0
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "p95 latency > 5s (SLO violation)"

- alert: HighP99Latency
  expr: histogram_quantile(0.99, rate({service}_request_duration_seconds_bucket[5m])) > 10.0
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "p99 latency > 10s (tail latency issue)"
```

---

## Transfer Checklist

When applying RED pattern to a new service:

1. **Identify request/response pattern**
   - [ ] Service is request-driven (not batch/streaming)
   - [ ] Request unit is well-defined (HTTP request, RPC call, message, query)

2. **Implement Rate metric**
   - [ ] Create `{service}_requests_total` counter
   - [ ] Add labels: endpoint, method, status
   - [ ] Increment on every request
   - [ ] Verify cardinality < 1000 series

3. **Implement Errors metric**
   - [ ] Create `{service}_errors_total` counter
   - [ ] Add labels: endpoint, error_type
   - [ ] Increment on every failure
   - [ ] Classify errors into 5-10 categories

4. **Implement Duration metric**
   - [ ] Create `{service}_request_duration_seconds` histogram
   - [ ] Add labels: endpoint, status
   - [ ] Choose appropriate buckets
   - [ ] Observe duration on request completion

5. **Create RED dashboard**
   - [ ] Add Rate panels (req/s, by endpoint)
   - [ ] Add Error panels (error rate %, by type)
   - [ ] Add Duration panels (p50/p95/p99, by endpoint)

6. **Configure alerting**
   - [ ] Alert on error rate > SLO (e.g., 5%)
   - [ ] Alert on p95 latency > SLO (e.g., 5s)
   - [ ] Alert on request rate anomalies

---

## Real-World Examples

### HTTP REST API
```
Rate:     http_requests_total{endpoint="/api/users", method="GET", status="200"}
Errors:   http_errors_total{endpoint="/api/users", error_type="validation_error"}
Duration: http_request_duration_seconds{endpoint="/api/users", status="200"}
```

### gRPC Service
```
Rate:     grpc_requests_total{method="GetUser", status="OK"}
Errors:   grpc_errors_total{method="GetUser", error_type="UNAUTHENTICATED"}
Duration: grpc_request_duration_seconds{method="GetUser", status="OK"}
```

### Message Queue Consumer
```
Rate:     queue_messages_processed_total{queue="orders", status="success"}
Errors:   queue_messages_failed_total{queue="orders", error_type="parse_error"}
Duration: queue_message_processing_duration_seconds{queue="orders", status="success"}
```

### Database Query Engine
```
Rate:     db_queries_total{query_type="SELECT", status="success"}
Errors:   db_query_errors_total{query_type="SELECT", error_type="timeout"}
Duration: db_query_duration_seconds{query_type="SELECT", status="success"}
```

---

## Relationship to Other Methodologies

### RED vs Four Golden Signals
- **Four Golden Signals**: Latency, Traffic, Errors, Saturation (Google SRE Book)
- **RED**: Rate (→ Traffic), Errors (→ Errors), Duration (→ Latency)
- **Difference**: Four Golden Signals adds Saturation (capacity/resource utilization)

### RED vs USE
- **RED**: Request flow (user perspective)
- **USE**: Resource health (system perspective: Utilization, Saturation, Errors)
- **Combined**: RED + USE = comprehensive service + system observability

### When to Combine
- Use **RED** for request-driven services (always)
- Add **USE** for capacity planning and resource monitoring
- Result: Full observability stack

---

## Implementation Checklist

- [ ] **Library**: Add Prometheus client library
- [ ] **Metrics**: Define Rate, Errors, Duration metrics with labels
- [ ] **Instrumentation**: Add metric collection to request handling
- [ ] **Cardinality**: Verify label cardinality < 1000 series
- [ ] **Exposition**: Expose /metrics endpoint (Prometheus format)
- [ ] **Scraping**: Configure Prometheus to scrape /metrics
- [ ] **Dashboard**: Create RED dashboard in Grafana
- [ ] **Alerting**: Configure alerting rules for SLO violations
- [ ] **Documentation**: Document metric meanings and thresholds

---

## References

- Tom Wilkie, "The RED Method: How to instrument your services" (2017)
- Prometheus Best Practices: https://prometheus.io/docs/practices/naming/
- Google SRE Book: "Monitoring Distributed Systems" (Four Golden Signals)
- Brendan Gregg, "The USE Method" (comparison)

---

**Pattern Status**: Validated (Bootstrap-009 Iteration 3)
**Transferability**: Universal (any request-driven service)
**Next Application**: Bootstrap-009 Iteration 4 (MCP server implementation)
