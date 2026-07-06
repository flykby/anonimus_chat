package handlers

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/flykby/anonimus_chat/internal/bot/apiclient"
	"github.com/flykby/anonimus_chat/internal/bot/language"
	"github.com/flykby/anonimus_chat/internal/bot/menu"
	"github.com/flykby/anonimus_chat/internal/shared"
)

func (a *App) sendLanguageChoice(ctx context.Context, b *bot.Bot, chatID int64, lang shared.Language) {
	a.sendInline(ctx, b, chatID, language.Prompt(lang), language.ChoiceButtons(lang))
}

func (a *App) onLangCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.CallbackQuery == nil || update.CallbackQuery.Message.Message == nil {
		return
	}

	telegramID := update.CallbackQuery.From.ID
	data := update.CallbackQuery.Data
	msg := update.CallbackQuery.Message.Message

	_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
	})
	a.clearInlineKeyboard(ctx, b, msg.Chat.ID, msg.ID)

	chosen, ok := language.ParseCallback(data)
	if !ok {
		return
	}

	profile, registered, err := a.API.GetByTelegramID(ctx, telegramID)
	if err != nil || !registered {
		a.promptRegistration(ctx, b, msg.Chat.ID)
		return
	}

	currentLang := menu.ParseLanguage(profile.Language)
	if chosen == currentLang {
		a.sendProfileView(ctx, b, msg.Chat.ID, telegramID, currentLang)
		return
	}

	langStr := string(chosen)
	_, err = a.API.UpdateProfile(ctx, apiclient.UpdateProfileRequest{
		TelegramID: telegramID,
		Language:   &langStr,
	})
	if err != nil {
		a.Logger.Error("update language failed", "err", err, "user_id", telegramID)
		a.sendText(ctx, b, msg.Chat.ID, language.SaveError(currentLang))
		return
	}

	a.sendText(ctx, b, msg.Chat.ID, language.Changed(chosen))
	a.sendProfileView(ctx, b, msg.Chat.ID, telegramID, chosen)
}
