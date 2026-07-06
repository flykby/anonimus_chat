package blockpair_test

import (
	"context"
	"testing"

	"github.com/flykby/anonimus_chat/internal/redis/blockpair"
	redistest "github.com/flykby/anonimus_chat/internal/redis/redistest"
)

func TestBlockAndIsBlocked(t *testing.T) {
	_, client := redistest.NewTestClient(t)
	store := blockpair.New(client)
	ctx := context.Background()

	blocked, err := store.IsBlocked(ctx, 1, 2)
	if err != nil {
		t.Fatalf("IsBlocked: %v", err)
	}
	if blocked {
		t.Fatal("expected not blocked initially")
	}

	if err := store.Block(ctx, 2, 1); err != nil {
		t.Fatalf("Block: %v", err)
	}

	blocked, err = store.IsBlocked(ctx, 1, 2)
	if err != nil || !blocked {
		t.Fatalf("blocked = %v err = %v", blocked, err)
	}
}
