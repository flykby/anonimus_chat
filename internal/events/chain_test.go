package events_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/flykby/anonimus_chat/internal/db"
	"github.com/flykby/anonimus_chat/internal/events"
	"github.com/flykby/anonimus_chat/internal/shared"
)

func TestIntegrationEventChain(t *testing.T) {
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

	dialogID := int64(9001)
	userID := up.User.ID
	if err := emitter.Emit(ctx, pool, events.Input{
		UserID:   &userID,
		DialogID: &dialogID,
		Type:     events.TypeDialogStarted,
		Metadata: events.DialogStartedMeta{
			Type:       "ai",
			MatchRoute: "m_seeks_f",
		},
	}); err != nil {
		t.Fatalf("emit dialog.started: %v", err)
	}
	if err := emitter.Emit(ctx, pool, events.Input{
		UserID:   &userID,
		DialogID: &dialogID,
		Type:     events.TypeMessageSent,
		Metadata: events.MessageSentMeta{ContentLength: 12},
	}); err != nil {
		t.Fatalf("emit message.sent: %v", err)
	}
	if err := emitter.Emit(ctx, pool, events.Input{
		UserID:   &userID,
		DialogID: &dialogID,
		Type:     events.TypeDialogEnded,
		Metadata: events.DialogEndedMeta{
			Reason:       "user_confirmed",
			DurationSec:  90,
			MessageCount: 1,
		},
	}); err != nil {
		t.Fatalf("emit dialog.ended: %v", err)
	}

	var count int
	err = pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM events
		WHERE user_id = $1
		  AND event_type IN ('user.registered', 'dialog.started', 'message.sent', 'dialog.ended')
	`, userID).Scan(&count)
	if err != nil {
		t.Fatalf("count events: %v", err)
	}
	if count != 4 {
		t.Fatalf("event count = %d, want 4", count)
	}
}
