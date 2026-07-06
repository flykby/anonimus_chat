package dialog_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/flykby/anonimus_chat/internal/db"
	"github.com/flykby/anonimus_chat/internal/dialog"
	"github.com/flykby/anonimus_chat/internal/events"
	"github.com/flykby/anonimus_chat/internal/match"
	"github.com/flykby/anonimus_chat/internal/redis/matchqueue"
	"github.com/flykby/anonimus_chat/internal/redis/ratelimit"
	redistest "github.com/flykby/anonimus_chat/internal/redis/redistest"
	"github.com/flykby/anonimus_chat/internal/shared"
)

func TestIntegrationEndDialog(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL not set")
	}

	ctx := context.Background()
	pool, err := db.Open(ctx, databaseURL)
	if err != nil {
		t.Fatalf("Open(): %v", err)
	}
	t.Cleanup(pool.Close)

	emitter := events.NewEmitter(nil)
	users := db.NewUsersRepo(pool, emitter)
	dialogs := db.NewDialogsRepo(pool)
	matchSvc := match.NewService(pool, users, dialogs, nil, emitter, nil, nil)
	dialogSvc := dialog.NewService(pool, dialogs, users, emitter, nil, nil, nil, nil)

	telegramID := time.Now().UnixNano()
	up, err := users.Register(ctx, db.RegisterInput{
		TelegramID: telegramID,
		Age:        22,
		Gender:     shared.GenderMale,
		Seeking:    shared.GenderFemale,
		Language:   shared.LanguageRU,
	})
	if err != nil {
		t.Fatalf("Register(): %v", err)
	}

	searching, err := matchSvc.Start(ctx, telegramID)
	if err != nil {
		t.Fatalf("Start(): %v", err)
	}
	matched, err := matchSvc.CompleteAI(ctx, telegramID, 3)
	if err != nil {
		t.Fatalf("CompleteAI(): %v", err)
	}
	_ = searching

	resp, err := dialogSvc.End(ctx, dialog.EndRequest{
		DialogID: *matched.DialogID,
		UserID:   up.User.ID,
		Reason:   "user_confirmed",
	})
	if err != nil {
		t.Fatalf("End(): %v", err)
	}
	if resp.Status != "ended" {
		t.Fatalf("status = %q", resp.Status)
	}

	active, err := users.HasActiveDialog(ctx, up.User.ID)
	if err != nil || active {
		t.Fatal("expected no active dialog after end")
	}

	resp2, err := dialogSvc.End(ctx, dialog.EndRequest{
		DialogID: *matched.DialogID,
		UserID:   up.User.ID,
		Reason:   "user_confirmed",
	})
	if err != nil {
		t.Fatalf("second End(): %v", err)
	}
	if resp2.Status != "already_ended" {
		t.Fatalf("second status = %q", resp2.Status)
	}
}

func TestIntegrationP2PRelayText(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL not set")
	}

	_, rdb := redistest.NewTestClient(t)
	ctx := context.Background()
	pool, err := db.Open(ctx, databaseURL)
	if err != nil {
		t.Fatalf("Open(): %v", err)
	}
	t.Cleanup(pool.Close)

	emitter := events.NewEmitter(nil)
	users := db.NewUsersRepo(pool, emitter)
	dialogs := db.NewDialogsRepo(pool)
	queue := matchqueue.New(rdb)
	matchSvc := match.NewService(pool, users, dialogs, queue, emitter, nil, nil)
	dialogSvc := dialog.NewService(pool, dialogs, users, emitter, nil, nil, ratelimit.New(rdb), nil)

	tgA := time.Now().UnixNano() + 500
	tgB := tgA + 1
	upA, err := users.Register(ctx, db.RegisterInput{
		TelegramID: tgA,
		Age:        25,
		Gender:     shared.GenderMale,
		Seeking:    shared.GenderMale,
		Language:   shared.LanguageRU,
	})
	if err != nil {
		t.Fatalf("register A: %v", err)
	}
	_, err = users.Register(ctx, db.RegisterInput{
		TelegramID: tgB,
		Age:        26,
		Gender:     shared.GenderMale,
		Seeking:    shared.GenderMale,
		Language:   shared.LanguageRU,
	})
	if err != nil {
		t.Fatalf("register B: %v", err)
	}

	if _, err := matchSvc.Start(ctx, tgA); err != nil {
		t.Fatalf("Start A: %v", err)
	}
	matched, err := matchSvc.Start(ctx, tgB)
	if err != nil {
		t.Fatalf("Start B: %v", err)
	}
	if matched.DialogID == nil {
		t.Fatal("expected dialog id")
	}

	dialogA, ok, err := dialogs.GetActiveByUserID(ctx, upA.User.ID)
	if err != nil || !ok {
		t.Fatalf("active dialog A: ok=%v err=%v", ok, err)
	}

	resp, err := dialogSvc.Relay(ctx, dialog.RelayRequest{
		DialogID: dialogA.ID,
		UserID:   upA.User.ID,
		Kind:     dialog.RelayKindText,
		Text:     "hello partner",
	})
	if err != nil {
		t.Fatalf("Relay(): %v", err)
	}
	if resp.PartnerTelegramID != tgB || resp.Text != "hello partner" {
		t.Fatalf("relay resp = %+v", resp)
	}

	count, err := dialogs.MessageCount(ctx, pool, dialogA.ID)
	if err != nil || count != 1 {
		t.Fatalf("message count = %d err=%v", count, err)
	}
}
