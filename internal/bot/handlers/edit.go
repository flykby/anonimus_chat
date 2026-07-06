package handlers

import (
	"context"
	"errors"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/flykby/anonimus_chat/internal/bot/apiclient"
	"github.com/flykby/anonimus_chat/internal/bot/edit"
	"github.com/flykby/anonimus_chat/internal/bot/menu"
	"github.com/flykby/anonimus_chat/internal/bot/registration"
	"github.com/flykby/anonimus_chat/internal/shared"
)

func (a *App) sendEditMenu(ctx context.Context, b *bot.Bot, chatID, telegramID int64, lang shared.Language) {
	_ = a.FSM.Delete(ctx, telegramID)
	a.showNavScreen(ctx, b, chatID, telegramID, []NavOutgoing{{
		Text: edit.MenuTitle(lang),
		Keyboard: models.InlineKeyboardMarkup{
			InlineKeyboard: edit.MenuButtons(lang),
		},
	}})
}

func (a *App) onEditCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.CallbackQuery == nil || update.CallbackQuery.Message.Message == nil {
		return
	}

	telegramID := update.CallbackQuery.From.ID
	data := update.CallbackQuery.Data
	msg := update.CallbackQuery.Message.Message

	_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
	})

	profile, ok, err := a.API.GetByTelegramID(ctx, telegramID)
	if err != nil || !ok {
		a.promptRegistration(ctx, b, msg.Chat.ID)
		return
	}
	lang := menu.ParseLanguage(profile.Language)
	labels := menu.LabelsFor(lang)

	switch data {
	case edit.CBBack:
		a.sendProfileView(ctx, b, msg.Chat.ID, telegramID, lang)
	case edit.CBAge:
		if err := a.FSM.Set(ctx, telegramID, edit.StateAge); err != nil {
			a.Logger.Error("fsm set failed", "err", err, "user_id", telegramID)
			return
		}
		a.showNavScreen(ctx, b, msg.Chat.ID, telegramID, []NavOutgoing{{
			Text: registration.AgePrompt(lang),
		}})
	case edit.CBGender:
		if profile.ActiveDialog {
			a.showNavScreen(ctx, b, msg.Chat.ID, telegramID, []NavOutgoing{{
				Text:     edit.ActiveDialog(lang),
				Keyboard: menu.MainKeyboard(labels),
			}})
			return
		}
		if err := a.FSM.Set(ctx, telegramID, edit.StateGender); err != nil {
			a.Logger.Error("fsm set failed", "err", err, "user_id", telegramID)
			return
		}
		a.showNavScreen(ctx, b, msg.Chat.ID, telegramID, []NavOutgoing{{
			Text: registration.GenderPrompt(lang),
			Keyboard: models.InlineKeyboardMarkup{
				InlineKeyboard: edit.GenderButtons(lang),
			},
		}})
	case edit.CBSeeking:
		if profile.ActiveDialog {
			a.showNavScreen(ctx, b, msg.Chat.ID, telegramID, []NavOutgoing{{
				Text:     edit.ActiveDialog(lang),
				Keyboard: menu.MainKeyboard(labels),
			}})
			return
		}
		if err := a.FSM.Set(ctx, telegramID, edit.StateSeeking); err != nil {
			a.Logger.Error("fsm set failed", "err", err, "user_id", telegramID)
			return
		}
		a.showNavScreen(ctx, b, msg.Chat.ID, telegramID, []NavOutgoing{{
			Text: registration.SeekingPrompt(lang),
			Keyboard: models.InlineKeyboardMarkup{
				InlineKeyboard: edit.SeekingButtons(lang),
			},
		}})
	case edit.CBGenderMale, edit.CBGenderFemale:
		a.handleEditGender(ctx, b, telegramID, msg.Chat.ID, data, lang)
	case edit.CBSeekingMale, edit.CBSeekingFemale:
		a.handleEditSeeking(ctx, b, telegramID, msg.Chat.ID, data, lang)
	}
}

func (a *App) handleEditMessage(ctx context.Context, b *bot.Bot, update *models.Update, state string) {
	telegramID := update.Message.From.ID
	chatID := update.Message.Chat.ID
	a.deleteUserMessage(ctx, b, chatID, update.Message.ID)

	profile, ok, err := a.API.GetByTelegramID(ctx, telegramID)
	if err != nil || !ok {
		a.promptRegistration(ctx, b, chatID)
		return
	}
	lang := menu.ParseLanguage(profile.Language)

	if state != edit.StateAge {
		a.sendText(ctx, b, chatID, registration.UseButtonsHint(lang))
		a.resumeEdit(ctx, b, chatID, telegramID, state, lang)
		return
	}

	text := update.Message.Text
	if registration.IsTooYoung(text) {
		a.sendText(ctx, b, chatID, registration.AgeTooYoung(lang))
		return
	}
	age, err := registration.ParseAge(text)
	if err != nil {
		a.sendText(ctx, b, chatID, registration.AgeInvalid(lang))
		return
	}

	a.saveProfileField(ctx, b, telegramID, chatID, lang, apiclient.UpdateProfileRequest{
		TelegramID: telegramID,
		Age:        &age,
	})
}

func (a *App) handleEditGender(ctx context.Context, b *bot.Bot, telegramID, chatID int64, data string, lang shared.Language) {
	if !a.requireEditState(ctx, b, chatID, telegramID, edit.StateGender, lang) {
		return
	}
	gender, ok := edit.ParseGenderCallback(data)
	if !ok {
		return
	}
	genderStr := string(gender)
	a.saveProfileField(ctx, b, telegramID, chatID, lang, apiclient.UpdateProfileRequest{
		TelegramID: telegramID,
		Gender:     &genderStr,
	})
}

func (a *App) handleEditSeeking(ctx context.Context, b *bot.Bot, telegramID, chatID int64, data string, lang shared.Language) {
	if !a.requireEditState(ctx, b, chatID, telegramID, edit.StateSeeking, lang) {
		return
	}
	seeking, ok := edit.ParseGenderCallback(data)
	if !ok {
		return
	}
	seekingStr := string(seeking)
	a.saveProfileField(ctx, b, telegramID, chatID, lang, apiclient.UpdateProfileRequest{
		TelegramID: telegramID,
		Seeking:    &seekingStr,
	})
}

func (a *App) saveProfileField(ctx context.Context, b *bot.Bot, telegramID, chatID int64, lang shared.Language, req apiclient.UpdateProfileRequest) {
	labels := menu.LabelsFor(lang)
	_, err := a.API.UpdateProfile(ctx, req)
	if errors.Is(err, apiclient.ErrActiveDialog) {
		_ = a.FSM.Delete(ctx, telegramID)
		a.showNavScreen(ctx, b, chatID, telegramID, []NavOutgoing{{
			Text:     edit.ActiveDialog(lang),
			Keyboard: menu.MainKeyboard(labels),
		}})
		return
	}
	if err != nil {
		a.Logger.Error("update profile failed", "err", err, "user_id", telegramID)
		a.sendText(ctx, b, chatID, edit.SaveError(lang))
		return
	}

	_ = a.FSM.Delete(ctx, telegramID)
	a.sendProfileView(ctx, b, chatID, telegramID, lang)
}

func (a *App) requireEditState(ctx context.Context, b *bot.Bot, chatID, telegramID int64, want string, lang shared.Language) bool {
	state, ok, err := a.FSM.Get(ctx, telegramID)
	if err != nil {
		a.Logger.Error("fsm get failed", "err", err, "user_id", telegramID)
		return false
	}
	if !ok || state != want {
		a.sendText(ctx, b, chatID, registration.UseButtonsHint(lang))
		if ok {
			a.resumeEdit(ctx, b, chatID, telegramID, state, lang)
		}
		return false
	}
	return true
}

func (a *App) resumeEdit(ctx context.Context, b *bot.Bot, chatID, telegramID int64, state string, lang shared.Language) {
	switch state {
	case edit.StateAge:
		a.showNavScreen(ctx, b, chatID, telegramID, []NavOutgoing{{
			Text: registration.AgePrompt(lang),
		}})
	case edit.StateGender:
		a.showNavScreen(ctx, b, chatID, telegramID, []NavOutgoing{{
			Text: registration.GenderPrompt(lang),
			Keyboard: models.InlineKeyboardMarkup{
				InlineKeyboard: edit.GenderButtons(lang),
			},
		}})
	case edit.StateSeeking:
		a.showNavScreen(ctx, b, chatID, telegramID, []NavOutgoing{{
			Text: registration.SeekingPrompt(lang),
			Keyboard: models.InlineKeyboardMarkup{
				InlineKeyboard: edit.SeekingButtons(lang),
			},
		}})
	default:
		a.sendEditMenu(ctx, b, chatID, telegramID, lang)
	}
}
