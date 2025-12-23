# RCIG Architecture

## Overview

RCIG (Real Time Code Intelligence Gateway) is a production-ready reverse proxy built in Go that provides real-time HTTP traffic analysis, anomaly detection, and alerting capabilities.

## System Components

```
┌─────────────────────────────────────────────────────────────────┐
│                           Client                                 │
└───────────────────────────────┬─────────────────────────────────┘
                                │
                                │ HTTP Request
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                        RCIG Proxy                                │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │   Proxy      │─▶│   Metrics    │─▶│   Analyzer   │          │
│  │   Handler    │  │   Registry   │  │              │          │
│  └──────────────┘  └──────────────┘  └──────┬───────┘          │
│         │                                    │                   │
│         │                                    ▼                   │
│         │                          ┌──────────────┐             │
│         │                          │    Alert     │             │
│         │                          │    Sinks     │             │
│         │                          └──────────────┘             │
│         │                                                        │
└─────────┼────────────────────────────────────────────────────────┘
          │
          │ Proxied Request
          ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Backend Service                             │
└─────────────────────────────────────────────────────────────────┘
```

## Component Details

### 1. Proxy Handler (`internal/proxy`)

**Responsibility**: Intercepts HTTP requests, forwards them to the backend, and records metrics.

**Key Features**:
- Reverse proxy using Go's `httputil.ReverseProxy`
- Request/response interception
- Latency measurement
- Status code tracking
- Thread-safe metrics recording

**Flow**:
1. Receive incoming HTTP request
2. Start timer
3. Forward request to backend via reverse proxy
4. Capture response status code
5. Calculate latency
6. Record metrics to registry
7. Return response to client

### 2. Metrics Registry (`internal/metrics`)

**Responsibility**: Stores and manages HTTP traffic metrics in memory.

**Key Features**:
- Thread-safe concurrent access
- Circular buffer for efficient storage
- Per-endpoint metric tracking (method + path)
- Snapshot generation for analysis
- HTTP endpoint for metrics export (`/rcig/metrics`)

**Data Structure**:
```go
type EndpointMetrics struct {
    key        EndpointKey       // Method + Path
    buffer     *CircularBuffer   // Latency history
    count      int64            // Total request count
    lastStatus int              // Most recent status code
}
```

**Circular Buffer**:
- Fixed-size ring buffer for memory efficiency
- Stores recent N latency measurements
- O(1) append and snapshot operations
- Automatically overwrites oldest data

### 3. Analyzer (`internal/analyzer`)

**Responsibility**: Detects anomalies and performance degradation by comparing metric windows.

**Key Features**:
- Periodic analysis (configurable tick interval)
- Drift detection algorithm
- Threshold-based alerting
- Window-based comparison

**Algorithm**:
1. Every tick interval (default: 5s):
   - Get snapshot of all endpoint metrics
   - For each endpoint with sufficient data:
     - Split data into "older" window (default: 80 samples)
     - And "recent" window (default: 20 samples)
     - Calculate average latency for each window
     - Compute percentage increase: `(recent - older) / older`
     - If increase >= threshold (default: 40%), emit alert

**Example**:
```
Older window (80 samples): 100ms average
Recent window (20 samples): 150ms average
Increase: (150 - 100) / 100 = 50% > 40% threshold → ALERT
```

### 4. Alert System (`internal/alert`)

**Responsibility**: Delivers alerts through various channels (sinks).

**Current Implementation**:
- Console sink (logs to stdout as JSON)
- Alert deduplication (prevents spam)

**Future Sinks**:
- Slack webhooks
- Email (SMTP)
- Generic webhooks
- PagerDuty
- Prometheus Alertmanager

**Alert Format**:
```json
{
  "type": "latency_drift",
  "method": "GET",
  "path": "/api/users",
  "percentage_increase": 50.0,
  "older_avg_ms": 100.0,
  "recent_avg_ms": 150.0,
  "timestamp": "2025-12-23T16:00:00Z"
}
```

## Request Flow Diagram

```
┌───────┐
│Client │
└───┬───┘
    │ 1. HTTP Request
    ▼
┌───────────────────┐
│  Proxy Handler    │
└────┬──────────────┘
     │ 2. Start Timer
     │ 3. Forward Request
     ▼
┌───────────────────┐
│     Backend       │
└────┬──────────────┘
     │ 4. Response
     ▼
┌───────────────────┐
│  Proxy Handler    │
└────┬──────────────┘
     │ 5. Calculate Latency
     │ 6. Record Metrics
     ▼
┌───────────────────┐
│ Metrics Registry  │
└───────────────────┘
     │
     │ 7. Return Response
     ▼
┌───────┐
│Client │
└───────┘

(Async) Every 5s:
┌───────────────────┐
│     Analyzer      │
└────┬──────────────┘
     │ 1. Get Snapshot
     ▼
┌───────────────────┐
│ Metrics Registry  │
└────┬──────────────┘
     │ 2. Analyze Windows
     ▼
┌───────────────────┐
│     Analyzer      │
└────┬──────────────┘
     │ 3. If drift detected
     ▼
┌───────────────────┐
│   Alert Sinks     │
└───────────────────┘
```

## Data Flow

### Metrics Collection
1. **Request arrives** at proxy handler
2. **Timer starts** before forwarding
3. **Request forwarded** to backend
4. **Response received** from backend
5. **Latency calculated**: `time.Since(start)`
6. **Metrics recorded**: `registry.Record(method, path, latency, status)`
7. **Data stored** in circular buffer for that endpoint
8. **Response returned** to client

### Drift Analysis
1. **Ticker fires** (every 5 seconds)
2. **Snapshot created** from all endpoint metrics
3. **For each endpoint**:
   - Check if sufficient samples (older + recent windows)
   - Split into older and recent windows
   - Calculate averages
   - Compute drift percentage
   - If drift > threshold, emit alert
4. **Alerts delivered** to configured sinks

## Concurrency Model

### Thread Safety
- **Metrics Registry**: Uses `sync.RWMutex` for concurrent reads/writes
- **Alert Sink**: Uses `sync.Mutex` for deduplication map
- **Proxy Handler**: Handles concurrent requests via Go's http server

### Goroutines
- **Main goroutine**: HTTP server
- **Analyzer goroutine**: Periodic drift detection
- **Per-request goroutines**: HTTP handler (managed by Go's http server)

## Configuration

### Environment Variables (Current)
- `RCIG_TARGET_URL`: Backend service URL (required)
- `RCIG_LISTEN_ADDR`: Listen address (default: `:8080`)

### Future Configuration
- YAML configuration files
- CLI flags
- Hot-reload support
- Per-endpoint settings

## Performance Characteristics

### Memory
- **Fixed overhead**: Circular buffers with configurable size
- **Per-endpoint**: ~1KB for 100 samples
- **Example**: 1000 endpoints × 100 samples = ~1MB

### CPU
- **Proxy overhead**: Minimal (~1-2% per request)
- **Analysis**: O(n) where n = number of endpoints
- **Ticker frequency**: Configurable (default: 5s)

### Latency
- **Proxy latency**: <1ms additional overhead
- **Does not block**: Analysis runs asynchronously

## Scalability

### Current Limits
- **Endpoints**: Thousands (limited by memory)
- **Requests/sec**: Tens of thousands (limited by backend)
- **History depth**: 100 samples per endpoint (configurable)

### Scaling Strategies
- **Horizontal**: Multiple RCIG instances behind load balancer
- **Vertical**: Increase memory for more history
- **Distributed**: External metrics store (Prometheus, InfluxDB)

## Security Considerations

### Current
- No authentication on admin endpoints
- No encryption in transit
- Trust all requests

### Future
- API key authentication
- TLS/HTTPS support
- Rate limiting
- Request sanitization
- IP whitelisting

## Extension Points

### Adding New Alert Sinks
1. Implement `alert.Sink` interface
2. Add configuration options
3. Register in main.go
4. Handle errors and retries

### Adding New Metrics
1. Extend `EndpointMetrics` struct
2. Update recording logic in proxy handler
3. Include in snapshots
4. Expose via HTTP endpoint

### Adding New Analysis
1. Create new analyzer type
2. Subscribe to metrics snapshots
3. Implement detection algorithm
4. Emit alerts on anomalies

## Monitoring RCIG Itself

### Health Check
- `GET /rcig/healthz`: Returns 200 OK if healthy

### Metrics Endpoint
- `GET /rcig/metrics`: JSON metrics for all endpoints

### Future Observability
- Prometheus metrics export
- OpenTelemetry tracing
- Structured logging
- pprof profiling endpoints

## Design Decisions

### Why Circular Buffer?
- **Fixed memory**: Prevents unbounded growth
- **Fast**: O(1) operations
- **Simple**: Easy to understand and debug

### Why In-Memory Metrics?
- **Fast**: No I/O overhead
- **Simple**: No external dependencies
- **Sufficient**: For real-time drift detection

### Why Window-Based Analysis?
- **Simple**: Easy to understand and configure
- **Effective**: Catches gradual degradation
- **Lightweight**: Low CPU overhead

### Why Separate Analyzer Goroutine?
- **Non-blocking**: Doesn't impact request handling
- **Periodic**: Configurable analysis frequency
- **Scalable**: Independent of request rate

## Future Architecture

### Planned Enhancements
1. **Configuration Management**: YAML files, hot-reload
2. **Enhanced Metrics**: Prometheus, OpenTelemetry
3. **Advanced Alerting**: Multiple sinks, rate limiting
4. **Authentication**: API keys, OAuth
5. **Advanced Proxy**: Rate limiting, circuit breakers, caching
6. **Distributed Tracing**: Request correlation across services
7. **High Availability**: Clustering, state synchronization

## References

- [Go net/http documentation](https://pkg.go.dev/net/http)
- [Reverse Proxy Pattern](https://en.wikipedia.org/wiki/Reverse_proxy)
- [Circuit Breaker Pattern](https://martinfowler.com/bliki/CircuitBreaker.html)
- [OpenTelemetry](https://opentelemetry.io/)
- [Prometheus](https://prometheus.io/)
