// Full-stack example: HTTP + gRPC servers with auth, metrics, and tracing.
//
// Run:
//
//	AUTH_TOKENS=dev-token METRICS_ENABLED=true TRACING_ENABLED=true go run ./examples/grpc
package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/imoudgil/servkit/auth"
	"github.com/imoudgil/servkit/config"
	"github.com/imoudgil/servkit/grpcserver"
	"github.com/imoudgil/servkit/metrics"
	timev1 "github.com/imoudgil/servkit/proto/time/v1"
	"github.com/imoudgil/servkit/server"
	"github.com/imoudgil/servkit/timegrpc"
	"github.com/imoudgil/servkit/tracing"
	"google.golang.org/grpc"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	if len(cfg.AuthTokens) == 0 {
		cfg.AuthTokens = []string{"dev-token"}
	}
	cfg.MetricsEnabled = true
	cfg.TracingEnabled = true

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	traceShutdown, err := tracing.Init(ctx, cfg.Name)
	if err != nil {
		log.Fatalf("tracing: %v", err)
	}
	defer func() { _ = traceShutdown(context.Background()) }()

	httpSrv := server.New(cfg)

	var prom *metrics.Registry
	if cfg.MetricsEnabled {
		prom = httpSrv.Metrics()
	}

	httpSrv.Mux().HandleFunc("GET /api/v1/time", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"unix":    time.Now().Unix(),
			"service": cfg.Name,
		})
	})

	grpcOpts := []grpcserver.Option{grpcserver.WithAuth(auth.New(cfg.AuthTokens...))}
	if prom != nil {
		grpcOpts = append(grpcOpts, grpcserver.WithMetrics(prom))
	}
	grpcSrv, err := grpcserver.New(cfg, func(gs *grpc.Server) {
		timev1.RegisterTimeServiceServer(gs, &timegrpc.Server{ServiceName: cfg.Name})
	}, grpcOpts...)
	if err != nil {
		log.Fatalf("grpc server: %v", err)
	}

	errCh := make(chan error, 2)
	go func() { errCh <- httpSrv.ListenAndServe(ctx) }()
	go func() { errCh <- grpcSrv.ListenAndServe(ctx) }()

	log.Printf("http=%s grpc=%s auth=Bearer %q metrics=/metrics tracing=stdout",
		cfg.HTTPAddr, cfg.GRPCAddr, cfg.AuthTokens[0])

	select {
	case <-ctx.Done():
	case err := <-errCh:
		if err != nil && err != context.Canceled {
			log.Fatalf("server: %v", err)
		}
	}
}

func writeJSON(w http.ResponseWriter, code int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(body)
}
