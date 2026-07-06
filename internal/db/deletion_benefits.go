package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DeletionBenefitsRepo struct {
	pool *pgxpool.Pool
}

func NewDeletionBenefitsRepo(pool *pgxpool.Pool) *DeletionBenefitsRepo {
	return &DeletionBenefitsRepo{pool: pool}
}

func (r *DeletionBenefitsRepo) FreeUnlockUsed(ctx context.Context, telegramID int64) (bool, error) {
	var usedAt *time.Time
	err := r.pool.QueryRow(ctx, `
		SELECT free_unlock_used_at
		FROM deletion_benefits
		WHERE telegram_id = $1
	`, telegramID).Scan(&usedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("get deletion benefits: %w", err)
	}
	return usedAt != nil, nil
}

func (r *DeletionBenefitsRepo) MarkFreeUnlockUsed(ctx context.Context, telegramID int64) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO deletion_benefits (telegram_id, free_unlock_used_at)
		VALUES ($1, NOW())
		ON CONFLICT (telegram_id) DO UPDATE
		SET free_unlock_used_at = EXCLUDED.free_unlock_used_at
	`, telegramID)
	if err != nil {
		return fmt.Errorf("mark free unlock used: %w", err)
	}
	return nil
}
