package users

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/flykby/anonimus_chat/internal/db"
	"github.com/flykby/anonimus_chat/internal/shared"
)

type Handler struct {
	Users *db.UsersRepo
}

type registerRequest struct {
	TelegramID int64  `json:"telegram_id"`
	Age        int16  `json:"age"`
	Gender     string `json:"gender"`
	Seeking    string `json:"seeking"`
	Language   string `json:"language"`
}

type profileResponse struct {
	TelegramID   int64  `json:"telegram_id"`
	Age          int16  `json:"age"`
	Gender       string `json:"gender"`
	Seeking      string `json:"seeking"`
	Language     string `json:"language"`
	ActiveDialog bool   `json:"active_dialog"`
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /users/by-telegram/{telegram_id}", h.getByTelegram)
	mux.HandleFunc("POST /users/register", h.register)
}

func (h *Handler) getByTelegram(w http.ResponseWriter, r *http.Request) {
	telegramID, err := strconv.ParseInt(r.PathValue("telegram_id"), 10, 64)
	if err != nil || telegramID <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid telegram_id"})
		return
	}

	up, ok, err := h.Users.GetByTelegramID(r.Context(), telegramID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}

	activeDialog, err := h.Users.HasActiveDialog(r.Context(), up.User.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	writeJSON(w, http.StatusOK, toProfileResponse(up, activeDialog))
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	if req.TelegramID <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "telegram_id required"})
		return
	}
	if req.Age < 18 || req.Age > 99 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "age must be 18-99"})
		return
	}

	gender, err := parseGender(req.Gender)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid gender"})
		return
	}
	seeking, err := parseGender(req.Seeking)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid seeking"})
		return
	}
	language, err := parseLanguage(req.Language)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid language"})
		return
	}

	up, err := h.Users.Register(r.Context(), db.RegisterInput{
		TelegramID: req.TelegramID,
		Age:        req.Age,
		Gender:     gender,
		Seeking:    seeking,
		Language:   language,
	})
	if errors.Is(err, db.ErrUserAlreadyRegistered) {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "already registered"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	writeJSON(w, http.StatusCreated, toProfileResponse(up, false))
}

func toProfileResponse(up db.UserProfile, activeDialog bool) profileResponse {
	return profileResponse{
		TelegramID:   up.User.TelegramID,
		Age:          up.Profile.Age,
		Gender:       string(up.Profile.Gender),
		Seeking:      string(up.Profile.Seeking),
		Language:     string(up.Profile.Language),
		ActiveDialog: activeDialog,
	}
}

func parseGender(v string) (shared.Gender, error) {
	switch shared.Gender(v) {
	case shared.GenderMale, shared.GenderFemale:
		return shared.Gender(v), nil
	default:
		return "", errors.New("invalid gender")
	}
}

func parseLanguage(v string) (shared.Language, error) {
	switch shared.Language(v) {
	case shared.LanguageRU, shared.LanguageEN:
		return shared.Language(v), nil
	default:
		return "", errors.New("invalid language")
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
