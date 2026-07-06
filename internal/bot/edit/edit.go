package edit

import (
	"github.com/go-telegram/bot/models"

	"github.com/flykby/anonimus_chat/internal/bot/locales"
	"github.com/flykby/anonimus_chat/internal/bot/registration"
	"github.com/flykby/anonimus_chat/internal/shared"
)

const (
	StateAge     = "edit:age"
	StateGender  = "edit:gender"
	StateSeeking = "edit:seeking"

	CBAge           = "edit:age"
	CBGender        = "edit:gender"
	CBSeeking       = "edit:seeking"
	CBBack          = "edit:back"
	CBGenderMale    = "edit:gender:male"
	CBGenderFemale  = "edit:gender:female"
	CBSeekingMale   = "edit:seeking:male"
	CBSeekingFemale = "edit:seeking:female"
)

func MenuTitle(lang shared.Language) string {
	return locales.T("profile.edit_flow.menu_title", lang, nil)
}

func ButtonAge(lang shared.Language) string {
	return locales.T("profile.edit_flow.button_age", lang, nil)
}

func ButtonGender(lang shared.Language) string {
	return locales.T("profile.edit_flow.button_gender", lang, nil)
}

func ButtonSeeking(lang shared.Language) string {
	return locales.T("profile.edit_flow.button_seeking", lang, nil)
}

func Updated(lang shared.Language) string {
	return locales.T("profile.edit_flow.updated", lang, nil)
}

func SaveError(lang shared.Language) string {
	return locales.T("profile.edit_flow.save_error", lang, nil)
}

func ActiveDialog(lang shared.Language) string {
	return locales.T("profile.edit_flow.active_dialog", lang, nil)
}

func MenuButtons(lang shared.Language) [][]models.InlineKeyboardButton {
	return [][]models.InlineKeyboardButton{
		{{Text: ButtonAge(lang), CallbackData: CBAge}},
		{{Text: ButtonGender(lang), CallbackData: CBGender}},
		{{Text: ButtonSeeking(lang), CallbackData: CBSeeking}},
		{{Text: locales.T("menu.back", lang, nil), CallbackData: CBBack}},
	}
}

func GenderButtons(lang shared.Language) [][]models.InlineKeyboardButton {
	return [][]models.InlineKeyboardButton{
		{
			{Text: registration.GenderButtonMale(lang), CallbackData: CBGenderMale},
			{Text: registration.GenderButtonFemale(lang), CallbackData: CBGenderFemale},
		},
	}
}

func SeekingButtons(lang shared.Language) [][]models.InlineKeyboardButton {
	return [][]models.InlineKeyboardButton{
		{
			{Text: registration.SeekingButtonMale(lang), CallbackData: CBSeekingMale},
			{Text: registration.SeekingButtonFemale(lang), CallbackData: CBSeekingFemale},
		},
	}
}

func ParseGenderCallback(data string) (shared.Gender, bool) {
	switch data {
	case CBGenderMale, CBSeekingMale:
		return shared.GenderMale, true
	case CBGenderFemale, CBSeekingFemale:
		return shared.GenderFemale, true
	default:
		return "", false
	}
}

func IsEditState(state string) bool {
	switch state {
	case StateAge, StateGender, StateSeeking:
		return true
	default:
		return false
	}
}
