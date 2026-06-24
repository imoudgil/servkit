package tracing

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.opentelemetry.io/otel"
)

func TestHTTPMiddlewareCreatesSpan(t *testing.T) {
	shutdown, err := Init(context.Background(), "servkit-test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = shutdown(context.Background()) })

	tracer := otel.Tracer("test")
	var spanName string
	h := HTTP("test_op", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, span := tracer.Start(r.Context(), "inner")
		spanName = span.SpanContext().TraceID().String()
		span.End()
		w.WriteHeader(http.StatusOK)
	}))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/ping", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}
	if spanName == "" {
		t.Fatal("expected trace id from nested span")
	}
}

func TestInitShutdown(t *testing.T) {
	shutdown, err := Init(context.Background(), "servkit-test")
	if err != nil {
		t.Fatal(err)
	}
	if err := shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown: %v", err)
	}
}
