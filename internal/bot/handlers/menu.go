package handlers

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/flykby/anonimus_chat/internal/bot/apiclient"
	"github.com/flykby/anonimus_chat/internal/bot/menu"
)

func (a *App) handleRegisteredMessage(ctx context.Context, b *bot.Bot, update *models.Update, profile apiclient.Profile) {
	lang := menu.ParseLanguage(profile.Language)
	labels := menu.LabelsFor(lang)

	if profile.ActiveDialog {
		a.handleDialogMessage(ctx, b, update, labels)
		return
	}

	action, _ := menu.ActionForText(update.Message.Text)
	switch action {
	case menu.ActionStartChat:
		a.handleStartChat(ctx, b, update.Message.Chat.ID, update.Message.From.ID, profile, labels)
	case menu.ActionCancelQueue:
		a.handleCancelQueue(ctx, b, update.Message.Chat.ID, update.Message.From.ID, labels)
	case menu.ActionProfile:
		a.sendProfileStub(ctx, b, update.Message.Chat.ID, profile, labels)
	case menu.ActionRules:
		a.sendRulesStub(ctx, b, update.Message.Chat.ID, labels)
	case menu.ActionEndDialog:
		a.sendReply(ctx, b, update.Message.Chat.ID, labels.MenuTitle, menu.MainKeyboard(labels))
	default:
		a.showMainMenu(ctx, b, update.Message.Chat.ID, profile)
	}
}

func (a *App) handleDialogMessage(ctx context.Context, b *bot.Bot, update *models.Update, labels menu.Labels) {
	action, _ := menu.ActionForText(update.Message.Text)
	if action == menu.ActionEndDialog {
		a.sendReply(ctx, b, update.Message.Chat.ID, labels.EndDialogMsg, menu.DialogKeyboard(labels))
		return
	}
	a.sendReply(ctx, b, update.Message.Chat.ID, labels.EndDialogMsg, menu.DialogKeyboard(labels))
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
	a.clearInlineKeyboard(ctx, b, msg.Chat.ID, msg.ID)

	if data != menu.CBBack {
		return
	}

	profile, ok, err := a.API.GetByTelegramID(ctx, telegramID)
	if err != nil || !ok {
		a.promptRegistration(ctx, b, msg.Chat.ID)
		return
	}
	a.showMainMenu(ctx, b, msg.Chat.ID, profile)
}

func (a *App) showMainMenu(ctx context.Context, b *bot.Bot, chatID int64, profile apiclient.Profile) {
	lang := menu.ParseLanguage(profile.Language)
	labels := menu.LabelsFor(lang)

	if profile.ActiveDialog {
		a.sendReply(ctx, b, chatID, labels.EndDialogMsg, menu.DialogKeyboard(labels))
		return
	}

	a.sendReply(ctx, b, chatID, labels.MenuTitle, menu.MainKeyboard(labels))
}

func (a *App) sendProfileStub(ctx context.Context, b *bot.Bot, chatID int64, profile apiclient.Profile, labels menu.Labels) {
	lang := menu.ParseLanguage(profile.Language)
	text := labels.ProfileMsg + "\n\n" + menu.ProfileSummary(profile.Age, profile.Gender, profile.Seeking, lang)
	a.sendInline(ctx, b, chatID, text, [][]models.InlineKeyboardButton{
		{menu.BackButton(labels)},
	})
}

func (a *App) sendRulesStub(ctx context.Context, b *bot.Bot, chatID int64, labels menu.Labels) {
	a.sendInline(ctx, b, chatID, labels.RulesMsg, [][]models.InlineKeyboardButton{
		{menu.BackButton(labels)},
	})
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
