package regdraft

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
	fieldAge      = "age"
	fieldGender   = "gender"
	fieldSeeking  = "seeking"
	fieldLanguage = "language"
)

const DefaultTTL = 24 * time.Hour

type Draft struct {
	Age      int16
	Gender   shared.Gender
	Seeking  shared.Gender
	Language shared.Language
}

type Store struct {
	rdb *goredis.Client
	ttl time.Duration
}

func New(rdb *goredis.Client) *Store {
	return &Store{rdb: rdb, ttl: DefaultTTL}
}

func (s *Store) Save(ctx context.Context, telegramID int64, d Draft) error {
	fields := map[string]string{
		fieldAge:      strconv.Itoa(int(d.Age)),
		fieldGender:   string(d.Gender),
		fieldSeeking:  string(d.Seeking),
		fieldLanguage: string(d.Language),
	}
	key := keys.RegDraft(telegramID)
	if err := s.rdb.HSet(ctx, key, fields).Err(); err != nil {
		return fmt.Errorf("save regdraft: %w", err)
	}
	if err := s.rdb.Expire(ctx, key, s.ttl).Err(); err != nil {
		return fmt.Errorf("expire regdraft: %w", err)
	}
	return nil
}

func (s *Store) Get(ctx context.Context, telegramID int64) (Draft, bool, error) {
	key := keys.RegDraft(telegramID)
	vals, err := s.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return Draft{}, false, fmt.Errorf("get regdraft: %w", err)
	}
	if len(vals) == 0 {
		return Draft{}, false, nil
	}

	age, err := strconv.ParseInt(vals[fieldAge], 10, 16)
	if err != nil {
		return Draft{}, false, fmt.Errorf("parse age: %w", err)
	}

	return Draft{
		Age:      int16(age),
		Gender:   shared.Gender(vals[fieldGender]),
		Seeking:  shared.Gender(vals[fieldSeeking]),
		Language: shared.Language(vals[fieldLanguage]),
	}, true, nil
}

func (s *Store) SetAge(ctx context.Context, telegramID int64, age int16) error {
	key := keys.RegDraft(telegramID)
	if err := s.rdb.HSet(ctx, key, fieldAge, strconv.Itoa(int(age))).Err(); err != nil {
		return fmt.Errorf("set age: %w", err)
	}
	return s.touch(ctx, key)
}

func (s *Store) SetGender(ctx context.Context, telegramID int64, gender shared.Gender) error {
	key := keys.RegDraft(telegramID)
	if err := s.rdb.HSet(ctx, key, fieldGender, string(gender)).Err(); err != nil {
		return fmt.Errorf("set gender: %w", err)
	}
	return s.touch(ctx, key)
}

func (s *Store) SetSeeking(ctx context.Context, telegramID int64, seeking shared.Gender) error {
	key := keys.RegDraft(telegramID)
	if err := s.rdb.HSet(ctx, key, fieldSeeking, string(seeking)).Err(); err != nil {
		return fmt.Errorf("set seeking: %w", err)
	}
	return s.touch(ctx, key)
}

func (s *Store) SetLanguage(ctx context.Context, telegramID int64, language shared.Language) error {
	key := keys.RegDraft(telegramID)
	if err := s.rdb.HSet(ctx, key, fieldLanguage, string(language)).Err(); err != nil {
		return fmt.Errorf("set language: %w", err)
	}
	return s.touch(ctx, key)
}

func (s *Store) Delete(ctx context.Context, telegramID int64) error {
	if err := s.rdb.Del(ctx, keys.RegDraft(telegramID)).Err(); err != nil {
		return fmt.Errorf("delete regdraft: %w", err)
	}
	return nil
}

func (s *Store) touch(ctx context.Context, key string) error {
	if err := s.rdb.Expire(ctx, key, s.ttl).Err(); err != nil {
		return fmt.Errorf("expire regdraft: %w", err)
	}
	return nil
}
