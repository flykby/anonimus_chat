package users

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/flykby/anonimus_chat/internal/db"
	"github.com/flykby/anonimus_chat/internal/shared"
)

type Handler struct {
	Users   *db.UsersRepo
	Dialogs *db.DialogsRepo
}

type registerRequest struct {
	TelegramID int64  `json:"telegram_id"`
	Age        int16  `json:"age"`
	Gender     string `json:"gender"`
	Seeking    string `json:"seeking"`
	Language   string `json:"language"`
}

type profileResponse struct {
	TelegramID       int64   `json:"telegram_id"`
	Age              int16   `json:"age"`
	Gender           string  `json:"gender"`
	Seeking          string  `json:"seeking"`
	Language         string  `json:"language"`
	ActiveDialog     bool    `json:"active_dialog"`
	ActiveDialogID   *int64  `json:"active_dialog_id,omitempty"`
	ActiveDialogType *string `json:"active_dialog_type,omitempty"`
}

type profileViewResponse struct {
	PublicUUID       string     `json:"public_uuid"`
	Age              int16      `json:"age"`
	Gender           string     `json:"gender"`
	Seeking          string     `json:"seeking"`
	Language         string     `json:"language"`
	PremiumActive    bool       `json:"premium_active"`
	PremiumExpiresAt *time.Time `json:"premium_expires_at,omitempty"`
}

type languageResponse struct {
	Language string `json:"language"`
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /users/by-telegram/{telegram_id}", h.getByTelegram)
	mux.HandleFunc("GET /users/by-telegram/{telegram_id}/profile", h.getProfileView)
	mux.HandleFunc("GET /users/by-telegram/{telegram_id}/language", h.getLanguage)
	mux.HandleFunc("GET /users/me/profile", h.getProfileViewMe)
	mux.HandleFunc("GET /users/me/language", h.getLanguageMe)
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

	var activeDialogID *int64
	var activeDialogType *string
	if activeDialog && h.Dialogs != nil {
		if d, ok, err := h.Dialogs.GetActiveByUserID(r.Context(), up.User.ID); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
			return
		} else if ok {
			id := d.ID
			activeDialogID = &id
			t := string(d.Type)
			activeDialogType = &t
		}
	}

	writeJSON(w, http.StatusOK, toProfileResponse(up, activeDialog, activeDialogID, activeDialogType))
}

func (h *Handler) getProfileViewMe(w http.ResponseWriter, r *http.Request) {
	telegramID, err := strconv.ParseInt(r.URL.Query().Get("telegram_id"), 10, 64)
	if err != nil || telegramID <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "telegram_id required"})
		return
	}
	h.writeProfileView(w, r, telegramID)
}

func (h *Handler) getProfileView(w http.ResponseWriter, r *http.Request) {
	telegramID, err := strconv.ParseInt(r.PathValue("telegram_id"), 10, 64)
	if err != nil || telegramID <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid telegram_id"})
		return
	}
	h.writeProfileView(w, r, telegramID)
}

func (h *Handler) writeProfileView(w http.ResponseWriter, r *http.Request, telegramID int64) {
	up, ok, err := h.Users.GetByTelegramID(r.Context(), telegramID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}

	premium, err := h.Users.GetPremiumStatus(r.Context(), up.User.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	writeJSON(w, http.StatusOK, toProfileViewResponse(up, premium))
}

func (h *Handler) getLanguageMe(w http.ResponseWriter, r *http.Request) {
	telegramID, err := strconv.ParseInt(r.URL.Query().Get("telegram_id"), 10, 64)
	if err != nil || telegramID <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "telegram_id required"})
		return
	}
	h.writeLanguage(w, r, telegramID)
}

func (h *Handler) getLanguage(w http.ResponseWriter, r *http.Request) {
	telegramID, err := strconv.ParseInt(r.PathValue("telegram_id"), 10, 64)
	if err != nil || telegramID <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid telegram_id"})
		return
	}
	h.writeLanguage(w, r, telegramID)
}

func (h *Handler) writeLanguage(w http.ResponseWriter, r *http.Request, telegramID int64) {
	up, ok, err := h.Users.GetByTelegramID(r.Context(), telegramID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	writeJSON(w, http.StatusOK, languageResponse{Language: string(up.Profile.Language)})
}

func toProfileViewResponse(up db.UserProfile, premium db.PremiumStatus) profileViewResponse {
	resp := profileViewResponse{
		PublicUUID:    up.User.PublicUUID,
		Age:           up.Profile.Age,
		Gender:        string(up.Profile.Gender),
		Seeking:       string(up.Profile.Seeking),
		Language:      string(up.Profile.Language),
		PremiumActive: premium.Active,
	}
	if premium.ExpiresAt != nil {
		expires := premium.ExpiresAt.UTC()
		resp.PremiumExpiresAt = &expires
	}
	return resp
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

	writeJSON(w, http.StatusCreated, toProfileResponse(up, false, nil, nil))
}

func toProfileResponse(up db.UserProfile, activeDialog bool, activeDialogID *int64, activeDialogType *string) profileResponse {
	return profileResponse{
		TelegramID:       up.User.TelegramID,
		Age:              up.Profile.Age,
		Gender:           string(up.Profile.Gender),
		Seeking:          string(up.Profile.Seeking),
		Language:         string(up.Profile.Language),
		ActiveDialog:     activeDialog,
		ActiveDialogID:   activeDialogID,
		ActiveDialogType: activeDialogType,
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
