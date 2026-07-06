package handlers

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/flykby/anonimus_chat/internal/bot/deleteprofile"
	"github.com/flykby/anonimus_chat/internal/bot/menu"
	"github.com/flykby/anonimus_chat/internal/shared"
)

func (a *App) sendDeleteConfirm1(ctx context.Context, b *bot.Bot, chatID int64, lang shared.Language) {
	a.sendInline(ctx, b, chatID, deleteprofile.Confirm1(lang), deleteprofile.Confirm1Buttons(lang))
}

func (a *App) sendDeleteConfirm2(ctx context.Context, b *bot.Bot, chatID int64, lang shared.Language) {
	a.sendInline(ctx, b, chatID, deleteprofile.Confirm2(lang), deleteprofile.Confirm2Buttons(lang))
}

func (a *App) onDeleteCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
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

	profile, ok, err := a.API.GetByTelegramID(ctx, telegramID)
	if err != nil || !ok {
		if data == deleteprofile.CBCancel {
			a.promptRegistration(ctx, b, msg.Chat.ID)
			return
		}
		a.promptRegistration(ctx, b, msg.Chat.ID)
		return
	}
	lang := menu.ParseLanguage(profile.Language)
	labels := menu.LabelsFor(lang)

	switch data {
	case deleteprofile.CBCancel:
		a.sendReply(ctx, b, msg.Chat.ID, deleteprofile.Cancelled(lang), menu.MainKeyboard(labels))
		a.sendProfileView(ctx, b, msg.Chat.ID, telegramID, lang)
	case deleteprofile.CBConfirm1:
		a.sendDeleteConfirm2(ctx, b, msg.Chat.ID, lang)
	case deleteprofile.CBConfirm2:
		a.executeDeleteProfile(ctx, b, msg.Chat.ID, telegramID, lang)
	}
}

func (a *App) executeDeleteProfile(ctx context.Context, b *bot.Bot, chatID, telegramID int64, lang shared.Language) {
	resp, err := a.API.DeleteProfile(ctx, telegramID)
	if err != nil {
		a.Logger.Error("delete profile failed", "err", err, "user_id", telegramID)
		a.sendText(ctx, b, chatID, deleteprofile.Error(lang))
		return
	}

	a.clearQueueWaitCancel(telegramID)
	_ = a.FSM.Delete(ctx, telegramID)
	_ = a.Draft.Delete(ctx, telegramID)

	a.sendReply(ctx, b, chatID, deleteprofile.Done(lang), menu.RemoveKeyboard())

	if resp.PartnerTelegramID != nil {
		partnerLabels := menu.LabelsFor(lang)
		if resp.PartnerLanguage != nil {
			partnerLabels = menu.LabelsFor(menu.ParseLanguage(*resp.PartnerLanguage))
		}
		a.sendReply(ctx, b, *resp.PartnerTelegramID, partnerLabels.PartnerEndedDialog, menu.MainKeyboard(partnerLabels))
	}
}
