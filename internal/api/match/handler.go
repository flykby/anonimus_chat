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

type telegramRequest struct {
	TelegramID int64 `json:"telegram_id"`
}

type completeRequest struct {
	TelegramID int64 `json:"telegram_id"`
	WaitSec    int   `json:"wait_sec"`
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /match/start", h.start)
	mux.HandleFunc("POST /match/complete", h.complete)
	mux.HandleFunc("POST /match/cancel", h.cancel)
}

func (h *Handler) start(w http.ResponseWriter, r *http.Request) {
	var req telegramRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	if req.TelegramID <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "telegram_id required"})
		return
	}

	resp, err := h.Match.Start(r.Context(), req.TelegramID)
	writeStartResponse(w, resp, err)
}

func (h *Handler) complete(w http.ResponseWriter, r *http.Request) {
	var req completeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	if req.TelegramID <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "telegram_id required"})
		return
	}
	if req.WaitSec < 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "wait_sec must be >= 0"})
		return
	}

	resp, err := h.Match.CompleteAI(r.Context(), req.TelegramID, req.WaitSec)
	writeStartResponse(w, resp, err)
}

func (h *Handler) cancel(w http.ResponseWriter, r *http.Request) {
	var req telegramRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	if req.TelegramID <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "telegram_id required"})
		return
	}

	err := h.Match.Cancel(r.Context(), req.TelegramID)
	if errors.Is(err, match.ErrUserNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not_found"})
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
	writeJSON(w, http.StatusOK, map[string]string{"status": "cancelled"})
}

func writeStartResponse(w http.ResponseWriter, resp match.StartResponse, err error) {
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
