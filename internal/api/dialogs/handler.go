package dialogapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/flykby/anonimus_chat/internal/db"
	"github.com/flykby/anonimus_chat/internal/dialog"
)

type Handler struct {
	Dialogs *dialog.Service
	Users   *db.UsersRepo
}

type endRequest struct {
	TelegramID int64  `json:"telegram_id"`
	Reason     string `json:"reason"`
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /dialogs/{id}/end", h.end)
}

func (h *Handler) end(w http.ResponseWriter, r *http.Request) {
	dialogID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || dialogID <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid dialog id"})
		return
	}

	var req endRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	if req.TelegramID <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "telegram_id required"})
		return
	}

	up, ok, err := h.Users.GetByTelegramID(r.Context(), req.TelegramID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not_found"})
		return
	}

	resp, err := h.Dialogs.End(r.Context(), dialog.EndRequest{
		DialogID: dialogID,
		UserID:   up.User.ID,
		Reason:   req.Reason,
	})
	if errors.Is(err, dialog.ErrDialogNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "dialog_not_found"})
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
