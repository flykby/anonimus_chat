package ratelimit_test

import (
	"context"
	"testing"
	"time"

	"github.com/flykby/anonimus_chat/internal/redis/ratelimit"
	"github.com/flykby/anonimus_chat/internal/redis/redistest"
)

func TestRateLimitBlocksSpam(t *testing.T) {
	_, client := redistest.NewTestClient(t)
	store := ratelimit.New(client)
	ctx := context.Background()

	const limit int64 = 3
	window := time.Minute

	for i := 0; i < 3; i++ {
		ok, err := store.Allow(ctx, 42, "message", limit, window)
		if err != nil || !ok {
			t.Fatalf("allow %d: ok=%v err=%v", i, ok, err)
		}
	}

	ok, err := store.Allow(ctx, 42, "message", limit, window)
	if err != nil {
		t.Fatalf("allow over limit: %v", err)
	}
	if ok {
		t.Fatal("expected rate limit block on 4th message")
	}
}

func TestRateLimitSeparateActions(t *testing.T) {
	_, client := redistest.NewTestClient(t)
	store := ratelimit.New(client)
	ctx := context.Background()

	ok, _ := store.Allow(ctx, 1, "photo", 1, time.Minute)
	if !ok {
		t.Fatal("expected photo allowed")
	}
	ok, _ = store.Allow(ctx, 1, "search", 1, time.Minute)
	if !ok {
		t.Fatal("expected search allowed independently")
	}
}
