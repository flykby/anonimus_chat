package handlers

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/flykby/anonimus_chat/internal/bot/menu"
	"github.com/flykby/anonimus_chat/internal/shared"
)

func (a *App) sendProfileView(ctx context.Context, b *bot.Bot, chatID, telegramID int64, lang shared.Language) {
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
	text := menu.ProfileViewText(menu.ProfileViewData{
		PublicUUID:       view.PublicUUID,
		Age:              view.Age,
		Gender:           view.Gender,
		Seeking:          view.Seeking,
		Language:         view.Language,
		PremiumActive:    view.PremiumActive,
		PremiumExpiresAt: view.PremiumExpiresAt,
	}, lang)

	a.showNavScreen(ctx, b, chatID, telegramID, []NavOutgoing{{
		Text: text,
		Keyboard: models.InlineKeyboardMarkup{
			InlineKeyboard: menu.ProfileViewButtons(labels, view.PremiumActive),
		},
	}})
}

func (a *App) handleProfileCallback(ctx context.Context, b *bot.Bot, chatID, telegramID int64, data string, labels menu.Labels, lang shared.Language) {
	switch data {
	case menu.CBProfilePremium:
		a.sendPremiumMenu(ctx, b, chatID, telegramID, lang)
	case menu.CBProfileEdit:
		a.sendEditMenu(ctx, b, chatID, telegramID, lang)
	case menu.CBProfileLanguage:
		a.sendLanguageChoice(ctx, b, chatID, telegramID, lang)
	case menu.CBProfileDelete:
		a.sendDeleteConfirm1(ctx, b, chatID, telegramID, lang)
	default:
		return
	}
}
