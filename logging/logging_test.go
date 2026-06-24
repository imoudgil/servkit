package logging

import (
	"log/slog"
	"testing"
)

func TestParseLevel(t *testing.T) {
	tests := []struct {
		in   string
		want slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"WARN", slog.LevelWarn},
		{"error", slog.LevelError},
		{"info", slog.LevelInfo},
		{"unknown", slog.LevelInfo},
	}
	for _, tc := range tests {
		if got := parseLevel(tc.in); got != tc.want {
			t.Errorf("parseLevel(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestRequestIDContext(t *testing.T) {
	ctx := ContextWithRequestID(t.Context(), "abc-123")
	if got := RequestIDFromContext(ctx); got != "abc-123" {
		t.Errorf("RequestIDFromContext() = %q", got)
	}
}

func TestWithRequestID(t *testing.T) {
	base := New("test", "info")
	child := WithRequestID(base, "req-1")
	if child == nil {
		t.Fatal("WithRequestID returned nil")
	}
}
