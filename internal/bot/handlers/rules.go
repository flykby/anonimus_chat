package handlers

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/flykby/anonimus_chat/internal/bot/menu"
	"github.com/flykby/anonimus_chat/internal/bot/rules"
	"github.com/flykby/anonimus_chat/internal/shared"
)

func (a *App) sendRulesPage(ctx context.Context, b *bot.Bot, chatID int64, lang shared.Language, labels menu.Labels) {
	messages, err := rules.Messages(lang)
	if err != nil {
		a.Logger.Error("load rules failed", "err", err, "lang", lang)
		a.sendReply(ctx, b, chatID, labels.StartChatError, menu.MainKeyboard(labels))
		return
	}

	for i, text := range messages {
		params := &bot.SendMessageParams{
			ChatID:    chatID,
			Text:      text,
			ParseMode: models.ParseModeHTML,
		}
		if i == len(messages)-1 {
			params.ReplyMarkup = models.InlineKeyboardMarkup{
				InlineKeyboard: [][]models.InlineKeyboardButton{
					{menu.BackButton(labels)},
				},
			}
		}
		if _, err := b.SendMessage(ctx, params); err != nil {
			a.Logger.Error("send rules message failed", "err", err, "chat_id", chatID, "part", i)
			return
		}
	}
}
