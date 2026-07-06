package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/flykby/anonimus_chat/internal/events"
	"github.com/flykby/anonimus_chat/internal/shared"
)

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
