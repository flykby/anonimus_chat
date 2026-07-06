package menu

import (
	"fmt"

	"github.com/go-telegram/bot/models"

	"github.com/flykby/anonimus_chat/internal/shared"
)

const (
	CBBack              = "menu:back"
	CBStartChat         = "menu:start_chat"
	CBProfile           = "menu:profile"
	CBRules             = "menu:rules"
	CBEndDialog         = "menu:end_dialog"
	CBEndConfirm        = "end:confirm"
	CBEndCancel         = "end:cancel"
	CBQueueCancel       = "menu:queue_cancel"
	CBP2PReport         = "p2p:report"
	CBP2PBlock          = "p2p:block"
	CBProfilePremium    = "menu:profile:premium"
	CBProfileEdit       = "menu:profile:edit"
	CBProfileLanguage   = "menu:profile:language"
	CBProfileDelete     = "menu:profile:delete"
)

type Action int

const (
	ActionUnknown Action = iota
	ActionStartChat
	ActionProfile
	ActionRules
	ActionEndDialog
	ActionCancelQueue
)

type Labels struct {
	MenuTitle           string
	StartChat           string
	Profile             string
	Rules               string
	EndDialog           string
	Back                string
	StartChatMsg        string
	StartChatAIMatched  string
	StartChatP2PQueued  string
	StartChatActive     string
	StartChatError      string
	QueueCancel         string
	QueueMatched        string
	QueueTimeout        string
	QueueCancelled      string
	ProfileMsg            string
	ProfileBuyPremium     string
	ProfileExtendPremium  string
	ProfileEdit           string
	ProfileChangeLanguage string
	ProfileDelete         string
	ProfilePremiumStub    string
	ProfileEditStub       string
	ProfileLanguageStub   string
	ProfileDeleteStub     string
	RulesMsg              string
	EndDialogMsg        string
	EndDialogConfirm    string
	EndDialogConfirmYes string
	EndDialogConfirmNo  string
	EndDialogCancelled  string
	EndDialogEnded      string
	PartnerEndedDialog  string
	PartnerBlockedDialog string
	P2PDisallowedMsg    string
	P2PRateLimited      string
	P2PPhotoLimit       string
	P2PRelayError       string
	P2PReport           string
	P2PBlock            string
	P2PReportSent       string
	P2PBlocked          string
	P2PModerationHint   string
	DialogActiveHint    string
}

func LabelsFor(lang shared.Language) Labels {
	switch lang {
	case shared.LanguageEN:
		return Labels{
			MenuTitle:           "Main menu",
			StartChat:           "Start chat",
			Profile:             "Profile",
			Rules:               "Rules",
			EndDialog:           "End dialog",
			Back:                "← Back",
			StartChatMsg:        "Matchmaking is coming soon.",
			StartChatAIMatched:  "Chat started. Send your first message.",
			StartChatP2PQueued:  "You are in the queue. We will notify you when a partner is found.",
			StartChatActive:     "You already have an active chat.",
			StartChatError:      "Could not start a chat. Please try again later.",
			QueueCancel:         "Cancel",
			QueueMatched:        "Partner found. Send your first message.",
			QueueTimeout:        "Still searching. Tap Cancel to leave the queue or wait a bit longer.",
			QueueCancelled:      "Search cancelled.",
			ProfileMsg:            "Profile section is coming soon.",
			ProfileBuyPremium:     "Buy premium",
			ProfileExtendPremium:  "Extend premium",
			ProfileEdit:           "Edit profile",
			ProfileChangeLanguage: "Change language",
			ProfileDelete:         "Delete profile",
			ProfilePremiumStub:    "Premium purchase is coming soon.",
			ProfileEditStub:       "Profile editing is coming soon.",
			ProfileLanguageStub:   "Language change is coming soon.",
			ProfileDeleteStub:     "Profile deletion is coming soon.",
			RulesMsg:              "1. Be respectful.\n2. Do not share personal data.\n3. No illegal content.\n\nFull rules page is coming soon.",
			EndDialogMsg:        "Ending a dialog is coming soon.",
			EndDialogConfirm:    "End this chat?",
			EndDialogConfirmYes: "End chat",
			EndDialogConfirmNo:  "Cancel",
			EndDialogCancelled:  "Staying in the chat.",
			EndDialogEnded:      "Chat ended.",
			PartnerEndedDialog:  "Your partner ended the chat.",
			PartnerBlockedDialog: "Your partner blocked you. The chat has ended.",
			P2PDisallowedMsg:    "Contacts, location, and forwards are not allowed in this chat.",
			P2PRateLimited:      "Too many messages. Please wait a minute.",
			P2PPhotoLimit:       "Photo limit reached for this chat (max 3).",
			P2PRelayError:       "Could not deliver the message. Please try again.",
			P2PReport:           "Report",
			P2PBlock:            "Block",
			P2PReportSent:       "Report sent. Moderators will review it.",
			P2PBlocked:          "User blocked. Chat ended.",
			P2PModerationHint:   "You can report or block your partner if needed.",
			DialogActiveHint:    "Chat is active. Tap End dialog to leave.",
		}
	default:
		return Labels{
			MenuTitle:           "Главное меню",
			StartChat:           "Начать разговор",
			Profile:             "Профиль",
			Rules:               "Правила",
			EndDialog:           "Завершить диалог",
			Back:                "← Назад",
			StartChatMsg:        "Поиск собеседника скоро будет доступен.",
			StartChatAIMatched:  "Диалог начат. Напиши первым сообщение.",
			StartChatP2PQueued:  "Ты в очереди на поиск собеседника. Подожди немного.",
			StartChatActive:     "У тебя уже есть активный диалог.",
			StartChatError:      "Не удалось начать разговор. Попробуй позже.",
			QueueCancel:         "Отмена",
			QueueMatched:        "Собеседник найден. Напиши первым сообщение.",
			QueueTimeout:        "Всё ещё ищем. Нажми «Отмена», чтобы выйти из очереди, или подожди ещё.",
			QueueCancelled:      "Поиск отменён.",
			ProfileMsg:            "Раздел профиля скоро будет доступен.",
			ProfileBuyPremium:     "Купить премиум",
			ProfileExtendPremium:  "Продлить премиум",
			ProfileEdit:           "Изменить анкету",
			ProfileChangeLanguage: "Сменить язык",
			ProfileDelete:         "Удалить профиль",
			ProfilePremiumStub:    "Покупка премиума скоро будет доступна.",
			ProfileEditStub:       "Редактирование анкеты скоро будет доступно.",
			ProfileLanguageStub:   "Смена языка скоро будет доступна.",
			ProfileDeleteStub:     "Удаление профиля скоро будет доступно.",
			RulesMsg:              "1. Будь уважителен к собеседнику.\n2. Не делись личными данными.\n3. Запрещён незаконный контент.\n\nПолная страница правил скоро появится.",
			EndDialogMsg:        "Завершение диалога скоро будет доступно.",
			EndDialogConfirm:    "Завершить диалог?",
			EndDialogConfirmYes: "Завершить",
			EndDialogConfirmNo:  "Отменить",
			EndDialogCancelled:  "Остаёмся в диалоге.",
			EndDialogEnded:      "Диалог завершён.",
			PartnerEndedDialog:  "Собеседник завершил диалог.",
			PartnerBlockedDialog: "Собеседник заблокировал тебя. Диалог завершён.",
			P2PDisallowedMsg:    "Контакты, геолокация и пересланные сообщения здесь запрещены.",
			P2PRateLimited:      "Слишком много сообщений. Подожди минуту.",
			P2PPhotoLimit:       "Лимит фото в этом диалоге исчерпан (макс. 3).",
			P2PRelayError:       "Не удалось доставить сообщение. Попробуй ещё раз.",
			P2PReport:           "Пожаловаться",
			P2PBlock:            "Заблокировать",
			P2PReportSent:       "Жалоба отправлена. Модераторы её проверят.",
			P2PBlocked:          "Пользователь заблокирован. Диалог завершён.",
			P2PModerationHint:   "При необходимости можно пожаловаться или заблокировать собеседника.",
			DialogActiveHint:    "Диалог активен. Нажми «Завершить диалог», чтобы выйти.",
		}
	}
}

func ParseLanguage(lang string) shared.Language {
	switch shared.Language(lang) {
	case shared.LanguageEN:
		return shared.LanguageEN
	default:
		return shared.LanguageRU
	}
}

func ActionForText(text string) (Action, shared.Language) {
	for _, lang := range []shared.Language{shared.LanguageRU, shared.LanguageEN} {
		labels := LabelsFor(lang)
		switch text {
		case labels.StartChat:
			return ActionStartChat, lang
		case labels.Profile:
			return ActionProfile, lang
		case labels.Rules:
			return ActionRules, lang
		case labels.EndDialog:
			return ActionEndDialog, lang
		case labels.QueueCancel:
			return ActionCancelQueue, lang
		}
	}
	return ActionUnknown, shared.LanguageRU
}

func MainKeyboard(labels Labels) models.ReplyKeyboardMarkup {
	return models.ReplyKeyboardMarkup{
		Keyboard: [][]models.KeyboardButton{
			{{Text: labels.StartChat}},
			{{Text: labels.Profile}, {Text: labels.Rules}},
		},
		ResizeKeyboard:  true,
		OneTimeKeyboard: false,
	}
}

func DialogKeyboard(labels Labels) models.ReplyKeyboardMarkup {
	return models.ReplyKeyboardMarkup{
		Keyboard: [][]models.KeyboardButton{
			{{Text: labels.EndDialog}},
		},
		ResizeKeyboard:  true,
		OneTimeKeyboard: false,
	}
}

func ProfileSummary(age int16, gender, seeking string, lang shared.Language) string {
	g := shared.Gender(gender)
	s := shared.Gender(seeking)
	if lang == shared.LanguageEN {
		return fmt.Sprintf("Age: %d\nGender: %s\nLooking for: %s", age, genderLabelEN(g), genderLabelEN(s))
	}
	return fmt.Sprintf("Возраст: %d\nПол: %s\nИщу: %s", age, genderLabelRU(g), seekingLabelRU(s))
}

func genderLabelRU(g shared.Gender) string {
	switch g {
	case shared.GenderMale:
		return "Парень"
	case shared.GenderFemale:
		return "Девушка"
	default:
		return string(g)
	}
}

func genderLabelEN(g shared.Gender) string {
	switch g {
	case shared.GenderMale:
		return "Male"
	case shared.GenderFemale:
		return "Female"
	default:
		return string(g)
	}
}

func seekingLabelRU(g shared.Gender) string {
	switch g {
	case shared.GenderMale:
		return "Парня"
	case shared.GenderFemale:
		return "Девушку"
	default:
		return string(g)
	}
}

func EndConfirmButtons(labels Labels) [][]models.InlineKeyboardButton {
	return [][]models.InlineKeyboardButton{
		{
			{Text: labels.EndDialogConfirmYes, CallbackData: CBEndConfirm},
			{Text: labels.EndDialogConfirmNo, CallbackData: CBEndCancel},
		},
	}
}

func P2PModerationButtons(labels Labels) [][]models.InlineKeyboardButton {
	return [][]models.InlineKeyboardButton{
		{
			{Text: labels.P2PReport, CallbackData: CBP2PReport},
			{Text: labels.P2PBlock, CallbackData: CBP2PBlock},
		},
	}
}

func BackButton(labels Labels) models.InlineKeyboardButton {
	return models.InlineKeyboardButton{Text: labels.Back, CallbackData: CBBack}
}

func RemoveKeyboard() models.ReplyKeyboardRemove {
	return models.ReplyKeyboardRemove{RemoveKeyboard: true}
}
