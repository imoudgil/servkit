# servkit

A small **Go shared library** for building consistent HTTP and gRPC microservices — production-ready building blocks platform teams ship so product teams move faster with fewer surprises.

Built as a portfolio project aligned with **service development** work: configuration, observability (logging, tracing, metrics), authentication, HTTP/gRPC servers and clients, and CI-tested packages other developers can import.

## Packages

| Package | Purpose |
|---------|---------|
| `config` | Environment-based service configuration with sane defaults |
| `logging` | JSON structured logging via `log/slog` and request-ID context helpers |
| `middleware` | Composable HTTP middleware: request ID, access logging, panic recovery |
| `health` | `/healthz` (liveness) and `/readyz` (readiness) handlers |
| `server` | HTTP server with auth, metrics, tracing, and graceful shutdown |
| `client` | HTTP client with timeout and retry/backoff on transient failures |
| `auth` | Bearer token authentication for HTTP handlers and gRPC unary RPCs |
| `metrics` | Prometheus counters/histograms and `/metrics` scrape endpoint |
| `tracing` | OpenTelemetry trace propagation for HTTP (W3C Trace Context) |
| `grpcserver` | gRPC server with auth, metrics, tracing, and graceful stop |
| `grpcclient` | gRPC client dial helper with bearer auth and timeouts |
| `timegrpc` | Sample `TimeService` gRPC implementation |

## Quick start (HTTP)

```bash
git clone https://github.com/imoudgil/servkit.git
cd servkit
go test ./...
go run ./examples/basic
```

In another terminal:

```bash
curl -i localhost:8080/healthz
curl -s localhost:8080/api/v1/time | jq
```

## Quick start (HTTP + gRPC + auth + metrics + tracing)

```bash
AUTH_TOKENS=dev-token METRICS_ENABLED=true TRACING_ENABLED=true go run ./examples/grpc
```

```bash
# HTTP (requires Bearer token on protected routes)
curl -s -H "Authorization: Bearer dev-token" localhost:8080/api/v1/time | jq
curl -s localhost:8080/metrics | head

# gRPC (requires grpcurl: brew install grpcurl)
grpcurl -H "authorization: Bearer dev-token" -plaintext localhost:9090 time.v1.TimeService/Now
```

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVICE_NAME` | `servkit` | Service name in logs and traces |
| `SERVICE_ENV` | `development` | Environment label |
| `HTTP_ADDR` | `:8080` | HTTP listen address |
| `GRPC_ADDR` | `:9090` | gRPC listen address |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |
| `SHUTDOWN_TIMEOUT` | `10s` | Graceful shutdown window |
| `CLIENT_TIMEOUT` | `5s` | Outbound HTTP/gRPC timeout |
| `CLIENT_RETRIES` | `3` | HTTP retries on 5xx / network errors |
| `AUTH_TOKENS` | _(empty)_ | Comma-separated Bearer tokens; enables auth when set |
| `METRICS_ENABLED` | `false` | Expose Prometheus metrics at `GET /metrics` |
| `TRACING_ENABLED` | `false` | Enable OpenTelemetry HTTP trace propagation |

## Example: register your own HTTP routes

```go
cfg, _ := config.Load()
srv := server.New(cfg)

srv.Mux().HandleFunc("GET /api/v1/hello", func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("hello"))
})

ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer stop()
_ = srv.ListenAndServe(ctx)
```

Protected routes automatically skip auth for `/healthz`, `/readyz`, and `/metrics`.

## Testing

```bash
go test ./... -count=1 -cover
go test ./... -race -count=1
```

CI runs the same suite on every push (see `.github/workflows/ci.yml`).

## Design notes

- **Inner-source mindset:** small, documented packages with runnable examples and changelog entries.
- **Observability:** structured logging, Prometheus metrics, OpenTelemetry tracing, request correlation IDs.
- **Security:** Bearer token auth on HTTP and gRPC; health/metrics endpoints remain unauthenticated.
- **Operability:** graceful shutdown, health/readiness endpoints, resilient outbound HTTP client.
- **Quality bar:** table-driven unit tests, `httptest` / `bufconn` integration tests, GitHub Actions CI with race detector.

## License

MIT — see [LICENSE](LICENSE).

## Author

[Ishan Moudgil](https://github.com/imoudgil)
