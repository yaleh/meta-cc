# Distributed Tracing Pattern (Universal)

**Pattern Type**: Observability / Distributed Tracing
**Domain**: Distributed systems, microservices, request-driven services
**Transferability**: Very high (applicable to any system with request flows)
**Source**: OpenTelemetry, W3C Trace Context, Dapper (Google)
**Created**: 2025-10-17 (Bootstrap-009 Iteration 6)

---

## Pattern Overview

**Distributed Tracing** is an observability methodology for tracking request flows through distributed systems. It provides end-to-end visibility by creating hierarchical traces that show how requests propagate across services, components, and operations.

### Core Concepts

1. **Trace**: A complete journey of a request through the system (tree structure)
2. **Span**: A single unit of work within a trace (node in the tree)
3. **Context Propagation**: Passing trace metadata across operation boundaries
4. **Sampling**: Selectively collecting traces to control overhead

### When to Use Distributed Tracing

- **Distributed systems**: Multiple services or components handling a request
- **Complex request flows**: Operations with multiple steps or dependencies
- **Performance debugging**: Identifying latency bottlenecks in request paths
- **Dependency analysis**: Understanding service interactions and call graphs

### When NOT to Use Distributed Tracing

- **Single-process applications**: Logging may be sufficient
- **Pure batch jobs**: No request flow to trace
- **Extremely high-throughput systems**: Sampling required, or overhead unacceptable

---

## Universal Implementation Pattern

### 1. Trace Provider Initialization

**Purpose**: Initialize the tracing framework and configure exporters

**OpenTelemetry (Go)**:
```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

func InitTracing() (func(), error) {
    // Create exporter (stdout for dev, OTLP for production)
    exporter, err := stdouttrace.New(
        stdouttrace.WithPrettyPrint(),
    )
    if err != nil {
        return nil, err
    }

    // Create resource with service information
    res, err := resource.New(context.Background(),
        resource.WithAttributes(
            semconv.ServiceName("myservice"),
            semconv.ServiceVersion("1.0.0"),
        ),
    )
    if err != nil {
        return nil, err
    }

    // Configure trace provider
    tp := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exporter),
        sdktrace.WithResource(res),
        sdktrace.WithSampler(sdktrace.AlwaysSample()), // or TraceIDRatioBased
    )

    // Register as global trace provider
    otel.SetTracerProvider(tp)

    // Return cleanup function
    cleanup := func() {
        _ = tp.Shutdown(context.Background())
    }

    return cleanup, nil
}
```

**Exporter Options**:
- **stdout**: Development/testing (human-readable)
- **OTLP**: Production (OpenTelemetry Protocol, to Jaeger/Zipkin/Tempo)
- **Jaeger**: Direct export to Jaeger backend
- **Zipkin**: Direct export to Zipkin backend

---

### 2. Span Creation (Request Tracing)

**Purpose**: Create root span for incoming requests

**Pattern**:
```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/codes"
    "go.opentelemetry.io/otel/trace"
)

func handleRequest(req Request) {
    // Get tracer
    tracer := otel.Tracer("myservice")

    // Create root span
    ctx, span := tracer.Start(context.Background(), "http.request",
        trace.WithAttributes(
            attribute.String("http.method", req.Method),
            attribute.String("http.url", req.URL),
            attribute.String("http.user_agent", req.UserAgent),
        ),
    )
    defer span.End()

    // Process request with context
    result, err := processRequest(ctx, req)

    // Record span status
    if err != nil {
        span.SetStatus(codes.Error, err.Error())
        span.RecordError(err)
    } else {
        span.SetStatus(codes.Ok, "success")
        span.SetAttributes(
            attribute.Int("http.status_code", 200),
            attribute.Int("response.size", len(result)),
        )
    }
}
```

**Span Naming**:
- Use semantic names: `http.request`, `db.query`, `rpc.call`
- Follow OpenTelemetry conventions: `{operation_type}.{operation_name}`
- Avoid high-cardinality names (no request IDs, user IDs)

---

### 3. Child Span Creation (Operation Tracing)

**Purpose**: Create child spans for sub-operations

**Pattern**:
```go
func processRequest(ctx context.Context, req Request) (Result, error) {
    tracer := otel.Tracer("myservice")

    // Create child span from parent context
    ctx, span := tracer.Start(ctx, "process.execute",
        trace.WithAttributes(
            attribute.String("operation", "data_processing"),
        ),
    )
    defer span.End()

    // Step 1: Database query (child span)
    data, err := fetchDataFromDB(ctx, req.ID)
    if err != nil {
        span.RecordError(err)
        return Result{}, err
    }

    // Step 2: Business logic (child span)
    result, err := applyBusinessLogic(ctx, data)
    if err != nil {
        span.RecordError(err)
        return Result{}, err
    }

    span.SetAttributes(
        attribute.Int("records.processed", len(data)),
    )

    return result, nil
}

func fetchDataFromDB(ctx context.Context, id string) (Data, error) {
    tracer := otel.Tracer("myservice")

    ctx, span := tracer.Start(ctx, "db.query",
        trace.WithAttributes(
            attribute.String("db.system", "postgresql"),
            attribute.String("db.statement", "SELECT * FROM users WHERE id = ?"),
        ),
    )
    defer span.End()

    // Execute query...
    return data, nil
}
```

**Hierarchy Example**:
```
Trace ID: abc123...
├─ Span: http.request (root)
   ├─ Span: process.execute
      ├─ Span: db.query
      ├─ Span: business_logic.apply
      └─ Span: cache.set
   └─ Span: http.response
```

---

### 4. Context Propagation (W3C Trace Context)

**Purpose**: Pass trace context across service boundaries

**HTTP Headers (W3C Trace Context)**:
```
traceparent: 00-<trace-id>-<span-id>-<flags>
tracestate: vendor1=value1,vendor2=value2
```

**Outgoing Request (Client)**:
```go
import (
    "go.opentelemetry.io/otel/propagation"
)

func callExternalService(ctx context.Context, url string) error {
    tracer := otel.Tracer("myservice")

    // Create span for outgoing request
    ctx, span := tracer.Start(ctx, "http.client.request",
        trace.WithAttributes(
            attribute.String("http.url", url),
        ),
    )
    defer span.End()

    // Create HTTP request
    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)

    // Inject trace context into HTTP headers
    propagator := otel.GetTextMapPropagator()
    propagator.Inject(ctx, propagation.HeaderCarrier(req.Header))

    // Execute request
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        span.RecordError(err)
        return err
    }

    return nil
}
```

**Incoming Request (Server)**:
```go
func handleHTTPRequest(w http.ResponseWriter, r *http.Request) {
    tracer := otel.Tracer("myservice")

    // Extract trace context from HTTP headers
    propagator := otel.GetTextMapPropagator()
    ctx := propagator.Extract(r.Context(), propagation.HeaderCarrier(r.Header))

    // Create span with extracted context (becomes child of remote span)
    ctx, span := tracer.Start(ctx, "http.server.request",
        trace.WithAttributes(
            attribute.String("http.method", r.Method),
        ),
    )
    defer span.End()

    // Process request with propagated context
    processRequest(ctx, r)
}
```

---

### 5. Trace-Log Correlation

**Purpose**: Link traces and logs for unified debugging

**Pattern**:
```go
import (
    "log/slog"
    "go.opentelemetry.io/otel/trace"
)

func GetTraceID(ctx context.Context) string {
    span := trace.SpanFromContext(ctx)
    if span.SpanContext().IsValid() {
        return span.SpanContext().TraceID().String()
    }
    return ""
}

func GetSpanID(ctx context.Context) string {
    span := trace.SpanFromContext(ctx)
    if span.SpanContext().IsValid() {
        return span.SpanContext().SpanID().String()
    }
    return ""
}

func processRequest(ctx context.Context, req Request) {
    traceID := GetTraceID(ctx)
    spanID := GetSpanID(ctx)

    slog.Info("processing request",
        "trace_id", traceID,
        "span_id", spanID,
        "request_id", req.ID,
    )

    // ... process request ...
}
```

**Workflow**:
1. Trace shows slow request (trace ID: `abc123...`)
2. Extract trace ID from trace viewer
3. Query logs: `SELECT * FROM logs WHERE trace_id = 'abc123...'`
4. See detailed logs for that specific request

---

### 6. Sampling Strategies

**Purpose**: Control overhead by sampling subset of traces

**Sampling Types**:

1. **AlwaysSample**: Sample 100% (development/testing)
   ```go
   sdktrace.WithSampler(sdktrace.AlwaysSample())
   ```

2. **TraceIDRatioBased**: Sample X% of traces
   ```go
   // Sample 10% of traces
   sdktrace.WithSampler(sdktrace.TraceIDRatioBased(0.1))
   ```

3. **ParentBased**: Inherit parent's sampling decision
   ```go
   sdktrace.WithSampler(sdktrace.ParentBased(
       sdktrace.TraceIDRatioBased(0.1),
   ))
   ```

4. **Custom**: Sample based on rules
   ```go
   type CustomSampler struct{}

   func (s CustomSampler) ShouldSample(params sdktrace.SamplingParameters) sdktrace.SamplingResult {
       // Always sample errors
       if isErrorRequest(params) {
           return sdktrace.SamplingResult{
               Decision: sdktrace.RecordAndSample,
           }
       }

       // Sample 1% of successful requests
       if rand.Float64() < 0.01 {
           return sdktrace.SamplingResult{
               Decision: sdktrace.RecordAndSample,
           }
       }

       return sdktrace.SamplingResult{
           Decision: sdktrace.Drop,
       }
   }
   ```

**Sampling Recommendations**:
- **Development**: AlwaysSample (100%)
- **Staging**: TraceIDRatioBased(0.5) (50%)
- **Production (low traffic)**: TraceIDRatioBased(0.1) (10%)
- **Production (high traffic)**: TraceIDRatioBased(0.01) (1%)
- **Production (critical paths)**: Custom sampler (always sample errors, 1% of success)

---

## Span Attributes (Semantic Conventions)

**HTTP Spans**:
```go
attribute.String("http.method", "GET")
attribute.String("http.url", "/api/users")
attribute.Int("http.status_code", 200)
attribute.String("http.user_agent", "...")
attribute.Int("http.request.body.size", 1024)
attribute.Int("http.response.body.size", 2048)
```

**Database Spans**:
```go
attribute.String("db.system", "postgresql")
attribute.String("db.name", "mydb")
attribute.String("db.statement", "SELECT * FROM users")
attribute.String("db.operation", "SELECT")
attribute.String("db.sql.table", "users")
```

**RPC Spans**:
```go
attribute.String("rpc.system", "grpc")
attribute.String("rpc.service", "UserService")
attribute.String("rpc.method", "GetUser")
attribute.String("rpc.grpc.status_code", "OK")
```

**Error Attributes**:
```go
attribute.String("error.type", "validation_error")
attribute.String("error.message", err.Error())
attribute.String("error.stack", stackTrace)
```

---

## Trace Analysis Workflows

### 1. Latency Analysis

**Workflow**:
1. Identify slow request in trace viewer (p95 > SLO)
2. View trace waterfall (time spent in each span)
3. Identify bottleneck span (longest duration)
4. Drill into span attributes (query, endpoint, etc.)
5. Correlate with logs using trace_id
6. Identify root cause (slow query, external API timeout, etc.)

**Example**:
```
Trace: abc123... (total: 5.2s, p95 SLO: 2.0s) ❌
├─ http.request (5.2s)
   ├─ auth.validate (0.1s) ✓
   ├─ db.query (4.8s) ⚠️ BOTTLENECK
   │  └─ SELECT * FROM users WHERE ... (missing index!)
   └─ response.serialize (0.3s) ✓

Root Cause: Missing index on users table → 4.8s query latency
```

### 2. Error Analysis

**Workflow**:
1. Filter traces by error status (span status = Error)
2. Group by error type (error.type attribute)
3. Identify error pattern (common span, service, endpoint)
4. View error trace timeline (when did error occur?)
5. Correlate with logs using trace_id
6. Identify root cause (validation, network, dependency failure)

**Example**:
```
Trace: def456... (status: Error)
├─ http.request (span status: OK)
   ├─ db.query (span status: OK)
   └─ external_api.call (span status: Error) ⚠️
      └─ error.type: "timeout"
         error.message: "context deadline exceeded"

Root Cause: External API timeout (no circuit breaker)
```

### 3. Dependency Analysis

**Workflow**:
1. View service dependency graph (aggregated traces)
2. Identify service-to-service call patterns
3. Measure latency contribution by dependency
4. Identify critical path (longest dependency chain)
5. Optimize high-latency dependencies

**Example**:
```
Service Dependency Graph:
API Gateway → Auth Service (20ms)
API Gateway → User Service (50ms)
User Service → DB (30ms)
User Service → Cache (5ms)

Critical Path: API Gateway → User Service → DB (80ms total)
Optimization: Add caching to reduce DB calls
```

---

## Integration with Metrics and Logs

### Three Pillars of Observability

**1. Logs**: Individual events (what happened?)
```
INFO: User login successful (user_id=123, trace_id=abc...)
```

**2. Metrics**: Aggregated trends (how often? how fast?)
```
http_requests_total{endpoint="/login"} = 1000
http_request_duration_seconds{endpoint="/login", p95} = 0.5s
```

**3. Traces**: Request flows (why slow? where failed?)
```
Trace abc...: http.request → auth.validate → db.query (500ms)
```

### Correlation Workflow

**Scenario**: High p95 latency alert

1. **Metrics**: Alert triggers (p95 latency > 2s)
2. **Traces**: Find slow traces (filter by duration > 2s)
3. **Logs**: Extract trace_id, query logs (`trace_id = abc123...`)
4. **Root Cause**: Logs show "Database connection pool exhausted"
5. **Fix**: Increase connection pool size

### Unified Debugging

**Tools**:
- **Jaeger**: Trace visualization
- **Grafana**: Metrics dashboards + trace correlation
- **Loki**: Log aggregation with trace_id filtering
- **Tempo**: Trace backend (integrates with Grafana)

**Grafana Integration**:
```
Panel: p95 Latency (Prometheus)
→ Click spike in graph
→ "View Traces" button (Tempo)
→ Show traces with latency > threshold
→ Click trace
→ "View Logs" button (Loki)
→ Show logs with trace_id from trace
```

---

## Transfer Checklist

When applying distributed tracing to a new service:

1. **Add OpenTelemetry SDK**
   - [ ] Add dependency: `go.opentelemetry.io/otel`
   - [ ] Initialize trace provider on startup
   - [ ] Configure exporter (stdout/OTLP/Jaeger)
   - [ ] Set service name and version

2. **Instrument Request Entry Points**
   - [ ] Create root span for incoming requests
   - [ ] Add request attributes (method, URL, headers)
   - [ ] Extract W3C Trace Context from headers (if distributed)

3. **Instrument Internal Operations**
   - [ ] Create child spans for key operations (DB, cache, RPC)
   - [ ] Add operation-specific attributes
   - [ ] Record errors in spans

4. **Add Trace-Log Correlation**
   - [ ] Extract trace_id and span_id from context
   - [ ] Add to all log statements
   - [ ] Verify correlation in log viewer

5. **Configure Sampling**
   - [ ] Development: AlwaysSample (100%)
   - [ ] Production: TraceIDRatioBased (1-10%)
   - [ ] Consider custom sampling for errors

6. **Set Up Trace Backend**
   - [ ] Deploy Jaeger/Zipkin/Tempo
   - [ ] Configure OTLP exporter
   - [ ] Verify traces appear in backend

7. **Integrate with Metrics/Logs**
   - [ ] Link traces from metrics dashboards
   - [ ] Filter logs by trace_id
   - [ ] Create unified debugging workflow

---

## Real-World Examples

### HTTP REST API
```go
func handleHTTPRequest(w http.ResponseWriter, r *http.Request) {
    tracer := otel.Tracer("api-gateway")

    ctx, span := tracer.Start(r.Context(), "http.request",
        trace.WithAttributes(
            attribute.String("http.method", r.Method),
            attribute.String("http.url", r.URL.Path),
        ),
    )
    defer span.End()

    // Route to handler
    result, err := routeRequest(ctx, r)
    if err != nil {
        span.SetStatus(codes.Error, err.Error())
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    span.SetStatus(codes.Ok, "success")
    json.NewEncoder(w).Encode(result)
}
```

### gRPC Service
```go
func (s *UserServiceServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
    tracer := otel.Tracer("user-service")

    ctx, span := tracer.Start(ctx, "grpc.GetUser",
        trace.WithAttributes(
            attribute.String("rpc.service", "UserService"),
            attribute.String("rpc.method", "GetUser"),
            attribute.String("user.id", req.UserId),
        ),
    )
    defer span.End()

    user, err := s.fetchUserFromDB(ctx, req.UserId)
    if err != nil {
        span.RecordError(err)
        return nil, err
    }

    return user, nil
}
```

### Database Query
```go
func fetchUserFromDB(ctx context.Context, id string) (*User, error) {
    tracer := otel.Tracer("user-service")

    ctx, span := tracer.Start(ctx, "db.query",
        trace.WithAttributes(
            attribute.String("db.system", "postgresql"),
            attribute.String("db.statement", "SELECT * FROM users WHERE id = ?"),
            attribute.String("db.sql.table", "users"),
        ),
    )
    defer span.End()

    var user User
    err := db.QueryRowContext(ctx, "SELECT * FROM users WHERE id = ?", id).Scan(&user)
    if err != nil {
        span.RecordError(err)
        return nil, err
    }

    return &user, nil
}
```

---

## Performance Considerations

### Overhead

**Span creation**: ~1-5 microseconds per span
**Sampling**: AlwaysSample (100%) adds ~0.1-1% CPU overhead
**Sampling**: TraceIDRatioBased(0.01) (1%) adds ~0.001-0.01% CPU overhead

### Optimization

1. **Use sampling in production** (1-10% sufficient)
2. **Batch export traces** (default: every 5 seconds)
3. **Limit span attributes** (avoid unbounded string sizes)
4. **Avoid high-cardinality span names** (no IDs in span names)

---

## References

- OpenTelemetry: https://opentelemetry.io/docs/
- W3C Trace Context: https://www.w3.org/TR/trace-context/
- Dapper Paper (Google): "Dapper, a Large-Scale Distributed Systems Tracing Infrastructure"
- Jaeger: https://www.jaegertracing.io/
- Zipkin: https://zipkin.io/

---

**Pattern Status**: Validated (Bootstrap-009 Iteration 6)
**Transferability**: Universal (any distributed system or request-driven service)
**Next Application**: Production deployment with OTLP exporter and Tempo backend
