package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PaymentType string

const (
	PaymentTypePremium PaymentType = "premium"
)

type Payment struct {
	ID               int64
	UserID           int64
	Type             PaymentType
	AmountStars      int
	TelegramChargeID string
	ProviderChargeID string
	CreatedAt        time.Time
}

type PaymentsRepo struct {
	pool *pgxpool.Pool
}

func NewPaymentsRepo(pool *pgxpool.Pool) *PaymentsRepo {
	return &PaymentsRepo{pool: pool}
}

func (r *PaymentsRepo) Insert(ctx context.Context, p Payment) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx, `
		INSERT INTO payments (user_id, type, amount_stars, telegram_charge_id, provider_charge_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, p.UserID, p.Type, p.AmountStars, p.TelegramChargeID, p.ProviderChargeID).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert payment: %w", err)
	}
	return id, nil
}

func (r *PaymentsRepo) ExistsByChargeID(ctx context.Context, telegramChargeID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM payments WHERE telegram_charge_id = $1)
	`, telegramChargeID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check payment exists: %w", err)
	}
	return exists, nil
}
