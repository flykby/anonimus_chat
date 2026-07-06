package users

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flykby/anonimus_chat/internal/profile"
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

func TestGetProfileViewInvalidID(t *testing.T) {
	t.Parallel()

	h := &Handler{Users: nil}
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/users/by-telegram/abc/profile", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestGetProfileViewMeMissingTelegramID(t *testing.T) {
	t.Parallel()

	h := &Handler{Users: nil}
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/users/me/profile", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestGetLanguageInvalidID(t *testing.T) {
	t.Parallel()

	h := &Handler{Users: nil}
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/users/by-telegram/abc/language", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestPatchProfileMissingTelegramID(t *testing.T) {
	t.Parallel()

	h := &Handler{Users: nil}
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	body, _ := json.Marshal(patchProfileRequest{Age: ptrInt16(25)})
	req := httptest.NewRequest(http.MethodPatch, "/users/me/profile", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestPatchProfileBadAge(t *testing.T) {
	t.Parallel()

	h := &Handler{Users: nil}
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	body, _ := json.Marshal(patchProfileRequest{
		TelegramID: 1,
		Age:        ptrInt16(17),
	})
	req := httptest.NewRequest(http.MethodPatch, "/users/me/profile", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestPatchProfileNoFields(t *testing.T) {
	t.Parallel()

	h := &Handler{Users: nil}
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	body, _ := json.Marshal(patchProfileRequest{TelegramID: 1})
	req := httptest.NewRequest(http.MethodPatch, "/users/me/profile", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestPatchProfileInvalidLanguage(t *testing.T) {
	t.Parallel()

	h := &Handler{Users: nil}
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	lang := "de"
	body, _ := json.Marshal(patchProfileRequest{
		TelegramID: 1,
		Language:   &lang,
	})
	req := httptest.NewRequest(http.MethodPatch, "/users/me/profile", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestDeleteMeMissingTelegramID(t *testing.T) {
	t.Parallel()

	h := &Handler{Users: nil, Delete: &profile.DeleteService{}}
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	body, _ := json.Marshal(deleteMeRequest{})
	req := httptest.NewRequest(http.MethodDelete, "/users/me", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func ptrInt16(v int16) *int16 {
	return &v
}
