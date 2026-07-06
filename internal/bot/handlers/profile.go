package handlers

import (
	"context"

	"github.com/go-telegram/bot"

	"github.com/flykby/anonimus_chat/internal/bot/menu"
	"github.com/flykby/anonimus_chat/internal/shared"
)

func (a *App) sendProfileView(ctx context.Context, b *bot.Bot, chatID, telegramID int64, lang shared.Language) {
	view, err := a.API.GetProfileView(ctx, telegramID)
	if err != nil {
		a.Logger.Error("load profile view failed", "err", err, "user_id", telegramID)
		a.sendReply(ctx, b, chatID, menu.LabelsFor(lang).StartChatError, menu.MainKeyboard(menu.LabelsFor(lang)))
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

	a.sendInline(ctx, b, chatID, text, menu.ProfileViewButtons(labels, view.PremiumActive))
}

func (a *App) handleProfileCallback(ctx context.Context, b *bot.Bot, chatID, telegramID int64, data string, labels menu.Labels, lang shared.Language) {
	var stub string
	switch data {
	case menu.CBProfilePremium:
		stub = labels.ProfilePremiumStub
	case menu.CBProfileEdit:
		a.sendEditMenu(ctx, b, chatID, telegramID, lang)
		return
	case menu.CBProfileLanguage:
		stub = labels.ProfileLanguageStub
	case menu.CBProfileDelete:
		stub = labels.ProfileDeleteStub
	default:
		return
	}
	a.sendReply(ctx, b, chatID, stub, menu.MainKeyboard(labels))
}
