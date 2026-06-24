package middleware

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestChainOrder(t *testing.T) {
	var order []string
	mk := func(name string) Middleware {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, name+"_before")
				next.ServeHTTP(w, r)
				order = append(order, name+"_after")
			})
		}
	}
	h := Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
	}), mk("outer"), mk("inner"))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

	want := []string{"outer_before", "inner_before", "handler", "inner_after", "outer_after"}
	if len(order) != len(want) {
		t.Fatalf("order = %v, want %v", order, want)
	}
	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("order[%d] = %q, want %q; full=%v", i, order[i], want[i], order)
		}
	}
}

func TestRequestIDGenerated(t *testing.T) {
	h := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/health", nil))
	id := rr.Header().Get("X-Request-ID")
	if id == "" {
		t.Fatal("missing X-Request-ID")
	}
	if len(id) < 8 {
		t.Errorf("request id too short: %q", id)
	}
}

func TestRequestIDPropagated(t *testing.T) {
	h := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", "client-supplied-id")
	h.ServeHTTP(rr, req)
	if got := rr.Header().Get("X-Request-ID"); got != "client-supplied-id" {
		t.Errorf("X-Request-ID = %q", got)
	}
}

func TestRecovery(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := Recovery(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	}))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status = %d", rr.Code)
	}
}

func TestLoggingMiddleware(t *testing.T) {
	var buf strings.Builder
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
	h := Chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte("ok"))
		}),
		RequestID,
		Logging(logger),
	)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/items", nil))
	if rr.Code != http.StatusCreated {
		t.Errorf("status = %d", rr.Code)
	}
	if !strings.Contains(buf.String(), "http_request") {
		t.Errorf("expected access log, got %q", buf.String())
	}
}
