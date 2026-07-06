package handlers

import (
	"context"
	"log/slog"
	"strings"
	"sync"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/flykby/anonimus_chat/internal/bot/apiclient"
	"github.com/flykby/anonimus_chat/internal/bot/edit"
	"github.com/flykby/anonimus_chat/internal/bot/locales"
	"github.com/flykby/anonimus_chat/internal/bot/menu"
	"github.com/flykby/anonimus_chat/internal/bot/registration"
	"github.com/flykby/anonimus_chat/internal/redis/fsm"
	"github.com/flykby/anonimus_chat/internal/redis/navscreen"
	"github.com/flykby/anonimus_chat/internal/redis/regdraft"
	"github.com/flykby/anonimus_chat/internal/shared"
)

type App struct {
	Logger       *slog.Logger
	FSM          *fsm.Store
	Draft        *regdraft.Store
	NavScreen    *navscreen.Store
	API          *apiclient.Client
	ReportChatID int64
	queueWait    sync.Map
}

func (a *App) Register(b *bot.Bot) {
	b.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact, a.start)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "reg:", bot.MatchTypePrefix, a.onRegCallback)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "edit:", bot.MatchTypePrefix, a.onEditCallback)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "lang:", bot.MatchTypePrefix, a.onLangCallback)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "delete:", bot.MatchTypePrefix, a.onDeleteCallback)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "menu:", bot.MatchTypePrefix, a.onMenuCallback)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "end:", bot.MatchTypePrefix, a.onEndCallback)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "p2p:", bot.MatchTypePrefix, a.onP2PCallback)
}

func (a *App) Default(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}
	if update.Message.Text != "" && strings.HasPrefix(update.Message.Text, "/") {
		return
	}

	telegramID := update.Message.From.ID
	state, ok, err := a.FSM.Get(ctx, telegramID)
	if err != nil {
		a.Logger.Error("fsm get failed", "err", err, "user_id", telegramID)
		return
	}
	if ok {
		if update.Message.Text == "" {
			return
		}
		if edit.IsEditState(state) {
			a.handleEditMessage(ctx, b, update, state)
			return
		}
		a.handleRegistrationMessage(ctx, b, update, state)
		return
	}

	profile, registered, err := a.API.GetByTelegramID(ctx, telegramID)
	if err != nil {
		a.Logger.Error("check registration failed", "err", err, "user_id", telegramID)
		a.sendText(ctx, b, update.Message.Chat.ID, locales.T("common.service_unavailable", shared.LanguageRU, nil))
		return
	}
	if !registered {
		if update.Message.Text == "" {
			return
		}
		a.sendWelcome(ctx, b, update.Message.Chat.ID)
		return
	}

	if profile.ActiveDialog && profile.ActiveDialogType != nil && *profile.ActiveDialogType == "p2p" {
		labels := menu.LabelsFor(menu.ParseLanguage(profile.Language))
		a.handleP2PUpdate(ctx, b, update, profile, labels)
		return
	}

	if update.Message.Text == "" {
		return
	}

	a.handleRegisteredMessage(ctx, b, update, profile)
}

func (a *App) start(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}
	telegramID := update.Message.From.ID
	a.Logger.Info("telegram update", "action", "start", "user_id", telegramID)

	profile, registered, err := a.API.GetByTelegramID(ctx, telegramID)
	if err != nil {
		a.Logger.Error("check registration failed", "err", err, "user_id", telegramID)
		a.sendText(ctx, b, update.Message.Chat.ID, locales.T("common.service_unavailable", shared.LanguageRU, nil))
		return
	}
	if registered {
		a.showMainMenu(ctx, b, update.Message.Chat.ID, update.Message.From.ID, profile)
		return
	}

	state, ok, err := a.FSM.Get(ctx, telegramID)
	if err != nil {
		a.Logger.Error("fsm get failed", "err", err, "user_id", telegramID)
		return
	}
	if ok {
		a.resumeRegistration(ctx, b, update.Message.Chat.ID, telegramID, state)
		return
	}

	a.sendWelcome(ctx, b, update.Message.Chat.ID)
}

func (a *App) onRegCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.CallbackQuery == nil {
		return
	}

	telegramID := update.CallbackQuery.From.ID
	data := update.CallbackQuery.Data
	a.Logger.Info("telegram update", "action", "callback", "user_id", telegramID, "data", data)

	_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
	})

	if update.CallbackQuery.Message.Message == nil {
		return
	}
	msg := update.CallbackQuery.Message.Message
	a.clearInlineKeyboard(ctx, b, msg.Chat.ID, msg.ID)

	switch data {
	case registration.CBStart:
		a.beginRegistration(ctx, b, telegramID, msg.Chat.ID)
	case registration.CBGenderMale, registration.CBGenderFemale:
		a.handleGender(ctx, b, telegramID, msg.Chat.ID, data)
	case registration.CBSeekingMale, registration.CBSeekingFemale:
		a.handleSeeking(ctx, b, telegramID, msg.Chat.ID, data)
	case registration.CBLanguageRU, registration.CBLanguageEN:
		a.handleLanguage(ctx, b, telegramID, msg.Chat.ID, data)
	case registration.CBConfirmYes:
		a.confirmRegistration(ctx, b, telegramID, msg.Chat.ID)
	case registration.CBConfirmRestart:
		a.restartRegistration(ctx, b, telegramID, msg.Chat.ID)
	}
}

func (a *App) registrationLang(ctx context.Context, telegramID int64) shared.Language {
	draft, ok, err := a.Draft.Get(ctx, telegramID)
	if err != nil || !ok {
		return shared.LanguageRU
	}
	return registration.RegLanguage(draft, ok)
}

func (a *App) handleRegistrationMessage(ctx context.Context, b *bot.Bot, update *models.Update, state string) {
	telegramID := update.Message.From.ID
	chatID := update.Message.Chat.ID
	lang := a.registrationLang(ctx, telegramID)

	if state != registration.StateAge {
		a.sendText(ctx, b, chatID, registration.UseButtonsHint(lang))
		a.resumeRegistration(ctx, b, chatID, telegramID, state)
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

	if err := a.Draft.SetAge(ctx, telegramID, age); err != nil {
		a.Logger.Error("save age failed", "err", err, "user_id", telegramID)
		return
	}
	if err := a.FSM.Set(ctx, telegramID, registration.StateGender); err != nil {
		a.Logger.Error("fsm set failed", "err", err, "user_id", telegramID)
		return
	}

	a.sendGenderQuestion(ctx, b, chatID, lang)
}

func (a *App) beginRegistration(ctx context.Context, b *bot.Bot, telegramID, chatID int64) {
	profile, registered, err := a.API.GetByTelegramID(ctx, telegramID)
	if err != nil {
		a.Logger.Error("check registration failed", "err", err, "user_id", telegramID)
		return
	}
	if registered {
		a.showMainMenu(ctx, b, chatID, telegramID, profile)
		return
	}

	_ = a.Draft.Delete(ctx, telegramID)
	if err := a.FSM.Set(ctx, telegramID, registration.StateAge); err != nil {
		a.Logger.Error("fsm set failed", "err", err, "user_id", telegramID)
		return
	}
	a.sendReply(ctx, b, chatID, registration.AgePrompt(a.registrationLang(ctx, telegramID)), menu.RemoveKeyboard())
}

func (a *App) handleGender(ctx context.Context, b *bot.Bot, telegramID, chatID int64, data string) {
	lang := a.registrationLang(ctx, telegramID)
	if !a.requireState(ctx, b, chatID, telegramID, registration.StateGender, lang) {
		return
	}
	gender, ok := registration.ParseGenderCallback(data)
	if !ok {
		return
	}
	if err := a.Draft.SetGender(ctx, telegramID, gender); err != nil {
		a.Logger.Error("save gender failed", "err", err, "user_id", telegramID)
		return
	}
	if err := a.FSM.Set(ctx, telegramID, registration.StateSeeking); err != nil {
		a.Logger.Error("fsm set failed", "err", err, "user_id", telegramID)
		return
	}
	a.sendSeekingQuestion(ctx, b, chatID, lang)
}

func (a *App) handleSeeking(ctx context.Context, b *bot.Bot, telegramID, chatID int64, data string) {
	lang := a.registrationLang(ctx, telegramID)
	if !a.requireState(ctx, b, chatID, telegramID, registration.StateSeeking, lang) {
		return
	}
	seeking, ok := registration.ParseGenderCallback(data)
	if !ok {
		return
	}
	if err := a.Draft.SetSeeking(ctx, telegramID, seeking); err != nil {
		a.Logger.Error("save seeking failed", "err", err, "user_id", telegramID)
		return
	}
	if err := a.FSM.Set(ctx, telegramID, registration.StateLanguage); err != nil {
		a.Logger.Error("fsm set failed", "err", err, "user_id", telegramID)
		return
	}
	a.sendLanguageQuestion(ctx, b, chatID, lang)
}

func (a *App) handleLanguage(ctx context.Context, b *bot.Bot, telegramID, chatID int64, data string) {
	lang := a.registrationLang(ctx, telegramID)
	if !a.requireState(ctx, b, chatID, telegramID, registration.StateLanguage, lang) {
		return
	}
	language, ok := registration.ParseLanguageCallback(data)
	if !ok {
		return
	}
	if err := a.Draft.SetLanguage(ctx, telegramID, language); err != nil {
		a.Logger.Error("save language failed", "err", err, "user_id", telegramID)
		return
	}
	if err := a.FSM.Set(ctx, telegramID, registration.StateConfirm); err != nil {
		a.Logger.Error("fsm set failed", "err", err, "user_id", telegramID)
		return
	}
	a.sendConfirmation(ctx, b, telegramID, chatID)
}

func (a *App) confirmRegistration(ctx context.Context, b *bot.Bot, telegramID, chatID int64) {
	lang := a.registrationLang(ctx, telegramID)
	if !a.requireState(ctx, b, chatID, telegramID, registration.StateConfirm, lang) {
		return
	}

	draft, ok, err := a.Draft.Get(ctx, telegramID)
	lang = registration.RegLanguage(draft, ok)
	if err != nil || !ok {
		a.Logger.Error("load draft failed", "err", err, "user_id", telegramID, "ok", ok)
		a.sendText(ctx, b, chatID, registration.LoadDraftError(lang))
		return
	}

	profile, err := a.API.Register(ctx, apiclient.RegisterRequest{
		TelegramID: telegramID,
		Age:        draft.Age,
		Gender:     string(draft.Gender),
		Seeking:    string(draft.Seeking),
		Language:   string(draft.Language),
	})
	if err != nil {
		a.Logger.Error("register failed", "err", err, "user_id", telegramID)
		a.sendText(ctx, b, chatID, registration.SaveProfileError(lang))
		return
	}

	_ = a.FSM.Delete(ctx, telegramID)
	_ = a.Draft.Delete(ctx, telegramID)
	a.showMainMenu(ctx, b, chatID, telegramID, profile)
}

func (a *App) restartRegistration(ctx context.Context, b *bot.Bot, telegramID, chatID int64) {
	_ = a.Draft.Delete(ctx, telegramID)
	_ = a.FSM.Delete(ctx, telegramID)
	a.beginRegistration(ctx, b, telegramID, chatID)
}

func (a *App) resumeRegistration(ctx context.Context, b *bot.Bot, chatID, telegramID int64, state string) {
	lang := a.registrationLang(ctx, telegramID)
	switch state {
	case registration.StateAge:
		a.sendReply(ctx, b, chatID, registration.AgePrompt(lang), menu.RemoveKeyboard())
	case registration.StateGender:
		a.sendGenderQuestion(ctx, b, chatID, lang)
	case registration.StateSeeking:
		a.sendSeekingQuestion(ctx, b, chatID, lang)
	case registration.StateLanguage:
		a.sendLanguageQuestion(ctx, b, chatID, shared.LanguageRU)
	case registration.StateConfirm:
		a.sendConfirmation(ctx, b, telegramID, chatID)
	default:
		a.sendWelcome(ctx, b, chatID)
	}
}

func (a *App) requireState(ctx context.Context, b *bot.Bot, chatID, telegramID int64, want string, lang shared.Language) bool {
	state, ok, err := a.FSM.Get(ctx, telegramID)
	if err != nil {
		a.Logger.Error("fsm get failed", "err", err, "user_id", telegramID)
		return false
	}
	if !ok || state != want {
		a.sendText(ctx, b, chatID, registration.UseButtonsHint(lang))
		if ok {
			a.resumeRegistration(ctx, b, chatID, telegramID, state)
		}
		return false
	}
	return true
}

func (a *App) sendWelcome(ctx context.Context, b *bot.Bot, chatID int64) {
	lang := shared.LanguageRU
	a.sendReply(ctx, b, chatID, registration.WelcomeText(lang), menu.RemoveKeyboard())
	a.sendInline(ctx, b, chatID, registration.WelcomeHint(lang), [][]models.InlineKeyboardButton{
		{{Text: registration.StartFormButton(lang), CallbackData: registration.CBStart}},
	})
}

func (a *App) sendGenderQuestion(ctx context.Context, b *bot.Bot, chatID int64, lang shared.Language) {
	a.sendInline(ctx, b, chatID, registration.GenderPrompt(lang), genderButtons(lang))
}

func (a *App) sendSeekingQuestion(ctx context.Context, b *bot.Bot, chatID int64, lang shared.Language) {
	a.sendInline(ctx, b, chatID, registration.SeekingPrompt(lang), genderButtonsSeeking(lang))
}

func (a *App) sendLanguageQuestion(ctx context.Context, b *bot.Bot, chatID int64, lang shared.Language) {
	if lang == "" {
		lang = shared.LanguageRU
	}
	a.sendInline(ctx, b, chatID, registration.LanguagePrompt(lang), [][]models.InlineKeyboardButton{
		{
			{Text: registration.LanguageButtonRU(lang), CallbackData: registration.CBLanguageRU},
			{Text: registration.LanguageButtonEN(lang), CallbackData: registration.CBLanguageEN},
		},
	})
}

func (a *App) sendConfirmation(ctx context.Context, b *bot.Bot, telegramID, chatID int64) {
	draft, ok, err := a.Draft.Get(ctx, telegramID)
	lang := registration.RegLanguage(draft, ok)
	if err != nil || !ok {
		a.Logger.Error("load draft for confirm failed", "err", err, "user_id", telegramID)
		a.sendText(ctx, b, chatID, registration.LoadDraftErrorShort(lang))
		return
	}
	a.sendInline(ctx, b, chatID, registration.ConfirmationText(draft), [][]models.InlineKeyboardButton{
		{
			{Text: registration.ConfirmYesButton(lang), CallbackData: registration.CBConfirmYes},
			{Text: registration.ConfirmRestartButton(lang), CallbackData: registration.CBConfirmRestart},
		},
	})
}

func genderButtons(lang shared.Language) [][]models.InlineKeyboardButton {
	return [][]models.InlineKeyboardButton{
		{
			{Text: registration.GenderButtonMale(lang), CallbackData: registration.CBGenderMale},
			{Text: registration.GenderButtonFemale(lang), CallbackData: registration.CBGenderFemale},
		},
	}
}

func genderButtonsSeeking(lang shared.Language) [][]models.InlineKeyboardButton {
	return [][]models.InlineKeyboardButton{
		{
			{Text: registration.SeekingButtonMale(lang), CallbackData: registration.CBSeekingMale},
			{Text: registration.SeekingButtonFemale(lang), CallbackData: registration.CBSeekingFemale},
		},
	}
}

func (a *App) sendText(ctx context.Context, b *bot.Bot, chatID int64, text string) {
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   text,
	})
	if err != nil {
		a.Logger.Error("send message failed", "err", err, "chat_id", chatID)
	}
}

func (a *App) sendInline(ctx context.Context, b *bot.Bot, chatID int64, text string, rows [][]models.InlineKeyboardButton) {
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   text,
		ReplyMarkup: models.InlineKeyboardMarkup{
			InlineKeyboard: rows,
		},
	})
	if err != nil {
		a.Logger.Error("send inline message failed", "err", err, "chat_id", chatID)
	}
}

func (a *App) clearInlineKeyboard(ctx context.Context, b *bot.Bot, chatID int64, messageID int) {
	_, err := b.EditMessageReplyMarkup(ctx, &bot.EditMessageReplyMarkupParams{
		ChatID:      chatID,
		MessageID:   messageID,
		ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: [][]models.InlineKeyboardButton{}},
	})
	if err != nil {
		a.Logger.Warn("clear inline keyboard failed", "err", err, "chat_id", chatID, "message_id", messageID)
	}
}
