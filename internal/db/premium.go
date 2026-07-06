package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

type PremiumStatus struct {
	Active    bool
	ExpiresAt *time.Time
}

func (r *UsersRepo) GetPremiumStatus(ctx context.Context, userID int64) (PremiumStatus, error) {
	var expiresAt time.Time
	err := r.pool.QueryRow(ctx, `
		SELECT expires_at
		FROM premium_subscriptions
		WHERE user_id = $1 AND expires_at > NOW()
		ORDER BY expires_at DESC
		LIMIT 1
	`, userID).Scan(&expiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return PremiumStatus{Active: false}, nil
	}
	if err != nil {
		return PremiumStatus{}, fmt.Errorf("get premium status: %w", err)
	}
	expiresAt = expiresAt.UTC()
	return PremiumStatus{Active: true, ExpiresAt: &expiresAt}, nil
}

func (r *UsersRepo) IsPremium(ctx context.Context, userID int64) (bool, error) {
	status, err := r.GetPremiumStatus(ctx, userID)
	if err != nil {
		return false, err
	}
	return status.Active, nil
}

func (r *UsersRepo) ExtendPremium(ctx context.Context, userID int64, durationDays int) (time.Time, error) {
	var expiresAt time.Time
	err := r.pool.QueryRow(ctx, `
		INSERT INTO premium_subscriptions (user_id, expires_at)
		VALUES ($1, GREATEST(NOW(), COALESCE(
			(SELECT expires_at FROM premium_subscriptions WHERE user_id = $1 ORDER BY expires_at DESC LIMIT 1),
			NOW()
		)) + make_interval(days => $2))
		RETURNING expires_at
	`, userID, durationDays).Scan(&expiresAt)
	if err != nil {
		return time.Time{}, fmt.Errorf("extend premium: %w", err)
	}
	return expiresAt.UTC(), nil
}
