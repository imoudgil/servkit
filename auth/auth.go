// Package auth provides Bearer token authentication for HTTP handlers and gRPC
// unary RPCs, with configurable path skips for health and metrics endpoints.
package auth

import (
	"context"
	"net/http"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ctxKey struct{}

// Validator checks Authorization: Bearer <token> against an allowlist.
type Validator struct {
	tokens map[string]struct{}
	skip   map[string]struct{}
}

// New builds a Validator from one or more bearer tokens. Empty tokens are ignored.
func New(tokens ...string) *Validator {
	allowed := make(map[string]struct{})
	for _, t := range tokens {
		t = strings.TrimSpace(t)
		if t != "" {
			allowed[t] = struct{}{}
		}
	}
	return &Validator{
		tokens: allowed,
		skip: map[string]struct{}{
			"/healthz": {},
			"/readyz":  {},
			"/metrics": {},
		},
	}
}

// Enabled reports whether any tokens were configured.
func (v *Validator) Enabled() bool {
	return len(v.tokens) > 0
}

// HTTP returns middleware that rejects unauthenticated requests.
func (v *Validator) HTTP(next http.Handler) http.Handler {
	if !v.Enabled() {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if v.skipPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}
		token, ok := bearerToken(r.Header.Get("Authorization"))
		if !ok || !v.valid(token) {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), ctxKey{}, token)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// UnaryServerInterceptor enforces bearer auth on gRPC unary calls.
func (v *Validator) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	if !v.Enabled() {
		return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
			return handler(ctx, req)
		}
	}
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		token, ok := tokenFromGRPC(ctx)
		if !ok || !v.valid(token) {
			return nil, status.Error(codes.Unauthenticated, "missing or invalid bearer token")
		}
		ctx = context.WithValue(ctx, ctxKey{}, token)
		return handler(ctx, req)
	}
}

// TokenFromContext returns the validated bearer token stored in ctx, if any.
func TokenFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(ctxKey{}).(string)
	return v, ok && v != ""
}

func (v *Validator) skipPath(path string) bool {
	_, ok := v.skip[path]
	return ok
}

func (v *Validator) valid(token string) bool {
	_, ok := v.tokens[token]
	return ok
}

func bearerToken(header string) (string, bool) {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", false
	}
	token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	if token == "" {
		return "", false
	}
	return token, true
}

func tokenFromGRPC(ctx context.Context) (string, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", false
	}
	vals := md.Get("authorization")
	if len(vals) == 0 {
		return "", false
	}
	return bearerToken(vals[0])
}
