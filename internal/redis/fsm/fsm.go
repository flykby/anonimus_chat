package fsm

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/flykby/anonimus_chat/internal/redis/keys"
)

const DefaultTTL = 24 * time.Hour

type Store struct {
	rdb *goredis.Client
	ttl time.Duration
}

func New(rdb *goredis.Client) *Store {
	return &Store{rdb: rdb, ttl: DefaultTTL}
}

func (s *Store) Set(ctx context.Context, telegramID int64, state string) error {
	key := keys.FSM(telegramID)
	if err := s.rdb.Set(ctx, key, state, s.ttl).Err(); err != nil {
		return fmt.Errorf("set fsm: %w", err)
	}
	return nil
}

func (s *Store) Get(ctx context.Context, telegramID int64) (string, bool, error) {
	key := keys.FSM(telegramID)
	val, err := s.rdb.Get(ctx, key).Result()
	if err == goredis.Nil {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("get fsm: %w", err)
	}
	return val, true, nil
}

func (s *Store) Delete(ctx context.Context, telegramID int64) error {
	if err := s.rdb.Del(ctx, keys.FSM(telegramID)).Err(); err != nil {
		return fmt.Errorf("delete fsm: %w", err)
	}
	return nil
}

func (s *Store) Refresh(ctx context.Context, telegramID int64) error {
	key := keys.FSM(telegramID)
	ok, err := s.rdb.Expire(ctx, key, s.ttl).Result()
	if err != nil {
		return fmt.Errorf("refresh fsm ttl: %w", err)
	}
	if !ok {
		return fmt.Errorf("fsm key not found")
	}
	return nil
}
