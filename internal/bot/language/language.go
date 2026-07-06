package language

import (
	"github.com/go-telegram/bot/models"

	"github.com/flykby/anonimus_chat/internal/bot/locales"
	"github.com/flykby/anonimus_chat/internal/bot/registration"
	"github.com/flykby/anonimus_chat/internal/shared"
)

const (
	CBRU = "lang:ru"
	CBEN = "lang:en"
)

func Prompt(lang shared.Language) string {
	return locales.T("profile.language_flow.prompt", lang, nil)
}

func Changed(lang shared.Language) string {
	return locales.T("profile.language_flow.changed", lang, map[string]string{
		"language": locales.LanguageCode(lang),
	})
}

func SaveError(lang shared.Language) string {
	return locales.T("profile.language_flow.save_error", lang, nil)
}

func ChoiceButtons(lang shared.Language) [][]models.InlineKeyboardButton {
	return [][]models.InlineKeyboardButton{
		{
			{Text: registration.LanguageButtonRU(lang), CallbackData: CBRU},
			{Text: registration.LanguageButtonEN(lang), CallbackData: CBEN},
		},
	}
}

func ParseCallback(data string) (shared.Language, bool) {
	switch data {
	case CBRU:
		return shared.LanguageRU, true
	case CBEN:
		return shared.LanguageEN, true
	default:
		return "", false
	}
}
