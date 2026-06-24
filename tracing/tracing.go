// Package tracing configures OpenTelemetry trace propagation for HTTP and gRPC.
package tracing

import (
	"context"
	"fmt"
	"net/http"

	"github.com/imoudgil/servkit/middleware"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

// Init installs a tracer provider with a stdout exporter suitable for local
// development. Call shutdown when the process exits.
func Init(ctx context.Context, serviceName string) (func(context.Context) error, error) {
	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, fmt.Errorf("stdout trace exporter: %w", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(semconv.ServiceName(serviceName)),
	)
	if err != nil {
		return nil, fmt.Errorf("trace resource: %w", err)
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return provider.Shutdown, nil
}

// HTTP wraps handlers with otelhttp instrumentation and W3C trace propagation.
func HTTP(operation string, next http.Handler) http.Handler {
	return otelhttp.NewHandler(next, operation)
}

// Middleware adapts tracing to the servkit middleware type.
func Middleware(operation string) middleware.Middleware {
	return func(next http.Handler) http.Handler {
		return HTTP(operation, next)
	}
}
