package users

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/flykby/anonimus_chat/internal/db"
	"github.com/flykby/anonimus_chat/internal/profile"
	"github.com/flykby/anonimus_chat/internal/shared"
)

type Handler struct {
	Users   *db.UsersRepo
	Dialogs *db.DialogsRepo
	Delete  *profile.DeleteService
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

type patchProfileRequest struct {
	TelegramID int64   `json:"telegram_id"`
	Age        *int16  `json:"age,omitempty"`
	Gender     *string `json:"gender,omitempty"`
	Seeking    *string `json:"seeking,omitempty"`
	Language   *string `json:"language,omitempty"`
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /users/by-telegram/{telegram_id}", h.getByTelegram)
	mux.HandleFunc("GET /users/by-telegram/{telegram_id}/profile", h.getProfileView)
	mux.HandleFunc("GET /users/by-telegram/{telegram_id}/language", h.getLanguage)
	mux.HandleFunc("GET /users/me/profile", h.getProfileViewMe)
	mux.HandleFunc("GET /users/me/language", h.getLanguageMe)
	mux.HandleFunc("PATCH /users/me/profile", h.patchProfileMe)
	mux.HandleFunc("DELETE /users/me", h.deleteMe)
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

func (h *Handler) patchProfileMe(w http.ResponseWriter, r *http.Request) {
	var req patchProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	if req.TelegramID <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "telegram_id required"})
		return
	}

	patch, err := parseProfilePatch(req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	up, err := h.Users.UpdateProfile(r.Context(), req.TelegramID, patch)
	if errors.Is(err, pgx.ErrNoRows) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	if errors.Is(err, db.ErrActiveDialog) {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "active_dialog"})
		return
	}
	if errors.Is(err, db.ErrNoProfileChanges) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "at least one field required"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	premium, err := h.Users.GetPremiumStatus(r.Context(), up.User.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, toProfileViewResponse(up, premium))
}

func parseProfilePatch(req patchProfileRequest) (db.UpdateProfilePatch, error) {
	var patch db.UpdateProfilePatch
	if req.Age != nil {
		if *req.Age < 18 || *req.Age > 99 {
			return patch, errors.New("age must be 18-99")
		}
		patch.Age = req.Age
	}
	if req.Gender != nil {
		gender, err := parseGender(*req.Gender)
		if err != nil {
			return patch, errors.New("invalid gender")
		}
		patch.Gender = &gender
	}
	if req.Seeking != nil {
		seeking, err := parseGender(*req.Seeking)
		if err != nil {
			return patch, errors.New("invalid seeking")
		}
		patch.Seeking = &seeking
	}
	if req.Language != nil {
		language, err := parseLanguage(*req.Language)
		if err != nil {
			return patch, errors.New("invalid language")
		}
		patch.Language = &language
	}
	if patch.Age == nil && patch.Gender == nil && patch.Seeking == nil && patch.Language == nil {
		return patch, errors.New("at least one field required")
	}
	return patch, nil
}

type deleteMeRequest struct {
	TelegramID int64 `json:"telegram_id"`
}

type deleteMeResponse struct {
	Status            string  `json:"status"`
	PartnerTelegramID *int64  `json:"partner_telegram_id,omitempty"`
	PartnerLanguage   *string `json:"partner_language,omitempty"`
}

func (h *Handler) deleteMe(w http.ResponseWriter, r *http.Request) {
	if h.Delete == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "delete unavailable"})
		return
	}

	var req deleteMeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	if req.TelegramID <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "telegram_id required"})
		return
	}

	result, err := h.Delete.Delete(r.Context(), req.TelegramID)
	if errors.Is(err, db.ErrUserNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	if errors.Is(err, db.ErrAlreadyDeleted) {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "already_deleted"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	writeJSON(w, http.StatusOK, deleteMeResponse{
		Status:            "deleted",
		PartnerTelegramID: result.PartnerTelegramID,
		PartnerLanguage:   result.PartnerLanguage,
	})
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
