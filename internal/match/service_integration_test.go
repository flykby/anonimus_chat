package match_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/flykby/anonimus_chat/internal/db"
	"github.com/flykby/anonimus_chat/internal/events"
	"github.com/flykby/anonimus_chat/internal/match"
	"github.com/flykby/anonimus_chat/internal/shared"
)

func TestIntegrationStartAIRoute(t *testing.T) {
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
	svc := match.NewService(pool, users, dialogs, nil, emitter, nil)

	telegramID := time.Now().UnixNano()
	up, err := users.Register(ctx, db.RegisterInput{
		TelegramID: telegramID,
		Age:        25,
		Gender:     shared.GenderMale,
		Seeking:    shared.GenderFemale,
		Language:   shared.LanguageRU,
	})
	if err != nil {
		t.Fatalf("Register(): %v", err)
	}

	searching, err := svc.Start(ctx, telegramID)
	if err != nil {
		t.Fatalf("Start(): %v", err)
	}
	if searching.Route != "ai" || searching.Status != "searching" || searching.DisplayCount == nil {
		t.Fatalf("unexpected searching response: %+v", searching)
	}

	matched, err := svc.CompleteAI(ctx, telegramID, 3)
	if err != nil {
		t.Fatalf("CompleteAI(): %v", err)
	}
	if matched.Status != "matched" || matched.DialogID == nil {
		t.Fatalf("unexpected matched response: %+v", matched)
	}

	active, err := users.HasActiveDialog(ctx, up.User.ID)
	if err != nil || !active {
		t.Fatalf("expected active dialog for user %d", up.User.ID)
	}
}

func TestIntegrationStartRejectsActiveDialog(t *testing.T) {
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
	svc := match.NewService(pool, users, dialogs, nil, emitter, nil)

	telegramID := time.Now().UnixNano() + 1
	up, err := users.Register(ctx, db.RegisterInput{
		TelegramID: telegramID,
		Age:        30,
		Gender:     shared.GenderFemale,
		Seeking:    shared.GenderFemale,
		Language:   shared.LanguageEN,
	})
	if err != nil {
		t.Fatalf("Register(): %v", err)
	}

	if _, err := svc.Start(ctx, telegramID); err != nil {
		t.Fatalf("first Start(): %v", err)
	}
	if _, err := svc.CompleteAI(ctx, telegramID, 2); err != nil {
		t.Fatalf("CompleteAI(): %v", err)
	}
	if _, err := svc.Start(ctx, telegramID); err != match.ErrActiveDialog {
		t.Fatalf("second Start() err = %v, want ErrActiveDialog", err)
	}

	active, err := users.HasActiveDialog(ctx, up.User.ID)
	if err != nil || !active {
		t.Fatalf("expected active dialog for user %d", up.User.ID)
	}
}
