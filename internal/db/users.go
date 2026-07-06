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

var ErrUserAlreadyRegistered = errors.New("user already registered")

type UserProfile struct {
	User    shared.User
	Profile shared.Profile
}

type UsersRepo struct {
	pool   *pgxpool.Pool
	events *events.Emitter
}

func NewUsersRepo(pool *pgxpool.Pool, emitter *events.Emitter) *UsersRepo {
	return &UsersRepo{pool: pool, events: emitter}
}

func (r *UsersRepo) GetByTelegramID(ctx context.Context, telegramID int64) (UserProfile, bool, error) {
	const q = `
		SELECT u.id, u.telegram_id, u.public_uuid::text, u.created_at, u.deleted_at,
		       p.gender, p.seeking, p.age, p.language
		FROM users u
		JOIN profiles p ON p.user_id = u.id
		WHERE u.telegram_id = $1 AND u.deleted_at IS NULL
	`
	var up UserProfile
	var deletedAt *time.Time
	row := r.pool.QueryRow(ctx, q, telegramID)
	err := row.Scan(
		&up.User.ID,
		&up.User.TelegramID,
		&up.User.PublicUUID,
		&up.User.CreatedAt,
		&deletedAt,
		&up.Profile.Gender,
		&up.Profile.Seeking,
		&up.Profile.Age,
		&up.Profile.Language,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return UserProfile{}, false, nil
	}
	if err != nil {
		return UserProfile{}, false, fmt.Errorf("get user by telegram id: %w", err)
	}
	up.User.DeletedAt = deletedAt
	up.Profile.UserID = up.User.ID
	return up, true, nil
}

func (r *UsersRepo) GetByUserID(ctx context.Context, userID int64) (UserProfile, bool, error) {
	const q = `
		SELECT u.id, u.telegram_id, u.public_uuid::text, u.created_at, u.deleted_at,
		       p.gender, p.seeking, p.age, p.language
		FROM users u
		JOIN profiles p ON p.user_id = u.id
		WHERE u.id = $1 AND u.deleted_at IS NULL
	`
	var up UserProfile
	var deletedAt *time.Time
	row := r.pool.QueryRow(ctx, q, userID)
	err := row.Scan(
		&up.User.ID,
		&up.User.TelegramID,
		&up.User.PublicUUID,
		&up.User.CreatedAt,
		&deletedAt,
		&up.Profile.Gender,
		&up.Profile.Seeking,
		&up.Profile.Age,
		&up.Profile.Language,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return UserProfile{}, false, nil
	}
	if err != nil {
		return UserProfile{}, false, fmt.Errorf("get user by id: %w", err)
	}
	up.User.DeletedAt = deletedAt
	up.Profile.UserID = up.User.ID
	return up, true, nil
}

func (r *UsersRepo) HasActiveDialog(ctx context.Context, userID int64) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM dialogs WHERE user_id = $1 AND ended_at IS NULL
		)
	`, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("has active dialog: %w", err)
	}
	return exists, nil
}

type RegisterInput struct {
	TelegramID int64
	Age        int16
	Gender     shared.Gender
	Seeking    shared.Gender
	Language   shared.Language
}

func (r *UsersRepo) Register(ctx context.Context, in RegisterInput) (UserProfile, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return UserProfile{}, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var exists bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM users WHERE telegram_id = $1 AND deleted_at IS NULL
		)
	`, in.TelegramID).Scan(&exists); err != nil {
		return UserProfile{}, fmt.Errorf("check user exists: %w", err)
	}
	if exists {
		return UserProfile{}, ErrUserAlreadyRegistered
	}

	var up UserProfile
	var deletedAt *time.Time
	err = tx.QueryRow(ctx, `
		INSERT INTO users (telegram_id)
		VALUES ($1)
		RETURNING id, telegram_id, public_uuid::text, created_at, deleted_at
	`, in.TelegramID).Scan(
		&up.User.ID,
		&up.User.TelegramID,
		&up.User.PublicUUID,
		&up.User.CreatedAt,
		&deletedAt,
	)
	if err != nil {
		return UserProfile{}, fmt.Errorf("insert user: %w", err)
	}
	up.User.DeletedAt = deletedAt

	up.Profile = shared.Profile{
		UserID:   up.User.ID,
		Gender:   in.Gender,
		Seeking:  in.Seeking,
		Age:      in.Age,
		Language: in.Language,
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO profiles (user_id, gender, seeking, age, language)
		VALUES ($1, $2, $3, $4, $5)
	`, up.Profile.UserID, up.Profile.Gender, up.Profile.Seeking, up.Profile.Age, up.Profile.Language)
	if err != nil {
		return UserProfile{}, fmt.Errorf("insert profile: %w", err)
	}

	userID := up.User.ID
	if err := r.events.Emit(ctx, tx, events.Input{
		UserID: &userID,
		Type:   events.TypeUserRegistered,
		Metadata: events.UserRegisteredMeta{
			TelegramID: in.TelegramID,
			Age:        in.Age,
			Gender:     string(in.Gender),
			Seeking:    string(in.Seeking),
			Language:   string(in.Language),
		},
	}); err != nil {
		return UserProfile{}, fmt.Errorf("emit user.registered: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return UserProfile{}, fmt.Errorf("commit tx: %w", err)
	}
	return up, nil
}
