package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestHTTPAllowsValidToken(t *testing.T) {
	v := New("secret-token")
	called := false
	h := v.HTTP(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if tok, ok := TokenFromContext(r.Context()); !ok || tok != "secret-token" {
			t.Fatalf("token in context = %q, ok=%v", tok, ok)
		}
		w.WriteHeader(http.StatusOK)
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/time", nil)
	req.Header.Set("Authorization", "Bearer secret-token")
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}
	if !called {
		t.Fatal("handler not called")
	}
}

func TestHTTPRejectsMissingToken(t *testing.T) {
	v := New("secret-token")
	h := v.HTTP(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not run")
	}))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/v1/time", nil))
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", rr.Code)
	}
}

func TestHTTPSkipsHealth(t *testing.T) {
	v := New("secret-token")
	h := v.HTTP(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}
}

func TestGRPCUnaryInterceptor(t *testing.T) {
	v := New("grpc-secret")
	ic := v.UnaryServerInterceptor()
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer grpc-secret"))

	_, err := ic(ctx, nil, &grpc.UnaryServerInfo{}, func(ctx context.Context, req any) (any, error) {
		if tok, ok := TokenFromContext(ctx); !ok || tok != "grpc-secret" {
			t.Fatalf("token = %q", tok)
		}
		return "ok", nil
	})
	if err != nil {
		t.Fatalf("interceptor error = %v", err)
	}
}

func TestGRPCUnaryRejectsInvalid(t *testing.T) {
	v := New("grpc-secret")
	ic := v.UnaryServerInterceptor()
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer wrong"))

	_, err := ic(ctx, nil, &grpc.UnaryServerInfo{}, func(ctx context.Context, req any) (any, error) {
		t.Fatal("handler should not run")
		return nil, nil
	})
	if err == nil {
		t.Fatal("expected error")
	}
}
