package match_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/flykby/anonimus_chat/internal/db"
	"github.com/flykby/anonimus_chat/internal/events"
	"github.com/flykby/anonimus_chat/internal/match"
	"github.com/flykby/anonimus_chat/internal/redis/matchqueue"
	redistest "github.com/flykby/anonimus_chat/internal/redis/redistest"
	"github.com/flykby/anonimus_chat/internal/redis/session"
	"github.com/flykby/anonimus_chat/internal/shared"
)

func TestIntegrationP2PMatchMalePair(t *testing.T) {
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
	sessions := session.New(rdb)
	svc := match.NewService(pool, users, dialogs, queue, emitter, sessions, nil)

	tgA := time.Now().UnixNano()
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
	upB, err := users.Register(ctx, db.RegisterInput{
		TelegramID: tgB,
		Age:        26,
		Gender:     shared.GenderMale,
		Seeking:    shared.GenderMale,
		Language:   shared.LanguageRU,
	})
	if err != nil {
		t.Fatalf("register B: %v", err)
	}

	first, err := svc.Start(ctx, tgA)
	if err != nil {
		t.Fatalf("Start A: %v", err)
	}
	if first.Status != "queued" {
		t.Fatalf("first start status = %q", first.Status)
	}

	second, err := svc.Start(ctx, tgB)
	if err != nil {
		t.Fatalf("Start B: %v", err)
	}
	if second.Status != "matched" || second.DialogID == nil {
		t.Fatalf("second start = %+v, want matched", second)
	}

	polled, err := svc.Poll(ctx, tgA)
	if err != nil {
		t.Fatalf("Poll A: %v", err)
	}
	if polled.Status != "matched" || polled.DialogID == nil {
		t.Fatalf("poll A = %+v, want matched", polled)
	}

	for _, userID := range []int64{upA.User.ID, upB.User.ID} {
		active, err := users.HasActiveDialog(ctx, userID)
		if err != nil || !active {
			t.Fatalf("user %d expected active dialog", userID)
		}
	}

	sessA, ok, err := sessions.Get(ctx, upA.User.ID)
	if err != nil || !ok || sessA.PartnerID != upB.User.ID {
		t.Fatalf("session A = %+v ok=%v err=%v", sessA, ok, err)
	}
	sessB, ok, err := sessions.Get(ctx, upB.User.ID)
	if err != nil || !ok || sessB.PartnerID != upA.User.ID {
		t.Fatalf("session B = %+v ok=%v err=%v", sessB, ok, err)
	}
	if sessA.DialogID == sessB.DialogID {
		t.Fatalf("expected distinct dialog ids, got %d and %d", sessA.DialogID, sessB.DialogID)
	}
}

func TestIntegrationP2PMatchHeteroPair(t *testing.T) {
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
	sessions := session.New(rdb)
	svc := match.NewService(pool, users, dialogs, queue, emitter, sessions, nil)

	tgF := time.Now().UnixNano() + 100
	tgM := tgF + 1
	upF, err := users.Register(ctx, db.RegisterInput{
		TelegramID: tgF,
		Age:        24,
		Gender:     shared.GenderFemale,
		Seeking:    shared.GenderMale,
		Language:   shared.LanguageEN,
	})
	if err != nil {
		t.Fatalf("register F: %v", err)
	}
	upM, err := users.Register(ctx, db.RegisterInput{
		TelegramID: tgM,
		Age:        27,
		Gender:     shared.GenderMale,
		Seeking:    shared.GenderFemale,
		Language:   shared.LanguageEN,
	})
	if err != nil {
		t.Fatalf("register M: %v", err)
	}

	if err := queue.EnqueueHetero(ctx, shared.GenderMale, upM.User.ID); err != nil {
		t.Fatalf("enqueue hetero male: %v", err)
	}

	queued, err := svc.Start(ctx, tgF)
	if err != nil {
		t.Fatalf("Start F: %v", err)
	}
	if queued.Status != "matched" || queued.DialogID == nil {
		t.Fatalf("start F = %+v, want matched", queued)
	}

	for _, userID := range []int64{upF.User.ID, upM.User.ID} {
		active, err := users.HasActiveDialog(ctx, userID)
		if err != nil || !active {
			t.Fatalf("user %d expected active dialog", userID)
		}
	}

	sessF, ok, err := sessions.Get(ctx, upF.User.ID)
	if err != nil || !ok || sessF.PartnerID != upM.User.ID {
		t.Fatalf("session F = %+v ok=%v err=%v", sessF, ok, err)
	}
}

func TestIntegrationP2PCancelLeavesQueue(t *testing.T) {
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
	svc := match.NewService(pool, users, dialogs, queue, emitter, nil, nil)

	tg := time.Now().UnixNano() + 200
	up, err := users.Register(ctx, db.RegisterInput{
		TelegramID: tg,
		Age:        28,
		Gender:     shared.GenderFemale,
		Seeking:    shared.GenderMale,
		Language:   shared.LanguageRU,
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	if _, err := svc.Start(ctx, tg); err != nil {
		t.Fatalf("Start(): %v", err)
	}
	if err := svc.Cancel(ctx, tg); err != nil {
		t.Fatalf("Cancel(): %v", err)
	}

	inQueue, err := queue.IsInHeteroQueue(ctx, shared.GenderFemale, up.User.ID)
	if err != nil {
		t.Fatalf("IsInHeteroQueue: %v", err)
	}
	if inQueue {
		t.Fatal("expected user removed from hetero queue")
	}

	if _, err := svc.Poll(ctx, tg); err != match.ErrNotInQueue {
		t.Fatalf("Poll() err = %v, want ErrNotInQueue", err)
	}
}
