package handlers

import (
	"context"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/flykby/anonimus_chat/internal/bot/apiclient"
	"github.com/flykby/anonimus_chat/internal/bot/menu"
)

func (a *App) handleRegisteredMessage(ctx context.Context, b *bot.Bot, update *models.Update, profile apiclient.Profile) {
	lang := menu.ParseLanguage(profile.Language)
	labels := menu.LabelsFor(lang)
	chatID := update.Message.Chat.ID
	telegramID := update.Message.From.ID

	if profile.ActiveDialog {
		a.handleDialogMessage(ctx, b, update, profile, labels)
		return
	}

	action, _ := menu.ActionForText(update.Message.Text)
	switch action {
	case menu.ActionStartChat:
		a.deleteUserMessage(ctx, b, chatID, update.Message.ID)
		a.handleStartChat(ctx, b, chatID, telegramID, profile, labels)
	case menu.ActionCancelQueue:
		a.deleteUserMessage(ctx, b, chatID, update.Message.ID)
		a.handleCancelQueue(ctx, b, chatID, telegramID, labels)
	case menu.ActionProfile:
		a.deleteUserMessage(ctx, b, chatID, update.Message.ID)
		a.sendProfileView(ctx, b, chatID, telegramID, lang)
	case menu.ActionRules:
		a.deleteUserMessage(ctx, b, chatID, update.Message.ID)
		a.sendRulesPage(ctx, b, chatID, telegramID, lang, labels)
	case menu.ActionEndDialog:
		a.showMainMenu(ctx, b, chatID, telegramID, profile)
	default:
		a.showMainMenu(ctx, b, chatID, telegramID, profile)
	}
}

func (a *App) onMenuCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.CallbackQuery == nil || update.CallbackQuery.Message.Message == nil {
		return
	}

	telegramID := update.CallbackQuery.From.ID
	data := update.CallbackQuery.Data

	_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
	})

	msg := update.CallbackQuery.Message.Message

	if data == menu.CBBack {
		profile, ok, err := a.API.GetByTelegramID(ctx, telegramID)
		if err != nil || !ok {
			a.promptRegistration(ctx, b, msg.Chat.ID)
			return
		}
		a.showMainMenu(ctx, b, msg.Chat.ID, telegramID, profile)
		return
	}

	if strings.HasPrefix(data, "menu:profile:") {
		profile, ok, err := a.API.GetByTelegramID(ctx, telegramID)
		if err != nil || !ok {
			a.promptRegistration(ctx, b, msg.Chat.ID)
			return
		}
		labels := menu.LabelsFor(menu.ParseLanguage(profile.Language))
		lang := menu.ParseLanguage(profile.Language)
		a.handleProfileCallback(ctx, b, msg.Chat.ID, telegramID, data, labels, lang)
	}
}

func (a *App) showMainMenu(ctx context.Context, b *bot.Bot, chatID, telegramID int64, profile apiclient.Profile) {
	lang := menu.ParseLanguage(profile.Language)
	labels := menu.LabelsFor(lang)

	if profile.ActiveDialog {
		a.showNavScreen(ctx, b, chatID, telegramID, []NavOutgoing{{
			Text:     labels.DialogActiveHint,
			Keyboard: menu.DialogKeyboard(labels),
		}})
		return
	}

	a.showNavScreen(ctx, b, chatID, telegramID, []NavOutgoing{{
		Text:     labels.MenuTitle,
		Keyboard: menu.MainKeyboard(labels),
	}})
}

func (a *App) promptRegistration(ctx context.Context, b *bot.Bot, chatID int64) {
	a.sendWelcome(ctx, b, chatID)
}

func (a *App) sendReply(ctx context.Context, b *bot.Bot, chatID int64, text string, markup any) {
	params := &bot.SendMessageParams{
		ChatID: chatID,
		Text:   text,
	}
	switch m := markup.(type) {
	case models.ReplyKeyboardMarkup:
		params.ReplyMarkup = m
	case models.ReplyKeyboardRemove:
		params.ReplyMarkup = m
	}
	_, err := b.SendMessage(ctx, params)
	if err != nil {
		a.Logger.Error("send reply message failed", "err", err, "chat_id", chatID)
	}
}
