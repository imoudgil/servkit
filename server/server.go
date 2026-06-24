// Package server wraps net/http with servkit defaults: structured logging middleware,
// graceful shutdown, and health endpoints.
package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/imoudgil/servkit/auth"
	"github.com/imoudgil/servkit/config"
	"github.com/imoudgil/servkit/health"
	"github.com/imoudgil/servkit/logging"
	"github.com/imoudgil/servkit/metrics"
	"github.com/imoudgil/servkit/middleware"
	"github.com/imoudgil/servkit/tracing"
)

// Server is a production-minded HTTP server with shared middleware and shutdown.
type Server struct {
	cfg     config.Service
	logger  *slog.Logger
	mux     *http.ServeMux
	http    *http.Server
	health  *health.Handler
	metrics *metrics.Registry
	auth    *auth.Validator
}

// New constructs a Server from configuration and registers default health routes.
func New(cfg config.Service) *Server {
	logger := logging.New(cfg.Name, cfg.LogLevel)
	mux := http.NewServeMux()
	h := health.New(nil)
	h.Register(mux)

	validator := auth.New(cfg.AuthTokens...)
	var reg *metrics.Registry
	if cfg.MetricsEnabled {
		reg = metrics.NewRegistry()
		mux.Handle("GET /metrics", reg.Handler())
	}

	s := &Server{
		cfg:     cfg,
		logger:  logger,
		mux:     mux,
		health:  h,
		metrics: reg,
		auth:    validator,
	}
	s.http = &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: s.handler(),
	}
	return s
}

// Mux returns the underlying ServeMux for registering application routes.
func (s *Server) Mux() *http.ServeMux {
	return s.mux
}

// Health returns the health handler for readiness toggling in tests or startup.
func (s *Server) Health() *health.Handler {
	return s.health
}

// Metrics returns the Prometheus registry when metrics are enabled.
func (s *Server) Metrics() *metrics.Registry {
	return s.metrics
}

// Logger returns the service-scoped structured logger.
func (s *Server) Logger() *slog.Logger {
	return s.logger
}

func (s *Server) handler() http.Handler {
	mws := []middleware.Middleware{
		middleware.RequestID,
	}
	if s.cfg.TracingEnabled {
		mws = append(mws, tracing.Middleware("http"))
	}
	if s.metrics != nil {
		mws = append(mws, s.metrics.Middleware())
	}
	mws = append(mws,
		middleware.Logging(s.logger),
		middleware.Recovery(s.logger),
	)
	if s.auth.Enabled() {
		mws = append(mws, func(next http.Handler) http.Handler {
			return s.auth.HTTP(next)
		})
	}
	return middleware.Chain(s.mux, mws...)
}

// ListenAndServe starts the HTTP server and blocks until ctx is cancelled, then
// shuts down gracefully within cfg.ShutdownTimeout.
func (s *Server) ListenAndServe(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		s.logger.Info("server_start", "addr", s.cfg.HTTPAddr, "env", s.cfg.Environment)
		if err := s.http.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
		defer cancel()
		s.logger.Info("server_shutdown", "timeout", s.cfg.ShutdownTimeout.String())
		if err := s.http.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown: %w", err)
		}
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}

// Addr returns the configured listen address.
func (s *Server) Addr() string {
	return s.cfg.HTTPAddr
}

// ShutdownTimeout exposes the configured graceful shutdown window.
func (s *Server) ShutdownTimeout() time.Duration {
	return s.cfg.ShutdownTimeout
}
