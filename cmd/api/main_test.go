package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestHealthHandlerWithoutDB(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler(nil))

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
	if body["database_ok"] != false {
		t.Fatalf("database_ok = %v, want false", body["database_ok"])
	}
}

func TestHealthHandlerConfiguredFlag(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("REDIS_URL", "redis://example")

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler(nil))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json: %v", err)
	}
	if body["database_configured"] != true {
		t.Fatalf("database_configured = %v", body["database_configured"])
	}
	if os.Getenv("DATABASE_URL") == "" {
		t.Fatal("expected DATABASE_URL in env")
	}
}
