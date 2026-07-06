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

var popHeteroPairScript = goredis.NewScript(`
local female = redis.call('ZRANGE', KEYS[1], 0, 0)
local male = redis.call('ZRANGE', KEYS[2], 0, 0)
if #female < 1 or #male < 1 then
  return {}
end
redis.call('ZREM', KEYS[1], female[1])
redis.call('ZREM', KEYS[2], male[1])
return {female[1], male[1]}
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

func (s *Store) IsInQueue(ctx context.Context, gender shared.Gender, userID int64) (bool, error) {
	score, err := s.rdb.ZScore(ctx, keys.P2PQueue(gender), strconv.FormatInt(userID, 10)).Result()
	if err == goredis.Nil {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("p2p queue membership: %w", err)
	}
	return score > 0 || score == 0, nil
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

func (s *Store) EnqueueHetero(ctx context.Context, gender shared.Gender, userID int64) error {
	key := keys.HeteroQueue(gender)
	score := float64(time.Now().UnixMilli())
	member := strconv.FormatInt(userID, 10)
	if err := s.rdb.ZAdd(ctx, key, goredis.Z{Score: score, Member: member}).Err(); err != nil {
		return fmt.Errorf("enqueue hetero: %w", err)
	}
	return nil
}

func (s *Store) LeaveHetero(ctx context.Context, gender shared.Gender, userID int64) error {
	key := keys.HeteroQueue(gender)
	member := strconv.FormatInt(userID, 10)
	if err := s.rdb.ZRem(ctx, key, member).Err(); err != nil {
		return fmt.Errorf("leave hetero queue: %w", err)
	}
	return nil
}

func (s *Store) HeteroQueueSize(ctx context.Context, gender shared.Gender) (int64, error) {
	n, err := s.rdb.ZCard(ctx, keys.HeteroQueue(gender)).Result()
	if err != nil {
		return 0, fmt.Errorf("hetero queue size: %w", err)
	}
	return n, nil
}

func (s *Store) IsInHeteroQueue(ctx context.Context, gender shared.Gender, userID int64) (bool, error) {
	_, err := s.rdb.ZScore(ctx, keys.HeteroQueue(gender), strconv.FormatInt(userID, 10)).Result()
	if err == goredis.Nil {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("hetero queue membership: %w", err)
	}
	return true, nil
}

// TryMatchHeteroPair atomically pairs the oldest female (seeking male) with the oldest male waiter.
func (s *Store) TryMatchHeteroPair(ctx context.Context) ([2]int64, bool, error) {
	res, err := popHeteroPairScript.Run(
		ctx,
		s.rdb,
		[]string{keys.HeteroQueue(shared.GenderFemale), keys.HeteroQueue(shared.GenderMale)},
	).StringSlice()
	if err != nil {
		return [2]int64{}, false, fmt.Errorf("try match hetero pair: %w", err)
	}
	if len(res) < 2 {
		return [2]int64{}, false, nil
	}
	femaleID, err := strconv.ParseInt(res[0], 10, 64)
	if err != nil {
		return [2]int64{}, false, fmt.Errorf("parse female user id: %w", err)
	}
	maleID, err := strconv.ParseInt(res[1], 10, 64)
	if err != nil {
		return [2]int64{}, false, fmt.Errorf("parse male user id: %w", err)
	}
	return [2]int64{femaleID, maleID}, true, nil
}
