package match

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/flykby/anonimus_chat/internal/db"
	"github.com/flykby/anonimus_chat/internal/events"
	"github.com/flykby/anonimus_chat/internal/redis/matchqueue"
)

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrActiveDialog     = errors.New("active dialog")
	ErrQueueUnavailable = errors.New("queue unavailable")
)

type StartResponse struct {
	Route      string `json:"route"`
	Status     string `json:"status"`
	DialogID   *int64 `json:"dialog_id,omitempty"`
	QueueSize  *int64 `json:"queue_size,omitempty"`
	MatchRoute string `json:"match_route"`
}

type Service struct {
	pool    *pgxpool.Pool
	users   *db.UsersRepo
	dialogs *db.DialogsRepo
	queue   *matchqueue.Store
	events  *events.Emitter
}

func NewService(
	pool *pgxpool.Pool,
	users *db.UsersRepo,
	dialogs *db.DialogsRepo,
	queue *matchqueue.Store,
	emitter *events.Emitter,
) *Service {
	return &Service{
		pool:    pool,
		users:   users,
		dialogs: dialogs,
		queue:   queue,
		events:  emitter,
	}
}

func (s *Service) Start(ctx context.Context, telegramID int64) (StartResponse, error) {
	up, ok, err := s.users.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return StartResponse{}, err
	}
	if !ok {
		return StartResponse{}, ErrUserNotFound
	}

	active, err := s.users.HasActiveDialog(ctx, up.User.ID)
	if err != nil {
		return StartResponse{}, err
	}
	if active {
		return StartResponse{}, ErrActiveDialog
	}

	resolved := Resolve(up.Profile.Gender, up.Profile.Seeking)
	switch resolved.Route {
	case RouteAI:
		return s.startAI(ctx, up, resolved)
	case RouteP2P:
		return s.startP2P(ctx, up, resolved)
	default:
		return StartResponse{}, fmt.Errorf("unsupported route %q", resolved.Route)
	}
}

func (s *Service) startAI(ctx context.Context, up db.UserProfile, resolved Result) (StartResponse, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return StartResponse{}, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	dialogID, err := s.dialogs.CreateAI(ctx, tx, up.User.ID, nil)
	if err != nil {
		return StartResponse{}, err
	}

	userID := up.User.ID
	dialogIDCopy := dialogID
	if err := s.events.Emit(ctx, tx, events.Input{
		UserID: &userID,
		Type:   events.TypeQueueEntered,
		Metadata: events.QueueEnteredMeta{
			Route:   string(RouteAI),
			Gender:  string(up.Profile.Gender),
			Seeking: string(up.Profile.Seeking),
		},
	}); err != nil {
		return StartResponse{}, fmt.Errorf("emit queue.entered: %w", err)
	}
	if err := s.events.Emit(ctx, tx, events.Input{
		UserID:   &userID,
		DialogID: &dialogIDCopy,
		Type:     events.TypeDialogStarted,
		Metadata: events.DialogStartedMeta{
			Type:       string(RouteAI),
			MatchRoute: resolved.MatchRoute,
		},
	}); err != nil {
		return StartResponse{}, fmt.Errorf("emit dialog.started: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return StartResponse{}, fmt.Errorf("commit tx: %w", err)
	}

	return StartResponse{
		Route:      string(RouteAI),
		Status:     "matched",
		DialogID:   &dialogID,
		MatchRoute: resolved.MatchRoute,
	}, nil
}

func (s *Service) startP2P(ctx context.Context, up db.UserProfile, resolved Result) (StartResponse, error) {
	if s.queue == nil {
		return StartResponse{}, ErrQueueUnavailable
	}

	if err := s.queue.Enqueue(ctx, up.Profile.Gender, up.User.ID); err != nil {
		return StartResponse{}, fmt.Errorf("enqueue p2p: %w", err)
	}

	userID := up.User.ID
	if err := s.events.Emit(ctx, s.pool, events.Input{
		UserID: &userID,
		Type:   events.TypeQueueEntered,
		Metadata: events.QueueEnteredMeta{
			Route:   string(RouteP2P),
			Gender:  string(up.Profile.Gender),
			Seeking: string(up.Profile.Seeking),
		},
	}); err != nil {
		_ = s.queue.Leave(ctx, up.Profile.Gender, up.User.ID)
		return StartResponse{}, fmt.Errorf("emit queue.entered: %w", err)
	}

	size, err := s.queue.QueueSize(ctx, up.Profile.Gender)
	if err != nil {
		return StartResponse{}, err
	}

	return StartResponse{
		Route:      string(RouteP2P),
		Status:     "queued",
		QueueSize:  &size,
		MatchRoute: resolved.MatchRoute,
	}, nil
}
