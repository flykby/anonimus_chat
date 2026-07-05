package fsm_test

import (
	"context"
	"testing"
	"time"

	"github.com/flykby/anonimus_chat/internal/redis/fsm"
	"github.com/flykby/anonimus_chat/internal/redis/keys"
	"github.com/flykby/anonimus_chat/internal/redis/redistest"
)

func TestFSMPersistAndTTL(t *testing.T) {
	mr, client := redistest.NewTestClient(t)
	store := fsm.New(client)
	ctx := context.Background()

	if err := store.Set(ctx, 12345, "registration:age"); err != nil {
		t.Fatalf("set: %v", err)
	}

	state, ok, err := store.Get(ctx, 12345)
	if err != nil || !ok || state != "registration:age" {
		t.Fatalf("get = %q ok=%v err=%v", state, ok, err)
	}

	ttl := mr.TTL(keys.FSM(12345))
	if ttl <= 0 || ttl > 24*time.Hour {
		t.Fatalf("ttl = %v", ttl)
	}

	if err := store.Delete(ctx, 12345); err != nil {
		t.Fatalf("delete: %v", err)
	}
	_, ok, _ = store.Get(ctx, 12345)
	if ok {
		t.Fatal("expected fsm cleared")
	}
}

func TestFSMSurvivesClientReconnect(t *testing.T) {
	mr, client := redistest.NewTestClient(t)
	store := fsm.New(client)
	ctx := context.Background()

	_ = store.Set(ctx, 999, "menu:main")
	_ = client.Close()

	client2 := redistest.ReopenClient(t, mr)
	store2 := fsm.New(client2)

	state, ok, err := store2.Get(ctx, 999)
	if err != nil || !ok || state != "menu:main" {
		t.Fatalf("state after reconnect = %q ok=%v err=%v", state, ok, err)
	}
}
