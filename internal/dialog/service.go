package dialog

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/flykby/anonimus_chat/internal/db"
	"github.com/flykby/anonimus_chat/internal/events"
	"github.com/flykby/anonimus_chat/internal/redis/blockpair"
	"github.com/flykby/anonimus_chat/internal/redis/dialogctx"
	"github.com/flykby/anonimus_chat/internal/redis/ratelimit"
	"github.com/flykby/anonimus_chat/internal/redis/session"
	"github.com/flykby/anonimus_chat/internal/shared"
)

var (
	ErrDialogNotFound = errors.New("dialog not found")
	ErrForbidden      = errors.New("forbidden")
)

type EndRequest struct {
	DialogID int64
	UserID   int64
	Reason   string
}

type EndResponse struct {
	Status            string  `json:"status"`
	DialogID          int64   `json:"dialog_id"`
	PartnerTelegramID *int64  `json:"partner_telegram_id,omitempty"`
	PartnerLanguage   *string `json:"partner_language,omitempty"`
}

type Service struct {
	pool      *pgxpool.Pool
	dialogs   *db.DialogsRepo
	users     *db.UsersRepo
	events    *events.Emitter
	sessions  *session.Store
	dialogctx *dialogctx.Store
	ratelimit *ratelimit.Store
	blockpair *blockpair.Store
}

func NewService(
	pool *pgxpool.Pool,
	dialogs *db.DialogsRepo,
	users *db.UsersRepo,
	emitter *events.Emitter,
	sessions *session.Store,
	dctx *dialogctx.Store,
	ratelimitStore *ratelimit.Store,
	blockpairStore *blockpair.Store,
) *Service {
	return &Service{
		pool:      pool,
		dialogs:   dialogs,
		users:     users,
		events:    emitter,
		sessions:  sessions,
		dialogctx: dctx,
		ratelimit: ratelimitStore,
		blockpair: blockpairStore,
	}
}

func (s *Service) End(ctx context.Context, req EndRequest) (EndResponse, error) {
	if req.Reason == "" {
		req.Reason = "user_confirmed"
	}

	dialog, ok, err := s.dialogs.GetByIDForUser(ctx, req.DialogID, req.UserID)
	if err != nil {
		return EndResponse{}, err
	}
	if !ok {
		return EndResponse{}, ErrDialogNotFound
	}
	if dialog.EndedAt != nil {
		return EndResponse{Status: "already_ended", DialogID: dialog.ID}, nil
	}

	endedAt := time.Now().UTC()
	var partnerDialog *db.DialogRow
	if dialog.Type == shared.DialogTypeP2P && dialog.PartnerUserID != nil {
		if pd, found, err := s.dialogs.GetActivePartnerDialog(ctx, s.pool, dialog.UserID, *dialog.PartnerUserID); err != nil {
			return EndResponse{}, err
		} else if found {
			partnerDialog = &pd
		}
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return EndResponse{}, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := s.endOne(ctx, tx, dialog, req.Reason, endedAt); err != nil {
		return EndResponse{}, err
	}
	if partnerDialog != nil {
		if err := s.endOne(ctx, tx, *partnerDialog, req.Reason, endedAt); err != nil {
			return EndResponse{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return EndResponse{}, fmt.Errorf("commit tx: %w", err)
	}

	s.cleanupRedis(ctx, dialog, partnerDialog)

	resp := EndResponse{Status: "ended", DialogID: dialog.ID}
	if partnerDialog != nil {
		partner, ok, err := s.users.GetByUserID(ctx, partnerDialog.UserID)
		if err != nil {
			return EndResponse{}, err
		}
		if ok {
			resp.PartnerTelegramID = &partner.User.TelegramID
			lang := string(partner.Profile.Language)
			resp.PartnerLanguage = &lang
		}
	}
	return resp, nil
}

func (s *Service) endOne(ctx context.Context, tx events.DBTX, dialog db.DialogRow, reason string, endedAt time.Time) error {
	if dialog.EndedAt != nil {
		return nil
	}

	count, err := s.dialogs.MessageCount(ctx, tx, dialog.ID)
	if err != nil {
		return err
	}
	durationSec := int(endedAt.Sub(dialog.StartedAt).Seconds())
	if durationSec < 0 {
		durationSec = 0
	}

	if err := s.dialogs.MarkEnded(ctx, tx, dialog.ID, reason, endedAt); err != nil {
		return err
	}

	userID := dialog.UserID
	dialogID := dialog.ID
	return s.events.Emit(ctx, tx, events.Input{
		UserID:   &userID,
		DialogID: &dialogID,
		Type:     events.TypeDialogEnded,
		Metadata: events.DialogEndedMeta{
			Reason:       reason,
			DurationSec:  durationSec,
			MessageCount: count,
		},
	})
}

func (s *Service) cleanupRedis(ctx context.Context, dialog db.DialogRow, partner *db.DialogRow) {
	if s.sessions != nil {
		_ = s.sessions.Delete(ctx, dialog.UserID)
		if partner != nil {
			_ = s.sessions.Delete(ctx, partner.UserID)
		}
	}
	if s.dialogctx != nil {
		_ = s.dialogctx.Delete(ctx, dialog.ID)
		if partner != nil && partner.ID != dialog.ID {
			_ = s.dialogctx.Delete(ctx, partner.ID)
		}
	}
}
