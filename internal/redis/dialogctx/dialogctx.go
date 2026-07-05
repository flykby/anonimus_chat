package dialogctx

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/flykby/anonimus_chat/internal/redis/keys"
	"github.com/flykby/anonimus_chat/internal/shared"
)

const (
	DefaultMaxMessages = 20
	DefaultTTL         = 4 * time.Hour
)

type Message struct {
	Role    shared.MessageRole `json:"role"`
	Content string             `json:"content"`
}

type Store struct {
	rdb         *goredis.Client
	maxMessages int
	ttl         time.Duration
}

func New(rdb *goredis.Client) *Store {
	return &Store{rdb: rdb, maxMessages: DefaultMaxMessages, ttl: DefaultTTL}
}

func (s *Store) Append(ctx context.Context, dialogID int64, msg Message) error {
	raw, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	key := keys.DialogContext(dialogID)
	pipe := s.rdb.Pipeline()
	pipe.LPush(ctx, key, raw)
	pipe.LTrim(ctx, key, 0, int64(s.maxMessages-1))
	pipe.Expire(ctx, key, s.ttl)
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("append dialog ctx: %w", err)
	}
	return nil
}

func (s *Store) List(ctx context.Context, dialogID int64) ([]Message, error) {
	key := keys.DialogContext(dialogID)
	raw, err := s.rdb.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("list dialog ctx: %w", err)
	}

	msgs := make([]Message, 0, len(raw))
	for i := len(raw) - 1; i >= 0; i-- {
		var m Message
		if err := json.Unmarshal([]byte(raw[i]), &m); err != nil {
			return nil, fmt.Errorf("unmarshal message: %w", err)
		}
		msgs = append(msgs, m)
	}
	return msgs, nil
}

func (s *Store) Delete(ctx context.Context, dialogID int64) error {
	if err := s.rdb.Del(ctx, keys.DialogContext(dialogID)).Err(); err != nil {
		return fmt.Errorf("delete dialog ctx: %w", err)
	}
	return nil
}

func (s *Store) Touch(ctx context.Context, dialogID int64) error {
	key := keys.DialogContext(dialogID)
	ok, err := s.rdb.Expire(ctx, key, s.ttl).Result()
	if err != nil {
		return fmt.Errorf("touch dialog ctx: %w", err)
	}
	if !ok {
		return nil
	}
	return nil
}
