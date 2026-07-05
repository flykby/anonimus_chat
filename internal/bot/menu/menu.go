package menu

import (
	"fmt"

	"github.com/go-telegram/bot/models"

	"github.com/flykby/anonimus_chat/internal/shared"
)

const (
	CBBack      = "menu:back"
	CBStartChat = "menu:start_chat"
	CBProfile   = "menu:profile"
	CBRules     = "menu:rules"
	CBEndDialog = "menu:end_dialog"
)

type Action int

const (
	ActionUnknown Action = iota
	ActionStartChat
	ActionProfile
	ActionRules
	ActionEndDialog
)

type Labels struct {
	MenuTitle    string
	StartChat    string
	Profile      string
	Rules        string
	EndDialog    string
	Back         string
	StartChatMsg string
	ProfileMsg   string
	RulesMsg     string
	EndDialogMsg string
}

func LabelsFor(lang shared.Language) Labels {
	switch lang {
	case shared.LanguageEN:
		return Labels{
			MenuTitle:    "Main menu",
			StartChat:    "Start chat",
			Profile:      "Profile",
			Rules:        "Rules",
			EndDialog:    "End dialog",
			Back:         "← Back",
			StartChatMsg: "Matchmaking is coming soon.",
			ProfileMsg:   "Profile section is coming soon.",
			RulesMsg:     "1. Be respectful.\n2. Do not share personal data.\n3. No illegal content.\n\nFull rules page is coming soon.",
			EndDialogMsg: "Ending a dialog is coming soon.",
		}
	default:
		return Labels{
			MenuTitle:    "Главное меню",
			StartChat:    "Начать разговор",
			Profile:      "Профиль",
			Rules:        "Правила",
			EndDialog:    "Завершить диалог",
			Back:         "← Назад",
			StartChatMsg: "Поиск собеседника скоро будет доступен.",
			ProfileMsg:   "Раздел профиля скоро будет доступен.",
			RulesMsg:     "1. Будь уважителен к собеседнику.\n2. Не делись личными данными.\n3. Запрещён незаконный контент.\n\nПолная страница правил скоро появится.",
			EndDialogMsg: "Завершение диалога скоро будет доступно.",
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

func BackButton(labels Labels) models.InlineKeyboardButton {
	return models.InlineKeyboardButton{Text: labels.Back, CallbackData: CBBack}
}

func RemoveKeyboard() models.ReplyKeyboardRemove {
	return models.ReplyKeyboardRemove{RemoveKeyboard: true}
}
