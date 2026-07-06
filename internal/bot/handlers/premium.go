package handlers

import (
	"context"
	"fmt"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/flykby/anonimus_chat/internal/bot/locales"
	"github.com/flykby/anonimus_chat/internal/bot/menu"
	"github.com/flykby/anonimus_chat/internal/shared"
)

func (a *App) matchPreCheckout(update *models.Update) bool {
	return update.PreCheckoutQuery != nil
}

func (a *App) matchSuccessfulPayment(update *models.Update) bool {
	return update.Message != nil && update.Message.SuccessfulPayment != nil
}

func (a *App) onPremiumCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.CallbackQuery == nil {
		return
	}

	telegramID := update.CallbackQuery.From.ID
	data := update.CallbackQuery.Data
	a.Logger.Info("premium callback", "user_id", telegramID, "data", data)

	_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
	})

	if update.CallbackQuery.Message.Message == nil {
		return
	}
	msg := update.CallbackQuery.Message.Message
	chatID := msg.Chat.ID

	lang := a.getUserLanguage(ctx, telegramID)

	switch data {
	case menu.CBPremiumBuy:
		a.sendPremiumInvoice(ctx, b, chatID, telegramID, lang)
	}
}

func (a *App) sendPremiumMenu(ctx context.Context, b *bot.Bot, chatID, telegramID int64, lang shared.Language) {
	view, err := a.API.GetProfileView(ctx, telegramID)
	if err != nil {
		a.Logger.Error("load profile view failed", "err", err, "user_id", telegramID)
		labels := menu.LabelsFor(lang)
		a.showNavScreen(ctx, b, chatID, telegramID, []NavOutgoing{{
			Text:     labels.StartChatError,
			Keyboard: menu.MainKeyboard(labels),
		}})
		return
	}

	lang = menu.ParseLanguage(view.Language)
	labels := menu.LabelsFor(lang)

	benefitsText := locales.T("premium.benefits", lang, map[string]string{
		"price": fmt.Sprintf("%d", a.PremiumPriceStars),
		"days":  fmt.Sprintf("%d", a.PremiumDurationDays),
	})

	var buttons [][]models.InlineKeyboardButton
	if view.PremiumActive {
		expiresText := ""
		if view.PremiumExpiresAt != nil {
			expiresText = view.PremiumExpiresAt.Format("02.01.2006 15:04")
		}
		benefitsText = locales.T("premium.active_status", lang, map[string]string{
			"date":  expiresText,
			"price": fmt.Sprintf("%d", a.PremiumPriceStars),
			"days":  fmt.Sprintf("%d", a.PremiumDurationDays),
		})
		buttons = [][]models.InlineKeyboardButton{
			{{Text: labels.ProfileExtendPremium, CallbackData: menu.CBPremiumBuy}},
			{{Text: labels.Back, CallbackData: menu.CBBack}},
		}
	} else {
		buttons = [][]models.InlineKeyboardButton{
			{{Text: labels.ProfileBuyPremium, CallbackData: menu.CBPremiumBuy}},
			{{Text: labels.Back, CallbackData: menu.CBBack}},
		}
	}

	a.showNavScreen(ctx, b, chatID, telegramID, []NavOutgoing{{
		Text: benefitsText,
		Keyboard: models.InlineKeyboardMarkup{
			InlineKeyboard: buttons,
		},
	}})
}

func (a *App) sendPremiumInvoice(ctx context.Context, b *bot.Bot, chatID, telegramID int64, lang shared.Language) {
	title := locales.T("premium.invoice.title", lang, nil)
	description := locales.T("premium.invoice.description", lang, map[string]string{
		"days": fmt.Sprintf("%d", a.PremiumDurationDays),
	})

	payload := fmt.Sprintf("premium:%d", telegramID)

	_, err := b.SendInvoice(ctx, &bot.SendInvoiceParams{
		ChatID:      chatID,
		Title:       title,
		Description: description,
		Payload:     payload,
		Currency:    "XTR",
		Prices: []models.LabeledPrice{
			{
				Label:  title,
				Amount: a.PremiumPriceStars,
			},
		},
	})
	if err != nil {
		a.Logger.Error("send invoice failed", "err", err, "user_id", telegramID)
		labels := menu.LabelsFor(lang)
		a.sendText(ctx, b, chatID, locales.T("premium.invoice.error", lang, nil))
		a.showNavScreen(ctx, b, chatID, telegramID, []NavOutgoing{{
			Text:     labels.MenuTitle,
			Keyboard: menu.MainKeyboard(labels),
		}})
	}
}

func (a *App) handlePreCheckoutQuery(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.PreCheckoutQuery == nil {
		return
	}

	query := update.PreCheckoutQuery
	a.Logger.Info("pre_checkout_query", "user_id", query.From.ID, "payload", query.InvoicePayload, "amount", query.TotalAmount)

	_, err := b.AnswerPreCheckoutQuery(ctx, &bot.AnswerPreCheckoutQueryParams{
		PreCheckoutQueryID: query.ID,
		OK:                 true,
	})
	if err != nil {
		a.Logger.Error("answer pre_checkout_query failed", "err", err, "user_id", query.From.ID)
	}
}

func (a *App) handleSuccessfulPayment(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil || update.Message.SuccessfulPayment == nil {
		return
	}

	payment := update.Message.SuccessfulPayment
	telegramID := update.Message.From.ID
	chatID := update.Message.Chat.ID

	a.Logger.Info("successful_payment",
		"user_id", telegramID,
		"payload", payment.InvoicePayload,
		"amount", payment.TotalAmount,
		"telegram_charge_id", payment.TelegramPaymentChargeID,
	)

	lang := a.getUserLanguage(ctx, telegramID)
	labels := menu.LabelsFor(lang)

	result, err := a.API.PurchasePremium(ctx, telegramID, payment.TotalAmount, a.PremiumDurationDays, payment.TelegramPaymentChargeID, payment.ProviderPaymentChargeID)
	if err != nil {
		a.Logger.Error("purchase premium failed", "err", err, "user_id", telegramID, "charge_id", payment.TelegramPaymentChargeID)
		a.sendText(ctx, b, chatID, locales.T("premium.purchase.error", lang, nil))
		a.showNavScreen(ctx, b, chatID, telegramID, []NavOutgoing{{
			Text:     labels.MenuTitle,
			Keyboard: menu.MainKeyboard(labels),
		}})
		return
	}

	expiresText := result.ExpiresAt.Format("02.01.2006 15:04")
	successText := locales.T("premium.purchase.success", lang, map[string]string{
		"date": expiresText,
	})

	a.sendText(ctx, b, chatID, successText)
	a.showNavScreen(ctx, b, chatID, telegramID, []NavOutgoing{{
		Text:     labels.MenuTitle,
		Keyboard: menu.MainKeyboard(labels),
	}})
}

func (a *App) getUserLanguage(ctx context.Context, telegramID int64) shared.Language {
	profile, registered, err := a.API.GetByTelegramID(ctx, telegramID)
	if err != nil || !registered {
		return shared.LanguageRU
	}
	return menu.ParseLanguage(profile.Language)
}
