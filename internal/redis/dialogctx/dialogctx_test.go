package dialogctx_test

import (
	"context"
	"testing"

	"github.com/flykby/anonimus_chat/internal/redis/dialogctx"
	"github.com/flykby/anonimus_chat/internal/redis/keys"
	"github.com/flykby/anonimus_chat/internal/redis/redistest"
	"github.com/flykby/anonimus_chat/internal/shared"
)

func TestAppendAndListMessages(t *testing.T) {
	_, client := redistest.NewTestClient(t)
	store := dialogctx.New(client)
	ctx := context.Background()

	_ = store.Append(ctx, 10, dialogctx.Message{Role: shared.MessageRoleUser, Content: "hi"})
	_ = store.Append(ctx, 10, dialogctx.Message{Role: shared.MessageRoleAssistant, Content: "hello"})

	msgs, err := store.List(ctx, 10)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(msgs) != 2 {
		t.Fatalf("len = %d", len(msgs))
	}
	if msgs[0].Content != "hi" || msgs[1].Content != "hello" {
		t.Fatalf("order = %+v", msgs)
	}
}

func TestTrimToMaxMessages(t *testing.T) {
	_, client := redistest.NewTestClient(t)
	store := dialogctx.New(client)
	ctx := context.Background()

	for i := 0; i < 25; i++ {
		_ = store.Append(ctx, 1, dialogctx.Message{Role: shared.MessageRoleUser, Content: "x"})
	}

	msgs, err := store.List(ctx, 1)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(msgs) != dialogctx.DefaultMaxMessages {
		t.Fatalf("len = %d, want %d", len(msgs), dialogctx.DefaultMaxMessages)
	}
}

func TestDialogContextHasTTL(t *testing.T) {
	mr, client := redistest.NewTestClient(t)
	store := dialogctx.New(client)
	ctx := context.Background()

	_ = store.Append(ctx, 5, dialogctx.Message{Role: shared.MessageRoleUser, Content: "ping"})
	if mr.TTL(keys.DialogContext(5)) <= 0 {
		t.Fatal("expected TTL on dialog context")
	}
}
