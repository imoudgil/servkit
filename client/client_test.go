package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/imoudgil/servkit/config"
)

func TestDoSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cfg := config.Service{ClientTimeout: time.Second, ClientRetries: 2}
	c := New(cfg)
	resp, err := c.Get(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
}

func TestDoRetriesOn500(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		if n <= 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cfg := config.Service{ClientTimeout: time.Second, ClientRetries: 3}
	c := New(cfg)
	resp, err := c.Get(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	if calls.Load() < 3 {
		t.Fatalf("expected retries, calls = %d", calls.Load())
	}
}

func TestDoRespectsContextCancel(t *testing.T) {
	block := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-block
	}))
	defer srv.Close()
	defer close(block)

	cfg := config.Service{ClientTimeout: time.Second, ClientRetries: 5}
	c := New(cfg)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := c.Get(ctx, srv.URL)
	if err == nil {
		t.Fatal("expected error on cancelled context")
	}
}

func TestBackoffCap(t *testing.T) {
	if got := backoff(10); got != 2*time.Second {
		t.Errorf("backoff(10) = %v", got)
	}
}
