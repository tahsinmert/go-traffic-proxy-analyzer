## Real Time Code Intelligence Gateway (RCIG)

RCIG is a lightweight reverse proxy written in Go that provides **real-time HTTP traffic metrics, anomaly detection, and alerting** for any upstream service.

It sits in front of your application, forwards requests to a target backend, and continuously analyzes responses to surface performance issues and error rate spikes as they happen.

---

## Features

- **Reverse proxy for any HTTP backend**
  - Forwards all incoming requests to a configured target URL.
  - Safe defaults with timeouts and graceful shutdown.

- **Real-time HTTP metrics**
  - Tracks request **count**, **status codes**, **latency (ms)** and **sample size** per endpoint.
  - Exposes a JSON metrics API at `GET /rcig/metrics`.

- **Health check endpoint**
  - Lightweight health probe at `GET /rcig/healthz` returning `200 OK` and body `ok`.

- **Drift / anomaly detection**
  - Compares **recent** vs **older** windows of metrics.
  - Detects significant changes in error rate or latency using a configurable drift threshold.

- **Alerting sink**
  - Pluggable alert sink (current implementation: console logging).
  - Designed to be extended with e.g. Slack, email, PagerDuty, or Prometheus Alertmanager integrations.

- **Small, focused codebase**
  - Idiomatic Go 1.22+.
  - Clear separation of concerns across `proxy`, `metrics`, `analyzer`, and `alert` packages.

---

## Architecture Overview

RCIG is structured as a small set of internal packages plus a single `cmd/proxy` entry point:

- `cmd/proxy/main.go`
  - Loads configuration from environment variables.
  - Creates the reverse proxy and HTTP server.
  - Wires together metrics registry, analyzer, and alert sink.
  - Exposes HTTP endpoints for metrics, health, and proxied traffic.

- `internal/proxy`
  - Contains the main HTTP handler used for proxying requests.
  - Wraps the standard library `httputil.ReverseProxy`.
  - Records per-request metrics into the registry.

- `internal/metrics`
  - Maintains in-memory statistics about HTTP traffic.
  - Produces **snapshots** and **summaries** used by the analyzer.
  - Exposes an HTTP handler (`NewHTTPHandler`) that returns a JSON payload:

    ```json
    {
      "generated_at": "2025-01-01T12:00:00Z",
      "endpoints": [
        {
          "method": "GET",
          "path": "/api/v1/users",
          "count": 1234,
          "last_status": 200,
          "average_ms": 15.2,
          "sample_size": 50,
          "last_updated_unix": 1735632000
        }
      ]
    }
    ```

- `internal/analyzer`
  - Periodically inspects metrics snapshots (on a configurable tick interval).
  - Compares **older** and **recent** windows (default: 80 vs 20 samples).
  - Emits alerts when drift exceeds a configured threshold (default: `0.40`).

- `internal/alert`
  - Defines the alert interface.
  - Provides a console implementation (`NewConsoleSink`) that prints alerts.

---

## Getting Started

### Prerequisites

- **Go**: version **1.22** or later.
- A reachable HTTP backend (e.g. your application running on `http://localhost:9000`).

### Installation

Clone the repository and fetch dependencies:

```bash
git clone https://github.com/example/rcig.git
cd rcig
go mod tidy
```

> Replace `https://github.com/example/rcig.git` with your actual repository URL if it differs.

---

## Configuration

RCIG is configured entirely via environment variables.

- **`RCIG_TARGET_URL`** (required)
  - The target HTTP backend to proxy traffic to.
  - Example: `http://localhost:9000`

- **`RCIG_LISTEN_ADDR`** (optional)
  - The address and port RCIG listens on.
  - Default: `:8080`
  - Examples:
    - `:8080` (all interfaces on port 8080)
    - `127.0.0.1:8080` (localhost only)

### Example

```bash
export RCIG_TARGET_URL=http://localhost:9000
export RCIG_LISTEN_ADDR=:8080

go run ./cmd/proxy
```

You should see a log message similar to:

```text
RCIG proxy :8080 adresinde dinliyor, hedef: http://localhost:9000
```

---

## Usage

With RCIG running in front of your backend:

- **Proxy all HTTP traffic**
  - Send your application traffic to RCIG instead of directly to the backend.
  - Example:

    ```bash
    curl -i http://localhost:8080/api/v1/users
    ```

- **Check health**

  ```bash
  curl -i http://localhost:8080/rcig/healthz
  # HTTP/1.1 200 OK
  # ok
  ```

- **Inspect metrics**

  ```bash
  curl -s http://localhost:8080/rcig/metrics | jq .
  ```

  The response includes aggregated metrics per `(method, path)` pair.

---

## Analyzer and Alerting

The analyzer runs in a separate goroutine and periodically evaluates metrics:

- Uses a configurable **tick interval** (default: every 5 seconds).
- Compares:
  - **Older window size** (default: 80 samples).
  - **Recent window size** (default: 20 samples).
- If the **drift** between windows exceeds the **drift threshold** (default: `0.40`),
  an alert is emitted via the configured alert sink.

The current implementation:

- Uses a **console alert sink** that logs alerts to standard output.
- Is designed so you can easily add new sinks (e.g. webhook, Slack, email).

---

## Running in Production

- **Logging**
  - RCIG uses the standard library `log` package.
  - Consider redirecting stdout/stderr to your log aggregation system.

- **Graceful shutdown**
  - Listens for `SIGINT` and `SIGTERM`.
  - Shuts down the HTTP server with a 10-second timeout.

- **Resource usage**
  - Metrics are held in memory with a bounded history (configured in `metrics.NewRegistry`).
  - Tune history size and analyzer windows according to your traffic profile.

---

## Development

### Project layout

- `cmd/proxy` – entrypoint for the RCIG proxy binary.
- `internal/metrics` – metrics registry and HTTP handler for `/rcig/metrics`.
- `internal/analyzer` – drift detection and alert generation.
- `internal/alert` – alert interfaces and implementations.
- `internal/proxy` – reverse proxy handler that records metrics.

### Running locally

```bash
export RCIG_TARGET_URL=http://localhost:9000
export RCIG_LISTEN_ADDR=:8080

go run ./cmd/proxy
```

### Running tests

If you add tests, you can run them with:

```bash
go test ./...
```

---

## Extending RCIG

Here are a few ideas for extending this project:

- **Custom alert sinks**
  - Implement new alert sinks (e.g. Slack, email, webhook, Prometheus).
  - Route analyzer alerts to multiple sinks.

- **Rich metrics storage**
  - Push metrics into Prometheus, OpenTelemetry, or a time-series database.
  - Build dashboards on top of these metrics.

- **Advanced anomaly detection**
  - Replace simple drift-based logic with more advanced statistical methods.
  - Add per-endpoint thresholds and dynamic baselines.

- **Authentication / authorization**
  - Protect `/rcig/*` endpoints with basic auth or tokens.

---

## Contributing

Contributions are welcome! Please see `CONTRIBUTING.md` for guidelines on how to:

- Propose new features or improvements.
- Report issues or bugs.
- Submit pull requests.

---

## Security

If you discover a security issue, **please do not open a public GitHub issue**.
Instead, follow the instructions in `SECURITY.md` to report the vulnerability responsibly.

---

## License

This project is licensed under the **MIT License**.  
See the `LICENSE` file for details.


