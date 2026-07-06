package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/flykby/anonimus_chat/internal/events"
	"github.com/flykby/anonimus_chat/internal/shared"
)

var ErrDialogNotFound = errors.New("dialog not found")

type DialogRow struct {
	ID            int64
	UserID        int64
	Type          shared.DialogType
	PersonaID     *int64
	PartnerUserID *int64
	StartedAt     time.Time
	EndedAt       *time.Time
	EndReason     *string
}

type DialogsRepo struct {
	pool *pgxpool.Pool
}

func NewDialogsRepo(pool *pgxpool.Pool) *DialogsRepo {
	return &DialogsRepo{pool: pool}
}

func (r *DialogsRepo) CreateAI(ctx context.Context, db events.DBTX, userID int64, personaID *int64) (int64, error) {
	var dialogID int64
	err := db.QueryRow(ctx, `
		INSERT INTO dialogs (user_id, type, persona_id)
		VALUES ($1, $2, $3)
		RETURNING id
	`, userID, shared.DialogTypeAI, personaID).Scan(&dialogID)
	if err != nil {
		return 0, fmt.Errorf("insert ai dialog: %w", err)
	}
	return dialogID, nil
}

func (r *DialogsRepo) CreateP2P(ctx context.Context, db events.DBTX, userA, userB int64) (dialogAID, dialogBID int64, err error) {
	err = db.QueryRow(ctx, `
		INSERT INTO dialogs (user_id, type, partner_user_id)
		VALUES ($1, $2, $3)
		RETURNING id
	`, userA, shared.DialogTypeP2P, userB).Scan(&dialogAID)
	if err != nil {
		return 0, 0, fmt.Errorf("insert p2p dialog for user a: %w", err)
	}

	err = db.QueryRow(ctx, `
		INSERT INTO dialogs (user_id, type, partner_user_id)
		VALUES ($1, $2, $3)
		RETURNING id
	`, userB, shared.DialogTypeP2P, userA).Scan(&dialogBID)
	if err != nil {
		return 0, 0, fmt.Errorf("insert p2p dialog for user b: %w", err)
	}
	return dialogAID, dialogBID, nil
}

func (r *DialogsRepo) GetActiveByUserID(ctx context.Context, userID int64) (DialogRow, bool, error) {
	return r.scanDialog(r.pool.QueryRow(ctx, `
		SELECT id, user_id, type, persona_id, partner_user_id, started_at, ended_at, end_reason
		FROM dialogs
		WHERE user_id = $1 AND ended_at IS NULL
		ORDER BY started_at DESC
		LIMIT 1
	`, userID))
}

func (r *DialogsRepo) GetByIDForUser(ctx context.Context, dialogID, userID int64) (DialogRow, bool, error) {
	return r.scanDialog(r.pool.QueryRow(ctx, `
		SELECT id, user_id, type, persona_id, partner_user_id, started_at, ended_at, end_reason
		FROM dialogs
		WHERE id = $1 AND user_id = $2
	`, dialogID, userID))
}

func (r *DialogsRepo) GetActivePartnerDialog(ctx context.Context, db events.DBTX, userID, partnerUserID int64) (DialogRow, bool, error) {
	return r.scanDialog(db.QueryRow(ctx, `
		SELECT id, user_id, type, persona_id, partner_user_id, started_at, ended_at, end_reason
		FROM dialogs
		WHERE user_id = $1 AND partner_user_id = $2 AND ended_at IS NULL
		ORDER BY started_at DESC
		LIMIT 1
	`, partnerUserID, userID))
}

func (r *DialogsRepo) MessageCount(ctx context.Context, db events.DBTX, dialogID int64) (int, error) {
	var count int
	err := db.QueryRow(ctx, `
		SELECT COUNT(*) FROM dialog_messages WHERE dialog_id = $1
	`, dialogID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count dialog messages: %w", err)
	}
	return count, nil
}

func (r *DialogsRepo) MarkEnded(ctx context.Context, db events.DBTX, dialogID int64, reason string, endedAt time.Time) error {
	tag, err := db.Exec(ctx, `
		UPDATE dialogs
		SET ended_at = $2, end_reason = $3
		WHERE id = $1 AND ended_at IS NULL
	`, dialogID, endedAt, reason)
	if err != nil {
		return fmt.Errorf("mark dialog ended: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return nil
	}
	return nil
}

func (r *DialogsRepo) scanDialog(row pgx.Row) (DialogRow, bool, error) {
	var d DialogRow
	err := row.Scan(
		&d.ID,
		&d.UserID,
		&d.Type,
		&d.PersonaID,
		&d.PartnerUserID,
		&d.StartedAt,
		&d.EndedAt,
		&d.EndReason,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return DialogRow{}, false, nil
	}
	if err != nil {
		return DialogRow{}, false, fmt.Errorf("scan dialog: %w", err)
	}
	return d, true, nil
}
