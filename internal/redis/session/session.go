package session

import (
	"context"
	"fmt"
	"strconv"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/flykby/anonimus_chat/internal/redis/keys"
	"github.com/flykby/anonimus_chat/internal/shared"
)

const (
	fieldDialogID  = "dialog_id"
	fieldType      = "type"
	fieldPartnerID = "partner_id"
	fieldPersonaID = "persona_id"
	fieldStartedAt = "started_at"
)

type ActiveSession struct {
	DialogID  int64
	Type      shared.DialogType
	PartnerID int64
	PersonaID int64
	StartedAt time.Time
}

type Store struct {
	rdb *goredis.Client
}

func New(rdb *goredis.Client) *Store {
	return &Store{rdb: rdb}
}

func (s *Store) Set(ctx context.Context, userID int64, sess ActiveSession) error {
	key := keys.Session(userID)
	fields := map[string]string{
		fieldDialogID:  strconv.FormatInt(sess.DialogID, 10),
		fieldType:      string(sess.Type),
		fieldPartnerID: strconv.FormatInt(sess.PartnerID, 10),
		fieldPersonaID: strconv.FormatInt(sess.PersonaID, 10),
		fieldStartedAt: sess.StartedAt.UTC().Format(time.RFC3339Nano),
	}
	if err := s.rdb.HSet(ctx, key, fields).Err(); err != nil {
		return fmt.Errorf("set session: %w", err)
	}
	return nil
}

func (s *Store) Get(ctx context.Context, userID int64) (ActiveSession, bool, error) {
	key := keys.Session(userID)
	vals, err := s.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return ActiveSession{}, false, fmt.Errorf("get session: %w", err)
	}
	if len(vals) == 0 {
		return ActiveSession{}, false, nil
	}

	dialogID, err := strconv.ParseInt(vals[fieldDialogID], 10, 64)
	if err != nil {
		return ActiveSession{}, false, fmt.Errorf("parse dialog_id: %w", err)
	}
	partnerID, err := strconv.ParseInt(vals[fieldPartnerID], 10, 64)
	if err != nil {
		return ActiveSession{}, false, fmt.Errorf("parse partner_id: %w", err)
	}
	personaID, err := strconv.ParseInt(vals[fieldPersonaID], 10, 64)
	if err != nil {
		return ActiveSession{}, false, fmt.Errorf("parse persona_id: %w", err)
	}
	startedAt, err := time.Parse(time.RFC3339Nano, vals[fieldStartedAt])
	if err != nil {
		return ActiveSession{}, false, fmt.Errorf("parse started_at: %w", err)
	}

	return ActiveSession{
		DialogID:  dialogID,
		Type:      shared.DialogType(vals[fieldType]),
		PartnerID: partnerID,
		PersonaID: personaID,
		StartedAt: startedAt,
	}, true, nil
}

func (s *Store) Delete(ctx context.Context, userID int64) error {
	if err := s.rdb.Del(ctx, keys.Session(userID)).Err(); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

// SetP2PPair creates mirrored sessions for both users in a P2P match.
func (s *Store) SetP2PPair(ctx context.Context, userA, userB, dialogID int64, startedAt time.Time) error {
	sessA := ActiveSession{
		DialogID:  dialogID,
		Type:      shared.DialogTypeP2P,
		PartnerID: userB,
		StartedAt: startedAt,
	}
	sessB := ActiveSession{
		DialogID:  dialogID,
		Type:      shared.DialogTypeP2P,
		PartnerID: userA,
		StartedAt: startedAt,
	}
	if err := s.Set(ctx, userA, sessA); err != nil {
		return err
	}
	if err := s.Set(ctx, userB, sessB); err != nil {
		return err
	}
	return nil
}
