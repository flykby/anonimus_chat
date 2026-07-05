package ratelimit

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/flykby/anonimus_chat/internal/redis/keys"
)

type Store struct {
	rdb *goredis.Client
}

func New(rdb *goredis.Client) *Store {
	return &Store{rdb: rdb}
}

// Allow increments the counter for user/action and returns false if limit exceeded.
func (s *Store) Allow(ctx context.Context, userID int64, action string, limit int64, window time.Duration) (bool, error) {
	if limit <= 0 {
		return true, nil
	}

	key := keys.RateLimit(userID, action)
	count, err := s.rdb.Incr(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("rate limit incr: %w", err)
	}
	if count == 1 {
		if err := s.rdb.Expire(ctx, key, window).Err(); err != nil {
			return false, fmt.Errorf("rate limit expire: %w", err)
		}
	}
	return count <= limit, nil
}

func (s *Store) Count(ctx context.Context, userID int64, action string) (int64, error) {
	key := keys.RateLimit(userID, action)
	n, err := s.rdb.Get(ctx, key).Int64()
	if err == goredis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("rate limit count: %w", err)
	}
	return n, nil
}

func (s *Store) Reset(ctx context.Context, userID int64, action string) error {
	if err := s.rdb.Del(ctx, keys.RateLimit(userID, action)).Err(); err != nil {
		return fmt.Errorf("rate limit reset: %w", err)
	}
	return nil
}
