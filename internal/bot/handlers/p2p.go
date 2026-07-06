package handlers

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/flykby/anonimus_chat/internal/bot/apiclient"
	"github.com/flykby/anonimus_chat/internal/bot/menu"
)

func (a *App) handleP2PUpdate(ctx context.Context, b *bot.Bot, update *models.Update, profile apiclient.Profile, labels menu.Labels) {
	msg := update.Message
	if msg == nil || profile.ActiveDialogID == nil {
		return
	}

	if action, _ := menu.ActionForText(msg.Text); action == menu.ActionEndDialog {
		a.promptEndDialogConfirm(ctx, b, msg.Chat.ID, labels)
		return
	}

	if isDisallowedP2PMessage(msg) {
		a.sendReply(ctx, b, msg.Chat.ID, labels.P2PDisallowedMsg, menu.DialogKeyboard(labels))
		return
	}

	switch {
	case msg.Text != "":
		a.relayP2PText(ctx, b, msg.Chat.ID, update.Message.From.ID, *profile.ActiveDialogID, msg.Text, labels)
	case len(msg.Photo) > 0:
		a.relayP2PPhoto(ctx, b, msg.Chat.ID, update.Message.From.ID, *profile.ActiveDialogID, largestPhotoFileID(msg.Photo), labels)
	case msg.Sticker != nil:
		a.relayP2PSticker(ctx, b, msg.Chat.ID, update.Message.From.ID, *profile.ActiveDialogID, msg.Sticker.FileID, labels)
	}
}

func (a *App) relayP2PText(ctx context.Context, b *bot.Bot, chatID, telegramID, dialogID int64, text string, labels menu.Labels) {
	resp, err := a.API.RelayDialog(ctx, dialogID, telegramID, "text", text, "")
	if errors.Is(err, apiclient.ErrRateLimited) {
		a.sendReply(ctx, b, chatID, labels.P2PRateLimited, menu.DialogKeyboard(labels))
		return
	}
	if err != nil {
		a.Logger.Warn("p2p relay text failed", "err", err, "user_id", telegramID, "dialog_id", dialogID)
		a.sendReply(ctx, b, chatID, labels.P2PRelayError, menu.DialogKeyboard(labels))
		return
	}
	a.deliverRelay(ctx, b, resp)
}

func (a *App) relayP2PPhoto(ctx context.Context, b *bot.Bot, chatID, telegramID, dialogID int64, fileID string, labels menu.Labels) {
	resp, err := a.API.RelayDialog(ctx, dialogID, telegramID, "photo", "", fileID)
	if errors.Is(err, apiclient.ErrRateLimited) {
		a.sendReply(ctx, b, chatID, labels.P2PRateLimited, menu.DialogKeyboard(labels))
		return
	}
	if errors.Is(err, apiclient.ErrPhotoLimit) {
		a.sendReply(ctx, b, chatID, labels.P2PPhotoLimit, menu.DialogKeyboard(labels))
		return
	}
	if err != nil {
		a.Logger.Warn("p2p relay photo failed", "err", err, "user_id", telegramID, "dialog_id", dialogID)
		a.sendReply(ctx, b, chatID, labels.P2PRelayError, menu.DialogKeyboard(labels))
		return
	}
	a.deliverRelay(ctx, b, resp)
}

func (a *App) relayP2PSticker(ctx context.Context, b *bot.Bot, chatID, telegramID, dialogID int64, fileID string, labels menu.Labels) {
	resp, err := a.API.RelayDialog(ctx, dialogID, telegramID, "sticker", "", fileID)
	if errors.Is(err, apiclient.ErrRateLimited) {
		a.sendReply(ctx, b, chatID, labels.P2PRateLimited, menu.DialogKeyboard(labels))
		return
	}
	if err != nil {
		a.Logger.Warn("p2p relay sticker failed", "err", err, "user_id", telegramID, "dialog_id", dialogID)
		a.sendReply(ctx, b, chatID, labels.P2PRelayError, menu.DialogKeyboard(labels))
		return
	}
	a.deliverRelay(ctx, b, resp)
}

func (a *App) deliverRelay(ctx context.Context, b *bot.Bot, resp apiclient.RelayResponse) {
	switch resp.Kind {
	case "text":
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: resp.PartnerTelegramID,
			Text:   resp.Text,
		})
		if err != nil {
			a.Logger.Warn("deliver p2p text failed", "err", err, "partner_id", resp.PartnerTelegramID)
		}
	case "photo":
		_, err := b.SendPhoto(ctx, &bot.SendPhotoParams{
			ChatID: resp.PartnerTelegramID,
			Photo:  &models.InputFileString{Data: resp.TelegramFileID},
		})
		if err != nil {
			a.Logger.Warn("deliver p2p photo failed", "err", err, "partner_id", resp.PartnerTelegramID)
		}
	case "sticker":
		_, err := b.SendSticker(ctx, &bot.SendStickerParams{
			ChatID:  resp.PartnerTelegramID,
			Sticker: &models.InputFileString{Data: resp.TelegramFileID},
		})
		if err != nil {
			a.Logger.Warn("deliver p2p sticker failed", "err", err, "partner_id", resp.PartnerTelegramID)
		}
	}
}

func (a *App) sendP2PModerationHint(ctx context.Context, b *bot.Bot, chatID int64, labels menu.Labels) {
	a.sendInline(ctx, b, chatID, labels.P2PModerationHint, menu.P2PModerationButtons(labels))
}

func (a *App) onP2PCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
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
	if err != nil || !ok || profile.ActiveDialogID == nil {
		return
	}
	labels := menu.LabelsFor(menu.ParseLanguage(profile.Language))

	switch data {
	case menu.CBP2PReport:
		a.handleP2PReport(ctx, b, msg.Chat.ID, telegramID, *profile.ActiveDialogID, profile, labels)
	case menu.CBP2PBlock:
		a.handleP2PBlock(ctx, b, msg.Chat.ID, telegramID, *profile.ActiveDialogID, labels)
	}
}

func (a *App) handleP2PReport(ctx context.Context, b *bot.Bot, chatID, telegramID, dialogID int64, profile apiclient.Profile, labels menu.Labels) {
	if _, err := a.API.ReportDialog(ctx, dialogID, telegramID, "user_report"); err != nil {
		a.Logger.Error("p2p report failed", "err", err, "user_id", telegramID, "dialog_id", dialogID)
		a.sendReply(ctx, b, chatID, labels.P2PRelayError, menu.DialogKeyboard(labels))
		return
	}

	a.sendReply(ctx, b, chatID, labels.P2PReportSent, menu.DialogKeyboard(labels))
	a.notifyReportAdmin(ctx, b, profile, dialogID)
}

func (a *App) handleP2PBlock(ctx context.Context, b *bot.Bot, chatID, telegramID, dialogID int64, labels menu.Labels) {
	resp, err := a.API.BlockDialog(ctx, dialogID, telegramID)
	if err != nil {
		a.Logger.Error("p2p block failed", "err", err, "user_id", telegramID, "dialog_id", dialogID)
		a.sendReply(ctx, b, chatID, labels.P2PRelayError, menu.DialogKeyboard(labels))
		return
	}

	a.sendReply(ctx, b, chatID, labels.P2PBlocked, menu.MainKeyboard(labels))

	if resp.PartnerTelegramID != nil {
		partnerLabels := labels
		if resp.PartnerLanguage != nil {
			partnerLabels = menu.LabelsFor(menu.ParseLanguage(*resp.PartnerLanguage))
		}
		a.sendReply(ctx, b, *resp.PartnerTelegramID, partnerLabels.PartnerBlockedDialog, menu.MainKeyboard(partnerLabels))
	}
}

func (a *App) notifyReportAdmin(ctx context.Context, b *bot.Bot, profile apiclient.Profile, dialogID int64) {
	if a.ReportChatID == 0 {
		return
	}
	text := fmt.Sprintf("P2P report\nuser_id=%d\ndialog_id=%d\ngender=%s seeking=%s",
		profile.TelegramID, dialogID, profile.Gender, profile.Seeking)
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: a.ReportChatID,
		Text:   text,
	})
	if err != nil {
		a.Logger.Warn("notify report admin failed", "err", err, "report_chat_id", a.ReportChatID)
	}
}

func isDisallowedP2PMessage(msg *models.Message) bool {
	if msg.Contact != nil || msg.Location != nil || msg.Venue != nil {
		return true
	}
	if msg.ForwardOrigin != nil || msg.IsAutomaticForward {
		return true
	}
	if msg.UsersShared != nil || msg.ChatShared != nil {
		return true
	}
	return false
}

func largestPhotoFileID(photos []models.PhotoSize) string {
	if len(photos) == 0 {
		return ""
	}
	best := photos[0]
	for _, photo := range photos[1:] {
		if photo.FileSize > best.FileSize {
			best = photo
		}
	}
	return best.FileID
}
