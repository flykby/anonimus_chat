package matchqueue_test

import (
	"context"
	"testing"
	"time"

	"github.com/flykby/anonimus_chat/internal/redis/matchqueue"
	redistest "github.com/flykby/anonimus_chat/internal/redis/redistest"
	"github.com/flykby/anonimus_chat/internal/shared"
)

func TestEnqueueAndMatchPair(t *testing.T) {
	_, client := redistest.NewTestClient(t)
	store := matchqueue.New(client)
	ctx := context.Background()

	if err := store.Enqueue(ctx, shared.GenderMale, 1); err != nil {
		t.Fatalf("enqueue 1: %v", err)
	}
	if err := store.Enqueue(ctx, shared.GenderMale, 2); err != nil {
		t.Fatalf("enqueue 2: %v", err)
	}

	pair, ok, err := store.TryMatchPair(ctx, shared.GenderMale)
	if err != nil {
		t.Fatalf("match: %v", err)
	}
	if !ok {
		t.Fatal("expected match")
	}
	if pair[0] != 1 || pair[1] != 2 {
		t.Fatalf("pair = %v", pair)
	}

	size, err := store.QueueSize(ctx, shared.GenderMale)
	if err != nil {
		t.Fatalf("size: %v", err)
	}
	if size != 0 {
		t.Fatalf("queue size = %d, want 0", size)
	}
}

func TestMatchPairRequiresTwoUsers(t *testing.T) {
	_, client := redistest.NewTestClient(t)
	store := matchqueue.New(client)
	ctx := context.Background()

	if err := store.Enqueue(ctx, shared.GenderFemale, 10); err != nil {
		t.Fatalf("enqueue: %v", err)
	}

	_, ok, err := store.TryMatchPair(ctx, shared.GenderFemale)
	if err != nil {
		t.Fatalf("match: %v", err)
	}
	if ok {
		t.Fatal("expected no match with one user")
	}
}

func TestLeaveQueue(t *testing.T) {
	_, client := redistest.NewTestClient(t)
	store := matchqueue.New(client)
	ctx := context.Background()

	_ = store.Enqueue(ctx, shared.GenderMale, 5)
	_ = store.Leave(ctx, shared.GenderMale, 5)

	size, _ := store.QueueSize(ctx, shared.GenderMale)
	if size != 0 {
		t.Fatalf("size = %d", size)
	}
}

func TestEnqueueOrderingByTime(t *testing.T) {
	_, client := redistest.NewTestClient(t)
	store := matchqueue.New(client)
	ctx := context.Background()

	_ = store.Enqueue(ctx, shared.GenderMale, 100)
	time.Sleep(2 * time.Millisecond)
	_ = store.Enqueue(ctx, shared.GenderMale, 200)

	pair, ok, err := store.TryMatchPair(ctx, shared.GenderMale)
	if err != nil || !ok {
		t.Fatalf("match failed: ok=%v err=%v", ok, err)
	}
	if pair[0] != 100 || pair[1] != 200 {
		t.Fatalf("pair order = %v", pair)
	}
}
