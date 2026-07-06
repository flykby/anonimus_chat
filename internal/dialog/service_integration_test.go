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
	matchSvc := match.NewService(pool, users, dialogs, nil, emitter, nil)
	dialogSvc := dialog.NewService(pool, dialogs, users, emitter, nil, nil)

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
