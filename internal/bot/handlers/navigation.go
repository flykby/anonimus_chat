package handlers

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type NavOutgoing struct {
	Text      string
	ParseMode models.ParseMode
	Keyboard  any
}

func (a *App) clearNavScreen(ctx context.Context, b *bot.Bot, chatID, telegramID int64) {
	if a.NavScreen == nil {
		return
	}
	ids, ok, err := a.NavScreen.Get(ctx, telegramID)
	if err != nil {
		a.Logger.Warn("nav screen get failed", "err", err, "user_id", telegramID)
		return
	}
	if !ok {
		return
	}
	for _, id := range ids {
		a.deleteBotMessage(ctx, b, chatID, int(id))
	}
	_ = a.NavScreen.Delete(ctx, telegramID)
}

func (a *App) showNavScreen(ctx context.Context, b *bot.Bot, chatID, telegramID int64, msgs []NavOutgoing) {
	a.clearNavScreen(ctx, b, chatID, telegramID)
	if len(msgs) == 0 {
		return
	}

	var ids []int64
	for _, msg := range msgs {
		params := &bot.SendMessageParams{
			ChatID: chatID,
			Text:   msg.Text,
		}
		if msg.ParseMode != "" {
			params.ParseMode = msg.ParseMode
		}
		if msg.Keyboard != nil {
			params.ReplyMarkup = msg.Keyboard
		}
		sent, err := b.SendMessage(ctx, params)
		if err != nil {
			a.Logger.Error("send nav screen failed", "err", err, "chat_id", chatID, "user_id", telegramID)
			return
		}
		ids = append(ids, int64(sent.ID))
	}

	if a.NavScreen != nil && len(ids) > 0 {
		if err := a.NavScreen.Set(ctx, telegramID, ids); err != nil {
			a.Logger.Warn("nav screen set failed", "err", err, "user_id", telegramID)
		}
	}
}

func (a *App) deleteBotMessage(ctx context.Context, b *bot.Bot, chatID int64, messageID int) {
	if messageID <= 0 {
		return
	}
	_, err := b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    chatID,
		MessageID: messageID,
	})
	if err != nil {
		a.Logger.Warn("delete message failed", "err", err, "chat_id", chatID, "message_id", messageID)
	}
}

func (a *App) deleteUserMessage(ctx context.Context, b *bot.Bot, chatID int64, messageID int) {
	a.deleteBotMessage(ctx, b, chatID, messageID)
}
