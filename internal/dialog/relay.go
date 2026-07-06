package dialog

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/flykby/anonimus_chat/internal/db"
	"github.com/flykby/anonimus_chat/internal/events"
	"github.com/flykby/anonimus_chat/internal/shared"
)

const (
	p2pMessageRateLimit  = 30
	p2pMessageRateWindow = time.Minute
	p2pPhotoMaxPerDialog = 3
	p2pRateAction        = "p2p_message"
)

type RelayKind string

const (
	RelayKindText    RelayKind = "text"
	RelayKindPhoto   RelayKind = "photo"
	RelayKindSticker RelayKind = "sticker"
)

var (
	ErrRateLimited    = errors.New("rate limited")
	ErrPhotoLimit     = errors.New("photo limit")
	ErrInvalidRelay   = errors.New("invalid relay")
	ErrNotP2PDialog   = errors.New("not p2p dialog")
	ErrDialogInactive = errors.New("dialog inactive")
)

type RelayRequest struct {
	DialogID       int64
	UserID         int64
	Kind           RelayKind
	Text           string
	TelegramFileID string
}

type RelayResponse struct {
	Status            string `json:"status"`
	PartnerTelegramID int64  `json:"partner_telegram_id"`
	PartnerLanguage   string `json:"partner_language"`
	Kind              string `json:"kind"`
	Text              string `json:"text,omitempty"`
	TelegramFileID    string `json:"telegram_file_id,omitempty"`
}

type ReportRequest struct {
	DialogID int64
	UserID   int64
	Reason   string
}

type BlockRequest struct {
	DialogID int64
	UserID   int64
}

type ReportResponse struct {
	Status   string `json:"status"`
	DialogID int64  `json:"dialog_id"`
}

func (s *Service) Relay(ctx context.Context, req RelayRequest) (RelayResponse, error) {
	dialogRow, partnerUserID, err := s.requireActiveP2P(ctx, req.DialogID, req.UserID)
	if err != nil {
		return RelayResponse{}, err
	}

	if s.ratelimit != nil {
		allowed, err := s.ratelimit.Allow(ctx, req.UserID, p2pRateAction, p2pMessageRateLimit, p2pMessageRateWindow)
		if err != nil {
			return RelayResponse{}, fmt.Errorf("rate limit: %w", err)
		}
		if !allowed {
			return RelayResponse{}, ErrRateLimited
		}
	}

	content, deliveryText, deliveryFileID, err := normalizeRelayPayload(req)
	if err != nil {
		return RelayResponse{}, err
	}

	if req.Kind == RelayKindPhoto {
		count, err := s.dialogs.PhotoCount(ctx, s.pool, req.DialogID)
		if err != nil {
			return RelayResponse{}, err
		}
		if count >= p2pPhotoMaxPerDialog {
			return RelayResponse{}, ErrPhotoLimit
		}
	}

	partner, ok, err := s.users.GetByUserID(ctx, partnerUserID)
	if err != nil {
		return RelayResponse{}, err
	}
	if !ok {
		return RelayResponse{}, fmt.Errorf("partner not found")
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return RelayResponse{}, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := s.dialogs.InsertMessage(ctx, tx, req.DialogID, shared.MessageRoleUser, content); err != nil {
		return RelayResponse{}, err
	}

	senderID := req.UserID
	dialogID := req.DialogID
	contentLen := len([]rune(deliveryText))
	if contentLen == 0 && deliveryFileID != "" {
		contentLen = len(deliveryFileID)
	}
	if err := s.events.Emit(ctx, tx, events.Input{
		UserID:   &senderID,
		DialogID: &dialogID,
		Type:     events.TypeMessageSent,
		Metadata: events.MessageSentMeta{ContentLength: contentLen},
	}); err != nil {
		return RelayResponse{}, fmt.Errorf("emit message.sent: %w", err)
	}

	receiverID := partnerUserID
	if err := s.events.Emit(ctx, tx, events.Input{
		UserID:   &receiverID,
		DialogID: &dialogID,
		Type:     events.TypeMessageReceived,
		Metadata: events.MessageReceivedMeta{
			Source:        "partner",
			ContentLength: contentLen,
		},
	}); err != nil {
		return RelayResponse{}, fmt.Errorf("emit message.received: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return RelayResponse{}, fmt.Errorf("commit tx: %w", err)
	}

	resp := RelayResponse{
		Status:            "relayed",
		PartnerTelegramID: partner.User.TelegramID,
		PartnerLanguage:   string(partner.Profile.Language),
		Kind:              string(req.Kind),
	}
	switch req.Kind {
	case RelayKindText:
		resp.Text = deliveryText
	case RelayKindPhoto, RelayKindSticker:
		resp.TelegramFileID = deliveryFileID
	}
	_ = dialogRow
	return resp, nil
}

func (s *Service) Report(ctx context.Context, req ReportRequest) (ReportResponse, error) {
	dialogRow, _, err := s.requireActiveP2P(ctx, req.DialogID, req.UserID)
	if err != nil {
		return ReportResponse{}, err
	}

	reporterID := req.UserID
	dialogID := dialogRow.ID
	if err := s.events.Emit(ctx, s.pool, events.Input{
		UserID:   &reporterID,
		DialogID: &dialogID,
		Type:     events.TypeDialogReported,
		Metadata: events.DialogReportedMeta{
			ReporterUserID: req.UserID,
			Reason:         req.Reason,
		},
	}); err != nil {
		return ReportResponse{}, fmt.Errorf("emit dialog.reported: %w", err)
	}

	return ReportResponse{Status: "reported", DialogID: dialogRow.ID}, nil
}

func (s *Service) Block(ctx context.Context, req BlockRequest) (EndResponse, error) {
	dialogRow, partnerUserID, err := s.requireActiveP2P(ctx, req.DialogID, req.UserID)
	if err != nil {
		return EndResponse{}, err
	}

	if s.blockpair != nil {
		if err := s.blockpair.Block(ctx, req.UserID, partnerUserID); err != nil {
			return EndResponse{}, err
		}
	}

	return s.End(ctx, EndRequest{
		DialogID: dialogRow.ID,
		UserID:   req.UserID,
		Reason:   "blocked",
	})
}

func (s *Service) requireActiveP2P(ctx context.Context, dialogID, userID int64) (db.DialogRow, int64, error) {
	dialogRow, ok, err := s.dialogs.GetByIDForUser(ctx, dialogID, userID)
	if err != nil {
		return db.DialogRow{}, 0, err
	}
	if !ok {
		return db.DialogRow{}, 0, ErrDialogNotFound
	}
	if dialogRow.EndedAt != nil {
		return db.DialogRow{}, 0, ErrDialogInactive
	}
	if dialogRow.Type != shared.DialogTypeP2P || dialogRow.PartnerUserID == nil {
		return db.DialogRow{}, 0, ErrNotP2PDialog
	}
	return dialogRow, *dialogRow.PartnerUserID, nil
}

func normalizeRelayPayload(req RelayRequest) (storedContent, deliveryText, deliveryFileID string, err error) {
	switch req.Kind {
	case RelayKindText:
		text := strings.TrimSpace(req.Text)
		if text == "" {
			return "", "", "", ErrInvalidRelay
		}
		return text, text, "", nil
	case RelayKindPhoto:
		fileID := strings.TrimSpace(req.TelegramFileID)
		if fileID == "" {
			return "", "", "", ErrInvalidRelay
		}
		return "photo:" + fileID, "", fileID, nil
	case RelayKindSticker:
		fileID := strings.TrimSpace(req.TelegramFileID)
		if fileID == "" {
			return "", "", "", ErrInvalidRelay
		}
		return "sticker:" + fileID, "", fileID, nil
	default:
		return "", "", "", ErrInvalidRelay
	}
}
