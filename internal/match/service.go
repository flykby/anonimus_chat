package match

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/flykby/anonimus_chat/internal/db"
	"github.com/flykby/anonimus_chat/internal/events"
	"github.com/flykby/anonimus_chat/internal/redis/blockpair"
	"github.com/flykby/anonimus_chat/internal/redis/matchqueue"
	"github.com/flykby/anonimus_chat/internal/redis/session"
	"github.com/flykby/anonimus_chat/internal/shared"
)

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrActiveDialog     = errors.New("active dialog")
	ErrQueueUnavailable = errors.New("queue unavailable")
	ErrNotInQueue       = errors.New("not in queue")
)

type StartResponse struct {
	Route        string `json:"route"`
	Status       string `json:"status"`
	DialogID     *int64 `json:"dialog_id,omitempty"`
	QueueSize    *int64 `json:"queue_size,omitempty"`
	DisplayCount *int64 `json:"display_count,omitempty"`
	MatchRoute   string `json:"match_route"`
}

type Service struct {
	pool      *pgxpool.Pool
	users     *db.UsersRepo
	dialogs   *db.DialogsRepo
	queue     *matchqueue.Store
	events    *events.Emitter
	sessions  *session.Store
	blockpair *blockpair.Store
}

func NewService(
	pool *pgxpool.Pool,
	users *db.UsersRepo,
	dialogs *db.DialogsRepo,
	queue *matchqueue.Store,
	emitter *events.Emitter,
	sessions *session.Store,
	blockpairStore *blockpair.Store,
) *Service {
	return &Service{
		pool:      pool,
		users:     users,
		dialogs:   dialogs,
		queue:     queue,
		events:    emitter,
		sessions:  sessions,
		blockpair: blockpairStore,
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

func (s *Service) Poll(ctx context.Context, telegramID int64) (StartResponse, error) {
	up, ok, err := s.users.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return StartResponse{}, err
	}
	if !ok {
		return StartResponse{}, ErrUserNotFound
	}

	resolved := Resolve(up.Profile.Gender, up.Profile.Seeking)
	if resolved.Route != RouteP2P {
		return StartResponse{}, fmt.Errorf("poll: user route is %q", resolved.Route)
	}
	if s.queue == nil {
		return StartResponse{}, ErrQueueUnavailable
	}

	if resp, matched, err := s.matchedResponseIfActive(ctx, up.User.ID, resolved); err != nil {
		return StartResponse{}, err
	} else if matched {
		return resp, nil
	}

	waiting, err := s.isWaitingInQueue(ctx, up)
	if err != nil {
		return StartResponse{}, err
	}
	if !waiting {
		return StartResponse{}, ErrNotInQueue
	}

	if err := s.tryMatch(ctx); err != nil {
		return StartResponse{}, err
	}

	if resp, matched, err := s.matchedResponseIfActive(ctx, up.User.ID, resolved); err != nil {
		return StartResponse{}, err
	} else if matched {
		return resp, nil
	}

	size, err := s.queueSizeForUser(ctx, up)
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

func (s *Service) startAI(ctx context.Context, up db.UserProfile, resolved Result) (StartResponse, error) {
	userID := up.User.ID
	if err := s.events.Emit(ctx, s.pool, events.Input{
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

	count := SyntheticDisplayCount()
	return StartResponse{
		Route:        string(RouteAI),
		Status:       "searching",
		DisplayCount: &count,
		MatchRoute:   resolved.MatchRoute,
	}, nil
}

func (s *Service) CompleteAI(ctx context.Context, telegramID int64, waitSec int) (StartResponse, error) {
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
	if resolved.Route != RouteAI {
		return StartResponse{}, fmt.Errorf("complete ai: user route is %q", resolved.Route)
	}

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
		Type:   events.TypeQueueMatched,
		Metadata: events.QueueMatchedMeta{
			Route:   string(RouteAI),
			WaitSec: waitSec,
		},
	}); err != nil {
		return StartResponse{}, fmt.Errorf("emit queue.matched: %w", err)
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

	if s.sessions != nil {
		startedAt := time.Now().UTC()
		if err := s.sessions.Set(ctx, up.User.ID, session.ActiveSession{
			DialogID:  dialogID,
			Type:      shared.DialogTypeAI,
			StartedAt: startedAt,
		}); err != nil {
			return StartResponse{}, fmt.Errorf("set session: %w", err)
		}
	}

	return StartResponse{
		Route:      string(RouteAI),
		Status:     "matched",
		DialogID:   &dialogID,
		MatchRoute: resolved.MatchRoute,
	}, nil
}

func (s *Service) Cancel(ctx context.Context, telegramID int64) error {
	up, ok, err := s.users.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrUserNotFound
	}

	resolved := Resolve(up.Profile.Gender, up.Profile.Seeking)
	if resolved.Route != RouteP2P {
		return nil
	}
	if s.queue == nil {
		return ErrQueueUnavailable
	}
	return s.leaveQueue(ctx, up)
}

func (s *Service) startP2P(ctx context.Context, up db.UserProfile, resolved Result) (StartResponse, error) {
	if s.queue == nil {
		return StartResponse{}, ErrQueueUnavailable
	}

	if err := s.enqueueUser(ctx, up); err != nil {
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
		_ = s.leaveQueue(ctx, up)
		return StartResponse{}, fmt.Errorf("emit queue.entered: %w", err)
	}

	size, err := s.queueSizeForUser(ctx, up)
	if err != nil {
		return StartResponse{}, err
	}

	if err := s.tryMatch(ctx); err != nil {
		return StartResponse{}, err
	}
	if resp, matched, err := s.matchedResponseIfActive(ctx, up.User.ID, resolved); err != nil {
		return StartResponse{}, err
	} else if matched {
		return resp, nil
	}

	return StartResponse{
		Route:      string(RouteP2P),
		Status:     "queued",
		QueueSize:  &size,
		MatchRoute: resolved.MatchRoute,
	}, nil
}

func (s *Service) tryMatch(ctx context.Context) error {
	pair, ok, err := s.queue.TryMatchPair(ctx, shared.GenderMale)
	if err != nil {
		return err
	}
	if ok {
		return s.createP2PMatch(ctx, pair[0], pair[1], "m_seeks_m")
	}

	heteroPair, ok, err := s.queue.TryMatchHeteroPair(ctx)
	if err != nil {
		return err
	}
	if ok {
		return s.createP2PMatch(ctx, heteroPair[0], heteroPair[1], "f_seeks_m")
	}

	return nil
}

func (s *Service) createP2PMatch(ctx context.Context, userAID, userBID int64, matchRoute string) error {
	if userAID == userBID {
		return nil
	}

	upA, okA, err := s.users.GetByUserID(ctx, userAID)
	if err != nil {
		return err
	}
	upB, okB, err := s.users.GetByUserID(ctx, userBID)
	if err != nil {
		return err
	}
	if !okA || !okB {
		return nil
	}

	activeA, err := s.users.HasActiveDialog(ctx, userAID)
	if err != nil {
		return err
	}
	activeB, err := s.users.HasActiveDialog(ctx, userBID)
	if err != nil {
		return err
	}
	if activeA || activeB {
		if !activeA {
			if err := s.enqueueUser(ctx, upA); err != nil {
				return err
			}
		}
		if !activeB {
			if err := s.enqueueUser(ctx, upB); err != nil {
				return err
			}
		}
		return nil
	}

	if s.blockpair != nil {
		blocked, err := s.blockpair.IsBlocked(ctx, userAID, userBID)
		if err != nil {
			return err
		}
		if blocked {
			if err := s.enqueueUser(ctx, upA); err != nil {
				return err
			}
			if err := s.enqueueUser(ctx, upB); err != nil {
				return err
			}
			return nil
		}
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	dialogAID, dialogBID, err := s.dialogs.CreateP2P(ctx, tx, userAID, userBID)
	if err != nil {
		return err
	}

	startedAt := time.Now().UTC()
	for _, item := range []struct {
		userID   int64
		dialogID int64
	}{
		{userID: userAID, dialogID: dialogAID},
		{userID: userBID, dialogID: dialogBID},
	} {
		uid := item.userID
		did := item.dialogID
		if err := s.events.Emit(ctx, tx, events.Input{
			UserID: &uid,
			Type:   events.TypeQueueMatched,
			Metadata: events.QueueMatchedMeta{
				Route: string(RouteP2P),
			},
		}); err != nil {
			return fmt.Errorf("emit queue.matched: %w", err)
		}
		if err := s.events.Emit(ctx, tx, events.Input{
			UserID:   &uid,
			DialogID: &did,
			Type:     events.TypeDialogStarted,
			Metadata: events.DialogStartedMeta{
				Type:       string(RouteP2P),
				MatchRoute: matchRoute,
			},
		}); err != nil {
			return fmt.Errorf("emit dialog.started: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	if s.sessions != nil {
		if err := s.sessions.SetP2PPair(ctx, userAID, userBID, dialogAID, dialogBID, startedAt); err != nil {
			return fmt.Errorf("set p2p sessions: %w", err)
		}
	}

	return nil
}

func (s *Service) matchedResponseIfActive(ctx context.Context, userID int64, resolved Result) (StartResponse, bool, error) {
	active, err := s.users.HasActiveDialog(ctx, userID)
	if err != nil {
		return StartResponse{}, false, err
	}
	if !active {
		return StartResponse{}, false, nil
	}

	dialog, ok, err := s.dialogs.GetActiveByUserID(ctx, userID)
	if err != nil {
		return StartResponse{}, false, err
	}
	if !ok {
		return StartResponse{}, false, nil
	}

	dialogID := dialog.ID
	return StartResponse{
		Route:      string(RouteP2P),
		Status:     "matched",
		DialogID:   &dialogID,
		MatchRoute: resolved.MatchRoute,
	}, true, nil
}

func (s *Service) enqueueUser(ctx context.Context, up db.UserProfile) error {
	target := p2pQueueFor(up.Profile.Gender, up.Profile.Seeking)
	switch target.kind {
	case p2pQueueSameGender:
		return s.queue.Enqueue(ctx, target.gender, up.User.ID)
	case p2pQueueHeteroFemale:
		return s.queue.EnqueueHetero(ctx, shared.GenderFemale, up.User.ID)
	default:
		return fmt.Errorf("unsupported p2p queue kind")
	}
}

func (s *Service) leaveQueue(ctx context.Context, up db.UserProfile) error {
	target := p2pQueueFor(up.Profile.Gender, up.Profile.Seeking)
	switch target.kind {
	case p2pQueueSameGender:
		return s.queue.Leave(ctx, target.gender, up.User.ID)
	case p2pQueueHeteroFemale:
		return s.queue.LeaveHetero(ctx, shared.GenderFemale, up.User.ID)
	default:
		return fmt.Errorf("unsupported p2p queue kind")
	}
}

func (s *Service) queueSizeForUser(ctx context.Context, up db.UserProfile) (int64, error) {
	target := p2pQueueFor(up.Profile.Gender, up.Profile.Seeking)
	switch target.kind {
	case p2pQueueSameGender:
		return s.queue.QueueSize(ctx, target.gender)
	case p2pQueueHeteroFemale:
		return s.queue.HeteroQueueSize(ctx, shared.GenderFemale)
	default:
		return 0, fmt.Errorf("unsupported p2p queue kind")
	}
}

func (s *Service) isWaitingInQueue(ctx context.Context, up db.UserProfile) (bool, error) {
	target := p2pQueueFor(up.Profile.Gender, up.Profile.Seeking)
	switch target.kind {
	case p2pQueueSameGender:
		return s.queue.IsInQueue(ctx, target.gender, up.User.ID)
	case p2pQueueHeteroFemale:
		return s.queue.IsInHeteroQueue(ctx, shared.GenderFemale, up.User.ID)
	default:
		return false, fmt.Errorf("unsupported p2p queue kind")
	}
}
