package session_test

import (
	"context"
	"testing"
	"time"

	"github.com/flykby/anonimus_chat/internal/redis/redistest"
	"github.com/flykby/anonimus_chat/internal/redis/session"
	"github.com/flykby/anonimus_chat/internal/shared"
)

func TestSetGetDeleteSession(t *testing.T) {
	_, client := redistest.NewTestClient(t)
	store := session.New(client)
	ctx := context.Background()

	started := time.Now().UTC().Truncate(time.Millisecond)
	sess := session.ActiveSession{
		DialogID:  99,
		Type:      shared.DialogTypeAI,
		PartnerID: 0,
		PersonaID: 3,
		StartedAt: started,
	}

	if err := store.Set(ctx, 7, sess); err != nil {
		t.Fatalf("set: %v", err)
	}

	got, ok, err := store.Get(ctx, 7)
	if err != nil || !ok {
		t.Fatalf("get: ok=%v err=%v", ok, err)
	}
	if got.DialogID != 99 || got.Type != shared.DialogTypeAI || got.PersonaID != 3 {
		t.Fatalf("session = %+v", got)
	}

	if err := store.Delete(ctx, 7); err != nil {
		t.Fatalf("delete: %v", err)
	}
	_, ok, err = store.Get(ctx, 7)
	if err != nil || ok {
		t.Fatalf("expected empty session after delete")
	}
}

func TestSetP2PPair(t *testing.T) {
	_, client := redistest.NewTestClient(t)
	store := session.New(client)
	ctx := context.Background()

	started := time.Now().UTC()
	if err := store.SetP2PPair(ctx, 1, 2, 50, 51, started); err != nil {
		t.Fatalf("set pair: %v", err)
	}

	a, ok, _ := store.Get(ctx, 1)
	if !ok || a.PartnerID != 2 || a.Type != shared.DialogTypeP2P || a.DialogID != 50 {
		t.Fatalf("user 1 session = %+v ok=%v", a, ok)
	}
	b, ok, _ := store.Get(ctx, 2)
	if !ok || b.PartnerID != 1 || b.DialogID != 51 {
		t.Fatalf("user 2 session = %+v ok=%v", b, ok)
	}
}

func TestSessionReadLatencyLocal(t *testing.T) {
	_, client := redistest.NewTestClient(t)
	store := session.New(client)
	ctx := context.Background()

	_ = store.Set(ctx, 1, session.ActiveSession{
		DialogID: 1, Type: shared.DialogTypeP2P, StartedAt: time.Now().UTC(),
	})

	start := time.Now()
	_, _, err := store.Get(ctx, 1)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if time.Since(start) > 5*time.Millisecond {
		t.Fatalf("read took %v, want < 5ms on local redis", time.Since(start))
	}
}
