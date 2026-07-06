package handlers

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/flykby/anonimus_chat/internal/bot/menu"
	"github.com/flykby/anonimus_chat/internal/bot/rules"
	"github.com/flykby/anonimus_chat/internal/shared"
)

func (a *App) sendRulesPage(ctx context.Context, b *bot.Bot, chatID, telegramID int64, lang shared.Language, labels menu.Labels) {
	messages, err := rules.Messages(lang)
	if err != nil {
		a.Logger.Error("load rules failed", "err", err, "lang", lang)
		a.showNavScreen(ctx, b, chatID, telegramID, []NavOutgoing{{
			Text:     labels.StartChatError,
			Keyboard: menu.MainKeyboard(labels),
		}})
		return
	}

	out := make([]NavOutgoing, 0, len(messages))
	for i, text := range messages {
		msg := NavOutgoing{
			Text:      text,
			ParseMode: models.ParseModeHTML,
		}
		if i == len(messages)-1 {
			msg.Keyboard = models.InlineKeyboardMarkup{
				InlineKeyboard: [][]models.InlineKeyboardButton{
					{menu.BackButton(labels)},
				},
			}
		}
		out = append(out, msg)
	}
	a.showNavScreen(ctx, b, chatID, telegramID, out)
}
