package handlers

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/flykby/anonimus_chat/internal/bot/apiclient"
	"github.com/flykby/anonimus_chat/internal/bot/menu"
	"github.com/flykby/anonimus_chat/internal/shared"
)

const (
	aiWaitMinSec      = 2
	aiWaitMaxSec      = 5
	p2pTimeoutSec     = 120
	typingInterval    = 3 * time.Second
	queuePollInterval = 5 * time.Second
)

func (a *App) handleStartChat(ctx context.Context, b *bot.Bot, chatID, telegramID int64, profile apiclient.Profile, labels menu.Labels) {
	result, err := a.API.StartMatch(ctx, telegramID)
	if errors.Is(err, apiclient.ErrActiveDialog) {
		a.sendReply(ctx, b, chatID, labels.StartChatActive, menu.DialogKeyboard(labels))
		return
	}
	if err != nil {
		a.Logger.Error("start match failed", "err", err, "user_id", telegramID)
		a.sendReply(ctx, b, chatID, labels.StartChatError, menu.MainKeyboard(labels))
		return
	}

	switch result.Status {
	case "matched":
		a.sendReply(ctx, b, chatID, labels.QueueMatched, menu.DialogKeyboard(labels))
		if result.Route == "p2p" {
			a.sendP2PModerationHint(ctx, b, chatID, labels)
		}
	case "searching", "queued":
		count := queueDisplayCount(result)
		gender := shared.Gender(profile.Gender)
		a.sendReply(ctx, b, chatID, menu.QueueWaitingText(count, gender, menu.ParseLanguage(profile.Language)), menu.QueueWaitingKeyboard(labels))
		waitCtx, cancel := context.WithCancel(context.Background())
		a.setQueueWaitCancel(telegramID, cancel)
		go a.runQueueWait(waitCtx, b, chatID, telegramID, profile, result, labels)
	default:
		a.sendReply(ctx, b, chatID, labels.StartChatError, menu.MainKeyboard(labels))
	}
}

func (a *App) handleCancelQueue(ctx context.Context, b *bot.Bot, chatID, telegramID int64, labels menu.Labels) {
	if cancel, ok := a.clearQueueWaitCancel(telegramID); ok {
		cancel()
	}
	if err := a.API.CancelMatch(ctx, telegramID); err != nil {
		a.Logger.Warn("cancel match failed", "err", err, "user_id", telegramID)
	}
	a.sendReply(ctx, b, chatID, labels.QueueCancelled, menu.MainKeyboard(labels))
}

func (a *App) runQueueWait(
	ctx context.Context,
	b *bot.Bot,
	chatID, telegramID int64,
	profile apiclient.Profile,
	start apiclient.StartMatchResponse,
	labels menu.Labels,
) {
	defer a.clearQueueWaitCancel(telegramID)

	if start.Route == "ai" {
		a.runAIQueueWait(ctx, b, chatID, telegramID, labels)
		return
	}
	a.runP2PQueueWait(ctx, b, chatID, telegramID, labels)
}

func (a *App) runAIQueueWait(ctx context.Context, b *bot.Bot, chatID, telegramID int64, labels menu.Labels) {
	waitSec := randomWaitSec(aiWaitMinSec, aiWaitMaxSec)
	deadline := time.Now().Add(time.Duration(waitSec) * time.Second)
	a.sendTypingUntil(ctx, b, chatID, deadline)

	select {
	case <-ctx.Done():
		return
	case <-time.After(time.Until(deadline)):
	}

	result, err := a.API.CompleteMatch(ctx, telegramID, waitSec)
	if errors.Is(err, context.Canceled) || ctx.Err() != nil {
		return
	}
	if err != nil {
		a.Logger.Error("complete match failed", "err", err, "user_id", telegramID)
		a.sendReply(ctx, b, chatID, labels.StartChatError, menu.MainKeyboard(labels))
		return
	}

	a.sendReply(ctx, b, chatID, labels.QueueMatched, menu.DialogKeyboard(labels))
	_ = result
}

func (a *App) runP2PQueueWait(ctx context.Context, b *bot.Bot, chatID, telegramID int64, labels menu.Labels) {
	timeoutAt := time.Now().Add(p2pTimeoutSec * time.Second)
	timedOut := false
	ticker := time.NewTicker(queuePollInterval)
	defer ticker.Stop()

	for {
		a.sendTyping(ctx, b, chatID)

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !timedOut && time.Now().After(timeoutAt) {
				timedOut = true
				a.sendReply(ctx, b, chatID, labels.QueueTimeout, menu.QueueWaitingKeyboard(labels))
			}

			result, err := a.API.PollMatch(ctx, telegramID)
			if errors.Is(err, context.Canceled) || ctx.Err() != nil {
				return
			}
			if err != nil {
				a.Logger.Warn("poll match failed", "err", err, "user_id", telegramID)
				continue
			}
			if result.Status == "matched" {
				a.sendReply(ctx, b, chatID, labels.QueueMatched, menu.DialogKeyboard(labels))
				a.sendP2PModerationHint(ctx, b, chatID, labels)
				return
			}
		}
	}
}

func (a *App) sendTypingUntil(ctx context.Context, b *bot.Bot, chatID int64, until time.Time) {
	for {
		if ctx.Err() != nil || time.Now().After(until) {
			return
		}
		a.sendTyping(ctx, b, chatID)
		select {
		case <-ctx.Done():
			return
		case <-time.After(typingInterval):
		}
	}
}

func (a *App) sendTyping(ctx context.Context, b *bot.Bot, chatID int64) {
	_, err := b.SendChatAction(ctx, &bot.SendChatActionParams{
		ChatID: chatID,
		Action: models.ChatActionTyping,
	})
	if err != nil {
		a.Logger.Warn("send typing failed", "err", err, "chat_id", chatID)
	}
}

func queueDisplayCount(result apiclient.StartMatchResponse) int64 {
	if result.DisplayCount != nil {
		return *result.DisplayCount
	}
	if result.QueueSize != nil {
		return *result.QueueSize
	}
	return 0
}

func randomWaitSec(min, max int) int {
	if max <= min {
		return min
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
	if err != nil {
		return min
	}
	return int(n.Int64()) + min
}

func (a *App) setQueueWaitCancel(telegramID int64, cancel context.CancelFunc) {
	if prev, ok := a.queueWait.LoadAndDelete(telegramID); ok {
		prev.(context.CancelFunc)()
	}
	a.queueWait.Store(telegramID, cancel)
}

func (a *App) clearQueueWaitCancel(telegramID int64) (context.CancelFunc, bool) {
	if v, ok := a.queueWait.LoadAndDelete(telegramID); ok {
		return v.(context.CancelFunc), true
	}
	return nil, false
}
