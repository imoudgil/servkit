package health

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLive(t *testing.T) {
	h := New(nil)
	mux := http.NewServeMux()
	h.Register(mux)

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}
}

func TestReadyWhenNotReady(t *testing.T) {
	h := New(nil)
	h.SetReady(false)
	rr := httptest.NewRecorder()
	h.Ready(rr, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d", rr.Code)
	}
}

func TestReadyCheckerFailure(t *testing.T) {
	h := New(func(r *http.Request) error {
		return errors.New("db down")
	})
	rr := httptest.NewRecorder()
	h.Ready(rr, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d", rr.Code)
	}
}

func TestReadySuccess(t *testing.T) {
	h := New(func(r *http.Request) error { return nil })
	rr := httptest.NewRecorder()
	h.Ready(rr, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}
}
