package webhook

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func TestHandlerProcessesWithIndependentContext(t *testing.T) {
	t.Parallel()

	done := make(chan struct{})

	tg, err := bot.New("123:ABC",
		bot.WithSkipGetMe(),
		bot.WithDefaultHandler(func(ctx context.Context, b *bot.Bot, update *models.Update) {
			time.Sleep(20 * time.Millisecond)
			if ctx.Err() != nil {
				t.Errorf("handler context canceled: %v", ctx.Err())
			}
			close(done)
		}))
	if err != nil {
		t.Fatal(err)
	}

	wh := New(tg, "", slog.Default())

	req := httptest.NewRequest(http.MethodPost, "/telegram/webhook", strings.NewReader(`{"update_id":1}`))
	rec := httptest.NewRecorder()

	wh.Handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("handler did not finish")
	}
}

func TestHandlerRejectsInvalidSecret(t *testing.T) {
	t.Parallel()

	tg, err := bot.New("123:ABC", bot.WithSkipGetMe())
	if err != nil {
		t.Fatal(err)
	}

	wh := New(tg, "secret", slog.Default())

	req := httptest.NewRequest(http.MethodPost, "/telegram/webhook", strings.NewReader(`{"update_id":1}`))
	req.Header.Set(secretTokenHeader, "wrong")
	rec := httptest.NewRecorder()

	wh.Handler(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
}
