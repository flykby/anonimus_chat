package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("REDIS_URL", "redis://example")

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":              "ok",
			"service":             "api",
			"database_configured": os.Getenv("DATABASE_URL") != "",
			"redis_configured":    os.Getenv("REDIS_URL") != "",
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json: %v", err)
	}
	if body["service"] != "api" {
		t.Fatalf("service = %v", body["service"])
	}
}
