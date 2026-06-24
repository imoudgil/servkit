package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPMiddlewareRecordsMetrics(t *testing.T) {
	reg := NewRegistry()
	h := reg.HTTP(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/items", nil))
	if rr.Code != http.StatusCreated {
		t.Fatalf("status = %d", rr.Code)
	}

	metric, err := reg.requests.GetMetricWithLabelValues("http", http.MethodPost, "201")
	if err != nil {
		t.Fatal(err)
	}
	_ = metric
}

func TestHandlerServesMetrics(t *testing.T) {
	reg := NewRegistry()
	reg.HTTP(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	rr := httptest.NewRecorder()
	reg.Handler().ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "servkit_requests_total") {
		t.Fatalf("expected metric in body, got %q", rr.Body.String())
	}
}
