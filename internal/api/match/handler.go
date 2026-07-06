package matchapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/flykby/anonimus_chat/internal/match"
)

type Handler struct {
	Match *match.Service
}

type startRequest struct {
	TelegramID int64 `json:"telegram_id"`
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /match/start", h.start)
}

func (h *Handler) start(w http.ResponseWriter, r *http.Request) {
	var req startRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	if req.TelegramID <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "telegram_id required"})
		return
	}

	resp, err := h.Match.Start(r.Context(), req.TelegramID)
	if errors.Is(err, match.ErrUserNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not_found"})
		return
	}
	if errors.Is(err, match.ErrActiveDialog) {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "active_dialog"})
		return
	}
	if errors.Is(err, match.ErrQueueUnavailable) {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "queue_unavailable"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
