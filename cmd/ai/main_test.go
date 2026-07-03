package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	t.Setenv("RUNPOD_LLM_URL", "https://runpod.example/llm")
	t.Setenv("RUNPOD_EMBEDDING_URL", "https://runpod.example/embed")

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":                      "ok",
			"service":                     "ai",
			"runpod_llm_configured":       os.Getenv("RUNPOD_LLM_URL") != "",
			"runpod_embedding_configured": os.Getenv("RUNPOD_EMBEDDING_URL") != "",
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}
