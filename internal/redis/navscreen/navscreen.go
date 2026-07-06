package navscreen

import (
	"context"
	"encoding/json"
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

func (s *Store) Set(ctx context.Context, telegramID int64, messageIDs []int64) error {
	raw, err := json.Marshal(messageIDs)
	if err != nil {
		return fmt.Errorf("marshal nav screen ids: %w", err)
	}
	if err := s.rdb.Set(ctx, keys.NavScreen(telegramID), raw, s.ttl).Err(); err != nil {
		return fmt.Errorf("set nav screen: %w", err)
	}
	return nil
}

func (s *Store) Get(ctx context.Context, telegramID int64) ([]int64, bool, error) {
	val, err := s.rdb.Get(ctx, keys.NavScreen(telegramID)).Result()
	if err == goredis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("get nav screen: %w", err)
	}
	var ids []int64
	if err := json.Unmarshal([]byte(val), &ids); err != nil {
		return nil, false, fmt.Errorf("unmarshal nav screen ids: %w", err)
	}
	return ids, true, nil
}

func (s *Store) Delete(ctx context.Context, telegramID int64) error {
	if err := s.rdb.Del(ctx, keys.NavScreen(telegramID)).Err(); err != nil {
		return fmt.Errorf("delete nav screen: %w", err)
	}
	return nil
}
