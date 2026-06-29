package handlers

import (
	"context"
	"log/slog"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type Echo struct {
	Logger *slog.Logger
}

func (h *Echo) Register(b *bot.Bot) {
	b.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact, h.start)
}

func (h *Echo) Default(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil || update.Message.Text == "" {
		return
	}
	if strings.HasPrefix(update.Message.Text, "/") {
		return
	}

	h.logMessage(update, "echo")
	h.sendText(ctx, b, update, EchoText(update.Message.Text))
}

func (h *Echo) start(ctx context.Context, b *bot.Bot, update *models.Update) {
	h.logMessage(update, "start")
	h.sendText(ctx, b, update, StartMessage)
}

func (h *Echo) sendText(ctx context.Context, b *bot.Bot, update *models.Update, text string) {
	if update.Message == nil {
		return
	}
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   text,
	})
	if err != nil {
		h.Logger.Error("send message failed",
			"err", err,
			"update_id", update.ID,
			"user_id", update.Message.From.ID,
		)
	}
}

func (h *Echo) logMessage(update *models.Update, action string) {
	if update.Message == nil {
		return
	}
	h.Logger.Info("telegram update",
		"action", action,
		"update_id", update.ID,
		"user_id", update.Message.From.ID,
		"message_len", len(update.Message.Text),
	)
}
