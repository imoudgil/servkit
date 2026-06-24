// Basic demonstrates wiring servkit's config, server, client, and health packages
// into a small HTTP service other teams could extend.
//
// Run:
//
//	go run ./examples/basic
package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/imoudgil/servkit/client"
	"github.com/imoudgil/servkit/config"
	"github.com/imoudgil/servkit/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	srv := server.New(cfg)
	srv.Mux().HandleFunc("GET /api/v1/time", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"unix":    time.Now().Unix(),
			"service": cfg.Name,
		})
	})

	srv.Mux().HandleFunc("GET /api/v1/upstream", func(w http.ResponseWriter, r *http.Request) {
		httpClient := client.New(cfg)
		resp, err := httpClient.Get(r.Context(), "https://httpbin.org/get")
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		defer resp.Body.Close()
		writeJSON(w, http.StatusOK, map[string]any{
			"upstream_status": resp.StatusCode,
		})
	})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Printf("listening on %s (env=%s)", cfg.HTTPAddr, cfg.Environment)
	if err := srv.ListenAndServe(ctx); err != nil && err != context.Canceled {
		log.Fatalf("server: %v", err)
	}
}

func writeJSON(w http.ResponseWriter, code int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(body)
}
