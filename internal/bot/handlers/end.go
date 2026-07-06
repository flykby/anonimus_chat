package handlers

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/flykby/anonimus_chat/internal/bot/apiclient"
	"github.com/flykby/anonimus_chat/internal/bot/menu"
)

func (a *App) handleDialogMessage(ctx context.Context, b *bot.Bot, update *models.Update, profile apiclient.Profile, labels menu.Labels) {
	action, _ := menu.ActionForText(update.Message.Text)
	if action == menu.ActionEndDialog {
		a.promptEndDialogConfirm(ctx, b, update.Message.Chat.ID, labels)
		return
	}

	if update.Message.Text == "" {
		return
	}

	a.sendReply(ctx, b, update.Message.Chat.ID, update.Message.Text, menu.DialogKeyboard(labels))
}

func (a *App) promptEndDialogConfirm(ctx context.Context, b *bot.Bot, chatID int64, labels menu.Labels) {
	a.sendInline(ctx, b, chatID, labels.EndDialogConfirm, menu.EndConfirmButtons(labels))
}

func (a *App) onEndCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
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

	profile, ok, err := a.API.GetByTelegramID(ctx, telegramID)
	if err != nil || !ok {
		a.promptRegistration(ctx, b, msg.Chat.ID)
		return
	}
	labels := menu.LabelsFor(menu.ParseLanguage(profile.Language))

	switch data {
	case menu.CBEndCancel:
		a.sendReply(ctx, b, msg.Chat.ID, labels.EndDialogCancelled, menu.DialogKeyboard(labels))
	case menu.CBEndConfirm:
		a.confirmEndDialog(ctx, b, msg.Chat.ID, telegramID, profile, labels)
	}
}

func (a *App) confirmEndDialog(ctx context.Context, b *bot.Bot, chatID, telegramID int64, profile apiclient.Profile, labels menu.Labels) {
	if profile.ActiveDialogID == nil {
		a.showMainMenu(ctx, b, chatID, telegramID, profile)
		return
	}

	resp, err := a.API.EndDialog(ctx, *profile.ActiveDialogID, telegramID, "user_confirmed")
	if err != nil {
		a.Logger.Error("end dialog failed", "err", err, "user_id", telegramID, "dialog_id", *profile.ActiveDialogID)
		a.sendReply(ctx, b, chatID, labels.StartChatError, menu.DialogKeyboard(labels))
		return
	}

	a.showNavScreen(ctx, b, chatID, telegramID, []NavOutgoing{{
		Text:     labels.EndDialogEnded,
		Keyboard: menu.MainKeyboard(labels),
	}})

	if resp.PartnerTelegramID != nil {
		partnerLabels := labels
		if resp.PartnerLanguage != nil {
			partnerLabels = menu.LabelsFor(menu.ParseLanguage(*resp.PartnerLanguage))
		}
		a.sendReply(ctx, b, *resp.PartnerTelegramID, partnerLabels.PartnerEndedDialog, menu.MainKeyboard(partnerLabels))
	}
}
