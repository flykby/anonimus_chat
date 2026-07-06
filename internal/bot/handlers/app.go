package handlers

import (
	"context"
	"log/slog"
	"strings"
	"sync"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/flykby/anonimus_chat/internal/bot/apiclient"
	"github.com/flykby/anonimus_chat/internal/bot/menu"
	"github.com/flykby/anonimus_chat/internal/bot/registration"
	"github.com/flykby/anonimus_chat/internal/redis/fsm"
	"github.com/flykby/anonimus_chat/internal/redis/regdraft"
)

type App struct {
	Logger    *slog.Logger
	FSM       *fsm.Store
	Draft     *regdraft.Store
	API       *apiclient.Client
	queueWait sync.Map
}

func (a *App) Register(b *bot.Bot) {
	b.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact, a.start)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "reg:", bot.MatchTypePrefix, a.onRegCallback)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "menu:", bot.MatchTypePrefix, a.onMenuCallback)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "end:", bot.MatchTypePrefix, a.onEndCallback)
}

func (a *App) Default(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil || update.Message.Text == "" {
		return
	}
	if strings.HasPrefix(update.Message.Text, "/") {
		return
	}

	telegramID := update.Message.From.ID
	state, ok, err := a.FSM.Get(ctx, telegramID)
	if err != nil {
		a.Logger.Error("fsm get failed", "err", err, "user_id", telegramID)
		return
	}
	if ok {
		a.handleRegistrationMessage(ctx, b, update, state)
		return
	}

	profile, registered, err := a.API.GetByTelegramID(ctx, telegramID)
	if err != nil {
		a.Logger.Error("check registration failed", "err", err, "user_id", telegramID)
		a.sendText(ctx, b, update.Message.Chat.ID, "Сервис временно недоступен. Попробуй позже.")
		return
	}
	if !registered {
		a.sendWelcome(ctx, b, update.Message.Chat.ID)
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
		a.sendText(ctx, b, update.Message.Chat.ID, "Сервис временно недоступен. Попробуй позже.")
		return
	}
	if registered {
		a.showMainMenu(ctx, b, update.Message.Chat.ID, profile)
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

func (a *App) handleRegistrationMessage(ctx context.Context, b *bot.Bot, update *models.Update, state string) {
	telegramID := update.Message.From.ID
	chatID := update.Message.Chat.ID

	if state != registration.StateAge {
		a.sendText(ctx, b, chatID, registration.UseButtonsHint)
		a.resumeRegistration(ctx, b, chatID, telegramID, state)
		return
	}

	text := update.Message.Text
	if registration.IsTooYoung(text) {
		a.sendText(ctx, b, chatID, registration.AgeTooYoung)
		return
	}

	age, err := registration.ParseAge(text)
	if err != nil {
		a.sendText(ctx, b, chatID, registration.AgeInvalid)
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

	a.sendGenderQuestion(ctx, b, chatID)
}

func (a *App) beginRegistration(ctx context.Context, b *bot.Bot, telegramID, chatID int64) {
	profile, registered, err := a.API.GetByTelegramID(ctx, telegramID)
	if err != nil {
		a.Logger.Error("check registration failed", "err", err, "user_id", telegramID)
		return
	}
	if registered {
		a.showMainMenu(ctx, b, chatID, profile)
		return
	}

	_ = a.Draft.Delete(ctx, telegramID)
	if err := a.FSM.Set(ctx, telegramID, registration.StateAge); err != nil {
		a.Logger.Error("fsm set failed", "err", err, "user_id", telegramID)
		return
	}
	a.sendReply(ctx, b, chatID, registration.AgePrompt, menu.RemoveKeyboard())
}

func (a *App) handleGender(ctx context.Context, b *bot.Bot, telegramID, chatID int64, data string) {
	if !a.requireState(ctx, b, chatID, telegramID, registration.StateGender) {
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
	a.sendSeekingQuestion(ctx, b, chatID)
}

func (a *App) handleSeeking(ctx context.Context, b *bot.Bot, telegramID, chatID int64, data string) {
	if !a.requireState(ctx, b, chatID, telegramID, registration.StateSeeking) {
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
	a.sendLanguageQuestion(ctx, b, chatID)
}

func (a *App) handleLanguage(ctx context.Context, b *bot.Bot, telegramID, chatID int64, data string) {
	if !a.requireState(ctx, b, chatID, telegramID, registration.StateLanguage) {
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
	if !a.requireState(ctx, b, chatID, telegramID, registration.StateConfirm) {
		return
	}

	draft, ok, err := a.Draft.Get(ctx, telegramID)
	if err != nil || !ok {
		a.Logger.Error("load draft failed", "err", err, "user_id", telegramID, "ok", ok)
		a.sendText(ctx, b, chatID, "Не удалось прочитать анкету. Нажми /start и заполни заново.")
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
		a.sendText(ctx, b, chatID, "Не удалось сохранить анкету. Попробуй позже.")
		return
	}

	_ = a.FSM.Delete(ctx, telegramID)
	_ = a.Draft.Delete(ctx, telegramID)
	a.showMainMenu(ctx, b, chatID, profile)
}

func (a *App) restartRegistration(ctx context.Context, b *bot.Bot, telegramID, chatID int64) {
	_ = a.Draft.Delete(ctx, telegramID)
	_ = a.FSM.Delete(ctx, telegramID)
	a.beginRegistration(ctx, b, telegramID, chatID)
}

func (a *App) resumeRegistration(ctx context.Context, b *bot.Bot, chatID, telegramID int64, state string) {
	switch state {
	case registration.StateAge:
		a.sendReply(ctx, b, chatID, registration.AgePrompt, menu.RemoveKeyboard())
	case registration.StateGender:
		a.sendGenderQuestion(ctx, b, chatID)
	case registration.StateSeeking:
		a.sendSeekingQuestion(ctx, b, chatID)
	case registration.StateLanguage:
		a.sendLanguageQuestion(ctx, b, chatID)
	case registration.StateConfirm:
		a.sendConfirmation(ctx, b, telegramID, chatID)
	default:
		a.sendWelcome(ctx, b, chatID)
	}
}

func (a *App) requireState(ctx context.Context, b *bot.Bot, chatID, telegramID int64, want string) bool {
	state, ok, err := a.FSM.Get(ctx, telegramID)
	if err != nil {
		a.Logger.Error("fsm get failed", "err", err, "user_id", telegramID)
		return false
	}
	if !ok || state != want {
		a.sendText(ctx, b, chatID, registration.UseButtonsHint)
		if ok {
			a.resumeRegistration(ctx, b, chatID, telegramID, state)
		}
		return false
	}
	return true
}

func (a *App) sendWelcome(ctx context.Context, b *bot.Bot, chatID int64) {
	a.sendReply(ctx, b, chatID, registration.WelcomeText, menu.RemoveKeyboard())
	a.sendInline(ctx, b, chatID, "👇", [][]models.InlineKeyboardButton{
		{{Text: "Заполнить анкету", CallbackData: registration.CBStart}},
	})
}

func (a *App) sendGenderQuestion(ctx context.Context, b *bot.Bot, chatID int64) {
	a.sendInline(ctx, b, chatID, registration.GenderPrompt, genderButtons())
}

func (a *App) sendSeekingQuestion(ctx context.Context, b *bot.Bot, chatID int64) {
	a.sendInline(ctx, b, chatID, registration.SeekingPrompt, genderButtonsSeeking())
}

func (a *App) sendLanguageQuestion(ctx context.Context, b *bot.Bot, chatID int64) {
	a.sendInline(ctx, b, chatID, registration.LanguagePrompt, [][]models.InlineKeyboardButton{
		{
			{Text: "Русский", CallbackData: registration.CBLanguageRU},
			{Text: "English", CallbackData: registration.CBLanguageEN},
		},
	})
}

func (a *App) sendConfirmation(ctx context.Context, b *bot.Bot, telegramID, chatID int64) {
	draft, ok, err := a.Draft.Get(ctx, telegramID)
	if err != nil || !ok {
		a.Logger.Error("load draft for confirm failed", "err", err, "user_id", telegramID)
		a.sendText(ctx, b, chatID, "Не удалось прочитать анкету. Нажми /start.")
		return
	}
	a.sendInline(ctx, b, chatID, registration.ConfirmationText(draft), [][]models.InlineKeyboardButton{
		{
			{Text: "Да, всё верно", CallbackData: registration.CBConfirmYes},
			{Text: "Заполнить заново", CallbackData: registration.CBConfirmRestart},
		},
	})
}

func genderButtons() [][]models.InlineKeyboardButton {
	return [][]models.InlineKeyboardButton{
		{
			{Text: "Парень", CallbackData: registration.CBGenderMale},
			{Text: "Девушка", CallbackData: registration.CBGenderFemale},
		},
	}
}

func genderButtonsSeeking() [][]models.InlineKeyboardButton {
	return [][]models.InlineKeyboardButton{
		{
			{Text: "Парня", CallbackData: registration.CBSeekingMale},
			{Text: "Девушку", CallbackData: registration.CBSeekingFemale},
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
