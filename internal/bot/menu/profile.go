package menu

import (
	"fmt"
	"time"

	"github.com/go-telegram/bot/models"

	"github.com/flykby/anonimus_chat/internal/shared"
)

type ProfileViewData struct {
	PublicUUID       string
	Age              int16
	Gender           string
	Seeking          string
	Language         string
	PremiumActive    bool
	PremiumExpiresAt *time.Time
}

func ProfileViewText(data ProfileViewData, lang shared.Language) string {
	if lang == shared.LanguageEN {
		return fmt.Sprintf(
			"User %s\n\nPremium: %s\n\nProfile:\nGender: %s\nLooking for: %s\nAge: %d\nLanguage: %s",
			data.PublicUUID,
			formatPremiumStatus(data.PremiumActive, data.PremiumExpiresAt, lang),
			profileGenderLabelEN(shared.Gender(data.Gender)),
			profileSeekingLabelEN(shared.Gender(data.Seeking)),
			data.Age,
			languageLabel(data.Language),
		)
	}

	return fmt.Sprintf(
		"Пользователь %s\n\nPremium: %s\n\nАнкета:\nПол: %s\nИщу: %s\nВозраст: %d\nЯзык: %s",
		data.PublicUUID,
		formatPremiumStatus(data.PremiumActive, data.PremiumExpiresAt, lang),
		genderLabelRU(shared.Gender(data.Gender)),
		seekingLabelRU(shared.Gender(data.Seeking)),
		data.Age,
		languageLabel(data.Language),
	)
}

func ProfileViewButtons(labels Labels, premiumActive bool) [][]models.InlineKeyboardButton {
	premiumLabel := labels.ProfileBuyPremium
	if premiumActive {
		premiumLabel = labels.ProfileExtendPremium
	}
	return [][]models.InlineKeyboardButton{
		{{Text: premiumLabel, CallbackData: CBProfilePremium}},
		{{Text: labels.ProfileEdit, CallbackData: CBProfileEdit}},
		{{Text: labels.ProfileChangeLanguage, CallbackData: CBProfileLanguage}},
		{{Text: labels.ProfileDelete, CallbackData: CBProfileDelete}},
		{BackButton(labels)},
	}
}

func formatPremiumStatus(active bool, expiresAt *time.Time, lang shared.Language) string {
	if !active || expiresAt == nil {
		if lang == shared.LanguageEN {
			return "none"
		}
		return "отсутствует"
	}

	formatted := expiresAt.UTC().Format("02.01.2006 15:04") + " UTC+0"
	if lang == shared.LanguageEN {
		return "active until " + formatted
	}
	return "действует до " + formatted
}

func languageLabel(code string) string {
	switch shared.Language(code) {
	case shared.LanguageEN:
		return "EN"
	default:
		return "RU"
	}
}

func profileGenderLabelEN(g shared.Gender) string {
	switch g {
	case shared.GenderMale:
		return "Guy"
	case shared.GenderFemale:
		return "Girl"
	default:
		return string(g)
	}
}

func profileSeekingLabelEN(g shared.Gender) string {
	return profileGenderLabelEN(g)
}
