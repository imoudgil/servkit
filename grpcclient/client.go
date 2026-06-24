// Package grpcclient dials gRPC services with servkit defaults: bearer auth,
// tracing propagation, and configurable timeouts.
package grpcclient

import (
	"context"
	"fmt"
	"time"

	"github.com/imoudgil/servkit/config"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// Dial connects to addr using insecure transport suitable for local development
// and sidecar-style service meshes. Production callers should inject TLS creds.
func Dial(ctx context.Context, addr string, cfg config.Service, token string) (*grpc.ClientConn, error) {
	if addr == "" {
		return nil, fmt.Errorf("grpc address is required")
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	}

	var interceptors []grpc.UnaryClientInterceptor
	if token != "" {
		interceptors = append(interceptors, bearerTokenInterceptor(token))
	}
	if cfg.ClientTimeout > 0 {
		interceptors = append(interceptors, timeoutInterceptor(cfg.ClientTimeout))
	}
	if len(interceptors) > 0 {
		opts = append(opts, grpc.WithChainUnaryInterceptor(interceptors...))
	}

	conn, err := grpc.NewClient(addr, opts...)
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", addr, err)
	}
	return conn, nil
}

func bearerTokenInterceptor(token string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func timeoutInterceptor(timeout time.Duration) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
