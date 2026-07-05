package users

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegisterBadAge(t *testing.T) {
	t.Parallel()

	h := &Handler{Users: nil}
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	body, _ := json.Marshal(registerRequest{
		TelegramID: 1,
		Age:        17,
		Gender:     "male",
		Seeking:    "female",
		Language:   "ru",
	})
	req := httptest.NewRequest(http.MethodPost, "/users/register", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestGetByTelegramInvalidID(t *testing.T) {
	t.Parallel()

	h := &Handler{Users: nil}
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/users/by-telegram/abc", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}
