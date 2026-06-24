// Package grpcserver wraps google.golang.org/grpc with servkit defaults:
// structured logging hooks, auth, metrics, tracing, and graceful shutdown.
package grpcserver

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/imoudgil/servkit/auth"
	"github.com/imoudgil/servkit/config"
	"github.com/imoudgil/servkit/logging"
	"github.com/imoudgil/servkit/metrics"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

// Server is a production-minded gRPC server with shared interceptors.
type Server struct {
	cfg    config.Service
	logger *slog.Logger
	srv    *grpc.Server
	lis    net.Listener
}

// RegisterFunc attaches service implementations to a grpc.Server.
type RegisterFunc func(*grpc.Server)

// New constructs a gRPC server and listener from configuration.
func New(cfg config.Service, register RegisterFunc, opts ...Option) (*Server, error) {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	logger := logging.New(cfg.Name, cfg.LogLevel)
	validator := o.auth
	if validator == nil {
		validator = auth.New(cfg.AuthTokens...)
	}

	var unary []grpc.UnaryServerInterceptor
	if validator.Enabled() {
		unary = append(unary, validator.UnaryServerInterceptor())
	}
	if o.metrics != nil {
		unary = append(unary, o.metrics.UnaryServerInterceptor())
	}
	unary = append(unary, loggingUnary(logger))

	grpcOpts := []grpc.ServerOption{
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	}
	if len(unary) > 0 {
		grpcOpts = append(grpcOpts, grpc.ChainUnaryInterceptor(unary...))
	}

	srv := grpc.NewServer(grpcOpts...)
	if register != nil {
		register(srv)
	}

	lis, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		return nil, fmt.Errorf("listen %s: %w", cfg.GRPCAddr, err)
	}

	return &Server{
		cfg:    cfg,
		logger: logger,
		srv:    srv,
		lis:    lis,
	}, nil
}

// Addr returns the configured listen address.
func (s *Server) Addr() string {
	return s.cfg.GRPCAddr
}

// GRPC returns the underlying grpc.Server for advanced registration in tests.
func (s *Server) GRPC() *grpc.Server {
	return s.srv
}

// ListenAndServe blocks until ctx is cancelled, then performs a graceful stop.
func (s *Server) ListenAndServe(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		s.logger.Info("grpc_server_start", "addr", s.lis.Addr().String(), "env", s.cfg.Environment)
		if err := s.srv.Serve(s.lis); err != nil {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		s.logger.Info("grpc_server_stop")
		stopped := make(chan struct{})
		go func() {
			s.srv.GracefulStop()
			close(stopped)
		}()
		select {
		case <-stopped:
			return ctx.Err()
		case <-time.After(s.cfg.ShutdownTimeout):
			s.srv.Stop()
			return ctx.Err()
		}
	case err := <-errCh:
		return err
	}
}

// Option configures optional grpcserver dependencies.
type Option func(*options)

type options struct {
	auth    *auth.Validator
	metrics *metrics.Registry
}

func defaultOptions() *options {
	return &options{}
}

// WithAuth sets a custom auth validator.
func WithAuth(v *auth.Validator) Option {
	return func(o *options) { o.auth = v }
}

// WithMetrics enables Prometheus RPC metrics.
func WithMetrics(reg *metrics.Registry) Option {
	return func(o *options) { o.metrics = reg }
}

func loggingUnary(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := logging.RequestIDFromContext(ctx)
		logger.Info("grpc_request", "method", info.FullMethod, "request_id", start)
		resp, err := handler(ctx, req)
		if err != nil {
			logger.Error("grpc_error", "method", info.FullMethod, "error", err.Error())
		}
		return resp, err
	}
}
