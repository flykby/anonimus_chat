package menu

import (
	"fmt"
	"time"

	"github.com/go-telegram/bot/models"

	"github.com/flykby/anonimus_chat/internal/bot/locales"
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
	return locales.T("profile.view.template", lang, map[string]string{
		"uuid":          data.PublicUUID,
		"premium":       formatPremiumStatus(data.PremiumActive, data.PremiumExpiresAt, lang),
		"gender":        locales.ProfileGenderLabel(shared.Gender(data.Gender), lang),
		"seeking":       locales.SeekingLabel(shared.Gender(data.Seeking), lang),
		"age":           fmtInt(data.Age),
		"language_code": locales.LanguageCode(shared.Language(data.Language)),
	})
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
		return locales.T("profile.premium.none", lang, nil)
	}
	return locales.T("profile.premium.active", lang, map[string]string{
		"date": expiresAt.UTC().Format("02.01.2006 15:04"),
	})
}

func fmtInt(v int16) string {
	return fmt.Sprint(v)
}
