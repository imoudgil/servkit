# Changelog

All notable changes to this project are documented here.

## [0.2.0] - 2026-06-24

### Added
- `auth` package with Bearer token middleware for HTTP and gRPC unary RPCs
- `metrics` package with Prometheus request counters/histograms and `/metrics` handler
- `tracing` package with OpenTelemetry HTTP instrumentation and stdout trace export
- `grpcserver` and `grpcclient` packages for gRPC servers/clients with auth, metrics, and tracing hooks
- `timegrpc` service implementation and `proto/time/v1` gRPC API definition
- `examples/grpc` demonstrating HTTP + gRPC with auth, metrics, and tracing enabled
- Config env vars: `GRPC_ADDR`, `AUTH_TOKENS`, `METRICS_ENABLED`, `TRACING_ENABLED`

## [0.1.0] - 2026-06-24

### Added
- `config` package for environment-based service configuration
- `logging` package with JSON `slog` helpers and request-ID context
- `middleware` package: RequestID, Logging, Recovery, and Chain compositor
- `health` package with `/healthz` and `/readyz` endpoints
- `server` package with default middleware stack and graceful shutdown
- `client` package with timeout and exponential backoff retry on 5xx
- Runnable example at `examples/basic`
- GitHub Actions CI workflow running `go test ./...`
