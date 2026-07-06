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

type relayRequest struct {
	TelegramID     int64  `json:"telegram_id"`
	Kind           string `json:"kind"`
	Text           string `json:"text,omitempty"`
	TelegramFileID string `json:"telegram_file_id,omitempty"`
}

type reportRequest struct {
	TelegramID int64  `json:"telegram_id"`
	Reason     string `json:"reason,omitempty"`
}

type telegramOnlyRequest struct {
	TelegramID int64 `json:"telegram_id"`
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /dialogs/{id}/end", h.end)
	mux.HandleFunc("POST /dialogs/{id}/relay", h.relay)
	mux.HandleFunc("POST /dialogs/{id}/report", h.report)
	mux.HandleFunc("POST /dialogs/{id}/block", h.block)
}

func (h *Handler) end(w http.ResponseWriter, r *http.Request) {
	dialogID, err := parseDialogID(r)
	if err != nil {
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
	writeEndResponse(w, resp, err)
}

func (h *Handler) relay(w http.ResponseWriter, r *http.Request) {
	dialogID, err := parseDialogID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid dialog id"})
		return
	}

	var req relayRequest
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

	resp, err := h.Dialogs.Relay(r.Context(), dialog.RelayRequest{
		DialogID:       dialogID,
		UserID:         up.User.ID,
		Kind:           dialog.RelayKind(req.Kind),
		Text:           req.Text,
		TelegramFileID: req.TelegramFileID,
	})
	if errors.Is(err, dialog.ErrDialogNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "dialog_not_found"})
		return
	}
	if errors.Is(err, dialog.ErrDialogInactive) {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "dialog_inactive"})
		return
	}
	if errors.Is(err, dialog.ErrNotP2PDialog) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "not_p2p"})
		return
	}
	if errors.Is(err, dialog.ErrRateLimited) {
		writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "rate_limited"})
		return
	}
	if errors.Is(err, dialog.ErrPhotoLimit) {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "photo_limit"})
		return
	}
	if errors.Is(err, dialog.ErrInvalidRelay) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_relay"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) report(w http.ResponseWriter, r *http.Request) {
	dialogID, err := parseDialogID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid dialog id"})
		return
	}

	var req reportRequest
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

	resp, err := h.Dialogs.Report(r.Context(), dialog.ReportRequest{
		DialogID: dialogID,
		UserID:   up.User.ID,
		Reason:   req.Reason,
	})
	writeReportResponse(w, resp, err)
}

func (h *Handler) block(w http.ResponseWriter, r *http.Request) {
	dialogID, err := parseDialogID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid dialog id"})
		return
	}

	var req telegramOnlyRequest
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

	resp, err := h.Dialogs.Block(r.Context(), dialog.BlockRequest{
		DialogID: dialogID,
		UserID:   up.User.ID,
	})
	writeEndResponse(w, resp, err)
}

func parseDialogID(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("id"), 10, 64)
}

func writeEndResponse(w http.ResponseWriter, resp dialog.EndResponse, err error) {
	if errors.Is(err, dialog.ErrDialogNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "dialog_not_found"})
		return
	}
	if errors.Is(err, dialog.ErrDialogInactive) {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "dialog_inactive"})
		return
	}
	if errors.Is(err, dialog.ErrNotP2PDialog) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "not_p2p"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func writeReportResponse(w http.ResponseWriter, resp dialog.ReportResponse, err error) {
	if errors.Is(err, dialog.ErrDialogNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "dialog_not_found"})
		return
	}
	if errors.Is(err, dialog.ErrDialogInactive) {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "dialog_inactive"})
		return
	}
	if errors.Is(err, dialog.ErrNotP2PDialog) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "not_p2p"})
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
