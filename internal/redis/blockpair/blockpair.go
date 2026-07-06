package blockpair

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/flykby/anonimus_chat/internal/redis/keys"
)

const defaultTTL = 24 * time.Hour

type Store struct {
	rdb *goredis.Client
	ttl time.Duration
}

func New(rdb *goredis.Client) *Store {
	return &Store{rdb: rdb, ttl: defaultTTL}
}

func (s *Store) Block(ctx context.Context, userA, userB int64) error {
	key := keys.BlockedPair(userA, userB)
	if err := s.rdb.Set(ctx, key, "1", s.ttl).Err(); err != nil {
		return fmt.Errorf("block pair: %w", err)
	}
	return nil
}

func (s *Store) IsBlocked(ctx context.Context, userA, userB int64) (bool, error) {
	n, err := s.rdb.Exists(ctx, keys.BlockedPair(userA, userB)).Result()
	if err != nil {
		return false, fmt.Errorf("blocked pair exists: %w", err)
	}
	return n > 0, nil
}
