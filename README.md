# servkit

A small **Go shared library** for building consistent HTTP microservices — the kind of production-ready building blocks platform teams ship so product teams move faster with fewer surprises.

Built as a learning/portfolio project aligned with internal **service development** work: configuration, structured logging, HTTP servers/clients, health checks, middleware, and CI-tested packages other developers can import.

## Packages

| Package | Purpose |
|---------|---------|
| `config` | Environment-based service configuration with sane defaults |
| `logging` | JSON structured logging via `log/slog` and request-ID context helpers |
| `middleware` | Composable HTTP middleware: request ID, access logging, panic recovery |
| `health` | `/healthz` (liveness) and `/readyz` (readiness) handlers |
| `server` | HTTP server wrapper with default middleware and graceful shutdown |
| `client` | HTTP client with timeout and retry/backoff on transient failures |

## Quick start

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

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVICE_NAME` | `servkit` | Service name in logs |
| `SERVICE_ENV` | `development` | Environment label |
| `HTTP_ADDR` | `:8080` | Listen address |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |
| `SHUTDOWN_TIMEOUT` | `10s` | Graceful shutdown window |
| `CLIENT_TIMEOUT` | `5s` | Outbound HTTP timeout |
| `CLIENT_RETRIES` | `3` | Retries on 5xx / network errors |

## Example: register your own routes

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

Every route automatically gets:

- `X-Request-ID` propagation/generation
- Structured JSON access logs (`method`, `path`, `status`, `duration_ms`)
- Panic recovery with error logging

## Testing

```bash
go test ./... -count=1 -cover
go test ./... -bench=. -benchmem ./middleware
```

CI runs the same suite on every push (see `.github/workflows/ci.yml`).

## Design notes

- **Inner-source mindset:** small, documented packages with runnable examples and changelog entries — intended to be imported by other services.
- **Observability:** structured logging and request correlation IDs; access logs suitable for log aggregation.
- **Operability:** graceful shutdown, health/readiness endpoints, resilient outbound HTTP client.
- **Quality bar:** table-driven unit tests, `httptest` integration tests, GitHub Actions CI.

## License

MIT — see [LICENSE](LICENSE).

## Author

[Ishan Moudgil](https://github.com/imoudgil)
