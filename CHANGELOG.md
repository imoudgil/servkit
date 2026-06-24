# Changelog

All notable changes to this project are documented here.

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
