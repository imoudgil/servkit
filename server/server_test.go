package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/imoudgil/servkit/config"
)

func TestServerHealthRoutes(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	cfg.HTTPAddr = ":0"

	s := New(cfg)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	s.handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("/healthz status = %d", rr.Code)
	}
}

func TestServerMiddlewareRequestID(t *testing.T) {
	cfg, _ := config.Load()
	s := New(cfg)
	s.Mux().HandleFunc("GET /ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rr := httptest.NewRecorder()
	s.handler().ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/ping", nil))
	if rr.Header().Get("X-Request-ID") == "" {
		t.Fatal("expected X-Request-ID header")
	}
}

func TestListenAndServeShutdown(t *testing.T) {
	t.Setenv("HTTP_ADDR", "127.0.0.1:0")
	t.Setenv("SHUTDOWN_TIMEOUT", "2s")
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	s := New(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- s.ListenAndServe(ctx) }()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err != nil && err != context.Canceled {
			t.Fatalf("ListenAndServe() = %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("server did not shut down in time")
	}
}
