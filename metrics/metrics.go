// Package metrics exposes Prometheus counters and histograms for HTTP and gRPC
// services, plus a standard /metrics scrape handler.
package metrics

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/imoudgil/servkit/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// Registry holds Prometheus collectors for request volume and latency.
type Registry struct {
	prom     *prometheus.Registry
	requests *prometheus.CounterVec
	duration *prometheus.HistogramVec
}

// NewRegistry registers HTTP and gRPC metric collectors on a dedicated registry.
func NewRegistry() *Registry {
	reg := prometheus.NewRegistry()
	r := &Registry{
		prom: reg,
		requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "servkit_requests_total",
			Help: "Total number of handled HTTP and gRPC requests.",
		}, []string{"protocol", "method", "code"}),
		duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "servkit_request_duration_seconds",
			Help:    "Request latency in seconds.",
			Buckets: prometheus.DefBuckets,
		}, []string{"protocol", "method"}),
	}
	reg.MustRegister(r.requests, r.duration)
	return r
}

// Handler serves Prometheus metrics at /metrics.
func (r *Registry) Handler() http.Handler {
	return promhttp.HandlerFor(r.prom, promhttp.HandlerOpts{})
}

// HTTP returns middleware that records request count and duration.
func (r *Registry) HTTP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, req)
		code := strconv.Itoa(rw.status)
		r.requests.WithLabelValues("http", req.Method, code).Inc()
		r.duration.WithLabelValues("http", req.Method).Observe(time.Since(start).Seconds())
	})
}

// Middleware adapts HTTP metrics to the servkit middleware type.
func (r *Registry) Middleware() middleware.Middleware {
	return func(next http.Handler) http.Handler {
		return r.HTTP(next)
	}
}

// UnaryServerInterceptor records gRPC unary RPC metrics.
func (r *Registry) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		code := status.Code(err).String()
		r.requests.WithLabelValues("grpc", info.FullMethod, code).Inc()
		r.duration.WithLabelValues("grpc", info.FullMethod).Observe(time.Since(start).Seconds())
		return resp, err
	}
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (w *responseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}
