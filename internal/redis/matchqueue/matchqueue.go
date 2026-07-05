package matchqueue

import (
	"context"
	"fmt"
	"strconv"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/flykby/anonimus_chat/internal/redis/keys"
	"github.com/flykby/anonimus_chat/internal/shared"
)

var popPairScript = goredis.NewScript(`
local members = redis.call('ZRANGE', KEYS[1], 0, 1)
if #members < 2 then
  return {}
end
redis.call('ZREM', KEYS[1], members[1], members[2])
return members
`)

type Store struct {
	rdb *goredis.Client
}

func New(rdb *goredis.Client) *Store {
	return &Store{rdb: rdb}
}

func (s *Store) Enqueue(ctx context.Context, gender shared.Gender, userID int64) error {
	key := keys.P2PQueue(gender)
	score := float64(time.Now().UnixMilli())
	member := strconv.FormatInt(userID, 10)
	if err := s.rdb.ZAdd(ctx, key, goredis.Z{Score: score, Member: member}).Err(); err != nil {
		return fmt.Errorf("enqueue p2p: %w", err)
	}
	return nil
}

func (s *Store) Leave(ctx context.Context, gender shared.Gender, userID int64) error {
	key := keys.P2PQueue(gender)
	member := strconv.FormatInt(userID, 10)
	if err := s.rdb.ZRem(ctx, key, member).Err(); err != nil {
		return fmt.Errorf("leave p2p queue: %w", err)
	}
	return nil
}

func (s *Store) QueueSize(ctx context.Context, gender shared.Gender) (int64, error) {
	n, err := s.rdb.ZCard(ctx, keys.P2PQueue(gender)).Result()
	if err != nil {
		return 0, fmt.Errorf("queue size: %w", err)
	}
	return n, nil
}

// TryMatchPair atomically removes and returns two waiting users, or nil if fewer than two.
func (s *Store) TryMatchPair(ctx context.Context, gender shared.Gender) ([2]int64, bool, error) {
	key := keys.P2PQueue(gender)
	res, err := popPairScript.Run(ctx, s.rdb, []string{key}).StringSlice()
	if err != nil {
		return [2]int64{}, false, fmt.Errorf("try match pair: %w", err)
	}
	if len(res) < 2 {
		return [2]int64{}, false, nil
	}
	a, err := strconv.ParseInt(res[0], 10, 64)
	if err != nil {
		return [2]int64{}, false, fmt.Errorf("parse user id: %w", err)
	}
	b, err := strconv.ParseInt(res[1], 10, 64)
	if err != nil {
		return [2]int64{}, false, fmt.Errorf("parse user id: %w", err)
	}
	return [2]int64{a, b}, true, nil
}
